package engine

import (
	"encoding/json"
	"math"
	"math/rand"
	"strings"
	"testing"
	"time"

	"genetica/internal/genetics"
)

// ---- helpers (all data is SYNTHETIC)

func fp(v float64) *float64 { return &v }

var t0 = time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)

func basis(trait string) Basis {
	return Basis{TraitCode: trait, EvaluationType: "DEP", Program: "SYN", Release: "r1", GeneticBase: "base-A", BreedCode: "NEL", Unit: "kg"}
}

// pct builds a percentile-bearing evaluation with ascending percentile semantics.
func pct(id, trait string, percentile float64, acc *float64) Evaluation {
	return Evaluation{ID: id, Basis: basis(trait), Value: percentile / 10, Accuracy: acc, Percentile: fp(percentile),
		DatasetID: "ds1", DatasetName: "Synthetic", DatasetVersion: "1", TraitName: trait, Direction: Maximize,
		Meta: TraitMeta{PercentileSemantics: PercentileAscending}, CreatedAt: "2026-01-01T00:00:00Z"}
}

func bull(id, name string, evals ...Evaluation) Candidate {
	return Candidate{Animal: Animal{ID: id, Name: name, Sex: "male", Status: "active"}, Evals: evals, Stock: Stock{Doses: 10, UnitCost: fp(50)}}
}

func dam(evals ...Evaluation) Female {
	return Female{Animal: Animal{ID: "dam", Name: "Matriz", Sex: "female", Status: "active"}, Evals: evals}
}

func profile(traits ...TraitWeight) Profile {
	return Profile{ID: "p", Name: "Perfil de teste", Traits: traits, Filters: Filters{RequireStock: true}, Normalization: NormAuto, MinCoverage: 1}
}

func tw(code string, w float64) TraitWeight {
	return TraitWeight{TraitCode: code, EvaluationType: "DEP", Weight: w}
}

func run(t *testing.T, in RunInput) RunOutput {
	t.Helper()
	if in.Now.IsZero() {
		in.Now = t0
	}
	return New().Run(in)
}

func near(t *testing.T, got, want float64, what string) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("%s: got %.12f, want %.12f", what, got, want)
	}
}

func byID(t *testing.T, out RunOutput, id string) CandidateResult {
	t.Helper()
	for _, r := range out.Results {
		if r.Bull.ID == id {
			return r
		}
	}
	t.Fatalf("candidate %s missing from results", id)
	return CandidateResult{}
}

// ---- normalization

func TestNormalizePercentile(t *testing.T) {
	n := DefaultNormalizer{}
	e := pct("e", "WW", 80, nil)
	r := n.Normalize(NormalizeInput{Weight: tw("WW", 1), Eval: e})
	near(t, *r.Score, 80, "ascending maximize")
	if r.Method != NormPercentile || r.Formula != "score = percentil" {
		t.Fatalf("%+v", r)
	}
	e.Direction = Minimize
	near(t, *n.Normalize(NormalizeInput{Weight: tw("WW", 1), Eval: e}).Score, 20, "ascending minimize")

	e.Direction = Maximize
	e.Meta.PercentileSemantics = PercentileDescending
	e.Percentile = fp(10)
	near(t, *n.Normalize(NormalizeInput{Weight: tw("WW", 1), Eval: e}).Score, 90, "descending maximize: top 10% => 90")
	w := tw("WW", 1)
	w.Direction = Minimize // profile override
	near(t, *n.Normalize(NormalizeInput{Weight: w, Eval: e}).Score, 10, "descending minimize")
}

func TestPercentileRequiresDeclaredSemantics(t *testing.T) {
	e := pct("e", "WW", 80, nil)
	e.Meta.PercentileSemantics = ""
	r := DefaultNormalizer{}.Normalize(NormalizeInput{Weight: tw("WW", 1), Eval: e})
	if r.Score != nil || !strings.Contains(r.Reason, "semântica do percentil") {
		t.Fatalf("an undeclared percentile meaning must not be guessed: %+v", r)
	}
	e.Meta.PercentileSemantics = PercentileAscending
	e.Percentile = fp(101)
	if (DefaultNormalizer{}).Normalize(NormalizeInput{Weight: tw("WW", 1), Eval: e}).Score != nil {
		t.Fatal("percentile out of range")
	}
}

func TestNormalizeZScore(t *testing.T) {
	e := Evaluation{Basis: basis("WW"), Value: 12, Direction: Maximize, Meta: TraitMeta{PopulationMean: fp(10), PopulationSD: fp(2)}}
	n := DefaultNormalizer{}
	r := n.Normalize(NormalizeInput{Weight: tw("WW", 1), Eval: e})
	near(t, *r.Score, 100*Phi(1), "z=1")
	if math.Abs(*r.Score-84.1344746) > 1e-5 || r.Method != NormZScore {
		t.Fatalf("%+v", r)
	}
	e.Direction = Minimize
	near(t, *n.Normalize(NormalizeInput{Weight: tw("WW", 1), Eval: e}).Score, 100-100*Phi(1), "minimize")
	e.Meta.PopulationSD = fp(0)
	if n.Normalize(NormalizeInput{Weight: tw("WW", 1), Eval: e}).Score != nil {
		t.Fatal("sd=0 must not produce a score")
	}
	near(t, Phi(0), 0.5, "Phi(0)")
}

