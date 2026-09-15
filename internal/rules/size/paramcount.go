package size

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"github.com/morphqdd/govanish/internal/astutil"
)

// maxParams is the most parameters a function may accept. A function that
// wants more wants a struct.
const maxParams = 4

// ParamCount reports functions with unreadably long parameter lists.
func ParamCount() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "paramcount",
		Doc:      fmt.Sprintf("functions may not take more than %d parameters", maxParams),
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runParamCount,
	}
}

func runParamCount(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		decl, ok := node.(*ast.FuncDecl)
		if !ok {
			return
		}

		count := fieldCount(decl.Type.Params)
		if count <= maxParams {
			return
		}

		pass.Reportf(decl.Pos(), "%s has %d parameters, limit is %d",
			astutil.Describe(decl), count, maxParams)
	})

	return nil, nil
}

// fieldCount counts declared names rather than field groups, so that
// (a, b, c int) counts as three.
func fieldCount(list *ast.FieldList) int {
	if list == nil {
		return 0
	}

	count := 0

	for _, field := range list.List {
		if len(field.Names) == 0 {
			count++

			continue
		}

		count += len(field.Names)
	}

	return count
}
