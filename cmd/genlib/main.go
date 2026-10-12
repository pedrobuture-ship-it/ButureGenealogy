// Command genlib inspects dataset packages.
//
//	genlib verify [--official] file.genlib...
//
// verify reads each package with the same safety checks the application uses
// and prints its license status. With --official it exits non-zero unless the
// license is identified and explicitly allows redistribution, which is the
// requirement for shipping a dataset with the application.
package main

import (
	"fmt"
	"os"

	"genetica/internal/datasets"
)

func main() {
	args := os.Args[1:]
	if len(args) < 2 || args[0] != "verify" {
		fmt.Fprintln(os.Stderr, "uso: genlib verify [--official] arquivo.genlib...")
		os.Exit(2)
	}
	official := false
	files := args[1:]
	if files[0] == "--official" {
		official, files = true, files[1:]
	}
	failed := false
	for _, f := range files {
		p, err := datasets.ReadPackageFile(f, datasets.DefaultLimits)
		if err != nil {
			fmt.Printf("FALHA  %s: %v\n", f, err)
			failed = true
			continue
		}
		m := p.Manifest
		fmt.Printf("OK     %s: %s@%s licença=%s (%s)\n", f, m.ID, m.Version, m.LicenseStatus(), m.License.Name)
		if official {
			if err := m.CheckBundleable(); err != nil {
				fmt.Printf("RECUSADO %s: %v\n", f, err)
				failed = true
			}
		}
	}
	if failed {
		os.Exit(1)
	}
}
