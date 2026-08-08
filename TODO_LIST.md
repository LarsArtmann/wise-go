# TODO List

Short- and mid-term actionable work for wise-go. Each item is bounded, has clear
ownership, and can be completed in one sitting. For long-term vision and raw ideas
see [ROADMAP.md](ROADMAP.md). For shipped features see [FEATURES.md](FEATURES.md).

## Status key

- `[ ]` not started
- `[~]` in progress
- `[x]` done

## P1 — Should ship in the next release (v0.2.1 / v0.2.2)

These are small, non-breaking, high-value cleanups identified in the 2026-07-18 review pass.

- [x] **Type `InvestmentState`** — promote bare `string` constants at `types.go:192` to
      `type InvestmentState string`; update filter at `balances.go:28`. Zero API break
      for typical callers. (data-model review Step 1)
- [x] **Export `DetailType` constants** — promote the unexported `wiseDetail*` constants
      in `transactions.go:137-145` so callers of `ListTransactionsRequest.Type` can
      discover the valid set without reading the README. Keep the field as `string` for
      backward compat. (data-model review Step 2)
- [x] **Accept `Doer` interface in `WithHTTPClient`** — change `options.go:58` and
      `client.go:25` from `*http.Client` to an unexported `doer interface{ Do(...) }`.
      `*http.Client` satisfies it implicitly. (architecture review Step 3)
- [ ] **Verify full `nix flake check`** — run the test + lint derivations end-to-end
      locally before merging the new `flake.nix`. The `--no-build` variant passes.
      (nix-flake-migration proposal)
- [ ] **Add `nix flake check` to CI** — add a `nix:` job to `.github/workflows/ci.yml`
      using `cachix/install-nix-action`. Caches recommended for the Go module fetch.

## P2 — v0.3.0 (breaking change release)

These are coordinated breaking changes; ship together.

- [ ] **Introduce `Money` + `Currency` types** — collapse paired `XxxCents int64` /
      `XxxCurrency string` fields across `Transaction`, `TransactionExchange`,
      `BalanceResult`, `ListTransactionsRequest`. (data-model review Step 3)
- [ ] **Normalize enum casing** — pick one rule (suggest lowercase for SDK enums) and
      apply to `ProfileType`, `BalanceType`, `TransactionType`. (data-model review Step 4)
- [ ] **Reconcile `TransactionTypeUnknown`** — either use it as the actual fallback
      in `classifyTransactionType` or remove the constant. (data-model review Step 4)
- [ ] **Drop the `Result` suffix or move raw types to `internal/raw`** — fix the
      naming inconsistency between `ProfileResult`/`BalanceResult` vs `Transaction`.
      (naming review)
- [ ] **Expose `EndOfStatementBalance`** — Wise returns it; the SDK decodes it and
      throws it away. Surface on `ListTransactionsResponse` (paired with `Money`).

## P3 — v1.0 (lock-in release)

- [ ] **Remove `ListTransactionsResponse.HasMore`** — return `[]Transaction` directly.
      The field is always `false`; the type lies about its capabilities.
- [ ] **Move raw wire types behind `internal/raw`** — `Profile`, `Balance`,
      `StatementTransaction`, etc. become invisible to consumers. (architecture review Step 5)
- [ ] **Lock the public API** — at v1.0, freeze the surface; future breaking changes
      require v2.

## P4 — Documentation polish

- [ ] **Add "Mocking the client" README section** — show consumers how to define
      narrow interfaces (`type ProfileLister interface { ... }`) for test mocking.
      (architecture review Step 1)
- [ ] **Add "Request middleware via WithHTTPClient" README section** — show how to
      inject tracing/logging via a custom `http.Client` Transport. (architecture review Step 2)
- [ ] **Document UTC assumption on `Transaction.Date`** — already added as a field
      comment on 2026-07-18; mirror in README's transaction section if/when it grows.

## Done

### 2026-08-08

- [x] Fix depguard config — allow failsafe-go, go-branded-id, go-error-family, onsi in `.golangci.yml`
- [x] Fix remaining lint issues (varnamelen, makezero, mnd, inamedparam, err113) — 0 issues

### 2026-07-18

- [x] Fix `CARD_PAYMENT` with positive amount being misclassified as `TransactionTypeRefund`
- [x] Normalize `%d` branded-ID formatting in `transactions.go` error paths (`.Get()` pattern)
- [x] Add 2 BDD tests for `ListTransactionsRequest.Type` filter forwarding
- [x] Add 2 unit tests for positive-amount `CARD_PAYMENT` classification
- [x] Document UTC timezone assumption on `parseWiseDate` and `Transaction.Date`
- [x] Create `flake.nix` (devShells + checks + treefmt)
- [x] Create `FEATURES.md`, `TODO_LIST.md`, `ROADMAP.md` (this file)
