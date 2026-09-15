package conc

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"github.com/morphqdd/govanish/internal/astutil"
)

// GoOwner reports a goroutine nobody waits for. A goroutine without an
// owner outlives the function that started it.
func GoOwner() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "goowner",
		Doc:      "a go statement needs a WaitGroup or errgroup in scope",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runGoOwner,
	}
}

// waiterTypes are the types that give a goroutine an owner able to wait
// for it.
func waiterTypes() []string {
	return []string{"sync.WaitGroup", "golang.org/x/sync/errgroup.Group"}
}

func runGoOwner(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		decl, ok := node.(*ast.FuncDecl)
		if !ok || decl.Body == nil || hasWaiter(pass, decl.Body) {
			return
		}

		reportGoStatements(pass, decl.Body)
	})

	return nil, nil
}

func reportGoStatements(pass *analysis.Pass, body *ast.BlockStmt) {
	ast.Inspect(body, func(node ast.Node) bool {
		stmt, ok := node.(*ast.GoStmt)
		if !ok {
			return true
		}

		pass.Reportf(stmt.Pos(),
			"go statement needs a sync.WaitGroup or errgroup.Group in scope to wait for it")

		return true
	})
}

// hasWaiter reports whether the function declares anything that can wait
// for a goroutine.
func hasWaiter(pass *analysis.Pass, body *ast.BlockStmt) bool {
	found := false

	ast.Inspect(body, func(node ast.Node) bool {
		ident, ok := node.(*ast.Ident)
		if !ok {
			return true
		}

		if isWaiterType(pass.TypesInfo.TypeOf(ident)) {
			found = true
		}

		return !found
	})

	return found
}

func isWaiterType(typ types.Type) bool {
	if typ == nil {
		return false
	}

	name := strings.TrimPrefix(typ.String(), "*")

	for _, waiter := range waiterTypes() {
		if name == waiter {
			return true
		}
	}

	return false
}
