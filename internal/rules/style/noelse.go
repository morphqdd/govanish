package style

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// NoElse reports an else branch whose if branch already ends in a return,
// which is an early return written the long way round.
func NoElse() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "noelse",
		Doc:      "an if branch ending in return must not be followed by else",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runNoElse,
	}
}

func runNoElse(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.IfStmt)(nil)}, func(node ast.Node) {
		stmt, ok := node.(*ast.IfStmt)
		if !ok || stmt.Else == nil {
			return
		}

		if !endsInTerminator(stmt.Body) {
			return
		}

		pass.Reportf(stmt.Else.Pos(),
			"else is unnecessary because the if branch ends in return")
	})

	return nil, nil
}

// endsInTerminator reports whether a block's last statement leaves the
// function or the enclosing loop, making a following else redundant.
func endsInTerminator(block *ast.BlockStmt) bool {
	if len(block.List) == 0 {
		return false
	}

	switch last := block.List[len(block.List)-1].(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.BranchStmt:
		return last.Tok != token.GOTO
	default:
		return false
	}
}
