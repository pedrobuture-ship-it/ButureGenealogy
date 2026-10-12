package service

import (
	"context"
	"database/sql"
	"strings"

	"genetica/internal/apperr"
	"genetica/internal/database"
	"genetica/internal/domain"
	"genetica/internal/genetics"
	"genetica/internal/repository"
)

const (
	MinGenerations = 2
	MaxGenerations = 5
)

// PedigreeService exposes pedigree views and checks. All algorithms live in
// internal/genetics; this layer only loads data and shapes DTOs.
type PedigreeService struct {
	db   *database.DB
	calc genetics.RelationshipCalculator
}

// PedigreeView is everything the pedigree screen needs.
type PedigreeView struct {
	AnimalID    string                     `json:"animalId"`
	Generations int                        `json:"generations"`
	Tree        *genetics.TreeNode         `json:"tree"`
	Depth       int                        `json:"depth"`
	Stats       []genetics.GenerationStat  `json:"stats"`
	Repeated    []genetics.Repeated        `json:"repeated"`
	Issues      []genetics.Issue           `json:"issues"`
	Inbreeding  genetics.CoefficientResult `json:"inbreeding"`
}

// loadGraph loads id and its ancestors up to `generations` generations.
func loadGraph(ctx context.Context, r *repository.Repos, id string, generations int) (genetics.Graph, error) {
	breeds, err := r.Breeds.List(ctx, false)
	if err != nil {
		return nil, err
	}
	breedName := map[string]string{}
	for _, b := range breeds {
		breedName[b.ID] = b.Name
	}
	g := genetics.Graph{}
	add := func(a domain.Animal) {
		g[a.ID] = &genetics.Node{ID: a.ID, Name: a.Name, Registration: a.Registration, Sex: a.Sex,
			BreedName: breedName[a.BreedID], BirthDate: a.BirthDate, FatherID: a.FatherID, MotherID: a.MotherID}
	}
	root, err := r.Animals.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	add(*root)
	frontier := []string{id}
	for gen := 0; gen < generations && len(frontier) > 0; gen++ {
		var want []string
		seen := map[string]bool{}
		for _, fid := range frontier {
			n := g[fid]
			for _, p := range []string{n.FatherID, n.MotherID} {
				if p != "" && g[p] == nil && !seen[p] {
					seen[p] = true
					want = append(want, p)
				}
			}
		}
		if len(want) == 0 {
			break
		}
		got, err := r.Animals.GetMany(ctx, want)
		if err != nil {
			return nil, err
		}
		frontier = frontier[:0]
		for _, a := range got {
			add(a)
			frontier = append(frontier, a.ID)
		}
	}
	return g, nil
}

func clampGenerations(n int) int {
	if n < MinGenerations {
		return MinGenerations
	}
	if n > MaxGenerations {
		return MaxGenerations
	}
	return n
}

// View builds the pedigree view for an animal.
func (s *PedigreeService) View(ctx context.Context, animalID string, generations int) (*PedigreeView, error) {
	generations = clampGenerations(generations)
	g, err := loadGraph(ctx, repos(s.db), animalID, generations)
	if err != nil {
		return nil, err
	}
	issues := genetics.CheckConsistency(g)
	if issues == nil {
		issues = []genetics.Issue{}
	}
	repeated := genetics.RepeatedAncestors(g, animalID, generations)
	if repeated == nil {
		repeated = []genetics.Repeated{}
	}
	return &PedigreeView{
		AnimalID:    animalID,
		Generations: generations,
		Tree:        genetics.BuildTree(g, animalID, generations),
		Depth:       genetics.Depth(g, animalID, generations),
		Stats:       genetics.Completeness(g, animalID, generations),
		Repeated:    repeated,
		Issues:      issues,
		Inbreeding:  s.calc.Inbreeding(g, animalID),
	}, nil
}

// CommonAncestorsResult lists what two animals share.
type CommonAncestorsResult struct {
	A, B         string                     `json:"-"`
	Ancestors    []CommonAncestorView       `json:"ancestors"`
	Relationship genetics.CoefficientResult `json:"relationship"`
}

type CommonAncestorView struct {
	genetics.CommonAncestor
	Name string `json:"name"`
}

