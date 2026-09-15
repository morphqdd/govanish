package arch_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"govanish/internal/rules/arch"
)

func TestNoInit(t *testing.T) {
	t.Parallel()

	analysistest.Run(t, analysistest.TestData(), arch.NoInit, "noinit")
}
