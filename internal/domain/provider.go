package domain

import "context"

// CatalogBull is a bull as published by an external catalogue.
type CatalogBull struct {
	Source       string `json:"source"`     // provider id
	ExternalID   string `json:"externalId"` // id inside the provider
	Name         string `json:"name"`
	Registration string `json:"registration"`
	BreedCode    string `json:"breedCode"`
}

// BullCatalogProvider is the extension point for future external catalogues
// (semen suppliers, breed associations). The domain depends only on this
// interface: no provider is implemented, no network access exists, and the
// application never depends on a provider to work offline.
type BullCatalogProvider interface {
	// ID is a stable identifier of the provider.
	ID() string
	SearchBulls(ctx context.Context, query string, limit int) ([]CatalogBull, error)
	GetBull(ctx context.Context, externalID string) (*CatalogBull, error)
}
