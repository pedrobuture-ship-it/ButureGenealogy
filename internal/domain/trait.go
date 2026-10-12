package domain

import (
	"math"
	"regexp"
	"strings"
)

// Trait categories.
const (
	CatGrowth       = "growth"
	CatReproduction = "reproduction"
	CatMaternal     = "maternal"
	CatCarcass      = "carcass"
	CatDairy        = "dairy"
	CatConformation = "conformation"
	CatHealth       = "health"
	CatTemperament  = "temperament"
	CatEfficiency   = "efficiency"
	CatOther        = "other"
)

// TraitCategories lists categories in display order.
var TraitCategories = []string{CatGrowth, CatReproduction, CatMaternal, CatCarcass, CatDairy, CatConformation, CatHealth, CatTemperament, CatEfficiency, CatOther}

// Optimization directions: what "better" means for a trait.
const (
	Maximize        = "MAXIMIZE"
	Minimize        = "MINIMIZE"
	TargetRange     = "TARGET_RANGE"
	InformationOnly = "INFORMATION_ONLY"
)

// EvalPhenotype is the value kind of a measured (non-genetic) trait.
const EvalPhenotype = "PHENOTYPE"

// ValueKinds lists every kind of value a trait may describe.
var ValueKinds = []string{EvalPhenotype, EvalDEP, EvalEPD, EvalEBV, EvalPTA, EvalIndex}

const (
	DataNumber  = "number"
	DataInteger = "integer"
	DataScore   = "score"
)

// ValueKindInfo explains a kind of value for the interface and documentation.
type ValueKindInfo struct {
	Kind        string `json:"kind"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Genetic     bool   `json:"genetic"`
}

// KindInfo describes each value kind. These are definitions, not conversions:
// values of different kinds (or different programs/bases) must never be mixed.
func KindInfo() []ValueKindInfo {
	return []ValueKindInfo{
		{EvalPhenotype, "Fenótipo", "Medida observada no animal. Mistura genética e ambiente; não é mérito genético.", false},
		{EvalDEP, "DEP", "Diferença Esperada na Progênie: valor genético publicado por um programa, relativo à base daquele programa.", true},
		{EvalEPD, "EPD", "Expected Progeny Difference: equivalente em inglês à DEP, relativo à base do programa que o publica.", true},
		{EvalEBV, "EBV", "Estimated Breeding Value: estimativa do valor genético aditivo do animal, relativo à base do programa.", true},
		{EvalPTA, "PTA", "Predicted Transmitting Ability: habilidade de transmissão prevista, relativa à base do programa.", true},
		{EvalIndex, "Índice", "Combinação ponderada de características definida pelo programa que o publica.", true},
	}
}

// AreComparable reports whether two values may be placed on the same scale
// without a documented conversion. Different kinds never are, and even equal
// kinds also require the same program, release, base, trait and unit (checked
// by the calculation engine).
func AreComparable(kindA, kindB string) bool { return kindA == kindB }

var traitCodeRE = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,63}$`)

// Trait is a generic characteristic definition.
type Trait struct {
	ID                    string   `json:"id"`
	Code                  string   `json:"code"`
	Name                  string   `json:"name"`
	Description           string   `json:"description"`
	Category              string   `json:"category"`
	Purpose               string   `json:"purpose"` // beef | dairy | dual | any
	Unit                  string   `json:"unit"`
	DataType              string   `json:"dataType"`
	EvaluationType        string   `json:"evaluationType"`
	OptimizationDirection string   `json:"optimizationDirection"`
	IdealMin              *float64 `json:"idealMin"`
	IdealMax              *float64 `json:"idealMax"`
	HigherIsBetter        *bool    `json:"higherIsBetter"`
	SourceDataset         string   `json:"sourceDataset"` // "" = built-in or user-defined
	Metadata              string   `json:"metadata"`      // JSON object
	IsBuiltin             bool     `json:"isBuiltin"`
	CreatedAt             string   `json:"createdAt"`
}

