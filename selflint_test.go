package govanish_test

import (
	"bytes"
	"testing"

	"govanish/internal/report"
	"govanish/internal/runner"
)

// TestSelfLint holds govanish to its own rules. It is expected to fail
// whenever a newly added rule finds a violation in govanish's own source;
// the fix is always to change govanish, never to weaken the rule.
func TestSelfLint(t *testing.T) {
	t.Parallel()

	findings, err := runner.Run(runner.Dir("."), []string{"./cmd/...", "./internal/..."})
	if err != nil {
		t.Fatalf("lint own source: %v", err)
	}

	if len(findings) == 0 {
		return
	}

	var buf bytes.Buffer
	if err := report.Text(&buf, findings); err != nil {
		t.Fatalf("render findings: %v", err)
	}

	t.Errorf("govanish does not satisfy its own rules:\n%s", buf.String())
}
