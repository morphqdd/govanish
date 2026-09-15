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

// TestEveryRuleIsRegistered pins the full rule set by name. Without it a
// rule can be dropped from the registry and every other test still
// passes: the rule's own test keeps working, and the linter simply stops
// enforcing it. That happened once.
func TestEveryRuleIsRegistered(t *testing.T) {
	t.Parallel()

	want := []string{
		"banned-imports", "comments-doc-only", "ctx-first", "ctx-origin",
		"cyclomatic", "decl-order", "defer-unlock", "errcheck", "error-shape",
		"errors-is", "exported-doc", "file-complexity", "file-len", "func-len",
		"go-needs-owner", "import-groups", "initialisms", "keyed-literals",
		"line-length", "nesting-depth", "no-any", "no-bool-param",
		"no-else-after-return", "no-empty-err-branch", "no-globals", "no-init",
		"no-naked-return", "no-panic", "no-shadow", "no-skip",
		"no-time-after-select", "no-underscore-names", "param-count", "pkg-doc",
		"return-count", "sized-chan", "struct-fields", "test-parallel",
		"test-shape", "wrap",
	}

	got := make(map[string]bool)
	for _, rule := range registry.New().Rules() {
		got[rule.ID] = true
	}

	for _, id := range want {
		t.Run(id, func(t *testing.T) {
			if !got[id] {
				t.Errorf("rule %q is not registered", id)
			}
		})
	}

	if len(got) != len(want) {
		t.Errorf("registry has %d rules, want %d", len(got), len(want))
	}
}
