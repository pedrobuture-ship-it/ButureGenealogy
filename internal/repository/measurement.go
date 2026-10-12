package repository

import (
	"context"
	"database/sql"

	"genetica/internal/domain"
)

type MeasurementRepo struct{ db DBTX }

const measCols = `id, animal_id, trait_code, value, unit, measured_at, source, notes, is_demo, created_at`

func (r *MeasurementRepo) Insert(ctx context.Context, m *domain.Measurement) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO animal_measurements (`+measCols+`) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		m.ID, m.AnimalID, m.TraitCode, m.Value, m.Unit, m.MeasuredAt, m.Source, m.Notes, boolInt(m.IsDemo), m.CreatedAt)
	return mapErr(err, "medição")
}

func (r *MeasurementRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM animal_measurements WHERE id=?`, id)
	if err != nil {
		return mapErr(err, "medição")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return mapErr(sql.ErrNoRows, "medição")
	}
	return nil
}

func (r *MeasurementRepo) ListByAnimal(ctx context.Context, animalID string) ([]domain.Measurement, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+measCols+` FROM animal_measurements WHERE animal_id=? ORDER BY trait_code, measured_at DESC`, animalID)
	if err != nil {
		return nil, mapErr(err, "medições")
	}
	defer rows.Close()
	out := []domain.Measurement{}
	for rows.Next() {
		var m domain.Measurement
		var demo int
		if err := rows.Scan(&m.ID, &m.AnimalID, &m.TraitCode, &m.Value, &m.Unit, &m.MeasuredAt, &m.Source, &m.Notes, &demo, &m.CreatedAt); err != nil {
			return nil, mapErr(err, "medições")
		}
		m.IsDemo = demo == 1
		out = append(out, m)
	}
	return out, mapErr(rows.Err(), "medições")
}
