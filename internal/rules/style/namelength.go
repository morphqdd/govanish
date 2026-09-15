package style

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/morphqdd/govanish/internal/astutil"
)

// minNameLength is the shortest an identifier may be. A name that fits in
// two characters has not been thought about.
const minNameLength = 3

// maxReceiverLength is the longest a method receiver may be. A receiver
// is read once per method and never needs to say more than which type it
// belongs to.
const maxReceiverLength = 2

// NameLength reports identifiers too short to say what they hold.
func NameLength() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "namelength",
		Doc:      "identifiers must be at least three characters, receivers at most two",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runNameLength,
	}
}

// shortNames are the names Go itself has made unambiguous. Every one of
// them means the same thing in every Go program ever written. The
// capitalized entries are the exported spellings, which the initialisms
// rule requires in full capitals.
func shortNames() map[string]bool {
	return map[string]bool{
		"err": true, "ok": true, "ctx": true, "id": true, "db": true,
		"tx": true, "mu": true, "wg": true, "fn": true, "_": true,
		"ID": true, "OK": true, "DB": true, "TX": true,
	}
}

// testNames are the names the testing package's own signatures dictate.
func testNames() map[string]bool {
	return map[string]bool{"t": true, "b": true, "f": true}
}

func runNameLength(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	receivers := receiverNames(insp)

	insp.Preorder([]ast.Node{(*ast.Ident)(nil)}, func(node ast.Node) {
		ident, ok := node.(*ast.Ident)
		if !ok || receivers[ident] {
			return
		}

		if !isShort(pass, ident) {
			return
		}

		pass.Reportf(ident.Pos(),
			"identifier %s is too short; names must be at least %d characters",
			ident.Name, minNameLength)
	})

	reportLongReceivers(pass, insp)

	return nil, nil
}

func isShort(pass *analysis.Pass, ident *ast.Ident) bool {
	if len(ident.Name) >= minNameLength || exemptShortName(pass, ident) {
		return false
	}

	_, declares := pass.TypesInfo.Defs[ident]

	return declares
}

func exemptShortName(pass *analysis.Pass, ident *ast.Ident) bool {
	if shortNames()[ident.Name] {
		return true
	}

	name := pass.Fset.Position(ident.Pos()).Filename

	return strings.HasSuffix(name, "_test.go") && testNames()[ident.Name]
}

// receiverNames collects the receiver identifier of every method, which
// the length rule exempts and the receiver rule checks instead.
func receiverNames(insp *inspector.Inspector) map[*ast.Ident]bool {
	names := make(map[*ast.Ident]bool)

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		decl, ok := node.(*ast.FuncDecl)
		if !ok || decl.Recv == nil {
			return
		}

		for _, field := range decl.Recv.List {
			for _, name := range field.Names {
				names[name] = true
			}
		}
	})

	return names
}

func reportLongReceivers(pass *analysis.Pass, insp *inspector.Inspector) {
	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		decl, ok := node.(*ast.FuncDecl)
		if !ok || decl.Recv == nil {
			return
		}

		for _, field := range decl.Recv.List {
			reportReceiverField(pass, field)
		}
	})
}

func reportReceiverField(pass *analysis.Pass, field *ast.Field) {
	for _, name := range field.Names {
		if name.Name == "_" || len(name.Name) <= maxReceiverLength {
			continue
		}

		pass.Reportf(name.Pos(), "receiver %s is too long; use at most %d characters",
			name.Name, maxReceiverLength)
	}
}
