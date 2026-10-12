package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"genetica/internal/domain"
	"genetica/internal/testutil"
)

type fixture struct {
	svc    *Services
	farmID string
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	svc := New(testutil.NewDB(t), t.TempDir())
	f, err := svc.Farms.Create(context.Background(), FarmInput{Name: "Fazenda Teste", State: "mg"})
	if err != nil {
		t.Fatal(err)
	}
	if f.State != "MG" {
		t.Fatalf("state not normalized: %s", f.State)
	}
	return fixture{svc, f.ID}
}

func (fx fixture) animal(t *testing.T, name, sex, cat string, mod ...func(*AnimalInput)) *domain.Animal {
	t.Helper()
	in := AnimalInput{FarmID: fx.farmID, Name: name, Sex: sex, Category: cat}
	for _, m := range mod {
		m(&in)
	}
	a, err := fx.svc.Animals.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("create %s: %v", name, err)
	}
	return a
}

func wantField(t *testing.T, err error, field string) {
	t.Helper()
	var ve *domain.ValidationError
	if !errors.As(err, &ve) || ve.Fields[field] == "" {
		t.Fatalf("expected validation error on %q, got %v", field, err)
	}
}

func TestParentRules(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	bull := fx.animal(t, "Touro", domain.SexMale, domain.CategoryBull)
	cow := fx.animal(t, "Vaca", domain.SexFemale, domain.CategoryCow)

	calf := fx.animal(t, "Bezerra", domain.SexFemale, domain.CategoryCalf, func(i *AnimalInput) { i.FatherID, i.MotherID = bull.ID, cow.ID })
	if calf.FatherID != bull.ID {
		t.Fatal("parents not stored")
	}
	// wrong sex
	_, err := fx.svc.Animals.Create(ctx, AnimalInput{FarmID: fx.farmID, Name: "X", Sex: domain.SexMale, Category: domain.CategoryCalf, FatherID: cow.ID})
	wantField(t, err, "fatherId")
	_, err = fx.svc.Animals.Create(ctx, AnimalInput{FarmID: fx.farmID, Name: "X", Sex: domain.SexMale, Category: domain.CategoryCalf, MotherID: bull.ID})
	wantField(t, err, "motherId")
	// unknown parent
	_, err = fx.svc.Animals.Create(ctx, AnimalInput{FarmID: fx.farmID, Name: "X", Sex: domain.SexMale, Category: domain.CategoryCalf, FatherID: "zzz"})
	wantField(t, err, "fatherId")
	// self parent on update
	_, err = fx.svc.Animals.Update(ctx, bull.ID, AnimalInput{FarmID: fx.farmID, Name: "Touro", Sex: domain.SexMale, Category: domain.CategoryBull, FatherID: bull.ID})
	wantField(t, err, "fatherId")
}

func TestPedigreeCycleBlocked(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	grand := fx.animal(t, "Avô", domain.SexMale, domain.CategoryBull)
	father := fx.animal(t, "Pai", domain.SexMale, domain.CategoryBull, func(i *AnimalInput) { i.FatherID = grand.ID })
	son := fx.animal(t, "Filho", domain.SexMale, domain.CategoryBull, func(i *AnimalInput) { i.FatherID = father.ID })
	// make the grandfather a son of his own grandson
	_, err := fx.svc.Animals.Update(ctx, grand.ID, AnimalInput{FarmID: fx.farmID, Name: "Avô", Sex: domain.SexMale, Category: domain.CategoryBull, FatherID: son.ID})
	wantField(t, err, "fatherId")
}

func TestBirthOrderAndArchive(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	dam := fx.animal(t, "Mãe", domain.SexFemale, domain.CategoryCow, func(i *AnimalInput) { i.BirthDate = "2020-01-01" })
	_, err := fx.svc.Animals.Create(ctx, AnimalInput{FarmID: fx.farmID, Name: "Filha", Sex: domain.SexFemale, Category: domain.CategoryCalf, BirthDate: "2019-01-01", MotherID: dam.ID})
	wantField(t, err, "motherId")

	if err := fx.svc.Animals.Archive(ctx, dam.ID); err != nil {
		t.Fatal(err)
	}
	page, _ := fx.svc.Animals.List(ctx, AnimalQuery{})
	if page.Total != 0 {
		t.Fatalf("archived animal should be hidden, got %d", page.Total)
	}
	page, _ = fx.svc.Animals.List(ctx, AnimalQuery{IncludeArchived: true})
	if page.Total != 1 || page.Items[0].Status != domain.StatusArchived {
		t.Fatalf("archived animal must still exist: %+v", page)
	}
	if err := fx.svc.Animals.Restore(ctx, dam.ID); err != nil {
		t.Fatal(err)
	}
}

