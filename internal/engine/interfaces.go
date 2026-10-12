package engine

import "genetica/internal/genetics"

// RunContext is the immutable, run-wide state shared by all stages.
type RunContext struct {
	Profile  Profile
	Female   Female
	Pedigree genetics.Graph
	// Bases are the reference bases chosen by the compatibility stage, by TraitWeight.Key().
	Bases map[string]*ReferenceBasis
}

// CalculationContext is the immutable input of every per-candidate stage:
// one female × one candidate within a run.
type CalculationContext struct {
	Run       RunContext
	Candidate Candidate
	// Selected holds, per TraitWeight.Key(), the candidate's evaluation on the
	// reference basis (nil when none exists). Filled by the compatibility stage.
	Selected map[string]*Evaluation
	// Incompatible holds, per TraitWeight.Key(), evaluations the candidate has
	// on bases other than the reference one.
	Incompatible map[string][]Evaluation
}

// Issue is a validation problem.
type Issue struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Stage 2 of the pipeline: Validation.
// CandidateValidator checks that the pair is well-formed (sexes, identity, …).
type CandidateValidator interface {
	Validate(run RunContext, c Candidate) []Issue
}

// RunValidator checks run-level input (profile, female). Errors abort the run.
type RunValidator interface {
	ValidateRun(in RunInput) []Issue
}

// Stage 3: Hard Filters. A filter either passes a candidate (nil) or excludes it with a reason.
type HardFilter interface {
	Name() string
	Apply(run RunContext, c Candidate) *Exclusion
}

// Stage 4: Trait Compatibility. Chooses, for every weighted trait, the single
// basis on which the eligible candidates are compared.
type BasisResolver interface {
	Resolve(run RunContext, eligible []Candidate) map[string]*ReferenceBasis
}

// SelectEvaluations picks each candidate's evaluation on the reference basis.
type EvaluationSelector interface {
	Select(run RunContext, c Candidate) (selected map[string]*Evaluation, incompatible map[string][]Evaluation, warnings []string)
}

// NormalizeInput asks a Normalizer for a 0..100 score of one value.
type NormalizeInput struct {
	Weight TraitWeight
	Eval   Evaluation
	Mode   string
}

type NormalizeResult struct {
	Score          *float64
	Method         string
	Transformation string
	Formula        string
	Reason         string // why no score could be produced
}

// Stage 5: Normalization. Turns a value into a 0..100 score using only
// documented, supported methods; otherwise it returns Reason, never a guess.
type Normalizer interface {
	Normalize(in NormalizeInput) NormalizeResult
}

// Stage 6: Weighted components. One ScoringComponent per weighted trait.
// Calculate returns the component with its normalized score; weights and
// contributions are applied afterwards by the ScoreAggregator because they
// depend on which components turned out to be calculable.
type ScoringComponent interface {
	Calculate(ctx CalculationContext) ComponentResult
}

// Stage 7: Complementarity (bonus points).
type ComplementarityScorer interface {
	Calculate(ctx CalculationContext, comps []ComponentResult) ComplementarityResult
}

// Stage 8: Penalties (points subtracted).
type PenaltyRule interface {
	Calculate(ctx CalculationContext, comps []ComponentResult) PenaltyResult
}

// Stage 9: Final score.
type ScoreAggregator interface {
	Aggregate(ctx CalculationContext, comps []ComponentResult, comp ComplementarityResult, pens []PenaltyResult) (final *FinalScore, weighted []ComponentResult, reasons []string)
}
