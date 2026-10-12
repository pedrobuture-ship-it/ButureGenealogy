package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"genetica/internal/apperr"
	"genetica/internal/datasets"
	"genetica/internal/domain"
)

func syntheticManifest(version string) datasets.Manifest {
	return datasets.Manifest{
		SchemaVersion: 1, ID: "synthetic.sires", Name: "Synthetic Sires", Version: version, Species: "Bos taurus",
		Breeds: []string{"NEL"}, Purpose: []string{"beef"}, SourceURL: "https://example.org/synthetic",
		License:    datasets.License{Name: "CC0-1.0", URL: "https://creativecommons.org/publicdomain/zero/1.0/", RedistributionAllowed: true},
		Evaluation: datasets.Evaluation{Program: "SYNTHETIC", GeneticBase: "synthetic-base"},
		Data:       datasets.DataSpec{Format: "csv", IDColumn: "registration", NameColumn: "name", AccuracyScale: "fraction"},
	}
}

var syntheticTraits = []datasets.TraitDef{
	{Code: "WW", Name: "Peso à desmama (synthetic)", Category: "growth", Unit: "kg", EvaluationType: "DEP", Column: "DEP_PD", AccuracyColumn: "ACC_PD", PercentileColumn: "PCT_PD"},
	{Code: "YW", Name: "Peso ao sobreano (synthetic)", Category: "growth", Unit: "kg", EvaluationType: "DEP", Column: "DEP_PS"},
}

const syntheticCSV = "registration,name,DEP_PD,ACC_PD,PCT_PD,DEP_PS\n" +
	"B-1,Bull One,5.2,0.81,70,8.1\n" +
	"B-2,Bull Two,3.0,0.50,40,\n" +
	"B-3,Bull Three,NA,,,1.0\n"

func writeGenlib(t *testing.T, dir, version, csv string) string {
	t.Helper()
	tb, err := datasets.ParseTable([]byte(csv), "csv")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := datasets.WritePackage(&datasets.Package{Manifest: syntheticManifest(version), Traits: syntheticTraits, Table: tb, License: "CC0 text", Readme: "# synthetic"})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "synthetic-"+version+".genlib")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func libFixture(t *testing.T) (fixture, string) {
	fx := newFixture(t)
	dir := t.TempDir()
	return fx, dir
}

func install(t *testing.T, fx fixture, req ImportRequest) *InstallResult {
	t.Helper()
	req.AcceptResponsibility = true
	res, err := fx.svc.Libraries.Install(context.Background(), req)
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	return res
}

