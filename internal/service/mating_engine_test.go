package service

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"genetica/internal/apperr"
	"genetica/internal/datasets"
	"genetica/internal/domain"
	"genetica/internal/engine"
)

func demoIDs(t *testing.T, fx fixture) map[string]string {
	t.Helper()
	ctx := context.Background()
	if err := fx.svc.Demo.Load(ctx); err != nil {
		t.Fatal(err)
	}
	page, err := fx.svc.Animals.List(ctx, AnimalQuery{IncludeArchived: true, PageSize: 500})
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]string{}
	for _, a := range page.Items {
		ids[a.Name] = a.ID
	}
	return ids
}

func nearf(t *testing.T, got, want float64, what string) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("%s: got %.10f want %.10f", what, got, want)
	}
}

func result(t *testing.T, out engine.RunOutput, bullID string) engine.CandidateResult {
	t.Helper()
	for _, r := range out.Results {
		if r.Bull.ID == bullID {
			return r
		}
	}
	t.Fatalf("bull %s not in results", bullID)
	return engine.CandidateResult{}
}

func TestRunDemoBeefMatchesHandCalculation(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	ids := demoIDs(t, fx)
	prof := engine.DemoBeefProfile()

	resp, err := fx.svc.Matings.Run(ctx, RunRequest{FemaleID: ids["DEMO Matriz 0472"], Profile: prof,
		BullIDs: []string{ids["DEMO Touro Alfa"], ids["DEMO Touro Beta"], ids["DEMO Touro Gama"]}})
	if err != nil || len(resp.Output.Errors) != 0 {
		t.Fatalf("%v %v", err, resp.Output.Errors)
	}
	a, b, g := result(t, resp.Output, ids["DEMO Touro Alfa"]), result(t, resp.Output, ids["DEMO Touro Beta"]), result(t, resp.Output, ids["DEMO Touro Gama"])

	// Alfa: base .4*82+.35*78+.25*55 = 73.85 ; bonus 10*(.4*.24+.35*.26+.25*0) = 1.87 ; no accuracy penalty
	nearf(t, a.Final.Base, 73.85, "Alfa base")
	nearf(t, a.Final.Complementarity, 1.87, "Alfa complementarity")
	nearf(t, a.Final.Final, 75.72, "Alfa final")
	// Beta: base 67.4 ; bonus 10*(.4*.02+.35*.12+.25*.13) = 0.825
	nearf(t, b.Final.Base, 67.4, "Beta base")
	nearf(t, b.Final.Final, 68.225, "Beta final")
	if a.Rank != 1 || b.Rank != 2 {
		t.Fatalf("%d %d", a.Rank, b.Rank)
	}
	// Gama was evaluated by ANOTHER program: his higher values must not be compared
	if g.Status != engine.StatusNotCalculable || g.Final != nil {
		t.Fatalf("Gama must be not calculable: %+v", g.Status)
	}
	found := false
	for _, w := range g.Warnings {
		found = found || w == "WW: "+engine.IncompatibleBasesMessage
	}
	if !found {
		t.Fatalf("%v", g.Warnings)
	}
	// provenance reaches the component: dataset name and version
	c0 := a.Components[0]
	if c0.Source.DatasetName != "DEMO / SYNTHETIC Avaliações" || c0.Source.DatasetVersion != "demo-1" || !c0.Source.IsDemo || c0.Method != "percentile" {
		t.Fatalf("%+v", c0.Source)
	}
	if !resp.Output.Profile.IsDemo || resp.Output.Profile.DemoNotice != "DEMO — NÃO REPRESENTA RECOMENDAÇÃO ZOOTÉCNICA VALIDADA" {
		t.Fatalf("%+v", resp.Output.Profile)
	}
	if resp.Saved || len(resp.MatingIDs) != 0 {
		t.Fatal("a preview must not save anything")
	}
	if list, _ := fx.svc.Matings.ListCalculations(ctx, ids["DEMO Matriz 0472"]); len(list) != 0 {
		t.Fatalf("preview left rows behind: %d", len(list))
	}
}

