package arch

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// NoGlobals reports package-level variables. A constant is fine: it
// cannot be changed from under a reader.
func NoGlobals() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "noglobals",
		Doc:  "package-level variables are banned; pass state explicitly",
		Run:  runNoGlobals,
	}
}

func runNoGlobals(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			reportGlobals(pass, decl)
		}
	}

	return nil, nil
}

func reportGlobals(pass *analysis.Pass, decl ast.Decl) {
	gen, ok := decl.(*ast.GenDecl)
	if !ok || gen.Tok != token.VAR {
		return
	}

	for _, spec := range gen.Specs {
		value, isValue := spec.(*ast.ValueSpec)
		if !isValue {
			continue
		}

		reportNames(pass, value)
	}
}

func reportNames(pass *analysis.Pass, spec *ast.ValueSpec) {
	for _, name := range spec.Names {
		if name.Name == "_" {
			continue
		}

		pass.Reportf(name.Pos(),
			"package-level var %s is banned; pass state explicitly", name.Name)
	}
}