func TestInstallGenlibMatchesExistingAndCreatesOnRequest(t *testing.T) {
	ctx := context.Background()
	fx, dir := libFixture(t)
	existing := fx.animal(t, "Bull One", domain.SexMale, domain.CategoryBull, func(i *AnimalInput) { i.Registration = " b-1 " }) // normalization: case + spaces
	path := writeGenlib(t, dir, "1.0.0", syntheticCSV)

	// inspection of a .genlib already validates it
	in, err := fx.svc.Libraries.Inspect(ctx, path)
	if err != nil || in.Kind != "genlib" || in.Validation == nil || !in.Validation.CanInstall {
		t.Fatalf("%+v %v", in, err)
	}
	rep := in.Validation.Report
	if rep.Records != 3 || rep.MatchedAnimals != 1 || rep.UnmatchedAnimals != 2 || rep.TraitsRecognized != 2 || rep.WithoutAccuracy != 2 {
		t.Fatalf("%+v", rep)
	}

	// default: skip animals that are not in the herd
	res := install(t, fx, ImportRequest{Path: path})
	if res.EvaluationsInserted != 2 || res.AnimalsCreated != 0 || res.RecordsSkipped != 2 {
		t.Fatalf("%+v", res)
	}
	lib := res.Library
	if lib.Name != "Synthetic Sires" || lib.TraitCount != 2 || lib.RecordCount != 3 || !lib.Active || lib.LicenseStatus != "redistributable" || len(lib.FileSHA256) != 64 {
		t.Fatalf("%+v", lib)
	}
	// the original package is kept on disk
	if _, err := os.Stat(filepath.Join(fx.svc.Libraries.dir, lib.ID+".genlib")); err != nil {
		t.Fatalf("package file not stored: %v", err)
	}
	groups, err := fx.svc.Traits.ForAnimal(ctx, existing.ID)
	if err != nil {
		t.Fatal(err)
	}
	var ww *TraitItem
	for gi := range groups {
		for ii := range groups[gi].Items {
			if groups[gi].Items[ii].TraitCode == "WW" {
				ww = &groups[gi].Items[ii]
			}
		}
	}
	if ww == nil || ww.TraitName != "Peso à desmama (synthetic)" || ww.Value != 5.2 || ww.Accuracy == nil || *ww.Accuracy != 0.81 || *ww.Percentile != 70 {
		t.Fatalf("%+v", ww)
	}
	p := ww.Provenance
	if p.DatasetName != "Synthetic Sires" || p.DatasetVersion != "1.0.0" || p.Program != "SYNTHETIC" || p.Release != "1.0.0" || p.GeneticBase != "synthetic-base" || !p.Genetic {
		t.Fatalf("provenance: %+v", p)
	}

	// a second version, creating unknown animals this time
	path2 := writeGenlib(t, dir, "2.0.0", syntheticCSV)
	res2 := install(t, fx, ImportRequest{Path: path2, UnmatchedPolicy: UnmatchedCreate, FarmID: fx.farmID})
	if res2.AnimalsCreated != 2 || res2.EvaluationsInserted != 4 {
		t.Fatalf("%+v", res2)
	}
	page, _ := fx.svc.Animals.List(ctx, AnimalQuery{Search: "Bull Two"})
	if page.Total != 1 || page.Items[0].Category != domain.CategoryBull || page.Items[0].Sex != domain.SexMale {
		t.Fatalf("%+v", page)
	}
	// both versions coexist as separate rows
	libs, _ := fx.svc.Libraries.List(ctx)
	if len(libs) != 2 {
		t.Fatalf("want 2 versions, got %d", len(libs))
	}
}

func TestInstallRequiresResponsibilityAndRejectsDuplicates(t *testing.T) {
	ctx := context.Background()
	fx, dir := libFixture(t)
	path := writeGenlib(t, dir, "1.0.0", syntheticCSV)

	_, err := fx.svc.Libraries.Install(ctx, ImportRequest{Path: path})
	wantField(t, err, "acceptResponsibility")
	_, err = fx.svc.Libraries.Install(ctx, ImportRequest{Path: path, AcceptResponsibility: true, UnmatchedPolicy: UnmatchedCreate})
	wantField(t, err, "farmId")

	install(t, fx, ImportRequest{Path: path})
	v, err := fx.svc.Libraries.Validate(ctx, ImportRequest{Path: path})
	if err != nil || !v.Duplicate || v.CanInstall {
		t.Fatalf("duplicate must be flagged: %+v %v", v, err)
	}
	if _, err := fx.svc.Libraries.Install(ctx, ImportRequest{Path: path, AcceptResponsibility: true}); apperr.CodeOf(err) != apperr.CodeValidation {
		t.Fatalf("duplicate install must fail, got %v", err)
	}
	libs, _ := fx.svc.Libraries.List(ctx)
	if len(libs) != 1 {
		t.Fatalf("a rejected install must leave nothing behind, got %d", len(libs))
	}
	// no stray temp files either
	entries, _ := os.ReadDir(fx.svc.Libraries.dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Fatalf("temp file left behind: %s", e.Name())
		}
	}
}

