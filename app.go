package main

import (
	"context"
	"log/slog"
	"path/filepath"
	"runtime"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"genetica/internal/backup"
	"genetica/internal/config"
	"genetica/internal/database"
	"genetica/internal/datasets"
	"genetica/internal/domain"
	"genetica/internal/engine"
	"genetica/internal/service"
	"genetica/migrations"
)

const appVersion = "0.1.0"

// App is the thin Wails binding layer. It holds no business rules: every
// method delegates to a service in internal/service.
type App struct {
	ctx context.Context
	cfg *config.Config
	log *slog.Logger
	db  *database.DB
	svc *service.Services
	bk  *backup.Service
}

func NewApp(cfg *config.Config, log *slog.Logger, db *database.DB) *App {
	libs := filepath.Join(cfg.DataDir, "libraries")
	return &App{cfg: cfg, log: log, db: db, svc: service.New(db, libs),
		bk: &backup.Service{DB: db, DataDir: cfg.DataDir, LibrariesDir: libs, Migrations: migrations.FS, AppVersion: appVersion, EngineVersion: engine.Version}}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.log.Info("application started", "version", appVersion, "dataDir", a.cfg.DataDir)
	// Integrity check in the background: it only reports, it never modifies anything.
	go func() {
		rep, err := a.bk.CheckIntegrity(context.Background())
		switch {
		case err != nil:
			a.log.Error("integrity check failed to run", "err", err)
		case !rep.OK:
			a.log.Error("integrity check found errors", "errors", rep.Errors, "warnings", rep.Warnings)
		default:
			a.log.Info("integrity check", "warnings", rep.Warnings)
		}
	}()
}

// shutdown closes resources when the window is closed.
func (a *App) shutdown(_ context.Context) {
	if err := a.db.Close(); err != nil {
		a.log.Error("closing database", "err", err)
		return
	}
	a.log.Info("database closed")
}

// GetAppInfo returns static information about the application.
func (a *App) GetAppInfo() service.AppInfo {
	return service.AppInfo{Name: config.AppName, Version: appVersion, DataDir: a.cfg.DataDir, Platform: runtime.GOOS}
}

// HealthCheck reports database health.
func (a *App) HealthCheck() database.Health { return a.svc.Health.Check(a.ctx) }

// ---- Propriedades e raças

func (a *App) ListFarms() ([]domain.Farm, error) {
	v, err := a.svc.Farms.List(a.ctx)
	return v, service.Wire(a.log, err)
}

func (a *App) CreateFarm(in service.FarmInput) (*domain.Farm, error) {
	v, err := a.svc.Farms.Create(a.ctx, in)
	return v, service.Wire(a.log, err)
}

func (a *App) UpdateFarm(id string, in service.FarmInput) (*domain.Farm, error) {
	v, err := a.svc.Farms.Update(a.ctx, id, in)
	return v, service.Wire(a.log, err)
}

func (a *App) ListBreeds() ([]domain.Breed, error) {
	v, err := a.svc.Breeds.List(a.ctx)
	return v, service.Wire(a.log, err)
}

// ---- Animais

func (a *App) ListAnimals(q service.AnimalQuery) (*service.AnimalPage, error) {
	v, err := a.svc.Animals.List(a.ctx, q)
	return v, service.Wire(a.log, err)
}

func (a *App) GetAnimal(id string) (*domain.Animal, error) {
	v, err := a.svc.Animals.Get(a.ctx, id)
	return v, service.Wire(a.log, err)
}

func (a *App) CreateAnimal(in service.AnimalInput) (*domain.Animal, error) {
	v, err := a.svc.Animals.Create(a.ctx, in)
	return v, service.Wire(a.log, err)
}

func (a *App) UpdateAnimal(id string, in service.AnimalInput) (*domain.Animal, error) {
	v, err := a.svc.Animals.Update(a.ctx, id, in)
	return v, service.Wire(a.log, err)
}

func (a *App) ArchiveAnimal(id string) error {
	return service.Wire(a.log, a.svc.Animals.Archive(a.ctx, id))
}
func (a *App) RestoreAnimal(id string) error {
	return service.Wire(a.log, a.svc.Animals.Restore(a.ctx, id))
}

// ---- Fenótipo e avaliações genéticas

func (a *App) ListMeasurements(animalID string) ([]domain.Measurement, error) {
	v, err := a.svc.Measurements.ListByAnimal(a.ctx, animalID)
	return v, service.Wire(a.log, err)
}

func (a *App) AddMeasurement(in service.MeasurementInput) (*domain.Measurement, error) {
	v, err := a.svc.Measurements.Add(a.ctx, in)
	return v, service.Wire(a.log, err)
}

func (a *App) DeleteMeasurement(id string) error {
	return service.Wire(a.log, a.svc.Measurements.Delete(a.ctx, id))
}

func (a *App) ListEvaluations(animalID string) ([]domain.GeneticEvaluation, error) {
	v, err := a.svc.Evaluations.ListByAnimal(a.ctx, animalID)
	return v, service.Wire(a.log, err)
}

