package tests_test

import (
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/morphqdd/govanish/internal/rules/tests"
)

func run(t *testing.T, analyzer *analysis.Analyzer, pkg string) {
	t.Helper()
	t.Parallel()

	analysistest.Run(t, analysistest.TestData(), analyzer, pkg)
}

func TestShape(t *testing.T)    { run(t, tests.Shape(), "shape") }
func TestNoSkip(t *testing.T)   { run(t, tests.NoSkip(), "skip") }
func TestParallel(t *testing.T) { run(t, tests.Parallel(), "parallel") }
func TestPkgDoc(t *testing.T)   { run(t, tests.PkgDoc(), "nopkgdoc") }
func TestPkgDocOK(t *testing.T) { run(t, tests.PkgDoc(), "pkgdoc") }

func TestExportedDoc(t *testing.T) { run(t, tests.ExportedDoc(), "exporteddoc") }
