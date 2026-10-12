package datasets

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"genetica/internal/domain"
)

// ---- builders for synthetic packages

type entry struct {
	name string
	data []byte
	mode os.FileMode
}

func buildZip(t *testing.T, entries ...entry) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		h := &zip.FileHeader{Name: e.name, Method: zip.Deflate}
		if e.mode != 0 {
			h.SetMode(e.mode)
		}
		w, err := zw.CreateHeader(h)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(e.data); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func manifestJSON(t *testing.T, mod func(*Manifest)) []byte {
	t.Helper()
	m := Manifest{
		SchemaVersion: 1, ID: "example.dataset", Name: "Example Dataset", Version: "1.0.0", Species: "Bos taurus",
		Breeds: []string{"example"}, Purpose: []string{"beef"}, Organization: "Org", SourceURL: "https://example.org",
		License:    License{Name: "CC-BY-4.0", URL: "https://creativecommons.org/licenses/by/4.0/", RedistributionAllowed: true},
		Evaluation: Evaluation{Program: "SYNTHETIC", GeneticBase: "synthetic-base"},
		Data:       DataSpec{Format: "csv", IDColumn: "registration", NameColumn: "name", AccuracyScale: "fraction"},
	}
	if mod != nil {
		mod(&m)
	}
	b, _ := json.Marshal(m)
	return b
}

const traitsOK = `[
 {"code":"WW","name":"Peso à desmama","category":"growth","unit":"kg","evaluationType":"DEP","column":"DEP_PD","accuracyColumn":"ACC_PD","percentileColumn":"PCT_PD"},
 {"code":"YW","name":"Peso ao sobreano","category":"growth","unit":"kg","evaluationType":"DEP","column":"DEP_PS"}
]`

const dataOK = "registration,name,DEP_PD,ACC_PD,PCT_PD,DEP_PS\n" +
	"R1,Bull One,5.2,0.81,70,8.1\n" +
	"R2,Bull Two,3.0,0.50,40,\n" +
	"R3,Bull Three,NA,,,1.0\n"

func validPackage(t *testing.T, extra ...entry) []byte {
	t.Helper()
	es := []entry{
		{name: "manifest.json", data: manifestJSON(t, nil)},
		{name: "traits.json", data: []byte(traitsOK)},
		{name: "data.csv", data: []byte(dataOK)},
		{name: "README.md", data: []byte("# readme")},
		{name: "LICENSE", data: []byte("CC-BY-4.0 text")},
	}
	return buildZip(t, append(es, extra...)...)
}

func wantReason(t *testing.T, err error, substr string) {
	t.Helper()
	var pe *PackageError
	if !errors.As(err, &pe) || !strings.Contains(pe.Reason, substr) {
		t.Fatalf("want package error containing %q, got %v", substr, err)
	}
}

// ---- happy path

func TestReadValidPackage(t *testing.T) {
	p, err := ReadPackage(validPackage(t), DefaultLimits)
	if err != nil {
		t.Fatal(err)
	}
	if p.Manifest.ID != "example.dataset" || len(p.Traits) != 2 || len(p.Table.Rows) != 3 || p.Readme != "# readme" || p.License == "" {
		t.Fatalf("%+v", p.Manifest)
	}
	if p.Manifest.Evaluation.Release != "1.0.0" {
		t.Fatalf("release should default to the version, got %q", p.Manifest.Evaluation.Release)
	}
	if len(p.SHA256) != 64 {
		t.Fatal("sha256 missing")
	}
}

func TestWriteThenReadRoundTrip(t *testing.T) {
	p, err := ReadPackage(validPackage(t), DefaultLimits)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := WritePackage(p)
	if err != nil {
		t.Fatal(err)
	}
	p2, err := ReadPackage(raw, DefaultLimits)
	if err != nil {
		t.Fatal(err)
	}
	if p2.Manifest.ID != p.Manifest.ID || len(p2.Table.Rows) != len(p.Table.Rows) || len(p2.Traits) != 2 {
		t.Fatal("round trip lost data")
	}
}

// ---- hostile ZIPs

