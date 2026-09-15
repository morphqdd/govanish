// Package registry is the single place where every govanish rule is
// listed. A rule that is not here does not run.
package registry

import (
	"golang.org/x/tools/go/analysis"

	"govanish/internal/rules/api"
	"govanish/internal/rules/arch"
	"govanish/internal/rules/conc"
	"govanish/internal/rules/errs"
	"govanish/internal/rules/size"
	"govanish/internal/rules/style"
	"govanish/internal/rules/tests"
)

// Rule pairs an analyzer with the identifier govanish prints for it.
// The two differ because analysis.Validate requires analyzer names to be
// valid Go identifiers, while diagnostics read better hyphenated.
type Rule struct {
	ID       string
	Analyzer *analysis.Analyzer
}

// Set is one construction of every rule. Analyzers are compared by
// pointer by the analysis driver, so a Set must be built once and then
// used throughout a run rather than rebuilt for each question asked of
// it.
type Set struct {
	rules []Rule
}

// New constructs every rule govanish runs.
func New() Set {
	var rules []Rule

	for _, group := range [][]Rule{
		structureRules(),
		sizeRules(),
		styleRules(),
		errorsRules(),
		apiRules(),
		concurrencyRules(),
		testsRules(),
	} {
		rules = append(rules, group...)
	}

	return Set{rules: rules}
}

func structureRules() []Rule {
	return []Rule{
		{ID: "no-init", Analyzer: arch.NoInit()},
		{ID: "no-globals", Analyzer: arch.NoGlobals()},
		{ID: "no-any", Analyzer: arch.NoAny()},
		{ID: "banned-imports", Analyzer: arch.BannedImports()},
		{ID: "no-shadow", Analyzer: arch.NoShadow()},
	}
}

func sizeRules() []Rule {
	return []Rule{
		{ID: "func-len", Analyzer: size.FuncLen()},
		{ID: "file-len", Analyzer: size.FileLen()},
		{ID: "line-length", Analyzer: size.LineLength()},
		{ID: "cyclomatic", Analyzer: size.Cyclo()},
		{ID: "nesting-depth", Analyzer: size.Nesting()},
		{ID: "param-count", Analyzer: size.ParamCount()},
		{ID: "return-count", Analyzer: size.ReturnCount()},
		{ID: "struct-fields", Analyzer: size.StructFields()},
		{ID: "file-complexity", Analyzer: size.FileComplexity()},
	}
}

func styleRules() []Rule {
	return []Rule{
		{ID: "no-else-after-return", Analyzer: style.NoElse()},
		{ID: "no-naked-return", Analyzer: style.NoNakedReturn()},
		{ID: "no-underscore-names", Analyzer: style.Underscores()},
		{ID: "initialisms", Analyzer: style.Initialisms()},
		{ID: "comments-doc-only", Analyzer: style.Comments()},
		{ID: "decl-order", Analyzer: style.DeclOrder()},
		{ID: "import-groups", Analyzer: style.ImportGroups()},
	}
}

func errorsRules() []Rule {
	return []Rule{
		{ID: "errcheck", Analyzer: errs.ErrCheck()},
		{ID: "no-empty-err-branch", Analyzer: errs.EmptyBranch()},
		{ID: "errors-is", Analyzer: errs.ErrorsIs()},
		{ID: "no-panic", Analyzer: errs.NoPanic()},
		{ID: "error-shape", Analyzer: errs.ErrShape()},
		{ID: "wrap", Analyzer: errs.Wrap()},
	}
}

func apiRules() []Rule {
	return []Rule{
		{ID: "no-bool-param", Analyzer: api.NoBoolParam()},
		{ID: "keyed-literals", Analyzer: api.KeyedLiterals()},
	}
}

func concurrencyRules() []Rule {
	return []Rule{
		{ID: "ctx-first", Analyzer: conc.CtxFirst()},
		{ID: "ctx-origin", Analyzer: conc.CtxOrigin()},
		{ID: "defer-unlock", Analyzer: conc.DeferUnlock()},
		{ID: "sized-chan", Analyzer: conc.SizedChan()},
		{ID: "go-needs-owner", Analyzer: conc.GoOwner()},
		{ID: "no-time-after-select", Analyzer: conc.TimeAfter()},
	}
}

func testsRules() []Rule {
	return []Rule{
		{ID: "test-shape", Analyzer: tests.Shape()},
		{ID: "no-skip", Analyzer: tests.NoSkip()},
		{ID: "test-parallel", Analyzer: tests.Parallel()},
		{ID: "pkg-doc", Analyzer: tests.PkgDoc()},
		{ID: "exported-doc", Analyzer: tests.ExportedDoc()},
	}
}

// Rules returns every rule in the set.
func (s Set) Rules() []Rule {
	out := make([]Rule, len(s.rules))
	copy(out, s.rules)

	return out
}

// Analyzers returns the analyzer of every rule in the set.
func (s Set) Analyzers() []*analysis.Analyzer {
	out := make([]*analysis.Analyzer, 0, len(s.rules))
	for _, rule := range s.rules {
		out = append(out, rule.Analyzer)
	}

	return out
}

// IDOf returns the printable identifier of an analyzer in this set, or
// the analyzer's own name if it belongs to another set.
func (s Set) IDOf(target *analysis.Analyzer) string {
	for _, rule := range s.rules {
		if rule.Analyzer == target {
			return rule.ID
		}
	}

	return target.Name
}
