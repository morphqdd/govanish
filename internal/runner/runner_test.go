package runner_test

import (
	"strings"
	"testing"

	"govanish/internal/report"
	"govanish/internal/runner"
)

const fixtureDir = "testdata/fixture"

// TestRunReportsViolations asserts on the rule under test rather than on
// the total number of findings, which grows with the rule set. Requiring
// exactly one occurrence also proves deduplication: the fixture package
// has a test file, so go/packages loads it twice.
func TestRunReportsViolations(t *testing.T) {
	t.Parallel()

	findings, err := runner.Run(fixtureDir, []string{"./..."})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	matched := findingsOf(findings, "no-init")
	if len(matched) != 1 {
		t.Fatalf("got %d no-init findings, want 1: %+v", len(matched), matched)
	}

	finding := matched[0]
	if !strings.HasSuffix(finding.File, "bad/bad.go") {
		t.Errorf("file = %q, want it to end with bad/bad.go", finding.File)
	}
	if finding.Line != 5 {
		t.Errorf("line = %d, want 5", finding.Line)
	}
}

func findingsOf(findings []report.Finding, rule string) []report.Finding {
	var matched []report.Finding

	for _, finding := range findings {
		if finding.Rule == rule {
			matched = append(matched, finding)
		}
	}

	return matched
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
