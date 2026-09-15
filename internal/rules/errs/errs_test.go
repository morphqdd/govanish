package errs_test

import (
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"

	"govanish/internal/rules/errs"
)

func run(t *testing.T, analyzer *analysis.Analyzer, pkg string) {
	t.Helper()
	t.Parallel()

	analysistest.Run(t, analysistest.TestData(), analyzer, pkg)
}

func TestErrCheck(t *testing.T)    { run(t, errs.ErrCheck(), "errcheck") }
func TestEmptyBranch(t *testing.T) { run(t, errs.EmptyBranch(), "emptybranch") }
func TestErrorsIs(t *testing.T)    { run(t, errs.ErrorsIs(), "errorsis") }
func TestNoPanic(t *testing.T)     { run(t, errs.NoPanic(), "nopanic") }
func TestErrShape(t *testing.T)    { run(t, errs.ErrShape(), "errshape") }
func TestWrap(t *testing.T)        { run(t, errs.Wrap(), "wrap") }
