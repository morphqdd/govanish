package tests

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// skipMethods are the testing methods that abandon a test at runtime.
var skipMethods = map[string]bool{
	"Skip":    true,
	"Skipf":   true,
	"SkipNow": true,
}

// NoSkip reports tests that skip themselves. A test that is skipped is a
// test that does not exist, but still looks like coverage.
var NoSkip = &analysis.Analyzer{
	Name:     "noskip",
	Doc:      "tests may not skip themselves; delete the test or fix it",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runNoSkip,
}

func runNoSkip(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(node ast.Node) {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return
		}

		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || !skipMethods[selector.Sel.Name] {
			return
		}

		pass.Reportf(call.Pos(), "t.%s disables a test without removing it",
			selector.Sel.Name)
	})

	return nil, nil
}