func TestRejectsPathTraversalAndOddNames(t *testing.T) {
	for _, name := range []string{"../evil.txt", "../../etc/passwd", "/etc/passwd", "sub/../manifest.json", "..\\evil", "dir/data.csv", "C:evil", "manifest.json\x00.txt"} {
		_, err := ReadPackage(validPackage(t, entry{name: name, data: []byte("x")}), DefaultLimits)
		if err == nil {
			t.Fatalf("name %q must be rejected", name)
		}
		var pe *PackageError
		if !errors.As(err, &pe) {
			t.Fatalf("name %q: wrong error type %v", name, err)
		}
	}
}

func TestRejectsUnexpectedFile(t *testing.T) {
	_, err := ReadPackage(validPackage(t, entry{name: "evil.sh", data: []byte("rm -rf /")}), DefaultLimits)
	wantReason(t, err, "inesperado")
}

func TestRejectsSymlinkAndDirectory(t *testing.T) {
	_, err := ReadPackage(validPackage(t, entry{name: "LICENSE.txt", data: []byte("/etc/passwd"), mode: os.ModeSymlink | 0o777}), DefaultLimits)
	wantReason(t, err, "regular")
	_, err = ReadPackage(validPackage(t, entry{name: "dir/", data: nil}), DefaultLimits)
	if err == nil {
		t.Fatal("directory entry must be rejected")
	}
}

func TestRejectsDuplicateEntries(t *testing.T) {
	_, err := ReadPackage(validPackage(t, entry{name: "manifest.json", data: manifestJSON(t, nil)}), DefaultLimits)
	wantReason(t, err, "duplicado")
}

func TestRejectsZipBomb(t *testing.T) {
	lim := DefaultLimits
	lim.MaxDataBytes = 1 << 20
	huge := bytes.Repeat([]byte("0"), 8<<20) // compresses to a few KB
	es := []entry{
		{name: "manifest.json", data: manifestJSON(t, nil)},
		{name: "traits.json", data: []byte(traitsOK)},
		{name: "data.csv", data: huge},
	}
	raw := buildZip(t, es...)
	if len(raw) > 100<<10 {
		t.Fatalf("test setup: bomb should be tiny, is %d bytes", len(raw))
	}
	_, err := ReadPackage(raw, lim)
	wantReason(t, err, "grande demais")

	// total-size and file-count limits
	lim2 := DefaultLimits
	lim2.MaxFiles = 3
	_, err = ReadPackage(validPackage(t), lim2)
	wantReason(t, err, "arquivos demais")

	lim3 := DefaultLimits
	lim3.MaxPackageBytes = 100
	_, err = ReadPackage(validPackage(t), lim3)
	wantReason(t, err, "grande demais")
}

func TestRejectsNotAZip(t *testing.T) {
	_, err := ReadPackage([]byte("this is not a zip"), DefaultLimits)
	wantReason(t, err, "ZIP")
}

// ---- structure and manifest validation

func TestRejectsMissingPieces(t *testing.T) {
	_, err := ReadPackage(buildZip(t, entry{name: "traits.json", data: []byte(traitsOK)}, entry{name: "data.csv", data: []byte(dataOK)}), DefaultLimits)
	wantReason(t, err, "manifest.json")
	_, err = ReadPackage(buildZip(t, entry{name: "manifest.json", data: manifestJSON(t, nil)}, entry{name: "data.csv", data: []byte(dataOK)}), DefaultLimits)
	wantReason(t, err, "traits.json")
	_, err = ReadPackage(buildZip(t, entry{name: "manifest.json", data: manifestJSON(t, nil)}, entry{name: "traits.json", data: []byte(traitsOK)}), DefaultLimits)
	wantReason(t, err, "exatamente um")
	_, err = ReadPackage(validPackage(t, entry{name: "data.json", data: []byte("[]")}), DefaultLimits)
	wantReason(t, err, "exatamente um")
}

func TestRejectsInvalidJSON(t *testing.T) {
	_, err := ReadPackage(buildZip(t, entry{name: "manifest.json", data: []byte("{")}, entry{name: "traits.json", data: []byte(traitsOK)}, entry{name: "data.csv", data: []byte(dataOK)}), DefaultLimits)
	wantReason(t, err, "manifest.json inválido")
	_, err = ReadPackage(buildZip(t, entry{name: "manifest.json", data: manifestJSON(t, nil)}, entry{name: "traits.json", data: []byte("{}")}, entry{name: "data.csv", data: []byte(dataOK)}), DefaultLimits)
	wantReason(t, err, "traits.json inválido")
}

