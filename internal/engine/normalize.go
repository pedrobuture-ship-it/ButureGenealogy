package engine

import (
	"fmt"
	"math"
	"strings"
)

// DefaultNormalizer supports exactly three documented methods:
//
//  1. percentile  — only when the dataset declares the percentile semantics
//     (trait metadata "percentileSemantics": ascending | descending).
//     MAXIMIZE: score = percentile (ascending) or 100 − percentile (descending).
//     MINIMIZE: the opposite.
//  2. zscore      — only when the dataset supplies the population mean and
//     standard deviation (trait metadata "populationMean", "populationSd").
//     z = (x − μ)/σ ; score = 100·Φ(z) for MAXIMIZE, 100·(1 − Φ(z)) for
//     MINIMIZE, with Φ the standard normal CDF (assumes approximate normality).
//  3. target_range — for TARGET_RANGE traits, using the profile's/trait's ideal
//     range and a tolerance given by the profile: 100 inside [min, max]; outside,
//     100·(1 − distance/tolerance) floored at 0.
//
// If none can be applied it returns a Reason: the value is "not calculable".
type DefaultNormalizer struct{}

// f2 formats a number with two decimals and a decimal comma (pt-BR) for explanations.
func f2(x float64) string { return strings.Replace(fmt.Sprintf("%.2f", x), ".", ",", 1) }

// Phi is the standard normal cumulative distribution function.
func Phi(z float64) float64 { return 0.5 * (1 + math.Erf(z/math.Sqrt2)) }

func (DefaultNormalizer) Normalize(in NormalizeInput) NormalizeResult {
	dir := in.Weight.Direction
	if dir == "" {
		dir = in.Eval.Direction
	}
	switch dir {
	case Maximize, Minimize:
	case TargetRange:
		return normalizeTarget(in)
	default:
		return NormalizeResult{Reason: "a direção de otimização da característica não está definida (informe MAXIMIZE, MINIMIZE ou TARGET_RANGE no perfil)"}
	}

	mode := in.Mode
	if mode == "" {
		mode = NormAuto
	}
	var reasons []string

	if mode == NormAuto || mode == NormPercentile {
		if r, why := byPercentile(in, dir); r != nil {
			return *r
		} else {
			reasons = append(reasons, why)
		}
	}
	if mode == NormAuto || mode == NormZScore {
		if r, why := byZScore(in, dir); r != nil {
			return *r
		} else {
			reasons = append(reasons, why)
		}
	}
	return NormalizeResult{Reason: "nenhum método de normalização aplicável: " + strings.Join(reasons, "; ")}
}

func byPercentile(in NormalizeInput, dir string) (*NormalizeResult, string) {
	p := in.Eval.Percentile
	if p == nil {
		return nil, "o valor não tem percentil"
	}
	if *p < 0 || *p > 100 || math.IsNaN(*p) {
		return nil, "percentil fora do intervalo 0–100"
	}
	sem := in.Eval.Meta.PercentileSemantics
	if sem != PercentileAscending && sem != PercentileDescending {
		return nil, "o dataset não declara a semântica do percentil (ascending/descending)"
	}
	score := *p
	desc := "percentil ascendente (valor maior ⇒ percentil maior)"
	if sem == PercentileDescending {
		score = 100 - *p
		desc = "percentil descendente (valor maior ⇒ percentil menor)"
	}
	formula := "score = percentil"
	if sem == PercentileDescending {
		formula = "score = 100 − percentil"
	}
	if dir == Minimize {
		score = 100 - score
		if sem == PercentileDescending {
			formula = "score = percentil"
		} else {
			formula = "score = 100 − percentil"
		}
		desc += "; MINIMIZE inverte o sentido"
	}
	return &NormalizeResult{Score: &score, Method: NormPercentile,
		Transformation: fmt.Sprintf("percentil %s → %s (%s)", f2(*p), f2(score), desc), Formula: formula}, ""
}

func byZScore(in NormalizeInput, dir string) (*NormalizeResult, string) {
	m := in.Eval.Meta
	if m.PopulationMean == nil || m.PopulationSD == nil {
		return nil, "o dataset não fornece média e desvio-padrão populacionais"
	}
	if !(*m.PopulationSD > 0) || math.IsInf(*m.PopulationSD, 0) || math.IsNaN(*m.PopulationSD) {
		return nil, "desvio-padrão populacional inválido"
	}
	z := (in.Eval.Value - *m.PopulationMean) / *m.PopulationSD
	score := 100 * Phi(z)
	formula := "z = (x − μ)/σ ; score = 100·Φ(z)"
	if dir == Minimize {
		score = 100 - score
		formula = "z = (x − μ)/σ ; score = 100·(1 − Φ(z))"
	}
	return &NormalizeResult{Score: &score, Method: NormZScore, Formula: formula,
		Transformation: fmt.Sprintf("z = (%s − %s)/%s = %s ; score = %s", f2(in.Eval.Value), f2(*m.PopulationMean), f2(*m.PopulationSD), f2(z), f2(score))}, ""
}

func normalizeTarget(in NormalizeInput) NormalizeResult {
	lo, hi := in.Weight.IdealMin, in.Weight.IdealMax
	if lo == nil || hi == nil {
		lo, hi = in.Eval.IdealMin, in.Eval.IdealMax
	}
	if lo == nil || hi == nil || *lo > *hi {
		return NormalizeResult{Reason: "faixa-alvo sem mínimo e máximo ideais válidos"}
	}
	tol := in.Weight.TargetTolerance
	if tol == nil || !(*tol > 0) {
		return NormalizeResult{Reason: "faixa-alvo sem tolerância definida no perfil (distância além da faixa em que o score chega a 0)"}
	}
	v := in.Eval.Value
	dist := 0.0
	switch {
	case v < *lo:
		dist = *lo - v
	case v > *hi:
		dist = v - *hi
	}
	score := 100 * math.Max(0, 1 - dist / *tol)
	var tr string
	if dist == 0 {
		tr = fmt.Sprintf("valor %s dentro da faixa [%s, %s] → 100", f2(v), f2(*lo), f2(*hi))
	} else {
		tr = fmt.Sprintf("valor %s a %s da faixa [%s, %s]; tolerância %s → %s", f2(v), f2(dist), f2(*lo), f2(*hi), f2(*tol), f2(score))
	}
	return NormalizeResult{Score: &score, Method: "target_range", Transformation: tr,
		Formula: "score = 100 · max(0, 1 − distância/tolerância)"}
}
