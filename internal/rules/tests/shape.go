package tests

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Shape reports test functions that do not follow Go's test conventions:
// a capitalized name, a single *testing.T, and subtests for table cases.
func Shape() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "testshape",
		Doc:  "tests must be named TestXxx and use t.Run for table cases",
		Run:  runShape,
	}
}

func runShape(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if !isTestFile(pass.Fset.Position(file.Pos()).Filename) {
			continue
		}

		for _, decl := range file.Decls {
			checkShape(pass, decl)
		}
	}

	return nil, nil
}

// checkShape validates one declaration. A Test-prefixed function with the
// wrong signature is ignored: it does not compile in a test file, so the
// toolchain has already rejected it.
func checkShape(pass *analysis.Pass, decl ast.Decl) {
	fn, ok := decl.(*ast.FuncDecl)
	if !ok || fn.Recv != nil || !strings.HasPrefix(fn.Name.Name, "Test") {
		return
	}

	receiver, isTest := testParam(fn)
	if !isTest {
		return
	}

	if !startsWithCapital(fn.Name.Name) {
		pass.Reportf(fn.Pos(), "test name %s must continue with a capital letter",
			fn.Name.Name)
	}

	checkSubtests(pass, fn, receiver)
}

// checkSubtests reports a test that loops over cases without giving each
// case its own subtest, which makes a failure name the case that failed.
func checkSubtests(pass *analysis.Pass, fn *ast.FuncDecl, receiver string) {
	if !rangesOverCases(pass, fn.Body) || calls(fn.Body, receiver, "Run") {
		return
	}

	pass.Reportf(fn.Pos(), "table test %s must run its cases with t.Run", fn.Name.Name)
}

// rangesOverCases reports whether the body loops over cases written out
// in the test itself. Looping over something a function returned is
// ordinary iteration, not a table, so it needs no subtests.
func rangesOverCases(pass *analysis.Pass, body *ast.BlockStmt) bool {
	literals := literalNames(pass, body)
	found := false

	ast.Inspect(body, func(node ast.Node) bool {
		stmt, ok := node.(*ast.RangeStmt)
		if ok && isCaseTable(pass, stmt.X, literals) {
			found = true
		}

		return !found
	})

	return found
}

func isCaseTable(pass *analysis.Pass, expr ast.Expr, literals map[string]bool) bool {
	if isCaseSlice(pass, expr) {
		return true
	}

	ident, ok := expr.(*ast.Ident)

	return ok && literals[ident.Name]
}

// isCaseSlice reports whether the expression is a slice of structs
// written out in place, which is the shape of a test table. A map or a
// slice of plain values is an expectation, not a table of cases.
func isCaseSlice(pass *analysis.Pass, expr ast.Expr) bool {
	if _, isLiteral := expr.(*ast.CompositeLit); !isLiteral {
		return false
	}

	typ := pass.TypesInfo.TypeOf(expr)
	if typ == nil {
		return false
	}

	slice, ok := typ.Underlying().(*types.Slice)
	if !ok {
		return false
	}

	_, isStruct := slice.Elem().Underlying().(*types.Struct)

	return isStruct
}

// literalNames collects the variables the body assigns a composite
// literal, which is how a table of cases is usually written.
func literalNames(pass *analysis.Pass, body *ast.BlockStmt) map[string]bool {
	names := make(map[string]bool)

	ast.Inspect(body, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}

		for index, value := range assign.Rhs {
			if !isCaseSlice(pass, value) {
				continue
			}

			if index < len(assign.Lhs) {
				recordName(names, assign.Lhs[index])
			}
		}

		return true
	})

	return names
}

func recordName(names map[string]bool, expr ast.Expr) {
	if ident, ok := expr.(*ast.Ident); ok {
		names[ident.Name] = true
	}
}
