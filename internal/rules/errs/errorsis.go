package errs

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// ErrorsIs reports errors compared with == or !=, which misses any error
// that has been wrapped.
func ErrorsIs() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "errorsis",
		Doc:      "errors must be compared with errors.Is, not == or !=",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runErrorsIs,
	}
}

func runErrorsIs(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.BinaryExpr)(nil)}, func(node ast.Node) {
		binary, ok := node.(*ast.BinaryExpr)
		if !ok || (binary.Op != token.EQL && binary.Op != token.NEQ) {
			return
		}

		if !comparesErrorValues(pass, binary) {
			return
		}

		pass.Reportf(binary.Pos(), "compare errors with errors.Is, not %s", binary.Op)
	})

	return nil, nil
}

// comparesErrorValues reports whether both sides are errors and neither
// is nil. Comparing an error to nil is how Go asks whether it failed.
func comparesErrorValues(pass *analysis.Pass, binary *ast.BinaryExpr) bool {
	if isNil(binary.X) || isNil(binary.Y) {
		return false
	}

	return astutil.IsError(pass.TypesInfo.TypeOf(binary.X)) &&
		astutil.IsError(pass.TypesInfo.TypeOf(binary.Y))
}
