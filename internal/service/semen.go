package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"genetica/internal/apperr"
	"genetica/internal/database"
	"genetica/internal/domain"
	"genetica/internal/repository"
)

// SemenService manages the semen bank. Bulls are ordinary animals (males);
// this service only adds stock lots and a movement history on top of them.
type SemenService struct{ db *database.DB }

// SemenInput creates a new lot (supplier + batch) with an initial quantity.
type SemenInput struct {
	BullID    string   `json:"bullId"`
	Supplier  string   `json:"supplier"`
	Batch     string   `json:"batch"`
	Quantity  int      `json:"quantity"`
	UnitCost  *float64 `json:"unitCost"`
	SemenType string   `json:"semenType"`
	Notes     string   `json:"notes"`
	Reason    string   `json:"reason"` // entry reason; defaults to "purchase"
}

// LotInfoInput edits descriptive fields of a lot. Quantity changes only through movements.
type LotInfoInput struct {
	Supplier  string   `json:"supplier"`
	Batch     string   `json:"batch"`
	UnitCost  *float64 `json:"unitCost"`
	SemenType string   `json:"semenType"`
	Notes     string   `json:"notes"`
}

// MovementInput is a stock movement request on one lot.
type MovementInput struct {
	InventoryID string `json:"inventoryId"`
	Quantity    int    `json:"quantity"` // doses (positive); for adjustments the signed delta
	Reason      string `json:"reason"`
	Note        string `json:"note"`
	MatingID    string `json:"matingId"`
}

func (s *SemenService) Types(ctx context.Context) ([]domain.SemenType, error) {
	return repos(s.db).Semen.ListTypes(ctx)
}

// CreateType adds a user-defined semen type; its code is derived from the name.
func (s *SemenService) CreateType(ctx context.Context, name string) (*domain.SemenType, error) {
	if err := domain.ValidateSemenTypeName(name); err != nil {
		return nil, err
	}
	code := slugCode(name)
	if code == "" {
		return nil, fieldError("name", "O nome precisa conter letras ou números.")
	}
	t := &domain.SemenType{Code: code, Name: strings.TrimSpace(name)}
	if err := repos(s.db).Semen.InsertType(ctx, t); err != nil {
		if apperr.CodeOf(err) == apperr.CodeConflict {
			return nil, fieldError("name", "Já existe um tipo de sêmen com esse nome.")
		}
		return nil, err
	}
	return t, nil
}