// Validate checks the definition in isolation.
func (t *Trait) Validate() error {
	var v validator
	if !traitCodeRE.MatchString(t.Code) {
		v.add("code", "Código inválido: use letras, números e _, começando por letra.")
	}
	required(&v, "name", t.Name, "Informe o nome da característica.")
	if !oneOf(t.Category, TraitCategories...) {
		v.add("category", "Categoria inválida.")
	}
	if !oneOf(t.Purpose, PurposeBeef, PurposeDairy, PurposeDual, "any") {
		v.add("purpose", "Finalidade inválida.")
	}
	if !oneOf(t.DataType, DataNumber, DataInteger, DataScore) {
		v.add("dataType", "Tipo de dado inválido.")
	}
	if !oneOf(t.EvaluationType, ValueKinds...) {
		v.add("evaluationType", "Tipo de avaliação inválido.")
	}
	if !oneOf(t.OptimizationDirection, Maximize, Minimize, TargetRange, InformationOnly) {
		v.add("optimizationDirection", "Direção de otimização inválida.")
	}
	for f, p := range map[string]*float64{"idealMin": t.IdealMin, "idealMax": t.IdealMax} {
		if p != nil && !finite(*p) {
			v.add(f, "Valor inválido.")
		}
	}
	if t.IdealMin != nil && t.IdealMax != nil && finite(*t.IdealMin) && finite(*t.IdealMax) && *t.IdealMin > *t.IdealMax {
		v.add("idealMin", "O mínimo ideal não pode ser maior que o máximo.")
	}
	switch t.OptimizationDirection {
	case TargetRange:
		if t.IdealMin == nil || t.IdealMax == nil {
			v.add("idealMin", "Faixa-alvo exige mínimo e máximo ideais.")
		}
		if t.HigherIsBetter != nil {
			v.add("higherIsBetter", "Não se aplica a faixa-alvo.")
		}
	case InformationOnly:
		if t.HigherIsBetter != nil {
			v.add("higherIsBetter", "Não se aplica a característica apenas informativa.")
		}
	case Maximize:
		if t.HigherIsBetter != nil && !*t.HigherIsBetter {
			v.add("higherIsBetter", "Contradiz a direção MAXIMIZE.")
		}
	case Minimize:
		if t.HigherIsBetter != nil && *t.HigherIsBetter {
			v.add("higherIsBetter", "Contradiz a direção MINIMIZE.")
		}
	}
	if !jsonObject(t.Metadata) {
		v.add("metadata", "Metadados devem ser um objeto JSON.")
	}
	return v.err()
}

// ValidateValue checks a value against the trait's data type.
func (t *Trait) ValidateValue(value float64) error {
	var v validator
	if !finite(value) {
		v.add("value", "Valor inválido.")
	} else if (t.DataType == DataInteger || t.DataType == DataScore) && value != math.Trunc(value) {
		v.add("value", "Esta característica aceita apenas valores inteiros.")
	}
	return v.err()
}

// ResolveUnit returns the unit to store for a value. Empty input adopts the
// trait's unit; a different unit is rejected (no silent conversion).
func (t *Trait) ResolveUnit(unit string) (string, error) {
	unit = strings.TrimSpace(unit)
	if unit == "" || unit == t.Unit {
		return t.Unit, nil
	}
	if t.Unit == "" {
		return unit, nil
	}
	return "", &ValidationError{Fields: map[string]string{"unit": "A unidade deve ser \"" + t.Unit + "\" para esta característica."}}
}

// EffectiveHigherIsBetter derives the direction for MAXIMIZE/MINIMIZE traits;
// it is nil for TARGET_RANGE and INFORMATION_ONLY.
func (t *Trait) EffectiveHigherIsBetter() *bool {
	var b bool
	switch t.OptimizationDirection {
	case Maximize:
		b = true
	case Minimize:
		b = false
	default:
		return nil
	}
	return &b
}