func TestInstallIsAtomicOnFailure(t *testing.T) {
	ctx := context.Background()
	fx, dir := libFixture(t)
	// a farm that does not exist makes the creation of the first animal fail mid-import
	path := writeGenlib(t, dir, "1.0.0", syntheticCSV)
	_, err := fx.svc.Libraries.Install(ctx, ImportRequest{Path: path, AcceptResponsibility: true, UnmatchedPolicy: UnmatchedCreate, FarmID: "no-such-farm"})
	if err == nil {
		t.Fatal("expected failure")
	}
	libs, _ := fx.svc.Libraries.List(ctx)
	traits, _ := fx.svc.Traits.List(ctx, TraitQuery{Kind: domain.EvalDEP})
	for _, tr := range traits {
		if tr.SourceDataset != "" {
			t.Fatalf("trait of a failed import left behind: %+v", tr)
		}
	}
	if len(libs) != 0 {
		t.Fatalf("dataset row left behind: %d", len(libs))
	}
	entries, _ := os.ReadDir(fx.svc.Libraries.dir)
	if len(entries) != 0 {
		t.Fatalf("files left behind: %v", entries)
	}
}

func TestImportCSVWithMapping(t *testing.T) {
	ctx := context.Background()
	fx, dir := libFixture(t)
	fx.animal(t, "Touro A", domain.SexMale, domain.CategoryBull, func(i *AnimalInput) { i.Registration = "RGD-10" })
	// semicolon-separated, decimal comma, accuracy in percent, Latin-1 name
	csv := "registro;nome;DEP_PD;ACC_PD;lixo\nRGD-10;Touro A;1,5;81;x\nRGD-11;Touro B;2,5;70;y\nRGD-10;Repetido;9;9;z\n;Sem id;1;1;1\n"
	path := filepath.Join(dir, "avaliacoes.csv")
	if err := os.WriteFile(path, []byte(csv), 0o600); err != nil {
		t.Fatal(err)
	}
	in, err := fx.svc.Libraries.Inspect(ctx, path)
	if err != nil || in.Kind != "csv" || !in.SuggestedDecimalComma || in.SuggestedIDColumn != "registro" || in.SuggestedNameColumn != "nome" || in.RowCount != 4 {
		t.Fatalf("%+v %v", in, err)
	}
	req := ImportRequest{
		Path: path,
		Manifest: datasets.Manifest{ID: "meu.csv", Name: "Meu CSV", Version: "2026.1", Species: "Bos indicus", Breeds: []string{"NEL"}, Purpose: []string{"beef"},
			Evaluation: datasets.Evaluation{Program: "Programa X"},
			Data:       datasets.DataSpec{IDColumn: "registro", NameColumn: "nome", DecimalComma: true, AccuracyScale: "percent"}},
		Traits:      []datasets.TraitDef{{Code: "PD", Name: "Peso à desmama", Category: "growth", Unit: "kg", EvaluationType: "DEP", Column: "DEP_PD", AccuracyColumn: "ACC_PD"}},
		LicenseText: "",
	}
	v, err := fx.svc.Libraries.Validate(ctx, req)
	if err != nil || !v.CanInstall {
		t.Fatalf("%+v %v", v, err)
	}
	r := v.Report
	if r.Records != 2 || r.Duplicates != 1 || r.RowsWithoutID != 1 || r.TraitsUnrecognized != 1 || r.UnrecognizedColumns[0] != "lixo" || r.MatchedAnimals != 1 || r.UnmatchedAnimals != 1 {
		t.Fatalf("%+v", r)
	}
	if v.LicenseStatus != "unknown" {
		t.Fatalf("a hand-imported dataset without a license must be 'unknown', got %s", v.LicenseStatus)
	}
	res := install(t, fx, req)
	if res.Library.Origin != "csv" || res.Library.LicenseStatus != "unknown" || res.EvaluationsInserted != 1 {
		t.Fatalf("%+v", res)
	}
	// the dataset is now a stored .genlib that can be read back
	if _, err := datasets.ReadPackageFile(filepath.Join(fx.svc.Libraries.dir, res.Library.ID+".genlib"), datasets.DefaultLimits); err != nil {
		t.Fatalf("stored package unreadable: %v", err)
	}
	d, err := fx.svc.Libraries.Get(ctx, res.Library.ID)
	if err != nil || d.ResponsibilityNotice != datasets.ResponsibilityNotice || len(d.Traits) != 1 || d.LicenseStatus != "unknown" || d.LicenseText != datasets.NoLicenseText || !strings.Contains(d.Readme, datasets.ResponsibilityNotice) {
		t.Fatalf("%+v %v", d, err)
	}
	// accuracy percent -> fraction
	animals, _ := fx.svc.Animals.List(ctx, AnimalQuery{Search: "Touro A"})
	evs, _ := fx.svc.Evaluations.ListByAnimal(ctx, animals.Items[0].ID)
	if len(evs) != 1 || evs[0].Value != 1.5 || evs[0].Accuracy == nil || *evs[0].Accuracy != 0.81 {
		t.Fatalf("%+v", evs)
	}
}

