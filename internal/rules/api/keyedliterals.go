package api

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// KeyedLiterals reports struct literals written by field order, which
// break silently when a field is inserted.
func KeyedLiterals() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "keyedliterals",
		Doc:      "struct literals must name their fields",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runKeyedLiterals,
	}
}

func runKeyedLiterals(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(node ast.Node) {
		literal, ok := node.(*ast.CompositeLit)
		if !ok || len(literal.Elts) == 0 {
			return
		}

		typ := pass.TypesInfo.TypeOf(literal)
		if !isStruct(typ) || hasKeys(literal) {
			return
		}

		pass.Reportf(literal.Pos(), "composite literal of %s must name its fields",
			typeName(typ))
	})

	return nil, nil
}

func isStruct(typ types.Type) bool {
	if typ == nil {
		return false
	}

	_, ok := typ.Underlying().(*types.Struct)

	return ok
}

func hasKeys(literal *ast.CompositeLit) bool {
	_, keyed := literal.Elts[0].(*ast.KeyValueExpr)

	return keyed
}

// typeName is the bare name of a type, without its package path, because
// a diagnostic already names the file it is in.
func typeName(typ types.Type) string {
	named, ok := typ.(*types.Named)
	if !ok {
		return typ.String()
	}

	return named.Obj().Name()
}
