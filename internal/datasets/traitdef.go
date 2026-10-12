package datasets

import (
	"encoding/json"
	"strconv"
	"strings"

	"genetica/internal/domain"
)

// TraitDef is one entry of traits.json: a trait definition plus the data file
// columns that carry its value, accuracy and percentile.
type TraitDef struct {
	Code                  string         `json:"code"`
	Name                  string         `json:"name"`
	Description           string         `json:"description"`
	Category              string         `json:"category"`
	Purpose               string         `json:"purpose"`
	Unit                  string         `json:"unit"`
	DataType              string         `json:"dataType"`
	EvaluationType        string         `json:"evaluationType"`
	OptimizationDirection string         `json:"optimizationDirection"`
	IdealMin              *float64       `json:"idealMin"`
	IdealMax              *float64       `json:"idealMax"`
	HigherIsBetter        *bool          `json:"higherIsBetter"`
	Metadata              map[string]any `json:"metadata"`
	Column                string         `json:"column"`
	AccuracyColumn        string         `json:"accuracyColumn"`
	PercentileColumn      string         `json:"percentileColumn"`

	// ExistingTraitID, when set (interactive imports only), reuses a catalogue
	// trait instead of creating a dataset-owned definition.
	ExistingTraitID string `json:"existingTraitId,omitempty"`
}

// geneticKinds are the value kinds a dataset may provide.
var geneticKinds = map[string]bool{
	domain.EvalDEP: true, domain.EvalEPD: true, domain.EvalEBV: true, domain.EvalPTA: true, domain.EvalIndex: true,
}

// ToTrait builds the domain trait owned by the dataset row datasetID.
func (d TraitDef) ToTrait(datasetID, id, createdAt string) *domain.Trait {
	meta := "{}"
	if len(d.Metadata) > 0 {
		if b, err := json.Marshal(d.Metadata); err == nil {
			meta = string(b)
		}
	}
	t := &domain.Trait{
		ID: id, Code: strings.TrimSpace(d.Code), Name: strings.TrimSpace(d.Name), Description: d.Description,
		Category: orDefault(d.Category, domain.CatOther), Purpose: orDefault(d.Purpose, "any"), Unit: strings.TrimSpace(d.Unit),
		DataType: orDefault(d.DataType, domain.DataNumber), EvaluationType: d.EvaluationType,
		OptimizationDirection: orDefault(d.OptimizationDirection, domain.InformationOnly),
		IdealMin:              d.IdealMin, IdealMax: d.IdealMax, HigherIsBetter: d.HigherIsBetter,
		SourceDataset: datasetID, Metadata: meta, CreatedAt: createdAt,
	}
	return t
}

func orDefault(s, d string) string {
	if strings.TrimSpace(s) == "" {
		return d
	}
	return s
}

// ValidateDefs checks definitions in isolation (not against the data file).
func ValidateDefs(defs []TraitDef) error {
	f := fieldErrs{}
	if len(defs) == 0 {
		f.add("traits", "O pacote não define nenhuma característica.")
	}
	seen := map[string]bool{}
	cols := map[string]bool{}
	for i, d := range defs {
		key := "traits[" + strconv.Itoa(i) + "]"
		if !geneticKinds[d.EvaluationType] {
			f.add(key+".evaluationType", "Bibliotecas fornecem DEP, EPD, EBV, PTA ou INDEX (não fenótipo).")
		}
		if d.ExistingTraitID == "" {
			if err := d.ToTrait("x", "x", "x").Validate(); err != nil {
				if ve, ok := err.(*domain.ValidationError); ok {
					for k, v := range ve.Fields {
						f.add(key+"."+k, v)
					}
				}
			}
		}
		if strings.TrimSpace(d.Column) == "" {
			f.add(key+".column", "Informe a coluna do valor.")
		} else if cols[strings.ToLower(d.Column)] {
			f.add(key+".column", "Coluna usada por mais de uma característica.")
		}
		cols[strings.ToLower(d.Column)] = true
		k := strings.ToLower(d.Code) + "|" + d.EvaluationType
		if seen[k] {
			f.add(key+".code", "Característica repetida (mesmo código e tipo).")
		}
		seen[k] = true
	}
	return f.err()
}
