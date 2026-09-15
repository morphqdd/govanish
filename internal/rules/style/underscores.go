package style

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// testPrefixes are the identifier prefixes Go's own testing conventions
// require to carry an underscore.
var testPrefixes = []string{"Test_", "Benchmark_", "Example_", "Fuzz_"}

// Underscores reports identifiers spelled with underscores instead of the
// camel case Go uses everywhere else.
var Underscores = &analysis.Analyzer{
	Name:     "underscores",
	Doc:      "identifiers must be camel case, not underscore separated",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runUnderscores,
}

func runUnderscores(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	insp.Preorder([]ast.Node{(*ast.Ident)(nil)}, func(node ast.Node) {
		ident, ok := node.(*ast.Ident)
		if !ok || !isDeclaration(pass, ident) {
			return
		}

		if !strings.Contains(ident.Name, "_") || exemptName(ident.Name) {
			return
		}

		pass.Reportf(ident.Pos(), "identifier %s contains an underscore", ident.Name)
	})

	return nil, nil
}

// isDeclaration reports whether this occurrence of the identifier is the
// one that introduces it, so that a badly named identifier is reported
// once rather than at every use.
func isDeclaration(pass *analysis.Pass, ident *ast.Ident) bool {
	_, declares := pass.TypesInfo.Defs[ident]

	return declares
}

// exemptName reports whether an underscore in this name is required
// rather than chosen: the blank identifier, Go's external test package
// suffix, and the testing framework's own prefixes.
func exemptName(name string) bool {
	if name == "_" {
		return true
	}

	if strings.HasSuffix(name, "_test") {
		return true
	}

	for _, prefix := range testPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	return false
}
