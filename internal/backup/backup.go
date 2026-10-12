// Package backup creates, validates and restores .genbackup files and checks
// the integrity of the local data. A backup is a ZIP holding a consistent
// copy of farm.db, the stored private library packages, optional attachments
// and a manifest with the SHA-256 of every file.
package backup

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"genetica/internal/apperr"
	"genetica/internal/database"
)

const (
	FormatVersion = 1
	Extension     = ".genbackup"

	manifestName = "backup-manifest.json"
	dbName       = "farm.db"
	configName   = "config.json"
)

// Limits protect restore against hostile files (zip bombs, huge entries).
type Limits struct {
	MaxFiles     int
	MaxFileBytes int64
	MaxTotal     int64
}

var DefaultLimits = Limits{MaxFiles: 20000, MaxFileBytes: 4 << 30, MaxTotal: 8 << 30}

// Service performs backup operations on one data directory.
type Service struct {
	DB            *database.DB
	DataDir       string
	LibrariesDir  string
	Migrations    fs.FS // migrations used to upgrade older backups
	AppVersion    string
	EngineVersion string
	Limits        Limits
	Now           func() time.Time
}

func (s *Service) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func (s *Service) limits() Limits {
	if s.Limits.MaxFiles == 0 {
		return DefaultLimits
	}
	return s.Limits
}

// FileEntry describes one file inside a backup.
type FileEntry struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

