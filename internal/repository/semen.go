package repository

import (
	"context"
	"database/sql"
	"strings"

	"genetica/internal/domain"
)

type SemenRepo struct{ db DBTX }

const semenCols = `id, bull_id, supplier, batch, quantity, unit_cost, semen_type, notes, is_demo, created_at, updated_at`

func scanSemen(sc interface{ Scan(...any) error }) (*domain.SemenInventory, error) {
	var s domain.SemenInventory
	var cost sql.NullFloat64
	var demo int
	if err := sc.Scan(&s.ID, &s.BullID, &s.Supplier, &s.Batch, &s.Quantity, &cost, &s.SemenType, &s.Notes, &demo, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return nil, err
	}
	s.UnitCost, s.IsDemo = floatPtr(cost), demo == 1
	return &s, nil
}

func (r *SemenRepo) Insert(ctx context.Context, s *domain.SemenInventory) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO semen_inventory (`+semenCols+`) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		s.ID, s.BullID, s.Supplier, s.Batch, s.Quantity, nullFloat(s.UnitCost), s.SemenType, s.Notes, boolInt(s.IsDemo), s.CreatedAt, s.UpdatedAt)
	return mapErr(err, "estoque de sêmen")
}

func (r *SemenRepo) Get(ctx context.Context, id string) (*domain.SemenInventory, error) {
	s, err := scanSemen(r.db.QueryRowContext(ctx, `SELECT `+semenCols+` FROM semen_inventory WHERE id=?`, id))
	return s, mapErr(err, "estoque de sêmen")
}

// SetQuantity updates the stock level; the CHECK constraint forbids negatives.
func (r *SemenRepo) SetQuantity(ctx context.Context, id string, qty int, updatedAt string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE semen_inventory SET quantity=?, updated_at=? WHERE id=?`, qty, updatedAt, id)
	if err != nil {
		return mapErr(err, "estoque de sêmen")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return mapErr(sql.ErrNoRows, "estoque de sêmen")
	}
	return nil
}

func (r *SemenRepo) ListByBull(ctx context.Context, bullID string) ([]domain.SemenInventory, error) {
	return r.list(ctx, `SELECT `+semenCols+` FROM semen_inventory WHERE bull_id=? ORDER BY created_at`, bullID)
}

func (r *SemenRepo) List(ctx context.Context) ([]domain.SemenInventory, error) {
	return r.list(ctx, `SELECT `+semenCols+` FROM semen_inventory ORDER BY created_at`)
}

func (r *SemenRepo) list(ctx context.Context, q string, args ...any) ([]domain.SemenInventory, error) {
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, mapErr(err, "estoque de sêmen")
	}
	defer rows.Close()
	out := []domain.SemenInventory{}
	for rows.Next() {
		s, err := scanSemen(rows)
		if err != nil {
			return nil, mapErr(err, "estoque de sêmen")
		}
		out = append(out, *s)
	}
	return out, mapErr(rows.Err(), "estoque de sêmen")
}

// ---- movements and types

// ErrInsufficientStock is returned when a movement would make stock negative.
type InsufficientStockError struct{ Available int }

func (e *InsufficientStockError) Error() string { return "estoque insuficiente" }

// Move changes a lot's quantity by delta atomically and records the movement.
// The UPDATE is conditional, so even concurrent writers cannot drive stock negative.
func (r *SemenRepo) Move(ctx context.Context, m *domain.SemenMovement, updatedAt string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE semen_inventory SET quantity = quantity + ?, updated_at = ? WHERE id = ? AND quantity + ? >= 0`,
		m.Delta, updatedAt, m.InventoryID, m.Delta)
	if err != nil {
		return mapErr(err, "estoque de sêmen")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		var q int
		if err := r.db.QueryRowContext(ctx, `SELECT quantity FROM semen_inventory WHERE id=?`, m.InventoryID).Scan(&q); err != nil {
			return mapErr(err, "lote de sêmen")
		}
		return &InsufficientStockError{Available: q}
	}
	if err := r.db.QueryRowContext(ctx, `SELECT quantity, bull_id FROM semen_inventory WHERE id=?`, m.InventoryID).Scan(&m.BalanceAfter, &m.BullID); err != nil {
		return mapErr(err, "lote de sêmen")
	}
	return r.InsertMovement(ctx, m)
}

// InsertMovement stores a movement whose balance has already been computed.
func (r *SemenRepo) InsertMovement(ctx context.Context, m *domain.SemenMovement) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO semen_movements (id, inventory_id, bull_id, movement_type, delta, balance_after, reason, note, mating_id, is_demo, created_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		m.ID, m.InventoryID, m.BullID, m.Type, m.Delta, m.BalanceAfter, m.Reason, m.Note, nullStr(m.MatingID), boolInt(m.IsDemo), m.CreatedAt)
	return mapErr(err, "movimentação")
}

