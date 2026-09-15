package conc

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// SizedChan reports make(chan T) without a size. Whether a send blocks is
// a design decision, and it should be written down rather than defaulted.
func SizedChan() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "sizedchan",
		Doc:      "channel construction must state a buffer size",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runSizedChan,
	}
}

func runSizedChan(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(node ast.Node) {
		call, ok := node.(*ast.CallExpr)
		if !ok || !isUnsizedMake(call) {
			return
		}

		pass.Reportf(call.Pos(), "make(chan) must state a buffer size")
	})

	return nil, nil
}

func isUnsizedMake(call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok || ident.Name != "make" || len(call.Args) != 1 {
		return false
	}

	_, isChan := call.Args[0].(*ast.ChanType)

	return isChan
}
