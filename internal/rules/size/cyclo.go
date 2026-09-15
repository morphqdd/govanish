package size

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"github.com/morphqdd/govanish/internal/astutil"
)

// maxComplexity is the most independent paths a function may contain.
const maxComplexity = 8

// Cyclo reports functions with too many branches to test exhaustively.
func Cyclo() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "cyclo",
		Doc:      fmt.Sprintf("functions may not exceed a cyclomatic complexity of %d", maxComplexity),
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runCyclo,
	}
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
func decisions(node ast.Node) int {
	switch typed := node.(type) {
	case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
		return 1
	case *ast.CaseClause:
		return caseDecision(typed)
	case *ast.CommClause:
		return commDecision(typed)
	case *ast.BinaryExpr:
		return logicalDecision(typed)
	default:
		return 0
	}
}

// caseDecision counts a case clause. A default clause introduces no new
// path: it is the path that remains.
func caseDecision(clause *ast.CaseClause) int {
	if len(clause.List) == 0 {
		return 0
	}

	return 1
}

func commDecision(clause *ast.CommClause) int {
	if clause.Comm == nil {
		return 0
	}

	return 1
}

func logicalDecision(expr *ast.BinaryExpr) int {
	if expr.Op != token.LAND && expr.Op != token.LOR {
		return 0
	}

	return 1
}
