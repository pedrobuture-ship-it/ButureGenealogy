package backup_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"genetica/internal/backup"
	"genetica/internal/database"
	"genetica/internal/datasets"
	"genetica/migrations"
)

func codes(r *backup.Report) map[string]string {
	m := map[string]string{}
	for _, i := range r.Issues {
		m[i.Code] = i.Severity
	}
	return m
}

func check(t *testing.T, e env) *backup.Report {
	t.Helper()
	r, err := e.bk.CheckIntegrity(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestIntegrityCleanData(t *testing.T) {
	e := newEnv(t)
	populate(t, e, datasets.License{})
	r := check(t, e)
	if !r.OK || r.Errors != 0 || r.Warnings != 0 || len(r.Issues) != 0 || r.Checked["animals"] != 3 || r.Checked["libraries"] != 1 {
		t.Fatalf("%+v", r)
	}
}

func TestIntegrityFindsEachKindOfProblem(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t)
	farmID, bullID, cowID := populate(t, e, datasets.License{})
	_ = farmID
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := e.db.ExecContext(ctx, q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	// library file changed / missing / orphan
	libs, _ := e.svc.Libraries.List(ctx)
	pkg := filepath.Join(e.libDir, libs[0].ID+".genlib")
	if err := os.WriteFile(filepath.Join(e.libDir, "00000000-0000-0000-0000-000000000000.genlib"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(pkg)
	_ = os.WriteFile(pkg, append(raw, 'x'), 0o600)
	c := codes(check(t, e))
	if c["dataset_file_changed"] != "error" || c["orphan_library_file"] != "warning" {
		t.Fatalf("%v", c)
	}
	_ = os.Remove(pkg)
	if codes(check(t, e))["dataset_file_missing"] != "warning" {
		t.Fatal("a missing library package is a warning (the values are still in the database)")
	}

	// calculation pointing at a library that is gone (snapshot keeps the values)
	exec(`INSERT INTO matings (id, farm_id, female_id, bull_id, status, score, engine_version, calculation_snapshot, created_at)
		VALUES ('m1', ?, ?, ?, 'draft', 70, '0.1.0', '{"datasets":[{"id":"gone","name":"Biblioteca Antiga","version":"9"}]}', 't')`, farmID, cowID, bullID)
	r := check(t, e)
	found := false
	for _, i := range r.Issues {
		if i.Code == "calculation_dataset_removed" && i.Severity == "warning" {
			found = true
		}
	}
	if !found || !r.OK {
		t.Fatalf("a historical calculation whose dataset was removed is reported but is not an error: %+v", r.Issues)
	}

	// stock that disagrees with the movement history
	exec(`UPDATE semen_inventory SET quantity = quantity + 5`)
	if codes(check(t, e))["stock_mismatch"] != "warning" {
		t.Fatal("stock_mismatch")
	}

	// pedigree cycle
	exec(`UPDATE animals SET father_id=? WHERE id=?`, bullID, cowID)
	exec(`UPDATE animals SET father_id=? WHERE id=?`, cowID, bullID)
	if codes(check(t, e))["pedigree_cycle"] != "warning" {
		t.Fatalf("%v", codes(check(t, e)))
	}

	// broken reference (foreign keys disabled to force it)
	exec(`PRAGMA foreign_keys=OFF`)
	exec(`INSERT INTO animals (id, farm_id, name, sex, category, status, notes, created_at, updated_at) VALUES ('orph', 'no-farm', 'X', 'male', 'bull', 'active', '', 't', 't')`)
	exec(`PRAGMA foreign_keys=ON`)
	r = check(t, e)
	if codes(r)["broken_reference"] != "error" || r.OK {
		t.Fatalf("%+v", r)
	}
	exec(`DELETE FROM animals WHERE id='orph'`)

	// incomplete migration / gap / future version / missing table
	exec(`DELETE FROM schema_migrations WHERE version = 3`)
	if codes(check(t, e))["migration_gap"] != "error" {
		t.Fatal("migration_gap")
	}
	exec(`DELETE FROM schema_migrations WHERE version = (SELECT MAX(version) FROM schema_migrations)`)
	if codes(check(t, e))["migration_incomplete"] != "error" {
		t.Fatal("migration_incomplete")
	}
	exec(`INSERT INTO schema_migrations (version, name) VALUES (999, 'future')`)
	if codes(check(t, e))["version_incompatible"] != "error" {
		t.Fatal("version_incompatible")
	}
	exec(`DROP TABLE semen_types`)
	if codes(check(t, e))["table_missing"] != "error" {
		t.Fatal("table_missing")
	}
}

func TestIntegrityDetectsACorruptedSQLiteFile(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "farm.db")
	db, _ := database.Open(path)
	_ = db.Migrate(ctx, migrations.FS)
	for i := 0; i < 200; i++ {
		_, _ = db.Exec(`INSERT INTO farms (id, name, notes, created_at, updated_at) VALUES (?, ?, ?, 't', 't')`, "f"+string(rune('a'+i%26))+string(rune('a'+i/26)), "Fazenda número "+string(rune('a'+i%26)), "nota longa para ocupar espaço nas páginas do banco de dados")
	}
	_, _ = db.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`)
	_ = db.Close()
	raw, _ := os.ReadFile(path)
	for i := 4096 * 2; i < len(raw) && i < 4096*2+3000; i++ {
		raw[i] = 0xFF
	}
	_ = os.WriteFile(path, raw, 0o600)

	cdb, err := database.Open(path)
	if err != nil {
		t.Skipf("sqlite refuses to open the damaged file at all (also a detection): %v", err)
	}
	defer cdb.Close()
	bk := &backup.Service{DB: cdb, DataDir: dir, LibrariesDir: filepath.Join(dir, "libraries"), Migrations: migrations.FS}
	r, err := bk.CheckIntegrity(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if r.OK || codes(r)["sqlite_corrupt"] != "error" {
		t.Fatalf("a corrupted database must be reported: %+v", r.Issues)
	}
}
