package errs

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// EmptyBranch reports an error check whose body does nothing, which reads
// as handling and behaves as ignoring.
func EmptyBranch() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "emptyerrbranch",
		Doc:      "an error branch may not be empty",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runEmptyBranch,
	}
}

func runEmptyBranch(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.IfStmt)(nil)}, func(node ast.Node) {
		stmt, ok := node.(*ast.IfStmt)
		if !ok || len(stmt.Body.List) > 0 || !comparesErrorToNil(pass, stmt.Cond) {
			return
		}

		pass.Reportf(stmt.Pos(),
			"error branch is empty, so the error is silently dropped")
	})

	return nil, nil
}

func comparesErrorToNil(pass *analysis.Pass, cond ast.Expr) bool {
	binary, ok := cond.(*ast.BinaryExpr)
	if !ok {
		return false
	}

	return astutil.IsError(pass.TypesInfo.TypeOf(binary.X)) && isNil(binary.Y)
}

func isNil(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)

	return ok && ident.Name == "nil"
}