func TestUnknownFarmAndBreed(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	_, err := fx.svc.Animals.Create(ctx, AnimalInput{FarmID: "nope", Name: "X", Sex: domain.SexMale, Category: domain.CategoryBull})
	wantField(t, err, "farmId")
	_, err = fx.svc.Animals.Create(ctx, AnimalInput{FarmID: fx.farmID, Name: "X", Sex: domain.SexMale, Category: domain.CategoryBull, BreedID: "nope"})
	wantField(t, err, "breedId")
}

func TestAgeFilter(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.animal(t, "Velha", domain.SexFemale, domain.CategoryCow, func(i *AnimalInput) { i.BirthDate = "2015-01-01" })
	fx.animal(t, "Nova", domain.SexFemale, domain.CategoryHeifer, func(i *AnimalInput) { i.BirthDate = domain.Now().AddDate(0, -6, 0).Format(domain.DateLayout) })
	max := 12
	p, _ := fx.svc.Animals.List(ctx, AnimalQuery{MaxAgeMonths: &max})
	if p.Total != 1 || p.Items[0].Name != "Nova" {
		t.Fatalf("%+v", p)
	}
	min := 24
	p, _ = fx.svc.Animals.List(ctx, AnimalQuery{MinAgeMonths: &min})
	if p.Total != 1 || p.Items[0].Name != "Velha" {
		t.Fatalf("%+v", p)
	}
}

func TestSemenAndMating(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	bull := fx.animal(t, "Touro", domain.SexMale, domain.CategoryBull)
	cow := fx.animal(t, "Vaca", domain.SexFemale, domain.CategoryCow)

	if _, err := fx.svc.Semen.Add(ctx, SemenInput{BullID: cow.ID, Quantity: 5}); err == nil {
		t.Fatal("semen for a female must fail")
	}
	if _, err := fx.svc.Semen.Add(ctx, SemenInput{BullID: bull.ID, Quantity: -1}); err == nil {
		t.Fatal("negative stock must fail")
	}
	inv, err := fx.svc.Semen.Add(ctx, SemenInput{BullID: bull.ID, Quantity: 20, Supplier: "X"})
	if err != nil || inv.SemenType != domain.SemenConventional {
		t.Fatalf("%v %v", inv, err)
	}

	m := domain.Mating{FarmID: fx.farmID, FemaleID: cow.ID, BullID: bull.ID, EngineVersion: "0.1.0", CalculationSnapshot: `{"ok":true}`}
	saved, err := fx.svc.Matings.Save(ctx, m)
	if err != nil || saved.Status != domain.MatingDraft {
		t.Fatalf("%v %v", saved, err)
	}
	swapped := domain.Mating{FarmID: fx.farmID, FemaleID: bull.ID, BullID: cow.ID, EngineVersion: "0.1.0", CalculationSnapshot: `{}`}
	if _, err := fx.svc.Matings.Save(ctx, swapped); err == nil {
		t.Fatal("swapped sexes must fail")
	}
}

func TestEvaluationAndMeasurement(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	a := fx.animal(t, "Touro", domain.SexMale, domain.CategoryBull)
	acc := 0.9
	e, err := fx.svc.Evaluations.Add(ctx, EvaluationInput{AnimalID: a.ID, EvaluationProgram: "P", EvaluationRelease: "2026.1", TraitCode: "WW", EvaluationType: domain.EvalDEP, Value: 4, Accuracy: &acc})
	if err != nil {
		t.Fatal(err)
	}
	_, err = fx.svc.Evaluations.Add(ctx, EvaluationInput{AnimalID: "nope", EvaluationProgram: "P", EvaluationRelease: "R", TraitCode: "WW", EvaluationType: domain.EvalDEP})
	wantField(t, err, "animalId")
	if _, err := fx.svc.Measurements.Add(ctx, MeasurementInput{AnimalID: a.ID, TraitCode: "BODY_WEIGHT", Value: 480, Unit: "kg", MeasuredAt: "2025-03-01"}); err != nil {
		t.Fatal(err)
	}
	list, _ := fx.svc.Evaluations.ListByAnimal(ctx, a.ID)
	if len(list) != 1 || list[0].ID != e.ID {
		t.Fatalf("%+v", list)
	}
}

