package arch

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"govanish/internal/astutil"
)

// NoShadow reports a variable that hides another of the same name from an
// enclosing scope, which makes the reader track which one is meant.
func NoShadow() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "noshadow",
		Doc:      "a declaration may not shadow a variable from an enclosing scope",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runNoShadow,
	}
}

func runNoShadow(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	parameters := literalParams(insp)

	insp.Preorder([]ast.Node{(*ast.Ident)(nil)}, func(node ast.Node) {
		ident, ok := node.(*ast.Ident)
		if !ok || ident.Name == "_" || parameters[ident] {
			return
		}

		variable, ok := pass.TypesInfo.Defs[ident].(*types.Var)
		if !ok || !shadowsOuter(pass, variable) {
			return
		}

		pass.Reportf(ident.Pos(), "declaration of %s shadows an outer variable",
			ident.Name)
	})

	return nil, nil
}

// literalParams collects the parameter and result names of function
// literals. A closure's signature is a new function's, not a shadowing
// declaration, and naming a parameter after the variable it replaces is
// how Go avoids capturing the outer one by accident.
func literalParams(insp *inspector.Inspector) map[*ast.Ident]bool {
	names := make(map[*ast.Ident]bool)

	insp.Preorder([]ast.Node{(*ast.FuncLit)(nil)}, func(node ast.Node) {
		literal, ok := node.(*ast.FuncLit)
		if !ok {
			return
		}

		for _, field := range fieldsOf(literal.Type) {
			for _, name := range field.Names {
				names[name] = true
			}
		}
	})

	return names
}

func fieldsOf(signature *ast.FuncType) []*ast.Field {
	fields := signature.Params.List
	if signature.Results != nil {
		fields = append(append([]*ast.Field{}, fields...), signature.Results.List...)
	}

	return fields
}

// shadowsOuter reports whether the variable hides one declared in an
// enclosing function scope. Package scope is excluded: a local named
// after a package-level identifier is ordinary, and shadowing across that
// boundary is what no-globals already prevents.
func shadowsOuter(pass *analysis.Pass, variable *types.Var) bool {
	scope := variable.Parent()
	if scope == nil {
		return false
	}

	for outer := scope.Parent(); outer != nil; outer = outer.Parent() {
		if outer == pass.Pkg.Scope() || outer == types.Universe {
			return false
		}

		if outer.Lookup(variable.Name()) != nil {
			return true
		}
	}

	return false
}
