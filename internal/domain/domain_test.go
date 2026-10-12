package domain

import (
	"errors"
	"math"
	"testing"
	"time"
)

func fieldsOf(t *testing.T, err error) map[string]string {
	t.Helper()
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	return ve.Fields
}

func validAnimal() Animal {
	return Animal{ID: "a1", FarmID: "f1", Name: "Mimosa", Sex: SexFemale, Category: CategoryCow, Status: StatusActive, BirthDate: "2020-01-15"}
}

func TestAnimalValid(t *testing.T) {
	a := validAnimal()
	if err := a.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestAnimalRules(t *testing.T) {
	tests := []struct {
		name  string
		mod   func(*Animal)
		field string
	}{
		{"sem identificação", func(a *Animal) { a.Name = ""; a.Registration = "" }, "name"},
		{"sexo inválido", func(a *Animal) { a.Sex = "x" }, "sex"},
		{"categoria incompatível", func(a *Animal) { a.Category = CategoryBull }, "category"},
		{"macho vaca", func(a *Animal) { a.Sex = SexMale; a.Category = CategoryCow }, "category"},
		{"status inválido", func(a *Animal) { a.Status = "?" }, "status"},
		{"data inválida", func(a *Animal) { a.BirthDate = "15/01/2020" }, "birthDate"},
		{"data futura", func(a *Animal) { a.BirthDate = time.Now().AddDate(0, 0, 3).Format(DateLayout) }, "birthDate"},
		{"próprio pai", func(a *Animal) { a.FatherID = a.ID }, "fatherId"},
		{"própria mãe", func(a *Animal) { a.MotherID = a.ID }, "motherId"},
		{"pai igual mãe", func(a *Animal) { a.FatherID, a.MotherID = "x", "x" }, "motherId"},
		{"sem fazenda", func(a *Animal) { a.FarmID = "" }, "farmId"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := validAnimal()
			tt.mod(&a)
			if _, ok := fieldsOf(t, a.Validate())[tt.field]; !ok {
				t.Fatalf("expected error on %s", tt.field)
			}
		})
	}
}

func TestMeasurementAndEvaluation(t *testing.T) {
	m := Measurement{AnimalID: "a", TraitCode: "WW", Value: math.NaN(), MeasuredAt: "2024-01-01"}
	if _, ok := fieldsOf(t, m.Validate())["value"]; !ok {
		t.Fatal("NaN must be rejected")
	}
	bad := 1.5
	e := GeneticEvaluation{AnimalID: "a", TraitCode: "WW", EvaluationProgram: "P", EvaluationRelease: "R", EvaluationType: EvalDEP, Value: 1, Accuracy: &bad}
	if _, ok := fieldsOf(t, e.Validate())["accuracy"]; !ok {
		t.Fatal("accuracy > 1 must be rejected")
	}
	e.Accuracy = nil
	e.EvaluationType = "FOO"
	if _, ok := fieldsOf(t, e.Validate())["evaluationType"]; !ok {
		t.Fatal("unknown type must be rejected")
	}
	e.EvaluationType = EvalEBV
	e.EvaluationRelease = ""
	if _, ok := fieldsOf(t, e.Validate())["evaluationRelease"]; !ok {
		t.Fatal("release is required for provenance")
	}
}

func TestSemenAndMating(t *testing.T) {
	neg := -1.0
	s := SemenInventory{BullID: "b", Quantity: -1, UnitCost: &neg, SemenType: "x"}
	f := fieldsOf(t, s.Validate())
	if f["quantity"] == "" || f["unitCost"] == "" {
		t.Fatalf("got %v", f)
	}
	m := Mating{FarmID: "f", FemaleID: "x", BullID: "x", Status: MatingDraft, EngineVersion: "0.1.0", CalculationSnapshot: "{}"}
	if fieldsOf(t, m.Validate())["bullId"] == "" {
		t.Fatal("same animal on both sides must be rejected")
	}
	m.BullID = "y"
	m.CalculationSnapshot = "not json"
	if fieldsOf(t, m.Validate())["calculationSnapshot"] == "" {
		t.Fatal("invalid snapshot must be rejected")
	}
}

func TestFarm(t *testing.T) {
	f := Farm{Name: " "}
	if fieldsOf(t, f.Validate())["name"] == "" {
		t.Fatal("name required")
	}
	f = Farm{Name: "X", State: "SAO"}
	if fieldsOf(t, f.Validate())["state"] == "" {
		t.Fatal("UF must be 2 letters")
	}
}

func TestDateTimeZoneTolerance(t *testing.T) {
	old := Now
	defer func() { Now = old }()
	Now = func() time.Time { return time.Date(2026, 10, 11, 23, 30, 0, 0, time.UTC) }
	ok := validAnimal()
	ok.BirthDate = "2026-10-12" // "today" for someone ahead of UTC
	if err := ok.Validate(); err != nil {
		t.Fatalf("a date one day ahead of UTC must be accepted: %v", err)
	}
	bad := validAnimal()
	bad.BirthDate = "2026-10-13"
	if _, ok := fieldsOf(t, bad.Validate())["birthDate"]; !ok {
		t.Fatal("two days ahead is the future")
	}
}
