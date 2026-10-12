package engine

func fptr(v float64) *float64 { return &v }

const (
	ProfileCustomID    = "builtin:custom"
	ProfileDemoBeefID  = "builtin:demo-beef"
	ProfileDemoDairyID = "builtin:demo-dairy"
)

// CustomProfile is the empty "Personalizado" template: the user chooses the
// traits and weights; nothing is pre-filled except neutral, documented defaults.
func CustomProfile() Profile {
	return Profile{
		ID: ProfileCustomID, Name: "Personalizado",
		Description:   "Escolha as características, os pesos e as direções. Nada é pré-definido.",
		Normalization: NormAuto, MinCoverage: 1,
		Filters: Filters{RequireStock: true, AncestorGenerations: 3},
		Traits:  []TraitWeight{},
	}
}

// DemoBeefProfile is a demonstration profile. Its weights and directions are
// arbitrary and exist only to exercise the engine and the interface.
func DemoBeefProfile() Profile {
	return Profile{
		ID: ProfileDemoBeefID, Name: "Demonstração Corte", IsDemo: true, DemoNotice: DemoNotice,
		Description: DemoNotice + ". Pesos e direções arbitrários, apenas para demonstrar o motor.",
		Traits: []TraitWeight{
			{TraitCode: "WW", EvaluationType: "DEP", Weight: 40, Direction: Maximize},
			{TraitCode: "YW", EvaluationType: "DEP", Weight: 35, Direction: Maximize},
			{TraitCode: "MILK", EvaluationType: "DEP", Weight: 25, Direction: Maximize},
		},
		Normalization: NormAuto, MinCoverage: 1,
		Filters:         Filters{RequireStock: true, AncestorGenerations: 3},
		Complementarity: ComplementarityConfig{MaxBonusPoints: 10},
		Penalties:       PenaltyConfig{LowAccuracy: LowAccuracyPenalty{Threshold: 0.6, MaxPoints: 5}},
	}
}

// DemoDairyProfile is the dairy demonstration profile (also MINIMIZE for one trait).
func DemoDairyProfile() Profile {
	return Profile{
		ID: ProfileDemoDairyID, Name: "Demonstração Leite", IsDemo: true, DemoNotice: DemoNotice,
		Description: DemoNotice + ". Pesos e direções arbitrários, apenas para demonstrar o motor.",
		Traits: []TraitWeight{
			{TraitCode: "MY", EvaluationType: "PTA", Weight: 50, Direction: Maximize},
			{TraitCode: "FERT", EvaluationType: "PTA", Weight: 30, Direction: Maximize},
			{TraitCode: "SCS", EvaluationType: "PTA", Weight: 20, Direction: Minimize},
		},
		Normalization: NormAuto, MinCoverage: 1,
		Filters:         Filters{RequireStock: true, AncestorGenerations: 3},
		Complementarity: ComplementarityConfig{MaxBonusPoints: 8},
		Penalties:       PenaltyConfig{LowAccuracy: LowAccuracyPenalty{Threshold: 0.6, MaxPoints: 5}},
	}
}

// Builtins lists the profiles shipped with the application.
func Builtins() []Profile { return []Profile{CustomProfile(), DemoBeefProfile(), DemoDairyProfile()} }