func TestDemoLoadAndRemove(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t) // real farm, must survive demo removal
	real := fx.animal(t, "Real", domain.SexFemale, domain.CategoryCow)

	if err := fx.svc.Demo.Load(ctx); err != nil {
		t.Fatal(err)
	}
	if err := fx.svc.Demo.Load(ctx); err == nil {
		t.Fatal("second load must be refused")
	}
	farms, _ := fx.svc.Farms.List(ctx)
	demo := 0
	for _, f := range farms {
		if f.IsDemo {
			demo++
			if !strings.Contains(f.Name, domain.DemoMarker) {
				t.Fatalf("demo farm without marker: %s", f.Name)
			}
		}
	}
	if demo != 1 {
		t.Fatalf("want 1 demo farm, got %d", demo)
	}
	all, _ := fx.svc.Animals.List(ctx, AnimalQuery{})
	for _, a := range all.Items {
		if a.IsDemo && !strings.Contains(a.Notes, domain.DemoMarker) {
			t.Fatalf("demo animal without marker: %s", a.Name)
		}
	}
	libs, _ := fx.svc.Libraries.List(ctx)
	if len(libs) != 2 || !libs[0].IsDemo || !strings.Contains(libs[0].Name, domain.DemoMarker) || !libs[1].IsDemo {
		t.Fatalf("demo evaluations must belong to a DEMO-marked library: %+v", libs)
	}
	if err := fx.svc.Demo.Remove(ctx); err != nil {
		t.Fatal(err)
	}
	if libs, _ = fx.svc.Libraries.List(ctx); len(libs) != 0 {
		t.Fatalf("demo library must be removed with the demo data: %+v", libs)
	}
	all, _ = fx.svc.Animals.List(ctx, AnimalQuery{})
	if all.Total != 1 || all.Items[0].ID != real.ID {
		t.Fatalf("real data must survive: %+v", all)
	}
}

func TestAnimalDetail(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	bull := fx.animal(t, "Touro", domain.SexMale, domain.CategoryBull, func(i *AnimalInput) { i.BreedID = "breed-nelore" })
	cow := fx.animal(t, "Vaca", domain.SexFemale, domain.CategoryCow)
	calf := fx.animal(t, "Bezerro", domain.SexMale, domain.CategoryCalf, func(i *AnimalInput) { i.FatherID, i.MotherID = bull.ID, cow.ID })
	if _, err := fx.svc.Semen.Add(ctx, SemenInput{BullID: bull.ID, Quantity: 7}); err != nil {
		t.Fatal(err)
	}
	d, err := fx.svc.Animals.Detail(ctx, bull.ID)
	if err != nil || d.BreedName != "Nelore" || len(d.Offspring) != 1 || d.Offspring[0].ID != calf.ID || d.SemenDoses != 7 {
		t.Fatalf("%+v %v", d, err)
	}
	d, _ = fx.svc.Animals.Detail(ctx, calf.ID)
	if d.Father == nil || d.Father.ID != bull.ID || d.Mother == nil || d.Mother.ID != cow.ID {
		t.Fatalf("parents not resolved: %+v", d)
	}
}

func TestTraitCatalogueAndValidation(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	a := fx.animal(t, "Touro", domain.SexMale, domain.CategoryBull)

	list, err := fx.svc.Traits.List(ctx, TraitQuery{Kind: domain.EvalDEP})
	if err != nil || len(list) == 0 {
		t.Fatalf("%v %v", list, err)
	}
	for _, tr := range list {
		if tr.OptimizationDirection != domain.InformationOnly {
			t.Fatalf("built-in traits must not impose a direction: %+v", tr)
		}
	}
	// measurement: unknown trait, wrong unit, default unit
	_, err = fx.svc.Measurements.Add(ctx, MeasurementInput{AnimalID: a.ID, TraitCode: "NOPE", Value: 1, MeasuredAt: "2025-01-01"})
	wantField(t, err, "traitCode")
	_, err = fx.svc.Measurements.Add(ctx, MeasurementInput{AnimalID: a.ID, TraitCode: "BODY_WEIGHT", Value: 500, Unit: "lb", MeasuredAt: "2025-01-01"})
	wantField(t, err, "unit")
	m, err := fx.svc.Measurements.Add(ctx, MeasurementInput{AnimalID: a.ID, TraitCode: "BODY_WEIGHT", Value: 500, MeasuredAt: "2025-01-01"})
	if err != nil || m.Unit != "kg" {
		t.Fatalf("unit should default to the trait unit: %+v %v", m, err)
	}
	_, err = fx.svc.Measurements.Add(ctx, MeasurementInput{AnimalID: a.ID, TraitCode: "BCS", Value: 3.5, MeasuredAt: "2025-01-01"})
	wantField(t, err, "value")
	// a DEP code does not exist as a phenotype, and vice versa
	_, err = fx.svc.Measurements.Add(ctx, MeasurementInput{AnimalID: a.ID, TraitCode: "WW", Value: 1, MeasuredAt: "2025-01-01"})
	wantField(t, err, "traitCode")
	_, err = fx.svc.Evaluations.Add(ctx, EvaluationInput{AnimalID: a.ID, EvaluationProgram: "P", EvaluationRelease: "R", TraitCode: "BODY_WEIGHT", EvaluationType: domain.EvalDEP, Value: 1})
	wantField(t, err, "traitCode")
	// the same code is not defined for PTA => not silently accepted as DEP
	_, err = fx.svc.Evaluations.Add(ctx, EvaluationInput{AnimalID: a.ID, EvaluationProgram: "P", EvaluationRelease: "R", TraitCode: "WW", EvaluationType: domain.EvalPTA, Value: 1})
	wantField(t, err, "traitCode")

	// custom trait
	_, err = fx.svc.Traits.Create(ctx, TraitInput{Code: "x y", Name: "Ruim", Category: domain.CatOther})
	wantField(t, err, "code")
	tr, err := fx.svc.Traits.Create(ctx, TraitInput{Code: "REA", Name: "Área de olho de lombo", Category: domain.CatCarcass, Unit: "cm²", EvaluationType: domain.EvalDEP})
	if err != nil || tr.IsBuiltin || tr.OptimizationDirection != domain.InformationOnly {
		t.Fatalf("%+v %v", tr, err)
	}
	if _, err := fx.svc.Traits.Create(ctx, TraitInput{Code: "REA", Name: "Dup", Category: domain.CatCarcass, EvaluationType: domain.EvalDEP}); err == nil {
		t.Fatal("duplicate (code, type) must be rejected")
	}
	// same code with another kind is a different trait
	if _, err := fx.svc.Traits.Create(ctx, TraitInput{Code: "REA", Name: "Área de olho de lombo (PTA)", Category: domain.CatCarcass, EvaluationType: domain.EvalPTA}); err != nil {
		t.Fatal(err)
	}
}

