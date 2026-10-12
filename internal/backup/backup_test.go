package backup_test

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"genetica/internal/apperr"
	"genetica/internal/backup"
	"genetica/internal/database"
	"genetica/internal/datasets"
	"genetica/internal/domain"
	"genetica/internal/service"
	"genetica/migrations"
)

type env struct {
	db      *database.DB
	svc     *service.Services
	bk      *backup.Service
	dataDir string
	libDir  string
}

func newEnv(t *testing.T) env {
	t.Helper()
	dir := t.TempDir()
	lib := filepath.Join(dir, "libraries")
	if err := os.MkdirAll(lib, 0o700); err != nil {
		t.Fatal(err)
	}
	db, err := database.Open(filepath.Join(dir, "farm.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate(context.Background(), migrations.FS); err != nil {
		t.Fatal(err)
	}
	return env{db: db, svc: service.New(db, lib), dataDir: dir, libDir: lib,
		bk: &backup.Service{DB: db, DataDir: dir, LibrariesDir: lib, Migrations: migrations.FS, AppVersion: "test", EngineVersion: "e1"}}
}

func count(t *testing.T, db *database.DB, table string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// populate creates a small but complete herd: pedigree, phenotype, a private library, stock and a saved calculation.
func populate(t *testing.T, e env, license datasets.License) (farmID, bullID, cowID string) {
	t.Helper()
	ctx := context.Background()
	farm, err := e.svc.Farms.Create(ctx, service.FarmInput{Name: "Fazenda São João"})
	if err != nil {
		t.Fatal(err)
	}
	mk := func(name, sex, cat, reg string) *domain.Animal {
		a, err := e.svc.Animals.Create(ctx, service.AnimalInput{FarmID: farm.ID, Name: name, Sex: sex, Category: cat, Registration: reg, BreedID: "breed-nelore"})
		if err != nil {
			t.Fatal(err)
		}
		return a
	}
	bull := mk("Touro", domain.SexMale, domain.CategoryBull, "B-1")
	cow := mk("Vaca", domain.SexFemale, domain.CategoryCow, "V-1")
	if _, err := e.svc.Animals.Create(ctx, service.AnimalInput{FarmID: farm.ID, Name: "Bezerra", Sex: domain.SexFemale, Category: domain.CategoryCalf, FatherID: bull.ID, MotherID: cow.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.svc.Measurements.Add(ctx, service.MeasurementInput{AnimalID: cow.ID, TraitCode: "BODY_WEIGHT", Value: 480, MeasuredAt: "2025-05-01"}); err != nil {
		t.Fatal(err)
	}
	// private library
	m := datasets.Manifest{SchemaVersion: 1, ID: "private.lib", Name: "Private Lib", Version: "1.0.0", Species: "Bos taurus", Breeds: []string{"NEL"}, Purpose: []string{"beef"},
		License: license, Evaluation: datasets.Evaluation{Program: "SYN", GeneticBase: "b"}, Data: datasets.DataSpec{Format: "csv", IDColumn: "reg", AccuracyScale: "fraction"}}
	tb, _ := datasets.ParseTable([]byte("reg,ww,pct\nB-1,5,80\nV-1,3,40\n"), "csv")
	raw, err := datasets.WritePackage(&datasets.Package{Manifest: m, Traits: []datasets.TraitDef{{Code: "WW", Name: "WW", Category: "growth", Unit: "kg", EvaluationType: "DEP", Column: "ww", PercentileColumn: "pct", Metadata: map[string]any{"percentileSemantics": "ascending"}, OptimizationDirection: "MAXIMIZE"}}, Table: tb})
	if err != nil {
		t.Fatal(err)
	}
	pkg := filepath.Join(t.TempDir(), "private.genlib")
	if err := os.WriteFile(pkg, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := e.svc.Libraries.Install(ctx, service.ImportRequest{Path: pkg, AcceptResponsibility: true}); err != nil {
		t.Fatal(err)
	}
	lot, err := e.svc.Semen.AddLot(ctx, service.SemenInput{BullID: bull.ID, Quantity: 10, Supplier: "Central"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.svc.Semen.Withdraw(ctx, service.MovementInput{InventoryID: lot.ID, Quantity: 2, Reason: "insemination"}); err != nil {
		t.Fatal(err)
	}
	prof := service.RunRequest{FemaleID: cow.ID, BullIDs: []string{bull.ID}, Save: true}
	prof.Profile.Name = "p"
	prof.Profile.Traits = nil
	return farm.ID, bull.ID, cow.ID
}

// ---- zip helpers

func readZip(t *testing.T, path string) map[string][]byte {
	t.Helper()
	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	out := map[string][]byte{}
	for _, f := range zr.File {
		rc, _ := f.Open()
		b, _ := io.ReadAll(rc)
		rc.Close()
		out[f.Name] = b
	}
	return out
}

func writeZip(t *testing.T, path string, files map[string][]byte, order []string) {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	names := order
	if names == nil {
		for n := range files {
			names = append(names, n)
		}
	}
	for _, n := range names {
		w, err := zw.Create(n)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write(files[n])
	}
	_ = zw.Close()
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
}

// tamper rewrites a backup, letting the test change entries; with rehash=true the manifest hashes are recomputed.
func tamper(t *testing.T, src, dst string, rehash bool, mod func(files map[string][]byte)) {
	t.Helper()
	files := readZip(t, src)
	mod(files)
	if rehash {
		var m backup.Manifest
		if err := json.Unmarshal(files["backup-manifest.json"], &m); err != nil {
			t.Fatal(err)
		}
		m.Files = nil
		var names []string
		for n := range files {
			if n != "backup-manifest.json" {
				names = append(names, n)
			}
		}
		for _, n := range names {
			sum := sha256.Sum256(files[n])
			m.Files = append(m.Files, backup.FileEntry{Path: n, Size: int64(len(files[n])), SHA256: hex.EncodeToString(sum[:])})
		}
		files["backup-manifest.json"], _ = json.Marshal(m)
	}
	writeZip(t, dst, files, nil)
}

func wantInvalid(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil || apperr.CodeOf(err) != apperr.CodeValidation || !strings.Contains(apperr.UserMessage(err), substr) {
		t.Fatalf("want validation error containing %q, got %v", substr, err)
	}
}

// ---- tests

func TestBackupRoundTrip(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t)
	populate(t, e, datasets.License{}) // unknown license => private => the package file is included

	name := e.bk.SuggestedName(ctx)
	if !strings.HasSuffix(name, ".genbackup") || !strings.Contains(name, "fazenda-sao-joao") {
		t.Fatalf("suggested name: %s", name)
	}
	dest := filepath.Join(t.TempDir(), "meu.genbackup")
	m, err := e.bk.Create(ctx, dest)
	if err != nil {
		t.Fatal(err)
	}
	if m.FormatVersion != 1 || m.Counts["animals"] != 3 || m.Counts["farms"] != 1 || m.SchemaVersion < 6 || m.AppVersion != "test" || len(m.Farms) != 1 {
		t.Fatalf("%+v", m)
	}
	files := readZip(t, dest)
	var libFile string
	for n := range files {
		if strings.HasPrefix(n, "libraries/") {
			libFile = n
		}
	}
	if libFile == "" || files["farm.db"] == nil || files["backup-manifest.json"] == nil {
		t.Fatalf("missing parts: %v", keys(files))
	}
	if len(m.Libraries) != 1 || !m.Libraries[0].Included || m.Libraries[0].SHA256 == "" {
		t.Fatalf("%+v", m.Libraries)
	}
	if _, err := e.bk.Inspect(ctx, dest); err != nil {
		t.Fatal(err)
	}

	// change the data after the backup
	farm2, _ := e.svc.Farms.Create(ctx, service.FarmInput{Name: "Outra"})
	if _, err := e.db.Exec(`DELETE FROM animal_measurements`); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(e.dataDir, "libraries", filepath.Base(libFile))); err != nil {
		t.Fatal(err)
	}
	if count(t, e.db, "farms") != 2 {
		t.Fatal("setup")
	}

	res, err := e.bk.Restore(ctx, dest)
	if err != nil {
		t.Fatal(err)
	}
	if count(t, e.db, "farms") != 1 || count(t, e.db, "animal_measurements") != 1 || count(t, e.db, "animals") != 3 || count(t, e.db, "semen_movements") != 2 {
		t.Fatal("the restored data must equal the backed-up data")
	}
	var got string
	if err := e.db.QueryRow(`SELECT name FROM farms WHERE id=?`, farm2.ID).Scan(&got); err == nil {
		t.Fatal("data created after the backup must be gone")
	}
	if _, err := os.Stat(filepath.Join(e.libDir, filepath.Base(libFile))); err != nil {
		t.Fatalf("the library package must be restored: %v", err)
	}
	// the services keep working on the same connection after the restore
	if page, err := e.svc.Animals.List(ctx, service.AnimalQuery{}); err != nil || page.Total != 3 {
		t.Fatalf("%v %v", page, err)
	}
	// automatic safety backup of the replaced data: valid and holding the pre-restore state
	if !strings.Contains(filepath.Base(res.SafetyBackup), "pre-restore-") {
		t.Fatalf("%s", res.SafetyBackup)
	}
	p, err := e.bk.Inspect(ctx, res.SafetyBackup)
	if err != nil || p.Manifest.Counts["farms"] != 2 {
		t.Fatalf("the safety backup must hold the state before restoring: %+v %v", p, err)
	}
	if list := e.bk.ListAutomatic(); len(list) != 1 {
		t.Fatalf("%v", list)
	}
	if rep, _ := e.bk.CheckIntegrity(ctx); !rep.OK {
		t.Fatalf("restored data must pass the integrity check: %+v", rep.Issues)
	}
}

func keys(m map[string][]byte) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestUserLibrariesAreAlwaysIncluded(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t)
	// a library the USER imported and declared redistributable: still included, the user may be the only holder of the file
	populate(t, e, datasets.License{Name: "CC0-1.0", URL: "https://creativecommons.org/publicdomain/zero/1.0/", RedistributionAllowed: true})
	userDest := filepath.Join(t.TempDir(), "user.genbackup")
	um, err := e.bk.Create(ctx, userDest)
	if err != nil {
		t.Fatal(err)
	}
	if len(um.Libraries) != 1 || !um.Libraries[0].Included || um.Libraries[0].IsOfficial {
		t.Fatalf("user-imported libraries must be included even when redistributable: %+v", um.Libraries)
	}
	n := 0
	for name := range readZip(t, userDest) {
		if strings.HasPrefix(name, "libraries/") {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("the package file must be in the backup, found %d", n)
	}

	// an OFFICIAL library (shipped with the application) is recorded as metadata only
	if _, err := e.db.Exec(`UPDATE datasets SET is_official=1`); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "b.genbackup")
	m, err := e.bk.Create(ctx, dest)
	if err != nil {
		t.Fatal(err)
	}
	for n := range readZip(t, dest) {
		if strings.HasPrefix(n, "libraries/") {
			t.Fatalf("an official library must not be copied into the backup: %s", n)
		}
	}
	if len(m.Libraries) != 1 || m.Libraries[0].Included || !m.Libraries[0].IsOfficial || m.Libraries[0].LicenseStatus != "redistributable" || m.Libraries[0].PackageID != "private.lib" || m.Libraries[0].Version != "1.0.0" || m.Libraries[0].SHA256 == "" {
		t.Fatalf("version metadata must be kept: %+v", m.Libraries)
	}
	res, err := e.bk.Restore(ctx, dest)
	if err != nil || len(res.MissingLibraries) != 1 {
		t.Fatalf("%+v %v", res, err)
	}
}

// Restoring must never destroy a library package that the restored data no longer references.
func TestRestoreMovesAsideUnreferencedPackages(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t)
	populate(t, e, datasets.License{})
	empty := newEnv(t) // a backup of an empty database
	dest := filepath.Join(t.TempDir(), "empty.genbackup")
	if _, err := empty.bk.Create(ctx, dest); err != nil {
		t.Fatal(err)
	}
	before, _ := filepath.Glob(filepath.Join(e.libDir, "*.genlib"))
	if len(before) != 1 {
		t.Fatal("setup")
	}
	if _, err := e.bk.Restore(ctx, dest); err != nil {
		t.Fatal(err)
	}
	if left, _ := filepath.Glob(filepath.Join(e.libDir, "*.genlib")); len(left) != 0 {
		t.Fatalf("the unreferenced package leaves the library folder: %v", left)
	}
	aside, _ := filepath.Glob(filepath.Join(e.dataDir, "backups", "replaced-libraries", "*.genlib"))
	if len(aside) != 1 {
		t.Fatalf("...but it is kept, not deleted: %v", aside)
	}
}

func TestRestoreRefusesTamperedBackups(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t)
	populate(t, e, datasets.License{})
	good := filepath.Join(t.TempDir(), "good.genbackup")
	if _, err := e.bk.Create(ctx, good); err != nil {
		t.Fatal(err)
	}
	before := count(t, e.db, "animals")
	dir := t.TempDir()

	cases := []struct {
		name   string
		rehash bool
		mod    func(map[string][]byte)
		want   string
	}{
		{"conteúdo alterado (hash não confere)", false, func(f map[string][]byte) { f["farm.db"][len(f["farm.db"])/2] ^= 0xFF }, "SHA-256"},
		{"path traversal", true, func(f map[string][]byte) { f["../evil.txt"] = []byte("x") }, "suspeito"},
		{"caminho absoluto", true, func(f map[string][]byte) { f["/etc/passwd"] = []byte("x") }, "suspeito"},
		{"traversal dentro de libraries", true, func(f map[string][]byte) { f["libraries/../../x.genlib"] = []byte("x") }, "suspeito"},
		{"barra invertida", true, func(f map[string][]byte) { f["libraries\\x.genlib"] = []byte("x") }, "suspeito"},
		{"arquivo inesperado", true, func(f map[string][]byte) { f["evil.sh"] = []byte("rm -rf /") }, "inesperado"},
		{"nome de biblioteca fora do padrão", true, func(f map[string][]byte) { f["libraries/nome-livre.genlib"] = []byte("x") }, "inesperado"},
		{"sem banco", true, func(f map[string][]byte) { delete(f, "farm.db") }, "não contém farm.db"},
		{"sem manifesto", false, func(f map[string][]byte) { delete(f, "backup-manifest.json") }, "backup-manifest.json"},
		{"formato futuro", false, func(f map[string][]byte) {
			var m map[string]any
			_ = json.Unmarshal(f["backup-manifest.json"], &m)
			m["formatVersion"] = 99
			f["backup-manifest.json"], _ = json.Marshal(m)
		}, "formato de backup versão 99"},
		{"arquivo fora do manifesto", false, func(f map[string][]byte) { f["attachments/foto.jpg"] = []byte("x") }, "não consta no manifesto"},
		{"esquema do manifesto diferente do banco", false, func(f map[string][]byte) {
			var m map[string]any
			_ = json.Unmarshal(f["backup-manifest.json"], &m)
			m["schemaVersion"] = 1
			f["backup-manifest.json"], _ = json.Marshal(m)
		}, "não confere"},
	}
	for i, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bad := filepath.Join(dir, "bad"+string(rune('a'+i))+".genbackup")
			tamper(t, good, bad, c.rehash, c.mod)
			_, err := e.bk.Restore(ctx, bad)
			wantInvalid(t, err, c.want)
			if _, err := e.bk.Inspect(ctx, bad); err == nil {
				t.Fatal("inspect must refuse too")
			}
			if count(t, e.db, "animals") != before {
				t.Fatal("a refused restore must not touch the data")
			}
		})
	}
	if len(e.bk.ListAutomatic()) != 0 {
		t.Fatal("no safety backup is created for a backup that fails validation")
	}
	// not a zip at all
	junk := filepath.Join(dir, "junk.genbackup")
	_ = os.WriteFile(junk, []byte("not a zip"), 0o600)
	_, err := e.bk.Restore(ctx, junk)
	wantInvalid(t, err, "ZIP")
	if _, err := e.bk.Restore(ctx, filepath.Join(dir, "missing.genbackup")); err == nil {
		t.Fatal("missing file")
	}
}

