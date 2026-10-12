package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"strings"

	"genetica/internal/apperr"
	"genetica/internal/domain"
	"genetica/internal/engine"
	"genetica/internal/genetics"
	"genetica/internal/repository"
)

// RunRequest asks for a ranking of candidate bulls for one female.
type RunRequest struct {
	FemaleID string         `json:"femaleId"`
	BullIDs  []string       `json:"bullIds"`
	Profile  engine.Profile `json:"profile"`
	Save     bool           `json:"save"`
}

// RunResponse carries the engine output plus the IDs under which pairs were saved.
type RunResponse struct {
	RunID     string            `json:"runId"`
	Saved     bool              `json:"saved"`
	MatingIDs map[string]string `json:"matingIds"` // bull ID -> mating ID
	Output    engine.RunOutput  `json:"output"`
	// Snapshots holds the self-contained snapshot of every pair that reached the
	// pipeline, so "Ver cálculo" works even for unsaved previews.
	Snapshots map[string]engine.PairSnapshot `json:"snapshots"`
}

// CalculationView is a saved pair calculation.
type CalculationView struct {
	ID            string          `json:"id"`
	RunID         string          `json:"runId"`
	FemaleID      string          `json:"femaleId"`
	BullID        string          `json:"bullId"`
	Status        string          `json:"status"`
	Score         *float64        `json:"score"`
	EngineVersion string          `json:"engineVersion"`
	CreatedAt     string          `json:"createdAt"`
	IsDemo        bool            `json:"isDemo"`
	Snapshot      json.RawMessage `json:"snapshot"`
}

// CalculationSummary is one row of a female's calculation history.
type CalculationSummary struct {
	ID            string   `json:"id"`
	RunID         string   `json:"runId"`
	BullID        string   `json:"bullId"`
	BullName      string   `json:"bullName"`
	ProfileName   string   `json:"profileName"`
	Status        string   `json:"status"`
	Score         *float64 `json:"score"`
	Rank          int      `json:"rank"`
	EngineVersion string   `json:"engineVersion"`
	CreatedAt     string   `json:"createdAt"`
	IsDemo        bool     `json:"isDemo"`
}

func parseTraitMeta(raw string) engine.TraitMeta {
	var m struct {
		PercentileSemantics string `json:"percentileSemantics"`
		Population          *struct {
			Mean *float64 `json:"mean"`
			SD   *float64 `json:"sd"`
		} `json:"population"`
		PopulationMean *float64 `json:"populationMean"`
		PopulationSD   *float64 `json:"populationSd"`
	}
	_ = json.Unmarshal([]byte(raw), &m)
	out := engine.TraitMeta{PercentileSemantics: m.PercentileSemantics, PopulationMean: m.PopulationMean, PopulationSD: m.PopulationSD}
	if m.Population != nil {
		out.PopulationMean, out.PopulationSD = m.Population.Mean, m.Population.SD
	}
	return out
}

type engineLoader struct {
	ctx context.Context
	r   *repository.Repos
	db  interface {
		QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row
	}
	breeds   map[string]string
	datasets map[string]engine.DatasetInfo
}

func (l *engineLoader) dataset(id string) (engine.DatasetInfo, bool, error) {
	if id == "" {
		return engine.DatasetInfo{}, false, nil
	}
	if d, ok := l.datasets[id]; ok {
		return d, true, nil
	}
	var d engine.DatasetInfo
	var demo int
	err := l.db.QueryRowContext(l.ctx, `SELECT id, package_id, name, version, program, release, file_sha256, is_demo, origin, license_status FROM datasets WHERE id=?`, id).
		Scan(&d.ID, &d.PackageID, &d.Name, &d.Version, &d.Program, &d.Release, &d.SHA256, &demo, &d.Origin, &d.LicenseStatus)
	if err != nil {
		return engine.DatasetInfo{}, false, apperr.Wrap(apperr.CodeDatabase, "Falha ao ler a biblioteca usada na avaliação.", err)
	}
	d.IsDemo = demo == 1
	l.datasets[id] = d
	return d, true, nil
}

func (l *engineLoader) animal(a domain.Animal) engine.Animal {
	return engine.Animal{ID: a.ID, Name: a.Name, Registration: a.Registration, Sex: a.Sex, Status: a.Status, BreedName: l.breeds[a.BreedID], IsDemo: a.IsDemo}
}