// LibraryEntry records an installed library. Official libraries (shipped with
// the application, always re-obtainable) are recorded as metadata only; every
// library the user imported is included, whatever its license says, because the
// user may be the only one who has the file.
type LibraryEntry struct {
	ID            string `json:"id"`
	PackageID     string `json:"packageId"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	FileName      string `json:"fileName"`
	SHA256        string `json:"sha256"`
	LicenseStatus string `json:"licenseStatus"`
	IsDemo        bool   `json:"isDemo"`
	IsOfficial    bool   `json:"isOfficial"`
	Included      bool   `json:"included"`
}

// Manifest is backup-manifest.json.
type Manifest struct {
	FormatVersion int            `json:"formatVersion"`
	CreatedAt     string         `json:"createdAt"`
	AppVersion    string         `json:"appVersion"`
	EngineVersion string         `json:"engineVersion"`
	SchemaVersion int            `json:"schemaVersion"`
	Farms         []string       `json:"farms"`
	Counts        map[string]int `json:"counts"`
	Files         []FileEntry    `json:"files"`
	Libraries     []LibraryEntry `json:"libraries"`
}

var countedTables = []string{"farms", "animals", "animal_measurements", "genetic_evaluations", "semen_inventory", "semen_movements", "matings", "datasets", "traits"}

func sha256File(p string) (string, int64, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil)), n, err
}

var slugRE = regexp.MustCompile(`[^a-z0-9]+`)

// SuggestedName returns e.g. "2026-10-11-fazenda-exemplo.genbackup".
func (s *Service) SuggestedName(ctx context.Context) string {
	name := "propriedade"
	var farm string
	if err := s.DB.QueryRowContext(ctx, `SELECT name FROM farms WHERE is_demo=0 ORDER BY created_at LIMIT 1`).Scan(&farm); err == nil && strings.TrimSpace(farm) != "" {
		name = strings.Trim(slugRE.ReplaceAllString(strings.ToLower(asciiFold(farm)), "-"), "-")
		if name == "" {
			name = "propriedade"
		}
		if len(name) > 40 {
			name = strings.Trim(name[:40], "-")
		}
	}
	return s.now().Format("2006-01-02") + "-" + name + Extension
}

func asciiFold(s string) string {
	r := strings.NewReplacer("á", "a", "à", "a", "â", "a", "ã", "a", "é", "e", "ê", "e", "í", "i", "ó", "o", "ô", "o", "õ", "o", "ú", "u", "ü", "u", "ç", "c",
		"Á", "A", "À", "A", "Â", "A", "Ã", "A", "É", "E", "Ê", "E", "Í", "I", "Ó", "O", "Ô", "O", "Õ", "O", "Ú", "U", "Ç", "C")
	return r.Replace(s)
}

// Create writes a consistent backup to dest (atomically: temp file + rename).
func (s *Service) Create(ctx context.Context, dest string) (*Manifest, error) {
	if !strings.EqualFold(filepath.Ext(dest), Extension) {
		dest += Extension
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "Não foi possível criar a pasta do backup.", err)
	}
	stage, err := os.MkdirTemp(s.DataDir, "backup-stage-*")
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "Não foi possível preparar o backup.", err)
	}
	defer os.RemoveAll(stage)

	// Consistent snapshot, safe even with WAL and concurrent readers.
	snap := filepath.Join(stage, dbName)
	if _, err := s.DB.ExecContext(ctx, "VACUUM INTO '"+strings.ReplaceAll(snap, "'", "''")+"'"); err != nil {
		return nil, apperr.Wrap(apperr.CodeDatabase, "Falha ao copiar o banco de dados.", err)
	}
	m := &Manifest{FormatVersion: FormatVersion, CreatedAt: s.now().Format(time.RFC3339), AppVersion: s.AppVersion, EngineVersion: s.EngineVersion,
		Counts: map[string]int{}, Farms: []string{}, Libraries: []LibraryEntry{}}
	_ = s.DB.QueryRowContext(ctx, `SELECT COALESCE(MAX(version),0) FROM schema_migrations`).Scan(&m.SchemaVersion)
	for _, t := range countedTables {
		var n int
		if err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+t).Scan(&n); err == nil {
			m.Counts[t] = n
		}
	}
	if rows, err := s.DB.QueryContext(ctx, `SELECT name FROM farms ORDER BY name`); err == nil {
		for rows.Next() {
			var n string
			if rows.Scan(&n) == nil {
				m.Farms = append(m.Farms, n)
			}
		}
		rows.Close()
	}

	type item struct{ src, rel string }
	items := []item{{snap, dbName}}
	if _, err := os.Stat(filepath.Join(s.DataDir, configName)); err == nil {
		items = append(items, item{filepath.Join(s.DataDir, configName), configName})
	}

	// Libraries: all user-imported packages are included; official ones only as metadata.
	lrows, err := s.DB.QueryContext(ctx, `SELECT id, package_id, name, version, file_name, file_sha256, license_status, is_demo, is_official FROM datasets ORDER BY name, version`)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDatabase, "Falha ao listar as bibliotecas.", err)
	}
	for lrows.Next() {
		var e LibraryEntry
		var demo, official int
		if err := lrows.Scan(&e.ID, &e.PackageID, &e.Name, &e.Version, &e.FileName, &e.SHA256, &e.LicenseStatus, &demo, &official); err != nil {
			lrows.Close()
			return nil, apperr.Wrap(apperr.CodeDatabase, "Falha ao ler as bibliotecas.", err)
		}
		e.IsDemo, e.IsOfficial = demo == 1, official == 1
		if e.FileName != "" && !e.IsOfficial {
			src := filepath.Join(s.LibrariesDir, e.FileName)
			if _, err := os.Stat(src); err == nil {
				e.Included = true
				items = append(items, item{src, "libraries/" + e.FileName})
			}
		}
		m.Libraries = append(m.Libraries, e)
	}
	lrows.Close()

	// Attachments / photos, when the directory exists.
	attDir := filepath.Join(s.DataDir, "attachments")
	if ents, err := os.ReadDir(attDir); err == nil {
		for _, e := range ents {
			if e.Type().IsRegular() {
				items = append(items, item{filepath.Join(attDir, e.Name()), "attachments/" + e.Name()})
			}
		}
	}

	for _, it := range items {
		sum, size, err := sha256File(it.src)
		if err != nil {
			return nil, apperr.Wrap(apperr.CodeInternal, "Não foi possível ler "+it.rel+".", err)
		}
		m.Files = append(m.Files, FileEntry{Path: it.rel, Size: size, SHA256: sum})
	}

	tmp := dest + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "Não foi possível criar o arquivo de backup.", err)
	}
	zw := zip.NewWriter(out)
	fail := func(err error) (*Manifest, error) {
		_ = zw.Close()
		_ = out.Close()
		_ = os.Remove(tmp)
		return nil, apperr.Wrap(apperr.CodeInternal, "Falha ao gravar o backup.", err)
	}
	mj, _ := json.MarshalIndent(m, "", "  ")
	w, err := zw.Create(manifestName)
	if err != nil {
		return fail(err)
	}
	if _, err := w.Write(mj); err != nil {
		return fail(err)
	}
	for _, it := range items {
		w, err := zw.Create(it.rel)
		if err != nil {
			return fail(err)
		}
		f, err := os.Open(it.src)
		if err != nil {
			return fail(err)
		}
		_, err = io.Copy(w, f)
		f.Close()
		if err != nil {
			return fail(err)
		}
	}
	if err := zw.Close(); err != nil {
		return fail(err)
	}
	if err := out.Sync(); err != nil {
		return fail(err)
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmp)
		return nil, apperr.Wrap(apperr.CodeInternal, "Falha ao gravar o backup.", err)
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return nil, apperr.Wrap(apperr.CodeInternal, "Falha ao finalizar o backup.", err)
	}
	return m, nil
}

// ---- reading and validating a backup

var (
	libraryEntryRE    = regexp.MustCompile(`^libraries/[0-9a-f-]{36}\.genlib$`)
	attachmentEntryRE = regexp.MustCompile(`^attachments/[A-Za-z0-9._-]{1,120}$`)
)

func allowedEntry(name string) error {
	switch {
	case name == "" || strings.ContainsAny(name, "\\\x00") || strings.HasPrefix(name, "/") || strings.Contains(name, "..") || strings.Contains(name, ":"):
		return fmt.Errorf("caminho suspeito no backup: %q", name)
	case path.Clean(name) != name:
		return fmt.Errorf("caminho não normalizado no backup: %q", name)
	case name == manifestName, name == dbName, name == configName, libraryEntryRE.MatchString(name), attachmentEntryRE.MatchString(name):
		return nil
	}
	return fmt.Errorf("arquivo inesperado no backup: %q", name)
}

// InvalidBackupError is a rejected backup; Reason is safe to show to the user.
type InvalidBackupError struct{ Reason string }

func (e *InvalidBackupError) Error() string { return "backup inválido: " + e.Reason }

func invalid(format string, a ...any) error { return &InvalidBackupError{fmt.Sprintf(format, a...)} }

// extracted is a backup unpacked into a private temporary directory.
type extracted struct {
	dir      string
	manifest Manifest
}

func (e *extracted) cleanup() { _ = os.RemoveAll(e.dir) }

// extract validates the ZIP structure and unpacks it. Nothing leaves the temp dir.
func (s *Service) extract(zipPath string) (*extracted, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, invalid("não é um arquivo .genbackup (ZIP) válido")
	}
	defer zr.Close()
	lim := s.limits()
	if len(zr.File) > lim.MaxFiles {
		return nil, invalid("arquivos demais")
	}
	dir, err := os.MkdirTemp(s.DataDir, "restore-*")
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "Não foi possível preparar a restauração.", err)
	}
	ex := &extracted{dir: dir}
	ok := false
	defer func() {
		if !ok {
			ex.cleanup()
		}
	}()

	seen := map[string]bool{}
	var total int64
	for _, f := range zr.File {
		if err := allowedEntry(f.Name); err != nil {
			return nil, invalid("%v", err)
		}
		if f.Mode()&os.ModeSymlink != 0 || f.Mode().IsDir() || !f.Mode().IsRegular() {
			return nil, invalid("entrada que não é arquivo regular: %q", f.Name)
		}
		if seen[f.Name] {
			return nil, invalid("arquivo duplicado: %q", f.Name)
		}
		seen[f.Name] = true
		if int64(f.UncompressedSize64) > lim.MaxFileBytes {
			return nil, invalid("%q é grande demais", f.Name)
		}
		rc, err := f.Open()
		if err != nil {
			return nil, invalid("não foi possível ler %q", f.Name)
		}
		target := filepath.Join(dir, filepath.FromSlash(f.Name))
		if rel, err := filepath.Rel(dir, target); err != nil || strings.HasPrefix(rel, "..") {
			rc.Close()
			return nil, invalid("caminho fora da pasta de restauração: %q", f.Name)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			rc.Close()
			return nil, apperr.Wrap(apperr.CodeInternal, "Falha ao extrair o backup.", err)
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			rc.Close()
			return nil, apperr.Wrap(apperr.CodeInternal, "Falha ao extrair o backup.", err)
		}
		n, err := io.Copy(out, io.LimitReader(rc, lim.MaxFileBytes+1))
		rc.Close()
		out.Close()
		if err != nil {
			return nil, invalid("erro ao descompactar %q", f.Name)
		}
		if n > lim.MaxFileBytes {
			return nil, invalid("%q excede o tamanho máximo", f.Name)
		}
		total += n
		if total > lim.MaxTotal {
			return nil, invalid("conteúdo total grande demais")
		}
	}

	mb, err := os.ReadFile(filepath.Join(dir, manifestName))
	if err != nil {
		return nil, invalid("faltando %s", manifestName)
	}
	if err := json.Unmarshal(mb, &ex.manifest); err != nil {
		return nil, invalid("%s ilegível: %v", manifestName, err)
	}
	if ex.manifest.FormatVersion != FormatVersion {
		return nil, invalid("formato de backup versão %d não suportado (este aplicativo lê a versão %d)", ex.manifest.FormatVersion, FormatVersion)
	}
	if !seen[dbName] {
		return nil, invalid("o backup não contém %s", dbName)
	}

	// integrity: every file listed must be present with the right hash, and nothing else.
	listed := map[string]bool{manifestName: true}
	for _, fe := range ex.manifest.Files {
		if err := allowedEntry(fe.Path); err != nil || fe.Path == manifestName {
			return nil, invalid("o manifesto lista um caminho inválido: %q", fe.Path)
		}
		listed[fe.Path] = true
		sum, size, err := sha256File(filepath.Join(dir, filepath.FromSlash(fe.Path)))
		if err != nil {
			return nil, invalid("arquivo listado no manifesto está ausente: %q", fe.Path)
		}
		if sum != fe.SHA256 || size != fe.Size {
			return nil, invalid("o arquivo %q está corrompido ou foi alterado (SHA-256 não confere)", fe.Path)
		}
	}
	for name := range seen {
		if !listed[name] {
			return nil, invalid("o arquivo %q não consta no manifesto", name)
		}
	}
	ok = true
	return ex, nil
}

// latestVersion is the highest migration version this application knows.
func latestVersion(fsys fs.FS) int {
	ents, _ := fs.ReadDir(fsys, ".")
	max := 0
	re := regexp.MustCompile(`^(\d{4})_.*\.sql$`)
	for _, e := range ents {
		if m := re.FindStringSubmatch(e.Name()); m != nil {
			var v int
			fmt.Sscanf(m[1], "%d", &v)
			if v > max {
				max = v
			}
		}
	}
	return max
}

// verifyDB opens the extracted database, checks it and upgrades older schemas.
func (s *Service) verifyDB(ctx context.Context, ex *extracted) (*Preview, error) {
	dbPath := filepath.Join(ex.dir, dbName)
	db, err := database.Open(dbPath)
	if err != nil {
		return nil, invalid("o banco de dados do backup não pôde ser aberto")
	}
	defer db.Close()
	var res string
	if err := db.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&res); err != nil || res != "ok" {
		return nil, invalid("o banco de dados do backup está corrompido (%s)", res)
	}
	var ver int
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version),0) FROM schema_migrations`).Scan(&ver); err != nil {
		return nil, invalid("o banco de dados do backup não tem controle de versão do esquema")
	}
	latest := latestVersion(s.Migrations)
	if ver > latest {
		return nil, invalid("o backup foi criado por uma versão mais nova do aplicativo (esquema %d; este aplicativo entende até %d). Atualize o aplicativo.", ver, latest)
	}
	if ex.manifest.SchemaVersion != ver {
		return nil, invalid("a versão do esquema do manifesto (%d) não confere com a do banco (%d)", ex.manifest.SchemaVersion, ver)
	}
	if s.Migrations != nil {
		if err := db.Migrate(ctx, s.Migrations); err != nil {
			return nil, invalid("não foi possível atualizar o backup para a versão atual: %v", err)
		}
	}
	if rows, err := db.QueryContext(ctx, `PRAGMA foreign_key_check`); err == nil {
		bad := rows.Next()
		rows.Close()
		if bad {
			return nil, invalid("o backup contém referências quebradas entre registros")
		}
	}
	_, _ = db.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`)
	return &Preview{Manifest: ex.manifest, UpgradedFrom: ver, SchemaVersion: latest}, nil
}

// Preview summarizes a validated backup, shown before the user confirms a restore.
type Preview struct {
	Manifest      Manifest `json:"manifest"`
	UpgradedFrom  int      `json:"upgradedFrom"`  // schema version stored in the backup
	SchemaVersion int      `json:"schemaVersion"` // version after upgrade (current)
	Path          string   `json:"path"`
}

// Inspect fully validates a backup (structure, hashes, SQLite integrity, version)
// without changing anything.
func (s *Service) Inspect(ctx context.Context, backupPath string) (*Preview, error) {
	ex, err := s.extract(backupPath)
	if err != nil {
		return nil, wrap(err)
	}
	defer ex.cleanup()
	p, err := s.verifyDB(ctx, ex)
	if err != nil {
		return nil, wrap(err)
	}
	p.Path = backupPath
	return p, nil
}

func wrap(err error) error {
	var ib *InvalidBackupError
	if errors.As(err, &ib) {
		return apperr.Wrap(apperr.CodeValidation, ib.Error(), err)
	}
	return err
}

// RestoreResult reports what happened.
type RestoreResult struct {
	SafetyBackup     string         `json:"safetyBackup"`
	Counts           map[string]int `json:"counts"`
	MissingLibraries []string       `json:"missingLibraries"` // libraries recorded only as metadata
}

// Restore validates the backup, writes an automatic safety backup of the
// current data, and then replaces the database contents in ONE transaction
// (all or nothing). It never overwrites anything silently.
func (s *Service) Restore(ctx context.Context, backupPath string) (*RestoreResult, error) {
	ex, err := s.extract(backupPath)
	if err != nil {
		return nil, wrap(err)
	}
	defer ex.cleanup()
	if _, err := s.verifyDB(ctx, ex); err != nil {
		return nil, wrap(err)
	}

	// 1. automatic safety backup of the current data
	if err := os.MkdirAll(filepath.Join(s.DataDir, "backups"), 0o700); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "Não foi possível criar a pasta de backups.", err)
	}
	safety := filepath.Join(s.DataDir, "backups", "pre-restore-"+s.now().Format("20060102-150405")+Extension)
	if _, err := s.Create(ctx, safety); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "Não foi possível criar o backup de segurança da base atual; a restauração foi cancelada.", err)
	}

	// 2. replace the contents atomically on one pinned connection
	if err := s.replaceDatabase(ctx, filepath.Join(ex.dir, dbName)); err != nil {
		return nil, apperr.Wrap(apperr.CodeDatabase, "A restauração falhou e a base atual foi mantida. Um backup de segurança foi salvo em "+safety+".", err)
	}

	// 3. files that belong to the restored data
	res := &RestoreResult{SafetyBackup: safety, Counts: ex.manifest.Counts}
	if err := s.restoreFiles(ctx, ex); err != nil {
		return res, apperr.Wrap(apperr.CodeInternal, "Os dados foram restaurados, mas alguns arquivos não puderam ser copiados: "+err.Error(), err)
	}
	for _, l := range ex.manifest.Libraries {
		if l.FileName != "" && !l.Included {
			res.MissingLibraries = append(res.MissingLibraries, l.Name+" "+l.Version)
		}
	}
	return res, nil
}

func quoteIdent(s string) string { return `"` + strings.ReplaceAll(s, `"`, `""`) + `"` }

