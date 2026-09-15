package runner_test

import (
	"strings"
	"testing"

	"govanish/internal/runner"
)

const fixtureDir = "testdata/fixture"

func TestRunReportsViolations(t *testing.T) {
	t.Parallel()

	findings, err := runner.Run(fixtureDir, []string{"./..."})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1: %+v", len(findings), findings)
	}

	finding := findings[0]
	if finding.Rule != "no-init" {
		t.Errorf("rule = %q, want %q", finding.Rule, "no-init")
	}
	if !strings.HasSuffix(finding.File, "bad/bad.go") {
		t.Errorf("file = %q, want it to end with bad/bad.go", finding.File)
	}
	if finding.Line != 5 {
		t.Errorf("line = %d, want 5", finding.Line)
	}
}

func TestRunSkipsGeneratedFiles(t *testing.T) {
	t.Parallel()

	findings, err := runner.Run(fixtureDir, []string{"./gen/..."})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if len(findings) != 0 {
		t.Errorf("got %d findings in generated code, want 0: %+v", len(findings), findings)
	}
}

func TestRunReturnsErrorForUnloadablePackages(t *testing.T) {
	t.Parallel()

	_, err := runner.Run(fixtureDir, []string{"./does-not-exist/..."})
	if err == nil {
		t.Fatal("want an error for a pattern matching no packages, got nil")
	}
}