func TestRunDemoDairyMinimizeTrait(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	ids := demoIDs(t, fx)
	resp, err := fx.svc.Matings.Run(ctx, RunRequest{FemaleID: ids["DEMO Vaca Holandesa V1"], Profile: engine.DemoDairyProfile(),
		BullIDs: []string{ids["DEMO Touro Holandês H1"], ids["DEMO Touro Holandês H2"]}})
	if err != nil || len(resp.Output.Errors) != 0 {
		t.Fatalf("%v %v", err, resp.Output.Errors)
	}
	h1, h2 := result(t, resp.Output, ids["DEMO Touro Holandês H1"]), result(t, resp.Output, ids["DEMO Touro Holandês H2"])
	// H1: base .5*85+.3*60+.2*(100-30) = 74.5 ; bonus 8*(.5*.45+.3*.10+.2*.40)=2.68
	nearf(t, h1.Final.Base, 74.5, "H1 base")
	nearf(t, h1.Final.Final, 77.18, "H1 final")
	// H2: base 63 ; bonus 1.76 ; low-accuracy penalty 5*(.5*(.05/.6)+.3*(.08/.6)+.2*(.1/.6)) = 0.575
	nearf(t, h2.Final.Base, 63, "H2 base")
	nearf(t, h2.Final.Penalties, 0.575, "H2 penalty")
	nearf(t, h2.Final.Final, 64.185, "H2 final")
	scs := h1.Components[2]
	if scs.TraitCode != "SCS" || scs.Direction != engine.Minimize || math.Abs(*scs.Score-70) > 1e-9 {
		t.Fatalf("MINIMIZE inverts the percentile: %+v", scs)
	}
}

func TestRunPedigreeWarningsAndAncestorFilter(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	ids := demoIDs(t, fx)
	// heifer1 is the daughter of bullA: bullA is her father
	req := RunRequest{FemaleID: ids["DEMO Novilha 1101"], Profile: engine.DemoBeefProfile(), BullIDs: []string{ids["DEMO Touro Alfa"], ids["DEMO Touro Beta"]}}
	resp, err := fx.svc.Matings.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	a := result(t, resp.Output, ids["DEMO Touro Alfa"])
	if len(a.PedigreeNotes) == 0 {
		t.Fatalf("the father must be flagged: %+v", a)
	}
	if a.Complementarity.Status != engine.StatusNotCalculable {
		t.Fatalf("the heifer has no evaluations, so complementarity is not calculated: %s", a.Complementarity.Status)
	}
	if result(t, resp.Output, ids["DEMO Touro Beta"]).Status != engine.StatusCalculated {
		t.Fatal("Beta stays calculable")
	}
	// with the filter on, the father is removed from the ranking
	req.Profile.Filters.ExcludeDirectAncestors = true
	resp, _ = fx.svc.Matings.Run(ctx, req)
	if r := result(t, resp.Output, ids["DEMO Touro Alfa"]); r.Status != engine.StatusExcluded {
		t.Fatalf("%+v", r.Status)
	}
}

func TestRunValidationErrors(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	ids := demoIDs(t, fx)
	_, err := fx.svc.Matings.Run(ctx, RunRequest{FemaleID: "nope", Profile: engine.DemoBeefProfile(), BullIDs: []string{ids["DEMO Touro Alfa"]}})
	wantField(t, err, "femaleId")
	_, err = fx.svc.Matings.Run(ctx, RunRequest{FemaleID: ids["DEMO Matriz 0472"], Profile: engine.DemoBeefProfile()})
	wantField(t, err, "bullIds")
	_, err = fx.svc.Matings.Run(ctx, RunRequest{FemaleID: ids["DEMO Matriz 0472"], Profile: engine.DemoBeefProfile(), BullIDs: []string{"nope"}})
	wantField(t, err, "bullIds")
	// an empty profile is reported by the engine as a run error, not as a crash or a made-up ranking
	resp, err := fx.svc.Matings.Run(ctx, RunRequest{FemaleID: ids["DEMO Matriz 0472"], Profile: engine.CustomProfile(), BullIDs: []string{ids["DEMO Touro Alfa"]}, Save: true})
	if err != nil || len(resp.Output.Errors) == 0 || resp.Saved || len(resp.Output.Results) != 0 {
		t.Fatalf("%v %+v", err, resp)
	}
	// a male as the dam
	resp, _ = fx.svc.Matings.Run(ctx, RunRequest{FemaleID: ids["DEMO Touro Beta"], Profile: engine.DemoBeefProfile(), BullIDs: []string{ids["DEMO Touro Alfa"]}})
	if len(resp.Output.Errors) == 0 {
		t.Fatal("the dam must be a female")
	}
}