func (l *engineLoader) evals(animalID string) ([]engine.Evaluation, error) {
	rows, err := l.r.Evaluations.ListByAnimal(l.ctx, animalID)
	if err != nil {
		return nil, err
	}
	out := make([]engine.Evaluation, 0, len(rows))
	for _, e := range rows {
		ev := engine.Evaluation{
			ID: e.ID, Value: e.Value, Accuracy: e.Accuracy, Percentile: e.Percentile, DatasetID: e.DatasetID, IsDemo: e.IsDemo, CreatedAt: e.CreatedAt,
			TraitName: e.TraitCode,
			Basis: engine.Basis{TraitCode: e.TraitCode, EvaluationType: e.EvaluationType, Program: e.EvaluationProgram, Release: e.EvaluationRelease,
				GeneticBase: e.GeneticBase, BreedCode: e.BreedCode, Unit: e.Unit},
		}
		if d, ok, err := l.dataset(e.DatasetID); err != nil {
			return nil, err
		} else if ok {
			ev.DatasetName, ev.DatasetVersion = d.Name, d.Version
		}
		if t, err := l.r.Traits.Resolve(l.ctx, e.TraitCode, e.EvaluationType, e.DatasetID); err == nil {
			ev.TraitName, ev.Direction, ev.IdealMin, ev.IdealMax = t.Name, t.OptimizationDirection, t.IdealMin, t.IdealMax
			ev.Meta = parseTraitMeta(t.Metadata)
		} else if apperr.CodeOf(err) != apperr.CodeNotFound {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, nil
}

// Run ranks candidate bulls for a female with the mating engine, optionally
// saving one snapshot per pair.
func (s *MatingService) Run(ctx context.Context, req RunRequest) (*RunResponse, error) {
	r := repos(s.db)
	female, err := requireAnimal(ctx, r, req.FemaleID, "femaleId")
	if err != nil {
		return nil, err
	}
	if len(req.BullIDs) == 0 {
		return nil, fieldError("bullIds", "Escolha ao menos um touro candidato.")
	}
	breeds, err := r.Breeds.List(ctx, false)
	if err != nil {
		return nil, err
	}
	l := &engineLoader{ctx: ctx, r: r, db: s.db, breeds: map[string]string{}, datasets: map[string]engine.DatasetInfo{}}
	for _, b := range breeds {
		l.breeds[b.ID] = b.Name
	}

	profile := req.Profile
	gens := profile.Filters.AncestorGenerations
	if gens <= 0 {
		gens = 3
	}
	gens = clampGenerations(gens)

	in := engine.RunInput{Profile: profile, Now: domain.Now(), Pedigree: genetics.Graph{}}
	in.Female.Animal = l.animal(*female)
	if in.Female.Evals, err = l.evals(female.ID); err != nil {
		return nil, err
	}
	merge := func(id string) error {
		g, err := loadGraph(ctx, r, id, gens)
		if err != nil {
			return err
		}
		for k, v := range g {
			in.Pedigree[k] = v
		}
		return nil
	}
	if err := merge(female.ID); err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	for _, id := range req.BullIDs {
		if seen[id] {
			continue // a duplicated id in the request is the same bull
		}
		seen[id] = true
		a, err := r.Animals.Get(ctx, id)
		if err != nil {
			if apperr.CodeOf(err) == apperr.CodeNotFound {
				return nil, fieldError("bullIds", "Um dos touros escolhidos não foi encontrado.")
			}
			return nil, err
		}
		c := engine.Candidate{Animal: l.animal(*a)}
		if c.Evals, err = l.evals(a.ID); err != nil {
			return nil, err
		}
		lots, err := r.Semen.ListByBull(ctx, a.ID)
		if err != nil {
			return nil, err
		}
		for _, lot := range lots {
			c.Stock.Doses += lot.Quantity
		}
		c.Stock.UnitCost = unitCost(lots)
		if err := merge(a.ID); err != nil {
			return nil, err
		}
		in.Candidates = append(in.Candidates, c)
	}
	// every dataset referenced by any evaluation
	for _, ev := range append(append([]engine.Evaluation{}, in.Female.Evals...), flattenEvals(in.Candidates)...) {
		if d, ok, err := l.dataset(ev.DatasetID); err != nil {
			return nil, err
		} else if ok {
			in.Datasets = appendDataset(in.Datasets, d)
		}
	}

	out := engine.New().Run(in)
	resp := &RunResponse{Output: out, MatingIDs: map[string]string{}, Snapshots: map[string]engine.PairSnapshot{}}
	if len(out.Errors) > 0 {
		return resp, nil
	}
	resp.Snapshots = engine.Snapshots(in, out)
	if !req.Save {
		return resp, nil
	}

	runID := newID()
	resp.RunID = runID
	err = s.db.WithTx(ctx, func(tx *sql.Tx) error {
		return s.saveRun(ctx, tx, female, runID, in, out, resp)
	})
	if err != nil {
		return nil, err
	}
	resp.Saved = true
	return resp, nil
}

func flattenEvals(cs []engine.Candidate) []engine.Evaluation {
	var out []engine.Evaluation
	for _, c := range cs {
		out = append(out, c.Evals...)
	}
	return out
}

func appendDataset(list []engine.DatasetInfo, d engine.DatasetInfo) []engine.DatasetInfo {
	for _, x := range list {
		if x.ID == d.ID {
			return list
		}
	}
	return append(list, d)
}

func (s *MatingService) saveRun(ctx context.Context, tx *sql.Tx, female *domain.Animal, runID string, in engine.RunInput, out engine.RunOutput, resp *RunResponse) error {
	type excluded struct {
		BullID     string             `json:"bullId"`
		Name       string             `json:"name"`
		Exclusions []engine.Exclusion `json:"exclusions"`
	}
	summary := struct {
		Bases    []engine.ReferenceBasis `json:"bases"`
		Warnings []string                `json:"warnings"`
		Top      []string                `json:"top"`
		Excluded []excluded              `json:"excluded"`
		Profile  engine.Profile          `json:"profile"`
	}{Bases: out.Bases, Warnings: out.Warnings, Top: out.Top, Profile: out.Profile, Excluded: []excluded{}}
	for _, r := range out.Results {
		if r.Status == engine.StatusExcluded {
			summary.Excluded = append(summary.Excluded, excluded{r.Bull.ID, r.Bull.Label(), r.Exclusions})
		}
	}
	sj, _ := json.Marshal(summary)
	demo := out.Profile.IsDemo || female.IsDemo
	for _, snap := range resp.Snapshots {
		demo = demo || snap.IsDemo
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO calculation_runs (id, farm_id, female_id, profile_name, engine_version, candidate_count, summary, is_demo, created_at)
		VALUES (?,?,?,?,?,?,?,?,?)`, runID, female.FarmID, female.ID, out.Profile.Name, out.EngineVersion, len(in.Candidates), string(sj), boolToInt(demo), out.CreatedAt); err != nil {
		return apperr.Wrap(apperr.CodeDatabase, "Falha ao registrar o cálculo.", err)
	}

	r := repository.New(tx)
	bullIDs := make([]string, 0, len(resp.Snapshots))
	for id := range resp.Snapshots {
		bullIDs = append(bullIDs, id)
	}
	sort.Strings(bullIDs)
	for _, bid := range bullIDs {
		snap := resp.Snapshots[bid]
		b, err := json.Marshal(snap)
		if err != nil {
			return apperr.Wrap(apperr.CodeInternal, "Falha ao montar o snapshot do cálculo.", err)
		}
		m := &domain.Mating{ID: newID(), FarmID: female.FarmID, FemaleID: female.ID, BullID: bid, Status: domain.MatingDraft,
			EngineVersion: out.EngineVersion, CalculationSnapshot: string(b), CreatedAt: out.CreatedAt}
		if snap.Result.Final != nil {
			v := snap.Result.Final.Final
			m.Score = &v
		}
		if err := m.Validate(); err != nil {
			return err
		}
		if err := r.Matings.Insert(ctx, m); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE matings SET run_id=?, is_demo=? WHERE id=?`, runID, boolToInt(snap.IsDemo), m.ID); err != nil {
			return apperr.Wrap(apperr.CodeDatabase, "Falha ao vincular o cálculo.", err)
		}
		for _, d := range snap.Datasets {
			if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO calculation_datasets (mating_id, dataset_id) VALUES (?, ?)`, m.ID, d.ID); err != nil {
				return apperr.Wrap(apperr.CodeDatabase, "Falha ao registrar as bibliotecas usadas.", err)
			}
		}
		resp.MatingIDs[bid] = m.ID
	}
	return nil
}

// GetCalculation returns a saved pair calculation exactly as it was computed.
func (s *MatingService) GetCalculation(ctx context.Context, matingID string) (*CalculationView, error) {
	var v CalculationView
	var run sql.NullString
	var score sql.NullFloat64
	var snap string
	var demo int
	err := s.db.QueryRowContext(ctx, `SELECT id, run_id, female_id, bull_id, status, score, engine_version, created_at, is_demo, calculation_snapshot FROM matings WHERE id=?`, matingID).
		Scan(&v.ID, &run, &v.FemaleID, &v.BullID, &v.Status, &score, &v.EngineVersion, &v.CreatedAt, &demo, &snap)
	if err == sql.ErrNoRows {
		return nil, apperr.New(apperr.CodeNotFound, "Cálculo não encontrado.")
	}
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDatabase, "Falha ao ler o cálculo.", err)
	}
	v.RunID, v.IsDemo, v.Snapshot = run.String, demo == 1, json.RawMessage(snap)
	if score.Valid {
		x := score.Float64
		v.Score = &x
	}
	return &v, nil
}

// ListCalculations returns a female's calculation history, newest first.
func (s *MatingService) ListCalculations(ctx context.Context, femaleID string) ([]CalculationSummary, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, COALESCE(run_id,''), bull_id, status, score, engine_version, created_at, is_demo,
		COALESCE(json_extract(calculation_snapshot, '$.bull.name'), ''), COALESCE(json_extract(calculation_snapshot, '$.profile.name'), ''),
		COALESCE(json_extract(calculation_snapshot, '$.result.rank'), 0), COALESCE(json_extract(calculation_snapshot, '$.result.status'), '')
		FROM matings WHERE female_id=? ORDER BY created_at DESC, rowid DESC`, femaleID)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDatabase, "Falha ao listar os cálculos.", err)
	}
	defer rows.Close()
	out := []CalculationSummary{}
	for rows.Next() {
		var c CalculationSummary
		var score sql.NullFloat64
		var demo int
		var calcStatus string
		if err := rows.Scan(&c.ID, &c.RunID, &c.BullID, &c.Status, &score, &c.EngineVersion, &c.CreatedAt, &demo, &c.BullName, &c.ProfileName, &c.Rank, &calcStatus); err != nil {
			return nil, apperr.Wrap(apperr.CodeDatabase, "Falha ao ler os cálculos.", err)
		}
		if score.Valid {
			x := score.Float64
			c.Score = &x
		}
		c.IsDemo = demo == 1
		out = append(out, c)
	}
	return out, rows.Err()
}

