package service

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"genetica/internal/domain"
)

func f64(v float64) *float64 { return &v }

func newBull(t *testing.T, fx fixture, name string, mod ...func(*AnimalInput)) *domain.Animal {
	t.Helper()
	return fx.animal(t, name, domain.SexMale, domain.CategoryBull, mod...)
}

func TestAddLotRecordsEntryAndValidates(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	bull := newBull(t, fx, "Touro")
	cow := fx.animal(t, "Vaca", domain.SexFemale, domain.CategoryCow)

	lot, err := fx.svc.Semen.AddLot(ctx, SemenInput{BullID: bull.ID, Supplier: "Central", Batch: "L1", Quantity: 20, UnitCost: f64(40)})
	if err != nil || lot.SemenType != domain.SemenConventional || lot.Quantity != 20 {
		t.Fatalf("%+v %v", lot, err)
	}
	mov, _ := fx.svc.Semen.Movements(ctx, bull.ID, 0)
	if len(mov) != 1 || mov[0].Type != domain.MovEntry || mov[0].Delta != 20 || mov[0].BalanceAfter != 20 || mov[0].InventoryID != lot.ID {
		t.Fatalf("initial stock must be recorded as a movement: %+v", mov)
	}
	if _, err := fx.svc.Semen.AddLot(ctx, SemenInput{BullID: cow.ID, Quantity: 5}); err == nil {
		t.Fatal("semen for a female must fail")
	}
	_, err = fx.svc.Semen.AddLot(ctx, SemenInput{BullID: bull.ID, Quantity: 5, SemenType: "inexistente"})
	wantField(t, err, "semenType")
	_, err = fx.svc.Semen.AddLot(ctx, SemenInput{BullID: bull.ID, Quantity: -1})
	wantField(t, err, "quantity")
	_, err = fx.svc.Semen.AddLot(ctx, SemenInput{BullID: bull.ID, Quantity: 5, UnitCost: f64(-1)})
	wantField(t, err, "unitCost")
	// an empty lot is allowed (stock arrives later) and records no movement
	empty, err := fx.svc.Semen.AddLot(ctx, SemenInput{BullID: bull.ID, Quantity: 0, Supplier: "Outra"})
	if err != nil || empty.Quantity != 0 {
		t.Fatal(err)
	}
	if mov, _ = fx.svc.Semen.Movements(ctx, bull.ID, 0); len(mov) != 1 {
		t.Fatalf("an empty lot must not create a movement: %d", len(mov))
	}
}

func TestWithdrawNeverGoesNegative(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	bull := newBull(t, fx, "Touro")
	lot, _ := fx.svc.Semen.AddLot(ctx, SemenInput{BullID: bull.ID, Quantity: 10})

	m, err := fx.svc.Semen.Withdraw(ctx, MovementInput{InventoryID: lot.ID, Quantity: 4, Reason: "insemination", Note: "vaca 0472"})
	if err != nil || m.Delta != -4 || m.BalanceAfter != 6 || m.Type != domain.MovWithdrawal {
		t.Fatalf("%+v %v", m, err)
	}
	// more than available: refused with a clear message, nothing changes, nothing recorded
	_, err = fx.svc.Semen.Withdraw(ctx, MovementInput{InventoryID: lot.ID, Quantity: 7, Reason: "sale"})
	wantField(t, err, "quantity")
	got, _ := fx.svc.Semen.ListByBull(ctx, bull.ID)
	if got[0].Quantity != 6 {
		t.Fatalf("stock changed by a refused withdrawal: %d", got[0].Quantity)
	}
	mov, _ := fx.svc.Semen.Movements(ctx, bull.ID, 0)
	if len(mov) != 2 {
		t.Fatalf("a refused withdrawal must not be recorded: %d", len(mov))
	}
	// zero / invalid reason / unknown lot
	_, err = fx.svc.Semen.Withdraw(ctx, MovementInput{InventoryID: lot.ID, Quantity: 0, Reason: "sale"})
	wantField(t, err, "quantity")
	_, err = fx.svc.Semen.Withdraw(ctx, MovementInput{InventoryID: lot.ID, Quantity: 1, Reason: "whim"})
	wantField(t, err, "reason")
	if _, err := fx.svc.Semen.Withdraw(ctx, MovementInput{InventoryID: "nope", Quantity: 1, Reason: "sale"}); err == nil {
		t.Fatal("unknown lot")
	}
	// exactly everything is fine
	if m, err := fx.svc.Semen.Withdraw(ctx, MovementInput{InventoryID: lot.ID, Quantity: 6, Reason: "loss"}); err != nil || m.BalanceAfter != 0 {
		t.Fatalf("%+v %v", m, err)
	}
	if _, err := fx.svc.Semen.Withdraw(ctx, MovementInput{InventoryID: lot.ID, Quantity: 1, Reason: "loss"}); err == nil {
		t.Fatal("empty lot cannot be withdrawn from")
	}
}

