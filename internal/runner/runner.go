// Package runner loads Go packages and applies every registered rule to
// them.
package runner

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/checker"
	"golang.org/x/tools/go/packages"

	"govanish/internal/registry"
	"govanish/internal/report"
)

const loadMode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedSyntax |
	packages.NeedTypes |
	packages.NeedTypesInfo |
	packages.NeedTypesSizes |
	packages.NeedDeps |
	packages.NeedImports |
	packages.NeedModule

// ErrNoPackages reports that the given patterns matched nothing.
var ErrNoPackages = errors.New("no packages matched the given patterns")

// Run lints the packages matching patterns, resolved relative to dir, and
// returns their findings in printing order.
func Run(dir string, patterns []string) ([]report.Finding, error) {
	pkgs, err := load(dir, patterns)
	if err != nil {
		return nil, err
	}

	graph, err := checker.Analyze(registry.Analyzers(), pkgs, nil)
	if err != nil {
		return nil, fmt.Errorf("analyze: %w", err)
	}

	findings, err := collect(graph)
	if err != nil {
		return nil, err
	}

	report.Sort(findings)

	return deduplicate(findings), nil
}

func load(dir string, patterns []string) ([]*packages.Package, error) {
	config := &packages.Config{Dir: dir, Mode: loadMode, Tests: true}

	pkgs, err := packages.Load(config, patterns...)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}

	if len(pkgs) == 0 {
		return nil, ErrNoPackages
	}

	if packages.PrintErrors(pkgs) > 0 {
		return nil, errors.New("packages contain errors; fix them before linting")
	}

	return pkgs, nil
}

// deduplicate drops findings that repeat. go/packages loads a package
// once plainly and once more with its test files, so any finding in a
// non-test file of a tested package is produced twice.
func deduplicate(findings []report.Finding) []report.Finding {
	seen := make(map[report.Finding]bool, len(findings))
	unique := make([]report.Finding, 0, len(findings))

	for _, finding := range findings {
		if seen[finding] {
			continue
		}

		seen[finding] = true
		unique = append(unique, finding)
	}

	return unique
}

func collect(graph *checker.Graph) ([]report.Finding, error) {
	findings := make([]report.Finding, 0)

	for action := range graph.All() {
		if !action.IsRoot {
			continue
		}

		if action.Err != nil {
			return nil, fmt.Errorf("rule %s on %s: %w",
				registry.IDOf(action.Analyzer), action.Package.PkgPath, action.Err)
		}

		generated := generatedFiles(action.Package)

		for _, diagnostic := range action.Diagnostics {
			position := action.Package.Fset.Position(diagnostic.Pos)
			if generated[position.Filename] {
				continue
			}

			findings = append(findings, finding(action, diagnostic, position))
		}
	}

	return findings, nil
}

func finding(
	action *checker.Action,
	diagnostic analysis.Diagnostic,
	position token.Position,
) report.Finding {
	end := position
	if diagnostic.End.IsValid() {
		end = action.Package.Fset.Position(diagnostic.End)
	}

	return report.Finding{
		File:    position.Filename,
		Line:    position.Line,
		Col:     position.Column,
		EndLine: end.Line,
		EndCol:  end.Column,
		Rule:    registry.IDOf(action.Analyzer),
		Message: diagnostic.Message,
	}
}

func generatedFiles(pkg *packages.Package) map[string]bool {
	generated := make(map[string]bool)

	for _, file := range pkg.Syntax {
		if !ast.IsGenerated(file) {
			continue
		}

		name := pkg.Fset.Position(file.Pos()).Filename
		generated[name] = true
	}

	return generated
}
