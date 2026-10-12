// Package testutil provides helpers shared by tests.
package testutil

import (
	"context"
	"path/filepath"
	"testing"

	"genetica/internal/database"
	"genetica/migrations"
)

// NewDB returns a migrated database in a temporary directory.
func NewDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "farm.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate(context.Background(), migrations.FS); err != nil {
		t.Fatal(err)
	}
	return db
}