func TestSaveRunKeepsReproducibleSnapshots(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	ids := demoIDs(t, fx)
	cow := ids["DEMO Matriz 0472"]
	resp, err := fx.svc.Matings.Run(ctx, RunRequest{FemaleID: cow, Profile: engine.DemoBeefProfile(), Save: true,
		BullIDs: []string{ids["DEMO Touro Alfa"], ids["DEMO Touro Beta"], ids["DEMO Touro Gama"]}})
	if err != nil || !resp.Saved || resp.RunID == "" || len(resp.MatingIDs) != 3 {
		t.Fatalf("%v %+v", err, resp)
	}
	alfaID := resp.MatingIDs[ids["DEMO Touro Alfa"]]
	view, err := fx.svc.Matings.GetCalculation(ctx, alfaID)
	if err != nil || view.EngineVersion != engine.Version || view.Score == nil || !view.IsDemo || view.RunID != resp.RunID {
		t.Fatalf("%+v %v", view, err)
	}
	nearf(t, *view.Score, 75.72, "stored score")
	var snap engine.PairSnapshot
	if err := json.Unmarshal(view.Snapshot, &snap); err != nil {
		t.Fatal(err)
	}
	if snap.SnapshotVersion != 1 || snap.EngineVersion != engine.Version || snap.Profile.Name != "Demonstração Corte" || len(snap.Datasets) != 1 ||
		snap.Datasets[0].Version != "demo-1" || snap.Datasets[0].Name != "DEMO / SYNTHETIC Avaliações" || !snap.IsDemo || len(snap.BullEvals) != 3 || len(snap.FemaleEvals) != 3 {
		t.Fatalf("%+v", snap)
	}
	nearf(t, snap.Result.Final.Final, 75.72, "snapshot final")

	// the not-calculable pair is saved too (score NULL), explaining itself
	gView, _ := fx.svc.Matings.GetCalculation(ctx, resp.MatingIDs[ids["DEMO Touro Gama"]])
	if gView.Score != nil {
		t.Fatal("a not-calculable pair has no score")
	}
	var gSnap engine.PairSnapshot
	_ = json.Unmarshal(gView.Snapshot, &gSnap)
	if gSnap.Result.Status != engine.StatusNotCalculable || len(gSnap.Datasets) == 0 {
		t.Fatalf("%+v", gSnap.Result.Status)
	}

	// history: deactivating or changing data later does NOT change the saved calculation
	libs, _ := fx.svc.Libraries.List(ctx)
	for _, l := range libs {
		if err := fx.svc.Libraries.SetActive(ctx, l.ID, false); err != nil {
			t.Fatal(err)
		}
		// a library used by a saved calculation cannot be removed
		if err := fx.svc.Libraries.Remove(ctx, l.ID); apperr.CodeOf(err) != apperr.CodeConflict {
			t.Fatalf("a library used by a calculation must be protected, got %v", err)
		}
	}
	again, _ := fx.svc.Matings.GetCalculation(ctx, alfaID)
	if string(again.Snapshot) != string(view.Snapshot) || *again.Score != *view.Score {
		t.Fatal("a saved calculation must never change")
	}
	// and a new run with the datasets deactivated no longer finds evaluations — proving the snapshot was the only source
	re, _ := fx.svc.Matings.Run(ctx, RunRequest{FemaleID: cow, Profile: engine.DemoBeefProfile(), BullIDs: []string{ids["DEMO Touro Alfa"]}})
	if r := result(t, re.Output, ids["DEMO Touro Alfa"]); r.Status != engine.StatusNotCalculable {
		t.Fatalf("with the library inactive the bull has no usable data: %s", r.Status)
	}

	// history list
	list, err := fx.svc.Matings.ListCalculations(ctx, cow)
	if err != nil || len(list) != 3 {
		t.Fatalf("%+v %v", list, err)
	}
	var top CalculationSummary
	for _, c := range list {
		if c.BullName == "DEMO Touro Alfa" {
			top = c
		}
	}
	if top.Rank != 1 || top.ProfileName != "Demonstração Corte" || top.Score == nil || top.Status != domain.MatingDraft {
		t.Fatalf("%+v", top)
	}
	if err := fx.svc.Matings.SetStatus(ctx, alfaID, domain.MatingPlanned); err != nil {
		t.Fatal(err)
	}
	if err := fx.svc.Matings.SetStatus(ctx, alfaID, "whatever"); err == nil {
		t.Fatal("invalid status")
	}
	if _, err := fx.svc.Matings.GetCalculation(ctx, "nope"); apperr.CodeOf(err) != apperr.CodeNotFound {
		t.Fatal("unknown calculation")
	}
	// DEMO data (and its calculations) can be removed together
	if err := fx.svc.Demo.Remove(ctx); err != nil {
		t.Fatal(err)
	}
	if list, _ := fx.svc.Matings.ListCalculations(ctx, cow); len(list) != 0 {
		t.Fatal("demo calculations must be removed with the demo data")
	}
}

