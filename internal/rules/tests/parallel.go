package tests

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Parallel reports tests that neither run in parallel nor hand their
// *testing.T to a helper that might.
var Parallel = &analysis.Analyzer{
	Name: "testparallel",
	Doc:  "tests must call t.Parallel",
	Run:  runParallel,
}

func runParallel(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if !isTestFile(pass.Fset.Position(file.Pos()).Filename) {
			continue
		}

		for _, decl := range file.Decls {
			checkParallel(pass, decl)
		}
	}

	return nil, nil
}

func checkParallel(pass *analysis.Pass, decl ast.Decl) {
	fn, ok := decl.(*ast.FuncDecl)
	if !ok {
		return
	}

	receiver, isTest := testParam(fn)
	if !isTest || fn.Body == nil {
		return
	}

	if calls(fn.Body, receiver, "Parallel") || delegates(fn.Body, receiver) {
		return
	}

	pass.Reportf(fn.Pos(), "test %s must call t.Parallel", fn.Name.Name)
}

// delegates reports whether the body passes its *testing.T to another
// function, which is then responsible for calling t.Parallel.
func delegates(body *ast.BlockStmt, receiver string) bool {
	found := false

	ast.Inspect(body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return !found
		}

		for _, arg := range call.Args {
			ident, isIdent := arg.(*ast.Ident)
			if isIdent && ident.Name == receiver {
				found = true
			}
		}

		return !found
	})

	return found
}
