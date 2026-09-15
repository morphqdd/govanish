package conc

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// TimeAfter reports time.After inside a select, whose timer stays alive
// until it fires even when another case wins.
func TimeAfter() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "timeafter",
		Doc:      "time.After in a select leaks its timer",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runTimeAfter,
	}
}

func runTimeAfter(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.SelectStmt)(nil)}, func(node ast.Node) {
		stmt, ok := node.(*ast.SelectStmt)
		if !ok {
			return
		}

		reportTimeAfter(pass, stmt)
	})

	return nil, nil
}

func reportTimeAfter(pass *analysis.Pass, stmt *ast.SelectStmt) {
	ast.Inspect(stmt, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || !isTimeAfter(call) {
			return true
		}

		pass.Reportf(call.Pos(),
			"time.After in a select leaks its timer; use time.NewTimer")

		return true
	})
}

func isTimeAfter(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "After" {
		return false
	}

	pkg, ok := selector.X.(*ast.Ident)

	return ok && pkg.Name == "time"
}