func TestRestoreRefusesSymlinkAndDuplicates(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t)
	populate(t, e, datasets.License{})
	good := filepath.Join(t.TempDir(), "good.genbackup")
	if _, err := e.bk.Create(ctx, good); err != nil {
		t.Fatal(err)
	}
	files := readZip(t, good)

	// symlink entry
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for n, b := range files {
		w, _ := zw.Create(n)
		_, _ = w.Write(b)
	}
	h := &zip.FileHeader{Name: "attachments/link"}
	h.SetMode(os.ModeSymlink | 0o777)
	w, _ := zw.CreateHeader(h)
	_, _ = w.Write([]byte("/etc/passwd"))
	_ = zw.Close()
	p := filepath.Join(t.TempDir(), "sym.genbackup")
	_ = os.WriteFile(p, buf.Bytes(), 0o600)
	_, err := e.bk.Restore(ctx, p)
	wantInvalid(t, err, "regular")

	// duplicate entry
	buf.Reset()
	zw = zip.NewWriter(&buf)
	for _, n := range []string{"backup-manifest.json", "farm.db", "farm.db"} {
		w, _ := zw.Create(n)
		_, _ = w.Write(files[n])
	}
	_ = zw.Close()
	p = filepath.Join(t.TempDir(), "dup.genbackup")
	_ = os.WriteFile(p, buf.Bytes(), 0o600)
	_, err = e.bk.Restore(ctx, p)
	wantInvalid(t, err, "duplicado")
}

