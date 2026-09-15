package size

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// maxFileComplexity is the most total decision paths a file may hold.
// It exists because func-len and cyclomatic are both per-function: a file
// of thirty individually simple functions passes every other size rule
// and is still too much to hold in the head at once.
const maxFileComplexity = 40

// FileComplexity reports a file whose functions are each simple enough
// but which together carry too much branching.
func FileComplexity() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "filecomplexity",
		Doc:  fmt.Sprintf("total complexity of a file may not exceed %d", maxFileComplexity),
		Run:  runFileComplexity,
	}
}

func runFileComplexity(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		total := fileComplexity(file)
		if total <= maxFileComplexity {
			continue
		}

		pass.Reportf(file.Package, "file has a total complexity of %d, limit is %d",
			total, maxFileComplexity)
	}

	return nil, nil
}

// fileComplexity sums the McCabe complexity of every function in a file,
// counting a function literal as part of the function that contains it.
func fileComplexity(file *ast.File) int {
	total := 0

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		total += complexity(fn.Body)
	}

	return total
}
