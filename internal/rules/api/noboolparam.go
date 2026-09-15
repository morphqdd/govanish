package api

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"github.com/morphqdd/govanish/internal/astutil"
)

// NoBoolParam reports boolean parameters. At a call site a bare true says
// nothing about what it turns on.
func NoBoolParam() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "noboolparam",
		Doc:      "boolean parameters are banned; use a named type",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runNoBoolParam,
	}
}

func runNoBoolParam(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		decl, ok := node.(*ast.FuncDecl)
		if !ok {
			return
		}

		for _, field := range decl.Type.Params.List {
			reportBoolNames(pass, field)
		}
	})

	return nil, nil
}

func reportBoolNames(pass *analysis.Pass, field *ast.Field) {
	if !isBool(pass.TypesInfo.TypeOf(field.Type)) {
		return
	}

	for _, name := range field.Names {
		pass.Reportf(name.Pos(),
			"parameter %s is a bool; a caller cannot read what true means", name.Name)
	}
}

// isBool reports whether a type is the predeclared bool, rather than a
// named type that happens to be defined as one.
func isBool(typ types.Type) bool {
	basic, ok := typ.(*types.Basic)

	return ok && basic.Kind() == types.Bool
}
