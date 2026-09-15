package conc

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"github.com/morphqdd/govanish/internal/astutil"
)

// DeferUnlock reports a Lock that is not immediately followed by a
// deferred Unlock, which is the only arrangement that survives an early
// return or a panic.
func DeferUnlock() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "deferunlock",
		Doc:      "a Lock must be followed immediately by defer Unlock",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runDeferUnlock,
	}
}

func runDeferUnlock(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.BlockStmt)(nil)}, func(node ast.Node) {
		block, ok := node.(*ast.BlockStmt)
		if !ok {
			return
		}

		checkLockOrder(pass, block)
	})

	return nil, nil
}

func checkLockOrder(pass *analysis.Pass, block *ast.BlockStmt) {
	for index, stmt := range block.List {
		receiver, ok := lockReceiver(stmt)
		if !ok {
			continue
		}

		if index+1 < len(block.List) && unlocksNext(block.List[index+1], receiver) {
			continue
		}

		pass.Reportf(stmt.Pos(), "Lock must be followed immediately by defer Unlock")
	}
}

// lockReceiver returns the textual receiver of a Lock call, so that
// c.mu.Lock and c.mu.Unlock can be matched against each other.
func lockReceiver(stmt ast.Stmt) (string, bool) {
	expr, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return "", false
	}

	return methodReceiver(expr.X, "Lock")
}

func unlocksNext(stmt ast.Stmt, receiver string) bool {
	deferred, ok := stmt.(*ast.DeferStmt)
	if !ok {
		return false
	}

	got, ok := methodReceiver(deferred.Call, "Unlock")

	return ok && got == receiver
}

func methodReceiver(expr ast.Expr, method string) (string, bool) {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return "", false
	}

	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != method {
		return "", false
	}

	return exprText(selector.X), true
}

// exprText renders the expression a method was called on, well enough to
// tell one mutex from another within a single block.
func exprText(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.SelectorExpr:
		return exprText(typed.X) + "." + typed.Sel.Name
	default:
		return ""
	}
}
