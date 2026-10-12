package repository

import (
	"context"
	"database/sql"

	"genetica/internal/domain"
)

type EvaluationRepo struct{ db DBTX }

const evalCols = `id, animal_id, dataset_id, evaluation_program, evaluation_release, breed_code, trait_code,
	evaluation_type, value, unit, accuracy, percentile, genetic_base, source_reference, evaluated_at, is_demo, created_at`

func (r *EvaluationRepo) Insert(ctx context.Context, e *domain.GeneticEvaluation) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO genetic_evaluations (`+evalCols+`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		e.ID, e.AnimalID, nullStr(e.DatasetID), e.EvaluationProgram, e.EvaluationRelease, e.BreedCode, e.TraitCode,
		e.EvaluationType, e.Value, e.Unit, nullFloat(e.Accuracy), nullFloat(e.Percentile),
		nullStr(e.GeneticBase), nullStr(e.SourceReference), nullStr(e.EvaluatedAt), boolInt(e.IsDemo), e.CreatedAt)
	return mapErr(err, "avaliação genética")
}

func (r *EvaluationRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM genetic_evaluations WHERE id=?`, id)
	if err != nil {
		return mapErr(err, "avaliação genética")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return mapErr(sql.ErrNoRows, "avaliação genética")
	}
	return nil
}

func (r *EvaluationRepo) ListByAnimal(ctx context.Context, animalID string) ([]domain.GeneticEvaluation, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+evalCols+` FROM genetic_evaluations WHERE animal_id=?
		AND (dataset_id IS NULL OR dataset_id IN (SELECT id FROM datasets WHERE active=1))
		ORDER BY trait_code, evaluation_release DESC`, animalID)
	if err != nil {
		return nil, mapErr(err, "avaliações genéticas")
	}
	defer rows.Close()
	out := []domain.GeneticEvaluation{}
	for rows.Next() {
		var e domain.GeneticEvaluation
		var ds, base, ref, at sql.NullString
		var acc, pct sql.NullFloat64
		var demo int
		if err := rows.Scan(&e.ID, &e.AnimalID, &ds, &e.EvaluationProgram, &e.EvaluationRelease, &e.BreedCode, &e.TraitCode,
			&e.EvaluationType, &e.Value, &e.Unit, &acc, &pct, &base, &ref, &at, &demo, &e.CreatedAt); err != nil {
			return nil, mapErr(err, "avaliações genéticas")
		}
		e.DatasetID, e.GeneticBase, e.SourceReference, e.EvaluatedAt = ds.String, base.String, ref.String, at.String
		e.Accuracy, e.Percentile, e.IsDemo = floatPtr(acc), floatPtr(pct), demo == 1
		out = append(out, e)
	}
	return out, mapErr(rows.Err(), "avaliações genéticas")
}
