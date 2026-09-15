package size_test

import (
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"

	"govanish/internal/rules/size"
)

func run(t *testing.T, analyzer *analysis.Analyzer, pkg string) {
	t.Helper()
	t.Parallel()

	analysistest.Run(t, analysistest.TestData(), analyzer, pkg)
}

func TestFuncLen(t *testing.T)      { run(t, size.FuncLen(), "funclen") }
func TestFileLen(t *testing.T)      { run(t, size.FileLen(), "filelen") }
func TestLineLength(t *testing.T)   { run(t, size.LineLength(), "linelength") }
func TestCyclo(t *testing.T)        { run(t, size.Cyclo(), "cyclo") }
func TestNesting(t *testing.T)      { run(t, size.Nesting(), "nesting") }
func TestParamCount(t *testing.T)   { run(t, size.ParamCount(), "paramcount") }
func TestReturnCount(t *testing.T)  { run(t, size.ReturnCount(), "returncount") }
func TestStructFields(t *testing.T) { run(t, size.StructFields(), "structfields") }
