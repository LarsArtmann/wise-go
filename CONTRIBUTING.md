# Contributing to wise-go

> **Thank you for contributing!** This guide covers everything you need to build, test, and submit changes to wise-go.

## Table of Contents

- [Quick Start](#quick-start)
- [Development Setup](#development-setup)
- [The GOEXPERIMENT=jsonv2 Requirement](#the-goexperimentjsonv2-requirement)
- [Project Layout](#project-layout)
- [Conventions](#conventions)
- [Testing](#testing)
- [Linting & Formatting](#linting--formatting)
- [Pre-commit Hooks](#pre-commit-hooks)
- [Pull Request Process](#pull-request-process)
- [Commit Messages](#commit-messages)
- [Getting Help](#getting-help)

---

## Quick Start

```bash
# 1. Clone
git clone https://github.com/LarsArtmann/wise-go.git
cd wise-go

# 2. Enter the dev shell (sets GOEXPERIMENT, Go 1.26, golangci-lint, gopls)
nix develop

# 3. Verify the build
go build ./...

# 4. Run the full suite
nix flake check      # tests + format + lint, all hermetic
```

If you do not use Nix, see [Development Setup](#development-setup) for the manual path.

---

## Development Setup

### Prerequisites

| Tool          | Version | Purpose                                                 |
| ------------- | ------- | ------------------------------------------------------- |
| Go            | 1.26+   | Language runtime. Required for the `jsonv2` experiment. |
| Nix (flakes)  | 2.18+   | Reproducible dev + CI environment (recommended)         |
| golangci-lint | v2.13   | Linting (the `nix develop` shell provides this)         |

### Recommended: Nix

```bash
nix develop          # enter the shell — GOEXPERIMENT is set automatically
```

Everything below works inside that shell without any prefix.

### Manual (non-Nix)

```bash
# Mandatory: the jsonv2 experiment must be on for every Go invocation
export GOEXPERIMENT=jsonv2   # add to ~/.bashrc or ~/.zshrc

# Tools
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.0
```

---

## The GOEXPERIMENT=jsonv2 Requirement

**This is non-negotiable.** The `go-branded-id` and `go-error-family` dependencies use `encoding/json/v2`, which only builds when the `jsonv2` experiment is enabled. Without it you will see:

```
build constraints exclude all Go files in encoding/json/v2
```

Every Go toolchain invocation needs the variable:

```bash
GOEXPERIMENT=jsonv2 go build ./...
GOEXPERIMENT=jsonv2 go test ./...
GOEXPERIMENT=jsonv2 go vet ./...
GOEXPERIMENT=jsonv2 go mod tidy
GOEXPERIMENT=jsonv2 golangci-lint run
```

Where it is already wired in:

- `flake.nix` — both devShells and the `buildGoModule` checkPhase.
- `.envrc` — `use flake` + `use_go_env`; direnv applies it on `cd` (needs direnv + nix-direnv).
- `.golangci.yml` — `run.build-tags: [goexperiment.jsonv2, ...]` so the analyzer sees the same code the compiler does.
- `.github/workflows/ci.yml` — top-level `env: GOEXPERIMENT: "jsonv2"`, inherited by every job.

`nix develop` sets it for you, so inside that shell plain `go test ./...` works.

### Editor/LSP environments need it too

Tools launched OUTSIDE the dev shell (IDE language servers, agent-launched
LSPs, hooks) do not inherit the variable. Their package load fails with the
same `build constraints exclude all Go files` error, and the degraded
analysis surfaces as phantom diagnostics (false "unused import",
`undefined: json`, bogus formatting findings) that the CLI does not show.
Inject `GOEXPERIMENT=jsonv2` into the editor's LSP server environment; this
repo's `.crushrc` does exactly that for Crush-launched `gopls` and
`golangci_lint_ls`. Root-cause write-up: AGENTS.md, "LSP phantoms were an env
problem" (2026-10-07).

---

## Project Layout

wise-go is a **single-package Go library** with one internal subpackage. The public SDK lives in `package wise` at the repository root; raw wire-format types live in `internal/raw`.

```
├── *.go              # package wise — the entire public SDK
├── flake.nix         # devShells, checks (tests + lint + format), treefmt
├── .golangci.yml     # curated linter config (63 linters)
├── go.mod / go.sum   # module github.com/larsartmann/wise-go
├── AGENTS.md         # session context + gotchas — READ THIS FIRST
├── docs/             # reviews, architecture, planning, status reports
└── README.md         # user-facing overview
```

Before contributing, read [AGENTS.md](AGENTS.md) — it documents the non-obvious behaviors (dual date formats, `Amount.Cents` vs `Total.Cents`, balance filtering, branded-ID usage, error families).

---

## Conventions

- **Money is `int64` cents** — never `float64`. `Amount.Cents` is absolute; `Total.Cents` preserves sign. Both are `Money` fields (cents + currency paired).
- **Branded IDs** — `ProfileID`, `BalanceID`, `TransactionID` are distinct phantom types from `go-branded-id`. Mixing them is a compile error. Construct with `NewProfileID` / `NewBalanceID` / `NewTransactionID`; unwrap with `.Get()`.
- **Behavioral errors** — domain error types implement `go-error-family` interfaces (`ErrorCode()`, `ErrorFamily()`, `IsRetryable()`). Never construct `AuthError` / `NotFoundError` etc. directly outside `newAPIError()` in `errors.go`.
- **Two-layer types** — raw wire structs (`raw.Profile`, `raw.Balance`, `raw.StatementTransaction` in `internal/raw`) match Wise's JSON exactly with primitives. Result types (`Profile`, `Balance`, `Transaction`) expose strong Go types (`Money`, branded IDs, enums). Mapping functions are the only bridge. Do not brand the JSON-decode layer.
- **Error wrapping at call sites** uses `fmt.Errorf("context: %w", err)`; the inner error carries the classification.
- **Retries** via `go-retry` — only 429, 5xx, and network errors are retried; Wise's `Retry-After` hint is honored as the delay, capped at `WithRetry`'s max delay.

---

## Testing

```bash
# Hermetic, includes race detector + coverage + the 90% coverage floor
nix flake check

# Or manually (remember GOEXPERIMENT)
go test -race -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -func=coverage.out | tail -1
```

Tests use `net/http/httptest` to mock the Wise API — **no network access, no API key required**. Coverage is currently ~90% (measured 90.1% on 2026-10-05 — keep an eye on it; the floor is close), and a **90% floor is enforced** in both the flake check and the CI coverage job: a silent drop below 90% fails the build.

### Test style

- BDD-style with Ginkgo for `wise_test.go` (black-box `package wise_test`).
- Internal unit tests in `internal_test.go` (white-box `package wise`).
- Use the `Given..._When..._Should...` naming pattern.

### Repeating specs

`go test -count=N` re-runs the Ginkgo bootstrap and reports false failures
(specs are not re-executed N times the way plain Go tests are). To repeat
the suite, use separate runs instead:

```bash
go test -count=1 ./...
go test -count=1 ./...   # each run is fresh; never stack -count
```

### Benchmarks

The hot parsing/mapping paths are benchmarked in `bench_test.go`. A
committed baseline lives in `docs/bench/` (`2026-10-07_v0120_baseline.txt`
at the time of writing). Regenerate and compare:

```bash
GOEXPERIMENT=jsonv2 go test -bench . -benchmem -count=6 -run '^$' . > /tmp/new.txt
benchstat docs/bench/2026-10-07_v0120_baseline.txt /tmp/new.txt
```

Commit a fresh baseline file (named `<date>_<label>.txt`) when benchmark
names or shapes change, or after landing a change that moves the numbers.

---

## Linting & Formatting

```bash
# Inside nix develop (GOEXPERIMENT already set):
golangci-lint run          # 63 curated linters, see .golangci.yml
nix fmt                    # gofumpt + goimports + nixfmt

# Manual:
GOEXPERIMENT=jsonv2 golangci-lint run
```

The linter config is deliberately curated. Do **not** run `buildflow auto-configure` or `buildflow --fix` — it replaces the curated list with 100+ generic linters (including ones that flag legitimate patterns in this codebase) and breaks the build.

### Quality gates

Before pushing, all of these must pass:

- [ ] `go build ./...`
- [ ] `go vet ./...`
- [ ] `go test -race ./...`
- [ ] `golangci-lint run` — 0 issues
- [ ] `nix flake check` — all checks pass

### Documentation & link checks

`nix flake check` runs an **offline link check** over the living docs: every
relative link in README.md must point at a file that exists **and** is listed
in the `links` fileset union in `flake.nix` (around the `markdown-links`
check). When you add a relative link to README, add its target to that union
or the check fails with "File not found".

Two README guardrails run with the test suite:

- `readme_guard_test.go` parses every Go code fence in README.md and fails if
  it references a `client.Method` or `wise.Symbol` that no longer exists —
  stale examples become test failures, not reader surprises.
- The coverage-badge CI job rewrites the README coverage percentage after
  every push to master; expect the in-repo number to lag your local
  measurement until the branch is pushed.

For a local link check (including external URLs), run the same tool the flake
check uses — `lychee` is in the devShell:

```bash
lychee --offline --no-progress README.md CONTRIBUTING.md
```

### Doc health & API-compat apps

Two flake apps automate the doc/compat checks that are easy to forget:

```bash
nix run .#doc-verify   # godoc render + count claims in living docs + links
nix run .#apidiff      # gorelease diff vs the latest tag (needs network)
```

`doc-verify` fails when any documented count claim (endpoint methods,
`Example` funcs) no longer matches the compiled surface, or when the claim
text has drifted so far the pattern extracts nothing. Run it after changing
the public API or the godoc examples — the claim-staleness class ("33
methods" outliving the 37-method surface) is exactly what it catches.
`apidiff` reports removed or changed exported API against the latest tag;
run it before any release.

### Release gates

```bash
nix run .#release-notes-check   # split code spans + repo-relative links in docs/releases/*.md
nix run .#pin-sync              # in-house flake pins match go.mod versions (needs network)
nix run .#pre-release           # the whole release gate chain, one command
nix run .#pre-release --shuffle # ...plus three randomized-order test passes, for tag day
```

`release-notes-check` fails on the two defect classes that shipped in the
v0.12.0 release body: a code span split across lines by a markdown reflow,
and repo-relative links (GitHub Release bodies cannot resolve repo-relative
paths — use absolute URLs). Run it before `gh release create`.
`pre-release` chains the full local gate sequence — dirty-tree check, build,
vet, race tests, lint, `nix flake check`, `doc-verify`,
`release-notes-check`, `pin-sync`, `apidiff` — so a release relies on one
command instead of session discipline. Pass `--shuffle` on tag day: Ginkgo
supports only `-count=1`, so repetition is three separate `-shuffle=on`
passes, not a `-count` flag.

**Flake-pin invariant.** A dependency version lives in two places — the
`require` block of `go.mod` and the matching `flake.nix` input `rev`. A
`go get` that bumps one without the other leaves `go test` and the hermetic
`nix flake check` verifying DIFFERENT versions; because a pinned `rev` is
immutable under `nix flake update`, the drift is silent (go-retry's pin
lagged its `go.mod` tag for ten days). Any `go.mod` version bump MUST update
the matching input `rev` in the same change. `checks.pin-sync` enforces this
hermetically in `nix flake check`; `nix run .#pin-sync` additionally resolves
each version tag on GitHub and fails on a stale rev.

### Quarterly API-surface rollover ritual

Wise versions some surfaces by quarter (`2026Q3`, `2026Q4`); the webhook
subscription CRUD and the OTT endpoints roll over INDEPENDENTLY (see
AGENTS.md). The 2026Q4 flip proved that a value-flip is not done until the
straggler grep is clean — a stale path in a comment or doc survives every
test. The ritual:

1. **Probe** — confirm the new prefix answers non-404 on the surface.
   Unauthenticated probes 404 on some surfaces even when live (e.g.
   `one-time-token`); those need sandbox credentials (see TODO_LIST.md).
2. **Flip** — change the version constant in `client.go`
   (`webhookSubscriptionsAPIVersion` / `ottAPIVersion`). The two roll over
   independently; never move both because one moved.
3. **Straggler grep** — `grep -rn "2026Q3" --include='*.go' .` (the old
   value) must come back empty outside changelog/history; update any
   comment, test path, or doc that still names the old prefix.
4. **Changelog** — record the flip under `[Unreleased]`.
5. **Verify** — `go test ./...` and `nix run .#doc-verify` (the spec
   conformance harness normalizes quarterly prefixes; a stale path fails
   there).

---

## Pre-commit Hooks

Pre-commit hooks are provided via [git-hooks.nix](https://github.com/cachix/git-hooks.nix) and wired into `flake.nix`. They run `nix fmt` and (if installed) `buildflow` validation including `govalid-generate`.

Because the hooks invoke the Go toolchain, **you must commit with `GOEXPERIMENT=jsonv2` in your environment**:

```bash
GOEXPERIMENT=jsonv2 git commit -m "..."
```

Inside `nix develop` this is handled automatically.

---

## Pull Request Process

### Branch naming

```
feat/description
fix/description
docs/description
refactor/description
test/description
chore/description
```

### Before opening a PR

1. **Self-review** — run the full quality gate suite locally.
2. **Small and focused** — one logical change per PR.
3. **Explain the "why"** — the PR description should motivate the change, not just list files.
4. **Update docs** — if your change affects behavior, update `README.md`, `AGENTS.md` gotchas, and `CHANGELOG.md` `[Unreleased]`.

### PR description template

```markdown
## Summary

Brief description of the change and why.

## Type

- [ ] Feature
- [ ] Bug fix
- [ ] Refactoring
- [ ] Documentation

## Test plan

- [ ] Unit tests added/updated
- [ ] `nix flake check` passes
- [ ] `golangci-lint run` is clean
```

---

## Commit Messages

Conventional Commits format:

```
<type>(<scope>): <subject>

<body — explain the why>

<footer>
```

### Types

| Type     | Description              |
| -------- | ------------------------ |
| feat     | New feature              |
| fix      | Bug fix                  |
| docs     | Documentation changes    |
| style    | Formatting, whitespace   |
| refactor | Code restructuring       |
| test     | Adding/updating tests    |
| chore    | Build, tooling, CI       |
| perf     | Performance improvements |
| ci       | CI/CD changes            |
| revert   | Reverting changes        |

### Examples

```bash
# Good
fix(transactions): classify CARD_PAYMENT separately from CARD_REFUND

CARD_PAYMENT and positive-amount CARD_REFUND were both mapped to
TransactionTypeCard, hiding refunds. Split the classification so
refunds surface correctly.

Closes #42

# Bad
fix stuff

# Good
ci: require GOEXPERIMENT=jsonv2 across all jobs

# Bad
updated CI
```

---

## Getting Help

- [AGENTS.md](AGENTS.md) — project gotchas and conventions
- [README.md](README.md) — user-facing overview and API examples
- [SECURITY.md](.github/SECURITY.md) — how to report vulnerabilities (never in public issues)
- [Go documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)

If you are stuck, open a discussion or check existing issues before struggling alone.

---

_Thank you for contributing to wise-go!_
