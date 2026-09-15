package arch

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// NoAny reports the empty interface in an exported signature or field,
// where it tells a caller nothing about the value it holds. Unexported
// code may still need it to satisfy an interface it does not own.
func NoAny() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "noany",
		Doc:      "the empty interface is banned from the exported surface",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runNoAny,
	}
}

func runNoAny(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	nodes := []ast.Node{(*ast.FuncDecl)(nil), (*ast.TypeSpec)(nil)}

	insp.Preorder(nodes, func(node ast.Node) {
		switch typed := node.(type) {
		case *ast.FuncDecl:
			checkSignature(pass, typed)
		case *ast.TypeSpec:
			checkStructFields(pass, typed)
		}
	})

	return nil, nil
}

func checkSignature(pass *analysis.Pass, decl *ast.FuncDecl) {
	if !decl.Name.IsExported() {
		return
	}

	reportFields(pass, decl.Type.Params)
	reportFields(pass, decl.Type.Results)
}

func checkStructFields(pass *analysis.Pass, spec *ast.TypeSpec) {
	structType, ok := spec.Type.(*ast.StructType)
	if !ok || !spec.Name.IsExported() {
		return
	}

	reportFields(pass, structType.Fields)
}

func reportFields(pass *analysis.Pass, fields *ast.FieldList) {
	if fields == nil {
		return
	}

	for _, field := range fields.List {
		if !isEmptyInterface(pass, field.Type) {
			continue
		}

		pass.Reportf(field.Type.Pos(), "any is banned; name the type you mean")
	}
}

func isEmptyInterface(pass *analysis.Pass, expr ast.Expr) bool {
	switch typed := expr.(type) {
	case *ast.InterfaceType:
		return len(typed.Methods.List) == 0
	case *ast.Ident:
		return typed.Name == "any" && pass.TypesInfo.Uses[typed] != nil
	default:
		return false
	}
}
