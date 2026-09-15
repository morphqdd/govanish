package conc

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// contextTypeName is the type every rule in this file is about.
const contextTypeName = "context.Context"

// CtxFirst reports contexts that are not the first parameter, are named
// something other than ctx, or are stored in a struct.
func CtxFirst() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "ctxfirst",
		Doc:      "context.Context must be the first parameter, named ctx, and never stored",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runCtxFirst,
	}
}

func runCtxFirst(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil), (*ast.StructType)(nil)}, func(node ast.Node) {
		switch typed := node.(type) {
		case *ast.FuncDecl:
			checkParams(pass, typed)
		case *ast.StructType:
			checkFields(pass, typed)
		}
	})

	return nil, nil
}

func checkParams(pass *analysis.Pass, decl *ast.FuncDecl) {
	for index, field := range decl.Type.Params.List {
		if !isContext(pass, field.Type) {
			continue
		}

		if index != 0 {
			pass.Reportf(field.Pos(), "context.Context must be the first parameter")
		}

		reportContextName(pass, field)
	}
}

func reportContextName(pass *analysis.Pass, field *ast.Field) {
	for _, name := range field.Names {
		if name.Name == "ctx" || name.Name == "_" {
			continue
		}

		pass.Reportf(name.Pos(), "context parameter must be named ctx, not %s", name.Name)
	}
}

func checkFields(pass *analysis.Pass, structType *ast.StructType) {
	for _, field := range structType.Fields.List {
		if !isContext(pass, field.Type) {
			continue
		}

		pass.Reportf(field.Pos(), "context.Context must not be stored in a struct field")
	}
}

func isContext(pass *analysis.Pass, expr ast.Expr) bool {
	typ := pass.TypesInfo.TypeOf(expr)
	if typ == nil {
		return false
	}

	named, ok := typ.(*types.Named)

	return ok && named.String() == contextTypeName
}
