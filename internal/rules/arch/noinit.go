package arch

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// NoInit reports package-level init functions.
func NoInit() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "noinit",
		Doc:      "func init is banned because it hides work behind package loading",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runNoInit,
	}
}

func runNoInit(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		decl, ok := node.(*ast.FuncDecl)
		if !ok {
			return
		}

		if decl.Recv != nil || decl.Name.Name != "init" {
			return
		}

		pass.Reportf(decl.Pos(),
			"func init is banned; do the work in an explicit constructor")
	})

	return nil, nil
}
