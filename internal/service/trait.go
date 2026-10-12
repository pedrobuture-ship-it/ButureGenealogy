package service

import (
	"context"
	"sort"
	"strings"

	"genetica/internal/apperr"
	"genetica/internal/database"
	"genetica/internal/domain"
	"genetica/internal/repository"
)

// TraitService manages the trait catalogue and builds per-animal trait views.
type TraitService struct{ db *database.DB }

type TraitInput struct {
	Code                  string   `json:"code"`
	Name                  string   `json:"name"`
	Description           string   `json:"description"`
	Category              string   `json:"category"`
	Purpose               string   `json:"purpose"`
	Unit                  string   `json:"unit"`
	DataType              string   `json:"dataType"`
	EvaluationType        string   `json:"evaluationType"`
	OptimizationDirection string   `json:"optimizationDirection"`
	IdealMin              *float64 `json:"idealMin"`
	IdealMax              *float64 `json:"idealMax"`
	HigherIsBetter        *bool    `json:"higherIsBetter"`
}

type TraitQuery struct {
	Kind     string `json:"kind"`
	Category string `json:"category"`
	Search   string `json:"search"`
}

func (s *TraitService) List(ctx context.Context, q TraitQuery) ([]domain.Trait, error) {
	return repos(s.db).Traits.List(ctx, repository.TraitFilter{Kind: q.Kind, Category: q.Category, Search: q.Search})
}

func (s *TraitService) Kinds() []domain.ValueKindInfo { return domain.KindInfo() }

