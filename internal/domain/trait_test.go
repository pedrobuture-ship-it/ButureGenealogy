package domain

import "testing"

func f(v float64) *float64 { return &v }
func b(v bool) *bool       { return &v }

func validTrait() Trait {
	return Trait{Code: "WW", Name: "Peso à desmama", Category: CatGrowth, Purpose: PurposeBeef, DataType: DataNumber,
		EvaluationType: EvalDEP, OptimizationDirection: InformationOnly, Metadata: "{}"}
}

func TestTraitValid(t *testing.T) {
	tr := validTrait()
	if err := tr.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestTraitRules(t *testing.T) {
	tests := []struct {
		name  string
		mod   func(*Trait)
		field string
	}{
		{"código inválido", func(t *Trait) { t.Code = "1 x" }, "code"},
		{"sem nome", func(t *Trait) { t.Name = " " }, "name"},
		{"categoria", func(t *Trait) { t.Category = "x" }, "category"},
		{"finalidade", func(t *Trait) { t.Purpose = "x" }, "purpose"},
		{"tipo de dado", func(t *Trait) { t.DataType = "x" }, "dataType"},
		{"tipo de avaliação", func(t *Trait) { t.EvaluationType = "BLUP" }, "evaluationType"},
		{"direção", func(t *Trait) { t.OptimizationDirection = "BEST" }, "optimizationDirection"},
		{"faixa sem limites", func(t *Trait) { t.OptimizationDirection = TargetRange }, "idealMin"},
		{"faixa só com mínimo", func(t *Trait) { t.OptimizationDirection = TargetRange; t.IdealMin = f(1) }, "idealMin"},
		{"limites invertidos", func(t *Trait) { t.IdealMin, t.IdealMax = f(5), f(1) }, "idealMin"},
		{"higher_is_better em info", func(t *Trait) { t.HigherIsBetter = b(true) }, "higherIsBetter"},
		{"maximize contradito", func(t *Trait) { t.OptimizationDirection = Maximize; t.HigherIsBetter = b(false) }, "higherIsBetter"},
		{"minimize contradito", func(t *Trait) { t.OptimizationDirection = Minimize; t.HigherIsBetter = b(true) }, "higherIsBetter"},
		{"faixa com higher_is_better", func(t *Trait) {
			t.OptimizationDirection = TargetRange
			t.IdealMin, t.IdealMax = f(1), f(2)
			t.HigherIsBetter = b(true)
		}, "higherIsBetter"},
		{"metadata não-objeto", func(t *Trait) { t.Metadata = "[1]" }, "metadata"},
		{"metadata inválido", func(t *Trait) { t.Metadata = "{" }, "metadata"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := validTrait()
			tt.mod(&tr)
			if _, ok := fieldsOf(t, tr.Validate())[tt.field]; !ok {
				t.Fatalf("expected error on %s", tt.field)
			}
		})
	}
}

func TestTraitDirectionsAccepted(t *testing.T) {
	cases := []Trait{
		func() Trait {
			x := validTrait()
			x.OptimizationDirection = Maximize
			x.HigherIsBetter = b(true)
			return x
		}(),
		func() Trait {
			x := validTrait()
			x.OptimizationDirection = Minimize
			x.HigherIsBetter = b(false)
			return x
		}(),
		func() Trait { x := validTrait(); x.OptimizationDirection = Maximize; return x }(),
		func() Trait {
			x := validTrait()
			x.OptimizationDirection = TargetRange
			x.IdealMin, x.IdealMax = f(30), f(40)
			return x
		}(),
	}
	for i, c := range cases {
		if err := c.Validate(); err != nil {
			t.Fatalf("case %d: %v", i, err)
		}
	}
	m := cases[2]
	if h := m.EffectiveHigherIsBetter(); h == nil || !*h {
		t.Fatal("MAXIMIZE => higher is better")
	}
	tr := cases[3]
	if tr.EffectiveHigherIsBetter() != nil {
		t.Fatal("TARGET_RANGE has no direction")
	}
}

func TestValueAndUnit(t *testing.T) {
	tr := validTrait()
	tr.Unit = "kg"
	if tr.ValidateValue(3.5) != nil {
		t.Fatal("number accepts decimals")
	}
	tr.DataType = DataScore
	if tr.ValidateValue(3.5) == nil || tr.ValidateValue(3) != nil {
		t.Fatal("score accepts only integers")
	}
	if u, err := tr.ResolveUnit(""); err != nil || u != "kg" {
		t.Fatalf("%q %v", u, err)
	}
	if _, err := tr.ResolveUnit("lb"); err == nil {
		t.Fatal("different unit must be rejected, not converted")
	}
	tr.Unit = ""
	if u, _ := tr.ResolveUnit("pts"); u != "pts" {
		t.Fatal("unit-less trait adopts the given unit")
	}
}

func TestKindsAreNotInterchangeable(t *testing.T) {
	if AreComparable(EvalDEP, EvalPTA) || AreComparable(EvalEBV, EvalEPD) || AreComparable(EvalPhenotype, EvalDEP) {
		t.Fatal("different kinds must not be comparable")
	}
	if !AreComparable(EvalDEP, EvalDEP) {
		t.Fatal("same kind is comparable (necessary, not sufficient)")
	}
	if len(KindInfo()) != len(ValueKinds) {
		t.Fatal("every kind needs a description")
	}
}
