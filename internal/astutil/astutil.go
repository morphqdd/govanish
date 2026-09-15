// Package astutil holds helpers shared by rules: retrieving the driver's
// shared inspector, and describing declarations in the words diagnostics
// use.
package astutil

import (
	"errors"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// ErrMissingInspector reports that the driver did not supply the shared
// inspector, which means the analyzer ran without declaring inspect in
// its Requires.
var ErrMissingInspector = errors.New("inspect analyzer result missing")

// Inspector returns the shared inspector for the package under analysis.
func Inspector(pass *analysis.Pass) (*inspector.Inspector, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, ErrMissingInspector
	}

	return insp, nil
}

// Describe names a function declaration the way a diagnostic should refer
// to it: "function Parse" or "method Thing.Parse".
func Describe(decl *ast.FuncDecl) string {
	if decl.Recv == nil {
		return "function " + decl.Name.Name
	}

	return "method " + decl.Name.Name
}
