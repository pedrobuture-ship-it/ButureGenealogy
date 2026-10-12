package service

// Input DTOs: what the UI may send. IDs and timestamps are assigned by the backend.

type FarmInput struct {
	Name      string `json:"name"`
	OwnerName string `json:"ownerName"`
	City      string `json:"city"`
	State     string `json:"state"`
	Notes     string `json:"notes"`
}

type AnimalInput struct {
	FarmID       string `json:"farmId"`
	Registration string `json:"registration"`
	Name         string `json:"name"`
	Sex          string `json:"sex"`
	BreedID      string `json:"breedId"`
	BirthDate    string `json:"birthDate"`
	Category     string `json:"category"`
	Status       string `json:"status"` // empty = active
	FatherID     string `json:"fatherId"`
	MotherID     string `json:"motherId"`
	Notes        string `json:"notes"`
}

type MeasurementInput struct {
	AnimalID   string  `json:"animalId"`
	TraitCode  string  `json:"traitCode"`
	Value      float64 `json:"value"`
	Unit       string  `json:"unit"`
	MeasuredAt string  `json:"measuredAt"`
	Source     string  `json:"source"`
	Notes      string  `json:"notes"`
}

type EvaluationInput struct {
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
}

// AnimalQuery is the listing request from the herd page.
type AnimalQuery struct {
	FarmID          string `json:"farmId"`
	BreedID         string `json:"breedId"`
	Sex             string `json:"sex"`
	Category        string `json:"category"`
	Status          string `json:"status"`
	IncludeArchived bool   `json:"includeArchived"`
	Search          string `json:"search"`
	MinAgeMonths    *int   `json:"minAgeMonths"`
	MaxAgeMonths    *int   `json:"maxAgeMonths"`
	SortBy          string `json:"sortBy"`
	SortDesc        bool   `json:"sortDesc"`
	Page            int    `json:"page"`     // 1-based
	PageSize        int    `json:"pageSize"` // default 50, max 500
}