func TestNormalizeModes(t *testing.T) {
	e := pct("e", "WW", 70, nil)
	e.Meta.PopulationMean, e.Meta.PopulationSD = fp(0), fp(1)
	e.Value = 1
	n := DefaultNormalizer{}
	if r := n.Normalize(NormalizeInput{Weight: tw("WW", 1), Eval: e, Mode: NormAuto}); r.Method != NormPercentile {
		t.Fatalf("auto prefers percentile: %+v", r)
	}
	if r := n.Normalize(NormalizeInput{Weight: tw("WW", 1), Eval: e, Mode: NormZScore}); r.Method != NormZScore {
		t.Fatalf("%+v", r)
	}
	e.Percentile = nil
	if r := n.Normalize(NormalizeInput{Weight: tw("WW", 1), Eval: e, Mode: NormAuto}); r.Method != NormZScore {
		t.Fatalf("auto falls back to z-score: %+v", r)
	}
	if r := n.Normalize(NormalizeInput{Weight: tw("WW", 1), Eval: e, Mode: NormPercentile}); r.Score != nil {
		t.Fatalf("forced percentile without percentile must fail: %+v", r)
	}
	e.Meta = TraitMeta{}
	if r := n.Normalize(NormalizeInput{Weight: tw("WW", 1), Eval: e}); r.Score != nil || !strings.Contains(r.Reason, "nenhum método") {
		t.Fatalf("no usable method => not calculable, never a guess: %+v", r)
	}
}

func TestNormalizeTargetRange(t *testing.T) {
	n := DefaultNormalizer{}
	w := tw("BW", 1)
	w.Direction, w.IdealMin, w.IdealMax, w.TargetTolerance = TargetRange, fp(30), fp(40), fp(10)
	val := func(v float64) *float64 {
		r := n.Normalize(NormalizeInput{Weight: w, Eval: Evaluation{Basis: basis("BW"), Value: v}})
		return r.Score
	}
	near(t, *val(35), 100, "inside")
	near(t, *val(30), 100, "edge")
	near(t, *val(43), 70, "3 above")
	near(t, *val(26), 60, "4 below")
	near(t, *val(55), 0, "far away is floored at 0")
	w.TargetTolerance = nil
	if r := n.Normalize(NormalizeInput{Weight: w, Eval: Evaluation{Basis: basis("BW"), Value: 35}}); r.Score != nil || !strings.Contains(r.Reason, "tolerância") {
		t.Fatalf("a target range without tolerance must be not calculable: %+v", r)
	}
	w.TargetTolerance, w.IdealMin = fp(10), nil
	if n.Normalize(NormalizeInput{Weight: w, Eval: Evaluation{Basis: basis("BW"), Value: 35}}).Score != nil {
		t.Fatal("missing range")
	}
	// trait-level range is used when the profile does not override it
	w.IdealMax = nil
	e := Evaluation{Basis: basis("BW"), Value: 45, IdealMin: fp(30), IdealMax: fp(40)}
	near(t, *n.Normalize(NormalizeInput{Weight: w, Eval: e}).Score, 50, "range from the trait definition")
}

func TestUndefinedDirectionIsNotCalculable(t *testing.T) {
	e := pct("e", "WW", 80, nil)
	e.Direction = InformationOnly
	if r := (DefaultNormalizer{}).Normalize(NormalizeInput{Weight: tw("WW", 1), Eval: e}); r.Score != nil || !strings.Contains(r.Reason, "direção") {
		t.Fatalf("%+v", r)
	}
	w := tw("WW", 1)
	w.Direction = Maximize
	if r := (DefaultNormalizer{}).Normalize(NormalizeInput{Weight: w, Eval: e}); r.Score == nil {
		t.Fatal("a profile direction override must work")
	}
}

// ---- end-to-end scoring with hand-computed numbers

func baseInput() RunInput {
	p := profile(tw("WW", 60), tw("YW", 40))
	p.Complementarity.MaxBonusPoints = 10
	return RunInput{
		Female:     dam(pct("d1", "WW", 40, nil), pct("d2", "YW", 60, nil)),
		Candidates: []Candidate{bull("b1", "Alfa", pct("a1", "WW", 80, nil), pct("a2", "YW", 70, nil)), bull("b2", "Beta", pct("c1", "WW", 50, nil), pct("c2", "YW", 90, nil))},
		Profile:    p,
	}
}

