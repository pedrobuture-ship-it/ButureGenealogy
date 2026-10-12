package datasets

import (
	"fmt"
	"strings"
)

// Value is one trait value read for one record.
type Value struct {
	TraitIndex int
	Value      float64
	Accuracy   *float64 // fraction in [0,1]
	Percentile *float64 // [0,100]
}

// Record is one animal of the data file with its values.
type Record struct {
	Row          int // 1-based data row, for messages
	Registration string
	Name         string
	Sex          string
	Breed        string
	Values       []Value
}

// Report is the validation summary shown before an import is confirmed.
type Report struct {
	Records             int      `json:"records"` // distinct animals that will be imported
	Rows                int      `json:"rows"`    // non-empty data rows
	RowsWithoutID       int      `json:"rowsWithoutId"`
	Duplicates          int      `json:"duplicates"` // repeated identifiers (first row kept)
	Values              int      `json:"values"`     // numeric values read
	TraitsRecognized    int      `json:"traitsRecognized"`
	TraitsUnrecognized  int      `json:"traitsUnrecognized"` // data columns not mapped to any trait
	UnrecognizedColumns []string `json:"unrecognizedColumns"`
	InvalidValues       int      `json:"invalidValues"`
	WithoutAccuracy     int      `json:"withoutAccuracy"` // values lacking an accuracy
	AccuracyOutOfRange  int      `json:"accuracyOutOfRange"`
	PercentileInvalid   int      `json:"percentileInvalid"`
	MatchedAnimals      int      `json:"matchedAnimals"`   // filled by the service
	UnmatchedAnimals    int      `json:"unmatchedAnimals"` // filled by the service
	AmbiguousAnimals    int      `json:"ambiguousAnimals"` // filled by the service
	Errors              []string `json:"errors"`           // blocking problems
	Warnings            []string `json:"warnings"`
}

// OK reports whether the import may proceed.
func (r *Report) OK() bool { return len(r.Errors) == 0 }

// NormalizeKey is how registrations are compared: trimmed, case-insensitive,
// internal whitespace collapsed.
func NormalizeKey(s string) string { return strings.ToUpper(strings.Join(strings.Fields(s), " ")) }

