package registry_test

import (
	"flag"
	"testing"

	"golang.org/x/tools/go/analysis"

	"govanish/internal/registry"
)

func TestAllRulesAreValid(t *testing.T) {
	t.Parallel()

	if err := analysis.Validate(registry.Analyzers()); err != nil {
		t.Fatalf("registry contains an invalid analyzer: %v", err)
	}
}

func TestRuleIdentifiersAreUnique(t *testing.T) {
	t.Parallel()

	seen := make(map[string]bool)
	for _, rule := range registry.All() {
		if seen[rule.ID] {
			t.Errorf("duplicate rule id %q", rule.ID)
		}
		seen[rule.ID] = true
	}
}

func TestEveryRuleIsDocumented(t *testing.T) {
	t.Parallel()

	for _, rule := range registry.All() {
		if rule.Analyzer.Doc == "" {
			t.Errorf("rule %q has no Doc", rule.ID)
		}
	}
}

func TestNoRuleDeclaresFlags(t *testing.T) {
	t.Parallel()

	for _, rule := range registry.All() {
		count := 0
		rule.Analyzer.Flags.VisitAll(func(*flag.Flag) { count++ })
		if count > 0 {
			t.Errorf("rule %q declares %d flags; govanish rules are not configurable", rule.ID, count)
		}
	}
}

func TestIDOfKnowsEveryRegisteredAnalyzer(t *testing.T) {
	t.Parallel()

	for _, rule := range registry.All() {
		if got := registry.IDOf(rule.Analyzer); got != rule.ID {
			t.Errorf("IDOf(%s) = %q, want %q", rule.Analyzer.Name, got, rule.ID)
		}
	}
}
