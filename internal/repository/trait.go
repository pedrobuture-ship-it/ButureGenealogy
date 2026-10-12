package repository

import (
	"context"
	"database/sql"
	"strings"

	"genetica/internal/domain"
)

type TraitRepo struct{ db DBTX }

const traitCols = `id, code, name, description, category, purpose, unit, data_type, evaluation_type, optimization_direction,
	ideal_min, ideal_max, higher_is_better, source_dataset, metadata, is_builtin, created_at`

func scanTrait(sc interface{ Scan(...any) error }) (*domain.Trait, error) {
	var t domain.Trait
	var min, max sql.NullFloat64
	var hib sql.NullInt64
	var builtin int
	if err := sc.Scan(&t.ID, &t.Code, &t.Name, &t.Description, &t.Category, &t.Purpose, &t.Unit, &t.DataType, &t.EvaluationType,
		&t.OptimizationDirection, &min, &max, &hib, &t.SourceDataset, &t.Metadata, &builtin, &t.CreatedAt); err != nil {
		return nil, err
	}
	t.IdealMin, t.IdealMax, t.IsBuiltin = floatPtr(min), floatPtr(max), builtin == 1
	if hib.Valid {
		v := hib.Int64 == 1
		t.HigherIsBetter = &v
	}
	return &t, nil
}

func (r *TraitRepo) Insert(ctx context.Context, t *domain.Trait) error {
	var hib any
	if t.HigherIsBetter != nil {
		hib = boolInt(*t.HigherIsBetter)
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO traits (`+traitCols+`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		t.ID, t.Code, t.Name, t.Description, t.Category, t.Purpose, t.Unit, t.DataType, t.EvaluationType, t.OptimizationDirection,
		nullFloat(t.IdealMin), nullFloat(t.IdealMax), hib, t.SourceDataset, t.Metadata, boolInt(t.IsBuiltin), t.CreatedAt)
	return mapErr(err, "característica")
}

// Resolve finds the trait definition for (code, kind). A definition owned by
// the dataset row of the value wins over a built-in/user one.
func (r *TraitRepo) Resolve(ctx context.Context, code, kind, datasetID string) (*domain.Trait, error) {
	t, err := scanTrait(r.db.QueryRowContext(ctx, `SELECT `+traitCols+` FROM traits
		WHERE code=? AND evaluation_type=? AND (source_dataset = '' OR source_dataset = ?)
		ORDER BY source_dataset DESC LIMIT 1`, code, kind, datasetID))
	return t, mapErr(err, "característica")
}

type TraitFilter struct {
	Kind     string
	Category string
	Search   string
}

func (r *TraitRepo) List(ctx context.Context, f TraitFilter) ([]domain.Trait, error) {
	var where []string
	var args []any
	if f.Kind != "" {
		where, args = append(where, "evaluation_type = ?"), append(args, f.Kind)
	}
	if f.Category != "" {
		where, args = append(where, "category = ?"), append(args, f.Category)
	}
	if q := strings.TrimSpace(f.Search); q != "" {
		like := "%" + escapeLike(q) + "%"
		where, args = append(where, `(code LIKE ? ESCAPE '\' OR name LIKE ? ESCAPE '\')`), append(args, like, like)
	}
	q := `SELECT ` + traitCols + ` FROM traits`
	if len(where) > 0 {
		q += ` WHERE ` + strings.Join(where, " AND ")
	}
	rows, err := r.db.QueryContext(ctx, q+` ORDER BY category, name COLLATE NOCASE, evaluation_type`, args...)
	if err != nil {
		return nil, mapErr(err, "características")
	}
	defer rows.Close()
	out := []domain.Trait{}
	for rows.Next() {
		t, err := scanTrait(rows)
		if err != nil {
			return nil, mapErr(err, "características")
		}
		out = append(out, *t)
	}
	return out, mapErr(rows.Err(), "características")
}

func (r *TraitRepo) Get(ctx context.Context, id string) (*domain.Trait, error) {
	t, err := scanTrait(r.db.QueryRowContext(ctx, `SELECT `+traitCols+` FROM traits WHERE id=?`, id))
	return t, mapErr(err, "característica")
}

// ListByDataset returns the traits defined by one dataset version.
func (r *TraitRepo) ListByDataset(ctx context.Context, datasetID string) ([]domain.Trait, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+traitCols+` FROM traits WHERE source_dataset=? ORDER BY category, name COLLATE NOCASE`, datasetID)
	if err != nil {
		return nil, mapErr(err, "características")
	}
	defer rows.Close()
	out := []domain.Trait{}
	for rows.Next() {
		t, err := scanTrait(rows)
		if err != nil {
			return nil, mapErr(err, "características")
		}
		out = append(out, *t)
	}
	return out, mapErr(rows.Err(), "características")
}
