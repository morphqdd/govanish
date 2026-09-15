package main

import (
	"bytes"
	"strings"
	"testing"
)

const fixtureDir = "../../internal/runner/testdata/fixture"

func TestRunCleanPackageExitsZero(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer

	code := run([]string{"-format", "text", "./good/..."}, fixtureDir, &out, &errOut)

	if code != exitClean {
		t.Errorf("exit code = %d, want %d; stderr: %s", code, exitClean, errOut.String())
	}
	if out.String() != "" {
		t.Errorf("output = %q, want empty", out.String())
	}
}

func TestRunDirtyPackageExitsOne(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer

	code := run([]string{"-format", "text", "./bad/..."}, fixtureDir, &out, &errOut)

	if code != exitFindings {
		t.Errorf("exit code = %d, want %d; stderr: %s", code, exitFindings, errOut.String())
	}
	if !strings.Contains(out.String(), "[no-init]") {
		t.Errorf("output = %q, want it to mention [no-init]", out.String())
	}
}

func TestRunLoadFailureExitsTwo(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer

	code := run([]string{"./does-not-exist/..."}, fixtureDir, &out, &errOut)

	if code != exitFailure {
		t.Errorf("exit code = %d, want %d", code, exitFailure)
	}
	if errOut.String() == "" {
		t.Error("want an explanation on stderr, got nothing")
	}
}

func TestUnknownFormatExitsTwo(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer

	code := run([]string{"-format", "xml", "./good/..."}, fixtureDir, &out, &errOut)

	if code != exitFailure {
		t.Errorf("exit code = %d, want %d", code, exitFailure)
	}
}
