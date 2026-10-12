package datasets

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Limits bound what a package may contain. They apply to the real number of
// bytes read, never to sizes merely declared in the ZIP headers.
type Limits struct {
	MaxPackageBytes int64 // the .genlib file itself
	MaxDataBytes    int64 // data.csv / data.json after decompression
	MaxTextBytes    int64 // manifest, traits, README, LICENSE
	MaxTotalBytes   int64 // everything after decompression
	MaxFiles        int
}

var DefaultLimits = Limits{
	MaxPackageBytes: 64 << 20,
	MaxDataBytes:    128 << 20,
	MaxTextBytes:    1 << 20,
	MaxTotalBytes:   160 << 20,
	MaxFiles:        10,
}

// allowedFiles is the whole set of names a package may contain (flat, no directories).
var allowedFiles = map[string]bool{
	"manifest.json": true, "traits.json": true, "data.csv": true, "data.json": true,
	"README.md": true, "LICENSE": true, "LICENSE.txt": true,
}

// Package is a fully parsed dataset package.
type Package struct {
	Manifest Manifest
	Traits   []TraitDef
	Table    *Table
	Readme   string
	License  string // license text
	SHA256   string // of the package bytes
	Raw      []byte // original package bytes (nil until WritePackage/ReadPackage)
}

// PackageError is a rejected package; Reason is safe to show to the user.
type PackageError struct{ Reason string }

func (e *PackageError) Error() string { return "pacote inválido: " + e.Reason }

func reject(format string, a ...any) error { return &PackageError{fmt.Sprintf(format, a...)} }

// ReadPackageFile reads a .genlib from disk, refusing oversized files before loading them.
func ReadPackageFile(path string, lim Limits) (*Package, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() {
		return nil, reject("não é um arquivo regular")
	}
	if st.Size() > lim.MaxPackageBytes {
		return nil, reject("arquivo grande demais (%d MB, máximo %d MB)", st.Size()>>20, lim.MaxPackageBytes>>20)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ReadPackage(b, lim)
}

func safeName(name string) error {
	switch {
	case name == "":
		return reject("entrada com nome vazio")
	case strings.ContainsAny(name, "\\\x00"):
		return reject("nome de arquivo suspeito: %q", name)
	case strings.HasPrefix(name, "/"), strings.Contains(name, ".."), strings.Contains(name, ":"):
		return reject("caminho suspeito no pacote: %q", name)
	case strings.Contains(name, "/"):
		return reject("o pacote não pode conter diretórios: %q", name)
	case !allowedFiles[name]:
		return reject("arquivo inesperado no pacote: %q", name)
	}
	return nil
}

// ReadPackage parses and validates a .genlib held in memory.
func ReadPackage(data []byte, lim Limits) (*Package, error) {
	if int64(len(data)) > lim.MaxPackageBytes {
		return nil, reject("arquivo grande demais")
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, reject("não é um arquivo .genlib (ZIP) válido")
	}
	if len(zr.File) > lim.MaxFiles {
		return nil, reject("arquivos demais no pacote (máximo %d)", lim.MaxFiles)
	}

	files := map[string][]byte{}
	var total int64
	for _, f := range zr.File {
		if err := safeName(f.Name); err != nil {
			return nil, err
		}
		if f.Mode()&os.ModeSymlink != 0 || f.Mode().IsDir() || !f.Mode().IsRegular() {
			return nil, reject("entrada que não é arquivo regular: %q", f.Name)
		}
		if _, dup := files[f.Name]; dup {
			return nil, reject("arquivo duplicado no pacote: %q", f.Name)
		}
		limit := lim.MaxTextBytes
		if f.Name == "data.csv" || f.Name == "data.json" {
			limit = lim.MaxDataBytes
		}
		if int64(f.UncompressedSize64) > limit {
			return nil, reject("%q é grande demais após descompactar", f.Name)
		}
		rc, err := f.Open()
		if err != nil {
			return nil, reject("não foi possível ler %q", f.Name)
		}
		b, err := io.ReadAll(io.LimitReader(rc, limit+1))
		rc.Close()
		if err != nil {
			return nil, reject("erro ao descompactar %q", f.Name)
		}
		if int64(len(b)) > limit {
			return nil, reject("%q excede o tamanho máximo permitido", f.Name)
		}
		total += int64(len(b))
		if total > lim.MaxTotalBytes {
			return nil, reject("conteúdo total grande demais")
		}
		files[f.Name] = b
	}

	for _, req := range []string{"manifest.json", "traits.json"} {
		if _, ok := files[req]; !ok {
			return nil, reject("faltando %s", req)
		}
	}
	_, hasCSV := files["data.csv"]
	_, hasJSON := files["data.json"]
	if hasCSV == hasJSON {
		return nil, reject("o pacote deve conter exatamente um arquivo de dados (data.csv ou data.json)")
	}

	p := &Package{Raw: data}
	sum := sha256.Sum256(data)
	p.SHA256 = hex.EncodeToString(sum[:])

	if err := json.Unmarshal(files["manifest.json"], &p.Manifest); err != nil {
		return nil, reject("manifest.json inválido: %v", err)
	}
	p.Manifest.Normalize()
	wantFormat := "csv"
	dataName := "data.csv"
	if hasJSON {
		wantFormat, dataName = "json", "data.json"
	}
	if p.Manifest.Data.Format != wantFormat {
		return nil, reject("manifest declara formato %q mas o arquivo é %s", p.Manifest.Data.Format, dataName)
	}
	if err := p.Manifest.Validate(); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(files["traits.json"], &p.Traits); err != nil {
		return nil, reject("traits.json inválido: %v", err)
	}
	if err := ValidateDefs(p.Traits); err != nil {
		return nil, err
	}
	if p.Table, err = ParseTable(files[dataName], wantFormat); err != nil {
		return nil, reject("%s: %v", dataName, err)
	}
	p.Readme = string(files["README.md"])
	if l, ok := files["LICENSE"]; ok {
		p.License = string(l)
	} else {
		p.License = string(files["LICENSE.txt"])
	}
	return p, nil
}

// WritePackage serializes a package (used for interactive CSV/JSON imports so
// that every installed dataset is stored as a .genlib).
func WritePackage(p *Package) ([]byte, error) {
	if p.Table == nil {
		return nil, errors.New("pacote sem dados")
	}
	m := p.Manifest
	m.Data.Format = "csv"
	mj, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	tj, err := json.MarshalIndent(p.Traits, "", "  ")
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	add := func(name string, content []byte) error {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = w.Write(content)
		return err
	}
	license, readme := p.License, p.Readme
	if strings.TrimSpace(license) == "" {
		license = NoLicenseText
	}
	if strings.TrimSpace(readme) == "" {
		readme = DefaultReadme(m)
	}
	for _, e := range []struct {
		name string
		b    []byte
	}{{"manifest.json", mj}, {"traits.json", tj}, {"data.csv", p.Table.CSV()}, {"README.md", []byte(readme)}, {"LICENSE", []byte(license)}} {
		if err := add(e.name, e.b); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// NoLicenseText is stored when a manually imported file comes without license text.
const NoLicenseText = "Licença não informada pelo usuário que criou este pacote."

// DefaultReadme documents a package created from a manually imported file.
func DefaultReadme(m Manifest) string {
	return "# " + m.Name + "\n\nPacote criado a partir de um arquivo importado manualmente.\n\n" + ResponsibilityNotice + "\n"
}