// Extract reads records from the table according to the manifest and trait
// definitions, producing the report. It never mutates the table.
func Extract(m *Manifest, defs []TraitDef, t *Table) ([]Record, *Report) {
	rep := &Report{Rows: len(t.Rows), Errors: []string{}, Warnings: []string{}, UnrecognizedColumns: []string{}}
	idCol := t.ColumnIndex(m.Data.IDColumn)
	if idCol < 0 {
		rep.Errors = append(rep.Errors, fmt.Sprintf("A coluna de identificação %q não existe no arquivo.", m.Data.IDColumn))
		return nil, rep
	}
	opt := func(name string) int {
		if strings.TrimSpace(name) == "" {
			return -1
		}
		return t.ColumnIndex(name)
	}
	nameCol, sexCol, breedCol := opt(m.Data.NameColumn), opt(m.Data.SexColumn), opt(m.Data.BreedColumn)
	for _, c := range []struct {
		label, name string
		idx         int
	}{{"nome", m.Data.NameColumn, nameCol}, {"sexo", m.Data.SexColumn, sexCol}, {"raça", m.Data.BreedColumn, breedCol}} {
		if c.name != "" && c.idx < 0 {
			rep.Errors = append(rep.Errors, fmt.Sprintf("A coluna de %s %q não existe no arquivo.", c.label, c.name))
		}
	}

	type cols struct{ val, acc, pct int }
	tcols := make([]cols, len(defs))
	used := map[int]bool{idCol: true, nameCol: true, sexCol: true, breedCol: true}
	for i, d := range defs {
		c := cols{val: t.ColumnIndex(d.Column), acc: opt(d.AccuracyColumn), pct: opt(d.PercentileColumn)}
		if c.val < 0 {
			rep.Errors = append(rep.Errors, fmt.Sprintf("A coluna %q (característica %s) não existe no arquivo.", d.Column, d.Code))
			continue
		}
		if d.AccuracyColumn != "" && c.acc < 0 {
			rep.Errors = append(rep.Errors, fmt.Sprintf("A coluna de acurácia %q (característica %s) não existe no arquivo.", d.AccuracyColumn, d.Code))
		}
		if d.PercentileColumn != "" && c.pct < 0 {
			rep.Errors = append(rep.Errors, fmt.Sprintf("A coluna de percentil %q (característica %s) não existe no arquivo.", d.PercentileColumn, d.Code))
		}
		tcols[i] = c
		rep.TraitsRecognized++
		used[c.val], used[c.acc], used[c.pct] = true, true, true
	}
	for i, name := range t.Columns {
		if !used[i] {
			rep.UnrecognizedColumns = append(rep.UnrecognizedColumns, name)
		}
	}
	rep.TraitsUnrecognized = len(rep.UnrecognizedColumns)
	if len(rep.Errors) > 0 {
		return nil, rep
	}

	seen := map[string]bool{}
	var out []Record
	for ri, row := range t.Rows {
		key := NormalizeKey(row[idCol])
		if key == "" {
			rep.RowsWithoutID++
			continue
		}
		if seen[key] {
			rep.Duplicates++
			continue
		}
		seen[key] = true
		rec := Record{Row: ri + 1, Registration: strings.Join(strings.Fields(row[idCol]), " ")}
		if nameCol >= 0 {
			rec.Name = row[nameCol]
		}
		if sexCol >= 0 {
			rec.Sex = normalizeSex(row[sexCol])
		}
		if rec.Sex == "" {
			rec.Sex = m.Data.DefaultSex
		}
		if breedCol >= 0 {
			rec.Breed = row[breedCol]
		}
		for i, c := range tcols {
			v, ok, missing := ParseNumber(row[c.val], m.Data.DecimalComma)
			if missing {
				continue
			}
			if !ok {
				rep.InvalidValues++
				continue
			}
			val := Value{TraitIndex: i, Value: v}
			rep.Values++
			if c.acc >= 0 {
				if a, ok, missing := ParseNumber(row[c.acc], m.Data.DecimalComma); ok {
					if m.Data.AccuracyScale == "percent" {
						a /= 100
					}
					if a < 0 || a > 1 {
						rep.AccuracyOutOfRange++
					} else {
						val.Accuracy = &a
					}
				} else if !missing {
					rep.AccuracyOutOfRange++
				}
			}
			if val.Accuracy == nil {
				rep.WithoutAccuracy++
			}
			if c.pct >= 0 {
				if p, ok, missing := ParseNumber(row[c.pct], m.Data.DecimalComma); ok {
					if p < 0 || p > 100 {
						rep.PercentileInvalid++
					} else {
						val.Percentile = &p
					}
				} else if !missing {
					rep.PercentileInvalid++
				}
			}
			rec.Values = append(rec.Values, val)
		}
		out = append(out, rec)
	}
	rep.Records = len(out)

	if rep.Records == 0 {
		rep.Errors = append(rep.Errors, "Nenhum registro válido encontrado no arquivo.")
	}
	if rep.Values == 0 && rep.Records > 0 {
		rep.Errors = append(rep.Errors, "Nenhum valor numérico foi lido para as características mapeadas.")
	}
	if rep.RowsWithoutID > 0 {
		rep.Warnings = append(rep.Warnings, fmt.Sprintf("%d linhas sem identificação serão ignoradas.", rep.RowsWithoutID))
	}
	if rep.Duplicates > 0 {
		rep.Warnings = append(rep.Warnings, fmt.Sprintf("%d linhas repetem uma identificação já lida; somente a primeira é usada.", rep.Duplicates))
	}
	if rep.InvalidValues > 0 {
		rep.Warnings = append(rep.Warnings, fmt.Sprintf("%d valores não numéricos serão ignorados.", rep.InvalidValues))
	}
	if rep.WithoutAccuracy > 0 {
		rep.Warnings = append(rep.Warnings, fmt.Sprintf("%d valores sem acurácia (ela será exibida como indisponível).", rep.WithoutAccuracy))
	}
	if rep.AccuracyOutOfRange > 0 {
		rep.Warnings = append(rep.Warnings, fmt.Sprintf("%d acurácias fora do intervalo esperado foram descartadas. Verifique a escala (fração ou %%).", rep.AccuracyOutOfRange))
	}
	return out, rep
}

func normalizeSex(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "m", "male", "macho", "masculino", "1":
		return "male"
	case "f", "female", "fêmea", "femea", "feminino", "2":
		return "female"
	}
	return ""
}
