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
| golangci-lint | v2.12   | Linting (the `nix develop` shell provides this)         |

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
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.0
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
- `.golangci.yml` — `run.build-tags: [goexperiment.jsonv2, ...]` so the analyzer sees the same code the compiler does.
- `.github/workflows/ci.yml` — top-level `env: GOEXPERIMENT: "jsonv2"`, inherited by every job.

`nix develop` sets it for you, so inside that shell plain `go test ./...` works.

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
- **Retries** via `failsafe-go` — only 429, 5xx, and network errors are retried.

---

## Testing

```bash
# Hermetic, includes race detector + coverage
nix flake check

# Or manually (remember GOEXPERIMENT)
go test -race -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -func=coverage.out | tail -1
```

Tests use `net/http/httptest` to mock the Wise API — **no network access, no API key required**. Coverage is currently ~95%.

### Test style

- BDD-style with Ginkgo for `wise_test.go` (black-box `package wise_test`).
- Internal unit tests in `internal_test.go` (white-box `package wise`).
- Use the `Given..._When..._Should...` naming pattern.

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
- [Go documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)

If you are stuck, open a discussion or check existing issues before struggling alone.

---

_Thank you for contributing to wise-go!_
