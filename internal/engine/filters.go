package engine

import (
	"fmt"

	"genetica/internal/genetics"
)

// StatusFilter keeps only active bulls.
type StatusFilter struct{}

func (StatusFilter) Name() string { return "status" }
func (StatusFilter) Apply(_ RunContext, c Candidate) *Exclusion {
	if c.Status != "" && c.Status != "active" {
		return &Exclusion{"status", fmt.Sprintf("O touro não está ativo (status: %s).", c.Status)}
	}
	return nil
}

// StockFilter requires semen in stock.
type StockFilter struct{}

func (StockFilter) Name() string { return "estoque" }
func (StockFilter) Apply(run RunContext, c Candidate) *Exclusion {
	f := run.Profile.Filters
	if !f.RequireStock && f.MinDoses <= 0 {
		return nil
	}
	min := f.MinDoses
	if f.RequireStock && min < 1 {
		min = 1
	}
	if c.Stock.Doses < min {
		return &Exclusion{"estoque", fmt.Sprintf("Estoque insuficiente: %d dose(s), mínimo exigido %d.", c.Stock.Doses, min)}
	}
	return nil
}

// CostFilter enforces the maximum cost per dose.
type CostFilter struct{}

func (CostFilter) Name() string { return "custo" }
func (CostFilter) Apply(run RunContext, c Candidate) *Exclusion {
	max := run.Profile.Filters.MaxUnitCost
	if max == nil {
		return nil
	}
	if c.Stock.UnitCost == nil {
		return &Exclusion{"custo", "Custo por dose desconhecido, mas o perfil define um custo máximo."}
	}
	if *c.Stock.UnitCost > *max {
		return &Exclusion{"custo", fmt.Sprintf("Custo por dose %.2f acima do máximo %.2f.", *c.Stock.UnitCost, *max)}
	}
	return nil
}

// RequiredTraitsFilter excludes candidates lacking any evaluation of a required trait.
type RequiredTraitsFilter struct{}

func (RequiredTraitsFilter) Name() string { return "características obrigatórias" }
func (RequiredTraitsFilter) Apply(run RunContext, c Candidate) *Exclusion {
	for _, t := range run.Profile.ActiveTraits() {
		if !t.Required {
			continue
		}
		found := false
		for _, e := range c.Evals {
			if e.Basis.TraitCode == t.TraitCode && e.Basis.EvaluationType == t.EvaluationType {
				found = true
				break
			}
		}
		if !found {
			return &Exclusion{"características obrigatórias", fmt.Sprintf("Sem avaliação %s de %s, exigida pelo perfil.", t.EvaluationType, t.TraitCode)}
		}
	}
	return nil
}

// MinAccuracyFilter excludes candidates whose evaluation of a weighted trait is
// below the minimum accuracy. When a minimum is set, an unknown accuracy cannot
// be verified and therefore fails.
type MinAccuracyFilter struct{}

func (MinAccuracyFilter) Name() string { return "acurácia mínima" }
func (MinAccuracyFilter) Apply(run RunContext, c Candidate) *Exclusion {
	min := run.Profile.Filters.MinAccuracy
	if min == nil {
		return nil
	}
	for _, t := range run.Profile.ActiveTraits() {
		best := -1.0
		known, any := false, false
		for _, e := range c.Evals {
			if e.Basis.TraitCode != t.TraitCode || e.Basis.EvaluationType != t.EvaluationType {
				continue
			}
			any = true
			if e.Accuracy != nil {
				known = true
				if *e.Accuracy > best {
					best = *e.Accuracy
				}
			}
		}
		if !any {
			continue // missing evaluations are handled by RequiredTraitsFilter / coverage
		}
		if !known {
			return &Exclusion{"acurácia mínima", fmt.Sprintf("A avaliação de %s não informa acurácia, mas o perfil exige mínimo de %.0f%%.", t.TraitCode, *min*100)}
		}
		if best < *min {
			return &Exclusion{"acurácia mínima", fmt.Sprintf("Acurácia de %s (%.0f%%) abaixo do mínimo de %.0f%%.", t.TraitCode, best*100, *min*100)}
		}
	}
	return nil
}

// AncestorFilter excludes bulls that are direct ancestors of the dam.
type AncestorFilter struct{}

func (AncestorFilter) Name() string { return "ancestral direto" }
func (AncestorFilter) Apply(run RunContext, c Candidate) *Exclusion {
	if !run.Profile.Filters.ExcludeDirectAncestors || run.Pedigree == nil {
		return nil
	}
	for _, o := range genetics.Ancestors(run.Pedigree, run.Female.ID, run.Profile.Filters.AncestorGenerations) {
		if o.ID == c.ID {
			return &Exclusion{"ancestral direto", fmt.Sprintf("O touro é ancestral da matriz (geração %d, caminho %s).", o.Generation, o.Path)}
		}
	}
	return nil
}

// DefaultFilters is the standard filter chain, in order.
func DefaultFilters() []HardFilter {
	return []HardFilter{StatusFilter{}, StockFilter{}, CostFilter{}, RequiredTraitsFilter{}, MinAccuracyFilter{}, AncestorFilter{}}
}
