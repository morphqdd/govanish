# govanish — a maximally strict Go linter

Date: 2026-09-15
Status: approved design, not yet implemented

## Purpose

A standalone linter that enforces one rigid code style for Go, with no
opt-outs. Where existing linters offer a menu, `govanish` offers a
verdict: the code either conforms or it does not.

The design goal is not to catch every possible bug. It is to remove
stylistic choice from Go code entirely, so that any two files written
under `govanish` look like the same person wrote them.

### Naming conflict

`yegor256/govanish` is an existing, unrelated linter that flags code the
compiler optimized away. The name collides if this tool is ever
published. Accepted for now; rename is a one-commit change since the
name appears only in the module path and `cmd/` directory.

## Decisions

These were settled during design and are not open in v1.

| Decision | Choice | Why |
|---|---|---|
| Form | Standalone binary on `go/analysis` | Rules become real `*analysis.Analyzer`s, so `analysistest` and golden files work with no custom harness. |
| Runner | Custom, on `packages.Load` | `multichecker` dictates a per-rule flag CLI, which contradicts zero-config. |
| Configuration | None | No config file, no per-rule flags, no inline suppression. Thresholds are constants in the rule source. |
| Autofix | Not in v1 | Most rules (complexity, architecture, types) have no mechanical fix. A partial `-fix` teaches the wrong habit. |
| Scope of v1 | All seven rule categories | ~45 rules. |
| Caching | None in v1 | `packages.Load` reuses the Go build cache for type-checking. A rule-result cache is a v2 concern, and only once the tool is measurably slow. |

The absence of inline suppression is deliberate and interacts with the
`comments-doc-only` rule: a `//nolint` directive would itself be an
illegal comment. The two rules are consistent only because neither
exists.

## Architecture

```
go.mod                      module govanish
cmd/govanish/main.go        argv parsing, runner invocation, exit code
internal/runner/            packages.Load -> analyzer execution -> diagnostic collection
internal/registry/          flat list of all rules, single registration point
internal/rules/
  style/                    naming, imports, comments, layout
  size/                     length and complexity limits
  errs/                     error handling
  arch/                     structural prohibitions
  conc/                     concurrency
  tests/                    tests and documentation
  api/                      exported surface and types
  <category>/testdata/src/<rule>/   analysistest fixtures
internal/astutil/           shared helpers: positions, traversal, predicates
```

### Rule model

Each rule is one file exporting one analyzer:

```go
var Analyzer = &analysis.Analyzer{
    Name:     "funclen",
    Doc:      "function exceeds 40 lines",
    Requires: []*analysis.Analyzer{inspect.Analyzer},
    Run:      run,
}
```

Three invariants hold the system together:

1. **No rule declares flags.** `Analyzer.Flags` is always empty. Changing
   a threshold means editing the linter, not a project config. A
   registry test enforces this.
2. **One rule = one file + one test + one testdata directory.** No rule
   reads another rule's state.
3. **Every AST-walking rule lists `inspect.Analyzer` in `Requires`,** so
   each package is traversed once rather than once per rule.

The registry is a flat `[]*analysis.Analyzer`. Execution order is
irrelevant to every rule except `min-export`; diagnostics are sorted at
print time.

## Rule set

Thresholds appear inline. Rules cut from v1 are listed at the end.

### style

- `name-length` — identifiers are at least 3 characters. Whitelist:
  `err ok ctx id db tx mu wg fn`. Receivers are 1-2 characters and use
  the same letters across every method on a type. This bans `i` as a
  loop variable, which is intended.
- `initialisms` — `Url` -> `URL`, `Http` -> `HTTP`, standard list.
- `no-underscore-names` — except `Test_` prefixes in `_test.go`.
- `no-stutter` — `user.UserService` is an error.
- `import-groups` — exactly three groups (stdlib, external, local),
  each sorted; no aliases unless the import genuinely collides; no
  dot-imports; no blank imports outside `main`.
- `decl-order` — const, then var, then type, then func; exported before
  unexported; methods follow the type they belong to.
- `comments-doc-only` — a comment is legal only as a doc comment on a
  declaration. It starts with the identifier's name and ends with a
  period. Inline, trailing, and in-function comments are errors.
  `TODO` and `FIXME` are errors. Build constraints and `//go:`
  directives are exempt.
- `no-else-after-return` — early return only.
- `line-length` — 100 columns.

