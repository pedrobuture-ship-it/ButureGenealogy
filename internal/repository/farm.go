package repository

import (
	"context"
	"database/sql"

	"genetica/internal/domain"
)

type FarmRepo struct{ db DBTX }

const farmCols = `id, name, owner_name, city, state, notes, is_demo, created_at, updated_at`

func scanFarm(sc interface{ Scan(...any) error }) (*domain.Farm, error) {
	var f domain.Farm
	var owner, city, state sql.NullString
	var demo int
	if err := sc.Scan(&f.ID, &f.Name, &owner, &city, &state, &f.Notes, &demo, &f.CreatedAt, &f.UpdatedAt); err != nil {
		return nil, err
	}
	f.OwnerName, f.City, f.State, f.IsDemo = owner.String, city.String, state.String, demo == 1
	return &f, nil
}

func (r *FarmRepo) Insert(ctx context.Context, f *domain.Farm) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO farms (`+farmCols+`) VALUES (?,?,?,?,?,?,?,?,?)`,
		f.ID, f.Name, nullStr(f.OwnerName), nullStr(f.City), nullStr(f.State), f.Notes, boolInt(f.IsDemo), f.CreatedAt, f.UpdatedAt)
	return mapErr(err, "propriedade")
}

func (r *FarmRepo) Update(ctx context.Context, f *domain.Farm) error {
	res, err := r.db.ExecContext(ctx, `UPDATE farms SET name=?, owner_name=?, city=?, state=?, notes=?, updated_at=? WHERE id=?`,
		f.Name, nullStr(f.OwnerName), nullStr(f.City), nullStr(f.State), f.Notes, f.UpdatedAt, f.ID)
	if err != nil {
		return mapErr(err, "propriedade")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return mapErr(sql.ErrNoRows, "propriedade")
	}
	return nil
}

func (r *FarmRepo) Get(ctx context.Context, id string) (*domain.Farm, error) {
	f, err := scanFarm(r.db.QueryRowContext(ctx, `SELECT `+farmCols+` FROM farms WHERE id=?`, id))
	return f, mapErr(err, "propriedade")
}

func (r *FarmRepo) List(ctx context.Context) ([]domain.Farm, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+farmCols+` FROM farms ORDER BY name COLLATE NOCASE`)
	if err != nil {
		return nil, mapErr(err, "propriedades")
	}
	defer rows.Close()
	out := []domain.Farm{}
	for rows.Next() {
		f, err := scanFarm(rows)
		if err != nil {
			return nil, mapErr(err, "propriedades")
		}
		out = append(out, *f)
	}
	return out, mapErr(rows.Err(), "propriedades")
}