func (s *Service) replaceDatabase(ctx context.Context, restored string) error {
	conn, err := s.DB.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	exec := func(q string, args ...any) error { _, err := conn.ExecContext(ctx, q, args...); return err }

	if err := exec(`ATTACH DATABASE ? AS restored`, restored); err != nil {
		return err
	}
	defer func() { _ = exec(`DETACH DATABASE restored`) }()
	if err := exec(`PRAGMA foreign_keys = OFF`); err != nil {
		return err
	}
	defer func() { _ = exec(`PRAGMA foreign_keys = ON`) }()

	tables, err := tableNames(ctx, conn, "restored")
	if err != nil {
		return err
	}
	current, err := tableNames(ctx, conn, "main")
	if err != nil {
		return err
	}
	if err := exec(`BEGIN IMMEDIATE`); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = exec(`ROLLBACK`)
		}
	}()
	for _, t := range current {
		if err := exec(`DELETE FROM main.` + quoteIdent(t)); err != nil {
			return fmt.Errorf("limpando %s: %w", t, err)
		}
	}
	inRestored := map[string]bool{}
	for _, t := range tables {
		inRestored[t] = true
	}
	for _, t := range current {
		if !inRestored[t] {
			continue
		}
		cols, err := sharedColumns(ctx, conn, t)
		if err != nil {
			return err
		}
		if len(cols) == 0 {
			continue
		}
		list := strings.Join(cols, ", ")
		if err := exec(`INSERT INTO main.` + quoteIdent(t) + ` (` + list + `) SELECT ` + list + ` FROM restored.` + quoteIdent(t)); err != nil {
			return fmt.Errorf("copiando %s: %w", t, err)
		}
	}
	rows, err := conn.QueryContext(ctx, `PRAGMA main.foreign_key_check`)
	if err != nil {
		return err
	}
	bad := rows.Next()
	rows.Close()
	if bad {
		return errors.New("a base restaurada tem referências quebradas")
	}
	if err := exec(`COMMIT`); err != nil {
		return err
	}
	committed = true
	return nil
}