func TestConcurrentWithdrawalsNeverOversell(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	bull := newBull(t, fx, "Touro")
	lot, _ := fx.svc.Semen.AddLot(ctx, SemenInput{BullID: bull.ID, Quantity: 30})

	var ok, refused int32
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := fx.svc.Semen.Withdraw(ctx, MovementInput{InventoryID: lot.ID, Quantity: 1, Reason: "insemination"}); err != nil {
				atomic.AddInt32(&refused, 1)
				return
			}
			atomic.AddInt32(&ok, 1)
		}()
	}
	wg.Wait()
	if ok != 30 || refused != 20 {
		t.Fatalf("ok=%d refused=%d, want 30/20", ok, refused)
	}
	got, _ := fx.svc.Semen.ListByBull(ctx, bull.ID)
	if got[0].Quantity != 0 {
		t.Fatalf("final stock %d", got[0].Quantity)
	}
	mov, _ := fx.svc.Semen.Movements(ctx, bull.ID, 0)
	sum := 0
	for _, m := range mov {
		sum += m.Delta
		if m.BalanceAfter < 0 {
			t.Fatalf("negative balance recorded: %+v", m)
		}
	}
	if sum != 0 || len(mov) != 31 {
		t.Fatalf("movement history must add up to the stock: sum=%d n=%d", sum, len(mov))
	}
}

func TestReceiveAndAdjust(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	bull := newBull(t, fx, "Touro")
	lot, _ := fx.svc.Semen.AddLot(ctx, SemenInput{BullID: bull.ID, Quantity: 5})

	m, err := fx.svc.Semen.Receive(ctx, MovementInput{InventoryID: lot.ID, Quantity: 15, Note: "reposição"})
	if err != nil || m.BalanceAfter != 20 || m.Reason != "purchase" {
		t.Fatalf("%+v %v", m, err)
	}
	// adjustments need a note and can go either way, but never below zero
	_, err = fx.svc.Semen.Adjust(ctx, MovementInput{InventoryID: lot.ID, Quantity: -2})
	wantField(t, err, "note")
	m, err = fx.svc.Semen.Adjust(ctx, MovementInput{InventoryID: lot.ID, Quantity: -2, Note: "contagem do botijão"})
	if err != nil || m.BalanceAfter != 18 || m.Reason != domain.ReasonCountCorrection {
		t.Fatalf("%+v %v", m, err)
	}
	if _, err := fx.svc.Semen.Adjust(ctx, MovementInput{InventoryID: lot.ID, Quantity: -19, Note: "x"}); err == nil {
		t.Fatal("an adjustment must not push stock below zero")
	}
	_, err = fx.svc.Semen.Adjust(ctx, MovementInput{InventoryID: lot.ID, Quantity: 0, Note: "x"})
	wantField(t, err, "quantity")
	mov, _ := fx.svc.Semen.Movements(ctx, bull.ID, 0)
	if mov[0].Type != domain.MovAdjustment || len(mov) != 3 {
		t.Fatalf("history is newest-first: %+v", mov)
	}
}

func TestUpdateLotNeverTouchesQuantity(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	bull := newBull(t, fx, "Touro")
	lot, _ := fx.svc.Semen.AddLot(ctx, SemenInput{BullID: bull.ID, Quantity: 8, Supplier: "A"})
	up, err := fx.svc.Semen.UpdateLot(ctx, lot.ID, LotInfoInput{Supplier: "B", Batch: "X9", UnitCost: f64(55.5), SemenType: domain.SemenSexed, Notes: "n"})
	if err != nil || up.Quantity != 8 || up.Supplier != "B" || up.SemenType != domain.SemenSexed || *up.UnitCost != 55.5 {
		t.Fatalf("%+v %v", up, err)
	}
	_, err = fx.svc.Semen.UpdateLot(ctx, lot.ID, LotInfoInput{SemenType: "zzz"})
	wantField(t, err, "semenType")
	if _, err := fx.svc.Semen.UpdateLot(ctx, "nope", LotInfoInput{}); err == nil {
		t.Fatal("unknown lot")
	}
}

