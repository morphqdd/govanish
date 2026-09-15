package errs

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// ErrCheck reports error results that are neither examined nor returned.
func ErrCheck() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "errcheck",
		Doc:      "every returned error must be examined",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runErrCheck,
	}
}

func runErrCheck(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	nodes := []ast.Node{(*ast.ExprStmt)(nil), (*ast.AssignStmt)(nil)}

	insp.Preorder(nodes, func(node ast.Node) {
		switch typed := node.(type) {
		case *ast.ExprStmt:
			checkExprStmt(pass, typed)
		case *ast.AssignStmt:
			checkAssign(pass, typed)
		}
	})

	return nil, nil
}

// checkExprStmt reports a call whose error result is thrown away by not
// being assigned to anything at all.
func checkExprStmt(pass *analysis.Pass, stmt *ast.ExprStmt) {
	call, ok := stmt.X.(*ast.CallExpr)
	if !ok || !returnsError(pass, call) {
		return
	}

	pass.Reportf(call.Pos(), "error result of %s is unchecked", callName(call))
}

// checkAssign reports an error assigned to the blank identifier, which is
// a deliberate way of not handling it.
func checkAssign(pass *analysis.Pass, stmt *ast.AssignStmt) {
	if len(stmt.Rhs) != 1 {
		return
	}

	call, ok := stmt.Rhs[0].(*ast.CallExpr)
	if !ok {
		return
	}

	results := astutil.ResultTypes(pass.TypesInfo, call)

	for index, target := range stmt.Lhs {
		if index >= len(results) || !astutil.IsError(results[index]) {
			continue
		}

		if ident, isIdent := target.(*ast.Ident); !isIdent || ident.Name != "_" {
			continue
		}

		pass.Reportf(target.Pos(), "error result of %s is discarded", callName(call))
	}
}

func returnsError(pass *analysis.Pass, call *ast.CallExpr) bool {
	for _, result := range astutil.ResultTypes(pass.TypesInfo, call) {
		if astutil.IsError(result) {
			return true
		}
	}

	return false
}

// callName is the function name a diagnostic should use for a call.
func callName(call *ast.CallExpr) string {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		return fun.Name
	case *ast.SelectorExpr:
		return fun.Sel.Name
	default:
		return "call"
	}
}
