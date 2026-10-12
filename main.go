package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"genetica/internal/config"
	"genetica/internal/database"
	"genetica/internal/logger"
	"genetica/migrations"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "erro fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	log, closeLog, err := logger.New(cfg.LogDir, cfg.LogLevel)
	if err != nil {
		return err
	}
	defer closeLog.Close()

	db, err := database.Open(cfg.DBPath)
	if err != nil {
		log.Error("cannot open the database", "path", cfg.DBPath, "err", err)
		return err
	}
	// Existing data is copied before pending migrations touch it; a failed migration leaves that copy behind.
	if saved, err := db.BackupBeforeMigrate(context.Background(), migrations.FS, filepath.Join(cfg.DataDir, "backups"), time.Now()); err != nil {
		log.Error("cannot back up the database before migrating; aborting to protect the data", "err", err)
		_ = db.Close()
		return err
	} else if saved != "" {
		log.Info("database copied before migration", "file", saved)
	}
	if err := db.Migrate(context.Background(), migrations.FS); err != nil {
		log.Error("migration failed; the previous state was kept", "err", err)
		_ = db.Close()
		return err
	}

	app := NewApp(cfg, log, db)
	err = wails.Run(&options.App{
		Title:     "Genética",
		Width:     1280,
		Height:    800,
		MinWidth:  1024,
		MinHeight: 640,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 246, G: 247, B: 244, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind:             []interface{}{app},
	})
	if err != nil {
		log.Error("wails run", "err", err)
	}
	return err
}
