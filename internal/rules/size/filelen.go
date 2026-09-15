package size

import (
	"fmt"

	"golang.org/x/tools/go/analysis"
)

// maxFileLines is the most lines a source file may contain.
const maxFileLines = 400

// FileLen reports source files that have grown past reading size.
func FileLen() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "filelen",
		Doc:  fmt.Sprintf("source files may not exceed %d lines", maxFileLines),
		Run:  runFileLen,
	}
}

func runFileLen(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		lines := pass.Fset.File(file.Pos()).LineCount()
		if lines <= maxFileLines {
			continue
		}

		pass.Reportf(file.Package, "file has %d lines, limit is %d", lines, maxFileLines)
	}

	return nil, nil
}
