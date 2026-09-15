// Package registry is the single place where every govanish rule is
// listed. A rule that is not here does not run.
package registry

import (
	"golang.org/x/tools/go/analysis"

	"govanish/internal/rules/arch"
	"govanish/internal/rules/size"
)

// Rule pairs an analyzer with the identifier govanish prints for it.
// The two differ because analysis.Validate requires analyzer names to be
// valid Go identifiers, while diagnostics read better hyphenated.
type Rule struct {
	ID       string
	Analyzer *analysis.Analyzer
}

var rules = []Rule{
	{ID: "no-init", Analyzer: arch.NoInit},

	{ID: "func-len", Analyzer: size.FuncLen},
	{ID: "file-len", Analyzer: size.FileLen},
	{ID: "line-length", Analyzer: size.LineLength},
	{ID: "cyclomatic", Analyzer: size.Cyclo},
	{ID: "nesting-depth", Analyzer: size.Nesting},
	{ID: "param-count", Analyzer: size.ParamCount},
	{ID: "return-count", Analyzer: size.ReturnCount},
	{ID: "struct-fields", Analyzer: size.StructFields},
}

// All returns every registered rule.
func All() []Rule {
	out := make([]Rule, len(rules))
	copy(out, rules)

	return out
}

// Analyzers returns the analyzer of every registered rule.
func Analyzers() []*analysis.Analyzer {
	out := make([]*analysis.Analyzer, 0, len(rules))
	for _, rule := range rules {
		out = append(out, rule.Analyzer)
	}

	return out
}

// IDOf returns the printable identifier of a registered analyzer, or the
// analyzer's own name if it was never registered.
func IDOf(target *analysis.Analyzer) string {
	for _, rule := range rules {
		if rule.Analyzer == target {
			return rule.ID
		}
	}

	return target.Name
}
