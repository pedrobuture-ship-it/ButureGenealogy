package engine

import (
	"fmt"
	"math"
)

// DefaultComplementarity awards a bonus for traits where the bull improves on
// the dam, using only evaluations on the same basis:
//
//	gain_t   = clamp((score_bull − score_dam)/100, 0, 1)
//	points   = MaxBonus × Σ_t share_t × gain_t
//
// share_t is the component's normalized weight over the calculable components.
// Traits where the dam has no comparable evaluation contribute 0 (a bonus is
// only granted where it is demonstrated). The bonus is never negative.
type DefaultComplementarity struct{ Normalizer Normalizer }

const complementarityFormula = "pontos = bônus_máximo × Σ(peso_normalizado × ganho) ; ganho = limitar((score_touro − score_matriz)/100, 0, 1)"

func traitWeightFor(p Profile, key string) (TraitWeight, bool) {
	for _, t := range p.Traits {
		if t.Key() == key {
			return t, true
		}
	}
	return TraitWeight{}, false
}

func calculated(comps []ComponentResult) ([]ComponentResult, float64) {
	var out []ComponentResult
	var raw float64
	for _, c := range comps {
		if c.Status == StatusCalculated {
			out = append(out, c)
			raw += c.RawWeight
		}
	}
	return out, raw
}

func (dc DefaultComplementarity) Calculate(ctx CalculationContext, comps []ComponentResult) ComplementarityResult {
	cfg := ctx.Run.Profile.Complementarity
	res := ComplementarityResult{MaxBonus: cfg.MaxBonusPoints, Items: []ComplementarityItem{}, Formula: complementarityFormula, Warnings: []string{},
		Source: Source{Parameter: fmt.Sprintf("perfil %q: bônus máximo de complementaridade = %s pontos", ctx.Run.Profile.Name, f2(cfg.MaxBonusPoints))}}
	if cfg.MaxBonusPoints <= 0 {
		res.Status = StatusNotApplied
		res.Explanation = "Complementaridade desativada no perfil (bônus máximo = 0)."
		return res
	}
	calc, rawSum := calculated(comps)
	if len(calc) == 0 || rawSum <= 0 {
		res.Status = StatusNotCalculable
		res.Explanation = "Nenhum componente calculável para comparar com a matriz."
		return res
	}
	if len(ctx.Run.Female.Evals) == 0 {
		res.Status = StatusNotCalculable
		res.Explanation = "A matriz não tem avaliações genéticas; a complementaridade não pode ser calculada."
		res.Warnings = append(res.Warnings, "Matriz sem avaliações genéticas.")
		return res
	}

	any := false
	for _, c := range calc {
		tw, _ := traitWeightFor(ctx.Run.Profile, c.TraitCode+"|"+c.EvaluationType)
		item := ComplementarityItem{TraitCode: c.TraitCode, Name: c.Name, Weight: c.RawWeight / rawSum, BullScore: c.Score}

		var dam *Evaluation
		var otherBases []Evaluation
		for i := range ctx.Run.Female.Evals {
			e := ctx.Run.Female.Evals[i]
			if e.Basis.TraitCode != c.TraitCode || e.Basis.EvaluationType != c.EvaluationType {
				continue
			}
			if c.Basis != nil && e.Basis.Key() == c.Basis.Key() {
				if dam == nil || e.CreatedAt > dam.CreatedAt {
					cp := e
					dam = &cp
				}
			} else {
				otherBases = append(otherBases, e)
			}
		}
		switch {
		case dam == nil && len(otherBases) > 0:
			item.Status = StatusNotCalculable
			item.Explanation = "A matriz foi avaliada em outra base: " + IncompatibleBasesMessage
			res.Warnings = append(res.Warnings, c.TraitCode+": "+IncompatibleBasesMessage)
		case dam == nil:
			item.Status = StatusNotCalculable
			item.Explanation = "A matriz não tem avaliação desta característica; contribui 0."
		default:
			n := dc.Normalizer.Normalize(NormalizeInput{Weight: tw, Eval: *dam, Mode: ctx.Run.Profile.Normalization})
			item.Source = sourceOf(dam)
			if n.Score == nil {
				item.Status = StatusNotCalculable
				item.Explanation = "A avaliação da matriz não pôde ser normalizada: " + n.Reason + "."
				break
			}
			gain := math.Min(1, math.Max(0, (*c.Score-*n.Score)/100))
			item.Status, item.DamScore, item.Gain = StatusCalculated, n.Score, &gain
			item.Points = cfg.MaxBonusPoints * item.Weight * gain
			item.Explanation = fmt.Sprintf("Score do touro %s, da matriz %s ⇒ ganho %s; pontos = %s × %s × %s = %s.",
				f2(*c.Score), f2(*n.Score), f2(gain), f2(cfg.MaxBonusPoints), f2(item.Weight), f2(gain), f2(item.Points))
			res.Points += item.Points
			any = true
		}
		res.Items = append(res.Items, item)
	}
	if any {
		res.Status = StatusCalculated
		res.Explanation = "Bônus pela melhora do touro sobre a matriz, nas características em que ambos têm avaliação na mesma base."
	} else {
		res.Status = StatusNotCalculable
		res.Explanation = "A matriz não tem avaliações comparáveis (mesma base) com as do touro; complementaridade não calculada (0 ponto)."
	}
	return res
}
