package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"genetica/internal/apperr"
	"genetica/internal/database"
	"genetica/internal/datasets"
	"genetica/internal/domain"
	"genetica/internal/repository"
)

// LibraryService installs and manages dataset packages ("bibliotecas").
type LibraryService struct {
	db  *database.DB
	dir string // directory holding the stored .genlib copies
}

const (
	UnmatchedSkip   = "skip"
	UnmatchedCreate = "create"
)

// LibrarySummary is one row of the installed-libraries list.
type LibrarySummary struct {
	ID                    string   `json:"id"`
	PackageID             string   `json:"packageId"`
	Name                  string   `json:"name"`
	Version               string   `json:"version"`
	Origin                string   `json:"origin"`
	Species               string   `json:"species"`
	Breeds                []string `json:"breeds"`
	Purposes              []string `json:"purposes"`
	Organization          string   `json:"organization"`
	SourceURL             string   `json:"sourceUrl"`
	Program               string   `json:"program"`
	Release               string   `json:"release"`
	LicenseName           string   `json:"licenseName"`
	LicenseURL            string   `json:"licenseUrl"`
	LicenseStatus         string   `json:"licenseStatus"`
	RedistributionAllowed bool     `json:"redistributionAllowed"`
	Active                bool     `json:"active"`
	IsDemo                bool     `json:"isDemo"`
	RecordCount           int      `json:"recordCount"`
	TraitCount            int      `json:"traitCount"`
	FileSHA256            string   `json:"fileSha256"`
	InstalledAt           string   `json:"installedAt"`
}

// LibraryDetail adds documentation, license text and traits.
type LibraryDetail struct {
	LibrarySummary
	Manifest             json.RawMessage `json:"manifest"`
	Readme               string          `json:"readme"`
	LicenseText          string          `json:"licenseText"`
	Traits               []domain.Trait  `json:"traits"`
	UsedByCalculations   int             `json:"usedByCalculations"`
	ResponsibilityNotice string          `json:"responsibilityNotice"`
}

const summaryCols = `id, package_id, name, version, origin, species, breeds, purposes, organization, source_url, program, release,
	license_name, license_url, license_status, redistribution_allowed, active, is_demo, record_count, trait_count, file_sha256, installed_at`

func scanSummary(sc interface{ Scan(...any) error }) (*LibrarySummary, error) {
	var l LibrarySummary
	var breeds, purposes string
	var redis, active, demo int
	if err := sc.Scan(&l.ID, &l.PackageID, &l.Name, &l.Version, &l.Origin, &l.Species, &breeds, &purposes, &l.Organization, &l.SourceURL,
		&l.Program, &l.Release, &l.LicenseName, &l.LicenseURL, &l.LicenseStatus, &redis, &active, &demo, &l.RecordCount, &l.TraitCount,
		&l.FileSHA256, &l.InstalledAt); err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(breeds), &l.Breeds)
	_ = json.Unmarshal([]byte(purposes), &l.Purposes)
	if l.Breeds == nil {
		l.Breeds = []string{}
	}
	if l.Purposes == nil {
		l.Purposes = []string{}
	}
	l.RedistributionAllowed, l.Active, l.IsDemo = redis == 1, active == 1, demo == 1
	return &l, nil
}

func (s *LibraryService) List(ctx context.Context) ([]LibrarySummary, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+summaryCols+` FROM datasets ORDER BY name COLLATE NOCASE, installed_at DESC`)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDatabase, "Falha ao listar as bibliotecas.", err)
	}
	defer rows.Close()
	out := []LibrarySummary{}
	for rows.Next() {
		l, err := scanSummary(rows)
		if err != nil {
			return nil, apperr.Wrap(apperr.CodeDatabase, "Falha ao ler as bibliotecas.", err)
		}
		out = append(out, *l)
	}
	return out, rows.Err()
}

