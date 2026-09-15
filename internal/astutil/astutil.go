// Package astutil holds helpers shared by rules: retrieving the driver's
// shared inspector, and describing declarations in the words diagnostics
// use.
package astutil

import (
	"errors"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Description is how a diagnostic refers to a declaration. It is a named
// type because govanish bans bare strings in exported signatures.
type Description string

// Inspector returns the shared inspector for the package under analysis.
func Inspector(pass *analysis.Pass) (*inspector.Inspector, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, errMissingInspector()
	}

	return insp, nil
}

// Describe names a function declaration the way a diagnostic should refer
// to it: "function Parse" or "method Parse".
func Describe(decl *ast.FuncDecl) Description {
	if decl.Recv == nil {
		return Description("function " + decl.Name.Name)
	}

	return Description("method " + decl.Name.Name)
}

// IsError reports whether a type is exactly the predeclared error
// interface, rather than merely implementing it.
func IsError(typ types.Type) bool {
	if typ == nil {
		return false
	}

	return types.Identical(typ, errorType())
}

// ResultTypes returns the result types of whatever a call expression
// evaluates to, flattening a tuple into its members.
func ResultTypes(info *types.Info, call *ast.CallExpr) []types.Type {
	typ := info.TypeOf(call)
	if typ == nil {
		return nil
	}

	tuple, ok := typ.(*types.Tuple)
	if !ok {
		return []types.Type{typ}
	}

	results := make([]types.Type, 0, tuple.Len())
	for index := range tuple.Len() {
		results = append(results, tuple.At(index).Type())
	}

	return results
}

// errMissingInspector reports that the driver did not supply the shared
// inspector, which means the analyzer ran without declaring inspect in
// its Requires.
func errMissingInspector() error {
	return errors.New("inspect analyzer result missing")
}

// errorType is the predeclared error interface.
func errorType() types.Type {
	return types.Universe.Lookup("error").Type()
}
