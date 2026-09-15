package tests

import (
	"strings"

	"golang.org/x/tools/go/analysis"
)

// PkgDoc reports a package with no package comment anywhere in it. An
// external test package is exempt: it documents nothing a consumer can
// import.
func PkgDoc() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "pkgdoc",
		Doc:  "every package must have a package comment",
		Run:  runPkgDoc,
	}
}

func runPkgDoc(pass *analysis.Pass) (any, error) {
	if len(pass.Files) == 0 || strings.HasSuffix(pass.Pkg.Name(), "_test") {
		return nil, nil
	}

	for _, file := range pass.Files {
		if file.Doc != nil {
			return nil, nil
		}
	}

	first := pass.Files[0]
	pass.Reportf(first.Package, "package %s has no doc comment", pass.Pkg.Name())

	return nil, nil
}