### size

`func-len` 40 body lines; `file-len` 400 lines; `cyclomatic` 8;
`nesting-depth` 3; `param-count` 4; `return-count` 3;
`struct-fields` 10.

### errs

- `errcheck` — every returned error is used. `_ = f()` on an
  error-returning call is an error.
- `wrap` — `fmt.Errorf` wrapping an error uses `%w`. Message is
  lowercase with no trailing punctuation.
- `error-shape` — error is the last result, named `err` when results
  are named.
- `errors-is` — errors are compared with `errors.Is` or `errors.As`,
  never `==` or string matching.
- `no-panic` — `panic`, `log.Fatal*`, and `os.Exit` are banned outside
  `main`.
- `no-naked-return`
- `no-empty-err-branch` — `if err != nil {}` with an empty or
  comment-only body.

### arch

- `no-globals` — no package-level `var`. `const` is permitted.
- `no-init` — `func init` is banned.
- `no-any` — `interface{}` and `any` are banned in signatures and
  struct fields.
- `banned-imports` — `unsafe`, `reflect`, `math/rand`, and `log` (use
  `log/slog`); `fmt.Print*` outside `cmd`.
- `no-shadow` — no shadowing of any variable, `err` included.

### conc

- `ctx-first` — `context.Context` is the first parameter, named `ctx`,
  and is never stored in a struct field.
- `ctx-origin` — `context.Background()` and `context.TODO()` appear
  only in `main` and in tests.
- `defer-unlock` — `defer mu.Unlock()` is the statement immediately
  following `mu.Lock()`.
- `sized-chan` — `make(chan T)` without a size is an error.
- `go-needs-owner` — a `go` statement requires a `sync.WaitGroup` or
  `errgroup.Group` in scope.
- `no-time-after-select` — `time.After` in a `select` leaks; use
  `time.NewTimer`.

### tests

- `exported-doc` — every exported identifier has a doc comment starting
  with its own name.
- `pkg-doc` — every package has a `doc.go` carrying the package comment.
- `test-shape` — `TestXxx` naming, subtests via `t.Run`.
- `no-skip` — `t.Skip` is banned.
- `test-parallel` — `t.Parallel()` is required in every test.

### api

- `no-iface-return` — exported functions do not return interface types.
  `error` is the exception.
- `no-bare-prim` — exported signatures do not use bare `string`, `int`,
  or `float64`; a domain type is required. `error`, `context.Context`,
  and `bool` results are exempt. This is the loudest rule in the set.
- `no-bool-param` — boolean parameters are banned.
- `keyed-literals` — composite literals name their fields.
- `min-export` — an exported identifier never used outside its own
  package is an error. Requires whole-program analysis via
  `analysis.Fact` and topological package ordering in the runner.

### Cut from v1

- `layers` — import-direction rules ("who may import whom"). Every
  correct answer is project-specific, and zero-config leaves nowhere to
  express it without hardcoding a directory convention. Revisit once
  there is real code to observe.

## CLI and output

```
govanish ./...
govanish ./internal/... ./cmd/...
govanish                      # defaults to ./...
```

Three flags, none affecting which rules run: `-format text|json`,
`-version`, `-h`.

### Runner pipeline

1. `packages.Load` with
   `NeedTypes|NeedSyntax|NeedTypesInfo|NeedDeps|NeedImports` over the
   given patterns. Load errors (syntax errors, missing dependencies)
   are reported and exit 2. Linting unparseable code produces
   meaningless results.
2. Packages are ordered topologically, because `min-export` emits facts
   its importers consume. No other rule depends on order.
3. Per package: `inspect.Analyzer` runs once, then every rule runs
   against that shared inspector. Packages are processed concurrently
   up to `GOMAXPROCS`; rules within a package run sequentially, so
   diagnostics accumulate without locking.
4. Diagnostics are collected, sorted by file, line, column, then rule
   id, printed, and the exit code is set.

### Text format

```
internal/user/service.go:42:6: [func-len] function CreateUser is 61 lines, limit is 40
internal/user/service.go:58:2: [comments-doc-only] comment is not a doc comment
```

### JSON format

A flat array of objects with `file`, `line`, `col`, `end_line`,
`end_col`, `rule`, and `message`. Flat rather than grouped by package,
so `jq` filters work without traversal.

### Exit codes

`0` clean; `1` diagnostics found; `2` load or internal failure. CI can
distinguish bad code from a broken linter.