func slugCode(name string) string {
	var b strings.Builder
	prevDash := true
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		case strings.ContainsRune("áàâãä", r):
			b.WriteRune('a')
			prevDash = false
		case strings.ContainsRune("éèêë", r):
			b.WriteRune('e')
			prevDash = false
		case strings.ContainsRune("íìîï", r):
			b.WriteRune('i')
			prevDash = false
		case strings.ContainsRune("óòôõö", r):
			b.WriteRune('o')
			prevDash = false
		case strings.ContainsRune("úùûü", r):
			b.WriteRune('u')
			prevDash = false
		case r == 'ç':
			b.WriteRune('c')
			prevDash = false
		default:
			if !prevDash {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	code := strings.Trim(b.String(), "-")
	if len(code) > 40 {
		code = strings.Trim(code[:40], "-")
	}
	if code == "" || code[0] < 'a' || code[0] > 'z' {
		if code == "" {
			return ""
		}
		code = "t-" + code
	}
	return code
}

func requireType(ctx context.Context, r *repository.Repos, code string) error {
	ok, err := r.Semen.TypeExists(ctx, code)
	if err != nil {
		return err
	}
	if !ok {
		return fieldError("semenType", "Tipo de sêmen não cadastrado.")
	}
	return nil
}

// AddLot registers a new lot for a bull and records its initial entry.
func (s *SemenService) AddLot(ctx context.Context, in SemenInput) (*domain.SemenInventory, error) {
	now := nowTS()
	lot := &domain.SemenInventory{ID: newID(), BullID: in.BullID, Supplier: strings.TrimSpace(in.Supplier), Batch: strings.TrimSpace(in.Batch),
		Quantity: in.Quantity, UnitCost: in.UnitCost, SemenType: defaultStr(in.SemenType, domain.SemenConventional), Notes: in.Notes, CreatedAt: now, UpdatedAt: now}
	if err := lot.Validate(); err != nil {
		return nil, err
	}
	reason := defaultStr(in.Reason, "purchase")
	if lot.Quantity > 0 {
		if err := domain.ValidateMovement(domain.MovEntry, lot.Quantity, reason, ""); err != nil {
			return nil, err
		}
	}
	err := s.db.WithTx(ctx, func(tx *sql.Tx) error {
		r := repository.New(tx)
		bull, err := requireAnimal(ctx, r, lot.BullID, "bullId")
		if err != nil {
			return err
		}
		if bull.Sex != domain.SexMale {
			return fieldError("bullId", "O sêmen só pode ser registrado para um macho.")
		}
		if err := requireType(ctx, r, lot.SemenType); err != nil {
			return err
		}
		if err := r.Semen.Insert(ctx, lot); err != nil {
			return err
		}
		if lot.Quantity > 0 {
			return r.Semen.InsertMovement(ctx, &domain.SemenMovement{ID: newID(), InventoryID: lot.ID, BullID: lot.BullID, Type: domain.MovEntry,
				Delta: lot.Quantity, BalanceAfter: lot.Quantity, Reason: reason, Note: "Lote criado", IsDemo: bull.IsDemo, CreatedAt: now})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return lot, nil
}

// Add is kept for callers that create a lot directly.
func (s *SemenService) Add(ctx context.Context, in SemenInput) (*domain.SemenInventory, error) {
	return s.AddLot(ctx, in)
}

func (s *SemenService) UpdateLot(ctx context.Context, id string, in LotInfoInput) (*domain.SemenInventory, error) {
	var out *domain.SemenInventory
	err := s.db.WithTx(ctx, func(tx *sql.Tx) error {
		r := repository.New(tx)
		lot, err := r.Semen.Get(ctx, id)
		if err != nil {
			return err
		}
		lot.Supplier, lot.Batch, lot.UnitCost, lot.Notes = strings.TrimSpace(in.Supplier), strings.TrimSpace(in.Batch), in.UnitCost, in.Notes
		lot.SemenType, lot.UpdatedAt = defaultStr(in.SemenType, lot.SemenType), nowTS()
		if err := lot.Validate(); err != nil {
			return err
		}
		if err := requireType(ctx, r, lot.SemenType); err != nil {
			return err
		}
		if err := r.Semen.UpdateInfo(ctx, lot); err != nil {
			return err
		}
		out = lot
		return nil
	})
	return out, err
}

func (s *SemenService) move(ctx context.Context, movType string, in MovementInput, sign int) (*domain.SemenMovement, error) {
	if err := domain.ValidateMovement(movType, in.Quantity, in.Reason, in.Note); err != nil {
		return nil, err
	}
	var m *domain.SemenMovement
	err := s.db.WithTx(ctx, func(tx *sql.Tx) error {
		r := repository.New(tx)
		lot, err := r.Semen.Get(ctx, in.InventoryID)
		if err != nil {
			return err
		}
		bull, err := r.Animals.Get(ctx, lot.BullID)
		if err != nil {
			return err
		}
		delta := in.Quantity * sign
		reason := in.Reason
		if movType == domain.MovAdjustment {
			delta, reason = in.Quantity, domain.ReasonCountCorrection
		}
		m = &domain.SemenMovement{ID: newID(), InventoryID: lot.ID, BullID: lot.BullID, Type: movType, Delta: delta, Reason: reason,
			Note: in.Note, MatingID: in.MatingID, IsDemo: bull.IsDemo, CreatedAt: nowTS()}
		if err := r.Semen.Move(ctx, m, m.CreatedAt); err != nil {
			var ins *repository.InsufficientStockError
			if errors.As(err, &ins) {
				return fieldError("quantity", fmt.Sprintf("Estoque insuficiente: há %d dose(s) neste lote.", ins.Available))
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Receive adds doses to an existing lot.
func (s *SemenService) Receive(ctx context.Context, in MovementInput) (*domain.SemenMovement, error) {
	if in.Reason == "" {
		in.Reason = "purchase"
	}
	return s.move(ctx, domain.MovEntry, in, +1)
}

// Withdraw removes doses from a lot. Stock can never become negative.
func (s *SemenService) Withdraw(ctx context.Context, in MovementInput) (*domain.SemenMovement, error) {
	return s.move(ctx, domain.MovWithdrawal, in, -1)
}

// Adjust corrects a lot's count by a signed delta; a note is mandatory.
func (s *SemenService) Adjust(ctx context.Context, in MovementInput) (*domain.SemenMovement, error) {
	return s.move(ctx, domain.MovAdjustment, in, +1)
}

func (s *SemenService) ListByBull(ctx context.Context, bullID string) ([]domain.SemenInventory, error) {
	return repos(s.db).Semen.ListByBull(ctx, bullID)
}

func (s *SemenService) Movements(ctx context.Context, bullID string, limit int) ([]domain.SemenMovement, error) {
	return repos(s.db).Semen.ListMovements(ctx, bullID, limit)
}

// ---- bank listing

const (
	StockInStock  = "in_stock"
	StockDepleted = "depleted"
	StockNoSemen  = "no_semen"

	FilterWithLots = "with_lots" // default: bulls that have at least one lot
	FilterInStock  = "in_stock"
	FilterAll      = "all"
)

type BankQuery struct {
	Search          string `json:"search"`
	BreedID         string `json:"breedId"`
	Supplier        string `json:"supplier"`
	StockFilter     string `json:"stockFilter"` // with_lots (default) | in_stock | all
	IncludeArchived bool   `json:"includeArchived"`
	SortBy          string `json:"sortBy"`
	SortDesc        bool   `json:"sortDesc"`
	Page            int    `json:"page"`
	PageSize        int    `json:"pageSize"`
}

// BankRow is one bull of the semen bank. UnitCost is the dose-weighted average
// over lots with a cost and stock; if no stock remains it is the simple mean of
// the lot costs; nil when no lot has a cost.
type BankRow struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Registration string   `json:"registration"`
	BreedName    string   `json:"breedName"`
	AnimalStatus string   `json:"animalStatus"`
	Doses        int      `json:"doses"`
	Lots         int      `json:"lots"`
	Suppliers    []string `json:"suppliers"`
	UnitCost     *float64 `json:"unitCost"`
	StockStatus  string   `json:"stockStatus"`
	IsDemo       bool     `json:"isDemo"`
}

type BankPage struct {
	Items []BankRow `json:"items"`
	Total int       `json:"total"`
	Page  int       `json:"page"`
	// TotalDoses sums every row that matched the filters, not just the page.
	TotalDoses int `json:"totalDoses"`
	// Suppliers lists every supplier with at least one lot (for the filter).
	Suppliers []string `json:"suppliers"`
}

func unitCost(lots []domain.SemenInventory) *float64 {
	var wSum, wDoses float64
	var plain []float64
	for _, l := range lots {
		if l.UnitCost == nil {
			continue
		}
		plain = append(plain, *l.UnitCost)
		if l.Quantity > 0 {
			wSum += *l.UnitCost * float64(l.Quantity)
			wDoses += float64(l.Quantity)
		}
	}
	switch {
	case wDoses > 0:
		v := wSum / wDoses
		return &v
	case len(plain) > 0:
		var t float64
		for _, c := range plain {
			t += c
		}
		v := t / float64(len(plain))
		return &v
	}
	return nil
}

func (s *SemenService) Bank(ctx context.Context, q BankQuery) (*BankPage, error) {
	r := repos(s.db)
	bulls, err := r.Animals.Bulls(ctx, repository.BullFilter{Search: q.Search, BreedID: q.BreedID, IncludeArchived: q.IncludeArchived})
	if err != nil {
		return nil, err
	}
	lots, err := r.Semen.List(ctx)
	if err != nil {
		return nil, err
	}
	byBull := map[string][]domain.SemenInventory{}
	for _, l := range lots {
		byBull[l.BullID] = append(byBull[l.BullID], l)
	}
	breeds, err := r.Breeds.List(ctx, false)
	if err != nil {
		return nil, err
	}
	breedName := map[string]string{}
	for _, b := range breeds {
		breedName[b.ID] = b.Name
	}

	filter := defaultStr(q.StockFilter, FilterWithLots)
	var rows []BankRow
	for _, a := range bulls {
		ls := byBull[a.ID]
		row := BankRow{ID: a.ID, Name: a.Name, Registration: a.Registration, BreedName: breedName[a.BreedID], AnimalStatus: a.Status,
			Lots: len(ls), Suppliers: []string{}, IsDemo: a.IsDemo, UnitCost: unitCost(ls)}
		seen := map[string]bool{}
		for _, l := range ls {
			row.Doses += l.Quantity
			if sp := strings.TrimSpace(l.Supplier); sp != "" && !seen[sp] {
				seen[sp] = true
				row.Suppliers = append(row.Suppliers, sp)
			}
		}
		sort.Strings(row.Suppliers)
		switch {
		case row.Lots == 0:
			row.StockStatus = StockNoSemen
		case row.Doses == 0:
			row.StockStatus = StockDepleted
		default:
			row.StockStatus = StockInStock
		}
		if q.Supplier != "" && !seen[q.Supplier] {
			continue
		}
		switch filter {
		case FilterInStock:
			if row.StockStatus != StockInStock {
				continue
			}
		case FilterWithLots:
			if row.Lots == 0 {
				continue
			}
		}
		rows = append(rows, row)
	}

	sortBank(rows, q.SortBy, q.SortDesc)
	total, totalDoses := len(rows), 0
	for _, r := range rows {
		totalDoses += r.Doses
	}
	size := q.PageSize
	if size <= 0 {
		size = 50
	}
	if size > 500 {
		size = 500
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	start := (page - 1) * size
	if start > len(rows) {
		start = len(rows)
	}
	end := start + size
	if end > len(rows) {
		end = len(rows)
	}
	items := rows[start:end]
	if items == nil {
		items = []BankRow{}
	}
	supSet := map[string]bool{}
	for _, l := range lots {
		if sp := strings.TrimSpace(l.Supplier); sp != "" {
			supSet[sp] = true
		}
	}
	suppliers := make([]string, 0, len(supSet))
	for sp := range supSet {
		suppliers = append(suppliers, sp)
	}
	sort.Strings(suppliers)
	return &BankPage{Items: items, Total: total, Page: page, TotalDoses: totalDoses, Suppliers: suppliers}, nil
}

func sortBank(rows []BankRow, by string, desc bool) {
	less := func(i, j int) bool { return strings.ToLower(rows[i].Name) < strings.ToLower(rows[j].Name) }
	switch by {
	case "breed":
		less = func(i, j int) bool { return strings.ToLower(rows[i].BreedName) < strings.ToLower(rows[j].BreedName) }
	case "registration":
		less = func(i, j int) bool {
			return strings.ToLower(rows[i].Registration) < strings.ToLower(rows[j].Registration)
		}
	case "supplier":
		less = func(i, j int) bool {
			return strings.Join(rows[i].Suppliers, ",") < strings.Join(rows[j].Suppliers, ",")
		}
	case "doses":
		less = func(i, j int) bool { return rows[i].Doses < rows[j].Doses }
	case "unitCost":
		less = func(i, j int) bool {
			a, b := rows[i].UnitCost, rows[j].UnitCost
			switch {
			case a == nil:
				return b != nil // costless rows first when ascending
			case b == nil:
				return false
			}
			return *a < *b
		}
	case "status":
		less = func(i, j int) bool { return rows[i].StockStatus < rows[j].StockStatus }
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if desc {
			return less(j, i)
		}
		return less(i, j)
	})
}

// BullSheet is the stock part of the bull's page; genetics and pedigree come
// from the trait and pedigree services, using the same animal id.
type BullSheet struct {
	Detail      *AnimalDetail           `json:"detail"`
	Lots        []domain.SemenInventory `json:"lots"`
	TotalDoses  int                     `json:"totalDoses"`
	UnitCost    *float64                `json:"unitCost"`
	Movements   []domain.SemenMovement  `json:"movements"`
	StockStatus string                  `json:"stockStatus"`
}

func (s *SemenService) Sheet(ctx context.Context, bullID string, movementLimit int) (*BullSheet, error) {
	animals := &AnimalService{db: s.db}
	d, err := animals.Detail(ctx, bullID)
	if err != nil {
		return nil, err
	}
	if d.Animal.Sex != domain.SexMale {
		return nil, apperr.New(apperr.CodeValidation, "Este animal não é um touro.")
	}
	r := repos(s.db)
	lots, err := r.Semen.ListByBull(ctx, bullID)
	if err != nil {
		return nil, err
	}
	mov, err := r.Semen.ListMovements(ctx, bullID, movementLimit)
	if err != nil {
		return nil, err
	}
	sh := &BullSheet{Detail: d, Lots: lots, Movements: mov, UnitCost: unitCost(lots)}
	for _, l := range lots {
		sh.TotalDoses += l.Quantity
	}
	switch {
	case len(lots) == 0:
		sh.StockStatus = StockNoSemen
	case sh.TotalDoses == 0:
		sh.StockStatus = StockDepleted
	default:
		sh.StockStatus = StockInStock
	}
	return sh, nil
}
