// Package datasets implements the dataset package format (.genlib), its safe
// reading, validation and record extraction. It is pure: no SQL, no UI.
// See docs/DATASET_FORMAT.md.
package datasets

import (
	"net/url"
	"regexp"
	"strings"
	"time"

	"genetica/internal/domain"
)

// SchemaVersion is the only manifest schema this version understands.
const SchemaVersion = 1

// License describes the terms of a dataset. Nothing is ever assumed: an empty
// name means "unknown", and redistribution is false unless explicitly true.
type License struct {
	Name                  string `json:"name"`
	URL                   string `json:"url"`
	RedistributionAllowed bool   `json:"redistributionAllowed"`
	CommercialUseAllowed  *bool  `json:"commercialUseAllowed"` // nil = unknown
}

// Known reports whether the license is identified at all.
func (l License) Known() bool { return strings.TrimSpace(l.Name) != "" }

// Evaluation carries the provenance shared by every value of the dataset.
type Evaluation struct {
	Program     string `json:"program"`
	Release     string `json:"release"` // defaults to the dataset version
	GeneticBase string `json:"geneticBase"`
}

// DataSpec tells how to read the data file.
type DataSpec struct {
	Format        string `json:"format"`        // "csv" | "json"
	IDColumn      string `json:"idColumn"`      // animal registration
	NameColumn    string `json:"nameColumn"`    // optional
	SexColumn     string `json:"sexColumn"`     // optional
	BreedColumn   string `json:"breedColumn"`   // optional
	DefaultSex    string `json:"defaultSex"`    // used when SexColumn is empty ("male" for sire catalogues)
	DecimalComma  bool   `json:"decimalComma"`  // numbers like 1.234,5
	AccuracyScale string `json:"accuracyScale"` // "fraction" (0-1, default) | "percent" (0-100)
}

// Manifest is manifest.json.
type Manifest struct {
	SchemaVersion int        `json:"schemaVersion"`
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Version       string     `json:"version"`
	Species       string     `json:"species"`
	Breeds        []string   `json:"breeds"`
	Purpose       []string   `json:"purpose"`
	Authors       []string   `json:"authors"`
	Organization  string     `json:"organization"`
	SourceURL     string     `json:"sourceUrl"`
	ReleaseDate   string     `json:"releaseDate"`
	License       License    `json:"license"`
	Evaluation    Evaluation `json:"evaluation"`
	Data          DataSpec   `json:"data"`
}

var (
	idRE      = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{1,99}$`)
	versionRE = regexp.MustCompile(`^[0-9A-Za-z][0-9A-Za-z._+-]{0,49}$`)
)

type fieldErrs map[string]string

func (f fieldErrs) add(k, msg string) {
	if _, ok := f[k]; !ok {
		f[k] = msg
	}
}

func (f fieldErrs) err() error {
	if len(f) == 0 {
		return nil
	}
	return &domain.ValidationError{Fields: f}
}

func validHTTPURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// Normalize fills defaults that do not change meaning.
func (m *Manifest) Normalize() {
	if m.Evaluation.Release == "" {
		m.Evaluation.Release = m.Version
	}
	if m.Data.AccuracyScale == "" {
		m.Data.AccuracyScale = "fraction"
	}
	if m.Data.Format == "" {
		m.Data.Format = "csv"
	}
	if m.Data.DefaultSex == "" && m.Data.SexColumn == "" {
		m.Data.DefaultSex = domain.SexMale
	}
}

// Validate checks the manifest. Field names are prefixed with "manifest.".
func (m *Manifest) Validate() error {
	f := fieldErrs{}
	if m.SchemaVersion != SchemaVersion {
		f.add("manifest.schemaVersion", "Versão de esquema incompatível: este aplicativo lê a versão 1.")
	}
	if !idRE.MatchString(m.ID) {
		f.add("manifest.id", "Identificador inválido: use minúsculas, números, ponto, hífen ou _ (2 a 100 caracteres).")
	}
	if strings.TrimSpace(m.Name) == "" || len(m.Name) > 200 {
		f.add("manifest.name", "Nome obrigatório (até 200 caracteres).")
	}
	if !versionRE.MatchString(m.Version) {
		f.add("manifest.version", "Versão inválida.")
	}
	if strings.TrimSpace(m.Species) == "" {
		f.add("manifest.species", "Informe a espécie.")
	}
	if len(m.Breeds) == 0 {
		f.add("manifest.breeds", "Informe ao menos uma raça.")
	}
	for _, b := range m.Breeds {
		if strings.TrimSpace(b) == "" {
			f.add("manifest.breeds", "Raça vazia.")
		}
	}
	if len(m.Purpose) == 0 {
		f.add("manifest.purpose", "Informe a finalidade (beef, dairy ou dual).")
	}
	for _, p := range m.Purpose {
		if p != domain.PurposeBeef && p != domain.PurposeDairy && p != domain.PurposeDual {
			f.add("manifest.purpose", "Finalidade inválida: "+p+".")
		}
	}
	if m.ReleaseDate != "" {
		if _, err := time.Parse(domain.DateLayout, m.ReleaseDate); err != nil {
			f.add("manifest.releaseDate", "Data de lançamento inválida (use AAAA-MM-DD).")
		}
	}
	if m.SourceURL != "" && !validHTTPURL(m.SourceURL) {
		f.add("manifest.sourceUrl", "URL da fonte inválida.")
	}
	if m.License.URL != "" && !validHTTPURL(m.License.URL) {
		f.add("manifest.license.url", "URL da licença inválida.")
	}
	if strings.TrimSpace(m.Evaluation.Program) == "" {
		f.add("manifest.evaluation.program", "Informe o programa de avaliação (necessário para a proveniência).")
	}
	if m.Data.Format != "csv" && m.Data.Format != "json" {
		f.add("manifest.data.format", "Formato de dados deve ser csv ou json.")
	}
	if strings.TrimSpace(m.Data.IDColumn) == "" {
		f.add("manifest.data.idColumn", "Informe a coluna de identificação do animal.")
	}
	if m.Data.AccuracyScale != "fraction" && m.Data.AccuracyScale != "percent" {
		f.add("manifest.data.accuracyScale", "Escala de acurácia deve ser fraction ou percent.")
	}
	if m.Data.SexColumn == "" && m.Data.DefaultSex != domain.SexMale && m.Data.DefaultSex != domain.SexFemale {
		f.add("manifest.data.defaultSex", "Informe a coluna de sexo ou o sexo padrão.")
	}
	return f.err()
}
