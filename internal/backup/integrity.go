package backup

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"genetica/internal/genetics"
)

// Issue severities.
const (
	SevError   = "error"
	SevWarning = "warning"
)

// Issue is one finding of the integrity check.
type Issue struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Detail   string `json:"detail"`
}

// Report is the result of an integrity check. OK is true when there are no errors
// (warnings do not make the data unusable).
type Report struct {
	OK        bool           `json:"ok"`
	CheckedAt string         `json:"checkedAt"`
	Errors    int            `json:"errors"`
	Warnings  int            `json:"warnings"`
	Issues    []Issue        `json:"issues"`
	Checked   map[string]int `json:"checked"`
}

var requiredTables = []string{"schema_migrations", "farms", "breeds", "animals", "animal_measurements", "genetic_evaluations", "semen_inventory", "semen_movements",
	"matings", "datasets", "traits", "calculation_datasets", "calculation_runs", "mating_profiles", "semen_types"}

// CheckIntegrity inspects the whole local data set. It never modifies anything.
func (s *Service) CheckIntegrity(ctx context.Context) (*Report, error) {
	r := &Report{CheckedAt: s.now().Format(time.RFC3339), Issues: []Issue{}, Checked: map[string]int{}}
	add := func(sev, code, msg, detail string) {
		r.Issues = append(r.Issues, Issue{sev, code, msg, detail})
		if sev == SevError {
			r.Errors++
		} else {
			r.Warnings++
		}
	}

	// 1. SQLite file integrity
	rows, err := s.DB.QueryContext(ctx, `PRAGMA integrity_check`)
	if err != nil {
		add(SevError, "sqlite_corrupt", "O banco de dados não pôde ser verificado.", err.Error())
	} else {
		var problems []string
		for rows.Next() {
			var line string
			if rows.Scan(&line) == nil && line != "ok" {
				problems = append(problems, line)
			}
		}
		rows.Close()
		if len(problems) > 0 {
			if len(problems) > 5 {
				problems = append(problems[:5], fmt.Sprintf("… e mais %d problema(s)", len(problems)-5))
			}
			add(SevError, "sqlite_corrupt", "O arquivo do banco de dados está corrompido.", strings.Join(problems, "; "))
		}
	}

	// 2. migrations: complete, contiguous, not newer than the application
	versions := map[int]bool{}
	maxV := 0
	if vr, err := s.DB.QueryContext(ctx, `SELECT version FROM schema_migrations ORDER BY version`); err != nil {
		add(SevError, "migration_incomplete", "O controle de migrations não pôde ser lido.", err.Error())
	} else {
		for vr.Next() {
			var v int
			if vr.Scan(&v) == nil {
				versions[v] = true
				if v > maxV {
					maxV = v
				}
			}
		}
		vr.Close()
	}
	latest := latestVersion(s.Migrations)
	switch {
	case latest > 0 && maxV > latest:
		add(SevError, "version_incompatible", fmt.Sprintf("O banco de dados (versão %d) é mais novo que este aplicativo (versão %d).", maxV, latest), "Atualize o aplicativo para evitar perda de dados.")
	case latest > 0 && maxV < latest:
		add(SevError, "migration_incomplete", fmt.Sprintf("Migrations pendentes: o banco está na versão %d e o aplicativo espera %d.", maxV, latest), "Reinicie o aplicativo para aplicar as migrations.")
	}
	for v := 1; v <= maxV; v++ {
		if !versions[v] {
			add(SevError, "migration_gap", fmt.Sprintf("A migration %04d não consta como aplicada.", v), "")
		}
	}
	r.Checked["migrations"] = len(versions)
	existing := map[string]bool{}
	if tr, err := s.DB.QueryContext(ctx, `SELECT name FROM sqlite_master WHERE type='table'`); err == nil {
		for tr.Next() {
			var n string
			if tr.Scan(&n) == nil {
				existing[n] = true
			}
		}
		tr.Close()
	}
	for _, t := range requiredTables {
		if !existing[t] {
			add(SevError, "table_missing", "Tabela ausente: "+t+".", "O esquema do banco está incompleto.")
		}
	}

	// 3. broken references
	if fr, err := s.DB.QueryContext(ctx, `PRAGMA foreign_key_check`); err == nil {
		n := 0
		for fr.Next() {
			var table, parent sql.NullString
			var rowid, fkid sql.NullInt64
			if fr.Scan(&table, &rowid, &parent, &fkid) == nil {
				n++
				if n <= 20 {
					add(SevError, "broken_reference", fmt.Sprintf("Referência quebrada: um registro de %s aponta para %s que não existe.", table.String, parent.String), fmt.Sprintf("linha %d", rowid.Int64))
				}
			}
		}
		fr.Close()
		if n > 20 {
			add(SevError, "broken_reference", fmt.Sprintf("… e mais %d referências quebradas.", n-20), "")
		}
	}

	// 4. library files: missing, changed, orphaned
	known := map[string]bool{}
	if dr, err := s.DB.QueryContext(ctx, `SELECT id, name, version, file_name, file_sha256 FROM datasets WHERE file_name <> ''`); err == nil {
		type ds struct{ id, name, version, file, sha string }
		var list []ds
		for dr.Next() {
			var d ds
			if dr.Scan(&d.id, &d.name, &d.version, &d.file, &d.sha) == nil {
				list = append(list, d)
			}
		}
		dr.Close()
		for _, d := range list {
			known[d.file] = true
			p := filepath.Join(s.LibrariesDir, d.file)
			if _, err := os.Stat(p); err != nil {
				add(SevWarning, "dataset_file_missing", fmt.Sprintf("O arquivo original da biblioteca %s %s não foi encontrado.", d.name, d.version),
					"Os valores continuam no banco, mas o pacote .genlib original está ausente (restaure um backup ou importe de novo).")
				continue
			}
			if d.sha != "" {
				if sum, _, err := sha256File(p); err == nil && sum != d.sha {
					add(SevError, "dataset_file_changed", fmt.Sprintf("O arquivo da biblioteca %s %s foi alterado (SHA-256 diferente do instalado).", d.name, d.version), "")
				}
			}
		}
		r.Checked["libraries"] = len(list)
	}
	if ents, err := os.ReadDir(s.LibrariesDir); err == nil {
		for _, e := range ents {
			if strings.HasSuffix(e.Name(), ".genlib") && !known[e.Name()] {
				add(SevWarning, "orphan_library_file", "Há um pacote de biblioteca sem registro no banco: "+e.Name()+".", "")
			}
		}
	}

	// 5. saved calculations pointing at libraries that are gone
	if cr, err := s.DB.QueryContext(ctx, `SELECT DISTINCT m.id, json_extract(d.value,'$.name'), json_extract(d.value,'$.version')
		FROM matings m, json_each(m.calculation_snapshot, '$.datasets') d
		WHERE json_extract(d.value,'$.id') NOT IN (SELECT id FROM datasets)`); err == nil {
		for cr.Next() {
			var id, name, version sql.NullString
			if cr.Scan(&id, &name, &version) == nil {
				add(SevWarning, "calculation_dataset_removed", fmt.Sprintf("Um cálculo salvo usou a biblioteca %s %s, que não está mais instalada.", name.String, version.String),
					"O snapshot do cálculo mantém os valores usados; ele continua reproduzível, mas a biblioteca não pode ser reativada. Cálculo: "+id.String)
			}
		}
		cr.Close()
	}

	// 6. stock vs movement history
	if sr, err := s.DB.QueryContext(ctx, `SELECT i.id, i.quantity, SUM(m.delta) FROM semen_inventory i JOIN semen_movements m ON m.inventory_id = i.id
		GROUP BY i.id HAVING i.quantity <> SUM(m.delta)`); err == nil {
		for sr.Next() {
			var id string
			var q, sum int
			if sr.Scan(&id, &q, &sum) == nil {
				add(SevWarning, "stock_mismatch", fmt.Sprintf("O estoque de um lote (%d) não confere com a soma das movimentações (%d).", q, sum), "Lote: "+id)
			}
		}
		sr.Close()
	}

	// 7. pedigree consistency
	g := genetics.Graph{}
	var animalCount int
	if ar, err := s.DB.QueryContext(ctx, `SELECT id, COALESCE(name,''), COALESCE(registration,''), sex, COALESCE(birth_date,''), COALESCE(father_id,''), COALESCE(mother_id,'') FROM animals`); err == nil {
		for ar.Next() {
			var n genetics.Node
			if ar.Scan(&n.ID, &n.Name, &n.Registration, &n.Sex, &n.BirthDate, &n.FatherID, &n.MotherID) == nil {
				cp := n
				g[n.ID] = &cp
				animalCount++
			}
		}
		ar.Close()
	}
	r.Checked["animals"] = animalCount
	issues := genetics.CheckConsistency(g)
	sort.SliceStable(issues, func(i, j int) bool { return issues[i].Code < issues[j].Code })
	for i, is := range issues {
		if i >= 30 {
			add(SevWarning, "pedigree", fmt.Sprintf("… e mais %d inconsistência(s) de pedigree.", len(issues)-30), "")
			break
		}
		add(SevWarning, "pedigree_"+is.Code, is.Message, "Animal: "+is.AnimalID)
	}

	r.OK = r.Errors == 0
	return r, nil
}
