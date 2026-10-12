package service

import (
	"context"
	"database/sql"
	"encoding/json"

	"genetica/internal/apperr"
	"genetica/internal/database"
	"genetica/internal/domain"
	"genetica/internal/repository"
)

// DemoService loads and removes DEMO / SYNTHETIC data. Nothing here is a
// scientific value: it exists only to exercise the interface and the engine.
type DemoService struct{ db *database.DB }

const demoFarmName = "[" + domain.DemoMarker + "] Fazenda Exemplo"

func (s *DemoService) HasDemo(ctx context.Context) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM farms WHERE is_demo=1`).Scan(&n)
	return n > 0, err
}

type demoAnimal struct {
	key, name, sex, cat, breed, birth, father, mother string
}

type demoTrait struct{ code, name, kind, category, unit string }

type demoEval struct {
	key, trait, kind string
	value, acc, pct  float64
}

// Load inserts the synthetic herd. It refuses to run twice.
func (s *DemoService) Load(ctx context.Context) error {
	has, err := s.HasDemo(ctx)
	if err != nil {
		return err
	}
	if has {
		return apperr.New(apperr.CodeConflict, "Os dados de demonstração já foram carregados.")
	}
	herd := []demoAnimal{
		{"bullA", "DEMO Touro Alfa", domain.SexMale, domain.CategoryBull, "breed-nelore", "2016-03-10", "", ""},
		{"bullB", "DEMO Touro Beta", domain.SexMale, domain.CategoryBull, "breed-nelore", "2017-08-22", "", ""},
		{"bullG", "DEMO Touro Gama", domain.SexMale, domain.CategoryBull, "breed-nelore", "2018-05-02", "", ""},
		{"cow1", "DEMO Matriz 0472", domain.SexFemale, domain.CategoryCow, "breed-nelore", "2017-02-05", "", ""},
		{"cow2", "DEMO Matriz 0518", domain.SexFemale, domain.CategoryCow, "breed-nelore", "2018-04-17", "", ""},
		{"cow3", "DEMO Matriz 0603", domain.SexFemale, domain.CategoryCow, "breed-angus", "2019-06-30", "", ""},
		{"heifer1", "DEMO Novilha 1101", domain.SexFemale, domain.CategoryHeifer, "breed-nelore", "2022-09-12", "bullA", "cow1"},
		{"heifer2", "DEMO Novilha 1102", domain.SexFemale, domain.CategoryHeifer, "breed-nelore", "2022-11-03", "bullB", "cow2"},
		{"calf1", "DEMO Bezerro 2301", domain.SexMale, domain.CategoryCalf, "breed-nelore", "2025-12-01", "bullA", "heifer1"},
		{"steer1", "DEMO Boi 0901", domain.SexMale, domain.CategorySteer, "breed-nelore", "2024-01-20", "bullA", "cow2"},
		{"bullH1", "DEMO Touro Holandês H1", domain.SexMale, domain.CategoryBull, "breed-holstein", "2017-01-15", "", ""},
		{"bullH2", "DEMO Touro Holandês H2", domain.SexMale, domain.CategoryBull, "breed-holstein", "2018-02-20", "", ""},
		{"cowV1", "DEMO Vaca Holandesa V1", domain.SexFemale, domain.CategoryCow, "breed-holstein", "2019-03-11", "", ""},
	}
	// Synthetic phenotype measurements (not real data).
	meas := []struct {
		key, trait, unit, date string
		v                      float64
	}{
		{"cow1", "BODY_WEIGHT", "kg", "2025-06-01", 512}, {"cow1", "HIP_HEIGHT", "cm", "2025-06-01", 138},
		{"cow2", "BODY_WEIGHT", "kg", "2025-06-01", 498}, {"bullA", "BODY_WEIGHT", "kg", "2025-06-01", 905},
		{"bullA", "SCROTAL_CIRC", "cm", "2025-06-01", 37},
	}
	traitsA := []demoTrait{
		{"WW", "Peso à desmama", "DEP", domain.CatGrowth, "kg"}, {"YW", "Peso ao sobreano", "DEP", domain.CatGrowth, "kg"},
		{"MILK", "Habilidade materna (leite)", "DEP", domain.CatMaternal, "kg"},
		{"MY", "Produção de leite", "PTA", domain.CatDairy, "kg"}, {"FERT", "Fertilidade das filhas", "PTA", domain.CatReproduction, ""},
		{"SCS", "Escore de células somáticas", "PTA", domain.CatHealth, ""},
	}
	traitsB := traitsA[:3]
	// All values below are invented. Percentiles are "ascending": higher value, higher percentile.
	evalsA := []demoEval{
		{"bullA", "WW", "DEP", 6.1, 0.88, 82}, {"bullA", "YW", "DEP", 9.4, 0.85, 78}, {"bullA", "MILK", "DEP", 1.2, 0.74, 55},
		{"bullB", "WW", "DEP", 3.8, 0.81, 60}, {"bullB", "YW", "DEP", 7.0, 0.79, 64}, {"bullB", "MILK", "DEP", 3.1, 0.70, 84},
		{"cow1", "WW", "DEP", 2.2, 0.55, 58}, {"cow1", "YW", "DEP", 4.0, 0.50, 52}, {"cow1", "MILK", "DEP", 2.5, 0.52, 71},
		{"cow2", "WW", "DEP", 4.0, 0.57, 70}, {"cow2", "MILK", "DEP", 0.8, 0.50, 40},
		{"bullH1", "MY", "PTA", 620, 0.90, 85}, {"bullH1", "FERT", "PTA", 1.1, 0.80, 60}, {"bullH1", "SCS", "PTA", 2.7, 0.78, 30},
		{"bullH2", "MY", "PTA", 410, 0.55, 60}, {"bullH2", "FERT", "PTA", 2.4, 0.52, 80}, {"bullH2", "SCS", "PTA", 2.9, 0.50, 55},
		{"cowV1", "MY", "PTA", 150, 0.45, 40}, {"cowV1", "FERT", "PTA", 0.4, 0.40, 50}, {"cowV1", "SCS", "PTA", 3.1, 0.42, 70},
	}
	evalsB := []demoEval{
		{"bullG", "WW", "DEP", 5.0, 0.90, 90}, {"bullG", "YW", "DEP", 8.0, 0.88, 85}, {"bullG", "MILK", "DEP", 2.0, 0.80, 70},
	}
	stock := []struct {
		key string
		qty int
	}{{"bullA", 40}, {"bullB", 25}, {"bullG", 15}, {"bullH1", 30}, {"bullH2", 12}}

	return s.db.WithTx(ctx, func(tx *sql.Tx) error {
		r := repository.New(tx)
		now := nowTS()
		farm := &domain.Farm{ID: newID(), Name: demoFarmName, City: "Demonstração", State: "XX", IsDemo: true,
			Notes: domain.DemoMarker + ": dados fictícios para testar a interface. Não representam animais reais.", CreatedAt: now, UpdatedAt: now}
		if err := r.Farms.Insert(ctx, farm); err != nil {
			return err
		}
		ids := map[string]string{}
		for _, d := range herd {
			a := &domain.Animal{ID: newID(), FarmID: farm.ID, Name: d.name, Sex: d.sex, Category: d.cat, BreedID: d.breed,
				BirthDate: d.birth, Status: domain.StatusActive, FatherID: ids[d.father], MotherID: ids[d.mother],
				Notes: domain.DemoMarker, IsDemo: true, CreatedAt: now, UpdatedAt: now}
			if err := r.Animals.Insert(ctx, a); err != nil {
				return err
			}
			ids[d.key] = a.ID
		}
		for _, m := range meas {
			if err := r.Measurements.Insert(ctx, &domain.Measurement{ID: newID(), AnimalID: ids[m.key], TraitCode: m.trait, Value: m.v,
				Unit: m.unit, MeasuredAt: m.date, Source: domain.DemoMarker, IsDemo: true, CreatedAt: now}); err != nil {
				return err
			}
		}

		// Two DEMO libraries from two different "programs": values from different
		// programs must never be compared, which the engine demonstrates with bull Gama.
		type lib struct {
			pkg, program, release, base string
			traits                      []demoTrait
			evals                       []demoEval
			records                     int
		}
		for _, l := range []lib{
			{"demo.synthetic", domain.DemoMarker, "demo-1", "DEMO base", traitsA, evalsA, 9},
			{"demo.synthetic-b", domain.DemoMarker + " (programa B)", "demo-b1", "DEMO base B", traitsB, evalsB, 1},
		} {
			datasetID := newID()
			name := domain.DemoMarker + " Avaliações"
			if l.pkg == "demo.synthetic-b" {
				name += " B"
			}
			manifest, _ := json.Marshal(map[string]any{"schemaVersion": 1, "id": l.pkg, "name": name, "version": l.release})
			if _, err := tx.ExecContext(ctx, `INSERT INTO datasets (id, package_id, name, version, is_demo, installed_at, origin, manifest, species,
				breeds, purposes, organization, program, release, license_name, license_status, redistribution_allowed, readme, license_text,
				record_count, trait_count, active, responsibility_accepted)
				VALUES (?, ?, ?, ?, 1, ?, 'csv', ?, 'Bos spp.', '["NEL","HOL"]', '["beef","dairy"]', ?, ?, ?,
				        'Dados sintéticos do aplicativo', 'redistributable', 1, ?, ?, ?, ?, 1, 1)`,
				datasetID, l.pkg, name, l.release, now, string(manifest), domain.DemoMarker, l.program, l.release,
				domain.DemoMarker+": valores inventados apenas para testar a interface e o motor. Não representam animais nem avaliações reais.",
				"Dados sintéticos gerados pelo próprio aplicativo; sem valor científico.", l.records, len(l.traits)); err != nil {
				return apperr.Wrap(apperr.CodeDatabase, "Falha ao criar a biblioteca de demonstração.", err)
			}
			unit := map[string]string{}
			for _, t := range l.traits {
				tr := &domain.Trait{ID: newID(), Code: t.code, Name: t.name, Description: domain.DemoMarker, Category: t.category, Purpose: "any", Unit: t.unit,
					DataType: domain.DataNumber, EvaluationType: t.kind, OptimizationDirection: domain.InformationOnly, SourceDataset: datasetID,
					Metadata: `{"percentileSemantics":"ascending","demo":true}`, CreatedAt: now}
				if err := r.Traits.Insert(ctx, tr); err != nil {
					return err
				}
				unit[t.code] = t.unit
			}
			for _, e := range l.evals {
				acc, pct := e.acc, e.pct
				if err := r.Evaluations.Insert(ctx, &domain.GeneticEvaluation{ID: newID(), AnimalID: ids[e.key], DatasetID: datasetID,
					EvaluationProgram: l.program, EvaluationRelease: l.release, BreedCode: demoBreedCode(e.trait),
					TraitCode: e.trait, EvaluationType: e.kind, Value: e.value, Unit: unit[e.trait], Accuracy: &acc, Percentile: &pct,
					GeneticBase: l.base, SourceReference: domain.DemoMarker, IsDemo: true, CreatedAt: now}); err != nil {
					return err
				}
			}
		}

		cost := 45.0
		for _, st := range stock {
			lot := &domain.SemenInventory{ID: newID(), BullID: ids[st.key], Supplier: "DEMO Central", Batch: "DEMO-001",
				Quantity: st.qty, UnitCost: &cost, SemenType: domain.SemenConventional, Notes: domain.DemoMarker, IsDemo: true,
				CreatedAt: now, UpdatedAt: now}
			if err := r.Semen.Insert(ctx, lot); err != nil {
				return err
			}
			if err := r.Semen.InsertMovement(ctx, &domain.SemenMovement{ID: newID(), InventoryID: lot.ID, BullID: lot.BullID, Type: domain.MovEntry,
				Delta: lot.Quantity, BalanceAfter: lot.Quantity, Reason: "purchase", Note: domain.DemoMarker, IsDemo: true, CreatedAt: now}); err != nil {
				return err
			}
		}
		return nil
	})
}

// Remove deletes every row flagged is_demo, leaving real data untouched.
func (s *DemoService) Remove(ctx context.Context) error {
	return s.db.WithTx(ctx, func(tx *sql.Tx) error {
		stmts := []string{
			`DELETE FROM calculation_datasets WHERE mating_id IN (SELECT id FROM matings WHERE is_demo=1)`,
			`DELETE FROM matings WHERE is_demo=1 OR farm_id IN (SELECT id FROM farms WHERE is_demo=1)`,
			`DELETE FROM calculation_runs WHERE is_demo=1 OR farm_id IN (SELECT id FROM farms WHERE is_demo=1)`,
			`DELETE FROM semen_movements WHERE is_demo=1`,
			`DELETE FROM semen_inventory WHERE is_demo=1`,
			`DELETE FROM genetic_evaluations WHERE is_demo=1`,
			`DELETE FROM animal_measurements WHERE is_demo=1`,
			`UPDATE animals SET father_id=NULL, mother_id=NULL WHERE is_demo=1`,
			`DELETE FROM animals WHERE is_demo=1`,
			`DELETE FROM farms WHERE is_demo=1`,
			`DELETE FROM traits WHERE source_dataset IN (SELECT id FROM datasets WHERE is_demo=1)`,
			`DELETE FROM datasets WHERE is_demo=1`,
		}
		for _, q := range stmts {
			if _, err := tx.ExecContext(ctx, q); err != nil {
				return apperr.Wrap(apperr.CodeConflict, "Não foi possível remover os dados de demonstração (há registros reais vinculados).", err)
			}
		}
		return nil
	})
}

// demoBreedCode: the dairy demo traits belong to Holstein, the beef ones to Nelore.
func demoBreedCode(trait string) string {
	switch trait {
	case "MY", "FERT", "SCS":
		return "HOL"
	}
	return "NEL"
}
