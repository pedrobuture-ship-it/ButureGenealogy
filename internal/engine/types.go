// Package engine is the mating engine: a deterministic, auditable pipeline
//
//	Candidate → Validation → Hard Filters → Trait Compatibility → Normalization
//	          → Weighted Components → Complementarity → Penalties → Final Score
//
// It is pure (no SQL, no UI, no randomness, no clock except the one passed in)
// and never invents science: a quantity that cannot be computed with the data
// and methods explicitly supported is reported as "not calculable", never as a
// plausible-looking number. Every parameter used comes from the Profile or from
// a dataset and is copied into the snapshot. See docs/CALCULATION_ENGINE.md.
package engine

import (
	"fmt"
	"strings"
	"time"

	"genetica/internal/genetics"
)

// Version identifies the engine. Bump it whenever scoring behaviour changes.
const Version = "0.1.0"

// SnapshotVersion is the format version of saved snapshots.
const SnapshotVersion = 1

// IncompatibleBasesMessage is shown whenever evaluations cannot be compared.
const IncompatibleBasesMessage = "⚠ Comparação genética não suportada entre essas bases."

// DemoNotice must accompany every demonstration profile.
const DemoNotice = "DEMO — NÃO REPRESENTA RECOMENDAÇÃO ZOOTÉCNICA VALIDADA"

// Optimization directions (mirror domain constants; the engine has no domain dependency).
const (
	Maximize        = "MAXIMIZE"
	Minimize        = "MINIMIZE"
	TargetRange     = "TARGET_RANGE"
	InformationOnly = "INFORMATION_ONLY"
)

// Normalization modes.
const (
	NormAuto       = "auto"
	NormPercentile = "percentile"
	NormZScore     = "zscore"
)

// Percentile semantics a dataset must declare for percentiles to be usable.
const (
	PercentileAscending  = "ascending"  // a higher value has a higher percentile
	PercentileDescending = "descending" // a higher value has a lower percentile (e.g. "top 10%")
)

// Statuses.
const (
	StatusCalculated    = "calculated"
	StatusNotCalculable = "not_calculable"
	StatusNotApplied    = "not_applied"
	StatusExcluded      = "excluded"
)

// Basis is everything that must be equal for two evaluations to be comparable:
// the trait, the kind of value, the program, release, genetic base, breed and unit.
type Basis struct {
	TraitCode      string `json:"traitCode"`
	EvaluationType string `json:"evaluationType"`
	Program        string `json:"program"`
	Release        string `json:"release"`
	GeneticBase    string `json:"geneticBase"`
	BreedCode      string `json:"breedCode"`
	Unit           string `json:"unit"`
}

func (b Basis) Key() string {
	return strings.Join([]string{b.TraitCode, b.EvaluationType, b.Program, b.Release, b.GeneticBase, strings.ToUpper(b.BreedCode), b.Unit}, "|")
}

// Describe is a human-readable label of the basis.
func (b Basis) Describe() string {
	base := b.GeneticBase
	if base == "" {
		base = "base não informada"
	}
	return fmt.Sprintf("%s %s · %s %s · %s · %s", b.EvaluationType, b.TraitCode, b.Program, b.Release, base, b.BreedCode)
}

// TraitMeta is the part of a trait definition's metadata the engine understands.
type TraitMeta struct {
	PercentileSemantics string   `json:"percentileSemantics"` // ascending | descending | "" (unknown)
	PopulationMean      *float64 `json:"populationMean"`
	PopulationSD        *float64 `json:"populationSd"`
}

// Evaluation is one value of one trait of one animal, with full provenance.
type Evaluation struct {
	ID             string    `json:"id"`
	Basis          Basis     `json:"basis"`
	Value          float64   `json:"value"`
	Accuracy       *float64  `json:"accuracy"`   // fraction 0..1
	Percentile     *float64  `json:"percentile"` // 0..100
	DatasetID      string    `json:"datasetId"`
	DatasetName    string    `json:"datasetName"`
	DatasetVersion string    `json:"datasetVersion"`
	IsDemo         bool      `json:"isDemo"`
	CreatedAt      string    `json:"createdAt"`
	TraitName      string    `json:"traitName"`
	Direction      string    `json:"direction"` // direction defined by the trait
	IdealMin       *float64  `json:"idealMin"`
	IdealMax       *float64  `json:"idealMax"`
	Meta           TraitMeta `json:"meta"`
}

