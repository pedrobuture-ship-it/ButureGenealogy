package repository

import (
	"context"
	"database/sql"
	"strings"

	"genetica/internal/domain"
)

type AnimalRepo struct{ db DBTX }

const animalCols = `id, farm_id, registration, name, sex, breed_id, birth_date, category, status,
	father_id, mother_id, notes, is_demo, created_at, updated_at`

func scanAnimal(sc interface{ Scan(...any) error }) (*domain.Animal, error) {
	var a domain.Animal
	var reg, breed, birth, father, mother sql.NullString
	var demo int
	if err := sc.Scan(&a.ID, &a.FarmID, &reg, &a.Name, &a.Sex, &breed, &birth, &a.Category, &a.Status,
		&father, &mother, &a.Notes, &demo, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return nil, err
	}
	a.Registration, a.BreedID, a.BirthDate = reg.String, breed.String, birth.String
	a.FatherID, a.MotherID, a.IsDemo = father.String, mother.String, demo == 1
	return &a, nil
}

func (r *AnimalRepo) Insert(ctx context.Context, a *domain.Animal) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO animals (`+animalCols+`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		a.ID, a.FarmID, nullStr(a.Registration), a.Name, a.Sex, nullStr(a.BreedID), nullStr(a.BirthDate), a.Category, a.Status,
		nullStr(a.FatherID), nullStr(a.MotherID), a.Notes, boolInt(a.IsDemo), a.CreatedAt, a.UpdatedAt)
	return mapErr(err, "animal")
}

func (r *AnimalRepo) Update(ctx context.Context, a *domain.Animal) error {
	res, err := r.db.ExecContext(ctx, `UPDATE animals SET farm_id=?, registration=?, name=?, sex=?, breed_id=?, birth_date=?,
		category=?, status=?, father_id=?, mother_id=?, notes=?, updated_at=? WHERE id=?`,
		a.FarmID, nullStr(a.Registration), a.Name, a.Sex, nullStr(a.BreedID), nullStr(a.BirthDate),
		a.Category, a.Status, nullStr(a.FatherID), nullStr(a.MotherID), a.Notes, a.UpdatedAt, a.ID)
	if err != nil {
		return mapErr(err, "animal")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return mapErr(sql.ErrNoRows, "animal")
	}
	return nil
}

