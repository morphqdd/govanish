package style_test

import (
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/morphqdd/govanish/internal/rules/style"
)

func run(t *testing.T, analyzer *analysis.Analyzer, pkg string) {
	t.Helper()
	t.Parallel()

	analysistest.Run(t, analysistest.TestData(), analyzer, pkg)
}

func TestNoElse(t *testing.T)        { run(t, style.NoElse(), "noelse") }
func TestNoNakedReturn(t *testing.T) { run(t, style.NoNakedReturn(), "nonakedreturn") }
func TestUnderscores(t *testing.T)   { run(t, style.Underscores(), "underscores") }
func TestInitialisms(t *testing.T)   { run(t, style.Initialisms(), "initialisms") }

func TestComments(t *testing.T)     { run(t, style.Comments(), "comments") }
func TestDeclOrder(t *testing.T)    { run(t, style.DeclOrder(), "declorder") }
func TestDeclOrderBad(t *testing.T) { run(t, style.DeclOrder(), "declorderbad") }

func TestImportGroups(t *testing.T) { run(t, style.ImportGroups(), "importgroups") }
func TestImportBad(t *testing.T)    { run(t, style.ImportGroups(), "importbad") }

func TestNameLength(t *testing.T) { run(t, style.NameLength(), "namelength") }
