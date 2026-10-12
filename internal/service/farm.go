package service

import (
	"context"
	"strings"

	"genetica/internal/database"
	"genetica/internal/domain"
)

type FarmService struct{ db *database.DB }

func (s *FarmService) Create(ctx context.Context, in FarmInput) (*domain.Farm, error) {
	now := nowTS()
	f := &domain.Farm{ID: newID(), Name: strings.TrimSpace(in.Name), OwnerName: in.OwnerName, City: in.City,
		State: strings.ToUpper(strings.TrimSpace(in.State)), Notes: in.Notes, CreatedAt: now, UpdatedAt: now}
	if err := f.Validate(); err != nil {
		return nil, err
	}
	if err := repos(s.db).Farms.Insert(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *FarmService) Update(ctx context.Context, id string, in FarmInput) (*domain.Farm, error) {
	r := repos(s.db).Farms
	f, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	f.Name, f.OwnerName, f.City, f.Notes = strings.TrimSpace(in.Name), in.OwnerName, in.City, in.Notes
	f.State = strings.ToUpper(strings.TrimSpace(in.State))
	f.UpdatedAt = nowTS()
	if err := f.Validate(); err != nil {
		return nil, err
	}
	if err := r.Update(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *FarmService) Get(ctx context.Context, id string) (*domain.Farm, error) {
	return repos(s.db).Farms.Get(ctx, id)
}

func (s *FarmService) List(ctx context.Context) ([]domain.Farm, error) {
	return repos(s.db).Farms.List(ctx)
}

type BreedService struct{ db *database.DB }

func (s *BreedService) List(ctx context.Context) ([]domain.Breed, error) {
	return repos(s.db).Breeds.List(ctx, true)
}
