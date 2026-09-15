package size

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// maxResults is the most values a function may return, error included.
const maxResults = 3

// ReturnCount reports functions returning more values than a caller can
// keep straight.
func ReturnCount() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "returncount",
		Doc:      fmt.Sprintf("functions may not return more than %d values", maxResults),
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runReturnCount,
	}
}

func runReturnCount(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		decl, ok := node.(*ast.FuncDecl)
		if !ok {
			return
		}

		count := fieldCount(decl.Type.Results)
		if count <= maxResults {
			return
		}

		pass.Reportf(decl.Pos(), "%s returns %d values, limit is %d",
			astutil.Describe(decl), count, maxResults)
	})

	return nil, nil
}