func (s *LibraryService) Get(ctx context.Context, id string) (*LibraryDetail, error) {
	l, err := scanSummary(s.db.QueryRowContext(ctx, `SELECT `+summaryCols+` FROM datasets WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperr.New(apperr.CodeNotFound, "Biblioteca não encontrada.")
	}
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDatabase, "Falha ao ler a biblioteca.", err)
	}
	d := &LibraryDetail{LibrarySummary: *l, ResponsibilityNotice: datasets.ResponsibilityNotice}
	var manifest string
	if err := s.db.QueryRowContext(ctx, `SELECT manifest, readme, license_text FROM datasets WHERE id=?`, id).Scan(&manifest, &d.Readme, &d.LicenseText); err != nil {
		return nil, apperr.Wrap(apperr.CodeDatabase, "Falha ao ler a biblioteca.", err)
	}
	d.Manifest = json.RawMessage(manifest)
	if d.Traits, err = repos(s.db).Traits.ListByDataset(ctx, id); err != nil {
		return nil, err
	}
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT mating_id) FROM calculation_datasets WHERE dataset_id=?`, id).Scan(&d.UsedByCalculations)
	return d, nil
}

func (s *LibraryService) SetActive(ctx context.Context, id string, active bool) error {
	res, err := s.db.ExecContext(ctx, `UPDATE datasets SET active=? WHERE id=?`, boolToInt(active), id)
	if err != nil {
		return apperr.Wrap(apperr.CodeDatabase, "Falha ao atualizar a biblioteca.", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return apperr.New(apperr.CodeNotFound, "Biblioteca não encontrada.")
	}
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Remove deletes a library version, its values and its traits. A version used
// by a saved calculation is never removed: it must stay reproducible.
func (s *LibraryService) Remove(ctx context.Context, id string) error {
	var fileName string
	err := s.db.WithTx(ctx, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx, `SELECT file_name FROM datasets WHERE id=?`, id).Scan(&fileName); errors.Is(err, sql.ErrNoRows) {
			return apperr.New(apperr.CodeNotFound, "Biblioteca não encontrada.")
		} else if err != nil {
			return apperr.Wrap(apperr.CodeDatabase, "Falha ao ler a biblioteca.", err)
		}
		var used int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM calculation_datasets WHERE dataset_id=?`, id).Scan(&used); err != nil {
			return apperr.Wrap(apperr.CodeDatabase, "Falha ao verificar o uso da biblioteca.", err)
		}
		if used > 0 {
			return apperr.New(apperr.CodeConflict, fmt.Sprintf(
				"Esta versão foi usada em %d cálculo(s) salvo(s) e não pode ser removida, para que eles continuem reproduzíveis. Desative-a em vez disso.", used))
		}
		for _, q := range []string{
			`DELETE FROM genetic_evaluations WHERE dataset_id=?`,
			`DELETE FROM traits WHERE source_dataset=?`,
			`DELETE FROM datasets WHERE id=?`,
		} {
			if _, err := tx.ExecContext(ctx, q, id); err != nil {
				return apperr.Wrap(apperr.CodeDatabase, "Falha ao remover a biblioteca.", err)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if fileName != "" {
		_ = os.Remove(filepath.Join(s.dir, fileName))
	}
	return nil
}

// ---- import: inspect -> validate -> install

// ImportInspection is what the wizard learns from a chosen file.
type ImportInspection struct {
	Kind                  string              `json:"kind"` // genlib | csv | json
	Path                  string              `json:"path"`
	Columns               []string            `json:"columns"`
	SampleRows            [][]string          `json:"sampleRows"`
	RowCount              int                 `json:"rowCount"`
	SuggestedDecimalComma bool                `json:"suggestedDecimalComma"`
	SuggestedIDColumn     string              `json:"suggestedIdColumn"`
	SuggestedNameColumn   string              `json:"suggestedNameColumn"`
	Manifest              *datasets.Manifest  `json:"manifest"` // genlib only
	Traits                []datasets.TraitDef `json:"traits"`   // genlib only
	Readme                string              `json:"readme"`
	LicenseText           string              `json:"licenseText"`
	Validation            *ImportValidation   `json:"validation"` // genlib only
}

// ImportRequest is the wizard's final state. For .genlib files the manifest and
// traits come from the package and these fields are ignored.
type ImportRequest struct {
	Path                 string              `json:"path"`
	Manifest             datasets.Manifest   `json:"manifest"`
	Traits               []datasets.TraitDef `json:"traits"`
	LicenseText          string              `json:"licenseText"`
	UnmatchedPolicy      string              `json:"unmatchedPolicy"`
	FarmID               string              `json:"farmId"`
	AcceptResponsibility bool                `json:"acceptResponsibility"`
}

// ImportValidation is the report shown before confirming.
type ImportValidation struct {
	Report        datasets.Report   `json:"report"`
	LicenseStatus string            `json:"licenseStatus"`
	Duplicate     bool              `json:"duplicate"` // same id + version already installed
	CanInstall    bool              `json:"canInstall"`
	Notice        string            `json:"notice"`
	Manifest      datasets.Manifest `json:"manifest"`
}

type InstallResult struct {
	Library             LibrarySummary  `json:"library"`
	Report              datasets.Report `json:"report"`
	EvaluationsInserted int             `json:"evaluationsInserted"`
	AnimalsCreated      int             `json:"animalsCreated"`
	RecordsSkipped      int             `json:"recordsSkipped"`
}

func kindOf(path string) (string, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".genlib":
		return "genlib", nil
	case ".csv", ".tsv", ".txt":
		return "csv", nil
	case ".json":
		return "json", nil
	}
	return "", apperr.New(apperr.CodeValidation, "Formato não suportado. Use .genlib, .csv ou .json.")
}

func pkgErr(err error) error {
	var pe *datasets.PackageError
	if errors.As(err, &pe) {
		return apperr.Wrap(apperr.CodeValidation, pe.Error(), err)
	}
	return err
}

// Inspect reads the file the user chose and reports what the wizard needs.
func (s *LibraryService) Inspect(ctx context.Context, path string) (*ImportInspection, error) {
	kind, err := kindOf(path)
	if err != nil {
		return nil, err
	}
	in := &ImportInspection{Kind: kind, Path: path, Columns: []string{}, SampleRows: [][]string{}, Traits: []datasets.TraitDef{}}
	if kind == "genlib" {
		p, err := datasets.ReadPackageFile(path, datasets.DefaultLimits)
		if err != nil {
			return nil, pkgErr(err)
		}
		in.Manifest, in.Traits, in.Readme, in.LicenseText = &p.Manifest, p.Traits, p.Readme, p.License
		fillTable(in, p.Table)
		v, err := s.validatePackage(ctx, p, ImportRequest{Path: path})
		if err != nil {
			return nil, err
		}
		in.Validation = v
		return in, nil
	}
	st, err := os.Stat(path)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeNotFound, "Não foi possível abrir o arquivo.", err)
	}
	if st.Size() > datasets.DefaultLimits.MaxDataBytes {
		return nil, apperr.New(apperr.CodeValidation, "Arquivo grande demais.")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "Não foi possível ler o arquivo.", err)
	}
	t, err := datasets.ParseTable(b, kind)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeValidation, "Arquivo inválido: "+err.Error(), err)
	}
	fillTable(in, t)
	in.SuggestedDecimalComma = kind == "csv" && t.Delimiter == ';'
	in.SuggestedIDColumn = guessColumn(t.Columns, "registro", "registration", "reg", "rgd", "rgn", "id", "codigo", "código", "brinco")
	in.SuggestedNameColumn = guessColumn(t.Columns, "nome", "name", "animal", "touro")
	return in, nil
}

func fillTable(in *ImportInspection, t *datasets.Table) {
	in.Columns, in.RowCount = t.Columns, len(t.Rows)
	for i := 0; i < len(t.Rows) && i < 5; i++ {
		in.SampleRows = append(in.SampleRows, t.Rows[i])
	}
}

func guessColumn(cols []string, names ...string) string {
	for _, n := range names {
		for _, c := range cols {
			if strings.EqualFold(c, n) {
				return c
			}
		}
	}
	return ""
}

// load builds the package described by the request.
func (s *LibraryService) load(req ImportRequest) (*datasets.Package, string, error) {
	kind, err := kindOf(req.Path)
	if err != nil {
		return nil, "", err
	}
	if kind == "genlib" {
		p, err := datasets.ReadPackageFile(req.Path, datasets.DefaultLimits)
		return p, kind, pkgErr(err)
	}
	st, err := os.Stat(req.Path)
	if err != nil {
		return nil, kind, apperr.Wrap(apperr.CodeNotFound, "Não foi possível abrir o arquivo.", err)
	}
	if st.Size() > datasets.DefaultLimits.MaxDataBytes {
		return nil, kind, apperr.New(apperr.CodeValidation, "Arquivo grande demais.")
	}
	b, err := os.ReadFile(req.Path)
	if err != nil {
		return nil, kind, apperr.Wrap(apperr.CodeInternal, "Não foi possível ler o arquivo.", err)
	}
	t, err := datasets.ParseTable(b, kind)
	if err != nil {
		return nil, kind, apperr.Wrap(apperr.CodeValidation, "Arquivo inválido: "+err.Error(), err)
	}
	m := req.Manifest
	m.SchemaVersion = datasets.SchemaVersion
	m.Data.Format = kind
	m.Normalize()
	license := req.LicenseText
	if strings.TrimSpace(license) == "" {
		license = datasets.NoLicenseText
	}
	return &datasets.Package{Manifest: m, Traits: req.Traits, Table: t, License: license, Readme: datasets.DefaultReadme(m)}, kind, nil
}

// Validate runs every check without writing anything.
func (s *LibraryService) Validate(ctx context.Context, req ImportRequest) (*ImportValidation, error) {
	p, _, err := s.load(req)
	if err != nil {
		return nil, err
	}
	return s.validatePackage(ctx, p, req)
}

func (s *LibraryService) validatePackage(ctx context.Context, p *datasets.Package, req ImportRequest) (*ImportValidation, error) {
	v := &ImportValidation{Notice: datasets.ResponsibilityNotice, Manifest: p.Manifest, LicenseStatus: string(p.Manifest.LicenseStatus())}
	v.Report.Errors, v.Report.Warnings, v.Report.UnrecognizedColumns = []string{}, []string{}, []string{}

	if err := p.Manifest.Validate(); err != nil {
		v.Report.Errors = append(v.Report.Errors, flattenValidation(err)...)
	}
	if err := datasets.ValidateDefs(p.Traits); err != nil {
		v.Report.Errors = append(v.Report.Errors, flattenValidation(err)...)
	}
	// A column mapped to an existing catalogue trait must point to a genetic trait of the
	// catalogue itself (not a phenotype, and not a trait owned by another library).
	for i, d := range p.Traits {
		if d.ExistingTraitID == "" {
			continue
		}
		t, err := repos(s.db).Traits.Get(ctx, d.ExistingTraitID)
		switch {
		case err != nil && apperr.CodeOf(err) == apperr.CodeNotFound:
			v.Report.Errors = append(v.Report.Errors, fmt.Sprintf("traits[%d]: a característica do catálogo escolhida não existe.", i))
		case err != nil:
			return nil, err
		case t.SourceDataset != "":
			v.Report.Errors = append(v.Report.Errors, fmt.Sprintf("traits[%d]: %s pertence a outra biblioteca; escolha uma característica do catálogo ou crie uma nova.", i, t.Code))
		case t.EvaluationType == domain.EvalPhenotype:
			v.Report.Errors = append(v.Report.Errors, fmt.Sprintf("traits[%d]: %s é um fenótipo; bibliotecas fornecem avaliações genéticas (DEP, EPD, EBV, PTA ou índice).", i, t.Code))
		}
	}
	if len(v.Report.Errors) > 0 {
		return v, nil
	}

	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) > 0 FROM datasets WHERE package_id=? AND version=?`, p.Manifest.ID, p.Manifest.Version).Scan(&v.Duplicate); err != nil {
		return nil, apperr.Wrap(apperr.CodeDatabase, "Falha ao verificar duplicidade.", err)
	}
	if v.Duplicate {
		v.Report.Errors = append(v.Report.Errors, fmt.Sprintf("A biblioteca %s versão %s já está instalada.", p.Manifest.ID, p.Manifest.Version))
	}

	recs, rep := datasets.Extract(&p.Manifest, p.Traits, p.Table)
	rep.Errors = append(rep.Errors, v.Report.Errors...)
	if recs != nil {
		idx, err := s.registrationIndex(ctx)
		if err != nil {
			return nil, err
		}
		for _, r := range recs {
			switch len(idx[datasets.NormalizeKey(r.Registration)]) {
			case 0:
				rep.UnmatchedAnimals++
			case 1:
				rep.MatchedAnimals++
			default:
				rep.AmbiguousAnimals++
			}
		}
		if rep.AmbiguousAnimals > 0 {
			rep.Warnings = append(rep.Warnings, fmt.Sprintf("%d registros correspondem a mais de um animal cadastrado e serão ignorados.", rep.AmbiguousAnimals))
		}
		if rep.UnmatchedAnimals > 0 {
			rep.Warnings = append(rep.Warnings, fmt.Sprintf("%d registros não correspondem a nenhum animal cadastrado (veja a opção de cadastro).", rep.UnmatchedAnimals))
		}
	}
	if p.Manifest.LicenseStatus() == datasets.LicenseUnknown {
		rep.Warnings = append(rep.Warnings, "Licença desconhecida: os dados ficam restritos a este computador e não devem ser redistribuídos.")
	}
	v.Report = *rep
	v.CanInstall = rep.OK()
	return v, nil
}

func flattenValidation(err error) []string {
	var ve *domain.ValidationError
	if errors.As(err, &ve) {
		out := make([]string, 0, len(ve.Fields))
		for k, m := range ve.Fields {
			out = append(out, k+": "+m)
		}
		sortStrings(out)
		return out
	}
	return []string{err.Error()}
}

func sortStrings(a []string) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j] < a[j-1]; j-- {
			a[j], a[j-1] = a[j-1], a[j]
		}
	}
}

