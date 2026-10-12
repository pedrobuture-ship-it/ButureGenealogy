// Command seeddemo loads the DEMO / SYNTHETIC herd into a data directory.
// Development aid only: GENETICA_DATA_DIR=/tmp/x go run ./cmd/seeddemo
package main

import (
	"context"
	"fmt"
	"os"

	"genetica/internal/config"
	"genetica/internal/database"
	"genetica/internal/service"
	"genetica/migrations"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	db, err := database.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx := context.Background()
	if err := db.Migrate(ctx, migrations.FS); err != nil {
		return err
	}
	if err := service.New(db, cfg.DataDir).Demo.Load(ctx); err != nil {
		return err
	}
	fmt.Println("dados DEMO / SYNTHETIC carregados em", cfg.DBPath)
	return nil
}
