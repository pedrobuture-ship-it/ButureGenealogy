package domain

import "strings"

// Enumerations (string-valued, validated here and by CHECK constraints in SQL).
const (
	SexMale   = "male"
	SexFemale = "female"

	CategoryCalf   = "calf"
	CategoryHeifer = "heifer"
	CategoryCow    = "cow"
	CategorySteer  = "steer"
	CategoryBull   = "bull"
	CategoryOther  = "other"

	StatusActive   = "active"
	StatusArchived = "archived"
	StatusSold     = "sold"
	StatusDeceased = "deceased"

	PurposeBeef  = "beef"
	PurposeDairy = "dairy"
	PurposeDual  = "dual"

	// Evaluation types are never interchangeable (see docs/TRAITS.md).
	EvalDEP   = "DEP"
	EvalEPD   = "EPD"
	EvalEBV   = "EBV"
	EvalPTA   = "PTA"
	EvalIndex = "INDEX"

	MatingDraft     = "draft"
	MatingPlanned   = "planned"
	MatingConfirmed = "confirmed"
	MatingDone      = "done"
	MatingCancelled = "cancelled"

	SemenConventional = "conventional"
	SemenSexed        = "sexed"
)

// DemoMarker is the label required on every synthetic record.
const DemoMarker = "DEMO / SYNTHETIC"

type Farm struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	OwnerName string `json:"ownerName"`
	City      string `json:"city"`
	State     string `json:"state"`
	Notes     string `json:"notes"`
	IsDemo    bool   `json:"isDemo"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func (f *Farm) Validate() error {
	var v validator
	required(&v, "name", f.Name, "Informe o nome da propriedade.")
	if len(f.Name) > 200 {
		v.add("name", "Nome muito longo.")
	}
	if f.State != "" && len(strings.TrimSpace(f.State)) != 2 {
		v.add("state", "Use a sigla da UF com 2 letras.")
	}
	return v.err()
}

type Breed struct {
	ID      string `json:"id"`
	Code    string `json:"code"`
	Name    string `json:"name"`
	Species string `json:"species"`
	Purpose string `json:"purpose"`
	Active  bool   `json:"active"`
}

type Animal struct {
	ID           string `json:"id"`
	FarmID       string `json:"farmId"`
	Registration string `json:"registration"`
	Name         string `json:"name"`
	Sex          string `json:"sex"`
	BreedID      string `json:"breedId"` // empty = unknown
	BirthDate    string `json:"birthDate"`
	Category     string `json:"category"`
	Status       string `json:"status"`
	FatherID     string `json:"fatherId"`
	MotherID     string `json:"motherId"`
	Notes        string `json:"notes"`
	IsDemo       bool   `json:"isDemo"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// Validate checks the animal in isolation. Rules that need other rows
// (parent sex, existence, pedigree cycles) live in the service layer.
func (a *Animal) Validate() error {
	var v validator
	required(&v, "farmId", a.FarmID, "Selecione a propriedade.")
	if strings.TrimSpace(a.Name) == "" && strings.TrimSpace(a.Registration) == "" {
		v.add("name", "Informe o nome ou o registro do animal.")
	}
	if !oneOf(a.Sex, SexMale, SexFemale) {
		v.add("sex", "Sexo inválido.")
	}
	if !oneOf(a.Category, CategoryCalf, CategoryHeifer, CategoryCow, CategorySteer, CategoryBull, CategoryOther) {
		v.add("category", "Categoria inválida.")
	} else {
		if a.Sex == SexMale && oneOf(a.Category, CategoryHeifer, CategoryCow) {
			v.add("category", "Categoria incompatível com o sexo.")
		}
		if a.Sex == SexFemale && oneOf(a.Category, CategorySteer, CategoryBull) {
			v.add("category", "Categoria incompatível com o sexo.")
		}
	}
	if !oneOf(a.Status, StatusActive, StatusArchived, StatusSold, StatusDeceased) {
		v.add("status", "Status inválido.")
	}
	if a.BirthDate != "" {
		checkDate(&v, "birthDate", a.BirthDate, false)
	}
	if a.ID != "" {
		if a.FatherID == a.ID {
			v.add("fatherId", "Um animal não pode ser o próprio pai.")
		}
		if a.MotherID == a.ID {
			v.add("motherId", "Um animal não pode ser a própria mãe.")
		}
	}
	if a.FatherID != "" && a.FatherID == a.MotherID {
		v.add("motherId", "Pai e mãe não podem ser o mesmo animal.")
	}
	return v.err()
}

type Measurement struct {
	ID         string  `json:"id"`
	AnimalID   string  `json:"animalId"`
	TraitCode  string  `json:"traitCode"`
	Value      float64 `json:"value"`
	Unit       string  `json:"unit"`
	MeasuredAt string  `json:"measuredAt"`
	Source     string  `json:"source"`
	Notes      string  `json:"notes"`
	IsDemo     bool    `json:"isDemo"`
	CreatedAt  string  `json:"createdAt"`
}

