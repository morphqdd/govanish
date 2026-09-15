package arch

import (
	"errors"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// errMissingInspector reports that the driver did not supply the shared
// inspector, which means the analyzer was run without its Requires.
var errMissingInspector = errors.New("inspect analyzer result missing")

// NoInit reports package-level init functions.
var NoInit = &analysis.Analyzer{
	Name:     "noinit",
	Doc:      "func init is banned because it hides work behind package loading",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runNoInit,
}

func runNoInit(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, errMissingInspector
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		decl, ok := node.(*ast.FuncDecl)
		if !ok {
			return
		}

		if decl.Recv != nil || decl.Name.Name != "init" {
			return
		}

		pass.Reportf(decl.Pos(), "func init is banned; do the work in an explicit constructor called from main")
	})

	return nil, nil
}
