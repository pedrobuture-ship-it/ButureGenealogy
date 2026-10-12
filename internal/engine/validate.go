package engine

import (
	"fmt"
	"math"
)

// DefaultRunValidator validates the profile and the female.
type DefaultRunValidator struct{}

func (DefaultRunValidator) ValidateRun(in RunInput) []Issue {
	var out []Issue
	add := func(field, msg string) { out = append(out, Issue{field, msg}) }

	if in.Female.ID == "" {
		add("female", "Selecione a matriz.")
	} else if in.Female.Sex != "female" {
		add("female", "A matriz deve ser uma fêmea.")
	}
	if len(in.Candidates) == 0 {
		add("candidates", "Escolha ao menos um touro candidato.")
	}
	seen := map[string]bool{}
	for _, c := range in.Candidates {
		if seen[c.ID] {
			add("candidates", fmt.Sprintf("O touro %s foi informado mais de uma vez.", c.Label()))
		}
		seen[c.ID] = true
	}

	if len(in.Profile.ActiveTraits()) == 0 {
		add("profile.traits", "O perfil não tem nenhuma característica com peso maior que zero.")
	}
	out = append(out, ValidateProfile(in.Profile)...)
	return out
}

// ValidateProfile checks a profile on its own (used when saving profiles too).
// It does not require any weighted trait: an incomplete profile may be saved.
func ValidateProfile(p Profile) []Issue {
	var out []Issue
	add := func(field, msg string) { out = append(out, Issue{field, msg}) }
	keys := map[string]bool{}
	for _, t := range p.Traits {
		if t.TraitCode == "" || t.EvaluationType == "" {
			add("profile.traits", "Toda característica do perfil precisa de código e tipo de avaliação.")
		}
		if math.IsNaN(t.Weight) || math.IsInf(t.Weight, 0) || t.Weight < 0 {
			add("profile.traits", fmt.Sprintf("Peso inválido para %s (use zero ou mais).", t.TraitCode))
		}
		if keys[t.Key()] {
			add("profile.traits", fmt.Sprintf("A característica %s (%s) aparece mais de uma vez no perfil.", t.TraitCode, t.EvaluationType))
		}
		keys[t.Key()] = true
		switch t.Direction {
		case "", Maximize, Minimize, TargetRange:
		default:
			add("profile.traits", fmt.Sprintf("Direção inválida para %s.", t.TraitCode))
		}
	}
	switch p.Normalization {
	case "", NormAuto, NormPercentile, NormZScore:
	default:
		add("profile.normalization", "Método de normalização desconhecido.")
	}
	if p.MinCoverage < 0 || p.MinCoverage > 1 {
		add("profile.minCoverage", "A cobertura mínima deve estar entre 0 e 1.")
	}
	if p.Complementarity.MaxBonusPoints < 0 || p.Complementarity.MaxBonusPoints > 100 || math.IsNaN(p.Complementarity.MaxBonusPoints) {
		add("profile.complementarity", "O bônus máximo de complementaridade deve estar entre 0 e 100 pontos.")
	}
	la := p.Penalties.LowAccuracy
	if la.Threshold < 0 || la.Threshold > 1 || la.MaxPoints < 0 || la.MaxPoints > 100 {
		add("profile.penalties", "Parâmetros de penalidade por baixa acurácia inválidos.")
	}
	if p.Filters.MinAccuracy != nil && (*p.Filters.MinAccuracy < 0 || *p.Filters.MinAccuracy > 1) {
		add("profile.filters", "A acurácia mínima deve estar entre 0 e 1.")
	}
	if p.Filters.MinDoses < 0 {
		add("profile.filters", "O mínimo de doses não pode ser negativo.")
	}
	return out
}

// DefaultCandidateValidator checks the pair itself.
type DefaultCandidateValidator struct{}

func (DefaultCandidateValidator) Validate(run RunContext, c Candidate) []Issue {
	var out []Issue
	if c.Sex != "male" {
		out = append(out, Issue{"candidate", fmt.Sprintf("%s não é um macho.", c.Label())})
	}
	if c.ID == run.Female.ID {
		out = append(out, Issue{"candidate", "O touro e a matriz são o mesmo animal."})
	}
	return out
}
