package api

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"github.com/morphqdd/govanish/internal/astutil"
)

// NoBarePrim reports bare primitives in an exported signature, where a
// domain type would say what the value means.
func NoBarePrim() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "nobareprim",
		Doc:      "exported signatures must use domain types, not bare primitives",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runNoBarePrim,
	}
}

// barePrimitives are the types that carry no meaning on their own. bool
// is absent because no-bool-param already bans it as a parameter, and a
// boolean result needs no name to be understood.
func barePrimitives() map[types.BasicKind]string {
	return map[types.BasicKind]string{
		types.String:  "string",
		types.Int:     "int",
		types.Int64:   "int64",
		types.Float64: "float64",
	}
}

func runNoBarePrim(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		decl, ok := node.(*ast.FuncDecl)
		if !ok || !decl.Name.IsExported() {
			return
		}

		reportBareParams(pass, decl)
		reportBareResults(pass, decl)
	})

	return nil, nil
}

func reportBareParams(pass *analysis.Pass, decl *ast.FuncDecl) {
	for _, field := range decl.Type.Params.List {
		name, bare := bareName(pass, field.Type)
		if !bare {
			continue
		}

		for _, param := range field.Names {
			pass.Reportf(param.Pos(), "parameter %s is a bare %s; define a domain type",
				param.Name, name)
		}
	}
}

func reportBareResults(pass *analysis.Pass, decl *ast.FuncDecl) {
	if decl.Type.Results == nil {
		return
	}

	for _, field := range decl.Type.Results.List {
		name, bare := bareName(pass, field.Type)
		if !bare {
			continue
		}

		pass.Reportf(field.Pos(), "result of %s is a bare %s; define a domain type",
			decl.Name.Name, name)
	}
}

// bareName reports whether the expression names a predeclared primitive
// directly, rather than a type defined in terms of one.
func bareName(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	typ := pass.TypesInfo.TypeOf(expr)

	basic, ok := typ.(*types.Basic)
	if !ok {
		return "", false
	}

	name, bare := barePrimitives()[basic.Kind()]

	return name, bare
}
