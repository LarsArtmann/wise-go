# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

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
