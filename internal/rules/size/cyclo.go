package size

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// maxComplexity is the most independent paths a function may contain.
const maxComplexity = 8

// Cyclo reports functions with too many branches to test exhaustively.
var Cyclo = &analysis.Analyzer{
	Name:     "cyclo",
	Doc:      fmt.Sprintf("functions may not exceed a cyclomatic complexity of %d", maxComplexity),
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runCyclo,
}

func runCyclo(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		decl, ok := node.(*ast.FuncDecl)
		if !ok || decl.Body == nil {
			return
		}

		score := complexity(decl.Body)
		if score <= maxComplexity {
			return
		}

		pass.Reportf(decl.Pos(), "%s has cyclomatic complexity %d, limit is %d",
			astutil.Describe(decl), score, maxComplexity)
	})

	return nil, nil
}

// complexity counts the decision points in a body, plus one for the entry
// path, which is the standard McCabe measure.
func complexity(body *ast.BlockStmt) int {
	score := 1

	ast.Inspect(body, func(node ast.Node) bool {
		score += decisions(node)

		return true
	})

	return score
}

// decisions reports how many independent paths a single node introduces.
// A default clause introduces none: it is the path that remains.
func decisions(node ast.Node) int {
	switch typed := node.(type) {
	case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
		return 1
	case *ast.CaseClause:
		return countIf(len(typed.List) > 0)
	case *ast.CommClause:
		return countIf(typed.Comm != nil)
	case *ast.BinaryExpr:
		return countIf(typed.Op == token.LAND || typed.Op == token.LOR)
	default:
		return 0
	}
}

func countIf(condition bool) int {
	if condition {
		return 1
	}

	return 0
}