// Create adds a user-defined trait. Dataset traits are created by the importer.
func (s *TraitService) Create(ctx context.Context, in TraitInput) (*domain.Trait, error) {
	t := &domain.Trait{
		ID: newID(), Code: strings.TrimSpace(in.Code), Name: strings.TrimSpace(in.Name), Description: in.Description,
		Category: in.Category, Purpose: defaultStr(in.Purpose, "any"), Unit: strings.TrimSpace(in.Unit),
		DataType: defaultStr(in.DataType, domain.DataNumber), EvaluationType: defaultStr(in.EvaluationType, domain.EvalPhenotype),
		OptimizationDirection: defaultStr(in.OptimizationDirection, domain.InformationOnly),
		IdealMin:              in.IdealMin, IdealMax: in.IdealMax, HigherIsBetter: in.HigherIsBetter,
		Metadata: "{}", CreatedAt: nowTS(),
	}
	if err := t.Validate(); err != nil {
		return nil, err
	}
	if err := repos(s.db).Traits.Insert(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func defaultStr(s, d string) string {
	if s == "" {
		return d
	}
	return s
}

// Provenance answers "where did this value come from?" for any trait value.
type Provenance struct {
	KindCode        string   `json:"kindCode"`
	KindLabel       string   `json:"kindLabel"`
	Genetic         bool     `json:"genetic"`
	TraitCode       string   `json:"traitCode"`
	TraitName       string   `json:"traitName"`
	DatasetID       string   `json:"datasetId"`
	DatasetName     string   `json:"datasetName"`
	DatasetVersion  string   `json:"datasetVersion"`
	Program         string   `json:"program"`
	Release         string   `json:"release"`
	GeneticBase     string   `json:"geneticBase"`
	SourceReference string   `json:"sourceReference"`
	Accuracy        *float64 `json:"accuracy"`
	Source          string   `json:"source"` // for measurements: who/what measured
	Date            string   `json:"date"`
	IsDemo          bool     `json:"isDemo"`
}

// TraitItem is one value of one trait of an animal, with its definition and provenance.
type TraitItem struct {
	ID         string     `json:"id"`
	Source     string     `json:"source"` // "measurement" | "evaluation"
	TraitCode  string     `json:"traitCode"`
	TraitName  string     `json:"traitName"`
	Defined    bool       `json:"defined"` // false when no catalogue entry exists
	Category   string     `json:"category"`
	ValueKind  string     `json:"valueKind"`
	Unit       string     `json:"unit"`
	Value      float64    `json:"value"`
	Accuracy   *float64   `json:"accuracy"`
	Percentile *float64   `json:"percentile"`
	Notes      string     `json:"notes"`
	Direction  string     `json:"optimizationDirection"`
	Provenance Provenance `json:"provenance"`
}

type TraitGroup struct {
	Category string      `json:"category"`
	Items    []TraitItem `json:"items"`
}

// ForAnimal returns the animal's phenotypes and genetic evaluations grouped by
// trait category. Phenotypes and each genetic kind stay separate items.
func (s *TraitService) ForAnimal(ctx context.Context, animalID string) ([]TraitGroup, error) {
	r := repos(s.db)
	if _, err := r.Animals.Get(ctx, animalID); err != nil {
		return nil, err
	}
	kindLabel := map[string]domain.ValueKindInfo{}
	for _, k := range domain.KindInfo() {
		kindLabel[k.Kind] = k
	}
	var items []TraitItem

	meas, err := r.Measurements.ListByAnimal(ctx, animalID)
	if err != nil {
		return nil, err
	}
	for _, m := range meas {
		it := TraitItem{ID: m.ID, Source: "measurement", TraitCode: m.TraitCode, TraitName: m.TraitCode, Category: domain.CatOther,
			ValueKind: domain.EvalPhenotype, Unit: m.Unit, Value: m.Value, Notes: m.Notes, Direction: domain.InformationOnly}
		it.Provenance = Provenance{KindCode: domain.EvalPhenotype, KindLabel: kindLabel[domain.EvalPhenotype].Label, TraitCode: m.TraitCode,
			TraitName: m.TraitCode, Source: m.Source, Date: m.MeasuredAt, IsDemo: m.IsDemo}
		if t, err := r.Traits.Resolve(ctx, m.TraitCode, domain.EvalPhenotype, ""); err == nil {
			applyTrait(&it, t)
		} else if apperr.CodeOf(err) != apperr.CodeNotFound {
			return nil, err
		}
		items = append(items, it)
	}

	evals, err := r.Evaluations.ListByAnimal(ctx, animalID)
	if err != nil {
		return nil, err
	}
	for _, e := range evals {
		k := kindLabel[e.EvaluationType]
		it := TraitItem{ID: e.ID, Source: "evaluation", TraitCode: e.TraitCode, TraitName: e.TraitCode, Category: domain.CatOther,
			ValueKind: e.EvaluationType, Unit: e.Unit, Value: e.Value, Accuracy: e.Accuracy, Percentile: e.Percentile, Direction: domain.InformationOnly}
		it.Provenance = Provenance{KindCode: e.EvaluationType, KindLabel: k.Label, Genetic: k.Genetic, TraitCode: e.TraitCode, TraitName: e.TraitCode,
			DatasetID: e.DatasetID, Program: e.EvaluationProgram, Release: e.EvaluationRelease, GeneticBase: e.GeneticBase,
			SourceReference: e.SourceReference, Accuracy: e.Accuracy, Date: e.EvaluatedAt, IsDemo: e.IsDemo}
		if e.DatasetID != "" {
			var name, ver string
			if err := s.db.QueryRowContext(ctx, `SELECT name, version FROM datasets WHERE id=?`, e.DatasetID).Scan(&name, &ver); err == nil {
				it.Provenance.DatasetName, it.Provenance.DatasetVersion = name, ver
			}
		}
		if t, err := r.Traits.Resolve(ctx, e.TraitCode, e.EvaluationType, e.DatasetID); err == nil {
			applyTrait(&it, t)
		} else if apperr.CodeOf(err) != apperr.CodeNotFound {
			return nil, err
		}
		items = append(items, it)
	}

	byCat := map[string][]TraitItem{}
	for _, it := range items {
		byCat[it.Category] = append(byCat[it.Category], it)
	}
	var groups []TraitGroup
	for _, c := range domain.TraitCategories {
		list := byCat[c]
		if len(list) == 0 {
			continue
		}
		sort.SliceStable(list, func(i, j int) bool {
			if list[i].TraitName != list[j].TraitName {
				return strings.ToLower(list[i].TraitName) < strings.ToLower(list[j].TraitName)
			}
			if list[i].ValueKind != list[j].ValueKind {
				return list[i].ValueKind < list[j].ValueKind
			}
			return list[i].Provenance.Date > list[j].Provenance.Date
		})
		groups = append(groups, TraitGroup{Category: c, Items: list})
	}
	if groups == nil {
		groups = []TraitGroup{}
	}
	return groups, nil
}

func applyTrait(it *TraitItem, t *domain.Trait) {
	it.Defined, it.TraitName, it.Category, it.Direction = true, t.Name, t.Category, t.OptimizationDirection
	if it.Unit == "" {
		it.Unit = t.Unit
	}
	it.Provenance.TraitName = t.Name
}
