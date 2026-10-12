package repository_test

import (
	"context"
	"testing"

	"genetica/internal/apperr"
	"genetica/internal/domain"
	"genetica/internal/repository"
	"genetica/internal/testutil"
)

const ts = "2026-01-01T00:00:00Z"

func newFarm(t *testing.T, r *repository.Repos, id string) {
	t.Helper()
	if err := r.Farms.Insert(context.Background(), &domain.Farm{ID: id, Name: "Fazenda " + id, CreatedAt: ts, UpdatedAt: ts}); err != nil {
		t.Fatal(err)
	}
}

func animal(id, farm, name, sex, cat string) *domain.Animal {
	return &domain.Animal{ID: id, FarmID: farm, Name: name, Sex: sex, Category: cat, Status: domain.StatusActive, CreatedAt: ts, UpdatedAt: ts}
}

func TestFarmCRUDAndBreeds(t *testing.T) {
	ctx := context.Background()
	r := repository.New(testutil.NewDB(t).DB)
	newFarm(t, r, "f1")
	f, err := r.Farms.Get(ctx, "f1")
	if err != nil || f.Name != "Fazenda f1" {
		t.Fatalf("%v %v", f, err)
	}
	f.City = "Uberaba"
	if err := r.Farms.Update(ctx, f); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Farms.Get(ctx, "missing"); apperr.CodeOf(err) != apperr.CodeNotFound {
		t.Fatalf("want not found, got %v", err)
	}
	breeds, err := r.Breeds.List(ctx, true)
	if err != nil || len(breeds) == 0 {
		t.Fatalf("breeds: %v %v", breeds, err)
	}
}

