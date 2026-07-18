# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Fixed

- **CARD_PAYMENT classification** — `classifyTransactionType` grouped `CARD_PAYMENT` and `CARD_REFUND` into a single amount-based case, causing positive-amount card payments (e.g. reversed charges) to be silently classified as `TransactionTypeRefund`. Split into two cases: `CARD_PAYMENT` always returns `TransactionTypeCard`; `CARD_REFUND` retains the amount-dependent dispatch.
- **Branded-ID formatting in error messages** — `ListTransactions` passed `req.ProfileID` and `req.BalanceID` directly to `%d` in two `fmt.Errorf` sites, inconsistent with the `.Get()` pattern used everywhere else. Normalized to `.Get()`.

### Added

- `flake.nix` — reproducible devShells (default + CI), treefmt (gofumpt + goimports + nixfmt), and a sandboxed test check via `buildGoModule`. Use `nix develop`, `nix fmt`, `nix flake check`.
- `FEATURES.md` — honest feature inventory by status (FULLY_FUNCTIONAL / PARTIALLY_FUNCTIONAL / BROKEN / PLANNED) with code evidence.
- `TODO_LIST.md` — short-term actionable tasks sorted by priority.
- `ROADMAP.md` — long-term direction across completeness, type-safety, and scale axes.
- BDD tests for `ListTransactionsRequest.Type` filter forwarding (previously untested code path).
- Unit tests for positive-amount `CARD_PAYMENT` classification (regression guard).

### Changed

- **`GOEXPERIMENT=jsonv2` is now a hard, end-to-end requirement.** `go-branded-id v0.3.2` and `go-error-family v0.7.0` import `encoding/json/v2`, which only builds with the `jsonv2` experiment enabled. Wired it into every build surface: the top-level `env` of `.github/workflows/ci.yml` (inherited by all jobs), both `flake.nix` devShells, the `buildGoModule` checkPhase, and the `.golangci.yml` `run.build-tags`. Documented the requirement in README (Installation + Testing), CONTRIBUTING.md (dedicated section), and AGENTS.md.
- Bump [go-branded-id](https://github.com/larsartmann/go-branded-id) from v0.3.1 to **v0.3.2**.
- Bump [go-error-family](https://github.com/larsartmann/go-error-family) from v0.6.1 to **v0.7.0**.
- Rewrote `CONTRIBUTING.md` to match the actual project (single `package wise`, `flake.nix` workflow, GOEXPERIMENT requirement, real conventions); the previous version described a fictional clean-architecture layout with `just`, `pkg/errors/`, and `cmd/` directories that do not exist.
- Documented UTC timezone assumption on `parseWiseDate` and `Transaction.Date` (Wise sends no timezone; `time.Parse` interprets as UTC).
- README: added coverage/lint/Go badges, sharpened value proposition, moved project status above the fold, linked to FEATURES.md and ROADMAP.md.

## [0.2.0] - 2026-07-05

### Changed

- **BREAKING:** `ListBalances` and `GetBalance` now take branded `ProfileID`/`BalanceID` parameters instead of raw `int64`, matching the existing `ListTransactions` API. Use `wise.NewProfileID(id)` / `wise.NewBalanceID(id)` to construct them. This makes mixing different entity IDs a compile-time error.
- **BREAKING:** Removed `WithNow` option and the entire `now func() time.Time` chain — `mapTransaction` never used it. If you passed `WithNow(...)`, simply drop the option.
- Bump [go-error-family](https://github.com/larsartmann/go-error-family) from v0.4.0 to **v0.6.0**.
- Bump Go toolchain requirement to **1.26.4**.
- Replace custom `joinStrings` helper with stdlib `strings.Join`.
- Strengthen linter configuration.

### Added

- `Transaction.RunningBalanceCents` and `Transaction.RunningBalanceCurrency` — Wise returns `runningBalance` on every transaction; previously silently dropped.
- `Transaction.Exchange` (`*TransactionExchange`) — Wise returns `exchangeDetails` on conversion transactions; previously never mapped.
- `InvestmentStateNotInvested` / `InvestmentStateInvested` constants — replaces the magic string `"NOT_INVESTED"` in balance filtering.
- Named wire constants for Wise `detail.type` values (`wiseDetailCardPayment`, `wiseDetailCardRefund`, etc.) — replaces repeated string literals in `classifyTransactionType`.
- `ListTransactionsRequest.validate()` — empty currency or inverted date range now fails fast with `wise.transactions.invalid_request` (no network round-trip).
- New `internal_test.go` with table-driven unit tests covering `Cents()`, ID constructors, error type assertions, `RetryAfter` threading, and parser error paths.
- Cross-currency transaction mapping test and `Retry-After` header retry test in `wise_test.go`.
- Test coverage 86.1% → **94.8%**; test count ~30 → 41.
- Code of Conduct.

### Fixed

- **Currency source in `mapTransaction`** — Used the request-level `Currency` parameter instead of each transaction's own `t.Amount.Currency`. For cross-currency statements this silently carried the wrong ISO 4217 code across `AmountCurrency`, `TotalCurrency`, and `RunningBalanceCurrency`.
- **`RetryAfter` hardcoded to 1 second** — `RateLimitError` always reported `time.Second` regardless of the HTTP `Retry-After` header. Added `parseRetryAfter` (delta-seconds + RFC 1123 HTTP-date), and `checkError` now reads and threads the header through.
- **False `WithHiddenBalances()` doc** — `ListBalances` documented an option that never existed. Rewrote to accurately describe the visible/non-invested filtering behavior.
- Remove unused `UserID` / `UserBrand` / `NewUserID` types (no Wise endpoint returns a standalone user ID).
- Inline trivial `jsonUnmarshal` wrapper into `parseErrorResponse`.
- Avoid an intermediate string allocation when joining API error messages.

## [0.1.0] - 2026-06-03

### Added

- Client with Bearer token authentication
- `ListProfiles` — list all profiles for the authenticated user
- `ListBalances` — list visible, non-investment balances for a profile
- `GetBalance` — get a specific balance by ID
- `ListTransactions` — list transactions for a balance within a date range
- `Authenticate` / `Health` — validate API key connectivity
- Automatic retries with exponential backoff (429, 5xx, network errors)
- Typed error hierarchy: `APIError`, `RateLimitError`, `AuthError`, `NotFoundError`, `ServerError`
- Strongly-typed result types with `int64` cents and `time.Time` dates
- Transaction type classification (card, credit, debit, exchange, fee, refund, transfer, payment)
- Sandbox environment support via `WithSandbox()`
- Functional options: `WithBaseURL`, `WithTimeout`, `WithRetry`, `WithHTTPClient`, `WithNow`