func TestAnimalTraitViewGroupsAndKeepsKindsApart(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	a := fx.animal(t, "Touro", domain.SexMale, domain.CategoryBull)
	acc, pct := 0.8, 70.0
	add := func(typ string) {
		if _, err := fx.svc.Evaluations.Add(ctx, EvaluationInput{AnimalID: a.ID, EvaluationProgram: "Prog", EvaluationRelease: "2026", TraitCode: "WW",
			EvaluationType: typ, Value: 5, Accuracy: &acc, Percentile: &pct, GeneticBase: "base-2020"}); err != nil {
			t.Fatal(err)
		}
	}
	add(domain.EvalDEP)
	if _, err := fx.svc.Traits.Create(ctx, TraitInput{Code: "WW", Name: "Peso à desmama (EBV)", Category: domain.CatGrowth, Unit: "kg", EvaluationType: domain.EvalEBV}); err != nil {
		t.Fatal(err)
	}
	add(domain.EvalEBV)
	if _, err := fx.svc.Measurements.Add(ctx, MeasurementInput{AnimalID: a.ID, TraitCode: "BODY_WEIGHT", Value: 800, MeasuredAt: "2025-05-01", Source: "balança"}); err != nil {
		t.Fatal(err)
	}
	if _, err := fx.svc.Measurements.Add(ctx, MeasurementInput{AnimalID: a.ID, TraitCode: "SCROTAL_CIRC", Value: 36, MeasuredAt: "2025-05-01"}); err != nil {
		t.Fatal(err)
	}

	groups, err := fx.svc.Traits.ForAnimal(ctx, a.ID)
	if err != nil {
		t.Fatal(err)
	}
	cats := []string{}
	for _, g := range groups {
		cats = append(cats, g.Category)
	}
	if len(cats) != 2 || cats[0] != domain.CatGrowth || cats[1] != domain.CatReproduction {
		t.Fatalf("groups in display order: %v", cats)
	}
	growth := groups[0].Items
	kinds := map[string]bool{}
	for _, it := range growth {
		kinds[it.ValueKind] = true
		if it.ValueKind == domain.EvalDEP || it.ValueKind == domain.EvalEBV {
			if !it.Provenance.Genetic || it.Provenance.Program != "Prog" || it.Provenance.GeneticBase != "base-2020" || it.Accuracy == nil {
				t.Fatalf("provenance incomplete: %+v", it.Provenance)
			}
		}
	}
	if len(growth) != 3 || !kinds[domain.EvalDEP] || !kinds[domain.EvalEBV] || !kinds[domain.EvalPhenotype] {
		t.Fatalf("DEP, EBV and phenotype must be separate items: %+v", growth)
	}
	if _, err := fx.svc.Traits.ForAnimal(ctx, "nope"); err == nil {
		t.Fatal("unknown animal")
	}
}
