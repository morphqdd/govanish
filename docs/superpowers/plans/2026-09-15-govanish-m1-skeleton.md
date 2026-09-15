# govanish M1 (Skeleton) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A working `govanish` binary that loads Go packages, runs a rule
registry against them, and prints sorted diagnostics in text or JSON with
correct exit codes — proven end to end by one real rule, `no-init`.

**Architecture:** Rules are `*analysis.Analyzer` values. A registry pairs
each analyzer with a hyphenated display ID. The runner calls
`packages.Load`, hands the packages to `checker.Analyze` from x/tools
(which handles dependency order, fact propagation and parallelism), then
flattens the resulting action graph into sorted findings. A report package
owns formatting. `main` owns only argv and the exit code.

**Tech Stack:** Go 1.26, `golang.org/x/tools` v0.44.0
(`go/packages`, `go/analysis`, `go/analysis/checker`,
`go/analysis/analysistest`, `go/ast/inspector`).

**Spec:** `docs/superpowers/specs/2026-09-15-govanish-design.md`

## Global Constraints

- Module path is `govanish`. Binary is `cmd/govanish`.
- Go 1.26. `golang.org/x/tools v0.44.0` is the only non-stdlib dependency.
- No rule may set `Analyzer.Flags`. Enforced by a registry test.
- `Analyzer.Name` must be a valid Go identifier — `analysis.Validate`
  rejects hyphens. The hyphenated form used in output lives in
  `registry.Rule.ID`.
- Every AST-walking rule lists `inspect.Analyzer` in `Requires`.
- Thresholds are constants in the rule's own file. No config, no flags,
  no inline suppression.
- Exit codes: `0` clean, `1` diagnostics found, `2` load or internal
  failure.
- Text output line format:
  `file:line:col: [rule-id] message`
- The project lints itself. Since the rule set is not yet written, M1's
  self-lint test asserts only against the rules that exist.

### Deviation from the spec

The spec describes hand-written topological ordering and per-package
concurrency in the runner. `golang.org/x/tools/go/analysis/checker`
provides `Analyze(analyzers, pkgs, opts) (*Graph, error)`, which does
exactly that, correctly, including fact import/export needed by
`min-export` in M5. This plan uses it. The spec's pipeline steps 2 and 3
are satisfied by that call; steps 1 and 4 remain ours.

---

### Task 1: The `no-init` rule

The first rule exists to prove the pipeline, so it is the simplest rule in
the whole set: no type information, no configuration, one AST node kind.

**Files:**
- Create: `go.mod`
- Create: `internal/rules/arch/doc.go`
- Create: `internal/rules/arch/noinit.go`
- Test: `internal/rules/arch/noinit_test.go`
- Test: `internal/rules/arch/testdata/src/noinit/noinit.go`

**Interfaces:**
- Consumes: nothing.
- Produces: `arch.NoInit` — a `*analysis.Analyzer` with
  `Name: "noinit"`.

- [ ] **Step 1: Initialize the module**

```bash
cd /home/morphe/utils
go mod init govanish
go get golang.org/x/tools@v0.44.0
```

- [ ] **Step 2: Write the test fixture**

`analysistest` reads expectations from `// want` comments in the fixture.
The regexp must match the diagnostic message.

Create `internal/rules/arch/testdata/src/noinit/noinit.go`:

```go
package noinit

import "os"

func init() { // want `func init is banned`
	_ = os.Getenv("HOME")
}

func NewThing() *Thing {
	return &Thing{}
}

type Thing struct{}

func (t *Thing) init() {}
```

Three cases in one file: the violation, a conforming function that must
stay silent, and the boundary case that catches the obvious bug — a
*method* named `init` is legal Go and must not be reported.

- [ ] **Step 3: Write the failing test**

Create `internal/rules/arch/noinit_test.go`:

```go
package arch_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"govanish/internal/rules/arch"
)

func TestNoInit(t *testing.T) {
	t.Parallel()

	analysistest.Run(t, analysistest.TestData(), arch.NoInit, "noinit")
}
```

- [ ] **Step 4: Run the test to verify it fails**

Run: `go test ./internal/rules/arch/`
Expected: FAIL — `undefined: arch.NoInit`.

- [ ] **Step 5: Write the package doc**

Create `internal/rules/arch/doc.go`:

```go
// Package arch holds rules governing the structural shape of a package:
// what may exist at package level, what may be imported, and what may run
// implicitly.
package arch
```

