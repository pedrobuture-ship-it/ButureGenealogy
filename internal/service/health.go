// Package service holds the application use-cases exposed to the UI.
package service

import (
	"context"

	"genetica/internal/database"
)

// AppInfo describes the running application.
type AppInfo struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	DataDir  string `json:"dataDir"`
	Platform string `json:"platform"`
}

// HealthService reports on application and database health.
type HealthService struct {
	db *database.DB
}

func NewHealthService(db *database.DB) *HealthService { return &HealthService{db: db} }

func (s *HealthService) Check(ctx context.Context) database.Health { return s.db.Health(ctx) }