func TestConfigurableSemenTypes(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	bull := newBull(t, fx, "Touro")
	types, _ := fx.svc.Semen.Types(ctx)
	if len(types) != 2 || !types[0].IsBuiltin {
		t.Fatalf("built-in types: %+v", types)
	}
	tp, err := fx.svc.Semen.CreateType(ctx, " Sêmen Congelado Premium ")
	if err != nil || tp.Code != "semen-congelado-premium" || tp.IsBuiltin {
		t.Fatalf("%+v %v", tp, err)
	}
	if _, err := fx.svc.Semen.CreateType(ctx, "Sêmen congelado premium"); err == nil {
		t.Fatal("duplicate type")
	}
	if _, err := fx.svc.Semen.CreateType(ctx, "  "); err == nil {
		t.Fatal("empty name")
	}
	if _, err := fx.svc.Semen.CreateType(ctx, "!!!"); err == nil {
		t.Fatal("name without letters")
	}
	if _, err := fx.svc.Semen.AddLot(ctx, SemenInput{BullID: bull.ID, Quantity: 1, SemenType: tp.Code}); err != nil {
		t.Fatalf("custom type must be usable: %v", err)
	}
}

func TestBankAggregationFiltersAndSorting(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	a := newBull(t, fx, "Alfa", func(i *AnimalInput) { i.BreedID = "breed-nelore"; i.Registration = "A-1" })
	b := newBull(t, fx, "Beta", func(i *AnimalInput) { i.BreedID = "breed-angus" })
	c := newBull(t, fx, "Gama") // no semen at all
	d := newBull(t, fx, "Delta")
	steer := fx.animal(t, "Boi", domain.SexMale, domain.CategorySteer) // male, not a bull, no semen => never listed
	_ = steer

	add := func(bull string, qty int, supplier string, cost *float64) *domain.SemenInventory {
		l, err := fx.svc.Semen.AddLot(ctx, SemenInput{BullID: bull, Quantity: qty, Supplier: supplier, UnitCost: cost})
		if err != nil {
			t.Fatal(err)
		}
		return l
	}
	add(a.ID, 10, "Central", f64(40))
	add(a.ID, 30, "Outra", f64(60)) // weighted: (10*40 + 30*60)/40 = 55
	lb := add(b.ID, 5, "Central", nil)
	ld := add(d.ID, 3, "Outra", f64(10))
	if _, err := fx.svc.Semen.Withdraw(ctx, MovementInput{InventoryID: ld.ID, Quantity: 3, Reason: "sale"}); err != nil { // depleted
		t.Fatal(err)
	}
	_ = lb

	page, err := fx.svc.Semen.Bank(ctx, BankQuery{})
	if err != nil || page.Total != 3 { // Alfa, Beta, Delta (Gama has no lots; default filter hides it)
		t.Fatalf("%+v %v", page, err)
	}
	rows := map[string]BankRow{}
	for _, r := range page.Items {
		rows[r.Name] = r
	}
	al := rows["Alfa"]
	if al.Doses != 40 || al.Lots != 2 || al.StockStatus != StockInStock || al.UnitCost == nil || *al.UnitCost != 55 || al.BreedName != "Nelore" || len(al.Suppliers) != 2 {
		t.Fatalf("Alfa: %+v", al)
	}
	if rows["Beta"].UnitCost != nil {
		t.Fatalf("no cost must stay nil, never 0: %+v", rows["Beta"])
	}
	if rows["Delta"].StockStatus != StockDepleted || rows["Delta"].Doses != 0 || rows["Delta"].UnitCost == nil || *rows["Delta"].UnitCost != 10 {
		t.Fatalf("Delta: %+v", rows["Delta"])
	}
	if len(page.Suppliers) != 2 || page.Suppliers[0] != "Central" {
		t.Fatalf("suppliers: %v", page.Suppliers)
	}
	if page.TotalDoses != 45 {
		t.Fatalf("total doses %d", page.TotalDoses)
	}

	if p, _ := fx.svc.Semen.Bank(ctx, BankQuery{StockFilter: FilterInStock}); p.Total != 2 {
		t.Fatalf("in stock: %d", p.Total)
	}
	all, _ := fx.svc.Semen.Bank(ctx, BankQuery{StockFilter: FilterAll})
	if all.Total != 4 {
		t.Fatalf("all bulls (incl. those without semen, excl. steers): %d", all.Total)
	}
	for _, r := range all.Items {
		if r.Name == "Gama" && r.StockStatus != StockNoSemen {
			t.Fatalf("%+v", r)
		}
	}
	if p, _ := fx.svc.Semen.Bank(ctx, BankQuery{Supplier: "Central"}); p.Total != 2 {
		t.Fatalf("supplier filter: %d", p.Total)
	}
	if p, _ := fx.svc.Semen.Bank(ctx, BankQuery{BreedID: "breed-angus"}); p.Total != 1 || p.Items[0].Name != "Beta" {
		t.Fatalf("breed filter: %+v", p)
	}
	if p, _ := fx.svc.Semen.Bank(ctx, BankQuery{Search: "A-1"}); p.Total != 1 {
		t.Fatalf("registration search: %+v", p)
	}
	if p, _ := fx.svc.Semen.Bank(ctx, BankQuery{SortBy: "doses", SortDesc: true}); p.Items[0].Name != "Alfa" {
		t.Fatalf("sort by doses desc: %+v", p.Items[0])
	}
	if p, _ := fx.svc.Semen.Bank(ctx, BankQuery{SortBy: "unitCost"}); p.Items[0].Name != "Beta" {
		t.Fatalf("cost-less first when ascending: %+v", p.Items[0])
	}
	if p, _ := fx.svc.Semen.Bank(ctx, BankQuery{PageSize: 2, Page: 2}); len(p.Items) != 1 || p.Total != 3 {
		t.Fatalf("paging: %+v", p)
	}
	// archived bulls are hidden unless asked for
	if err := fx.svc.Animals.Archive(ctx, c.ID); err != nil {
		t.Fatal(err)
	}
	if p, _ := fx.svc.Semen.Bank(ctx, BankQuery{StockFilter: FilterAll}); p.Total != 3 {
		t.Fatalf("archived hidden: %d", p.Total)
	}
	if p, _ := fx.svc.Semen.Bank(ctx, BankQuery{StockFilter: FilterAll, IncludeArchived: true}); p.Total != 4 {
		t.Fatalf("archived shown: %d", p.Total)
	}
}

