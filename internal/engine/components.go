package engine

import (
	"fmt"
)

// TraitComponent is the ScoringComponent of one weighted trait. It normalizes
// the candidate's evaluation on the reference basis, or explains why it cannot.
type TraitComponent struct {
	Weight     TraitWeight
	Normalizer Normalizer
}

func traitName(tw TraitWeight, e *Evaluation) string {
	name := tw.TraitCode
	if e != nil && e.TraitName != "" {
		name = e.TraitName
	}
	return fmt.Sprintf("%s (%s)", name, tw.EvaluationType)
}

func direction(tw TraitWeight, e *Evaluation) string {
	if tw.Direction != "" {
		return tw.Direction
	}
	if e != nil {
		return e.Direction
	}
	return ""
}

func (tc TraitComponent) Calculate(ctx CalculationContext) ComponentResult {
	tw := tc.Weight
	key := tw.Key()
	sel := ctx.Selected[key]
	weightSource := fmt.Sprintf("perfil %q: peso %s", ctx.Run.Profile.Name, f2(tw.Weight))
	res := ComponentResult{
		Name: traitName(tw, sel), TraitCode: tw.TraitCode, EvaluationType: tw.EvaluationType,
		Direction: direction(tw, sel), RawWeight: tw.Weight, Warnings: []string{},
		Source: Source{Parameter: weightSource},
	}
	bull := ctx.Candidate.Label()

	if sel == nil {
		res.Status = StatusNotCalculable
		if inc := ctx.Incompatible[key]; len(inc) > 0 {
			ref := ctx.Run.Bases[key]
			refText := "indefinida"
			if ref != nil && ref.Basis != nil {
				refText = ref.Basis.Describe()
			}
			res.Warnings = append(res.Warnings, IncompatibleBasesMessage)
			res.Explanation = fmt.Sprintf("%s tem %s apenas em base diferente da usada para comparar os candidatos (%s). Avaliações de bases diferentes não são equivalentes e não são convertidas.",
				bull, tw.EvaluationType+" de "+tw.TraitCode, refText)
			for _, e := range inc {
				res.Warnings = append(res.Warnings, fmt.Sprintf("Base do touro: %s", e.Basis.Describe()))
			}
			e := inc[0]
			res.Source = sourceOf(&e)
			res.Source.Parameter = weightSource
		} else {
			res.Explanation = fmt.Sprintf("%s não tem avaliação %s de %s.", bull, tw.EvaluationType, tw.TraitCode)
		}
		return res
	}

	b := sel.Basis
	v := sel.Value
	res.Basis, res.InputValue, res.Unit = &b, &v, b.Unit
	res.Accuracy, res.Percentile = sel.Accuracy, sel.Percentile
	res.Source = sourceOf(sel)
	res.Source.Parameter = weightSource // the value comes from the dataset; the weight comes from the profile
	if sel.IsDemo {
		res.Warnings = append(res.Warnings, "Valor DEMO / SYNTHETIC: dado fictício, sem valor científico.")
	}
	if sel.Accuracy == nil {
		res.Warnings = append(res.Warnings, "Avaliação sem acurácia informada.")
	}
	for _, e := range ctx.Run.Female.Evals {
		if e.Basis.Key() == b.Key() {
			dv := e.Value
			res.DamValue = &dv
			break
		}
	}

	n := tc.Normalizer.Normalize(NormalizeInput{Weight: tw, Eval: *sel, Mode: ctx.Run.Profile.Normalization})
	if n.Score == nil {
		res.Status = StatusNotCalculable
		res.Explanation = fmt.Sprintf("%s: %s = %s %s não pôde ser normalizado — %s.", bull, tw.TraitCode, f2(v), b.Unit, n.Reason)
		res.Warnings = append(res.Warnings, "Não calculável: "+n.Reason)
		return res
	}
	res.Status = StatusCalculated
	res.Score, res.Method, res.Transformation, res.Formula = n.Score, n.Method, n.Transformation, n.Formula
	res.Explanation = fmt.Sprintf("%s: %s de %s = %s %s (%s %s; programa %s, release %s). %s.",
		bull, tw.EvaluationType, tw.TraitCode, f2(v), b.Unit, sel.DatasetName, sel.DatasetVersion, b.Program, b.Release, n.Transformation)
	return res
}
