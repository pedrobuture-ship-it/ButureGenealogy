package service

import (
	"context"

	"genetica/internal/apperr"
	"genetica/internal/domain"
)

// AnimalRef is a compact reference to another animal.
type AnimalRef struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Registration string `json:"registration"`
	Sex          string `json:"sex"`
	BreedName    string `json:"breedName"`
}

// AnimalDetail is everything the animal page needs to render the summary tab.
type AnimalDetail struct {
	Animal     domain.Animal `json:"animal"`
	BreedName  string        `json:"breedName"`
	FarmName   string        `json:"farmName"`
	Father     *AnimalRef    `json:"father"`
	Mother     *AnimalRef    `json:"mother"`
	Offspring  []AnimalRef   `json:"offspring"`
	SemenDoses int           `json:"semenDoses"`
}

func (s *AnimalService) Detail(ctx context.Context, id string) (*AnimalDetail, error) {
	r := repos(s.db)
	a, err := r.Animals.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	breeds, err := r.Breeds.List(ctx, false)
	if err != nil {
		return nil, err
	}
	breedName := map[string]string{}
	for _, b := range breeds {
		breedName[b.ID] = b.Name
	}
	d := &AnimalDetail{Animal: *a, BreedName: breedName[a.BreedID], Offspring: []AnimalRef{}}
	if f, err := r.Farms.Get(ctx, a.FarmID); err == nil {
		d.FarmName = f.Name
	}
	ref := func(pid string) (*AnimalRef, error) {
		if pid == "" {
			return nil, nil
		}
		p, err := r.Animals.Get(ctx, pid)
		if err != nil {
			if apperr.CodeOf(err) == apperr.CodeNotFound {
				return nil, nil
			}
			return nil, err
		}
		return &AnimalRef{ID: p.ID, Name: p.Name, Registration: p.Registration, Sex: p.Sex, BreedName: breedName[p.BreedID]}, nil
	}
	if d.Father, err = ref(a.FatherID); err != nil {
		return nil, err
	}
	if d.Mother, err = ref(a.MotherID); err != nil {
		return nil, err
	}
	kids, err := r.Animals.Children(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, k := range kids {
		d.Offspring = append(d.Offspring, AnimalRef{ID: k.ID, Name: k.Name, Registration: k.Registration, Sex: k.Sex, BreedName: breedName[k.BreedID]})
	}
	stock, err := r.Semen.ListByBull(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, inv := range stock {
		d.SemenDoses += inv.Quantity
	}
	return d, nil
}