func TestRestoreRefusesZipBombs(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t)
	populate(t, e, datasets.License{})
	good := filepath.Join(t.TempDir(), "good.genbackup")
	if _, err := e.bk.Create(ctx, good); err != nil {
		t.Fatal(err)
	}
	huge := filepath.Join(t.TempDir(), "huge.genbackup")
	tamper(t, good, huge, true, func(f map[string][]byte) { f["attachments/big.bin"] = bytes.Repeat([]byte{0}, 3<<20) })
	e.bk.Limits = backup.Limits{MaxFiles: 100, MaxFileBytes: 1 << 20, MaxTotal: 8 << 20}
	_, err := e.bk.Restore(ctx, huge)
	wantInvalid(t, err, "grande demais")
	e.bk.Limits = backup.Limits{MaxFiles: 2, MaxFileBytes: 1 << 30, MaxTotal: 1 << 30}
	_, err = e.bk.Restore(ctx, good)
	wantInvalid(t, err, "arquivos demais")
}

func TestRestoreRefusesNewerAndCorruptDatabases(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t)
	populate(t, e, datasets.License{})
	good := filepath.Join(t.TempDir(), "good.genbackup")
	if _, err := e.bk.Create(ctx, good); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()

	// a backup made by a NEWER application version
	newer := filepath.Join(dir, "newer.genbackup")
	tamper(t, good, newer, true, func(f map[string][]byte) {
		p := filepath.Join(dir, "newer.db")
		_ = os.WriteFile(p, f["farm.db"], 0o600)
		db, _ := database.Open(p)
		_, _ = db.Exec(`INSERT INTO schema_migrations (version, name) VALUES (999, '0999_future.sql')`)
		_, _ = db.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`)
		_ = db.Close()
		f["farm.db"], _ = os.ReadFile(p)
		var m map[string]any
		_ = json.Unmarshal(f["backup-manifest.json"], &m)
		m["schemaVersion"] = 999
		f["backup-manifest.json"], _ = json.Marshal(m)
	})
	// manifest edited after hashing is irrelevant (the manifest itself is not hashed): rehash only the db
	files := readZip(t, newer)
	var m backup.Manifest
	_ = json.Unmarshal(files["backup-manifest.json"], &m)
	sum := sha256.Sum256(files["farm.db"])
	for i := range m.Files {
		if m.Files[i].Path == "farm.db" {
			m.Files[i].SHA256, m.Files[i].Size = hex.EncodeToString(sum[:]), int64(len(files["farm.db"]))
		}
	}
	files["backup-manifest.json"], _ = json.Marshal(m)
	writeZip(t, newer, files, nil)
	_, err := e.bk.Restore(ctx, newer)
	wantInvalid(t, err, "versão mais nova")

	// a corrupted SQLite file whose hash is nevertheless correct (corrupted before the backup was made)
	corrupt := filepath.Join(dir, "corrupt.genbackup")
	tamper(t, good, corrupt, true, func(f map[string][]byte) {
		db := f["farm.db"]
		for i := 4096; i < len(db) && i < 4096+2048; i++ {
			db[i] = 0xFF
		}
	})
	_, err = e.bk.Restore(ctx, corrupt)
	wantInvalid(t, err, "banco de dados")
	if count(t, e.db, "animals") != 3 {
		t.Fatal("untouched")
	}
}

func TestRestoreUpgradesOlderBackups(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t)
	populate(t, e, datasets.License{})
	dir := t.TempDir()

	// build a database that only knows migrations 0001–0003 (an old version of the application)
	old := fstest.MapFS{}
	for _, n := range []string{"0001_init.sql", "0002_domain.sql", "0003_traits.sql"} {
		b, _ := migrations.FS.ReadFile(n)
		old[n] = &fstest.MapFile{Data: b}
	}
	oldPath := filepath.Join(dir, "old.db")
	odb, err := database.Open(oldPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := odb.Migrate(ctx, old); err != nil {
		t.Fatal(err)
	}
	if _, err := odb.Exec(`INSERT INTO farms (id, name, notes, created_at, updated_at) VALUES ('old-farm', 'Fazenda Antiga', '', '2025-01-01T00:00:00Z', '2025-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	_, _ = odb.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`)
	_ = odb.Close()
	raw, _ := os.ReadFile(oldPath)
	sum := sha256.Sum256(raw)
	m := backup.Manifest{FormatVersion: 1, CreatedAt: "2025-01-01T00:00:00Z", AppVersion: "0.0.1", SchemaVersion: 3, Counts: map[string]int{"farms": 1}, Farms: []string{"Fazenda Antiga"},
		Files: []backup.FileEntry{{Path: "farm.db", Size: int64(len(raw)), SHA256: hex.EncodeToString(sum[:])}}, Libraries: []backup.LibraryEntry{}}
	mj, _ := json.Marshal(m)
	oldBackup := filepath.Join(dir, "old.genbackup")
	writeZip(t, oldBackup, map[string][]byte{"backup-manifest.json": mj, "farm.db": raw}, []string{"backup-manifest.json", "farm.db"})

	prev, err := e.bk.Inspect(ctx, oldBackup)
	if err != nil || prev.UpgradedFrom != 3 || prev.SchemaVersion < 6 {
		t.Fatalf("%+v %v", prev, err)
	}
	if _, err := e.bk.Restore(ctx, oldBackup); err != nil {
		t.Fatal(err)
	}
	var name string
	if err := e.db.QueryRow(`SELECT name FROM farms WHERE id='old-farm'`).Scan(&name); err != nil || name != "Fazenda Antiga" {
		t.Fatalf("%v %q", err, name)
	}
	if count(t, e.db, "animals") != 0 {
		t.Fatal("an old backup replaces the data (nothing from the newer state remains)")
	}
	var v int
	_ = e.db.QueryRow(`SELECT MAX(version) FROM schema_migrations`).Scan(&v)
	if v < 6 {
		t.Fatalf("the restored data must be at the current schema version, got %d", v)
	}
	// the application still works after restoring an upgraded old backup
	if _, err := e.svc.Libraries.List(ctx); err != nil {
		t.Fatal(err)
	}
	if rep, _ := e.bk.CheckIntegrity(ctx); !rep.OK {
		t.Fatalf("%+v", rep.Issues)
	}
}

func TestFailedRestoreKeepsCurrentData(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t)
	populate(t, e, datasets.License{})
	// a database that passes a superficial look but has a dangling reference (made with FKs off)
	badPath := filepath.Join(t.TempDir(), "bad.db")
	bdb, _ := database.Open(badPath)
	_ = bdb.Migrate(ctx, migrations.FS)
	_, _ = bdb.Exec(`PRAGMA foreign_keys=OFF`)
	_, _ = bdb.Exec(`INSERT INTO animals (id, farm_id, name, sex, category, status, notes, created_at, updated_at) VALUES ('a', 'no-such-farm', 'X', 'male', 'bull', 'active', '', 't', 't')`)
	_, _ = bdb.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`)
	_ = bdb.Close()
	raw, _ := os.ReadFile(badPath)
	var v int
	_ = e.db.QueryRow(`SELECT MAX(version) FROM schema_migrations`).Scan(&v)
	sum := sha256.Sum256(raw)
	m := backup.Manifest{FormatVersion: 1, SchemaVersion: v, Counts: map[string]int{}, Files: []backup.FileEntry{{Path: "farm.db", Size: int64(len(raw)), SHA256: hex.EncodeToString(sum[:])}}, Libraries: []backup.LibraryEntry{}}
	mj, _ := json.Marshal(m)
	p := filepath.Join(t.TempDir(), "bad.genbackup")
	writeZip(t, p, map[string][]byte{"backup-manifest.json": mj, "farm.db": raw}, []string{"backup-manifest.json", "farm.db"})

	_, err := e.bk.Restore(ctx, p)
	if err == nil {
		t.Fatal("a backup with broken references must be refused")
	}
	if count(t, e.db, "animals") != 3 || count(t, e.db, "farms") != 1 {
		t.Fatal("the current data must be untouched")
	}
	var fk int
	_ = e.db.QueryRow(`PRAGMA foreign_keys`).Scan(&fk)
	if fk != 1 {
		t.Fatal("foreign keys must be re-enabled after a failed restore")
	}
}

func TestCreateIsAtomicAndAddsExtension(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t)
	populate(t, e, datasets.License{})
	dest := filepath.Join(t.TempDir(), "sem-extensao")
	if _, err := e.bk.Create(ctx, dest); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dest + ".genbackup"); err != nil {
		t.Fatal("the .genbackup extension must be added")
	}
	ents, _ := os.ReadDir(filepath.Dir(dest))
	for _, en := range ents {
		if strings.HasSuffix(en.Name(), ".tmp") {
			t.Fatalf("leftover temp file %s", en.Name())
		}
	}
	// the staging directory is cleaned up
	left, _ := filepath.Glob(filepath.Join(e.dataDir, "backup-stage-*"))
	if len(left) != 0 {
		t.Fatalf("%v", left)
	}
	// Create never reveals the live database in the zip: a consistent copy is made
	if bytes.Contains(readZip(t, dest+".genbackup")["farm.db"], []byte("-wal")) {
		t.Fatal("unexpected")
	}
}
