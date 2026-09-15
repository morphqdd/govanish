package tests

import (
	"go/ast"
	"strings"
	"unicode"
)

// isTestFile reports whether a file is a Go test file.
func isTestFile(name string) bool {
	return strings.HasSuffix(name, "_test.go")
}

// testParam returns the name of a test function's *testing.T parameter,
// and whether the declaration is a test function at all.
func testParam(decl *ast.FuncDecl) (string, bool) {
	if decl.Recv != nil || !strings.HasPrefix(decl.Name.Name, "Test") {
		return "", false
	}

	params := decl.Type.Params.List
	if len(params) != 1 || len(params[0].Names) != 1 {
		return "", false
	}

	if !isTestingT(params[0].Type) {
		return "", false
	}

	return params[0].Names[0].Name, true
}

func isTestingT(expr ast.Expr) bool {
	pointer, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}

	selector, ok := pointer.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	pkg, ok := selector.X.(*ast.Ident)

	return ok && pkg.Name == "testing" && selector.Sel.Name == "T"
}

// calls reports whether the body calls method on the named receiver, as
// in t.Parallel.
func calls(body *ast.BlockStmt, receiver, method string) bool {
	found := false

	ast.Inspect(body, func(node ast.Node) bool {
		if isMethodCall(node, receiver, method) {
			found = true
		}

		return !found
	})

	return found
}

func isMethodCall(node ast.Node, receiver, method string) bool {
	call, ok := node.(*ast.CallExpr)
	if !ok {
		return false
	}

	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != method {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)

	return ok && ident.Name == receiver
}

// startsWithCapital reports whether the part of a test name after the
// Test prefix begins with a capital, as TestFoo does and Testfoo does not.
func startsWithCapital(name string) bool {
	rest := strings.TrimPrefix(name, "Test")
	if rest == "" {
		return true
	}

	return unicode.IsUpper([]rune(rest)[0])
}