func tableNames(ctx context.Context, conn *sql.Conn, schema string) ([]string, error) {
	rows, err := conn.QueryContext(ctx, `SELECT name FROM `+schema+`.sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// sharedColumns returns the quoted columns present in both main and restored.
func sharedColumns(ctx context.Context, conn *sql.Conn, table string) ([]string, error) {
	get := func(schema string) (map[string]int, []string, error) {
		rows, err := conn.QueryContext(ctx, `PRAGMA `+schema+`.table_info(`+quoteIdent(table)+`)`)
		if err != nil {
			return nil, nil, err
		}
		defer rows.Close()
		set := map[string]int{}
		var order []string
		for rows.Next() {
			var cid int
			var name, typ string
			var notnull, pk int
			var dflt sql.NullString
			if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
				return nil, nil, err
			}
			set[name] = cid
			order = append(order, name)
		}
		return set, order, rows.Err()
	}
	rset, _, err := get("restored")
	if err != nil {
		return nil, err
	}
	_, morder, err := get("main")
	if err != nil {
		return nil, err
	}
	var cols []string
	for _, c := range morder {
		if _, ok := rset[c]; ok {
			cols = append(cols, quoteIdent(c))
		}
	}
	return cols, nil
}

// restoreFiles writes the library packages, config and attachments from the backup
// and removes library packages that no longer belong to the restored data.
func (s *Service) restoreFiles(ctx context.Context, ex *extracted) error {
	if err := os.MkdirAll(s.LibrariesDir, 0o700); err != nil {
		return err
	}
	keep := map[string]bool{}
	rows, err := s.DB.QueryContext(ctx, `SELECT file_name FROM datasets WHERE file_name <> ''`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var n string
		if rows.Scan(&n) == nil {
			keep[n] = true
		}
	}
	rows.Close()

	var firstErr error
	note := func(err error) {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	copyFile := func(src, dst string) {
		tmp := dst + ".tmp"
		in, err := os.Open(src)
		if err != nil {
			note(err)
			return
		}
		defer in.Close()
		out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
		if err != nil {
			note(err)
			return
		}
		_, err = io.Copy(out, in)
		if cerr := out.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			_ = os.Remove(tmp)
			note(err)
			return
		}
		note(os.Rename(tmp, dst))
	}

	for _, fe := range ex.manifest.Files {
		src := filepath.Join(ex.dir, filepath.FromSlash(fe.Path))
		switch {
		case strings.HasPrefix(fe.Path, "libraries/"):
			copyFile(src, filepath.Join(s.LibrariesDir, path.Base(fe.Path)))
		case fe.Path == configName:
			copyFile(src, filepath.Join(s.DataDir, configName))
		case strings.HasPrefix(fe.Path, "attachments/"):
			_ = os.MkdirAll(filepath.Join(s.DataDir, "attachments"), 0o700)
			copyFile(src, filepath.Join(s.DataDir, "attachments", path.Base(fe.Path)))
		}
	}
	// Packages that are no longer part of the data are moved aside, not deleted:
	// they are also in the safety backup, and this keeps the step recoverable.
	if ents, err := os.ReadDir(s.LibrariesDir); err == nil {
		for _, e := range ents {
			if strings.HasSuffix(e.Name(), ".genlib") && !keep[e.Name()] {
				aside := filepath.Join(s.DataDir, "backups", "replaced-libraries")
				if err := os.MkdirAll(aside, 0o700); err != nil {
					note(err)
					continue
				}
				note(os.Rename(filepath.Join(s.LibrariesDir, e.Name()), filepath.Join(aside, e.Name())))
			}
		}
	}
	return firstErr
}

// ListAutomatic lists the automatic (pre-restore) backups, newest first.
func (s *Service) ListAutomatic() []string {
	ents, err := os.ReadDir(filepath.Join(s.DataDir, "backups"))
	if err != nil {
		return []string{}
	}
	var out []string
	for _, e := range ents {
		if e.Type().IsRegular() && strings.HasSuffix(e.Name(), Extension) {
			out = append(out, filepath.Join(s.DataDir, "backups", e.Name()))
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(out)))
	if out == nil {
		out = []string{}
	}
	return out
}
