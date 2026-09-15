package api

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"github.com/morphqdd/govanish/internal/astutil"
)

// NoIfaceReturn reports an exported function returning an interface. The
// caller should be given the concrete thing and decide for itself what
// abstraction to hold it behind.
func NoIfaceReturn() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "noifacereturn",
		Doc:      "exported functions must return concrete types, not interfaces",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runNoIfaceReturn,
	}
}

func runNoIfaceReturn(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		decl, ok := node.(*ast.FuncDecl)
		if !ok || !decl.Name.IsExported() || decl.Type.Results == nil {
			return
		}

		reportInterfaceResults(pass, decl)
	})

	return nil, nil
}

func reportInterfaceResults(pass *analysis.Pass, decl *ast.FuncDecl) {
	for _, field := range decl.Type.Results.List {
		typ := pass.TypesInfo.TypeOf(field.Type)
		if !isInterface(typ) || astutil.IsError(typ) {
			continue
		}

		pass.Reportf(decl.Pos(), "exported function %s must not return an interface",
			decl.Name.Name)
	}
}

func isInterface(typ types.Type) bool {
	if typ == nil {
		return false
	}

	_, ok := typ.Underlying().(*types.Interface)

	return ok
}