func TestScoreMatchesHandCalculation(t *testing.T) {
	out := run(t, baseInput())
	if len(out.Errors) != 0 {
		t.Fatal(out.Errors)
	}
	a, b := byID(t, out, "b1"), byID(t, out, "b2")
	// Alfa: base = 0.6*80 + 0.4*70 = 76 ; bonus = 10*(0.6*0.4 + 0.4*0.1) = 2.8
	near(t, a.Final.Base, 76, "Alfa base")
	near(t, a.Final.Complementarity, 2.8, "Alfa complementarity")
	near(t, a.Final.Final, 78.8, "Alfa final")
	// Beta: base = 0.6*50 + 0.4*90 = 66 ; bonus = 10*(0.6*0.1 + 0.4*0.3) = 1.8
	near(t, b.Final.Base, 66, "Beta base")
	near(t, b.Final.Complementarity, 1.8, "Beta complementarity")
	near(t, b.Final.Final, 67.8, "Beta final")
	if a.Rank != 1 || b.Rank != 2 || out.Results[0].Bull.ID != "b1" {
		t.Fatalf("ranking: %d %d", a.Rank, b.Rank)
	}
	if len(out.Top) != 2 || out.Top[0] != "b1" {
		t.Fatalf("top: %v", out.Top)
	}
	// components: weights are fractions that add to 1, contributions add to base
	var wsum, csum float64
	for _, c := range a.Components {
		wsum += c.Weight
		csum += *c.Contribution
		if c.Source.DatasetName != "Synthetic" || c.Source.EvaluationID == "" || c.Transformation == "" || c.Formula == "" || c.Explanation == "" {
			t.Fatalf("component without provenance/explanation: %+v", c)
		}
	}
	near(t, wsum, 1, "weights")
	near(t, csum, a.Final.Base, "contributions")
	near(t, a.Components[0].Weight, 0.6, "WW weight")
	near(t, *a.Components[0].Contribution, 48, "WW contribution")
	if a.Components[0].DamValue == nil || *a.Components[0].DamValue != 4 {
		t.Fatalf("dam value should be shown: %+v", a.Components[0].DamValue)
	}
	if !strings.Contains(a.Components[0].Source.Parameter, "peso 60,00") {
		t.Fatalf("the weight must say where it came from: %q", a.Components[0].Source.Parameter)
	}
	if a.Final.Coverage != 1 || a.Final.Clamped {
		t.Fatalf("%+v", a.Final)
	}
}

func TestMinimizeAndTargetRangeIntegration(t *testing.T) {
	scs := pct("s", "SCS", 30, nil) // ascending percentile 30
	scs.Direction = Minimize
	bw := Evaluation{ID: "bw", Basis: basis("BW"), Value: 43, Direction: TargetRange, TraitName: "BW", IdealMin: fp(30), IdealMax: fp(40), DatasetID: "ds1", DatasetName: "Synthetic", DatasetVersion: "1"}
	p := profile(tw("SCS", 50), TraitWeight{TraitCode: "BW", EvaluationType: "DEP", Weight: 50, TargetTolerance: fp(10)})
	out := run(t, RunInput{Female: dam(), Candidates: []Candidate{bull("b1", "Alfa", scs, bw)}, Profile: p})
	r := byID(t, out, "b1")
	// SCS minimize with ascending percentile 30 => 70 ; BW 3 above range, tol 10 => 70
	near(t, r.Final.Base, 70, "base")
	if r.Complementarity.Status != StatusNotApplied {
		t.Fatalf("complementarity defaults to off: %s", r.Complementarity.Status)
	}
}

