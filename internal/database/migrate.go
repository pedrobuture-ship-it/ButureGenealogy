package database

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"genetica/internal/apperr"
)

var migrationName = regexp.MustCompile(`^(\d{4})_[a-z0-9_]+\.sql$`)

type migration struct {
	version int
	name    string
	sql     string
}

func loadMigrations(fsys fs.FS) ([]migration, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}
	var out []migration
	seen := map[int]string{}
	for _, e := range entries {
		m := migrationName.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		v, _ := strconv.Atoi(m[1])
		if prev, dup := seen[v]; dup {
			return nil, fmt.Errorf("duplicate migration version %d: %s and %s", v, prev, e.Name())
		}
		seen[v] = e.Name()
		b, err := fs.ReadFile(fsys, e.Name())
		if err != nil {
			return nil, err
		}
		out = append(out, migration{version: v, name: e.Name(), sql: string(b)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].version < out[j].version })
	return out, nil
}

// Migrate applies pending migrations in order. Each runs in its own
// transaction together with its schema_migrations record, so a failure
// leaves the database at the last fully applied version.
func (d *DB) Migrate(ctx context.Context, fsys fs.FS) error {
	migs, err := loadMigrations(fsys)
	if err != nil {
		return apperr.Wrap(apperr.CodeDatabase, "Falha ao ler as migrations.", err)
	}
	if _, err := d.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		name       TEXT NOT NULL,
		applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	)`); err != nil {
		return apperr.Wrap(apperr.CodeDatabase, "Falha ao preparar o controle de migrations.", err)
	}
	current, err := d.SchemaVersion(ctx)
	if err != nil {
		return err
	}
	if known := len(migs); known > 0 && current > migs[known-1].version {
		return apperr.New(apperr.CodeConflict,
			fmt.Sprintf("O banco de dados (versão %d) é mais novo que este aplicativo (versão %d). Atualize o aplicativo.", current, migs[known-1].version))
	}
	for _, m := range migs {
		if m.version <= current {
			continue
		}
		if err := d.apply(ctx, m); err != nil {
			return apperr.Wrap(apperr.CodeDatabase, fmt.Sprintf("Falha ao aplicar a migration %s.", m.name), err)
		}
	}
	return nil
}

func (d *DB) apply(ctx context.Context, m migration) error {
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, m.sql); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version, name) VALUES (?, ?)`, m.version, m.name); err != nil {
		return err
	}
	return tx.Commit()
}

// SchemaVersion returns the highest applied migration version (0 if none).
func (d *DB) SchemaVersion(ctx context.Context) (int, error) {
	var v int
	err := d.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&v)
	if err != nil {
		return 0, apperr.Wrap(apperr.CodeDatabase, "Falha ao ler a versão do esquema.", err)
	}
	return v, nil
}

// BackupBeforeMigrate keeps a consistent copy of an existing database before
// pending migrations change it. It returns the copy's path, or "" when there is
// nothing to protect (new database) or nothing to apply (already up to date).
func (d *DB) BackupBeforeMigrate(ctx context.Context, fsys fs.FS, dir string, now time.Time) (string, error) {
	var tables int
	if err := d.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'`).Scan(&tables); err != nil || tables == 0 {
		return "", err
	}
	cur, err := d.SchemaVersion(ctx)
	if err != nil || cur == 0 {
		return "", err
	}
	migs, err := loadMigrations(fsys)
	if err != nil || len(migs) == 0 || cur >= migs[len(migs)-1].version {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	dest := filepath.Join(dir, fmt.Sprintf("pre-migration-v%d-to-v%d-%s.db", cur, migs[len(migs)-1].version, now.UTC().Format("20060102-150405")))
	if _, err := d.ExecContext(ctx, "VACUUM INTO '"+strings.ReplaceAll(dest, "'", "''")+"'"); err != nil {
		return "", apperr.Wrap(apperr.CodeDatabase, "Não foi possível guardar uma cópia do banco antes de atualizá-lo.", err)
	}
	return dest, nil
}
