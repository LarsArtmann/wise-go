# TODO List

Short- and mid-term actionable work for wise-go. Each item is bounded, has clear
ownership, and can be completed in one sitting. For long-term vision and raw ideas
see [ROADMAP.md](ROADMAP.md). For shipped features see [FEATURES.md](FEATURES.md).

## P1 — v1.0 release (API lock)

- [ ] **Lock the public API at v1.0** — formal audit of every exported symbol,
      godoc review pass, then tag `v1.0.0`. All v0.4.0 breaking changes are shipped
      and tested; the lock is the logical next milestone. Requires explicit approval
      (tagging is irreversible). Deferred from v0.4.0 session
      (`docs/status/2026-08-08_02-53_pareto-plan-v040-implementation.md`).

## P2 — Type-safety refinements

- [ ] **`Money` arithmetic API** — add `Add`, `Sub`, `IsZero`, `IsNegative`, `Equal`
      methods with currency-mismatch checks. Consumers currently handle cents math
      and currency comparison themselves. (`types.go:55`)
- [ ] **`classifyTransactionType` takes `Money` not `float64`** — the classifier
      leaks the raw API's `float64` representation into the clean layer
      (`transactions.go:180`). Refactor to accept `int64` cents or `Money`.
- [ ] **`ListTransactionsRequest.Type` typed enum** — currently `string` (`types.go:182`);
      should be a typed enum matching the exported `DetailType*` constants
      (`transactions.go:165`).

## P3 — Test coverage gaps

- [ ] **Test `toMoney` currency validation failure path** — the error branches in
      `toMoney` (`helpers.go:17`) are not covered by BDD tests. A malformed currency
      code in a Wise response currently fails the entire `ListTransactions` call.
- [ ] **Test `EndOfStatementBalance` with empty/zero values** — the BDD test covers
      the happy path (`types.go:188`) but not the edge case where Wise returns an
      empty balance object.
- [ ] **Test `internal/raw.BalanceAmount.Cents()`** — the `math.Round` path
      (`internal/raw/types.go:46`) has no direct test. Move or duplicate from
      `internal_test.go`.

## P4 — Documentation

- [ ] **Update CONTRIBUTING.md for v0.4.0 API** — three stale references: still
      cites `AmountCents`/`TotalCents` (`:120`), `ProfileResult`/`BalanceResult`
      (`:123`), and describes raw types as being in `types.go` not `internal/raw`.
- [ ] **Godoc examples for `Money` and `Currency`** — add testable examples
      (`ExampleMoney_String`, `ExampleNewCurrency`) so `go doc` shows usage.

## P5 — CI / tooling polish

- [ ] **Run full `nix flake check` in CI** — currently the GitHub Actions `nix:`
      job may skip the expensive build; ensure the full check (format + sandboxed
      test) runs.
- [ ] **Add `gofumpt` to CI lint job** — currently only runs locally via `nix fmt`.
- [ ] **Add `go mod tidy` check to `nix flake check`** — currently only in GitHub Actions.