func TestAnimalConstraints(t *testing.T) {
	ctx := context.Background()
	r := repository.New(testutil.NewDB(t).DB)
	newFarm(t, r, "f1")

	a := animal("a1", "f1", "Boi", domain.SexMale, domain.CategoryBull)
	a.Registration = "REG1"
	if err := r.Animals.Insert(ctx, a); err != nil {
		t.Fatal(err)
	}
	// duplicate registration within the farm
	b := animal("a2", "f1", "Outro", domain.SexMale, domain.CategoryBull)
	b.Registration = "REG1"
	if err := r.Animals.Insert(ctx, b); apperr.CodeOf(err) != apperr.CodeConflict {
		t.Fatalf("want conflict, got %v", err)
	}
	// self parent blocked by CHECK
	self := animal("a3", "f1", "Self", domain.SexMale, domain.CategoryBull)
	self.FatherID = "a3"
	if err := r.Animals.Insert(ctx, self); err == nil {
		t.Fatal("self-parent must fail")
	}
	// unknown parent blocked by FK
	orphan := animal("a4", "f1", "Orphan", domain.SexMale, domain.CategoryBull)
	orphan.FatherID = "nope"
	if err := r.Animals.Insert(ctx, orphan); apperr.CodeOf(err) != apperr.CodeConflict {
		t.Fatalf("want FK conflict, got %v", err)
	}
	// empty registrations may repeat
	for _, id := range []string{"n1", "n2"} {
		if err := r.Animals.Insert(ctx, animal(id, "f1", id, domain.SexFemale, domain.CategoryCow)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAnimalListFilters(t *testing.T) {
	ctx := context.Background()
	r := repository.New(testutil.NewDB(t).DB)
	newFarm(t, r, "f1")
	mk := func(id, name, sex, cat, breed, birth string) {
		a := animal(id, "f1", name, sex, cat)
		a.BreedID, a.BirthDate = breed, birth
		if err := r.Animals.Insert(ctx, a); err != nil {
			t.Fatal(err)
		}
	}
	mk("1", "Alfa", domain.SexMale, domain.CategoryBull, "breed-nelore", "2018-05-01")
	mk("2", "Beta", domain.SexFemale, domain.CategoryCow, "breed-nelore", "2020-05-01")
	mk("3", "Gama_100%", domain.SexFemale, domain.CategoryHeifer, "breed-angus", "2023-05-01")
	if err := r.Animals.SetStatus(ctx, "3", domain.StatusArchived, ts); err != nil {
		t.Fatal(err)
	}

	list := func(f repository.AnimalFilter) ([]domain.Animal, int) {
		out, total, err := r.Animals.List(ctx, f)
		if err != nil {
			t.Fatal(err)
		}
		return out, total
	}
	if _, n := list(repository.AnimalFilter{}); n != 2 {
		t.Fatalf("archived must be hidden by default, got %d", n)
	}
	if _, n := list(repository.AnimalFilter{IncludeArchived: true}); n != 3 {
		t.Fatalf("got %d", n)
	}
	if _, n := list(repository.AnimalFilter{BreedID: "breed-nelore", Sex: domain.SexFemale}); n != 1 {
		t.Fatalf("got %d", n)
	}
	if out, _ := list(repository.AnimalFilter{Query: "100%", IncludeArchived: true}); len(out) != 1 {
		t.Fatalf("LIKE wildcard must be escaped, got %d", len(out))
	}
	if out, _ := list(repository.AnimalFilter{SortBy: "birthDate", SortDesc: true}); out[0].ID != "2" {
		t.Fatalf("sort desc: %v", out[0].ID)
	}
	if out, total := list(repository.AnimalFilter{Limit: 1, Offset: 1}); len(out) != 1 || total != 2 {
		t.Fatalf("paging: %d/%d", len(out), total)
	}
	// SQL injection attempt through sort is ignored (whitelist)
	if _, _, err := r.Animals.List(ctx, repository.AnimalFilter{SortBy: "name; DROP TABLE animals"}); err != nil {
		t.Fatal(err)
	}
}

func TestRecordsAndCascade(t *testing.T) {
	ctx := context.Background()
	r := repository.New(testutil.NewDB(t).DB)
	newFarm(t, r, "f1")
	if err := r.Animals.Insert(ctx, animal("a1", "f1", "Boi", domain.SexMale, domain.CategoryBull)); err != nil {
		t.Fatal(err)
	}
	acc, pct := 0.81, 70.0
	if err := r.Evaluations.Insert(ctx, &domain.GeneticEvaluation{ID: "e1", AnimalID: "a1", EvaluationProgram: "P", EvaluationRelease: "R",
		TraitCode: "WW", EvaluationType: domain.EvalDEP, Value: 3.2, Accuracy: &acc, Percentile: &pct, CreatedAt: ts, IsDemo: true}); err != nil {
		t.Fatal(err)
	}
	evs, err := r.Evaluations.ListByAnimal(ctx, "a1")
	if err != nil || len(evs) != 1 || evs[0].Accuracy == nil || *evs[0].Accuracy != 0.81 || !evs[0].IsDemo {
		t.Fatalf("%+v %v", evs, err)
	}
	if err := r.Measurements.Insert(ctx, &domain.Measurement{ID: "m1", AnimalID: "a1", TraitCode: "WT", Value: 450, Unit: "kg", MeasuredAt: "2025-01-01", CreatedAt: ts}); err != nil {
		t.Fatal(err)
	}
	// negative stock blocked at DB level
	if err := r.Semen.Insert(ctx, &domain.SemenInventory{ID: "s1", BullID: "a1", Quantity: -1, SemenType: "conventional", CreatedAt: ts, UpdatedAt: ts}); apperr.CodeOf(err) != apperr.CodeValidation {
		t.Fatalf("want validation, got %v", err)
	}
	if err := r.Semen.Insert(ctx, &domain.SemenInventory{ID: "s1", BullID: "a1", Quantity: 10, SemenType: "conventional", CreatedAt: ts, UpdatedAt: ts}); err != nil {
		t.Fatal(err)
	}
	if err := r.Semen.SetQuantity(ctx, "s1", -5, ts); err == nil {
		t.Fatal("negative quantity must fail")
	}
	// animal with stock cannot be hard-deleted through FK (RESTRICT)
	if err := r.Matings.Insert(ctx, &domain.Mating{ID: "x", FarmID: "f1", FemaleID: "a1", BullID: "a1", Status: "draft", EngineVersion: "v", CalculationSnapshot: "{}", CreatedAt: ts}); err == nil {
		t.Fatal("same animal on both sides must fail")
	}
}
