package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"genetica/internal/domain"
)

// TestLargeHerdStaysResponsive guards the listing paths against full scans and N+1 queries:
// 20 000 animals with 8 000 bulls holding semen.
func TestLargeHerdStaysResponsive(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()
	fx := newFixture(t)
	now := nowTS()
	tx, err := fx.svc.Animals.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20000; i++ {
		sex, cat := domain.SexFemale, domain.CategoryCow
		if i%5 < 2 {
			sex, cat = domain.SexMale, domain.CategoryBull
		}
		id := fmt.Sprintf("a%06d", i)
		if _, err := tx.ExecContext(ctx, `INSERT INTO animals (id, farm_id, registration, name, sex, breed_id, birth_date, category, status, notes, created_at, updated_at)
			VALUES (?,?,?,?,?,?,?,?,'active','',?,?)`, id, fx.farmID, fmt.Sprintf("REG-%06d", i), fmt.Sprintf("Animal %d", i), sex, "breed-nelore", fmt.Sprintf("20%02d-01-15", 10+i%15), cat, now, now); err != nil {
			t.Fatal(err)
		}
		if sex == domain.SexMale {
			lotID := "l" + id
			if _, err := tx.ExecContext(ctx, `INSERT INTO semen_inventory (id, bull_id, supplier, batch, quantity, unit_cost, semen_type, notes, created_at, updated_at)
				VALUES (?,?,?,?,?,?, 'conventional', '', ?, ?)`, lotID, id, "Central", "L1", 5+i%30, 40.0+float64(i%20), now, now); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	measure := func(name string, max time.Duration, fn func() error) {
		t.Helper()
		start := time.Now()
		if err := fn(); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		d := time.Since(start)
		t.Logf("%-28s %v", name, d)
		if d > max {
			t.Fatalf("%s took %v (limit %v)", name, d, max)
		}
	}
	measure("list page 1", 300*time.Millisecond, func() error {
		p, err := fx.svc.Animals.List(ctx, AnimalQuery{PageSize: 50})
		if err == nil && (p.Total != 20000 || len(p.Items) != 50) {
			t.Fatalf("%d %d", p.Total, len(p.Items))
		}
		return err
	})
	measure("search by registration", 300*time.Millisecond, func() error {
		p, err := fx.svc.Animals.List(ctx, AnimalQuery{Search: "REG-01999", PageSize: 50})
		if err == nil && p.Total != 10 {
			t.Fatalf("%d", p.Total)
		}
		return err
	})
	measure("filter + sort + deep page", 400*time.Millisecond, func() error {
		_, err := fx.svc.Animals.List(ctx, AnimalQuery{Sex: "male", BreedID: "breed-nelore", SortBy: "birthDate", SortDesc: true, Page: 100, PageSize: 50})
		return err
	})
	measure("semen bank (8 000 bulls)", 1500*time.Millisecond, func() error {
		p, err := fx.svc.Semen.Bank(ctx, BankQuery{PageSize: 50, SortBy: "doses", SortDesc: true})
		if err == nil && p.Total != 8000 {
			t.Fatalf("%d", p.Total)
		}
		return err
	})
	measure("pedigree view", 200*time.Millisecond, func() error {
		_, err := fx.svc.Pedigree.View(ctx, "a000001", 5)
		return err
	})
}
