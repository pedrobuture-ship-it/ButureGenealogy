package engine

import (
	"fmt"
	"math"
	"strings"
)

// DefaultAggregator turns components into the final score:
//
//	weight_i       = raw_weight_i / Σ raw_weight (calculable components only)
//	contribution_i = score_i × weight_i
//	base           = Σ contribution_i
//	final          = clamp(base + complementarity − penalties, 0, 100)
//
// If the calculable share of the profile weight is below MinCoverage the pair
// has no score at all ("not calculable"): a partial number would mislead.
type DefaultAggregator struct{}

const finalFormula = "final = limitar(base + complementaridade − penalidades, 0, 100) ; base = Σ(score × peso_normalizado)"

const eps = 1e-9

func (DefaultAggregator) Aggregate(ctx CalculationContext, comps []ComponentResult, comp ComplementarityResult, pens []PenaltyResult) (*FinalScore, []ComponentResult, []string) {
	weighted := make([]ComponentResult, len(comps))
	copy(weighted, comps)

	var totalRaw, calcRaw float64
	var missing []string
	for _, c := range comps {
		totalRaw += c.RawWeight
		if c.Status == StatusCalculated {
			calcRaw += c.RawWeight
		} else {
			missing = append(missing, c.Name)
		}
	}
	if len(comps) == 0 || totalRaw <= 0 {
		return nil, weighted, []string{"O perfil não tem nenhuma característica com peso."}
	}
	coverage := calcRaw / totalRaw
	min := ctx.Run.Profile.MinCoverage
	if min <= 0 || min > 1 {
		min = 1
	}
	if calcRaw <= 0 || coverage+eps < min {
		reasons := []string{fmt.Sprintf("Apenas %s%% do peso do perfil é calculável (mínimo exigido: %s%%).", f2(coverage*100), f2(min*100))}
		if len(missing) > 0 {
			reasons = append(reasons, "Componentes não calculáveis: "+strings.Join(missing, ", ")+".")
		}
		return nil, weighted, reasons
	}

	var base float64
	for i := range weighted {
		c := &weighted[i]
		if c.Status != StatusCalculated {
			continue
		}
		c.Weight = c.RawWeight / calcRaw
		contrib := *c.Score * c.Weight
		c.Contribution = &contrib
		base += contrib
	}
	var pen float64
	for _, p := range pens {
		if p.Status == StatusCalculated {
			pen += p.Points
		}
	}
	compPts := 0.0
	if comp.Status == StatusCalculated {
		compPts = comp.Points
	}
	un := base + compPts - pen
	final := math.Min(100, math.Max(0, un))
	return &FinalScore{Base: base, Complementarity: compPts, Penalties: pen, Unclamped: un, Final: final,
		Clamped: math.Abs(final-un) > eps, Coverage: coverage, Formula: finalFormula}, weighted, nil
}