func (s *PedigreeService) CommonAncestors(ctx context.Context, a, b string, generations int) (*CommonAncestorsResult, error) {
	generations = clampGenerations(generations)
	r := repos(s.db)
	ga, err := loadGraph(ctx, r, a, generations)
	if err != nil {
		return nil, err
	}
	gb, err := loadGraph(ctx, r, b, generations)
	if err != nil {
		return nil, err
	}
	merged := genetics.Graph{}
	for id, n := range ga {
		merged[id] = n
	}
	for id, n := range gb {
		merged[id] = n
	}
	res := &CommonAncestorsResult{A: a, B: b, Ancestors: []CommonAncestorView{}, Relationship: s.calc.Relationship(merged, a, b)}
	for _, c := range genetics.CommonAncestors(merged, a, b, generations) {
		name := ""
		if n := merged[c.ID]; n != nil {
			name = n.Name
			if name == "" {
				name = n.Registration
			}
		}
		res.Ancestors = append(res.Ancestors, CommonAncestorView{CommonAncestor: c, Name: name})
	}
	return res, nil
}

// AuditIssue is a pedigree issue located in a named animal.
type AuditIssue = genetics.Issue

// Audit checks the whole herd (or one farm) for pedigree inconsistencies.
func (s *PedigreeService) Audit(ctx context.Context, farmID string) ([]AuditIssue, error) {
	all, err := repos(s.db).Animals.All(ctx, farmID)
	if err != nil {
		return nil, err
	}
	g := genetics.Graph{}
	for _, a := range all {
		g[a.ID] = &genetics.Node{ID: a.ID, Name: a.Name, Registration: a.Registration, Sex: a.Sex,
			BirthDate: a.BirthDate, FatherID: a.FatherID, MotherID: a.MotherID}
	}
	issues := genetics.CheckConsistency(g)
	if issues == nil {
		issues = []AuditIssue{}
	}
	return issues, nil
}

// MinimalAncestorInput creates a placeholder-free minimal record for a missing parent.
type MinimalAncestorInput struct {
	ChildID      string `json:"childId"`
	Role         string `json:"role"` // "sire" | "dam"
	Name         string `json:"name"`
	Registration string `json:"registration"`
	BreedID      string `json:"breedId"`
}

// CreateMinimalAncestor creates the missing parent and links it to the child,
// atomically. The new record only carries what the user knows.
func (s *PedigreeService) CreateMinimalAncestor(ctx context.Context, in MinimalAncestorInput) (*domain.Animal, error) {
	if in.Role != "sire" && in.Role != "dam" {
		return nil, fieldError("role", "Informe se é pai ou mãe.")
	}
	var created *domain.Animal
	err := s.db.WithTx(ctx, func(tx *sql.Tx) error {
		r := repository.New(tx)
		child, err := r.Animals.Get(ctx, in.ChildID)
		if err != nil {
			return err
		}
		a := &domain.Animal{
			ID: newID(), FarmID: child.FarmID, Registration: strings.TrimSpace(in.Registration), Name: strings.TrimSpace(in.Name),
			BreedID: in.BreedID, Status: domain.StatusActive, IsDemo: child.IsDemo, CreatedAt: nowTS(),
		}
		a.UpdatedAt = a.CreatedAt
		if in.Role == "sire" {
			if child.FatherID != "" {
				return apperr.New(apperr.CodeConflict, "Este animal já possui um pai registrado.")
			}
			a.Sex, a.Category = domain.SexMale, domain.CategoryBull
		} else {
			if child.MotherID != "" {
				return apperr.New(apperr.CodeConflict, "Este animal já possui uma mãe registrada.")
			}
			a.Sex, a.Category = domain.SexFemale, domain.CategoryCow
		}
		if err := validateAnimal(ctx, r, a); err != nil {
			return err
		}
		if err := r.Animals.Insert(ctx, a); err != nil {
			return err
		}
		if in.Role == "sire" {
			child.FatherID = a.ID
		} else {
			child.MotherID = a.ID
		}
		child.UpdatedAt = nowTS()
		if err := validateAnimal(ctx, r, child); err != nil {
			return err
		}
		if err := r.Animals.Update(ctx, child); err != nil {
			return err
		}
		created = a
		return nil
	})
	return created, err
}