- [ ] **Step 6: Implement the rule**

Create `internal/rules/arch/noinit.go`:

```go
package arch

import (
	"errors"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// errMissingInspector reports that the driver did not supply the shared
// inspector, which means the analyzer was run without its Requires.
var errMissingInspector = errors.New("inspect analyzer result missing")

// NoInit reports package-level init functions.
var NoInit = &analysis.Analyzer{
	Name:     "noinit",
	Doc:      "func init is banned because it hides work behind package loading",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runNoInit,
}

func runNoInit(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, errMissingInspector
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		decl, ok := node.(*ast.FuncDecl)
		if !ok {
			return
		}
		if decl.Recv != nil || decl.Name.Name != "init" {
			return
		}
		pass.Reportf(decl.Pos(), "func init is banned; do the work in an explicit constructor called from main")
	})

	return nil, nil
}
```

The `decl.Recv != nil` guard is the one that matters: without it the rule
reports `func (t *Thing) init()`, which is ordinary legal code.

- [ ] **Step 7: Run the test to verify it passes**

Run: `go test ./internal/rules/arch/`
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add go.mod go.sum internal/rules/arch/
git commit -m "feat(rules): add no-init rule"
```

---

### Task 2: The rule registry

**Files:**
- Create: `internal/registry/registry.go`
- Test: `internal/registry/registry_test.go`

**Interfaces:**
- Consumes: `arch.NoInit`.
- Produces:
  - `type Rule struct { ID string; Analyzer *analysis.Analyzer }`
  - `func All() []Rule` — every rule in the linter.
  - `func Analyzers() []*analysis.Analyzer` — the analyzers alone, for
    passing to `checker.Analyze`.
  - `func IDOf(a *analysis.Analyzer) string` — display ID for an
    analyzer, used by the runner when building findings.

- [ ] **Step 1: Write the failing test**

Create `internal/registry/registry_test.go`:

```go
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
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/registry/`
Expected: FAIL — package `govanish/internal/registry` does not exist.

- [ ] **Step 3: Implement the registry**

Create `internal/registry/registry.go`:

```go
// Package registry is the single place where every govanish rule is
// listed. A rule that is not here does not run.
package registry