func (a *App) AddEvaluation(in service.EvaluationInput) (*domain.GeneticEvaluation, error) {
	v, err := a.svc.Evaluations.Add(a.ctx, in)
	return v, service.Wire(a.log, err)
}

func (a *App) DeleteEvaluation(id string) error {
	return service.Wire(a.log, a.svc.Evaluations.Delete(a.ctx, id))
}

func (a *App) GetAnimalDetail(id string) (*service.AnimalDetail, error) {
	v, err := a.svc.Animals.Detail(a.ctx, id)
	return v, service.Wire(a.log, err)
}

// ---- Dados de demonstração (DEMO / SYNTHETIC)

func (a *App) HasDemoData() (bool, error) {
	v, err := a.svc.Demo.HasDemo(a.ctx)
	return v, service.Wire(a.log, err)
}
func (a *App) LoadDemoData() error   { return service.Wire(a.log, a.svc.Demo.Load(a.ctx)) }
func (a *App) RemoveDemoData() error { return service.Wire(a.log, a.svc.Demo.Remove(a.ctx)) }

// ---- Genealogia

func (a *App) GetPedigree(animalID string, generations int) (*service.PedigreeView, error) {
	v, err := a.svc.Pedigree.View(a.ctx, animalID, generations)
	return v, service.Wire(a.log, err)
}

func (a *App) GetCommonAncestors(animalA, animalB string, generations int) (*service.CommonAncestorsResult, error) {
	v, err := a.svc.Pedigree.CommonAncestors(a.ctx, animalA, animalB, generations)
	return v, service.Wire(a.log, err)
}

func (a *App) AuditPedigrees(farmID string) ([]service.AuditIssue, error) {
	v, err := a.svc.Pedigree.Audit(a.ctx, farmID)
	return v, service.Wire(a.log, err)
}

func (a *App) CreateMinimalAncestor(in service.MinimalAncestorInput) (*domain.Animal, error) {
	v, err := a.svc.Pedigree.CreateMinimalAncestor(a.ctx, in)
	return v, service.Wire(a.log, err)
}

// ---- Características (traits)

func (a *App) ListTraits(q service.TraitQuery) ([]domain.Trait, error) {
	v, err := a.svc.Traits.List(a.ctx, q)
	return v, service.Wire(a.log, err)
}

func (a *App) CreateTrait(in service.TraitInput) (*domain.Trait, error) {
	v, err := a.svc.Traits.Create(a.ctx, in)
	return v, service.Wire(a.log, err)
}

func (a *App) ListValueKinds() []domain.ValueKindInfo { return a.svc.Traits.Kinds() }

func (a *App) GetAnimalTraits(animalID string) ([]service.TraitGroup, error) {
	v, err := a.svc.Traits.ForAnimal(a.ctx, animalID)
	return v, service.Wire(a.log, err)
}

// ---- Bibliotecas (datasets)

// PickLibraryFile opens the native file dialog and returns the chosen path ("" if cancelled).
func (a *App) PickLibraryFile() (string, error) {
	path, err := wruntime.OpenFileDialog(a.ctx, wruntime.OpenDialogOptions{
		Title: "Importar biblioteca",
		Filters: []wruntime.FileFilter{
			{DisplayName: "Bibliotecas e dados (*.genlib, *.csv, *.json)", Pattern: "*.genlib;*.csv;*.tsv;*.json"},
		},
	})
	return path, service.Wire(a.log, err)
}

func (a *App) ListLibraries() ([]service.LibrarySummary, error) {
	v, err := a.svc.Libraries.List(a.ctx)
	return v, service.Wire(a.log, err)
}

func (a *App) GetLibrary(id string) (*service.LibraryDetail, error) {
	v, err := a.svc.Libraries.Get(a.ctx, id)
	return v, service.Wire(a.log, err)
}

func (a *App) SetLibraryActive(id string, active bool) error {
	return service.Wire(a.log, a.svc.Libraries.SetActive(a.ctx, id, active))
}

func (a *App) RemoveLibrary(id string) error {
	return service.Wire(a.log, a.svc.Libraries.Remove(a.ctx, id))
}

func (a *App) InspectImport(path string) (*service.ImportInspection, error) {
	v, err := a.svc.Libraries.Inspect(a.ctx, path)
	return v, service.Wire(a.log, err)
}

func (a *App) ValidateImport(req service.ImportRequest) (*service.ImportValidation, error) {
	v, err := a.svc.Libraries.Validate(a.ctx, req)
	return v, service.Wire(a.log, err)
}

func (a *App) InstallLibrary(req service.ImportRequest) (*service.InstallResult, error) {
	v, err := a.svc.Libraries.Install(a.ctx, req)
	return v, service.Wire(a.log, err)
}

// ResponsibilityNotice is the fixed text shown when installing manually imported data.
func (a *App) ResponsibilityNotice() string { return datasets.ResponsibilityNotice }

// ---- Sêmen

func (a *App) SemenBank(q service.BankQuery) (*service.BankPage, error) {
	v, err := a.svc.Semen.Bank(a.ctx, q)
	return v, service.Wire(a.log, err)
}

