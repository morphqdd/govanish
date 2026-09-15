package size

import (
	"bytes"
	"fmt"
	"go/token"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
)

// maxLineColumns is the widest a source line may be.
const maxLineColumns = 100

// LineLength reports source lines too wide to read in a split window.
var LineLength = &analysis.Analyzer{
	Name: "linelength",
	Doc:  fmt.Sprintf("source lines may not exceed %d columns", maxLineColumns),
	Run:  runLineLength,
}

func runLineLength(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		tokenFile := pass.Fset.File(file.Pos())

		content, err := pass.ReadFile(tokenFile.Name())
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", tokenFile.Name(), err)
		}

		reportLongLines(pass, tokenFile, content)
	}

	return nil, nil
}

func reportLongLines(pass *analysis.Pass, tokenFile *token.File, content []byte) {
	for index, line := range bytes.Split(content, []byte("\n")) {
		columns := utf8.RuneCount(bytes.TrimSuffix(line, []byte("\r")))
		if columns <= maxLineColumns {
			continue
		}

		number := index + 1
		if number > tokenFile.LineCount() {
			continue
		}

		pass.Reportf(tokenFile.LineStart(number), "line is %d columns, limit is %d",
			columns, maxLineColumns)
	}
}
