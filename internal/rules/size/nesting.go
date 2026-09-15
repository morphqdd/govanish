package size

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// maxNesting is the deepest a block may sit inside a function body.
const maxNesting = 3

// Nesting reports blocks buried too deep to follow.
func Nesting() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "nesting",
		Doc:      fmt.Sprintf("blocks may not nest deeper than %d levels", maxNesting),
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runNesting,
	}
}

func runNesting(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		decl, ok := node.(*ast.FuncDecl)
		if !ok || decl.Body == nil {
			return
		}

		walkNesting(pass, decl.Body, 0)
	})

	return nil, nil
}

// walkNesting descends a statement tree, reporting the first statement at
// each path that exceeds the limit. Deeper statements below a reported one
// are not reported again: one diagnostic per offending path is enough.
func walkNesting(pass *analysis.Pass, node ast.Node, depth int) {
	for _, child := range nestedStatements(node) {
		next := depth + 1
		if next > maxNesting {
			pass.Reportf(child.Pos(), "block is nested %d levels deep, limit is %d",
				next, maxNesting)

			continue
		}

		walkNesting(pass, child, next)
	}
}

// nestedStatements returns the statements directly inside node that open a
// new level of nesting.
func nestedStatements(node ast.Node) []ast.Stmt {
	var found []ast.Stmt

	for _, stmt := range directStatements(node) {
		switch stmt.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt,
			*ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.SelectStmt:
			found = append(found, stmt)
		}
	}

	return found
}

// directStatements returns the statements one level inside node, without
// descending into nested blocks.
func directStatements(node ast.Node) []ast.Stmt {
	switch typed := node.(type) {
	case *ast.BlockStmt:
		return typed.List
	case *ast.IfStmt:
		return ifStatements(typed)
	case *ast.ForStmt:
		return typed.Body.List
	case *ast.RangeStmt:
		return typed.Body.List
	case *ast.SwitchStmt:
		return clauseStatements(typed.Body)
	case *ast.TypeSwitchStmt:
		return clauseStatements(typed.Body)
	case *ast.SelectStmt:
		return clauseStatements(typed.Body)
	default:
		return nil
	}
}

// ifStatements flattens an if/else-if/else chain. An else-if sits at the
// same depth as the if it continues, not one level deeper.
func ifStatements(stmt *ast.IfStmt) []ast.Stmt {
	found := append([]ast.Stmt{}, stmt.Body.List...)

	switch typed := stmt.Else.(type) {
	case *ast.BlockStmt:
		found = append(found, typed.List...)
	case *ast.IfStmt:
		found = append(found, ifStatements(typed)...)
	}

	return found
}

func clauseStatements(body *ast.BlockStmt) []ast.Stmt {
	var found []ast.Stmt

	for _, clause := range body.List {
		switch typed := clause.(type) {
		case *ast.CaseClause:
			found = append(found, typed.Body...)
		case *ast.CommClause:
			found = append(found, typed.Body...)
		}
	}

	return found
}