// Animal is the identity part of a female or candidate.
type Animal struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Registration string `json:"registration"`
	Sex          string `json:"sex"`
	Status       string `json:"status"`
	BreedName    string `json:"breedName"`
	IsDemo       bool   `json:"isDemo"`
}

func (a Animal) Label() string {
	switch {
	case a.Name != "" && a.Registration != "":
		return a.Name + " (" + a.Registration + ")"
	case a.Name != "":
		return a.Name
	}
	return a.Registration
}

type Female struct {
	Animal
	Evals []Evaluation `json:"evals"`
}

type Stock struct {
	Doses    int      `json:"doses"`
	UnitCost *float64 `json:"unitCost"`
}

type Candidate struct {
	Animal
	Evals []Evaluation `json:"evals"`
	Stock Stock        `json:"stock"`
}

// DatasetInfo identifies the exact data version used by a calculation.
type DatasetInfo struct {
	ID            string `json:"id"`
	PackageID     string `json:"packageId"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	Program       string `json:"program"`
	Release       string `json:"release"`
	SHA256        string `json:"sha256"`
	IsDemo        bool   `json:"isDemo"`
	Origin        string `json:"origin"`
	LicenseStatus string `json:"licenseStatus"`
}

// ---- profile

// TraitWeight is one weighted trait of a profile. Weights are relative: they
// are normalized over the components that can actually be calculated.
type TraitWeight struct {
	TraitCode       string   `json:"traitCode"`
	EvaluationType  string   `json:"evaluationType"`
	Weight          float64  `json:"weight"`
	Direction       string   `json:"direction"` // overrides the trait's own direction when set
	IdealMin        *float64 `json:"idealMin"`
	IdealMax        *float64 `json:"idealMax"`
	TargetTolerance *float64 `json:"targetTolerance"` // distance beyond the range at which the score reaches 0
	Required        bool     `json:"required"`        // candidates without any evaluation are excluded
}

func (t TraitWeight) Key() string { return t.TraitCode + "|" + t.EvaluationType }

type Filters struct {
	RequireStock           bool     `json:"requireStock"`
	MinDoses               int      `json:"minDoses"`
	MaxUnitCost            *float64 `json:"maxUnitCost"`
	MinAccuracy            *float64 `json:"minAccuracy"`            // fraction; an unknown accuracy fails when set
	ExcludeDirectAncestors bool     `json:"excludeDirectAncestors"` // exclude bulls that are ancestors of the dam
	AncestorGenerations    int      `json:"ancestorGenerations"`    // generations inspected for ancestors/common ancestors (default 3)
}

type ComplementarityConfig struct {
	MaxBonusPoints float64 `json:"maxBonusPoints"` // 0 disables complementarity
}

type LowAccuracyPenalty struct {
	Threshold float64 `json:"threshold"` // fraction; 0 disables
	MaxPoints float64 `json:"maxPoints"`
}

type PenaltyConfig struct {
	LowAccuracy LowAccuracyPenalty `json:"lowAccuracy"`
}

// Profile is a breeding objective. Everything the engine uses beyond the data
// itself is here, and is copied into every snapshot.
type Profile struct {
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	IsDemo          bool                  `json:"isDemo"`
	DemoNotice      string                `json:"demoNotice"`
	Description     string                `json:"description"`
	Traits          []TraitWeight         `json:"traits"`
	Filters         Filters               `json:"filters"`
	Normalization   string                `json:"normalization"`
	MinCoverage     float64               `json:"minCoverage"` // fraction of the weight that must be calculable (default 1)
	Complementarity ComplementarityConfig `json:"complementarity"`
	Penalties       PenaltyConfig         `json:"penalties"`
}

// normalized returns a copy with defaults applied.
func (p Profile) normalized() Profile {
	if p.Normalization == "" {
		p.Normalization = NormAuto
	}
	if p.MinCoverage <= 0 || p.MinCoverage > 1 {
		p.MinCoverage = 1
	}
	if p.Filters.AncestorGenerations <= 0 {
		p.Filters.AncestorGenerations = 3
	}
	return p
}

// ActiveTraits are the weighted traits (weight > 0), in profile order.
func (p Profile) ActiveTraits() []TraitWeight {
	var out []TraitWeight
	for _, t := range p.Traits {
		if t.Weight > 0 {
			out = append(out, t)
		}
	}
	return out
}

// ---- results

// Source says where a number came from: a dataset evaluation or a profile parameter.
type Source struct {
	EvaluationID   string `json:"evaluationId"`
	DatasetID      string `json:"datasetId"`
	DatasetName    string `json:"datasetName"`
	DatasetVersion string `json:"datasetVersion"`
	Program        string `json:"program"`
	Release        string `json:"release"`
	GeneticBase    string `json:"geneticBase"`
	IsDemo         bool   `json:"isDemo"`
	Parameter      string `json:"parameter"` // e.g. "perfil: peso do componente" for profile-supplied values
}

func sourceOf(e *Evaluation) Source {
	if e == nil {
		return Source{}
	}
	return Source{EvaluationID: e.ID, DatasetID: e.DatasetID, DatasetName: e.DatasetName, DatasetVersion: e.DatasetVersion,
		Program: e.Basis.Program, Release: e.Basis.Release, GeneticBase: e.Basis.GeneticBase, IsDemo: e.IsDemo}
}

// ComponentResult is one weighted trait component, fully explained.
type ComponentResult struct {
	Name           string   `json:"name"`
	TraitCode      string   `json:"traitCode"`
	EvaluationType string   `json:"evaluationType"`
	Status         string   `json:"status"`
	Direction      string   `json:"direction"`
	InputValue     *float64 `json:"inputValue"` // the bull's value
	DamValue       *float64 `json:"damValue"`   // the dam's value on the same basis, when available
	Unit           string   `json:"unit"`
	Accuracy       *float64 `json:"accuracy"`
	Percentile     *float64 `json:"percentile"`
	Score          *float64 `json:"score"`  // normalized, 0..100
	Method         string   `json:"method"` // percentile | zscore | target_range
	Transformation string   `json:"transformation"`
	Formula        string   `json:"formula"`
	RawWeight      float64  `json:"rawWeight"`    // as set in the profile
	Weight         float64  `json:"weight"`       // fraction actually used (normalized over calculable components)
	Contribution   *float64 `json:"contribution"` // score × weight
	Basis          *Basis   `json:"basis"`
	Source         Source   `json:"source"`
	Explanation    string   `json:"explanation"`
	Warnings       []string `json:"warnings"`
}

// ComplementarityItem is the per-trait part of the complementarity bonus.
type ComplementarityItem struct {
	TraitCode   string   `json:"traitCode"`
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	BullScore   *float64 `json:"bullScore"`
	DamScore    *float64 `json:"damScore"`
	Gain        *float64 `json:"gain"`   // clamp((bull−dam)/100, 0, 1)
	Weight      float64  `json:"weight"` // normalized share of the component
	Points      float64  `json:"points"`
	Explanation string   `json:"explanation"`
	Source      Source   `json:"source"` // the dam's evaluation
}

type ComplementarityResult struct {
	Status      string                `json:"status"`
	Points      float64               `json:"points"`
	MaxBonus    float64               `json:"maxBonus"`
	Formula     string                `json:"formula"`
	Items       []ComplementarityItem `json:"items"`
	Source      Source                `json:"source"`
	Explanation string                `json:"explanation"`
	Warnings    []string              `json:"warnings"`
}

// PenaltyResult is one penalty rule's outcome. Points are subtracted from the score.
type PenaltyResult struct {
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	Points      float64  `json:"points"`
	Formula     string   `json:"formula"`
	Source      Source   `json:"source"`
	Explanation string   `json:"explanation"`
	Warnings    []string `json:"warnings"`
}

type FinalScore struct {
	Base            float64 `json:"base"`
	Complementarity float64 `json:"complementarity"`
	Penalties       float64 `json:"penalties"`
	Unclamped       float64 `json:"unclamped"`
	Final           float64 `json:"final"`
	Clamped         bool    `json:"clamped"`
	Coverage        float64 `json:"coverage"` // fraction of the profile weight that was calculable
	Formula         string  `json:"formula"`
}

// Exclusion records why a hard filter removed a candidate.
type Exclusion struct {
	Filter string `json:"filter"`
	Reason string `json:"reason"`
}

// CandidateResult is the complete outcome for one female × bull pair.
type CandidateResult struct {
	Bull            Animal                `json:"bull"`
	Stock           Stock                 `json:"stock"`
	Status          string                `json:"status"`
	Rank            int                   `json:"rank"`
	Tied            bool                  `json:"tied"`
	Exclusions      []Exclusion           `json:"exclusions"`
	Reasons         []string              `json:"reasons"` // why the pair is not calculable
	Components      []ComponentResult     `json:"components"`
	Complementarity ComplementarityResult `json:"complementarity"`
	Penalties       []PenaltyResult       `json:"penalties"`
	Final           *FinalScore           `json:"final"`
	Warnings        []string              `json:"warnings"`
	PedigreeNotes   []string              `json:"pedigreeNotes"`
}

// ReferenceBasis is the evaluation basis chosen to compare candidates on one trait.
type ReferenceBasis struct {
	TraitCode      string       `json:"traitCode"`
	EvaluationType string       `json:"evaluationType"`
	Basis          *Basis       `json:"basis"`
	Candidates     int          `json:"candidates"` // eligible candidates evaluated on this basis
	Eligible       int          `json:"eligible"`   // eligible candidates in total
	DamMatches     bool         `json:"damMatches"`
	Reason         string       `json:"reason"`
	Alternatives   []BasisCount `json:"alternatives"`
}

type BasisCount struct {
	Basis      Basis `json:"basis"`
	Candidates int   `json:"candidates"`
}

// RunInput is everything a run needs; the engine never reads anything else.
type RunInput struct {
	Female     Female
	Candidates []Candidate
	Profile    Profile
	Pedigree   genetics.Graph
	Datasets   []DatasetInfo
	Now        time.Time
}

// RunOutput is the full result of ranking candidates for one female.
type RunOutput struct {
	EngineVersion string            `json:"engineVersion"`
	CreatedAt     string            `json:"createdAt"`
	Female        Animal            `json:"female"`
	Profile       Profile           `json:"profile"`
	Bases         []ReferenceBasis  `json:"bases"`
	Results       []CandidateResult `json:"results"` // calculated (ranked), then not calculable, then excluded
	Top           []string          `json:"top"`     // bull IDs of the best calculated candidates
	Datasets      []DatasetInfo     `json:"datasets"`
	Errors        []string          `json:"errors"` // run-level problems that prevented the run
	Warnings      []string          `json:"warnings"`
}

// PairSnapshot is the self-contained record saved for each female × bull pair.
// It carries everything needed to explain the number without consulting
// anything that may have changed since (datasets, profiles, the herd).
type PairSnapshot struct {
	SnapshotVersion int              `json:"snapshotVersion"`
	EngineVersion   string           `json:"engineVersion"`
	CreatedAt       string           `json:"createdAt"`
	Female          Animal           `json:"female"`
	FemaleEvals     []Evaluation     `json:"femaleEvals"`
	Bull            Animal           `json:"bull"`
	BullEvals       []Evaluation     `json:"bullEvals"`
	Profile         Profile          `json:"profile"`
	Bases           []ReferenceBasis `json:"bases"`
	Datasets        []DatasetInfo    `json:"datasets"`
	Result          CandidateResult  `json:"result"`
	Candidates      int              `json:"candidates"` // size of the run (for context of the rank)
	RunWarnings     []string         `json:"runWarnings"`
	IsDemo          bool             `json:"isDemo"`
}