func TestManifestValidation(t *testing.T) {
	cases := []struct {
		name  string
		mod   func(*Manifest)
		field string
	}{
		{"schema incompatível", func(m *Manifest) { m.SchemaVersion = 2 }, "manifest.schemaVersion"},
		{"id inválido", func(m *Manifest) { m.ID = "../x" }, "manifest.id"},
		{"id vazio", func(m *Manifest) { m.ID = "" }, "manifest.id"},
		{"nome vazio", func(m *Manifest) { m.Name = " " }, "manifest.name"},
		{"versão com barra", func(m *Manifest) { m.Version = "1/2" }, "manifest.version"},
		{"sem espécie", func(m *Manifest) { m.Species = "" }, "manifest.species"},
		{"sem raças", func(m *Manifest) { m.Breeds = nil }, "manifest.breeds"},
		{"finalidade inválida", func(m *Manifest) { m.Purpose = []string{"pets"} }, "manifest.purpose"},
		{"data inválida", func(m *Manifest) { m.ReleaseDate = "ontem" }, "manifest.releaseDate"},
		{"url inválida", func(m *Manifest) { m.SourceURL = "javascript:alert(1)" }, "manifest.sourceUrl"},
		{"url de licença inválida", func(m *Manifest) { m.License.URL = "file:///etc/passwd" }, "manifest.license.url"},
		{"sem programa", func(m *Manifest) { m.Evaluation.Program = "" }, "manifest.evaluation.program"},
		{"sem coluna id", func(m *Manifest) { m.Data.IDColumn = "" }, "manifest.data.idColumn"},
		{"escala acurácia", func(m *Manifest) { m.Data.AccuracyScale = "x" }, "manifest.data.accuracyScale"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ReadPackage(buildZip(t,
				entry{name: "manifest.json", data: manifestJSON(t, c.mod)},
				entry{name: "traits.json", data: []byte(traitsOK)},
				entry{name: "data.csv", data: []byte(dataOK)}), DefaultLimits)
			var ve *domain.ValidationError
			if !errors.As(err, &ve) || ve.Fields[c.field] == "" {
				t.Fatalf("want error on %s, got %v", c.field, err)
			}
		})
	}
}

func TestManifestFormatMustMatchDataFile(t *testing.T) {
	_, err := ReadPackage(buildZip(t,
		entry{name: "manifest.json", data: manifestJSON(t, func(m *Manifest) { m.Data.Format = "json" })},
		entry{name: "traits.json", data: []byte(traitsOK)},
		entry{name: "data.csv", data: []byte(dataOK)}), DefaultLimits)
	wantReason(t, err, "formato")
}

func TestTraitDefValidation(t *testing.T) {
	bad := func(traits string) error {
		_, err := ReadPackage(buildZip(t,
			entry{name: "manifest.json", data: manifestJSON(t, nil)},
			entry{name: "traits.json", data: []byte(traits)},
			entry{name: "data.csv", data: []byte(dataOK)}), DefaultLimits)
		return err
	}
	if bad(`[]`) == nil {
		t.Fatal("empty trait list must be rejected")
	}
	if bad(`[{"code":"WW","name":"x","evaluationType":"PHENOTYPE","column":"DEP_PD"}]`) == nil {
		t.Fatal("datasets cannot provide phenotypes")
	}
	if bad(`[{"code":"W W","name":"x","evaluationType":"DEP","column":"DEP_PD"}]`) == nil {
		t.Fatal("invalid code")
	}
	if bad(`[{"code":"A","name":"x","evaluationType":"DEP","column":"DEP_PD"},{"code":"A","name":"y","evaluationType":"DEP","column":"DEP_PS"}]`) == nil {
		t.Fatal("duplicate code+type")
	}
	if bad(`[{"code":"A","name":"x","evaluationType":"DEP","column":"DEP_PD"},{"code":"B","name":"y","evaluationType":"DEP","column":"dep_pd"}]`) == nil {
		t.Fatal("column reused")
	}
	if bad(`[{"code":"A","name":"x","evaluationType":"DEP","column":"DEP_PD","optimizationDirection":"TARGET_RANGE"}]`) == nil {
		t.Fatal("TARGET_RANGE without bounds")
	}
	// same code with different types is fine
	if err := bad(`[{"code":"A","name":"x","evaluationType":"DEP","column":"DEP_PD"},{"code":"A","name":"y","evaluationType":"EBV","column":"DEP_PS"}]`); err != nil {
		t.Fatal(err)
	}
}

