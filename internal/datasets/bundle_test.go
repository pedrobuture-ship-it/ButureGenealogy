package datasets

import (
	"path/filepath"
	"testing"
)

// TestBundledDatasetsAreRedistributable guards the official distribution:
// every package in /datasets must have a known license that permits redistribution.
func TestBundledDatasetsAreRedistributable(t *testing.T) {
	files, err := filepath.Glob("../../datasets/*.genlib")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		p, err := ReadPackageFile(f, DefaultLimits)
		if err != nil {
			t.Errorf("%s: %v", f, err)
			continue
		}
		if err := p.Manifest.CheckBundleable(); err != nil {
			t.Errorf("%s não pode ser distribuído com o aplicativo: %v", f, err)
		}
	}
	t.Logf("%d pacote(s) verificado(s)", len(files))
}
