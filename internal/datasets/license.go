package datasets

import "errors"

// LicenseStatus is the user-facing summary of a license. It is computed, never
// assumed: "free to download" is not a license.
type LicenseStatus string

const (
	LicenseUnknown         LicenseStatus = "unknown"         // no license identified
	LicenseNoRedistribute  LicenseStatus = "no_redistribute" // known, redistribution not permitted
	LicenseRedistributable LicenseStatus = "redistributable" // known and redistribution explicitly permitted
)

// Status classifies the license of a manifest.
func (m *Manifest) LicenseStatus() LicenseStatus {
	switch {
	case !m.License.Known():
		return LicenseUnknown
	case m.License.RedistributionAllowed:
		return LicenseRedistributable
	default:
		return LicenseNoRedistribute
	}
}

// ResponsibilityNotice must be shown whenever a user installs data by hand.
const ResponsibilityNotice = "A responsabilidade pela autorização de uso e distribuição de dados importados manualmente é do usuário."

// ErrNotBundleable explains why a dataset cannot ship with the application.
var ErrNotBundleable = errors.New("dataset não pode ser distribuído com o aplicativo")

// CheckBundleable is the gate for the official distribution: the license must
// be identified (name and URL) and must explicitly allow redistribution.
// Anything else, including an unknown license, is refused.
func (m *Manifest) CheckBundleable() error {
	if !m.License.Known() {
		return errors.Join(ErrNotBundleable, errors.New("licença desconhecida"))
	}
	if m.License.URL == "" {
		return errors.Join(ErrNotBundleable, errors.New("licença sem URL de referência"))
	}
	if !m.License.RedistributionAllowed {
		return errors.Join(ErrNotBundleable, errors.New("a licença não permite redistribuição"))
	}
	return nil
}
