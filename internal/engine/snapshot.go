package engine

import "sort"

// Snapshots builds the self-contained snapshot for every pair of a run that went
// through the pipeline (calculated or not calculable). Candidates removed by a
// hard filter have no pair snapshot; they are listed in the run result.
func Snapshots(in RunInput, out RunOutput) map[string]PairSnapshot {
	res := map[string]PairSnapshot{}
	active := map[string]bool{}
	for _, t := range out.Profile.ActiveTraits() {
		active[t.Key()] = true
	}
	weighted := func(evals []Evaluation) []Evaluation {
		var list []Evaluation
		for _, e := range evals {
			if active[e.Basis.TraitCode+"|"+e.Basis.EvaluationType] {
				list = append(list, e)
			}
		}
		sort.SliceStable(list, func(i, j int) bool { return list[i].ID < list[j].ID })
		if list == nil {
			list = []Evaluation{}
		}
		return list
	}
	cands := map[string]Candidate{}
	for _, c := range in.Candidates {
		cands[c.ID] = c
	}
	for _, r := range out.Results {
		if r.Status == StatusExcluded {
			continue
		}
		c := cands[r.Bull.ID]
		snap := PairSnapshot{
			SnapshotVersion: SnapshotVersion, EngineVersion: out.EngineVersion, CreatedAt: out.CreatedAt,
			Female: out.Female, FemaleEvals: weighted(in.Female.Evals), Bull: r.Bull, BullEvals: weighted(c.Evals),
			Profile: out.Profile, Bases: out.Bases, Result: r, Candidates: len(in.Candidates), RunWarnings: out.Warnings,
			Datasets: pairDatasets(in, r),
		}
		snap.IsDemo = out.Profile.IsDemo || out.Female.IsDemo || r.Bull.IsDemo
		for _, d := range snap.Datasets {
			snap.IsDemo = snap.IsDemo || d.IsDemo
		}
		for _, e := range snap.BullEvals {
			snap.IsDemo = snap.IsDemo || e.IsDemo
		}
		for _, e := range snap.FemaleEvals {
			snap.IsDemo = snap.IsDemo || e.IsDemo
		}
		res[r.Bull.ID] = snap
	}
	return res
}

func pairDatasets(in RunInput, r CandidateResult) []DatasetInfo {
	byID := map[string]DatasetInfo{}
	for _, d := range in.Datasets {
		byID[d.ID] = d
	}
	seen := map[string]bool{}
	res := []DatasetInfo{}
	add := func(id string) {
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		if d, ok := byID[id]; ok {
			res = append(res, d)
		}
	}
	for _, c := range r.Components {
		add(c.Source.DatasetID)
	}
	for _, it := range r.Complementarity.Items {
		add(it.Source.DatasetID)
	}
	sort.Slice(res, func(i, j int) bool { return res[i].ID < res[j].ID })
	return res
}
