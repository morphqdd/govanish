package arch_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"govanish/internal/rules/arch"
)

func TestNoInit(t *testing.T) {
	t.Parallel()

	analysistest.Run(t, analysistest.TestData(), arch.NoInit(), "noinit")
}

func TestNoGlobals(t *testing.T) {
	t.Parallel()

	analysistest.Run(t, analysistest.TestData(), arch.NoGlobals(), "noglobals")
}

func TestNoAny(t *testing.T) {
	t.Parallel()

	analysistest.Run(t, analysistest.TestData(), arch.NoAny(), "noany")
}

func TestBannedImports(t *testing.T) {
	t.Parallel()

	analysistest.Run(t, analysistest.TestData(), arch.BannedImports(), "bannedimports")
}