// ---- table parsing

func TestParseCSVVariants(t *testing.T) {
	semi, err := ParseTable([]byte("\xEF\xBB\xBFreg;nome;valor\nR1;José;1,5\n;;\nR2;Ana;2,5\n"), "csv")
	if err != nil {
		t.Fatal(err)
	}
	if semi.Delimiter != ';' || len(semi.Rows) != 2 || semi.Columns[0] != "reg" {
		t.Fatalf("%+v", semi)
	}
	// Latin-1 bytes (é = 0xE9)
	lat, err := ParseTable([]byte("reg,nome\nR1,Jos\xE9\n"), "csv")
	if err != nil || lat.Rows[0][1] != "José" {
		t.Fatalf("latin-1: %+v %v", lat, err)
	}
	if _, err := ParseTable([]byte("a,a\n1,2\n"), "csv"); err == nil {
		t.Fatal("duplicate header must be rejected")
	}
	if _, err := ParseTable([]byte(""), "csv"); err == nil {
		t.Fatal("empty csv")
	}
	if _, err := ParseTable([]byte("a,b\n1,2\n"), "xml"); err == nil {
		t.Fatal("unknown format")
	}
}

func TestParseJSON(t *testing.T) {
	tb, err := ParseTable([]byte(`[{"reg":"R1","v":1.5,"ok":true},{"reg":"R2","v":null}]`), "json")
	if err != nil || len(tb.Rows) != 2 {
		t.Fatalf("%+v %v", tb, err)
	}
	if _, err := ParseTable([]byte(`{"a":1}`), "json"); err == nil {
		t.Fatal("object instead of array")
	}
	if _, err := ParseTable([]byte(`[{"a":{"nested":1}}]`), "json"); err == nil {
		t.Fatal("nested value")
	}
}

func TestParseNumber(t *testing.T) {
	cases := []struct {
		in          string
		comma       bool
		want        float64
		ok, missing bool
	}{
		{"1.5", false, 1.5, true, false},
		{"1,5", true, 1.5, true, false},
		{"1.234,5", true, 1234.5, true, false},
		{"1,234.5", false, 1234.5, true, false},
		{"-3", false, -3, true, false},
		{"", false, 0, false, true},
		{"NA", false, 0, false, true},
		{"nan", false, 0, false, true},
		{"abc", false, 0, false, false},
		{"1e999", false, 0, false, false},
	}
	for _, c := range cases {
		v, ok, missing := ParseNumber(c.in, c.comma)
		if ok != c.ok || missing != c.missing || (ok && v != c.want) {
			t.Fatalf("%q: got %v %v %v", c.in, v, ok, missing)
		}
	}
}

// ---- extraction and validation report

func TestExtractReport(t *testing.T) {
	p, err := ReadPackage(validPackage(t), DefaultLimits)
	if err != nil {
		t.Fatal(err)
	}
	recs, rep := Extract(&p.Manifest, p.Traits, p.Table)
	if !rep.OK() {
		t.Fatalf("%+v", rep)
	}
	if rep.Records != 3 || rep.TraitsRecognized != 2 || rep.Values != 4 {
		t.Fatalf("%+v", rep)
	}
	// values: R1 has WW(acc) + YW(no acc col); R2 has WW(acc); R3 has only YW(no acc col)
	// => accuracy missing for: R1.YW, R3.YW = 2
	if rep.WithoutAccuracy != 2 {
		t.Fatalf("withoutAccuracy = %d, want 2 (YW has no accuracy column)", rep.WithoutAccuracy)
	}
	if len(rep.UnrecognizedColumns) != 0 {
		t.Fatalf("%v", rep.UnrecognizedColumns)
	}
	r1 := recs[0]
	if r1.Registration != "R1" || r1.Sex != "male" || len(r1.Values) != 2 {
		t.Fatalf("%+v", r1)
	}
	if r1.Values[0].Accuracy == nil || *r1.Values[0].Accuracy != 0.81 || r1.Values[0].Percentile == nil || *r1.Values[0].Percentile != 70 {
		t.Fatalf("%+v", r1.Values[0])
	}
	if r1.Values[1].Accuracy != nil {
		t.Fatal("no accuracy column => nil accuracy, not a made-up number")
	}
}