func TestIncompatibleBasesAreNotConverted(t *testing.T) {
	other := pct("o1", "WW", 99, nil)
	other.Basis.Program = "OTHER-PROGRAM"
	in := RunInput{
		Female: dam(pct("d1", "WW", 40, nil)),
		Candidates: []Candidate{
			bull("b1", "Alfa", pct("a1", "WW", 60, nil)),
			bull("b2", "Beta", pct("c1", "WW", 55, nil)),
			bull("b3", "Gama", other), // evaluated by another program: 99 must NOT beat 60
		},
		Profile: profile(tw("WW", 100)),
	}
	out := run(t, in)
	g := byID(t, out, "b3")
	if g.Status != StatusNotCalculable || g.Final != nil {
		t.Fatalf("a candidate on another basis must be not calculable, got %s", g.Status)
	}
	if !contains(g.Warnings, "WW: "+IncompatibleBasesMessage) || !contains(g.Components[0].Warnings, IncompatibleBasesMessage) {
		t.Fatalf("the unsupported-comparison warning must be shown: %v / %v", g.Warnings, g.Components[0].Warnings)
	}
	if IncompatibleBasesMessage != "⚠ Comparação genética não suportada entre essas bases." {
		t.Fatal("message text changed")
	}
	if byID(t, out, "b1").Rank != 1 || byID(t, out, "b2").Rank != 2 {
		t.Fatal("compatible candidates are still ranked")
	}
	// results order: calculated, then not calculable
	if out.Results[2].Bull.ID != "b3" {
		t.Fatalf("not-calculable candidates come after the ranking: %v", out.Results[2].Bull.ID)
	}
	if len(out.Bases) != 1 || out.Bases[0].Basis.Program != "SYN" || out.Bases[0].Candidates != 2 || len(out.Bases[0].Alternatives) != 2 {
		t.Fatalf("%+v", out.Bases)
	}
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func TestReferenceBasisChoice(t *testing.T) {
	mk := func(id string, rel string) Evaluation {
		e := pct(id, "WW", 50, nil)
		e.Basis.Release = rel
		return e
	}
	// 2 candidates on r2, 1 on r1 => r2 wins by count, even if the dam is on r1
	in := RunInput{Female: dam(mk("d", "r1")), Profile: profile(tw("WW", 1)), Candidates: []Candidate{
		bull("b1", "A", mk("1", "r2")), bull("b2", "B", mk("2", "r2")), bull("b3", "C", mk("3", "r1"))}}
	out := run(t, in)
	if out.Bases[0].Basis.Release != "r2" || out.Bases[0].DamMatches {
		t.Fatalf("%+v", out.Bases[0])
	}
	if !strings.Contains(strings.Join(out.Warnings, " "), IncompatibleBasesMessage) {
		t.Fatalf("the dam on another basis must be flagged: %v", out.Warnings)
	}
	// tie 1–1: the basis the dam is on wins
	in.Candidates = []Candidate{bull("b1", "A", mk("1", "r2")), bull("b3", "C", mk("3", "r1"))}
	out = run(t, in)
	if out.Bases[0].Basis.Release != "r1" || !out.Bases[0].DamMatches || !strings.Contains(out.Bases[0].Reason, "empate") {
		t.Fatalf("%+v", out.Bases[0])
	}
	// tie and the dam has neither: smallest key (deterministic)
	in.Female = dam()
	out = run(t, in)
	if out.Bases[0].Basis.Release != "r1" {
		t.Fatalf("%+v", out.Bases[0])
	}
}

func TestEvaluationTypesAreNeverMixed(t *testing.T) {
	ebv := pct("e", "WW", 90, nil)
	ebv.Basis.EvaluationType = "EBV"
	out := run(t, RunInput{Female: dam(), Candidates: []Candidate{bull("b1", "A", ebv)}, Profile: profile(tw("WW", 100))}) // profile asks for DEP
	r := byID(t, out, "b1")
	if r.Status != StatusNotCalculable || !strings.Contains(r.Components[0].Explanation, "não tem avaliação DEP") {
		t.Fatalf("an EBV must never stand in for a DEP: %+v", r.Components[0])
	}
}

func TestSeveralEvaluationsOnTheSameBasisUseTheNewest(t *testing.T) {
	old := pct("old", "WW", 10, nil)
	old.CreatedAt = "2026-01-01T00:00:00Z"
	neu := pct("new", "WW", 90, nil)
	neu.CreatedAt = "2026-06-01T00:00:00Z"
	out := run(t, RunInput{Female: dam(), Candidates: []Candidate{bull("b1", "A", old, neu)}, Profile: profile(tw("WW", 100))})
	r := byID(t, out, "b1")
	near(t, *r.Components[0].Score, 90, "newest")
	if !strings.Contains(strings.Join(r.Warnings, " "), "2 avaliações na mesma base") {
		t.Fatalf("%v", r.Warnings)
	}
}

// ---- coverage and renormalization

func TestCoverageAndMinCoverage(t *testing.T) {
	p := profile(tw("WW", 50), tw("YW", 30), tw("MILK", 20))
	cands := []Candidate{
		bull("full", "Completo", pct("1", "WW", 80, nil), pct("2", "YW", 60, nil), pct("3", "MILK", 40, nil)),
		bull("part", "Parcial", pct("4", "WW", 80, nil), pct("5", "YW", 60, nil)), // lacks MILK (20% of the weight)
	}
	out := run(t, RunInput{Female: dam(), Candidates: cands, Profile: p})
	part := byID(t, out, "part")
	if part.Status != StatusNotCalculable || part.Final != nil || !strings.Contains(strings.Join(part.Reasons, " "), "80,00%") {
		t.Fatalf("with the default MinCoverage=100%%, a partial candidate has NO score: %+v", part)
	}
	if byID(t, out, "full").Status != StatusCalculated {
		t.Fatal("the full candidate is calculated")
	}

	p.MinCoverage = 0.8
	out = run(t, RunInput{Female: dam(), Candidates: cands, Profile: p})
	part = byID(t, out, "part")
	if part.Status != StatusCalculated {
		t.Fatalf("%+v", part.Reasons)
	}
	// weights are renormalized over the calculable components: 50/80 and 30/80
	near(t, part.Components[0].Weight, 0.625, "WW")
	near(t, part.Components[1].Weight, 0.375, "YW")
	near(t, part.Components[2].Weight, 0, "MILK not used")
	near(t, part.Final.Base, 0.625*80+0.375*60, "base")
	near(t, part.Final.Coverage, 0.8, "coverage is reported")
	if part.Components[2].Status != StatusNotCalculable || part.Components[2].Contribution != nil {
		t.Fatalf("%+v", part.Components[2])
	}
	// the complete candidate has 100% coverage, so the two scores are not equivalent bases — coverage tells the user
	near(t, byID(t, out, "full").Final.Coverage, 1, "full coverage")
}

func TestZeroWeightTraitsAreIgnored(t *testing.T) {
	p := profile(tw("WW", 100), tw("YW", 0))
	out := run(t, RunInput{Female: dam(), Candidates: []Candidate{bull("b1", "A", pct("1", "WW", 70, nil))}, Profile: p})
	r := byID(t, out, "b1")
	if len(r.Components) != 1 || r.Status != StatusCalculated {
		t.Fatalf("a trait with weight 0 must not require data: %+v", r.Components)
	}
}

// ---- complementarity

func TestComplementarityNeedsComparableDam(t *testing.T) {
	p := profile(tw("WW", 100))
	p.Complementarity.MaxBonusPoints = 10
	cand := bull("b1", "A", pct("1", "WW", 80, nil))

	// no dam evaluations at all
	r := byID(t, run(t, RunInput{Female: dam(), Candidates: []Candidate{cand}, Profile: p}), "b1")
	if r.Complementarity.Status != StatusNotCalculable || r.Final.Complementarity != 0 || r.Status != StatusCalculated {
		t.Fatalf("%+v", r.Complementarity)
	}
	// dam on another basis
	other := pct("d", "WW", 10, nil)
	other.Basis.Program = "OTHER"
	r = byID(t, run(t, RunInput{Female: dam(other), Candidates: []Candidate{cand}, Profile: p}), "b1")
	if r.Complementarity.Items[0].Status != StatusNotCalculable || !contains(r.Complementarity.Warnings, "WW: "+IncompatibleBasesMessage) || r.Final.Complementarity != 0 {
		t.Fatalf("%+v", r.Complementarity)
	}
	// dam better than the bull: no bonus, and never a negative one
	r = byID(t, run(t, RunInput{Female: dam(pct("d", "WW", 95, nil)), Candidates: []Candidate{cand}, Profile: p}), "b1")
	near(t, r.Complementarity.Points, 0, "no bonus when the dam is better")
	near(t, *r.Complementarity.Items[0].Gain, 0, "gain clamped at 0")
	// complementarity disabled
	p.Complementarity.MaxBonusPoints = 0
	r = byID(t, run(t, RunInput{Female: dam(pct("d", "WW", 10, nil)), Candidates: []Candidate{cand}, Profile: p}), "b1")
	if r.Complementarity.Status != StatusNotApplied || r.Complementarity.Points != 0 {
		t.Fatalf("%+v", r.Complementarity)
	}
}

func TestComplementarityOnlyCountsTraitsWhereTheDamIsKnown(t *testing.T) {
	p := profile(tw("WW", 50), tw("YW", 50))
	p.Complementarity.MaxBonusPoints = 10
	cand := bull("b1", "A", pct("1", "WW", 100, nil), pct("2", "YW", 100, nil))
	// dam known only for WW (score 0): gain 1 on WW (share .5) and unknown on YW => 10*0.5*1 = 5
	r := byID(t, run(t, RunInput{Female: dam(pct("d", "WW", 0, nil)), Candidates: []Candidate{cand}, Profile: p}), "b1")
	near(t, r.Complementarity.Points, 5, "only the demonstrated half is rewarded")
	if r.Complementarity.Items[1].Status != StatusNotCalculable {
		t.Fatalf("%+v", r.Complementarity.Items[1])
	}
}

// ---- penalties

func TestLowAccuracyPenalty(t *testing.T) {
	p := profile(tw("WW", 60), tw("YW", 40))
	p.Penalties.LowAccuracy = LowAccuracyPenalty{Threshold: 0.6, MaxPoints: 10}
	cand := bull("b1", "A", pct("1", "WW", 80, fp(0.3)), pct("2", "YW", 80, fp(0.9)))
	r := byID(t, run(t, RunInput{Female: dam(), Candidates: []Candidate{cand}, Profile: p}), "b1")
	// WW: shortfall (0.6-0.3)/0.6 = 0.5, share .6 => 10*.6*.5 = 3 ; YW: none
	near(t, r.Penalties[0].Points, 3, "penalty")
	near(t, r.Final.Penalties, 3, "total penalties")
	near(t, r.Final.Final, 80-3, "final")
	if r.Penalties[0].Status != StatusCalculated || !strings.Contains(r.Penalties[0].Source.Parameter, "limiar 0,60") {
		t.Fatalf("%+v", r.Penalties[0])
	}
	// unknown accuracy is not "low"
	cand = bull("b1", "A", pct("1", "WW", 80, nil), pct("2", "YW", 80, nil))
	r = byID(t, run(t, RunInput{Female: dam(), Candidates: []Candidate{cand}, Profile: p}), "b1")
	near(t, r.Penalties[0].Points, 0, "no penalty for unknown accuracy")
	if len(r.Penalties[0].Warnings) == 0 {
		t.Fatal("unknown accuracy must be reported")
	}
}

func TestInbreedingIsNeverInvented(t *testing.T) {
	r := byID(t, run(t, baseInput()), "b1")
	var inb *PenaltyResult
	for i := range r.Penalties {
		if strings.Contains(r.Penalties[i].Name, "endogamia") {
			inb = &r.Penalties[i]
		}
	}
	if inb == nil || inb.Status != StatusNotApplied || inb.Points != 0 || !strings.Contains(inb.Explanation, "Planejado — ainda não utilizado no score do MVP") {
		t.Fatalf("%+v", inb)
	}
}

func TestClamping(t *testing.T) {
	p := profile(tw("WW", 100))
	p.Complementarity.MaxBonusPoints = 10
	r := byID(t, run(t, RunInput{Female: dam(pct("d", "WW", 0, nil)), Candidates: []Candidate{bull("b", "A", pct("1", "WW", 99.5, nil))}, Profile: p}), "b")
	// gain = 0.995 => bonus 9.95 ; base 99.5 ; unclamped 109.45
	near(t, r.Final.Unclamped, 99.5+9.95, "unclamped is recorded")
	near(t, r.Final.Final, 100, "capped at 100")
	if !r.Final.Clamped {
		t.Fatal("clamping must be visible")
	}
	p = profile(tw("WW", 100))
	p.Penalties.LowAccuracy = LowAccuracyPenalty{Threshold: 1, MaxPoints: 50}
	r = byID(t, run(t, RunInput{Female: dam(), Candidates: []Candidate{bull("b", "A", pct("1", "WW", 10, fp(0)))}, Profile: p}), "b")
	near(t, r.Final.Unclamped, 10-50, "unclamped")
	near(t, r.Final.Final, 0, "floored at 0")
}

// ---- hard filters

func TestHardFilters(t *testing.T) {
	p := profile(tw("WW", 100))
	p.Filters = Filters{RequireStock: true, MinDoses: 5, MaxUnitCost: fp(60), MinAccuracy: fp(0.5)}
	ok := bull("ok", "Ok", pct("1", "WW", 70, fp(0.8)))
	noStock := bull("stock", "SemEstoque", pct("2", "WW", 90, fp(0.8)))
	noStock.Stock.Doses = 2
	costly := bull("cost", "Caro", pct("3", "WW", 90, fp(0.8)))
	costly.Stock.UnitCost = fp(80)
	noCost := bull("nocost", "SemCusto", pct("4", "WW", 90, fp(0.8)))
	noCost.Stock.UnitCost = nil
	lowAcc := bull("low", "Baixa", pct("5", "WW", 90, fp(0.2)))
	unkAcc := bull("unk", "AcuraciaDesconhecida", pct("6", "WW", 90, nil))
	archived := bull("arch", "Arquivado", pct("7", "WW", 90, fp(0.8)))
	archived.Status = "archived"

	out := run(t, RunInput{Female: dam(), Candidates: []Candidate{ok, noStock, costly, noCost, lowAcc, unkAcc, archived}, Profile: p})
	for id, filter := range map[string]string{"stock": "estoque", "cost": "custo", "nocost": "custo", "low": "acurácia mínima", "unk": "acurácia mínima", "arch": "status"} {
		r := byID(t, out, id)
		if r.Status != StatusExcluded || len(r.Exclusions) == 0 || r.Exclusions[0].Filter != filter {
			t.Fatalf("%s: %+v", id, r)
		}
	}
	if byID(t, out, "ok").Status != StatusCalculated || out.Results[0].Bull.ID != "ok" {
		t.Fatal("only the compliant candidate survives")
	}
	// excluded candidates come last and never receive a rank
	if byID(t, out, "stock").Rank != 0 || byID(t, out, "stock").Final != nil {
		t.Fatal("excluded candidates have no rank or score")
	}
	// excluded candidates do not influence the reference basis
	if out.Bases[0].Eligible != 1 {
		t.Fatalf("eligible: %d", out.Bases[0].Eligible)
	}
	// every failing filter is reported, not just the first
	both := bull("both", "Ambos", pct("8", "WW", 50, fp(0.1)))
	both.Stock.Doses = 0
	out = run(t, RunInput{Female: dam(), Candidates: []Candidate{both}, Profile: p})
	if got := byID(t, out, "both").Exclusions; len(got) < 2 {
		t.Fatalf("all reasons must be listed: %+v", got)
	}
}

func TestRequiredTraitFilter(t *testing.T) {
	p := profile(tw("WW", 50), TraitWeight{TraitCode: "YW", EvaluationType: "DEP", Weight: 50, Required: true})
	p.MinCoverage = 0.5
	out := run(t, RunInput{Female: dam(), Candidates: []Candidate{bull("a", "A", pct("1", "WW", 50, nil)), bull("b", "B", pct("2", "WW", 50, nil), pct("3", "YW", 50, nil))}, Profile: p})
	if byID(t, out, "a").Status != StatusExcluded || byID(t, out, "b").Status != StatusCalculated {
		t.Fatal("a required trait without any evaluation excludes the candidate")
	}
}

func pedigree() genetics.Graph {
	g := genetics.Graph{}
	add := func(id, name, sex, f, m string) {
		g[id] = &genetics.Node{ID: id, Name: name, Sex: sex, FatherID: f, MotherID: m}
	}
	add("dam", "Matriz", "female", "grand", "gdam")
	add("grand", "Avô", "male", "", "")
	add("gdam", "Avó", "female", "", "")
	add("b1", "Alfa", "male", "grand", "") // son of the dam's father: half brother of the dam
	add("b2", "Beta", "male", "", "")
	add("b3", "Avô", "male", "", "") // placeholder, replaced below
	delete(g, "b3")
	return g
}

func TestAncestorFilterAndPedigreeWarnings(t *testing.T) {
	p := profile(tw("WW", 100))
	cands := []Candidate{bull("grand", "Avô", pct("1", "WW", 90, nil)), bull("b1", "Alfa", pct("2", "WW", 80, nil)), bull("b2", "Beta", pct("3", "WW", 70, nil))}
	in := RunInput{Female: dam(), Candidates: cands, Profile: p, Pedigree: pedigree()}

	// default: warn only
	out := run(t, in)
	a := byID(t, out, "b1")
	if a.Status != StatusCalculated || len(a.PedigreeNotes) == 0 || !strings.Contains(a.PedigreeNotes[0], "Ancestral comum: Avô") {
		t.Fatalf("common ancestors are reported: %+v", a.PedigreeNotes)
	}
	if !strings.Contains(strings.Join(a.Warnings, " "), "coeficiente de endogamia não calculado nesta versão") {
		t.Fatalf("%v", a.Warnings)
	}
	if len(byID(t, out, "b2").PedigreeNotes) != 0 {
		t.Fatal("unrelated bull has no notes")
	}
	if byID(t, out, "grand").Status != StatusCalculated {
		t.Fatal("not excluded unless the profile asks")
	}
	// the profile excludes direct ancestors of the dam
	in.Profile.Filters.ExcludeDirectAncestors = true
	out = run(t, in)
	g := byID(t, out, "grand")
	if g.Status != StatusExcluded || g.Exclusions[0].Filter != "ancestral direto" {
		t.Fatalf("%+v", g)
	}
	if byID(t, out, "b1").Status != StatusCalculated {
		t.Fatal("a half-brother is not a direct ancestor")
	}
}

// ---- validation

func TestRunValidation(t *testing.T) {
	good := baseInput()
	bad := func(mod func(*RunInput), want string) {
		t.Helper()
		in := baseInput()
		mod(&in)
		out := run(t, in)
		if len(out.Errors) == 0 || !strings.Contains(strings.Join(out.Errors, " "), want) || len(out.Results) != 0 {
			t.Fatalf("want error %q, got %v (results %d)", want, out.Errors, len(out.Results))
		}
	}
	if out := run(t, good); len(out.Errors) != 0 {
		t.Fatal(out.Errors)
	}
	bad(func(in *RunInput) { in.Female.Sex = "male" }, "fêmea")
	bad(func(in *RunInput) { in.Female.ID = "" }, "matriz")
	bad(func(in *RunInput) { in.Candidates = nil }, "ao menos um touro")
	bad(func(in *RunInput) { in.Candidates = append(in.Candidates, in.Candidates[0]) }, "mais de uma vez")
	bad(func(in *RunInput) { in.Profile.Traits = nil }, "nenhuma característica")
	bad(func(in *RunInput) { in.Profile.Traits[0].Weight = 0; in.Profile.Traits[1].Weight = 0 }, "nenhuma característica")
	bad(func(in *RunInput) { in.Profile.Traits[0].Weight = -1 }, "Peso inválido")
	bad(func(in *RunInput) { in.Profile.Traits[0].Weight = math.NaN() }, "Peso inválido")
	bad(func(in *RunInput) { in.Profile.Traits[1] = in.Profile.Traits[0] }, "mais de uma vez")
	bad(func(in *RunInput) { in.Profile.Traits[0].Direction = "BEST" }, "Direção inválida")
	bad(func(in *RunInput) { in.Profile.Normalization = "magic" }, "normalização")
	bad(func(in *RunInput) { in.Profile.Complementarity.MaxBonusPoints = 500 }, "bônus máximo")
	bad(func(in *RunInput) { in.Profile.Filters.MinAccuracy = fp(2) }, "acurácia mínima")
}

func TestCandidateValidation(t *testing.T) {
	in := baseInput()
	cow := bull("cow", "Vaca", pct("9", "WW", 99, nil))
	cow.Sex = "female"
	self := bull("dam", "Ela mesma", pct("8", "WW", 99, nil))
	in.Candidates = append(in.Candidates, cow, self)
	out := run(t, in)
	for _, id := range []string{"cow", "dam"} {
		if r := byID(t, out, id); r.Status != StatusExcluded || r.Exclusions[0].Filter != "validação" {
			t.Fatalf("%s: %+v", id, r)
		}
	}
	if byID(t, out, "b1").Status != StatusCalculated {
		t.Fatal("valid candidates are unaffected")
	}
}

// ---- ranking, determinism, snapshot

func TestTiesAndDeterministicOrdering(t *testing.T) {
	p := profile(tw("WW", 100))
	mk := func(id, name string) Candidate { return bull(id, name, pct("e"+id, "WW", 70, nil)) }
	cands := []Candidate{mk("3", "Zeta"), mk("1", "alfa"), mk("2", "Beta"), bull("4", "Topo", pct("e4", "WW", 90, nil))}
	out := run(t, RunInput{Female: dam(), Candidates: cands, Profile: p})
	var names []string
	for _, r := range out.Results {
		names = append(names, r.Bull.Name)
	}
	if strings.Join(names, ",") != "Topo,alfa,Beta,Zeta" {
		t.Fatalf("ties are broken by name: %v", names)
	}
	if out.Results[0].Tied || !out.Results[1].Tied || !out.Results[3].Tied {
		t.Fatalf("tied flags: %v %v %v", out.Results[0].Tied, out.Results[1].Tied, out.Results[3].Tied)
	}
	// shuffling the input must not change anything
	rng := rand.New(rand.NewSource(1))
	first, _ := json.Marshal(out)
	for i := 0; i < 20; i++ {
		rng.Shuffle(len(cands), func(a, b int) { cands[a], cands[b] = cands[b], cands[a] })
		again, _ := json.Marshal(run(t, RunInput{Female: dam(), Candidates: cands, Profile: p}))
		if string(first) != string(again) {
			t.Fatal("the result depends on input order")
		}
	}
}

func TestRunIsReproducible(t *testing.T) {
	a, _ := json.Marshal(run(t, baseInput()))
	b, _ := json.Marshal(run(t, baseInput()))
	if string(a) != string(b) {
		t.Fatal("two runs on the same input must be byte-identical")
	}
	out := run(t, baseInput())
	if out.EngineVersion != Version || out.CreatedAt != "2026-10-11T12:00:00Z" {
		t.Fatalf("%s %s", out.EngineVersion, out.CreatedAt)
	}
}

func TestTopNIsLimited(t *testing.T) {
	var cands []Candidate
	for i := 0; i < 9; i++ {
		cands = append(cands, bull(string(rune('a'+i)), string(rune('A'+i)), pct("e"+string(rune('a'+i)), "WW", float64(10+i*5), nil)))
	}
	out := run(t, RunInput{Female: dam(), Candidates: cands, Profile: profile(tw("WW", 1))})
	if len(out.Top) != TopN || out.Top[0] != "i" {
		t.Fatalf("%v", out.Top)
	}
}

func TestSnapshotIsSelfContained(t *testing.T) {
	in := baseInput()
	in.Datasets = []DatasetInfo{{ID: "ds1", PackageID: "synthetic", Name: "Synthetic", Version: "1", Program: "SYN", Release: "r1", SHA256: "abc", IsDemo: true}, {ID: "unused", Name: "Unused", Version: "9"}}
	in.Profile.IsDemo = true
	in.Profile.DemoNotice = DemoNotice
	extra := bull("x", "Sem estoque", pct("9", "WW", 99, nil))
	extra.Stock.Doses = 0
	in.Candidates = append(in.Candidates, extra)
	out := run(t, in)
	snaps := Snapshots(in, out)
	if _, ok := snaps["x"]; ok {
		t.Fatal("excluded candidates have no pair snapshot")
	}
	s := snaps["b1"]
	if s.SnapshotVersion != SnapshotVersion || s.EngineVersion != Version || s.CreatedAt == "" || s.Female.ID != "dam" || s.Bull.ID != "b1" {
		t.Fatalf("%+v", s)
	}
	if len(s.Datasets) != 1 || s.Datasets[0].ID != "ds1" || s.Datasets[0].Version != "1" || s.Datasets[0].SHA256 != "abc" {
		t.Fatalf("only the datasets actually used are recorded: %+v", s.Datasets)
	}
	if len(s.BullEvals) != 2 || len(s.FemaleEvals) != 2 || s.Profile.Name != "Perfil de teste" || len(s.Bases) != 2 || s.Candidates != 3 {
		t.Fatalf("%d %d %d", len(s.BullEvals), len(s.FemaleEvals), len(s.Bases))
	}
	if !s.IsDemo {
		t.Fatal("DEMO must propagate to the snapshot")
	}
	// round trip: the saved JSON reproduces the exact result numbers
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var back PairSnapshot
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	near(t, back.Result.Final.Final, s.Result.Final.Final, "final after round trip")
	near(t, *back.Result.Components[0].Contribution, *s.Result.Components[0].Contribution, "contribution after round trip")
	if back.Result.Components[0].Source.DatasetVersion != "1" {
		t.Fatal("dataset version must be in the snapshot")
	}
}

// ---- invariants over random synthetic inputs

func TestInvariantsOnRandomInputs(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	checked, withBonus, withPenalty := 0, 0, 0
	traits := []string{"WW", "YW", "MILK", "SC"}
	for iter := 0; iter < 300; iter++ {
		var tws []TraitWeight
		for _, tr := range traits {
			if rng.Intn(4) > 0 {
				tws = append(tws, tw(tr, float64(rng.Intn(100))))
			}
		}
		p := profile(tws...)
		p.MinCoverage = []float64{1, 0.5, 0.2}[rng.Intn(3)]
		p.Complementarity.MaxBonusPoints = float64(rng.Intn(21))
		if rng.Intn(2) == 0 {
			p.Penalties.LowAccuracy = LowAccuracyPenalty{Threshold: rng.Float64(), MaxPoints: float64(rng.Intn(21))}
		}
		mk := func(id string) []Evaluation {
			var evs []Evaluation
			for _, tr := range traits {
				if rng.Intn(5) == 0 {
					continue
				}
				var acc *float64
				if rng.Intn(3) > 0 {
					acc = fp(rng.Float64())
				}
				evs = append(evs, pct(id+tr, tr, rng.Float64()*100, acc))
			}
			return evs
		}
		var cands []Candidate
		for i := 0; i < 1+rng.Intn(6); i++ {
			id := string(rune('a' + i))
			cands = append(cands, bull(id, "B"+id, mk(id)...))
		}
		out := New().Run(RunInput{Female: dam(mk("d")...), Candidates: cands, Profile: p, Now: t0})
		if len(out.Errors) > 0 {
			continue // e.g. every weight drawn as 0
		}
		prev := math.Inf(1)
		for _, r := range out.Results {
			if r.Status != StatusCalculated {
				if r.Final != nil || r.Rank != 0 {
					t.Fatalf("iter %d: a non-calculated pair must have no score/rank", iter)
				}
				continue
			}
			f := r.Final
			checked++
			if f.Complementarity > 0 {
				withBonus++
			}
			if f.Penalties > 0 {
				withPenalty++
			}
			var wsum, csum float64
			for _, c := range r.Components {
				if c.Status == StatusCalculated {
					wsum += c.Weight
					csum += *c.Contribution
					if *c.Score < -1e-9 || *c.Score > 100+1e-9 {
						t.Fatalf("iter %d: score out of range: %v", iter, *c.Score)
					}
				} else if c.Weight != 0 || c.Contribution != nil {
					t.Fatalf("iter %d: uncalculated component carries weight", iter)
				}
			}
			if math.Abs(wsum-1) > 1e-9 || math.Abs(csum-f.Base) > 1e-9 {
				t.Fatalf("iter %d: weights %v / contributions %v vs base %v", iter, wsum, csum, f.Base)
			}
			if math.Abs(f.Unclamped-(f.Base+f.Complementarity-f.Penalties)) > 1e-9 {
				t.Fatalf("iter %d: unclamped formula broken", iter)
			}
			if f.Final < 0 || f.Final > 100 || f.Complementarity < -1e-12 || f.Penalties < -1e-12 {
				t.Fatalf("iter %d: %+v", iter, f)
			}
			if f.Coverage+1e-9 < p.MinCoverage {
				t.Fatalf("iter %d: coverage %v below the minimum %v yet calculated", iter, f.Coverage, p.MinCoverage)
			}
			if f.Final > prev+1e-9 {
				t.Fatalf("iter %d: ranking not sorted by score", iter)
			}
			prev = f.Final
		}
	}
	if checked < 300 || withBonus < 50 || withPenalty < 50 {
		t.Fatalf("the random scenarios must exercise the engine: checked=%d bonus=%d penalty=%d", checked, withBonus, withPenalty)
	}
}
