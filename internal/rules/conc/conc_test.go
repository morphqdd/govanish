package conc_test

import (
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/morphqdd/govanish/internal/rules/conc"
)

func run(t *testing.T, analyzer *analysis.Analyzer, pkg string) {
	t.Helper()
	t.Parallel()

	analysistest.Run(t, analysistest.TestData(), analyzer, pkg)
}

func TestCtxFirst(t *testing.T)    { run(t, conc.CtxFirst(), "ctxfirst") }
func TestCtxOrigin(t *testing.T)   { run(t, conc.CtxOrigin(), "ctxorigin") }
func TestDeferUnlock(t *testing.T) { run(t, conc.DeferUnlock(), "deferunlock") }
func TestSizedChan(t *testing.T)   { run(t, conc.SizedChan(), "sizedchan") }
func TestGoOwner(t *testing.T)     { run(t, conc.GoOwner(), "goowner") }
func TestTimeAfter(t *testing.T)   { run(t, conc.TimeAfter(), "timeafter") }
