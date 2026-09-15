# govanish

A deliberately inflexible Go linter. There is no configuration file, no
per-rule flag, and no way to suppress a diagnostic in source. Code either
conforms to the style or it does not.

44 rules, all of them always on.

## Usage

```
govanish ./...
govanish -format json ./internal/...
```

Exit codes: `0` clean, `1` findings, `2` the linter could not run.

Output is one finding per line:

```
internal/user/service.go:42:6: [func-len] function CreateUser is 61 lines, limit is 40
```

## The rules

**Size** — `func-len` 40 lines · `file-len` 400 · `line-length` 100 ·
`cyclomatic` 8 · `nesting-depth` 3 · `param-count` 4 · `return-count` 3 ·
`struct-fields` 10 · `file-complexity` 40 (the total branching of a file,
which catches thirty individually simple functions in one place)

**Style** — `name-length` (3 characters, short whitelist, receivers of 1–2) ·
`initialisms` (`Url` → `URL`) · `no-underscore-names` · `decl-order` ·
`import-groups` (three sorted groups) · `comments-doc-only` (a comment is
legal only as a doc comment; `TODO` is an error) · `no-else-after-return` ·
`no-naked-return`

**Errors** — `errcheck` · `wrap` (`%w`, lowercase, no trailing period) ·
`error-shape` (last result, named `err`) · `errors-is` · `no-panic` ·
`no-empty-err-branch`

**Structure** — `no-globals` · `no-init` · `no-any` · `banned-imports`
(`unsafe`, `reflect`, `math/rand`, `log`) · `no-shadow`

**Concurrency** — `ctx-first` · `ctx-origin` · `defer-unlock` ·
`sized-chan` · `go-needs-owner` · `no-time-after-select`

**API** — `no-iface-return` · `no-bare-prim` · `no-bool-param` ·
`keyed-literals` · `min-export`

**Tests and docs** — `exported-doc` · `pkg-doc` · `test-shape` ·
`test-parallel` · `no-skip`

## Why no suppression

Every escape hatch becomes the default path under deadline pressure. The
rules are strict enough to be worth arguing about, so the argument happens
once, here, rather than in every code review.

The corollary is that govanish is for code written under it from the
start. Pointed at an existing codebase it will produce thousands of
findings and offer no way to silence them.

## govanish lints itself

`make lint` runs the linter over its own source, and it is clean. This is
not decoration: it is the only honest test of whether the rule set can be
lived with, and it has repeatedly been the thing that exposed a rule as
wrong rather than the code.

`no-globals` is why every rule is a constructor function rather than a
package-level var. `no-bare-prim` is why `RuleID`, `Dir` and `Description`
exist. `min-export` is why four helpers are unexported. `file-complexity`
is why the import rules live in two files.

## Development

```
make all     # gofmt check, tests, self-lint, build
make test
make lint
```

Adding a rule takes two edits: a new file under `internal/rules/<category>/`
exporting a constructor, and one line in `internal/registry/registry.go`.
Every rule is an `analysis.Analyzer` tested with `analysistest`, so its
test is a fixture with `// want` comments and a three-line test function.

One rule is not an analyzer. `min-export` asks whether anything outside a
package uses an export, and analysis facts flow from a package to its
importers, never the reverse. It lives in `internal/wholeprogram` and runs
over the whole loaded package set.

See `docs/superpowers/specs/` for the design and `docs/superpowers/plans/`
for the build order.
