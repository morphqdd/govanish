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

	return loadConfig(t, packages.Config{})
}

// loadWithTests loads the fixture the way the runner does, so a package
// with in-package tests arrives twice: plain and test-augmented.
func loadWithTests(t *testing.T) []*packages.Package {
	t.Helper()

	return loadConfig(t, packages.Config{Tests: true})
}

func loadConfig(t *testing.T, config packages.Config) []*packages.Package {
	t.Helper()

	config.Dir = "testdata/fixture"
	config.Mode = packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
		packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps |
		packages.NeedImports

	pkgs, err := packages.Load(&config, "./...")
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

// TestMinExportCountsTestVariants checks that a use from another package
// still counts when the used package also has a test-augmented variant,
// whose objects are distinct from the ones the importer references.
func TestMinExportCountsTestVariants(t *testing.T) {
	t.Parallel()

	reportedOrphan := false

	for _, name := range names(wholeprogram.MinExport(loadWithTests(t))) {
		if name == "Consumed" {
			t.Errorf("reported %q, which another package uses", name)
		}

		reportedOrphan = reportedOrphan || name == "Orphan"
	}

	if !reportedOrphan {
		t.Error(`did not report "Orphan", which nothing uses`)
	}
}
