package errs

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"github.com/morphqdd/govanish/internal/astutil"
)

// ErrShape reports signatures that put the error anywhere but last, or
// name it anything but err.
func ErrShape() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "errshape",
		Doc:      "the error result must come last and be named err",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runErrShape,
	}
}

func runErrShape(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		decl, ok := node.(*ast.FuncDecl)
		if !ok || decl.Type.Results == nil {
			return
		}

		checkResults(pass, decl)
	})

	return nil, nil
}

func checkResults(pass *analysis.Pass, decl *ast.FuncDecl) {
	fields := decl.Type.Results.List

	for index, field := range fields {
		if !astutil.IsError(pass.TypesInfo.TypeOf(field.Type)) {
			continue
		}

		if index != len(fields)-1 {
			pass.Reportf(decl.Pos(), "error must be the last result")
		}

		reportErrorName(pass, decl, field)
	}
}

func reportErrorName(pass *analysis.Pass, decl *ast.FuncDecl, field *ast.Field) {
	for _, name := range field.Names {
		if name.Name == "err" || name.Name == "_" {
			continue
		}

		pass.Reportf(decl.Pos(), "named error result must be called err, not %s",
			name.Name)
	}
}
