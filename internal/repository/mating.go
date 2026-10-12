package repository

import (
	"context"
	"database/sql"

	"genetica/internal/domain"
)

type MatingRepo struct{ db DBTX }

const matingCols = `id, farm_id, female_id, bull_id, status, score, engine_version, calculation_snapshot, created_at`

func (r *MatingRepo) Insert(ctx context.Context, m *domain.Mating) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO matings (`+matingCols+`) VALUES (?,?,?,?,?,?,?,?,?)`,
		m.ID, m.FarmID, m.FemaleID, m.BullID, m.Status, nullFloat(m.Score), m.EngineVersion, m.CalculationSnapshot, m.CreatedAt)
	return mapErr(err, "acasalamento")
}

func (r *MatingRepo) Get(ctx context.Context, id string) (*domain.Mating, error) {
	var m domain.Mating
	var score sql.NullFloat64
	err := r.db.QueryRowContext(ctx, `SELECT `+matingCols+` FROM matings WHERE id=?`, id).
		Scan(&m.ID, &m.FarmID, &m.FemaleID, &m.BullID, &m.Status, &score, &m.EngineVersion, &m.CalculationSnapshot, &m.CreatedAt)
	if err != nil {
		return nil, mapErr(err, "acasalamento")
	}
	m.Score = floatPtr(score)
	return &m, nil
}

func (r *MatingRepo) SetStatus(ctx context.Context, id, status string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE matings SET status=? WHERE id=?`, status, id)
	if err != nil {
		return mapErr(err, "acasalamento")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return mapErr(sql.ErrNoRows, "acasalamento")
	}
	return nil
}

func (r *MatingRepo) ListByFemale(ctx context.Context, femaleID string) ([]domain.Mating, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+matingCols+` FROM matings WHERE female_id=? ORDER BY created_at DESC`, femaleID)
	if err != nil {
		return nil, mapErr(err, "acasalamentos")
	}
	defer rows.Close()
	out := []domain.Mating{}
	for rows.Next() {
		var m domain.Mating
		var score sql.NullFloat64
		if err := rows.Scan(&m.ID, &m.FarmID, &m.FemaleID, &m.BullID, &m.Status, &score, &m.EngineVersion, &m.CalculationSnapshot, &m.CreatedAt); err != nil {
			return nil, mapErr(err, "acasalamentos")
		}
		m.Score = floatPtr(score)
		out = append(out, m)
	}
	return out, mapErr(rows.Err(), "acasalamentos")
}
