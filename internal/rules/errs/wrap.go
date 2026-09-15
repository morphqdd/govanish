package errs

import (
	"go/ast"
	"go/constant"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// trailingPunctuation is what an error message may not end with, because
// errors are composed into sentences by their callers.
const trailingPunctuation = ".!?:;"

// Wrap reports errors that lose their cause, and error messages that read
// as sentences rather than as clauses.
func Wrap() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "wrap",
		Doc:      "errors must wrap their cause with %w and read as lowercase clauses",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runWrap,
	}
}

func runWrap(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(node ast.Node) {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return
		}

		pkg, name, ok := qualifiedName(call)
		if !ok || !isErrorConstructor(pkg, name) {
			return
		}

		checkMessage(pass, call)

		if name == "Errorf" {
			checkVerb(pass, call)
		}
	})

	return nil, nil
}

func isErrorConstructor(pkg, name string) bool {
	if pkg == "errors" && name == "New" {
		return true
	}

	return pkg == "fmt" && name == "Errorf"
}

// checkVerb reports an error argument formatted with %v, which prints the
// cause but does not keep it reachable through errors.Is.
func checkVerb(pass *analysis.Pass, call *ast.CallExpr) {
	format, ok := stringValue(pass, call.Args[0])
	if !ok || strings.Contains(format, "%w") {
		return
	}

	if !wrapsAnError(pass, call.Args[1:]) {
		return
	}

	pass.Reportf(call.Pos(), "wrap the error with %%w, not %%v")
}

func wrapsAnError(pass *analysis.Pass, args []ast.Expr) bool {
	for _, arg := range args {
		if astutil.IsError(pass.TypesInfo.TypeOf(arg)) {
			return true
		}
	}

	return false
}

func checkMessage(pass *analysis.Pass, call *ast.CallExpr) {
	if len(call.Args) == 0 {
		return
	}

	message, ok := stringValue(pass, call.Args[0])
	if !ok || message == "" {
		return
	}

	if unicode.IsUpper([]rune(message)[0]) {
		pass.Reportf(call.Args[0].Pos(), "error message must start lowercase")
	}

	if strings.ContainsRune(trailingPunctuation, rune(message[len(message)-1])) {
		pass.Reportf(call.Args[0].Pos(), "error message must not end with punctuation")
	}
}

func stringValue(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	value := pass.TypesInfo.Types[expr].Value
	if value == nil || value.Kind() != constant.String {
		return "", false
	}

	return constant.StringVal(value), true
}

// qualifiedName splits a call of the form pkg.Name into its parts.
func qualifiedName(call *ast.CallExpr) (string, string, bool) {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", "", false
	}

	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return "", "", false
	}

	return ident.Name, selector.Sel.Name, true
}