func TestExtractCountsProblems(t *testing.T) {
	m := Manifest{Data: DataSpec{IDColumn: "reg", DefaultSex: "male", AccuracyScale: "percent"}}
	defs := []TraitDef{{Code: "WW", Column: "ww", AccuracyColumn: "acc"}, {Code: "YW", Column: "yw"}}
	tb := &Table{Columns: []string{"reg", "ww", "acc", "yw", "extra1", "extra2"}, Rows: [][]string{
		{"A", "1", "81", "2", "x", "y"},
		{"a ", "9", "9", "9", "", ""},     // duplicate (case/whitespace-insensitive)
		{"", "1", "1", "1", "", ""},       // no id
		{"B", "oops", "150", "3", "", ""}, // invalid value; accuracy out of range (percent > 100) but ww invalid so not read
		{"C", "4", "150", "", "", ""},     // accuracy out of range
		{"D", "5", "", "6", "", ""},       // missing accuracy
	}}
	recs, rep := Extract(&m, defs, tb)
	if rep.Records != 4 || rep.Duplicates != 1 || rep.RowsWithoutID != 1 || rep.InvalidValues != 1 || rep.AccuracyOutOfRange != 1 {
		t.Fatalf("%+v", rep)
	}
	if rep.TraitsUnrecognized != 2 || strings.Join(rep.UnrecognizedColumns, ",") != "extra1,extra2" {
		t.Fatalf("unrecognized: %+v", rep.UnrecognizedColumns)
	}
	if recs[0].Values[0].Accuracy == nil || *recs[0].Values[0].Accuracy != 0.81 {
		t.Fatalf("percent scale must convert to a fraction: %+v", recs[0].Values[0])
	}
	if len(rep.Warnings) == 0 {
		t.Fatal("expected warnings")
	}
}

func TestExtractErrors(t *testing.T) {
	m := Manifest{Data: DataSpec{IDColumn: "nope", DefaultSex: "male"}}
	_, rep := Extract(&m, nil, &Table{Columns: []string{"a"}})
	if rep.OK() {
		t.Fatal("missing id column must block")
	}
	m.Data.IDColumn = "a"
	_, rep = Extract(&m, []TraitDef{{Code: "X", Column: "zzz"}}, &Table{Columns: []string{"a"}, Rows: [][]string{{"1"}}})
	if rep.OK() {
		t.Fatal("missing trait column must block")
	}
	_, rep = Extract(&m, []TraitDef{{Code: "X", Column: "b"}}, &Table{Columns: []string{"a", "b"}, Rows: [][]string{{"1", "NA"}}})
	if rep.OK() {
		t.Fatal("no numeric values at all must block")
	}
}

// ---- license policy

func TestLicensePolicy(t *testing.T) {
	m := Manifest{}
	if m.LicenseStatus() != LicenseUnknown || m.CheckBundleable() == nil {
		t.Fatal("empty license is unknown and not bundleable")
	}
	m.License = License{Name: "Proprietary", RedistributionAllowed: false}
	if m.LicenseStatus() != LicenseNoRedistribute || !errors.Is(m.CheckBundleable(), ErrNotBundleable) {
		t.Fatal("known but non-redistributable must not be bundleable")
	}
	m.License = License{Name: "CC-BY-4.0", RedistributionAllowed: true}
	if m.CheckBundleable() == nil {
		t.Fatal("a license without URL must not be bundleable")
	}
	m.License.URL = "https://creativecommons.org/licenses/by/4.0/"
	if m.LicenseStatus() != LicenseRedistributable || m.CheckBundleable() != nil {
		t.Fatal("explicit license + url + redistribution => bundleable")
	}
	// "free to download" is represented by nothing: a free download has no license fields at all
	free := Manifest{SourceURL: "https://example.org/free-download"}
	if free.CheckBundleable() == nil {
		t.Fatal("a public source URL alone never makes a dataset redistributable")
	}
}
