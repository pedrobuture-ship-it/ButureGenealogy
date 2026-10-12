package repository

import (
	"context"

	"genetica/internal/domain"
)

type BreedRepo struct{ db DBTX }

func (r *BreedRepo) List(ctx context.Context, onlyActive bool) ([]domain.Breed, error) {
	q := `SELECT id, code, name, species, purpose, active FROM breeds`
	if onlyActive {
		q += ` WHERE active=1`
	}
	rows, err := r.db.QueryContext(ctx, q+` ORDER BY name COLLATE NOCASE`)
	if err != nil {
		return nil, mapErr(err, "raças")
	}
	defer rows.Close()
	out := []domain.Breed{}
	for rows.Next() {
		var b domain.Breed
		var active int
		if err := rows.Scan(&b.ID, &b.Code, &b.Name, &b.Species, &b.Purpose, &active); err != nil {
			return nil, mapErr(err, "raças")
		}
		b.Active = active == 1
		out = append(out, b)
	}
	return out, mapErr(rows.Err(), "raças")
}

func (r *BreedRepo) Exists(ctx context.Context, id string) (bool, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM breeds WHERE id=?`, id).Scan(&n)
	return n > 0, mapErr(err, "raça")
}
