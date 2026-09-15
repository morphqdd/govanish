package api_test

import (
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/morphqdd/govanish/internal/rules/api"
)

func run(t *testing.T, analyzer *analysis.Analyzer, pkg string) {
	t.Helper()
	t.Parallel()

	analysistest.Run(t, analysistest.TestData(), analyzer, pkg)
}

func TestNoBoolParam(t *testing.T)   { run(t, api.NoBoolParam(), "noboolparam") }
func TestKeyedLiterals(t *testing.T) { run(t, api.KeyedLiterals(), "keyedliterals") }

func TestNoIfaceReturn(t *testing.T) { run(t, api.NoIfaceReturn(), "noifacereturn") }
func TestNoBarePrim(t *testing.T)    { run(t, api.NoBarePrim(), "nobareprim") }