func TestProfilesCRUD(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	ps, err := fx.svc.Matings.Profiles(ctx)
	if err != nil || len(ps) != 3 || ps[0].Name != "Personalizado" || !ps[1].IsDemo || !ps[2].IsDemo {
		t.Fatalf("%+v %v", ps, err)
	}
	for _, p := range ps[1:] {
		if p.DemoNotice != engine.DemoNotice {
			t.Fatalf("every demo profile carries the notice: %+v", p)
		}
	}
	p := engine.DemoBeefProfile()
	p.Name = "Meu perfil"
	saved, err := fx.svc.Matings.SaveProfile(ctx, p) // copying a built-in creates a user profile
	if err != nil || saved.IsDemo || saved.DemoNotice != "" || saved.ID == engine.ProfileDemoBeefID || saved.ID == "" {
		t.Fatalf("a copy of a DEMO profile is the user's own profile, not DEMO: %+v %v", saved, err)
	}
	saved.Traits[0].Weight = 99
	if _, err := fx.svc.Matings.SaveProfile(ctx, *saved); err != nil {
		t.Fatal(err)
	}
	ps, _ = fx.svc.Matings.Profiles(ctx)
	if len(ps) != 4 || ps[3].Traits[0].Weight != 99 {
		t.Fatalf("%+v", ps)
	}
	if _, err := fx.svc.Matings.SaveProfile(ctx, engine.Profile{Name: "  "}); err == nil {
		t.Fatal("name required")
	}
	dup := engine.DemoBeefProfile()
	dup.Name = "Meu perfil"
	if _, err := fx.svc.Matings.SaveProfile(ctx, dup); err == nil {
		t.Fatal("duplicate name")
	}
	bad := engine.DemoBeefProfile()
	bad.Name = "Inválido"
	bad.Traits[0].Weight = -5
	if _, err := fx.svc.Matings.SaveProfile(ctx, bad); err == nil {
		t.Fatal("negative weight")
	}
	if err := fx.svc.Matings.DeleteProfile(ctx, engine.ProfileDemoBeefID); err == nil {
		t.Fatal("built-in profiles cannot be deleted")
	}
	if err := fx.svc.Matings.DeleteProfile(ctx, saved.ID); err != nil {
		t.Fatal(err)
	}
	if err := fx.svc.Matings.DeleteProfile(ctx, saved.ID); apperr.CodeOf(err) != apperr.CodeNotFound {
		t.Fatal("already gone")
	}
}

// A percentile imported from a user's CSV is only usable when the user declared what it means.
func TestImportedPercentilesNeedDeclaredSemantics(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name      string
		semantics map[string]any
		wantCalc  bool
	}{
		{"declarada", map[string]any{"percentileSemantics": "ascending"}, true},
		{"não declarada", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx := newFixture(t)
			fx.animal(t, "Touro CSV", domain.SexMale, domain.CategoryBull, func(i *AnimalInput) { i.Registration = "B-1" })
			bull, _ := fx.svc.Animals.List(ctx, AnimalQuery{Search: "Touro CSV"})
			cow := fx.animal(t, "Vaca", domain.SexFemale, domain.CategoryCow)
			if _, err := fx.svc.Semen.AddLot(ctx, SemenInput{BullID: bull.Items[0].ID, Quantity: 5}); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "a.csv")
			if err := os.WriteFile(path, []byte("reg,ww,pct\nB-1,5,80\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := fx.svc.Libraries.Install(ctx, ImportRequest{Path: path, AcceptResponsibility: true,
				Manifest: datasets.Manifest{ID: "csv.x", Name: "CSV X", Version: "1", Species: "s", Breeds: []string{"NEL"}, Purpose: []string{"beef"},
					Evaluation: datasets.Evaluation{Program: "P"}, Data: datasets.DataSpec{IDColumn: "reg"}},
				Traits: []datasets.TraitDef{{Code: "WW", Name: "WW", Category: "growth", Unit: "kg", EvaluationType: "DEP", Column: "ww", PercentileColumn: "pct", Metadata: tc.semantics}}})
			if err != nil {
				t.Fatal(err)
			}
			prof := engine.CustomProfile()
			prof.Traits = []engine.TraitWeight{{TraitCode: "WW", EvaluationType: "DEP", Weight: 1, Direction: engine.Maximize}}
			resp, err := fx.svc.Matings.Run(ctx, RunRequest{FemaleID: cow.ID, BullIDs: []string{bull.Items[0].ID}, Profile: prof})
			if err != nil || len(resp.Output.Errors) != 0 {
				t.Fatalf("%v %v", err, resp.Output.Errors)
			}
			r := resp.Output.Results[0]
			if tc.wantCalc {
				if r.Status != engine.StatusCalculated || r.Components[0].Method != "percentile" || *r.Components[0].Score != 80 {
					t.Fatalf("%+v", r.Components[0])
				}
				return
			}
			if r.Status != engine.StatusNotCalculable || !strings.Contains(r.Components[0].Explanation, "semântica do percentil") {
				t.Fatalf("an undeclared percentile meaning must not be guessed: %+v", r.Components[0])
			}
		})
	}
}
