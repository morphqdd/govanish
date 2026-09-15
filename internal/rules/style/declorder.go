package style

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// rankFunc is the rank of a function declaration, which follows every
// kind of general declaration.
const rankFunc = 4

// DeclOrder reports top-level declarations that appear out of order.
func DeclOrder() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "declorder",
		Doc:  "declarations must be ordered const, var, type, then func",
		Run:  runDeclOrder,
	}
}

// declRank() orders the kinds of top-level declaration. A file reads from
// the most constant thing in it to the most active.
func declRank() map[token.Token]int {
	return map[token.Token]int{
		token.IMPORT: 0,
		token.CONST:  1,
		token.VAR:    2,
		token.TYPE:   3,
	}
}

// rankName() names a rank for a diagnostic.
func rankName() map[int]string {
	return map[int]string{
		0: "import",
		1: "const",
		2: "var",
		3: "type",
		4: "func",
	}
}

func runDeclOrder(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		checkDeclOrder(pass, file)

		if !isTestFile(pass, file) {
			checkExportOrder(pass, file)
		}
	}

	return nil, nil
}

// isTestFile reports whether a file is a test. A test file reads as a
// narrative, with helpers introduced before the tests that use them, so
// export order does not apply to it.
func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	name := pass.Fset.Position(file.Pos()).Filename

	return strings.HasSuffix(name, "_test.go")
}

func checkDeclOrder(pass *analysis.Pass, file *ast.File) {
	highest := 0

	for _, decl := range file.Decls {
		rank := rankOf(decl)
		if rank >= highest {
			highest = rank

			continue
		}

		pass.Reportf(decl.Pos(), "%s declaration must come before %s declarations",
			rankName()[rank], rankName()[highest])
	}
}

// checkExportOrder reports an exported function declared after an
// unexported one, so that a reader meets a package's surface first.
func checkExportOrder(pass *analysis.Pass, file *ast.File) {
	seenUnexported := false

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil {
			continue
		}

		if !fn.Name.IsExported() {
			seenUnexported = true

			continue
		}

		if !seenUnexported {
			continue
		}

		pass.Reportf(fn.Pos(),
			"exported function %s must be declared before unexported functions",
			fn.Name.Name)
	}
}

func rankOf(decl ast.Decl) int {
	gen, ok := decl.(*ast.GenDecl)
	if !ok {
		return rankFunc
	}

	return declRank()[gen.Tok]
}
