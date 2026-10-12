package database

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"genetica/migrations"
)

func openTemp(t *testing.T) *DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "farm.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestMigrateAndHealth(t *testing.T) {
	db := openTemp(t)
	ctx := context.Background()
	if err := db.Migrate(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	// idempotent
	if err := db.Migrate(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	h := db.Health(ctx)
	if !h.OK || h.SchemaVersion < 1 || !h.ForeignKeys || h.JournalMode != "wal" {
		t.Fatalf("unexpected health: %+v", h)
	}
}

func TestMigrateRollsBackFailedMigration(t *testing.T) {
	db := openTemp(t)
	ctx := context.Background()
	fsys := fstest.MapFS{
		"0001_a.sql": {Data: []byte(`CREATE TABLE a (id INTEGER);`)},
		"0002_b.sql": {Data: []byte(`CREATE TABLE b (id INTEGER); THIS IS INVALID;`)},
	}
	if err := db.Migrate(ctx, fsys); err == nil {
		t.Fatal("expected error")
	}
	if v, _ := db.SchemaVersion(ctx); v != 1 {
		t.Fatalf("want version 1, got %d", v)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE name='b'`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("table b must not exist (n=%d err=%v)", n, err)
	}
}

func TestMigrateRejectsNewerDatabase(t *testing.T) {
	db := openTemp(t)
	ctx := context.Background()
	if err := db.Migrate(ctx, fstest.MapFS{"0002_x.sql": {Data: []byte(`SELECT 1;`)}}); err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(ctx, fstest.MapFS{"0001_x.sql": {Data: []byte(`SELECT 1;`)}}); err == nil {
		t.Fatal("expected error for database newer than app")
	}
}

func TestDuplicateVersion(t *testing.T) {
	_, err := loadMigrations(fstest.MapFS{
		"0001_a.sql": {Data: []byte("")}, "0001_b.sql": {Data: []byte("")},
	})
	if err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestBackupBeforeMigrate(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	now := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)

	// a brand-new database has nothing to protect
	db := openTemp(t)
	if p, err := db.BackupBeforeMigrate(ctx, migrations.FS, dir, now); err != nil || p != "" {
		t.Fatalf("new database: %q %v", p, err)
	}
	// an old database (only 0001-0003 applied) with data is copied before the rest is applied
	old := fstest.MapFS{}
	for _, n := range []string{"0001_init.sql", "0002_domain.sql", "0003_traits.sql"} {
		b, _ := migrations.FS.ReadFile(n)
		old[n] = &fstest.MapFile{Data: b}
	}
	if err := db.Migrate(ctx, old); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO farms (id, name, notes, created_at, updated_at) VALUES ('f', 'Minha Fazenda', '', 't', 't')`); err != nil {
		t.Fatal(err)
	}
	p, err := db.BackupBeforeMigrate(ctx, migrations.FS, dir, now)
	if err != nil || p == "" || !strings.Contains(p, "pre-migration-v3-to-v") || !strings.Contains(p, "20261011-120000") {
		t.Fatalf("%q %v", p, err)
	}
	cp, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer cp.Close()
	var name string
	var v int
	if err := cp.QueryRow(`SELECT name FROM farms WHERE id='f'`).Scan(&name); err != nil || name != "Minha Fazenda" {
		t.Fatalf("the copy must contain the data: %v %q", err, name)
	}
	_ = cp.QueryRow(`SELECT MAX(version) FROM schema_migrations`).Scan(&v)
	if v != 3 {
		t.Fatalf("the copy holds the PRE-migration state (v3), got v%d", v)
	}
	// after migrating there is nothing left to protect
	if err := db.Migrate(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	if p, err := db.BackupBeforeMigrate(ctx, migrations.FS, dir, now); err != nil || p != "" {
		t.Fatalf("up to date: %q %v", p, err)
	}
}
