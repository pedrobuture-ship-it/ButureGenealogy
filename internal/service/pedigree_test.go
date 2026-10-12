package service

import (
	"context"
	"testing"

	"genetica/internal/domain"
)

func TestPedigreeViewLoadsGenerations(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	gs := fx.animal(t, "Avô", domain.SexMale, domain.CategoryBull, func(i *AnimalInput) { i.BirthDate = "2010-01-01" })
	gd := fx.animal(t, "Avó", domain.SexFemale, domain.CategoryCow, func(i *AnimalInput) { i.BirthDate = "2010-02-01" })
	sire := fx.animal(t, "Pai", domain.SexMale, domain.CategoryBull, func(i *AnimalInput) { i.FatherID, i.MotherID, i.BirthDate = gs.ID, gd.ID, "2015-01-01" })
	dam := fx.animal(t, "Mãe", domain.SexFemale, domain.CategoryCow, func(i *AnimalInput) { i.FatherID, i.BirthDate = gs.ID, "2016-01-01" }) // same grandsire => collapse
	calf := fx.animal(t, "Bezerro", domain.SexMale, domain.CategoryCalf, func(i *AnimalInput) { i.FatherID, i.MotherID, i.BirthDate = sire.ID, dam.ID, "2022-01-01" })

	v, err := fx.svc.Pedigree.View(ctx, calf.ID, 3)
	if err != nil {
		t.Fatal(err)
	}
	if v.Tree.Sire.Name != "Pai" || v.Tree.Sire.Sire.Name != "Avô" || v.Tree.Dam.Sire.Name != "Avô" {
		t.Fatalf("tree: %+v", v.Tree)
	}
	if v.Tree.Dam.Dam.Known {
		t.Fatal("unknown maternal grandmother must be a placeholder")
	}
	if len(v.Repeated) != 1 || v.Repeated[0].ID != gs.ID {
		t.Fatalf("repeated: %+v", v.Repeated)
	}
	if v.Depth != 2 || len(v.Issues) != 0 {
		t.Fatalf("depth=%d issues=%v", v.Depth, v.Issues)
	}
	if v.Inbreeding.Status != "not_calculated" || v.Inbreeding.Value != nil {
		t.Fatalf("inbreeding must not be invented: %+v", v.Inbreeding)
	}
	// generation clamp
	if v, _ := fx.svc.Pedigree.View(ctx, calf.ID, 99); v.Generations != MaxGenerations {
		t.Fatalf("generations=%d", v.Generations)
	}
	// 2 generations must not expose great-grandparents
	v2, _ := fx.svc.Pedigree.View(ctx, calf.ID, 2)
	if v2.Tree.Sire.Sire.Sire != nil {
		t.Fatal("tree deeper than requested")
	}

	ca, err := fx.svc.Pedigree.CommonAncestors(ctx, sire.ID, dam.ID, 3)
	if err != nil || len(ca.Ancestors) != 1 || ca.Ancestors[0].Name != "Avô" {
		t.Fatalf("%+v %v", ca, err)
	}
	if ca.Relationship.Status != "not_calculated" {
		t.Fatal("relationship must be not_calculated")
	}
}

func TestCreateMinimalAncestor(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	calf := fx.animal(t, "Bezerra", domain.SexFemale, domain.CategoryCalf, func(i *AnimalInput) { i.BirthDate = "2024-01-01" })

	sire, err := fx.svc.Pedigree.CreateMinimalAncestor(ctx, MinimalAncestorInput{ChildID: calf.ID, Role: "sire", Name: "Touro X"})
	if err != nil || sire.Sex != domain.SexMale || sire.FarmID != calf.FarmID {
		t.Fatalf("%+v %v", sire, err)
	}
	got, _ := fx.svc.Animals.Get(ctx, calf.ID)
	if got.FatherID != sire.ID {
		t.Fatal("child not linked to new sire")
	}
	// the slot is now taken
	if _, err := fx.svc.Pedigree.CreateMinimalAncestor(ctx, MinimalAncestorInput{ChildID: calf.ID, Role: "sire", Name: "Outro"}); err == nil {
		t.Fatal("must refuse to overwrite an existing parent")
	}
	// must have identification
	if _, err := fx.svc.Pedigree.CreateMinimalAncestor(ctx, MinimalAncestorInput{ChildID: calf.ID, Role: "dam"}); err == nil {
		t.Fatal("name or registration required")
	}
	got, _ = fx.svc.Animals.Get(ctx, calf.ID)
	if got.MotherID != "" {
		t.Fatal("failed creation must not link anything")
	}
	if _, err := fx.svc.Pedigree.CreateMinimalAncestor(ctx, MinimalAncestorInput{ChildID: calf.ID, Role: "dam", Registration: "M-1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := fx.svc.Pedigree.CreateMinimalAncestor(ctx, MinimalAncestorInput{ChildID: calf.ID, Role: "x", Name: "y"}); err == nil {
		t.Fatal("invalid role")
	}
}

func TestAuditFindsCorruptedPedigree(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	a := fx.animal(t, "A", domain.SexMale, domain.CategoryBull)
	b := fx.animal(t, "B", domain.SexMale, domain.CategoryBull, func(i *AnimalInput) { i.FatherID = a.ID })
	// bypass the service guard to simulate bad imported data: A's father becomes B
	if _, err := fx.svc.Pedigree.db.ExecContext(ctx, `UPDATE animals SET father_id=? WHERE id=?`, b.ID, a.ID); err != nil {
		t.Fatal(err)
	}
	issues, err := fx.svc.Pedigree.Audit(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	cycle := false
	for _, i := range issues {
		cycle = cycle || i.Code == "cycle"
	}
	if !cycle {
		t.Fatalf("cycle not reported: %+v", issues)
	}
	// and the viewer must still terminate
	if _, err := fx.svc.Pedigree.View(ctx, a.ID, 5); err != nil {
		t.Fatal(err)
	}
}
