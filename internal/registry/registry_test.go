package registry_test

import (
	"flag"
	"testing"

	"golang.org/x/tools/go/analysis"

	"govanish/internal/registry"
)

func TestAllRulesAreValid(t *testing.T) {
	t.Parallel()

	if err := analysis.Validate(registry.New().Analyzers()); err != nil {
		t.Fatalf("registry contains an invalid analyzer: %v", err)
	}
}

func TestRuleIdentifiersAreUnique(t *testing.T) {
	t.Parallel()

	seen := make(map[string]bool)
	for _, rule := range registry.New().Rules() {
		if seen[rule.ID] {
			t.Errorf("duplicate rule id %q", rule.ID)
		}
		seen[rule.ID] = true
	}
}

func TestEveryRuleIsDocumented(t *testing.T) {
	t.Parallel()

	for _, rule := range registry.New().Rules() {
		if rule.Analyzer.Doc == "" {
			t.Errorf("rule %q has no Doc", rule.ID)
		}
	}
}

func TestNoRuleDeclaresFlags(t *testing.T) {
	t.Parallel()

	for _, rule := range registry.New().Rules() {
		count := 0
		rule.Analyzer.Flags.VisitAll(func(*flag.Flag) { count++ })
		if count > 0 {
			t.Errorf("rule %q declares %d flags; govanish rules are not configurable", rule.ID, count)
		}
	}
}

func TestIDOfKnowsEveryRegisteredAnalyzer(t *testing.T) {
	t.Parallel()

	set := registry.New()
	for _, rule := range set.Rules() {
		if got := set.IDOf(rule.Analyzer); got != rule.ID {
			t.Errorf("IDOf(%s) = %q, want %q", rule.Analyzer.Name, got, rule.ID)
		}
	}
}

// TestSetIsStableWithinItself checks the invariant the analysis driver
// relies on: analyzers are identified by pointer, so the analyzers a Set
// hands out must be the same ones it can name.
func TestSetIsStableWithinItself(t *testing.T) {
	t.Parallel()

	set := registry.New()

	analyzers := set.Analyzers()
	rules := set.Rules()

	for index, analyzer := range analyzers {
		if analyzer != rules[index].Analyzer {
			t.Fatalf("analyzer %d differs between Analyzers and Rules", index)
		}
	}
}