func (r *AnimalRepo) SetStatus(ctx context.Context, id, status, updatedAt string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE animals SET status=?, updated_at=? WHERE id=?`, status, updatedAt, id)
	if err != nil {
		return mapErr(err, "animal")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return mapErr(sql.ErrNoRows, "animal")
	}
	return nil
}

func (r *AnimalRepo) Get(ctx context.Context, id string) (*domain.Animal, error) {
	a, err := scanAnimal(r.db.QueryRowContext(ctx, `SELECT `+animalCols+` FROM animals WHERE id=?`, id))
	return a, mapErr(err, "animal")
}

// AnimalFilter describes a listing query. Empty fields do not filter.
type AnimalFilter struct {
	FarmID   string
	BreedID  string
	Sex      string
	Category string
	Status   string // empty = all statuses except none; use IncludeArchived to widen
	// IncludeArchived: when Status is empty, archived animals are hidden unless true.
	IncludeArchived bool
	Query           string // matches name or registration
	BornOnOrAfter   string // YYYY-MM-DD
	BornOnOrBefore  string
	SortBy          string
	SortDesc        bool
	Limit, Offset   int
}

var animalSortColumns = map[string]string{
	"name":         "a.name COLLATE NOCASE",
	"registration": "a.registration COLLATE NOCASE",
	"birthDate":    "a.birth_date",
	"category":     "a.category",
	"status":       "a.status",
	"sex":          "a.sex",
	"breed":        "b.name COLLATE NOCASE",
	"createdAt":    "a.created_at",
}

func (r *AnimalRepo) List(ctx context.Context, f AnimalFilter) ([]domain.Animal, int, error) {
	var where []string
	var args []any
	add := func(cond string, v ...any) { where = append(where, cond); args = append(args, v...) }
	if f.FarmID != "" {
		add("a.farm_id = ?", f.FarmID)
	}
	if f.BreedID != "" {
		add("a.breed_id = ?", f.BreedID)
	}
	if f.Sex != "" {
		add("a.sex = ?", f.Sex)
	}
	if f.Category != "" {
		add("a.category = ?", f.Category)
	}
	if f.Status != "" {
		add("a.status = ?", f.Status)
	} else if !f.IncludeArchived {
		add("a.status <> 'archived'")
	}
	if q := strings.TrimSpace(f.Query); q != "" {
		like := "%" + escapeLike(q) + "%"
		add(`(a.name LIKE ? ESCAPE '\' OR a.registration LIKE ? ESCAPE '\')`, like, like)
	}
	if f.BornOnOrAfter != "" {
		add("a.birth_date >= ?", f.BornOnOrAfter)
	}
	if f.BornOnOrBefore != "" {
		add("a.birth_date <= ?", f.BornOnOrBefore)
	}
	cond := ""
	if len(where) > 0 {
		cond = " WHERE " + strings.Join(where, " AND ")
	}

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM animals a`+cond, args...).Scan(&total); err != nil {
		return nil, 0, mapErr(err, "animais")
	}

	order := "a.name COLLATE NOCASE"
	if col, ok := animalSortColumns[f.SortBy]; ok {
		order = col
	}
	dir := "ASC"
	if f.SortDesc {
		dir = "DESC"
	}
	q := `SELECT ` + prefixCols("a.", animalCols) + ` FROM animals a LEFT JOIN breeds b ON b.id = a.breed_id` +
		cond + ` ORDER BY ` + order + ` ` + dir + `, a.id`
	qargs := append([]any{}, args...)
	if f.Limit > 0 {
		q += ` LIMIT ? OFFSET ?`
		qargs = append(qargs, f.Limit, f.Offset)
	}
	rows, err := r.db.QueryContext(ctx, q, qargs...)
	if err != nil {
		return nil, 0, mapErr(err, "animais")
	}
	defer rows.Close()
	out := []domain.Animal{}
	for rows.Next() {
		a, err := scanAnimal(rows)
		if err != nil {
			return nil, 0, mapErr(err, "animais")
		}
		out = append(out, *a)
	}
	return out, total, mapErr(rows.Err(), "animais")
}

// Children returns animals whose father or mother is id.
func (r *AnimalRepo) Children(ctx context.Context, id string) ([]domain.Animal, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+animalCols+` FROM animals WHERE father_id=? OR mother_id=? ORDER BY birth_date, name`, id, id)
	if err != nil {
		return nil, mapErr(err, "descendentes")
	}
	defer rows.Close()
	out := []domain.Animal{}
	for rows.Next() {
		a, err := scanAnimal(rows)
		if err != nil {
			return nil, mapErr(err, "descendentes")
		}
		out = append(out, *a)
	}
	return out, mapErr(rows.Err(), "descendentes")
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func prefixCols(prefix, cols string) string {
	parts := strings.Split(cols, ",")
	for i, p := range parts {
		parts[i] = prefix + strings.TrimSpace(p)
	}
	return strings.Join(parts, ", ")
}

// GetMany returns the animals with the given IDs (unknown IDs are skipped).
func (r *AnimalRepo) GetMany(ctx context.Context, ids []string) ([]domain.Animal, error) {
	out := make([]domain.Animal, 0, len(ids))
	const chunk = 400 // stay below SQLite's variable limit
	for start := 0; start < len(ids); start += chunk {
		end := start + chunk
		if end > len(ids) {
			end = len(ids)
		}
		part := ids[start:end]
		q := `SELECT ` + animalCols + ` FROM animals WHERE id IN (?` + strings.Repeat(",?", len(part)-1) + `)`
		args := make([]any, len(part))
		for i, id := range part {
			args[i] = id
		}
		rows, err := r.db.QueryContext(ctx, q, args...)
		if err != nil {
			return nil, mapErr(err, "animais")
		}
		for rows.Next() {
			a, err := scanAnimal(rows)
			if err != nil {
				rows.Close()
				return nil, mapErr(err, "animais")
			}
			out = append(out, *a)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, mapErr(err, "animais")
		}
		rows.Close()
	}
	return out, nil
}

// All returns every animal (optionally of one farm), including archived ones.
func (r *AnimalRepo) All(ctx context.Context, farmID string) ([]domain.Animal, error) {
	q, args := `SELECT `+animalCols+` FROM animals`, []any{}
	if farmID != "" {
		q += ` WHERE farm_id=?`
		args = append(args, farmID)
	}
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, mapErr(err, "animais")
	}
	defer rows.Close()
	out := []domain.Animal{}
	for rows.Next() {
		a, err := scanAnimal(rows)
		if err != nil {
			return nil, mapErr(err, "animais")
		}
		out = append(out, *a)
	}
	return out, mapErr(rows.Err(), "animais")
}