// SetStatus moves a saved mating through draft → planned → confirmed → done / cancelled.
func (s *MatingService) SetStatus(ctx context.Context, matingID, status string) error {
	switch status {
	case domain.MatingDraft, domain.MatingPlanned, domain.MatingConfirmed, domain.MatingDone, domain.MatingCancelled:
	default:
		return fieldError("status", "Status inválido.")
	}
	return repos(s.db).Matings.SetStatus(ctx, matingID, status)
}

// ---- profiles

// Profiles returns the built-in profiles followed by the user's saved ones.
func (s *MatingService) Profiles(ctx context.Context) ([]engine.Profile, error) {
	out := engine.Builtins()
	rows, err := s.db.QueryContext(ctx, `SELECT id, config FROM mating_profiles ORDER BY name COLLATE NOCASE`)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDatabase, "Falha ao listar os perfis.", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id, cfg string
		if err := rows.Scan(&id, &cfg); err != nil {
			return nil, apperr.Wrap(apperr.CodeDatabase, "Falha ao ler os perfis.", err)
		}
		var p engine.Profile
		if err := json.Unmarshal([]byte(cfg), &p); err != nil {
			continue // a corrupted saved profile must not hide the others
		}
		p.ID, p.IsDemo, p.DemoNotice = id, false, ""
		out = append(out, p)
	}
	return out, rows.Err()
}

