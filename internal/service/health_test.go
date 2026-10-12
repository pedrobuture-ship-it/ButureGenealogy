package service

import (
	"context"
	"path/filepath"
	"testing"

	"genetica/internal/database"
	"genetica/migrations"
)

func TestHealthCheck(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "farm.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Migrate(context.Background(), migrations.FS); err != nil {
		t.Fatal(err)
	}
	if h := NewHealthService(db).Check(context.Background()); !h.OK {
		t.Fatalf("expected healthy: %+v", h)
	}
}
