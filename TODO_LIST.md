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

- [x] **Type `InvestmentState`** — promote bare `string` constants to
      `type InvestmentState string`. (data-model review Step 1)
- [x] **Export `DetailType` constants** — callers of `ListTransactionsRequest.Type` can
      discover the valid set. (data-model review Step 2)
- [x] **Accept `Doer` interface in `WithHTTPClient`** — `*http.Client` satisfies it implicitly.
      (architecture review Step 3)
- [x] **Verify full `nix flake check`** — passes (format + sandboxed test via buildGoModule).
- [x] **Add `nix flake check` to CI** — `nix:` job added to `.github/workflows/ci.yml`
      using `cachix/install-nix-action`.

## P2 — v0.4.0 (breaking change release)

These were coordinated breaking changes; shipped together in one release.

- [x] **Introduce `Money` + `Currency` types** — collapse paired `XxxCents int64` /
      `XxxCurrency string` fields across `Transaction`, `TransactionExchange`,
      `Balance`, `ListTransactionsRequest`. Mismatched currency/amount now unrepresentable.
- [x] **Normalize enum casing** — lowercase for all SDK-facing enums (`BalanceType` normalized
      from `"STANDARD"` to `"standard"`; `ProfileType` and `TransactionType` already lowercase).
- [x] **Reconcile `TransactionTypeUnknown`** — removed; `classifyTransactionType` never
      returned it (default falls back to credit/debit).
- [x] **Drop the `Result` suffix + move raw types to `internal/raw`** — `ProfileResult` → `Profile`,
      `BalanceResult` → `Balance`. Raw wire types moved to `internal/raw` package.
- [x] **Expose `EndOfStatementBalance`** — surfaced as `Money` on `ListTransactionsResponse`.

## P3 — v1.0 (lock-in release)

- [x] **Remove `ListTransactionsResponse.HasMore`** — field was always `false`; removed.
- [x] **Move raw wire types behind `internal/raw`** — `Profile`, `Balance`,
      `StatementTransaction`, etc. are invisible to consumers.
- [ ] **Lock the public API** — at v1.0, freeze the surface; future breaking changes
      require v2. (Pending: final API audit + tag)

## P4 — Documentation polish

- [x] **Add "Mocking the client" README section** — narrow consumer-side interface pattern.
- [x] **Add "Request middleware via WithHTTPClient" README section** — Transport wrapping.
- [x] **Document UTC assumption on `Transaction.Date`** — added to README transactions section.

## Done

### 2026-08-08

- [x] Fix depguard config — allow failsafe-go, go-branded-id, go-error-family, onsi in `.golangci.yml`
- [x] Fix remaining lint issues (varnamelen, makezero, mnd, inamedparam, err113) — 0 issues
- [x] Introduce `Money` + `Currency` value objects with ISO 4217 validation
- [x] Refactor all structs to use `Money` (Transaction, TransactionExchange, Balance, ListTransactionsResponse)
- [x] Normalize `BalanceType` enum casing to lowercase
- [x] Remove dead `TransactionTypeUnknown` constant
- [x] Move raw wire types to `internal/raw` package
- [x] Drop `Result` suffix: `ProfileResult` → `Profile`, `BalanceResult` → `Balance`
- [x] Remove `HasMore` from `ListTransactionsResponse`
- [x] Expose `EndOfStatementBalance` as `Money`
- [x] Add README sections: Mocking, Middleware, UTC timezone note
- [x] Add `nix:` CI job to `.github/workflows/ci.yml`

### 2026-07-18

- [x] Fix `CARD_PAYMENT` with positive amount being misclassified as `TransactionTypeRefund`
- [x] Normalize `%d` branded-ID formatting in `transactions.go` error paths (`.Get()` pattern)
- [x] Add 2 BDD tests for `ListTransactionsRequest.Type` filter forwarding
- [x] Add 2 unit tests for positive-amount `CARD_PAYMENT` classification
- [x] Document UTC timezone assumption on `parseWiseDate` and `Transaction.Date`
- [x] Create `flake.nix` (devShells + checks + treefmt)
- [x] Create `FEATURES.md`, `TODO_LIST.md`, `ROADMAP.md` (this file)