func TestBullSheet(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	bull := newBull(t, fx, "Touro", func(i *AnimalInput) { i.BreedID = "breed-nelore" })
	cow := fx.animal(t, "Vaca", domain.SexFemale, domain.CategoryCow)
	lot, _ := fx.svc.Semen.AddLot(ctx, SemenInput{BullID: bull.ID, Quantity: 12, UnitCost: f64(30)})
	_, _ = fx.svc.Semen.Withdraw(ctx, MovementInput{InventoryID: lot.ID, Quantity: 2, Reason: "insemination"})

	sh, err := fx.svc.Semen.Sheet(ctx, bull.ID, 10)
	if err != nil || sh.TotalDoses != 10 || len(sh.Lots) != 1 || len(sh.Movements) != 2 || sh.StockStatus != StockInStock || sh.Detail.BreedName != "Nelore" {
		t.Fatalf("%+v %v", sh, err)
	}
	if sh.Movements[0].Type != domain.MovWithdrawal {
		t.Fatalf("newest first: %+v", sh.Movements)
	}
	if _, err := fx.svc.Semen.Sheet(ctx, cow.ID, 10); err == nil {
		t.Fatal("a cow has no bull sheet")
	}
	if _, err := fx.svc.Semen.Sheet(ctx, "nope", 10); err == nil {
		t.Fatal("unknown animal")
	}
	if sh, _ := fx.svc.Semen.Sheet(ctx, bull.ID, 1); len(sh.Movements) != 1 {
		t.Fatal("movement limit")
	}
}

// The domain only knows the interface of external catalogues; nothing in the
// application depends on an implementation (offline-first).
type stubProvider struct{}

func (stubProvider) ID() string { return "stub" }
func (stubProvider) SearchBulls(context.Context, string, int) ([]domain.CatalogBull, error) {
	return nil, nil
}
func (stubProvider) GetBull(context.Context, string) (*domain.CatalogBull, error) { return nil, nil }

var _ domain.BullCatalogProvider = stubProvider{}
