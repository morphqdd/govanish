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
