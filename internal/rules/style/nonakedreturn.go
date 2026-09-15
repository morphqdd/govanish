package style

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// NoNakedReturn reports bare returns in functions with named results.
// Function literals are skipped: their results are checked against their
// own signature, not the enclosing one.
var NoNakedReturn = &analysis.Analyzer{
	Name:     "nonakedreturn",
	Doc:      "functions with named results must return their values explicitly",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runNoNakedReturn,
}

func runNoNakedReturn(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		decl, ok := node.(*ast.FuncDecl)
		if !ok || decl.Body == nil || !hasNamedResults(decl.Type) {
			return
		}

		reportNakedReturns(pass, decl.Body)
	})

	return nil, nil
}

func hasNamedResults(signature *ast.FuncType) bool {
	if signature.Results == nil {
		return false
	}

	for _, field := range signature.Results.List {
		if len(field.Names) > 0 {
			return true
		}
	}

	return false
}

func reportNakedReturns(pass *analysis.Pass, body *ast.BlockStmt) {
	ast.Inspect(body, func(node ast.Node) bool {
		if _, nested := node.(*ast.FuncLit); nested {
			return false
		}

		stmt, ok := node.(*ast.ReturnStmt)
		if !ok || len(stmt.Results) > 0 {
			return true
		}

		pass.Reportf(stmt.Pos(), "naked return hides which values are returned")

		return true
	})
}
