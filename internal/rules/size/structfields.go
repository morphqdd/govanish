package size

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"github.com/morphqdd/govanish/internal/astutil"
)

// maxStructFields is the most fields a struct may declare.
const maxStructFields = 10

// StructFields reports structs that have accumulated too many fields to
// have a single responsibility.
func StructFields() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "structfields",
		Doc:      fmt.Sprintf("structs may not declare more than %d fields", maxStructFields),
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runStructFields,
	}
}

func runStructFields(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.TypeSpec)(nil)}, func(node ast.Node) {
		spec, ok := node.(*ast.TypeSpec)
		if !ok {
			return
		}

		structType, ok := spec.Type.(*ast.StructType)
		if !ok {
			return
		}

		count := fieldCount(structType.Fields)
		if count <= maxStructFields {
			return
		}

		pass.Reportf(spec.Pos(), "struct %s has %d fields, limit is %d",
			spec.Name.Name, count, maxStructFields)
	})

	return nil, nil
}