import (
	"golang.org/x/tools/go/analysis"

	"govanish/internal/rules/arch"
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
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/registry/`
Expected: PASS, five tests.

- [ ] **Step 5: Commit**

```bash
git add internal/registry/
git commit -m "feat(registry): add rule registry with invariant tests"
```

---

### Task 3: Findings and formatting

Formatting is built before the runner so the runner has something concrete
to produce.

**Files:**
- Create: `internal/report/report.go`
- Test: `internal/report/report_test.go`
- Test: `internal/report/testdata/text.golden`
- Test: `internal/report/testdata/json.golden`

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `type Finding struct { File string; Line, Col, EndLine, EndCol int; Rule, Message string }`
    with JSON tags `file`, `line`, `col`, `end_line`, `end_col`, `rule`,
    `message`.
  - `func Sort(findings []Finding)` — in place, by file, line, col, rule.
  - `func Text(w io.Writer, findings []Finding) error`
  - `func JSON(w io.Writer, findings []Finding) error`

- [ ] **Step 1: Write the golden files**

Create `internal/report/testdata/text.golden`:

```
internal/user/service.go:12:1: [no-init] func init is banned
internal/user/service.go:42:6: [func-len] function CreateUser is 61 lines, limit is 40
internal/user/service.go:42:6: [no-bare-prim] parameter name is a bare string
```

Create `internal/report/testdata/json.golden`:

```json
[
  {
    "file": "internal/user/service.go",
    "line": 12,
    "col": 1,
    "end_line": 12,
    "end_col": 10,
    "rule": "no-init",
    "message": "func init is banned"
  },
  {
    "file": "internal/user/service.go",
    "line": 42,
    "col": 6,
    "end_line": 42,
    "end_col": 16,
    "rule": "func-len",
    "message": "function CreateUser is 61 lines, limit is 40"
  },
  {
    "file": "internal/user/service.go",
    "line": 42,
    "col": 6,
    "end_line": 42,
    "end_col": 16,
    "rule": "no-bare-prim",
    "message": "parameter name is a bare string"
  }
]
```

- [ ] **Step 2: Write the failing test**

Create `internal/report/report_test.go`:

```go
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
```

The third test exists because `json.Marshal` renders a nil slice as
`null`, which breaks every consumer that expects to iterate an array.

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./internal/report/`
Expected: FAIL — package `govanish/internal/report` does not exist.

- [ ] **Step 4: Implement the report package**

Create `internal/report/report.go`:

```go
// Package report renders findings in the formats govanish supports.
package report

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"slices"
)

// Finding is one rule violation at one source position.
type Finding struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Col     int    `json:"col"`
	EndLine int    `json:"end_line"`
	EndCol  int    `json:"end_col"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// Sort orders findings by file, line, column, then rule identifier, so
// that output is stable across runs regardless of analysis order.
func Sort(findings []Finding) {
	slices.SortFunc(findings, func(left, right Finding) int {
		return cmp.Or(
			cmp.Compare(left.File, right.File),
			cmp.Compare(left.Line, right.Line),
			cmp.Compare(left.Col, right.Col),
			cmp.Compare(left.Rule, right.Rule),
		)
	})
}

// Text writes findings one per line as file:line:col: [rule] message.
func Text(writer io.Writer, findings []Finding) error {
	for _, finding := range findings {
		_, err := fmt.Fprintf(writer, "%s:%d:%d: [%s] %s\n",
			finding.File, finding.Line, finding.Col, finding.Rule, finding.Message)
		if err != nil {
			return fmt.Errorf("write finding: %w", err)
		}
	}

	return nil
}

// JSON writes findings as a flat array, empty rather than null when there
// are none.
func JSON(writer io.Writer, findings []Finding) error {
	if findings == nil {
		findings = []Finding{}
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(findings); err != nil {
		return fmt.Errorf("encode findings: %w", err)
	}

	return nil
}
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/report/`
Expected: PASS, three tests. If the JSON golden mismatches on trailing
whitespace, fix the golden file to match `encoding/json` output rather
than changing the encoder — `Encode` appends a newline, and the golden
file must end with one.

- [ ] **Step 6: Commit**

```bash
git add internal/report/
git commit -m "feat(report): add finding type with text and json output"
```

---

### Task 4: The runner

**Files:**
- Create: `internal/runner/runner.go`
- Test: `internal/runner/runner_test.go`
- Test: `internal/runner/testdata/fixture/go.mod`
- Test: `internal/runner/testdata/fixture/bad/bad.go`
- Test: `internal/runner/testdata/fixture/good/good.go`
- Test: `internal/runner/testdata/fixture/gen/gen.go`

**Interfaces:**
- Consumes: `registry.Analyzers()`, `registry.IDOf`, `report.Finding`,
  `report.Sort`.
- Produces: `func Run(dir string, patterns []string) ([]report.Finding, error)`
  — `dir` is the working directory to load from (empty means the process
  working directory), `patterns` are `go list` patterns. Returns findings
  sorted, and a non-nil error only for load or driver failure, never for
  the presence of findings.

- [ ] **Step 1: Write the fixture module**

Create `internal/runner/testdata/fixture/go.mod`:

```
module fixture

go 1.26
```

Create `internal/runner/testdata/fixture/bad/bad.go`:

```go
package bad

var loaded bool

func init() {
	loaded = true
}

func Loaded() bool {
	return loaded
}
```

Create `internal/runner/testdata/fixture/good/good.go`:

```go
package good

func Answer() int {
	return 42
}
```

Create `internal/runner/testdata/fixture/gen/gen.go`:

```go
// Code generated by fixture. DO NOT EDIT.

package gen

func init() {
	_ = 1
}
```

The `gen` package is the one that matters: it violates `no-init` and must
still produce no finding, because generated files are not linted.

- [ ] **Step 2: Write the failing test**

Create `internal/runner/runner_test.go`:

```go
package runner_test

import (
	"strings"
	"testing"

	"govanish/internal/runner"
)

const fixtureDir = "testdata/fixture"

func TestRunReportsViolations(t *testing.T) {
	t.Parallel()

	findings, err := runner.Run(fixtureDir, []string{"./..."})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1: %+v", len(findings), findings)
	}

	finding := findings[0]
	if finding.Rule != "no-init" {
		t.Errorf("rule = %q, want %q", finding.Rule, "no-init")
	}
	if !strings.HasSuffix(finding.File, "bad/bad.go") {
		t.Errorf("file = %q, want it to end with bad/bad.go", finding.File)
	}
	if finding.Line != 5 {
		t.Errorf("line = %d, want 5", finding.Line)
	}
}

func TestRunSkipsGeneratedFiles(t *testing.T) {
	t.Parallel()

	findings, err := runner.Run(fixtureDir, []string{"./gen/..."})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if len(findings) != 0 {
		t.Errorf("got %d findings in generated code, want 0: %+v", len(findings), findings)
	}
}

func TestRunReturnsErrorForUnloadablePackages(t *testing.T) {
	t.Parallel()

	_, err := runner.Run(fixtureDir, []string{"./does-not-exist/..."})
	if err == nil {
		t.Fatal("want an error for a pattern matching no packages, got nil")
	}
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./internal/runner/`
Expected: FAIL — package `govanish/internal/runner` does not exist.

- [ ] **Step 4: Implement the runner**

Create `internal/runner/runner.go`:

```go
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

	return findings, nil
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

func finding(action *checker.Action, diagnostic analysis.Diagnostic, position token.Position) report.Finding {
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
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/runner/`
Expected: PASS, three tests.

If `TestRunReportsViolations` reports two findings rather than one, the
cause is `Tests: true` loading the same package twice (once plain, once
with its test files). The fixture has no test files, so this should not
happen; if it does, deduplicate findings by file, line, col and rule in
`collect` before returning.

- [ ] **Step 6: Commit**

```bash
git add internal/runner/
git commit -m "feat(runner): load packages and collect rule findings"
```

---

### Task 5: The command

**Files:**
- Create: `cmd/govanish/main.go`
- Test: `cmd/govanish/main_test.go`

**Interfaces:**
- Consumes: `runner.Run`, `report.Text`, `report.JSON`.
- Produces: the `govanish` binary. No exported Go API.

- [ ] **Step 1: Write the failing test**

Create `cmd/govanish/main_test.go`:

```go
package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunCleanPackageExitsZero(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer

	code := run([]string{"-format", "text", "./good/..."}, "../../internal/runner/testdata/fixture", &out, &errOut)

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

	code := run([]string{"-format", "text", "./bad/..."}, "../../internal/runner/testdata/fixture", &out, &errOut)

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

	code := run([]string{"./does-not-exist/..."}, "../../internal/runner/testdata/fixture", &out, &errOut)

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

	code := run([]string{"-format", "xml", "./good/..."}, "../../internal/runner/testdata/fixture", &out, &errOut)

	if code != exitFailure {
		t.Errorf("exit code = %d, want %d", code, exitFailure)
	}
}
```

The test drives `run`, not `main`, so exit codes are values to assert on
rather than process termination.

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./cmd/govanish/`
Expected: FAIL — `undefined: run`, `undefined: exitClean`.

- [ ] **Step 3: Implement the command**

Create `cmd/govanish/main.go`:

```go
// Command govanish lints Go packages against a fixed, non-negotiable
// style. It has no configuration file and no suppression directives: code
// either conforms or it does not.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"govanish/internal/report"
	"govanish/internal/runner"
)

const (
	exitClean    = 0
	exitFindings = 1
	exitFailure  = 2
)

// version is overridden at build time with -ldflags.
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], "", os.Stdout, os.Stderr))
}

func run(args []string, dir string, out, errOut io.Writer) int {
	flags := flag.NewFlagSet("govanish", flag.ContinueOnError)
	flags.SetOutput(errOut)

	format := flags.String("format", "text", "output format: text or json")
	showVersion := flags.Bool("version", false, "print the version and exit")

	if err := flags.Parse(args); err != nil {
		return exitFailure
	}

	if *showVersion {
		fmt.Fprintf(out, "govanish %s\n", version)

		return exitClean
	}

	render, err := renderer(*format)
	if err != nil {
		fmt.Fprintf(errOut, "govanish: %v\n", err)

		return exitFailure
	}

	patterns := flags.Args()
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}

	findings, err := runner.Run(dir, patterns)
	if err != nil {
		fmt.Fprintf(errOut, "govanish: %v\n", err)

		return exitFailure
	}

	if err := render(out, findings); err != nil {
		fmt.Fprintf(errOut, "govanish: %v\n", err)

		return exitFailure
	}

	if len(findings) > 0 {
		return exitFindings
	}

	return exitClean
}

func renderer(format string) (func(io.Writer, []report.Finding) error, error) {
	switch format {
	case "text":
		return report.Text, nil
	case "json":
		return report.JSON, nil
	default:
		return nil, fmt.Errorf("unknown format %q: want text or json", format)
	}
}
```

Note that `report.JSON` prints `[]` for a clean run, so
`TestRunCleanPackageExitsZero` uses `-format text` deliberately.

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./cmd/govanish/`
Expected: PASS, four tests.

- [ ] **Step 5: Verify the binary end to end**

```bash
go build -o /tmp/govanish ./cmd/govanish
cd internal/runner/testdata/fixture && /tmp/govanish ./... ; echo "exit=$?"
```

Expected: one `[no-init]` line pointing at `bad/bad.go:5:1`, and
`exit=1`.

- [ ] **Step 6: Commit**

```bash
git add cmd/govanish/
git commit -m "feat(cmd): add govanish command with text and json output"
```

---

### Task 6: Self-lint

The linter runs on its own source and stays clean from here on. Every
later milestone adds rules; this test is what forces the linter's own code
to satisfy them as they land, rather than accumulating a cleanup debt to
be paid at M6.

**Files:**
- Create: `selflint_test.go` (module root)
- Create: `Makefile`
- Create: `README.md`

**Interfaces:**
- Consumes: `runner.Run`.
- Produces: nothing importable.

- [ ] **Step 1: Write the test**

Create `selflint_test.go` at the module root:

```go
package govanish_test

import (
	"bytes"
	"testing"

	"govanish/internal/report"
	"govanish/internal/runner"
)

// TestSelfLint holds govanish to its own rules. It is expected to fail
// whenever a newly added rule finds a violation in govanish's own source;
// the fix is always to change govanish, never to weaken the rule.
func TestSelfLint(t *testing.T) {
	t.Parallel()

	findings, err := runner.Run(".", []string{"./..."})
	if err != nil {
		t.Fatalf("lint own source: %v", err)
	}

	if len(findings) == 0 {
		return
	}

	var buf bytes.Buffer
	if err := report.Text(&buf, findings); err != nil {
		t.Fatalf("render findings: %v", err)
	}

	t.Errorf("govanish does not satisfy its own rules:\n%s", buf.String())
}
```

- [ ] **Step 2: Run the test**

Run: `go test .`
Expected: PASS. Only `no-init` exists, and govanish has no `init`
functions.

Note: this test loads `./...` from the module root, which includes the
`testdata` fixture directories. `go list` excludes any directory named
`testdata`, so the fixtures are not linted. If findings from fixture files
appear, that assumption is wrong and the pattern needs narrowing to
`./cmd/...` and `./internal/...` explicitly.

- [ ] **Step 3: Write the Makefile**

Create `Makefile`:

```make
.PHONY: build test lint all

all: test lint build

build:
	go build -o bin/govanish ./cmd/govanish

test:
	go test ./...

lint: build
	./bin/govanish ./cmd/... ./internal/...
```

- [ ] **Step 4: Write the README**

Create `README.md`:

```markdown
# govanish

A deliberately inflexible Go linter. There is no configuration file, no
per-rule flag, and no way to suppress a diagnostic in source. Code either
conforms to the style or it does not.

## Usage

    govanish ./...
    govanish -format json ./internal/...

Exit codes: `0` clean, `1` findings, `2` the linter could not run.

## Why no suppression

Every escape hatch becomes the default path under deadline pressure. The
rules are strict enough to be worth arguing about, so the argument happens
once, in this repository, rather than in every code review.

## Status

Under construction. See `docs/superpowers/specs/` for the design and
`docs/superpowers/plans/` for the build order.
```

- [ ] **Step 5: Verify everything passes together**

Run: `make all`
Expected: tests pass, binary builds, self-lint prints nothing.

- [ ] **Step 6: Commit**

```bash
git add selflint_test.go Makefile README.md
git commit -m "test: lint govanish with its own rules"
```

---

## Done when

- `make all` is green.
- `govanish ./...` in the fixture module reports the planted `func init`
  and exits 1.
- `govanish -format json` emits a flat array, `[]` when clean.
- Generated files produce no findings.
- Adding a rule in M2 requires touching exactly two places: the rule's own
  new file, and one line in `internal/registry/registry.go`.

## Next

M2 (AST-only rules) gets its own plan once this one is green.
