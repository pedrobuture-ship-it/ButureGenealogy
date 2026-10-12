// Package repository isolates all SQL. Repositories accept either *sql.DB or
// *sql.Tx through DBTX, so services can compose them inside transactions.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"genetica/internal/apperr"
)

// DBTX is satisfied by *sql.DB and *sql.Tx.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Repos bundles every repository over one DBTX.
type Repos struct {
	Farms        *FarmRepo
	Breeds       *BreedRepo
	Animals      *AnimalRepo
	Measurements *MeasurementRepo
	Evaluations  *EvaluationRepo
	Semen        *SemenRepo
	Matings      *MatingRepo
	Traits       *TraitRepo
}

func New(db DBTX) *Repos {
	return &Repos{
		Farms:        &FarmRepo{db},
		Breeds:       &BreedRepo{db},
		Animals:      &AnimalRepo{db},
		Measurements: &MeasurementRepo{db},
		Evaluations:  &EvaluationRepo{db},
		Semen:        &SemenRepo{db},
		Matings:      &MatingRepo{db},
		Traits:       &TraitRepo{db},
	}
}

func nullStr(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func nullFloat(f *float64) any {
	if f == nil {
		return nil
	}
	return *f
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func floatPtr(n sql.NullFloat64) *float64 {
	if !n.Valid {
		return nil
	}
	v := n.Float64
	return &v
}

// mapErr translates driver errors into application errors.
func mapErr(err error, what string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return apperr.Wrap(apperr.CodeNotFound, what+" não encontrado(a).", err)
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "UNIQUE constraint failed"):
		return apperr.Wrap(apperr.CodeConflict, "Já existe um registro com esses dados ("+what+").", err)
	case strings.Contains(msg, "FOREIGN KEY constraint failed"):
		return apperr.Wrap(apperr.CodeConflict, "Referência inválida ou registro em uso ("+what+").", err)
	case strings.Contains(msg, "CHECK constraint failed"):
		return apperr.Wrap(apperr.CodeValidation, "Dados inválidos ("+what+").", err)
	}
	return apperr.Wrap(apperr.CodeDatabase, "Erro ao acessar o banco de dados ("+what+").", err)
}
