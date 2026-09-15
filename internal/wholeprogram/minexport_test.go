package wholeprogram_test

import (
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"

	"github.com/morphqdd/govanish/internal/report"
	"github.com/morphqdd/govanish/internal/wholeprogram"
)

func load(t *testing.T) []*packages.Package {
	t.Helper()

	mode := packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
		packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps |
		packages.NeedImports

	pkgs, err := packages.Load(&packages.Config{
		Dir:  "testdata/fixture",
		Mode: mode,
	}, "./...")
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	return pkgs
}

func names(findings []report.Finding) []string {
	var found []string

	for _, finding := range findings {
		fields := strings.Fields(finding.Message)
		found = append(found, fields[2])
	}

	return found
}

// TestMinExportFindsOrphans checks both directions at once: an export
// another package calls is left alone, and one nobody calls is reported.
func TestMinExportFindsOrphans(t *testing.T) {
	t.Parallel()

	found := names(wholeprogram.MinExport(load(t)))

	wants := map[string]bool{"Orphan": true}
	forbidden := map[string]bool{"Consumed": true, "helper": true}

	for _, name := range found {
		if forbidden[name] {
			t.Errorf("reported %q, which is used or unexported", name)
		}

		delete(wants, name)
	}

	for name := range wants {
		t.Errorf("did not report %q, which nothing uses", name)
	}
}
