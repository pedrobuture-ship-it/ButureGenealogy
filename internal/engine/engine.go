package engine

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"genetica/internal/genetics"
)

// TopN is how many ranked candidates are highlighted as "Top".
const TopN = 5

// Engine wires one implementation per pipeline stage. Every field is an
// interface, so a stage can be replaced or extended without touching the others.
type Engine struct {
	RunValidator       RunValidator
	CandidateValidator CandidateValidator
	Filters            []HardFilter
	BasisResolver      BasisResolver
	Selector           EvaluationSelector
	Normalizer         Normalizer
	NewComponent       func(tw TraitWeight, n Normalizer) ScoringComponent
	Complementarity    ComplementarityScorer
	Penalties          []PenaltyRule
	Aggregator         ScoreAggregator
}

// New returns the default engine.
func New() *Engine {
	n := DefaultNormalizer{}
	return &Engine{
		RunValidator:       DefaultRunValidator{},
		CandidateValidator: DefaultCandidateValidator{},
		Filters:            DefaultFilters(),
		BasisResolver:      DefaultBasisResolver{},
		Selector:           DefaultEvaluationSelector{},
		Normalizer:         n,
		NewComponent:       func(tw TraitWeight, n Normalizer) ScoringComponent { return TraitComponent{Weight: tw, Normalizer: n} },
		Complementarity:    DefaultComplementarity{Normalizer: n},
		Penalties:          []PenaltyRule{LowAccuracyRule{}, InbreedingRule{}},
		Aggregator:         DefaultAggregator{},
	}
}

func label(a Animal) string { return a.Label() }

// Run executes the whole pipeline for one female and her candidate bulls.
func (e *Engine) Run(in RunInput) RunOutput {
	profile := in.Profile.normalized()
	in.Profile = profile
	now := in.Now
	out := RunOutput{
		EngineVersion: Version, CreatedAt: now.UTC().Format(time.RFC3339), Female: in.Female.Animal, Profile: profile,
		Bases: []ReferenceBasis{}, Results: []CandidateResult{}, Top: []string{}, Datasets: []DatasetInfo{}, Errors: []string{}, Warnings: []string{},
	}

	// Stage 1–2: run and candidate validation.
	if issues := e.RunValidator.ValidateRun(in); len(issues) > 0 {
		for _, i := range issues {
			out.Errors = append(out.Errors, i.Message)
		}
		return out
	}
	run := RunContext{Profile: profile, Female: in.Female, Pedigree: in.Pedigree}

	var eligible []Candidate
	results := map[string]*CandidateResult{}
	order := make([]string, 0, len(in.Candidates))
	for _, c := range in.Candidates {
		r := &CandidateResult{Bull: c.Animal, Stock: c.Stock, Exclusions: []Exclusion{}, Reasons: []string{}, Components: []ComponentResult{},
			Penalties: []PenaltyResult{}, Warnings: []string{}, PedigreeNotes: []string{}}
		results[c.ID] = r
		order = append(order, c.ID)

		// Stage 2: validation.
		if issues := e.CandidateValidator.Validate(run, c); len(issues) > 0 {
			for _, i := range issues {
				r.Exclusions = append(r.Exclusions, Exclusion{"validação", i.Message})
			}
		}
		// Stage 3: hard filters (all are applied so every reason is shown).
		if len(r.Exclusions) == 0 {
			for _, f := range e.Filters {
				if ex := f.Apply(run, c); ex != nil {
					r.Exclusions = append(r.Exclusions, *ex)
				}
			}
		}
		if len(r.Exclusions) > 0 {
			r.Status = StatusExcluded
			continue
		}
		eligible = append(eligible, c)
	}

	// Stage 4: trait compatibility — one reference basis per trait.
	run.Bases = e.BasisResolver.Resolve(run, eligible)
	for _, t := range profile.ActiveTraits() {
		if rb := run.Bases[t.Key()]; rb != nil {
			out.Bases = append(out.Bases, *rb)
			if rb.Basis != nil && !rb.DamMatches && len(in.Female.Evals) > 0 {
				hasTrait := false
				for _, ev := range in.Female.Evals {
					if ev.Basis.TraitCode == t.TraitCode && ev.Basis.EvaluationType == t.EvaluationType {
						hasTrait = true
					}
				}
				if hasTrait {
					out.Warnings = append(out.Warnings, fmt.Sprintf("%s: a matriz foi avaliada em base diferente da usada para comparar os touros. %s", t.TraitCode, IncompatibleBasesMessage))
				}
			}
		}
	}

	// Stages 5–9 for each eligible candidate.
	for _, c := range eligible {
		r := results[c.ID]
		ctx := CalculationContext{Run: run, Candidate: c}
		var selWarnings []string
		ctx.Selected, ctx.Incompatible, selWarnings = e.Selector.Select(run, c)
		r.Warnings = append(r.Warnings, selWarnings...)

		comps := make([]ComponentResult, 0, len(profile.ActiveTraits()))
		for _, tw := range profile.ActiveTraits() {
			comps = append(comps, e.NewComponent(tw, e.Normalizer).Calculate(ctx))
		}
		comp := e.Complementarity.Calculate(ctx, comps)
		pens := make([]PenaltyResult, 0, len(e.Penalties))
		for _, p := range e.Penalties {
			pens = append(pens, p.Calculate(ctx, comps))
		}
		final, weighted, reasons := e.Aggregator.Aggregate(ctx, comps, comp, pens)
		r.Components = weighted
		r.Pedigree(in, run)
		if final == nil {
			r.Status, r.Reasons = StatusNotCalculable, reasons
			comp.Status, comp.Points = StatusNotCalculable, 0
			comp.Explanation = "Não aplicada: o score final deste touro não é calculável."
			for i := range pens {
				pens[i].Points = 0
				if pens[i].Status == StatusCalculated {
					pens[i].Status = StatusNotApplied
				}
			}
		} else {
			r.Status, r.Final = StatusCalculated, final
		}
		r.Complementarity, r.Penalties = comp, pens
		for _, w := range weighted {
			for _, msg := range w.Warnings {
				if msg == IncompatibleBasesMessage {
					r.Warnings = appendUnique(r.Warnings, fmt.Sprintf("%s: %s", w.TraitCode, msg))
				}
			}
		}
	}

	out.Results = rank(results, order)
	for _, r := range out.Results {
		if r.Status == StatusCalculated && len(out.Top) < TopN {
			out.Top = append(out.Top, r.Bull.ID)
		}
	}
	out.Datasets = usedDatasets(in, out)
	return out
}

