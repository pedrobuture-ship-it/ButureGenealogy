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

type AnimalService struct{ db *database.DB }

func (in AnimalInput) toAnimal() *domain.Animal {
	status := in.Status
	if status == "" {
		status = domain.StatusActive
	}
	return &domain.Animal{
		FarmID: in.FarmID, Registration: strings.TrimSpace(in.Registration), Name: strings.TrimSpace(in.Name),
		Sex: in.Sex, BreedID: in.BreedID, BirthDate: in.BirthDate, Category: in.Category, Status: status,
		FatherID: in.FatherID, MotherID: in.MotherID, Notes: in.Notes,
	}
}

func (s *AnimalService) Create(ctx context.Context, in AnimalInput) (*domain.Animal, error) {
	a := in.toAnimal()
	a.ID = newID()
	a.CreatedAt = nowTS()
	a.UpdatedAt = a.CreatedAt
	err := s.db.WithTx(ctx, func(tx *sql.Tx) error {
		r := repository.New(tx)
		if err := validateAnimal(ctx, r, a); err != nil {
			return err
		}
		return r.Animals.Insert(ctx, a)
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *AnimalService) Update(ctx context.Context, id string, in AnimalInput) (*domain.Animal, error) {
	var out *domain.Animal
	err := s.db.WithTx(ctx, func(tx *sql.Tx) error {
		r := repository.New(tx)
		cur, err := r.Animals.Get(ctx, id)
		if err != nil {
			return err
		}
		a := in.toAnimal()
		a.ID, a.IsDemo, a.CreatedAt, a.UpdatedAt = cur.ID, cur.IsDemo, cur.CreatedAt, nowTS()
		if err := validateAnimal(ctx, r, a); err != nil {
			return err
		}
		if err := r.Animals.Update(ctx, a); err != nil {
			return err
		}
		out = a
		return nil
	})
	return out, err
}

// Archive soft-deletes: the animal stays in the database and in pedigrees.
func (s *AnimalService) Archive(ctx context.Context, id string) error {
	return repos(s.db).Animals.SetStatus(ctx, id, domain.StatusArchived, nowTS())
}

func (s *AnimalService) Restore(ctx context.Context, id string) error {
	return repos(s.db).Animals.SetStatus(ctx, id, domain.StatusActive, nowTS())
}

func (s *AnimalService) Get(ctx context.Context, id string) (*domain.Animal, error) {
	return repos(s.db).Animals.Get(ctx, id)
}

// AnimalPage is one page of a listing.
type AnimalPage struct {
	Items []domain.Animal `json:"items"`
	Total int             `json:"total"`
	Page  int             `json:"page"`
}

func (s *AnimalService) List(ctx context.Context, q AnimalQuery) (*AnimalPage, error) {
	size := q.PageSize
	if size <= 0 {
		size = 50
	}
	if size > 500 {
		size = 500
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	f := repository.AnimalFilter{
		FarmID: q.FarmID, BreedID: q.BreedID, Sex: q.Sex, Category: q.Category, Status: q.Status,
		IncludeArchived: q.IncludeArchived, Query: q.Search, SortBy: q.SortBy, SortDesc: q.SortDesc,
		Limit: size, Offset: (page - 1) * size,
	}
	today := domain.Now()
	// Older animals have earlier birth dates: max age => lower bound on the date.
	if q.MaxAgeMonths != nil {
		f.BornOnOrAfter = today.AddDate(0, -*q.MaxAgeMonths, 0).Format(domain.DateLayout)
	}
	if q.MinAgeMonths != nil {
		f.BornOnOrBefore = today.AddDate(0, -*q.MinAgeMonths, 0).Format(domain.DateLayout)
	}
	items, total, err := repos(s.db).Animals.List(ctx, f)
	if err != nil {
		return nil, err
	}
	return &AnimalPage{Items: items, Total: total, Page: page}, nil
}

// validateAnimal runs the isolated validation plus rules that need other rows:
// existing farm/breed/parents, parent sex and pedigree cycles.
func validateAnimal(ctx context.Context, r *repository.Repos, a *domain.Animal) error {
	if err := a.Validate(); err != nil {
		return err
	}
	if _, err := r.Farms.Get(ctx, a.FarmID); err != nil {
		if apperr.CodeOf(err) == apperr.CodeNotFound {
			return fieldError("farmId", "Propriedade não encontrada.")
		}
		return err
	}
	if a.BreedID != "" {
		ok, err := r.Breeds.Exists(ctx, a.BreedID)
		if err != nil {
			return err
		}
		if !ok {
			return fieldError("breedId", "Raça não encontrada.")
		}
	}
	if err := checkParent(ctx, r, a, a.FatherID, domain.SexMale, "fatherId", "O pai deve ser um macho."); err != nil {
		return err
	}
	return checkParent(ctx, r, a, a.MotherID, domain.SexFemale, "motherId", "A mãe deve ser uma fêmea.")
}

func checkParent(ctx context.Context, r *repository.Repos, a *domain.Animal, parentID, wantSex, field, sexMsg string) error {
	if parentID == "" {
		return nil
	}
	p, err := r.Animals.Get(ctx, parentID)
	if err != nil {
		if apperr.CodeOf(err) == apperr.CodeNotFound {
			return fieldError(field, "Animal não encontrado.")
		}
		return err
	}
	if p.Sex != wantSex {
		return fieldError(field, sexMsg)
	}
	if a.BirthDate != "" && p.BirthDate != "" && p.BirthDate >= a.BirthDate {
		return fieldError(field, "O ascendente deve ter nascido antes do animal.")
	}
	if a.ID != "" {
		cyc, err := reachesAncestor(ctx, r, parentID, a.ID)
		if err != nil {
			return err
		}
		if cyc {
			return fieldError(field, "Esta ligação criaria um ciclo no pedigree.")
		}
	}
	return nil
}

// reachesAncestor reports whether target is `start` or one of its ancestors.
func reachesAncestor(ctx context.Context, r *repository.Repos, start, target string) (bool, error) {
	seen := map[string]bool{}
	stack := []string{start}
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if id == target {
			return true, nil
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		cur, err := r.Animals.Get(ctx, id)
		if err != nil {
			return false, err
		}
		if cur.FatherID != "" {
			stack = append(stack, cur.FatherID)
		}
		if cur.MotherID != "" {
			stack = append(stack, cur.MotherID)
		}
	}
	return false, nil
}