func (m *Measurement) Validate() error {
	var v validator
	required(&v, "animalId", m.AnimalID, "Animal obrigatório.")
	required(&v, "traitCode", m.TraitCode, "Informe a característica.")
	if !finite(m.Value) {
		v.add("value", "Valor inválido.")
	}
	if m.MeasuredAt == "" {
		v.add("measuredAt", "Informe a data da medição.")
	} else {
		checkDate(&v, "measuredAt", m.MeasuredAt, false)
	}
	return v.err()
}

// GeneticEvaluation is one published value for one trait of one animal.
// Accuracy is a fraction in [0,1]; Percentile is in [0,100].
type GeneticEvaluation struct {
	ID                string   `json:"id"`
	AnimalID          string   `json:"animalId"`
	DatasetID         string   `json:"datasetId"`
	EvaluationProgram string   `json:"evaluationProgram"`
	EvaluationRelease string   `json:"evaluationRelease"`
	BreedCode         string   `json:"breedCode"`
	TraitCode         string   `json:"traitCode"`
	EvaluationType    string   `json:"evaluationType"`
	Value             float64  `json:"value"`
	Unit              string   `json:"unit"`
	Accuracy          *float64 `json:"accuracy"`
	Percentile        *float64 `json:"percentile"`
	GeneticBase       string   `json:"geneticBase"`
	SourceReference   string   `json:"sourceReference"`
	EvaluatedAt       string   `json:"evaluatedAt"`
	IsDemo            bool     `json:"isDemo"`
	CreatedAt         string   `json:"createdAt"`
}

func (e *GeneticEvaluation) Validate() error {
	var v validator
	required(&v, "animalId", e.AnimalID, "Animal obrigatório.")
	required(&v, "traitCode", e.TraitCode, "Informe a característica.")
	required(&v, "evaluationProgram", e.EvaluationProgram, "Informe o programa de avaliação.")
	required(&v, "evaluationRelease", e.EvaluationRelease, "Informe a versão (release) da avaliação.")
	if !oneOf(e.EvaluationType, EvalDEP, EvalEPD, EvalEBV, EvalPTA, EvalIndex) {
		v.add("evaluationType", "Tipo de avaliação inválido.")
	}
	if !finite(e.Value) {
		v.add("value", "Valor inválido.")
	}
	if e.Accuracy != nil && (!finite(*e.Accuracy) || *e.Accuracy < 0 || *e.Accuracy > 1) {
		v.add("accuracy", "A acurácia deve estar entre 0 e 1.")
	}
	if e.Percentile != nil && (!finite(*e.Percentile) || *e.Percentile < 0 || *e.Percentile > 100) {
		v.add("percentile", "O percentil deve estar entre 0 e 100.")
	}
	if e.EvaluatedAt != "" {
		checkDate(&v, "evaluatedAt", e.EvaluatedAt, false)
	}
	return v.err()
}

type SemenInventory struct {
	ID        string   `json:"id"`
	BullID    string   `json:"bullId"`
	Supplier  string   `json:"supplier"`
	Batch     string   `json:"batch"`
	Quantity  int      `json:"quantity"`
	UnitCost  *float64 `json:"unitCost"`
	SemenType string   `json:"semenType"`
	Notes     string   `json:"notes"`
	IsDemo    bool     `json:"isDemo"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
}

func (s *SemenInventory) Validate() error {
	var v validator
	required(&v, "bullId", s.BullID, "Selecione o touro.")
	if s.Quantity < 0 {
		v.add("quantity", "A quantidade não pode ser negativa.")
	}
	if s.UnitCost != nil && (!finite(*s.UnitCost) || *s.UnitCost < 0) {
		v.add("unitCost", "Custo inválido.")
	}
	required(&v, "semenType", s.SemenType, "Informe o tipo de sêmen.")
	return v.err()
}

type Mating struct {
	ID                  string   `json:"id"`
	FarmID              string   `json:"farmId"`
	FemaleID            string   `json:"femaleId"`
	BullID              string   `json:"bullId"`
	Status              string   `json:"status"`
	Score               *float64 `json:"score"`
	EngineVersion       string   `json:"engineVersion"`
	CalculationSnapshot string   `json:"calculationSnapshot"` // JSON document
	CreatedAt           string   `json:"createdAt"`
}

func (m *Mating) Validate() error {
	var v validator
	required(&v, "farmId", m.FarmID, "Propriedade obrigatória.")
	required(&v, "femaleId", m.FemaleID, "Selecione a matriz.")
	required(&v, "bullId", m.BullID, "Selecione o touro.")
	if m.FemaleID != "" && m.FemaleID == m.BullID {
		v.add("bullId", "Matriz e touro devem ser animais diferentes.")
	}
	if !oneOf(m.Status, MatingDraft, MatingPlanned, MatingConfirmed, MatingDone, MatingCancelled) {
		v.add("status", "Status inválido.")
	}
	required(&v, "engineVersion", m.EngineVersion, "Versão do motor obrigatória.")
	if m.Score != nil && !finite(*m.Score) {
		v.add("score", "Score inválido.")
	}
	if !jsonValid(m.CalculationSnapshot) {
		v.add("calculationSnapshot", "Snapshot do cálculo não é um JSON válido.")
	}
	return v.err()
}