func TestImportRejectsBadInputs(t *testing.T) {
	ctx := context.Background()
	fx, dir := libFixture(t)
	if _, err := fx.svc.Libraries.Inspect(ctx, filepath.Join(dir, "x.exe")); apperr.CodeOf(err) != apperr.CodeValidation {
		t.Fatalf("unsupported extension: %v", err)
	}
	junk := filepath.Join(dir, "junk.genlib")
	_ = os.WriteFile(junk, []byte("not a zip"), 0o600)
	if _, err := fx.svc.Libraries.Inspect(ctx, junk); apperr.CodeOf(err) != apperr.CodeValidation {
		t.Fatalf("junk package: %v", err)
	}
	if _, err := fx.svc.Libraries.Inspect(ctx, filepath.Join(dir, "missing.csv")); err == nil {
		t.Fatal("missing file")
	}
	// mapping pointing at a column that does not exist blocks the install
	path := filepath.Join(dir, "a.csv")
	_ = os.WriteFile(path, []byte("reg,v\nR1,1\n"), 0o600)
	v, err := fx.svc.Libraries.Validate(ctx, ImportRequest{Path: path,
		Manifest: datasets.Manifest{ID: "x.y", Name: "X", Version: "1", Species: "s", Breeds: []string{"b"}, Purpose: []string{"beef"},
			Evaluation: datasets.Evaluation{Program: "P"}, Data: datasets.DataSpec{IDColumn: "reg"}},
		Traits: []datasets.TraitDef{{Code: "A", Name: "A", EvaluationType: "DEP", Column: "nope"}}})
	if err != nil || v.CanInstall || len(v.Report.Errors) == 0 {
		t.Fatalf("%+v %v", v, err)
	}
	// an invalid manifest never reaches the data
	v, _ = fx.svc.Libraries.Validate(ctx, ImportRequest{Path: path, Manifest: datasets.Manifest{ID: "../evil"}})
	if v.CanInstall {
		t.Fatal("invalid manifest must block")
	}
}

func TestAmbiguousRegistrationsAreSkipped(t *testing.T) {
	ctx := context.Background()
	fx, dir := libFixture(t)
	other, _ := fx.svc.Farms.Create(ctx, FarmInput{Name: "Outra"})
	fx.animal(t, "A1", domain.SexMale, domain.CategoryBull, func(i *AnimalInput) { i.Registration = "DUP" })
	if _, err := fx.svc.Animals.Create(ctx, AnimalInput{FarmID: other.ID, Name: "A2", Sex: domain.SexMale, Category: domain.CategoryBull, Registration: "DUP"}); err != nil {
		t.Fatal(err)
	}
	path := writeGenlib(t, dir, "1.0.0", "registration,name,DEP_PD,ACC_PD,PCT_PD,DEP_PS\nDUP,x,1,0.5,50,2\n")
	v, _ := fx.svc.Libraries.Validate(ctx, ImportRequest{Path: path})
	if v.Report.AmbiguousAnimals != 1 {
		t.Fatalf("%+v", v.Report)
	}
	res, err := fx.svc.Libraries.Install(ctx, ImportRequest{Path: path, AcceptResponsibility: true})
	if err != nil || res.EvaluationsInserted != 0 || res.RecordsSkipped != 1 {
		t.Fatalf("an ambiguous match must never be guessed: %+v %v", res, err)
	}
}