// SaveProfile stores a user profile (new, or an update when the ID is a saved one).
func (s *MatingService) SaveProfile(ctx context.Context, p engine.Profile) (*engine.Profile, error) {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return nil, fieldError("name", "Informe um nome para o perfil.")
	}
	if issues := engine.ValidateProfile(p); len(issues) > 0 {
		f := map[string]string{}
		for _, i := range issues {
			if _, ok := f[i.Field]; !ok {
				f[i.Field] = i.Message
			}
		}
		return nil, &domain.ValidationError{Fields: f}
	}
	p.IsDemo, p.DemoNotice = false, ""
	now := nowTS()
	update := p.ID != "" && !strings.HasPrefix(p.ID, "builtin:")
	if !update {
		p.ID = newID()
	}
	cfg, _ := json.Marshal(p)
	var err error
	if update {
		var res sql.Result
		res, err = s.db.ExecContext(ctx, `UPDATE mating_profiles SET name=?, config=?, updated_at=? WHERE id=?`, p.Name, string(cfg), now, p.ID)
		if err == nil {
			if n, _ := res.RowsAffected(); n == 0 {
				return nil, apperr.New(apperr.CodeNotFound, "Perfil não encontrado.")
			}
		}
	} else {
		_, err = s.db.ExecContext(ctx, `INSERT INTO mating_profiles (id, name, config, created_at, updated_at) VALUES (?,?,?,?,?)`, p.ID, p.Name, string(cfg), now, now)
	}
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return nil, fieldError("name", "Já existe um perfil com esse nome.")
		}
		return nil, apperr.Wrap(apperr.CodeDatabase, "Falha ao salvar o perfil.", err)
	}
	return &p, nil
}

func (s *MatingService) DeleteProfile(ctx context.Context, id string) error {
	if strings.HasPrefix(id, "builtin:") {
		return apperr.New(apperr.CodeValidation, "Perfis embutidos não podem ser excluídos.")
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM mating_profiles WHERE id=?`, id)
	if err != nil {
		return apperr.Wrap(apperr.CodeDatabase, "Falha ao excluir o perfil.", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return apperr.New(apperr.CodeNotFound, "Perfil não encontrado.")
	}
	return nil
}
