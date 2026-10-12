package datasets

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Limits protect against hostile or accidental huge inputs.
const (
	MaxRows    = 500_000
	MaxColumns = 500
)

// Table is a rectangular text table: Rows[i][j] belongs to Columns[j].
type Table struct {
	Columns   []string
	Rows      [][]string
	Delimiter rune // CSV only
}

// ColumnIndex returns the index of a column (exact match, then case-insensitive), or -1.
func (t *Table) ColumnIndex(name string) int {
	name = strings.TrimSpace(name)
	for i, c := range t.Columns {
		if c == name {
			return i
		}
	}
	for i, c := range t.Columns {
		if strings.EqualFold(c, name) {
			return i
		}
	}
	return -1
}

// decodeText returns UTF-8 text. Invalid UTF-8 is read as Latin-1 (what
// spreadsheet exports commonly produce); a UTF-8 BOM is stripped.
func decodeText(b []byte) string {
	b = bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
	if utf8.Valid(b) {
		return string(b)
	}
	r := make([]rune, len(b))
	for i, c := range b {
		r[i] = rune(c)
	}
	return string(r)
}

func detectDelimiter(firstLine string) rune {
	best, bestN := ',', -1
	for _, d := range []rune{',', ';', '\t'} {
		if n := strings.Count(firstLine, string(d)); n > bestN {
			best, bestN = d, n
		}
	}
	return best
}

// ParseTable reads CSV or a JSON array of flat objects.
func ParseTable(data []byte, format string) (*Table, error) {
	switch format {
	case "csv":
		return parseCSV(data)
	case "json":
		return parseJSON(data)
	default:
		return nil, fmt.Errorf("formato desconhecido: %q", format)
	}
}

func parseCSV(data []byte) (*Table, error) {
	text := decodeText(data)
	first, _, _ := strings.Cut(text, "\n")
	delim := detectDelimiter(first)
	r := csv.NewReader(strings.NewReader(text))
	r.Comma = delim
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true
	header, err := r.Read()
	if err != nil {
		return nil, errors.New("arquivo CSV vazio ou sem cabeçalho")
	}
	if len(header) > MaxColumns {
		return nil, fmt.Errorf("colunas demais (máximo %d)", MaxColumns)
	}
	t := &Table{Delimiter: delim}
	seen := map[string]bool{}
	for i, h := range header {
		h = strings.TrimSpace(h)
		if h == "" {
			h = fmt.Sprintf("coluna_%d", i+1)
		}
		if seen[strings.ToLower(h)] {
			return nil, fmt.Errorf("coluna duplicada no cabeçalho: %q", h)
		}
		seen[strings.ToLower(h)] = true
		t.Columns = append(t.Columns, h)
	}
	for {
		rec, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("CSV inválido: %w", err)
		}
		if len(t.Rows) >= MaxRows {
			return nil, fmt.Errorf("linhas demais (máximo %d)", MaxRows)
		}
		row := make([]string, len(t.Columns))
		empty := true
		for i := range row {
			if i < len(rec) {
				row[i] = strings.TrimSpace(rec[i])
				if row[i] != "" {
					empty = false
				}
			}
		}
		if !empty {
			t.Rows = append(t.Rows, row)
		}
	}
	return t, nil
}

func parseJSON(data []byte) (*Table, error) {
	dec := json.NewDecoder(bytes.NewReader(bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})))
	dec.UseNumber()
	var objs []map[string]any
	if err := dec.Decode(&objs); err != nil {
		return nil, fmt.Errorf("JSON inválido (esperado um array de objetos): %w", err)
	}
	if len(objs) > MaxRows {
		return nil, fmt.Errorf("linhas demais (máximo %d)", MaxRows)
	}
	colSet := map[string]bool{}
	for _, o := range objs {
		for k := range o {
			colSet[k] = true
		}
	}
	if len(colSet) > MaxColumns {
		return nil, fmt.Errorf("colunas demais (máximo %d)", MaxColumns)
	}
	t := &Table{}
	for k := range colSet {
		t.Columns = append(t.Columns, k)
	}
	sort.Strings(t.Columns)
	for _, o := range objs {
		row := make([]string, len(t.Columns))
		for i, c := range t.Columns {
			switch v := o[c].(type) {
			case nil:
			case string:
				row[i] = strings.TrimSpace(v)
			case json.Number:
				row[i] = v.String()
			case bool:
				row[i] = strconv.FormatBool(v)
			default:
				return nil, fmt.Errorf("valor não escalar na coluna %q", c)
			}
		}
		t.Rows = append(t.Rows, row)
	}
	return t, nil
}

// CSV re-serializes the table as UTF-8 comma-separated text.
func (t *Table) CSV() []byte {
	var b bytes.Buffer
	w := csv.NewWriter(&b)
	_ = w.Write(t.Columns)
	for _, r := range t.Rows {
		_ = w.Write(r)
	}
	w.Flush()
	return b.Bytes()
}

var missingTokens = map[string]bool{"": true, "na": true, "n/a": true, "nan": true, "null": true, "-": true, "--": true, ".": true}

// ParseNumber parses a numeric cell. ok=false with missing=true means "no
// value" (blank, NA, ...); ok=false with missing=false means malformed text.
func ParseNumber(s string, decimalComma bool) (v float64, ok, missing bool) {
	s = strings.TrimSpace(s)
	if missingTokens[strings.ToLower(s)] {
		return 0, false, true
	}
	if decimalComma {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.Replace(s, ",", ".", 1)
	} else {
		s = strings.ReplaceAll(s, ",", "")
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || f != f || f > 1e300 || f < -1e300 {
		return 0, false, false
	}
	return f, true, false
}