func TestActivateDeactivateHidesValues(t *testing.T) {
	ctx := context.Background()
	fx, dir := libFixture(t)
	a := fx.animal(t, "Bull One", domain.SexMale, domain.CategoryBull, func(i *AnimalInput) { i.Registration = "B-1" })
	res := install(t, fx, ImportRequest{Path: writeGenlib(t, dir, "1.0.0", syntheticCSV)})

	count := func() int {
		evs, _ := fx.svc.Evaluations.ListByAnimal(ctx, a.ID)
		return len(evs)
	}
	if count() != 2 {
		t.Fatalf("got %d", count())
	}
	if err := fx.svc.Libraries.SetActive(ctx, res.Library.ID, false); err != nil {
		t.Fatal(err)
	}
	if count() != 0 {
		t.Fatal("an inactive library must not feed the animal's evaluations")
	}
	groups, _ := fx.svc.Traits.ForAnimal(ctx, a.ID)
	if len(groups) != 0 {
		t.Fatal("inactive library values must not appear in the trait view")
	}
	_ = fx.svc.Libraries.SetActive(ctx, res.Library.ID, true)
	if count() != 2 {
		t.Fatal("reactivation must restore the values")
	}
	if err := fx.svc.Libraries.SetActive(ctx, "nope", true); apperr.CodeOf(err) != apperr.CodeNotFound {
		t.Fatalf("%v", err)
	}
}

func TestRemoveAndHistoricalProtection(t *testing.T) {
	ctx := context.Background()
	fx, dir := libFixture(t)
	bull := fx.animal(t, "Bull One", domain.SexMale, domain.CategoryBull, func(i *AnimalInput) { i.Registration = "B-1" })
	cow := fx.animal(t, "Cow", domain.SexFemale, domain.CategoryCow)
	used := install(t, fx, ImportRequest{Path: writeGenlib(t, dir, "1.0.0", syntheticCSV)})
	unused := install(t, fx, ImportRequest{Path: writeGenlib(t, dir, "2.0.0", syntheticCSV)})

	// a saved calculation that used version 1.0.0
	m, err := fx.svc.Matings.Save(ctx, domain.Mating{FarmID: fx.farmID, FemaleID: cow.ID, BullID: bull.ID, EngineVersion: "t", CalculationSnapshot: `{}`})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fx.svc.Libraries.db.ExecContext(ctx, `INSERT INTO calculation_datasets (mating_id, dataset_id) VALUES (?, ?)`, m.ID, used.Library.ID); err != nil {
		t.Fatal(err)
	}
	err = fx.svc.Libraries.Remove(ctx, used.Library.ID)
	if apperr.CodeOf(err) != apperr.CodeConflict {
		t.Fatalf("a version used by a calculation must not be removable, got %v", err)
	}
	d, _ := fx.svc.Libraries.Get(ctx, used.Library.ID)
	if d.UsedByCalculations != 1 {
		t.Fatalf("%+v", d)
	}

	if err := fx.svc.Libraries.Remove(ctx, unused.Library.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(fx.svc.Libraries.dir, unused.Library.ID+".genlib")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("stored package must be deleted with the library")
	}
	libs, _ := fx.svc.Libraries.List(ctx)
	if len(libs) != 1 || libs[0].ID != used.Library.ID {
		t.Fatalf("%+v", libs)
	}
	// the removed version's traits and values are gone; the other version's remain
	evs, _ := fx.svc.Evaluations.ListByAnimal(ctx, bull.ID)
	if len(evs) != 2 {
		t.Fatalf("only the removed version's values must be deleted, got %d", len(evs))
	}
	if err := fx.svc.Libraries.Remove(ctx, "nope"); apperr.CodeOf(err) != apperr.CodeNotFound {
		t.Fatalf("%v", err)
	}
}

