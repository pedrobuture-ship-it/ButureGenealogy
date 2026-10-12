package engine

import (
	"fmt"
	"sort"
)

// DefaultBasisResolver chooses, per weighted trait, the reference basis on which
// candidates are compared. Candidates are only ever compared on one basis per
// trait. Choice rule (deterministic): most eligible candidates evaluated on it;
// then a basis the dam also has; then the lexicographically smallest key.
type DefaultBasisResolver struct{}

func (DefaultBasisResolver) Resolve(run RunContext, eligible []Candidate) map[string]*ReferenceBasis {
	out := map[string]*ReferenceBasis{}
	for _, t := range run.Profile.ActiveTraits() {
		rb := &ReferenceBasis{TraitCode: t.TraitCode, EvaluationType: t.EvaluationType, Eligible: len(eligible), Alternatives: []BasisCount{}}
		counts := map[string]int{}
		bases := map[string]Basis{}
		for _, c := range eligible {
			seen := map[string]bool{}
			for _, e := range c.Evals {
				if e.Basis.TraitCode != t.TraitCode || e.Basis.EvaluationType != t.EvaluationType {
					continue
				}
				k := e.Basis.Key()
				if !seen[k] {
					seen[k] = true
					counts[k]++
					bases[k] = e.Basis
				}
			}
		}
		damHas := map[string]bool{}
		for _, e := range run.Female.Evals {
			if e.Basis.TraitCode == t.TraitCode && e.Basis.EvaluationType == t.EvaluationType {
				damHas[e.Basis.Key()] = true
			}
		}
		keys := make([]string, 0, len(counts))
		for k := range counts {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			a, b := keys[i], keys[j]
			if counts[a] != counts[b] {
				return counts[a] > counts[b]
			}
			if damHas[a] != damHas[b] {
				return damHas[a]
			}
			return a < b
		})
		for _, k := range keys {
			rb.Alternatives = append(rb.Alternatives, BasisCount{Basis: bases[k], Candidates: counts[k]})
		}
		if len(keys) == 0 {
			rb.Reason = "Nenhum candidato elegível tem avaliação desta característica."
		} else {
			b := bases[keys[0]]
			rb.Basis, rb.Candidates, rb.DamMatches = &b, counts[keys[0]], damHas[keys[0]]
			rb.Reason = fmt.Sprintf("Base com mais candidatos avaliados (%d de %d)", rb.Candidates, rb.Eligible)
			if len(keys) > 1 && counts[keys[0]] == counts[keys[1]] {
				if damHas[keys[0]] && !damHas[keys[1]] {
					rb.Reason += "; empate resolvido a favor da base em que a matriz também foi avaliada"
				} else {
					rb.Reason += "; empate resolvido pela ordem alfabética da chave da base"
				}
			}
			if !rb.DamMatches {
				rb.Reason += "; a matriz não tem avaliação nesta base"
			}
		}
		out[t.Key()] = rb
	}
	return out
}

// DefaultEvaluationSelector picks the candidate's evaluation on the reference basis.
type DefaultEvaluationSelector struct{}

func (DefaultEvaluationSelector) Select(run RunContext, c Candidate) (map[string]*Evaluation, map[string][]Evaluation, []string) {
	sel := map[string]*Evaluation{}
	incompatible := map[string][]Evaluation{}
	var warnings []string
	for _, t := range run.Profile.ActiveTraits() {
		rb := run.Bases[t.Key()]
		var same []Evaluation
		for _, e := range c.Evals {
			if e.Basis.TraitCode != t.TraitCode || e.Basis.EvaluationType != t.EvaluationType {
				continue
			}
			if rb != nil && rb.Basis != nil && e.Basis.Key() == rb.Basis.Key() {
				same = append(same, e)
			} else {
				incompatible[t.Key()] = append(incompatible[t.Key()], e)
			}
		}
		if len(same) == 0 {
			continue
		}
		// several evaluations on the same basis (e.g. two installed versions of one dataset):
		// use the most recently installed, deterministically, and say so.
		sort.SliceStable(same, func(i, j int) bool {
			if same[i].CreatedAt != same[j].CreatedAt {
				return same[i].CreatedAt > same[j].CreatedAt
			}
			return same[i].ID < same[j].ID
		})
		chosen := same[0]
		sel[t.Key()] = &chosen
		if len(same) > 1 {
			warnings = append(warnings, fmt.Sprintf("%s: %d avaliações na mesma base; usada a mais recentemente instalada (%s %s).",
				t.TraitCode, len(same), chosen.DatasetName, chosen.DatasetVersion))
		}
	}
	return sel, incompatible, warnings
}
