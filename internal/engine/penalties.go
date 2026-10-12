package engine

import (
	"fmt"
	"math"
)

// LowAccuracyRule subtracts points when the accuracy of used evaluations is
// below a threshold set by the profile:
//
//	shortfall_t = max(0, (T − acc_t)/T)
//	points      = MaxPoints × Σ_t share_t × shortfall_t
//
// Unknown accuracies add no penalty (unknown is not the same as low) and are
// reported as warnings.
type LowAccuracyRule struct{}

const lowAccuracyFormula = "pontos = máximo × Σ(peso_normalizado × max(0, (limiar − acurácia)/limiar))"

func (LowAccuracyRule) Calculate(ctx CalculationContext, comps []ComponentResult) PenaltyResult {
	cfg := ctx.Run.Profile.Penalties.LowAccuracy
	res := PenaltyResult{Name: "Baixa acurácia", Formula: lowAccuracyFormula, Warnings: []string{},
		Source: Source{Parameter: fmt.Sprintf("perfil %q: limiar %s, máximo %s pontos", ctx.Run.Profile.Name, f2(cfg.Threshold), f2(cfg.MaxPoints))}}
	if cfg.Threshold <= 0 || cfg.MaxPoints <= 0 {
		res.Status = StatusNotApplied
		res.Explanation = "Penalidade por baixa acurácia desativada no perfil."
		return res
	}
	calc, rawSum := calculated(comps)
	if len(calc) == 0 || rawSum <= 0 {
		res.Status = StatusNotCalculable
		res.Explanation = "Nenhum componente calculável."
		return res
	}
	unknown := 0
	for _, c := range calc {
		if c.Accuracy == nil {
			unknown++
			continue
		}
		short := math.Max(0, (cfg.Threshold-*c.Accuracy)/cfg.Threshold)
		res.Points += cfg.MaxPoints * (c.RawWeight / rawSum) * short
	}
	if unknown > 0 {
		res.Warnings = append(res.Warnings, fmt.Sprintf("%d componente(s) sem acurácia informada não foram penalizados.", unknown))
	}
	res.Status = StatusCalculated
	if res.Points == 0 {
		res.Explanation = fmt.Sprintf("Nenhuma acurácia abaixo de %s%%.", f2(cfg.Threshold*100))
	} else {
		res.Explanation = fmt.Sprintf("Acurácias abaixo de %s%% reduzem o score em %s ponto(s).", f2(cfg.Threshold*100), f2(res.Points))
	}
	return res
}

// InbreedingRule is a placeholder that never invents a number: the formal
// relationship/inbreeding coefficient is not implemented in this version.
type InbreedingRule struct{}

func (InbreedingRule) Calculate(CalculationContext, []ComponentResult) PenaltyResult {
	return PenaltyResult{
		Name: "Parentesco / endogamia", Status: StatusNotApplied, Points: 0, Warnings: []string{},
		Explanation: "Planejado — ainda não utilizado no score do MVP. O coeficiente de parentesco/endogamia não é calculado nesta versão; ancestrais comuns são apenas listados como alerta.",
	}
}