func TestOfficialDistributionGate(t *testing.T) {
	// the rule that keeps non-redistributable data out of the official build
	m := syntheticManifest("1")
	if err := m.CheckBundleable(); err != nil {
		t.Fatal(err)
	}
	m.License.RedistributionAllowed = false
	if err := m.CheckBundleable(); err == nil {
		t.Fatal("non-redistributable data must be refused")
	}
	m.License = datasets.License{}
	if err := m.CheckBundleable(); err == nil {
		t.Fatal("unknown license must be refused")
	}
}

func TestInstallLargeDatasetIsFast(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	fx, dir := libFixture(t)
	var b strings.Builder
	b.WriteString("registration,name,DEP_PD,ACC_PD,PCT_PD,DEP_PS\n")
	for i := 0; i < 8412; i++ {
		fmt.Fprintf(&b, "SYN-%05d,Bull %d,%.2f,0.%02d,%d,%.2f\n", i, i, float64(i%50)/10, 50+i%49, i%100, float64(i%70)/10)
	}
	path := writeGenlib(t, dir, "1.0.0", b.String())
	start := time.Now()
	res := install(t, fx, ImportRequest{Path: path, UnmatchedPolicy: UnmatchedCreate, FarmID: fx.farmID})
	if res.AnimalsCreated != 8412 || res.EvaluationsInserted != 8412*2 {
		t.Fatalf("%+v", res)
	}
	if d := time.Since(start); d > 20*time.Second {
		t.Fatalf("install took %v", d)
	}
	t.Logf("installed %d animals / %d values in %v", res.AnimalsCreated, res.EvaluationsInserted, time.Since(start))
}

func TestMappingToExistingCatalogueTraits(t *testing.T) {
	ctx := context.Background()
	fx, dir := libFixture(t)
	path := filepath.Join(dir, "a.csv")
	_ = os.WriteFile(path, []byte("reg,ww\nR1,5\n"), 0o600)
	mk := func(traitID string) ImportRequest {
		return ImportRequest{Path: path, AcceptResponsibility: true,
			Manifest: datasets.Manifest{ID: "m.x", Name: "M", Version: "1", Species: "s", Breeds: []string{"NEL"}, Purpose: []string{"beef"},
				Evaluation: datasets.Evaluation{Program: "P"}, Data: datasets.DataSpec{IDColumn: "reg"}},
			Traits: []datasets.TraitDef{{Code: "WW", Name: "WW", EvaluationType: "DEP", Column: "ww", ExistingTraitID: traitID}}}
	}
	check := func(id, wantErr string) {
		t.Helper()
		v, err := fx.svc.Libraries.Validate(ctx, mk(id))
		if err != nil {
			t.Fatal(err)
		}
		if wantErr == "" {
			if !v.CanInstall {
				t.Fatalf("%+v", v.Report.Errors)
			}
			return
		}
		if v.CanInstall || !strings.Contains(strings.Join(v.Report.Errors, " "), wantErr) {
			t.Fatalf("want error %q, got %+v", wantErr, v.Report.Errors)
		}
	}
	check("builtin:DEP:WW", "")                        // a catalogue genetic trait
	check("builtin:PHENOTYPE:BODY_WEIGHT", "fenótipo") // phenotypes are not library values
	check("no-such-trait", "não existe")
	// a trait owned by another library cannot be borrowed
	first := install(t, fx, ImportRequest{Path: writeGenlib(t, dir, "1.0.0", syntheticCSV)})
	owned, _ := fx.svc.Libraries.Get(ctx, first.Library.ID)
	check(owned.Traits[0].ID, "outra biblioteca")
}
