// Package database opens SQLite safely, applies migrations and reports health.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"path/filepath"

	_ "modernc.org/sqlite" // pure-Go driver, no CGO

	"genetica/internal/apperr"
)

// DB wraps *sql.DB.
type DB struct {
	*sql.DB
	path string
}

// Open opens (creating if needed) the SQLite file with safe pragmas:
// foreign keys on, WAL journal, busy timeout. It does not run migrations.
func Open(path string) (*DB, error) {
	q := url.Values{}
	q.Add("_pragma", "foreign_keys(1)")
	q.Add("_pragma", "journal_mode(WAL)")
	q.Add("_pragma", "busy_timeout(5000)")
	q.Add("_pragma", "synchronous(NORMAL)")
	dsn := (&url.URL{Scheme: "file", Opaque: filepath.ToSlash(path), RawQuery: q.Encode()}).String()

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDatabase, "Não foi possível abrir o banco de dados.", err)
	}
	// SQLite allows a single writer; one connection avoids SQLITE_BUSY surprises
	// and keeps per-connection pragmas consistent.
	sqlDB.SetMaxOpenConns(1)
	if err := sqlDB.PingContext(context.Background()); err != nil {
		_ = sqlDB.Close()
		return nil, apperr.Wrap(apperr.CodeDatabase, "Não foi possível acessar o banco de dados.", err)
	}
	return &DB{DB: sqlDB, path: path}, nil
}

// Path returns the database file path.
func (d *DB) Path() string { return d.path }

// Close checkpoints the WAL and closes the connection.
func (d *DB) Close() error {
	_, _ = d.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	if err := d.DB.Close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}
	return nil
}

// WithTx runs fn inside a transaction, committing on nil and rolling back on error or panic.
func (d *DB) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) (err error) {
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return apperr.Wrap(apperr.CodeDatabase, "Não foi possível iniciar a transação.", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if err = fn(tx); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return apperr.Wrap(apperr.CodeDatabase, "Falha ao confirmar a transação.", err)
	}
	return nil
}
