package service

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/google/uuid"

	"genetica/internal/database"
	"genetica/internal/domain"
	"genetica/internal/genetics"
	"genetica/internal/repository"
)

// Services bundles every application service.
type Services struct {
	Health       *HealthService
	Farms        *FarmService
	Animals      *AnimalService
	Measurements *MeasurementService
	Evaluations  *EvaluationService
	Semen        *SemenService
	Matings      *MatingService
	Breeds       *BreedService
	Demo         *DemoService
	Pedigree     *PedigreeService
	Traits       *TraitService
	Libraries    *LibraryService
}

// New wires every service. librariesDir is where installed .genlib files are kept.
func New(db *database.DB, librariesDir string) *Services {
	return &Services{
		Health:       NewHealthService(db),
		Farms:        &FarmService{db},
		Animals:      &AnimalService{db},
		Measurements: &MeasurementService{db},
		Evaluations:  &EvaluationService{db},
		Semen:        &SemenService{db},
		Matings:      &MatingService{db},
		Breeds:       &BreedService{db},
		Demo:         &DemoService{db},
		Traits:       &TraitService{db},
		Libraries:    &LibraryService{db: db, dir: librariesDir},
		Pedigree:     &PedigreeService{db: db, calc: genetics.UnavailableCalculator{}},
	}
}

func newID() string { return uuid.NewString() }
func nowTS() string { return domain.Timestamp(domain.Now()) }

func repos(db *database.DB) *repository.Repos { return repository.New(db.DB) }

// notFoundOrNil converts a not-found lookup into a validation error on a field.
func fieldError(field, msg string) error {
	return &domain.ValidationError{Fields: map[string]string{field: msg}}
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