// registrationIndex maps normalized registration -> animal IDs.
func (s *LibraryService) registrationIndex(ctx context.Context) (map[string][]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, registration FROM animals WHERE registration IS NOT NULL AND registration <> ''`)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDatabase, "Falha ao consultar os animais.", err)
	}
	defer rows.Close()
	idx := map[string][]string{}
	for rows.Next() {
		var id, reg string
		if err := rows.Scan(&id, &reg); err != nil {
			return nil, apperr.Wrap(apperr.CodeDatabase, "Falha ao consultar os animais.", err)
		}
		k := datasets.NormalizeKey(reg)
		idx[k] = append(idx[k], id)
	}
	return idx, rows.Err()
}

// Install validates and then installs the library atomically.
func (s *LibraryService) Install(ctx context.Context, req ImportRequest) (*InstallResult, error) {
	if !req.AcceptResponsibility {
		return nil, fieldError("acceptResponsibility", datasets.ResponsibilityNotice+" Confirme para continuar.")
	}
	if req.UnmatchedPolicy != UnmatchedCreate && req.UnmatchedPolicy != UnmatchedSkip {
		req.UnmatchedPolicy = UnmatchedSkip
	}
	if req.UnmatchedPolicy == UnmatchedCreate && req.FarmID == "" {
		return nil, fieldError("farmId", "Escolha a propriedade onde os animais serão cadastrados.")
	}
	p, kind, err := s.load(req)
	if err != nil {
		return nil, err
	}
	val, err := s.validatePackage(ctx, p, req)
	if err != nil {
		return nil, err
	}
	if !val.CanInstall {
		return nil, apperr.New(apperr.CodeValidation, "A validação encontrou problemas: "+strings.Join(val.Report.Errors, " "))
	}
	recs, _ := datasets.Extract(&p.Manifest, p.Traits, p.Table)

	// Every installed dataset is stored as a .genlib. A package keeps its original bytes.
	raw := p.Raw
	if raw == nil {
		if raw, err = datasets.WritePackage(p); err != nil {
			return nil, apperr.Wrap(apperr.CodeInternal, "Falha ao empacotar a biblioteca.", err)
		}
		sum := sha256Hex(raw)
		p.SHA256 = sum
	}
	datasetID := newID()
	fileName := datasetID + ".genlib"
	tmp := filepath.Join(s.dir, fileName+".tmp")
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "Não foi possível criar a pasta de bibliotecas.", err)
	}
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "Não foi possível gravar a biblioteca.", err)
	}

	res := &InstallResult{Report: val.Report}
	err = s.db.WithTx(ctx, func(tx *sql.Tx) error {
		return s.installTx(ctx, tx, p, kind, recs, req, datasetID, fileName, res)
	})
	if err != nil {
		_ = os.Remove(tmp)
		return nil, err
	}
	if err := os.Rename(tmp, filepath.Join(s.dir, fileName)); err != nil {
		// The database already holds the data; keep the tmp copy so nothing is lost.
		return nil, apperr.Wrap(apperr.CodeInternal, "A biblioteca foi instalada, mas o arquivo original não pôde ser guardado.", err)
	}
	l, err := scanSummary(s.db.QueryRowContext(ctx, `SELECT `+summaryCols+` FROM datasets WHERE id=?`, datasetID))
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDatabase, "Falha ao ler a biblioteca instalada.", err)
	}
	res.Library = *l
	return res, nil
}

func (s *LibraryService) installTx(ctx context.Context, tx *sql.Tx, p *datasets.Package, kind string, recs []datasets.Record,
	req ImportRequest, datasetID, fileName string, res *InstallResult) error {

	r := repository.New(tx)
	now := nowTS()
	m := p.Manifest
	manifestJSON, _ := json.Marshal(m)
	breeds, _ := json.Marshal(m.Breeds)
	purposes, _ := json.Marshal(m.Purpose)

	if _, err := tx.ExecContext(ctx, `INSERT INTO datasets (id, package_id, name, version, is_demo, installed_at, origin, manifest, species, breeds, purposes,
		organization, source_url, program, release, license_name, license_url, license_status, redistribution_allowed, readme, license_text,
		record_count, trait_count, active, file_name, file_sha256, responsibility_accepted)
		VALUES (?,?,?,?,0,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,1,?,?,1)`,
		datasetID, m.ID, m.Name, m.Version, now, kind, string(manifestJSON), m.Species, string(breeds), string(purposes),
		m.Organization, m.SourceURL, m.Evaluation.Program, m.Evaluation.Release, m.License.Name, m.License.URL, string(m.LicenseStatus()),
		boolToInt(m.License.RedistributionAllowed), p.Readme, p.License, len(recs), len(p.Traits), fileName, p.SHA256); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return apperr.New(apperr.CodeConflict, fmt.Sprintf("A biblioteca %s versão %s já está instalada.", m.ID, m.Version))
		}
		return apperr.Wrap(apperr.CodeDatabase, "Falha ao registrar a biblioteca.", err)
	}

	// traits: dataset-owned definitions, or reused catalogue entries
	type resolved struct {
		code, kind, unit, dataType string
		t                          *domain.Trait
	}
	rt := make([]resolved, len(p.Traits))
	for i, d := range p.Traits {
		var t *domain.Trait
		if d.ExistingTraitID != "" {
			ex, err := r.Traits.Get(ctx, d.ExistingTraitID)
			if err != nil {
				return err
			}
			t = ex
		} else {
			t = d.ToTrait(datasetID, newID(), now)
			if err := t.Validate(); err != nil {
				return err
			}
			if err := r.Traits.Insert(ctx, t); err != nil {
				return err
			}
		}
		rt[i] = resolved{code: t.Code, kind: t.EvaluationType, unit: t.Unit, t: t}
	}

	idx := map[string][]string{}
	rows, err := tx.QueryContext(ctx, `SELECT id, registration FROM animals WHERE registration IS NOT NULL AND registration <> ''`)
	if err != nil {
		return apperr.Wrap(apperr.CodeDatabase, "Falha ao consultar os animais.", err)
	}
	for rows.Next() {
		var id, reg string
		if err := rows.Scan(&id, &reg); err != nil {
			rows.Close()
			return apperr.Wrap(apperr.CodeDatabase, "Falha ao consultar os animais.", err)
		}
		k := datasets.NormalizeKey(reg)
		idx[k] = append(idx[k], id)
	}
	rows.Close()

	breedByCode := map[string]string{}
	breedList, err := r.Breeds.List(ctx, false)
	if err != nil {
		return err
	}
	for _, b := range breedList {
		breedByCode[strings.ToUpper(b.Code)] = b.ID
	}
	defaultBreed := ""
	if len(m.Breeds) > 0 {
		defaultBreed = m.Breeds[0]
	}

	for _, rec := range recs {
		ids := idx[datasets.NormalizeKey(rec.Registration)]
		var animalID string
		switch {
		case len(ids) == 1:
			animalID = ids[0]
		case len(ids) > 1:
			res.RecordsSkipped++
			continue
		case req.UnmatchedPolicy == UnmatchedCreate:
			a := &domain.Animal{ID: newID(), FarmID: req.FarmID, Registration: rec.Registration, Name: rec.Name, Sex: rec.Sex,
				Status: domain.StatusActive, BreedID: breedByCode[strings.ToUpper(firstNonEmpty(rec.Breed, defaultBreed))],
				Notes: fmt.Sprintf("Cadastrado pela importação da biblioteca %s %s.", m.Name, m.Version), CreatedAt: now, UpdatedAt: now}
			if a.Name == "" {
				a.Name = rec.Registration
			}
			if a.Sex == domain.SexFemale {
				a.Category = domain.CategoryCow
			} else {
				a.Sex, a.Category = domain.SexMale, domain.CategoryBull
			}
			if err := validateAnimal(ctx, r, a); err != nil {
				return err
			}
			if err := r.Animals.Insert(ctx, a); err != nil {
				return err
			}
			animalID = a.ID
			idx[datasets.NormalizeKey(rec.Registration)] = []string{a.ID}
			res.AnimalsCreated++
		default:
			res.RecordsSkipped++
			continue
		}
		for _, v := range rec.Values {
			t := rt[v.TraitIndex]
			if err := t.t.ValidateValue(v.Value); err != nil {
				continue // e.g. a decimal in an integer trait: ignored, not coerced
			}
			e := &domain.GeneticEvaluation{ID: newID(), AnimalID: animalID, DatasetID: datasetID,
				EvaluationProgram: m.Evaluation.Program, EvaluationRelease: m.Evaluation.Release,
				BreedCode: firstNonEmpty(rec.Breed, defaultBreed), TraitCode: t.code, EvaluationType: t.kind, Value: v.Value, Unit: t.unit,
				Accuracy: v.Accuracy, Percentile: v.Percentile, GeneticBase: m.Evaluation.GeneticBase, SourceReference: m.SourceURL,
				EvaluatedAt: m.ReleaseDate, CreatedAt: now}
			if err := e.Validate(); err != nil {
				return err
			}
			if err := r.Evaluations.Insert(ctx, e); err != nil {
				return err
			}
			res.EvaluationsInserted++
		}
	}
	return nil
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
