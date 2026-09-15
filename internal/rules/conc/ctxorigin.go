package conc

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"github.com/morphqdd/govanish/internal/astutil"
)

// CtxOrigin reports a context created where one should have been passed
// in. Only the top of the program knows what the root context should be.
func CtxOrigin() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "ctxorigin",
		Doc:      "context.Background and context.TODO belong in main or a test",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runCtxOrigin,
	}
}

func runCtxOrigin(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() == "main" {
		return nil, nil
	}

	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(node ast.Node) {
		call, ok := node.(*ast.CallExpr)
		if !ok || isTestPosition(pass, call) {
			return
		}

		name, ok := contextRoot(call)
		if !ok {
			return
		}

		pass.Reportf(call.Pos(), "context.%s may only be called in main or a test", name)
	})

	return nil, nil
}

func contextRoot(call *ast.CallExpr) (string, bool) {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}

	pkg, ok := selector.X.(*ast.Ident)
	if !ok || pkg.Name != "context" {
		return "", false
	}

	if selector.Sel.Name != "Background" && selector.Sel.Name != "TODO" {
		return "", false
	}

	return selector.Sel.Name, true
}

func isTestPosition(pass *analysis.Pass, node ast.Node) bool {
	name := pass.Fset.Position(node.Pos()).Filename

	return strings.HasSuffix(name, "_test.go")
}