func appendUnique(list []string, s string) []string {
	for _, x := range list {
		if x == s {
			return list
		}
	}
	return append(list, s)
}

// Pedigree adds common-ancestor notes. The coefficient of inbreeding is NOT
// calculated; relatedness found here is only reported as a warning.
func (r *CandidateResult) Pedigree(in RunInput, run RunContext) {
	if in.Pedigree == nil {
		return
	}
	gens := run.Profile.Filters.AncestorGenerations
	for _, ca := range genetics.CommonAncestors(in.Pedigree, in.Female.ID, r.Bull.ID, gens) {
		name := ca.ID
		if n := in.Pedigree[ca.ID]; n != nil {
			name = n.Name
			if name == "" {
				name = n.Registration
			}
		}
		note := fmt.Sprintf("Ancestral comum: %s (geração %s da matriz, %s do touro).", name, joinInts(ca.GenA), joinInts(ca.GenB))
		r.PedigreeNotes = append(r.PedigreeNotes, note)
	}
	if len(r.PedigreeNotes) > 0 {
		r.Warnings = append(r.Warnings, "Parentesco detectado no pedigree (coeficiente de endogamia não calculado nesta versão): "+strings.Join(r.PedigreeNotes, " "))
	}
}

func joinInts(a []int) string {
	parts := make([]string, len(a))
	for i, v := range a {
		parts[i] = fmt.Sprint(v)
	}
	return strings.Join(parts, "/")
}

// rank orders results: calculated by score (desc), then not calculable, then excluded.
func rank(results map[string]*CandidateResult, order []string) []CandidateResult {
	var calc, notCalc, excl []*CandidateResult
	for _, id := range order {
		r := results[id]
		switch r.Status {
		case StatusCalculated:
			calc = append(calc, r)
		case StatusNotCalculable:
			notCalc = append(notCalc, r)
		default:
			excl = append(excl, r)
		}
	}
	byName := func(a, b *CandidateResult) bool {
		an, bn := strings.ToLower(label(a.Bull)), strings.ToLower(label(b.Bull))
		if an != bn {
			return an < bn
		}
		return a.Bull.ID < b.Bull.ID
	}
	sort.SliceStable(calc, func(i, j int) bool {
		a, b := calc[i], calc[j]
		if math.Abs(a.Final.Final-b.Final.Final) > eps {
			return a.Final.Final > b.Final.Final
		}
		return byName(a, b)
	})
	sort.SliceStable(notCalc, func(i, j int) bool { return byName(notCalc[i], notCalc[j]) })
	sort.SliceStable(excl, func(i, j int) bool { return byName(excl[i], excl[j]) })
	for i, r := range calc {
		r.Rank = i + 1
		if i > 0 && math.Abs(calc[i-1].Final.Final-r.Final.Final) <= eps {
			r.Tied, calc[i-1].Tied = true, true
		}
	}
	out := make([]CandidateResult, 0, len(order))
	for _, group := range [][]*CandidateResult{calc, notCalc, excl} {
		for _, r := range group {
			out = append(out, *r)
		}
	}
	return out
}

// usedDatasets lists the datasets behind every evaluation that took part in the run.
func usedDatasets(in RunInput, out RunOutput) []DatasetInfo {
	byID := map[string]DatasetInfo{}
	for _, d := range in.Datasets {
		byID[d.ID] = d
	}
	seen := map[string]bool{}
	var res []DatasetInfo
	add := func(id string) {
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		if d, ok := byID[id]; ok {
			res = append(res, d)
		}
	}
	for _, r := range out.Results {
		for _, c := range r.Components {
			add(c.Source.DatasetID)
		}
		for _, it := range r.Complementarity.Items {
			add(it.Source.DatasetID)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i].Name != res[j].Name {
			return res[i].Name < res[j].Name
		}
		return res[i].Version < res[j].Version
	})
	if res == nil {
		res = []DatasetInfo{}
	}
	return res
}
