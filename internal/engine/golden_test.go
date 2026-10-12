package engine

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the golden file of the current engine version")

// goldenInput exercises every stage: percentile and z-score normalization, MINIMIZE, TARGET_RANGE,
// complementarity, the low-accuracy penalty, the (not applied) inbreeding rule, an incompatible
// basis, a coverage failure and a hard-filter exclusion. All data is synthetic.
func goldenInput() RunInput {
	z := func(id, trait string, value, mean, sd float64) Evaluation {
		return Evaluation{ID: id, Basis: basis(trait), Value: value, DatasetID: "ds1", DatasetName: "Synthetic", DatasetVersion: "1", TraitName: trait,
			Direction: Maximize, Meta: TraitMeta{PopulationMean: fp(mean), PopulationSD: fp(sd)}, CreatedAt: "2026-01-01T00:00:00Z"}
	}
	scs := func(id string, p float64) Evaluation {
		e := pct(id, "SCS", p, fp(0.7))
		e.Direction = Minimize
		return e
	}
	bw := func(id string, v float64) Evaluation {
		return Evaluation{ID: id, Basis: basis("BW"), Value: v, DatasetID: "ds1", DatasetName: "Synthetic", DatasetVersion: "1", TraitName: "BW",
			Direction: TargetRange, IdealMin: fp(30), IdealMax: fp(40), CreatedAt: "2026-01-01T00:00:00Z"}
	}
	otherProgram := pct("g1", "WW", 99, fp(0.9))
	otherProgram.Basis.Program = "OTHER"
	p := profile(tw("WW", 40), tw("YW", 25), tw("SCS", 15), TraitWeight{TraitCode: "BW", EvaluationType: "DEP", Weight: 20, TargetTolerance: fp(10)})
	p.Name, p.IsDemo, p.DemoNotice = "Golden", true, DemoNotice
	p.Complementarity.MaxBonusPoints = 10
	p.Penalties.LowAccuracy = LowAccuracyPenalty{Threshold: 0.6, MaxPoints: 5}
	broke := bull("d", "Delta", pct("d1", "WW", 70, nil), z("d2", "YW", 12, 10, 2))
	broke.Stock.Doses = 0
	return RunInput{
		Female: dam(pct("m1", "WW", 40, fp(0.5)), z("m2", "YW", 9, 10, 2), scs("m3", 60), bw("m4", 36)),
		Candidates: []Candidate{
			bull("a", "Alfa", pct("a1", "WW", 80, fp(0.8)), z("a2", "YW", 12, 10, 2), scs("a3", 30), bw("a4", 43)),
			bull("b", "Beta", pct("b1", "WW", 55, fp(0.4)), z("b2", "YW", 10, 10, 2), scs("b3", 70), bw("b4", 35)),
			bull("c", "Gama", otherProgram, z("c2", "YW", 11, 10, 2), scs("c3", 40), bw("c4", 38)),
			bull("e", "Epsilon", pct("e1", "WW", 60, nil)), // lacks 3 of 4 traits => not calculable
			broke,
		},
		Profile:  p,
		Datasets: []DatasetInfo{{ID: "ds1", PackageID: "synthetic", Name: "Synthetic", Version: "1", Program: "SYN", Release: "r1", SHA256: "0", IsDemo: true}},
		Now:      t0,
	}
}

// TestGoldenRun pins the exact output of the engine for a rich synthetic scenario.
// If this test fails you changed the engine's behaviour: bump engine.Version, then
// regenerate with `go test ./internal/engine -run Golden -update`. A new version gets
// its own golden file, so history stays comparable.
func TestGoldenRun(t *testing.T) {
	in := goldenInput()
	out := New().Run(in)
	got, err := json.MarshalIndent(struct {
		Output    RunOutput               `json:"output"`
		Snapshots map[string]PairSnapshot `json:"snapshots"`
	}{out, Snapshots(in, out)}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join("testdata", "golden_run_v"+Version+".json")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, append(got, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("no golden file for engine version %s (%v): run `go test ./internal/engine -run Golden -update`", Version, err)
	}
	if string(append(got, '\n')) != string(want) {
		t.Fatalf("the engine output changed but engine.Version is still %s.\n"+
			"If the change is intended: bump engine.Version and regenerate the golden file (-update).\nIf not: you introduced a regression.", Version)
	}
}

// The scenario must actually cover the interesting paths, otherwise the golden file protects nothing.
func TestGoldenScenarioCoverage(t *testing.T) {
	out := New().Run(goldenInput())
	statuses := map[string]int{}
	methods := map[string]bool{}
	var bonus, penalty, incompatible bool
	for _, r := range out.Results {
		statuses[r.Status]++
		for _, c := range r.Components {
			methods[c.Method] = true
			for _, w := range c.Warnings {
				incompatible = incompatible || w == IncompatibleBasesMessage
			}
		}
		if r.Final != nil {
			bonus = bonus || r.Final.Complementarity > 0
			penalty = penalty || r.Final.Penalties > 0
		}
	}
	if statuses[StatusCalculated] < 2 || statuses[StatusNotCalculable] < 2 || statuses[StatusExcluded] != 1 {
		t.Fatalf("%v", statuses)
	}
	if !methods["percentile"] || !methods["zscore"] || !methods["target_range"] || !bonus || !penalty || !incompatible {
		t.Fatalf("methods=%v bonus=%v penalty=%v incompatible=%v", methods, bonus, penalty, incompatible)
	}
}
