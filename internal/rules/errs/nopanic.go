package errs

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// NoPanic reports code that ends the process instead of returning an
// error. Only main may decide that the program is over.
func NoPanic() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "nopanic",
		Doc:      "panic and process exit are banned outside package main",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runNoPanic,
	}
}

// fatalCalls() are the package-qualified calls that end the process.
func fatalCalls() map[string]map[string]bool {
	return map[string]map[string]bool{
		"os":  {"Exit": true},
		"log": {"Fatal": true, "Fatalf": true, "Fatalln": true, "Panic": true},
	}
}

func runNoPanic(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() == "main" {
		return nil, nil
	}

	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(node ast.Node) {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return
		}

		reportFatal(pass, call)
	})

	return nil, nil
}

func reportFatal(pass *analysis.Pass, call *ast.CallExpr) {
	if ident, isIdent := call.Fun.(*ast.Ident); isIdent && ident.Name == "panic" {
		pass.Reportf(call.Pos(), "panic is banned; return an error instead")

		return
	}

	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	pkg, ok := selector.X.(*ast.Ident)
	if !ok || !fatalCalls()[pkg.Name][selector.Sel.Name] {
		return
	}

	pass.Reportf(call.Pos(), "%s.%s is banned outside main; return an error instead",
		pkg.Name, selector.Sel.Name)
}