### Skipped files

Generated files (`// Code generated ... DO NOT EDIT.`) and `vendor/`
are not linted.

## Testing

Every rule is tested through `analysistest`, which is the reason rules
are real analyzers.

```
internal/rules/size/testdata/src/funclen/funclen.go
internal/rules/size/funclen_test.go
```

Expectations live in the fixture:

```go
func CreateUser() { // want `function CreateUser is 61 lines, limit is 40`
```

and the test is a single call:

```go
func TestFuncLen(t *testing.T) {
    analysistest.Run(t, analysistest.TestData(), size.Analyzer)
}
```

Each rule's testdata covers at least three cases: code that triggers
the rule, conforming code that must stay silent (the false-positive
guard), and the boundary condition (exactly 40 lines passes, 41 fails).

Four suite-level tests sit above the rules:

- **registry** — names are unique, every rule has a non-empty `Doc`, no
  rule declares flags. Catches copy-paste registration errors.
- **runner** — a fixture module with known violations, asserting exact
  diagnostic count, sort order, and exit code.
- **self-lint** — `govanish` runs over its own source and reports
  nothing. This is the real acceptance test for whether the rule set is
  livable.
- **golden output** — text and JSON output compared byte-for-byte
  against checked-in files.

Development is test-first per rule: write the testdata with `// want`
markers, watch the test fail because the analyzer reports nothing,
implement, reach green. `no-bare-prim` and `min-export` get their
testdata written before any implementation, since those two are where
false positives will concentrate.

Self-lint is bootstrapped rather than retrofitted: it passes from the
first rule onward, so the linter's own source is written under the
rules as they land.

## Build order

Each milestone ends with a working binary and passing tests.

**M1 — skeleton.** `go.mod`, `cmd/govanish/main.go`, `internal/runner`,
`internal/registry`, both printers, exit codes. One trivial rule
(`no-init`) proves the pipeline end to end. The runner fixture test and
golden-output test land here. Done when `govanish ./...` finds a
planted `func init` and exits 1.

**M2 — AST-only rules.** Everything needing no type information: all of
`size`, plus `no-else-after-return`, `no-underscore-names`,
`initialisms`, `decl-order`, `import-groups`, `comments-doc-only`,
`no-naked-return`, `no-skip`, `test-parallel`,
`test-shape`, `pkg-doc`. Most of the rule count, least difficulty per
rule. `astutil` grows here and nowhere else.

**M3 — type-aware rules.** Requiring `TypesInfo`: `errcheck`,
`errors-is`, `wrap`, `error-shape`, `no-empty-err-branch`, `no-panic`,
`no-any`, `banned-imports`, `no-globals`, `exported-doc`,
`no-bool-param`, `keyed-literals` (distinguishing a struct literal from a
slice or map literal requires type information). Error handling is the category most prone to false
positives, so testdata here is larger than elsewhere.

**M4 — concurrency and shadowing.** `ctx-first`, `ctx-origin`,
`defer-unlock`, `sized-chan`, `go-needs-owner`, `no-time-after-select`,
`no-shadow`. Grouped separately because these are flow-sensitive rather
than shape-sensitive; `go-needs-owner` and `no-shadow` need scope
walking no earlier rule requires.

**M5 — the hard rules.** `no-iface-return`, `no-bare-prim`, and
`min-export`. `min-export` introduces `analysis.Fact` plumbing and the
topological ordering in the runner, making this the only milestone that
modifies the runner after M1, and the only rule that cannot be tested
with a single-package fixture.

**M6 — `name-length` and self-lint hardening.** `name-length` lands
last deliberately: it churns more identifiers in our own source than
any other rule, so it arrives when there is the most code to rename in
one pass. Then a full pass bringing self-lint green under all rules,
which is where wrong thresholds surface.

Only M1 and M5 touch `internal/runner`. M2, M3, M4, and M6 add files
under `internal/rules/` alone, so they parallelize cleanly.

## Expected outcomes

Self-lint at M6 will fail loudly. That is the design working: it is the
only honest signal about whether `no-bare-prim` and `name-length` are
livable, and it surfaces on the linter's own code rather than on a real
project.

Running `govanish` against an existing codebase will produce thousands
of diagnostics with no way to silence them. This is inherent to
zero-config and is the accepted cost of the design. The tool is
intended for code written under it from the start.