func (a *App) GetBullSheet(bullID string) (*service.BullSheet, error) {
	v, err := a.svc.Semen.Sheet(a.ctx, bullID, 200)
	return v, service.Wire(a.log, err)
}

func (a *App) ListSemenTypes() ([]domain.SemenType, error) {
	v, err := a.svc.Semen.Types(a.ctx)
	return v, service.Wire(a.log, err)
}

func (a *App) CreateSemenType(name string) (*domain.SemenType, error) {
	v, err := a.svc.Semen.CreateType(a.ctx, name)
	return v, service.Wire(a.log, err)
}

func (a *App) AddSemenLot(in service.SemenInput) (*domain.SemenInventory, error) {
	v, err := a.svc.Semen.AddLot(a.ctx, in)
	return v, service.Wire(a.log, err)
}

func (a *App) UpdateSemenLot(id string, in service.LotInfoInput) (*domain.SemenInventory, error) {
	v, err := a.svc.Semen.UpdateLot(a.ctx, id, in)
	return v, service.Wire(a.log, err)
}

func (a *App) ReceiveSemen(in service.MovementInput) (*domain.SemenMovement, error) {
	v, err := a.svc.Semen.Receive(a.ctx, in)
	return v, service.Wire(a.log, err)
}

func (a *App) WithdrawSemen(in service.MovementInput) (*domain.SemenMovement, error) {
	v, err := a.svc.Semen.Withdraw(a.ctx, in)
	return v, service.Wire(a.log, err)
}

func (a *App) AdjustSemen(in service.MovementInput) (*domain.SemenMovement, error) {
	v, err := a.svc.Semen.Adjust(a.ctx, in)
	return v, service.Wire(a.log, err)
}

// ---- Acasalamentos (motor)

func (a *App) RunMating(req service.RunRequest) (*service.RunResponse, error) {
	v, err := a.svc.Matings.Run(a.ctx, req)
	return v, service.Wire(a.log, err)
}

func (a *App) GetCalculation(matingID string) (*service.CalculationView, error) {
	v, err := a.svc.Matings.GetCalculation(a.ctx, matingID)
	return v, service.Wire(a.log, err)
}

func (a *App) ListCalculations(femaleID string) ([]service.CalculationSummary, error) {
	v, err := a.svc.Matings.ListCalculations(a.ctx, femaleID)
	return v, service.Wire(a.log, err)
}

func (a *App) SetMatingStatus(matingID, status string) error {
	return service.Wire(a.log, a.svc.Matings.SetStatus(a.ctx, matingID, status))
}

func (a *App) ListMatingProfiles() ([]engine.Profile, error) {
	v, err := a.svc.Matings.Profiles(a.ctx)
	return v, service.Wire(a.log, err)
}

func (a *App) SaveMatingProfile(p engine.Profile) (*engine.Profile, error) {
	v, err := a.svc.Matings.SaveProfile(a.ctx, p)
	return v, service.Wire(a.log, err)
}

func (a *App) DeleteMatingProfile(id string) error {
	return service.Wire(a.log, a.svc.Matings.DeleteProfile(a.ctx, id))
}

// EngineVersion reports the version of the mating engine.
func (a *App) EngineVersion() string { return engine.Version }

// ---- Backup, restauração e integridade

func (a *App) SuggestedBackupName() string { return a.bk.SuggestedName(a.ctx) }

// PickBackupSavePath opens the native "save as" dialog ("" if cancelled).
func (a *App) PickBackupSavePath(defaultName string) (string, error) {
	p, err := wruntime.SaveFileDialog(a.ctx, wruntime.SaveDialogOptions{
		Title: "Criar backup", DefaultFilename: defaultName,
		Filters: []wruntime.FileFilter{{DisplayName: "Backup (*.genbackup)", Pattern: "*.genbackup"}},
	})
	return p, service.Wire(a.log, err)
}

// PickBackupOpenPath opens the native file dialog to choose a backup ("" if cancelled).
func (a *App) PickBackupOpenPath() (string, error) {
	p, err := wruntime.OpenFileDialog(a.ctx, wruntime.OpenDialogOptions{
		Title:   "Restaurar backup",
		Filters: []wruntime.FileFilter{{DisplayName: "Backup (*.genbackup)", Pattern: "*.genbackup"}},
	})
	return p, service.Wire(a.log, err)
}

func (a *App) CreateBackup(path string) (*backup.Manifest, error) {
	v, err := a.bk.Create(a.ctx, path)
	return v, service.Wire(a.log, err)
}

func (a *App) InspectBackup(path string) (*backup.Preview, error) {
	v, err := a.bk.Inspect(a.ctx, path)
	return v, service.Wire(a.log, err)
}

func (a *App) RestoreBackup(path string) (*backup.RestoreResult, error) {
	v, err := a.bk.Restore(a.ctx, path)
	return v, service.Wire(a.log, err)
}

func (a *App) ListAutomaticBackups() []string { return a.bk.ListAutomatic() }

func (a *App) RunIntegrityCheck() (*backup.Report, error) {
	v, err := a.bk.CheckIntegrity(a.ctx)
	return v, service.Wire(a.log, err)
}