// UpdateInfo edits descriptive fields of a lot (never the quantity).
func (r *SemenRepo) UpdateInfo(ctx context.Context, s *domain.SemenInventory) error {
	res, err := r.db.ExecContext(ctx, `UPDATE semen_inventory SET supplier=?, batch=?, unit_cost=?, semen_type=?, notes=?, updated_at=? WHERE id=?`,
		s.Supplier, s.Batch, nullFloat(s.UnitCost), s.SemenType, s.Notes, s.UpdatedAt, s.ID)
	if err != nil {
		return mapErr(err, "lote de sêmen")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return mapErr(sql.ErrNoRows, "lote de sêmen")
	}
	return nil
}

func (r *SemenRepo) ListMovements(ctx context.Context, bullID string, limit int) ([]domain.SemenMovement, error) {
	q := `SELECT id, inventory_id, bull_id, movement_type, delta, balance_after, reason, note, mating_id, is_demo, created_at
		FROM semen_movements WHERE bull_id=? ORDER BY created_at DESC, rowid DESC`
	args := []any{bullID}
	if limit > 0 {
		q += ` LIMIT ?`
		args = append(args, limit)
	}
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, mapErr(err, "movimentações")
	}
	defer rows.Close()
	out := []domain.SemenMovement{}
	for rows.Next() {
		var m domain.SemenMovement
		var mating sql.NullString
		var demo int
		if err := rows.Scan(&m.ID, &m.InventoryID, &m.BullID, &m.Type, &m.Delta, &m.BalanceAfter, &m.Reason, &m.Note, &mating, &demo, &m.CreatedAt); err != nil {
			return nil, mapErr(err, "movimentações")
		}
		m.MatingID, m.IsDemo = mating.String, demo == 1
		out = append(out, m)
	}
	return out, mapErr(rows.Err(), "movimentações")
}

func (r *SemenRepo) ListTypes(ctx context.Context) ([]domain.SemenType, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT code, name, is_builtin FROM semen_types ORDER BY is_builtin DESC, name COLLATE NOCASE`)
	if err != nil {
		return nil, mapErr(err, "tipos de sêmen")
	}
	defer rows.Close()
	out := []domain.SemenType{}
	for rows.Next() {
		var t domain.SemenType
		var b int
		if err := rows.Scan(&t.Code, &t.Name, &b); err != nil {
			return nil, mapErr(err, "tipos de sêmen")
		}
		t.IsBuiltin = b == 1
		out = append(out, t)
	}
	return out, mapErr(rows.Err(), "tipos de sêmen")
}

func (r *SemenRepo) InsertType(ctx context.Context, t *domain.SemenType) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO semen_types (code, name, is_builtin) VALUES (?,?,?)`, t.Code, t.Name, boolInt(t.IsBuiltin))
	return mapErr(err, "tipo de sêmen")
}

func (r *SemenRepo) TypeExists(ctx context.Context, code string) (bool, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM semen_types WHERE code=?`, code).Scan(&n)
	return n > 0, mapErr(err, "tipo de sêmen")
}

// BullRow is one candidate row of the semen bank before stock aggregation.
type BullFilter struct {
	Search          string
	BreedID         string
	IncludeArchived bool
}

// Bulls returns males that are bulls or hold any semen lot.
func (r *AnimalRepo) Bulls(ctx context.Context, f BullFilter) ([]domain.Animal, error) {
	where := []string{`a.sex = 'male'`, `(a.category = 'bull' OR EXISTS (SELECT 1 FROM semen_inventory s WHERE s.bull_id = a.id))`}
	var args []any
	if !f.IncludeArchived {
		where = append(where, `a.status <> 'archived'`)
	}
	if f.BreedID != "" {
		where, args = append(where, `a.breed_id = ?`), append(args, f.BreedID)
	}
	if q := strings.TrimSpace(f.Search); q != "" {
		like := "%" + escapeLike(q) + "%"
		where, args = append(where, `(a.name LIKE ? ESCAPE '\' OR a.registration LIKE ? ESCAPE '\')`), append(args, like, like)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT `+prefixCols("a.", animalCols)+` FROM animals a WHERE `+strings.Join(where, " AND ")+` ORDER BY a.name COLLATE NOCASE, a.id`, args...)
	if err != nil {
		return nil, mapErr(err, "touros")
	}
	defer rows.Close()
	out := []domain.Animal{}
	for rows.Next() {
		a, err := scanAnimal(rows)
		if err != nil {
			return nil, mapErr(err, "touros")
		}
		out = append(out, *a)
	}
	return out, mapErr(rows.Err(), "touros")
}
