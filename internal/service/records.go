package service

import (
	"context"
	"database/sql"
	"strings"

	"genetica/internal/apperr"
	"genetica/internal/database"
	"genetica/internal/domain"
	"genetica/internal/repository"
)

func requireAnimal(ctx context.Context, r *repository.Repos, id, field string) (*domain.Animal, error) {
	a, err := r.Animals.Get(ctx, id)
	if err != nil {
		if apperr.CodeOf(err) == apperr.CodeNotFound {
			return nil, fieldError(field, "Animal não encontrado.")
		}
		return nil, err
	}
	return a, nil
}

// ---- phenotype measurements

type MeasurementService struct{ db *database.DB }

func (s *MeasurementService) Add(ctx context.Context, in MeasurementInput) (*domain.Measurement, error) {
	m := &domain.Measurement{ID: newID(), AnimalID: in.AnimalID, TraitCode: strings.TrimSpace(in.TraitCode), Value: in.Value,
		Unit: in.Unit, MeasuredAt: in.MeasuredAt, Source: in.Source, Notes: in.Notes, CreatedAt: nowTS()}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	r := repos(s.db)
	if _, err := requireAnimal(ctx, r, m.AnimalID, "animalId"); err != nil {
		return nil, err
	}
	trait, err := r.Traits.Resolve(ctx, m.TraitCode, domain.EvalPhenotype, "")
	if err != nil {
		if apperr.CodeOf(err) == apperr.CodeNotFound {
			return nil, fieldError("traitCode", "Característica fenotípica não cadastrada.")
		}
		return nil, err
	}
	if err := trait.ValidateValue(m.Value); err != nil {
		return nil, err
	}
	if m.Unit, err = trait.ResolveUnit(m.Unit); err != nil {
		return nil, err
	}
	if err := r.Measurements.Insert(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *MeasurementService) Delete(ctx context.Context, id string) error {
	return repos(s.db).Measurements.Delete(ctx, id)
}

func (s *MeasurementService) ListByAnimal(ctx context.Context, animalID string) ([]domain.Measurement, error) {
	return repos(s.db).Measurements.ListByAnimal(ctx, animalID)
}

// ---- genetic evaluations

type EvaluationService struct{ db *database.DB }

func (s *EvaluationService) Add(ctx context.Context, in EvaluationInput) (*domain.GeneticEvaluation, error) {
	e := &domain.GeneticEvaluation{ID: newID(), AnimalID: in.AnimalID, DatasetID: in.DatasetID,
		EvaluationProgram: strings.TrimSpace(in.EvaluationProgram), EvaluationRelease: strings.TrimSpace(in.EvaluationRelease),
		BreedCode: in.BreedCode, TraitCode: strings.TrimSpace(in.TraitCode), EvaluationType: in.EvaluationType,
		Value: in.Value, Unit: in.Unit, Accuracy: in.Accuracy, Percentile: in.Percentile, GeneticBase: in.GeneticBase,
		SourceReference: in.SourceReference, EvaluatedAt: in.EvaluatedAt, CreatedAt: nowTS()}
	if err := e.Validate(); err != nil {
		return nil, err
	}
	r := repos(s.db)
	if _, err := requireAnimal(ctx, r, e.AnimalID, "animalId"); err != nil {
		return nil, err
	}
	trait, err := r.Traits.Resolve(ctx, e.TraitCode, e.EvaluationType, e.DatasetID)
	if err != nil {
		if apperr.CodeOf(err) == apperr.CodeNotFound {
			return nil, fieldError("traitCode", "Característica não cadastrada para o tipo "+e.EvaluationType+".")
		}
		return nil, err
	}
	if err := trait.ValidateValue(e.Value); err != nil {
		return nil, err
	}
	if e.Unit, err = trait.ResolveUnit(e.Unit); err != nil {
		return nil, err
	}
	if err := r.Evaluations.Insert(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *EvaluationService) Delete(ctx context.Context, id string) error {
	return repos(s.db).Evaluations.Delete(ctx, id)
}

func (s *EvaluationService) ListByAnimal(ctx context.Context, animalID string) ([]domain.GeneticEvaluation, error) {
	return repos(s.db).Evaluations.ListByAnimal(ctx, animalID)
}

// ---- matings

type MatingService struct{ db *database.DB }

// Save stores a mating with its full calculation snapshot. The calculation
// itself is produced by the genetics engine (later stage); this layer only validates and persists.
func (s *MatingService) Save(ctx context.Context, m domain.Mating) (*domain.Mating, error) {
	m.ID, m.CreatedAt = newID(), nowTS()
	if m.Status == "" {
		m.Status = domain.MatingDraft
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	err := s.db.WithTx(ctx, func(tx *sql.Tx) error {
		r := repository.New(tx)
		f, err := requireAnimal(ctx, r, m.FemaleID, "femaleId")
		if err != nil {
			return err
		}
		if f.Sex != domain.SexFemale {
			return fieldError("femaleId", "A matriz deve ser uma fêmea.")
		}
		b, err := requireAnimal(ctx, r, m.BullID, "bullId")
		if err != nil {
			return err
		}
		if b.Sex != domain.SexMale {
			return fieldError("bullId", "O touro deve ser um macho.")
		}
		return r.Matings.Insert(ctx, &m)
	})
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *MatingService) Get(ctx context.Context, id string) (*domain.Mating, error) {
	return repos(s.db).Matings.Get(ctx, id)
}

func (s *MatingService) ListByFemale(ctx context.Context, femaleID string) ([]domain.Mating, error) {
	return repos(s.db).Matings.ListByFemale(ctx, femaleID)
}
