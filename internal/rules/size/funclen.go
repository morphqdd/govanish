package size

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// maxFuncLines is the most lines a function body may contain, counting
// neither of the braces that delimit it.
const maxFuncLines = 40

// FuncLen reports functions whose body is too long to read at once.
func FuncLen() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "funclen",
		Doc:      fmt.Sprintf("function bodies may not exceed %d lines", maxFuncLines),
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runFuncLen,
	}
}

func runFuncLen(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		decl, ok := node.(*ast.FuncDecl)
		if !ok || decl.Body == nil {
			return
		}

		lines := bodyLines(pass, decl.Body)
		if lines <= maxFuncLines {
			return
		}

		pass.Reportf(decl.Pos(), "%s is %d lines, limit is %d",
			astutil.Describe(decl), lines, maxFuncLines)
	})

	return nil, nil
}

func bodyLines(pass *analysis.Pass, body *ast.BlockStmt) int {
	first := pass.Fset.Position(body.Lbrace).Line
	last := pass.Fset.Position(body.Rbrace).Line

	return last - first - 1
}
