package report_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"govanish/internal/report"
)

func unsorted() []report.Finding {
	return []report.Finding{
		{
			File: "internal/user/service.go", Line: 42, Col: 6,
			EndLine: 42, EndCol: 16,
			Rule:    "no-bare-prim",
			Message: "parameter name is a bare string",
		},
		{
			File: "internal/user/service.go", Line: 12, Col: 1,
			EndLine: 12, EndCol: 10,
			Rule:    "no-init",
			Message: "func init is banned",
		},
		{
			File: "internal/user/service.go", Line: 42, Col: 6,
			EndLine: 42, EndCol: 16,
			Rule:    "func-len",
			Message: "function CreateUser is 61 lines, limit is 40",
		},
	}
}

func golden(t *testing.T, name string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}

	return string(content)
}

func TestTextOutputMatchesGolden(t *testing.T) {
	t.Parallel()

	findings := unsorted()
	report.Sort(findings)

	var buf bytes.Buffer
	if err := report.Text(&buf, findings); err != nil {
		t.Fatalf("render text: %v", err)
	}

	if buf.String() != golden(t, "text.golden") {
		t.Errorf("text output mismatch:\ngot:\n%s\nwant:\n%s", buf.String(), golden(t, "text.golden"))
	}
}

func TestJSONOutputMatchesGolden(t *testing.T) {
	t.Parallel()

	findings := unsorted()
	report.Sort(findings)

	var buf bytes.Buffer
	if err := report.JSON(&buf, findings); err != nil {
		t.Fatalf("render json: %v", err)
	}

	if buf.String() != golden(t, "json.golden") {
		t.Errorf("json output mismatch:\ngot:\n%s\nwant:\n%s", buf.String(), golden(t, "json.golden"))
	}
}

func TestEmptyFindingsRenderAsEmptyJSONArray(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := report.JSON(&buf, nil); err != nil {
		t.Fatalf("render json: %v", err)
	}

	if buf.String() != "[]\n" {
		t.Errorf("empty json = %q, want %q", buf.String(), "[]\n")
	}
}
