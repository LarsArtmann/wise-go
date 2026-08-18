# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

## [0.5.1] - 2026-08-18

### Fixed

- **Zoneless timestamps from the live Wise API now parse.** The `/v2/profiles` endpoint returns `createdAt` without a zone designator (e.g. `"2020-05-27T10:27:22"`), which made `ListProfiles` (and therefore `Authenticate`) fail with `parsing time ... cannot parse "" as "Z07:00"`. All timestamps now go through one tolerant parser (`parseWiseTimestamp`) that accepts RFC3339 (with `Z` or offset), zoneless `T`-separated, and space-separated statement dates; zoneless values are interpreted as UTC. Replaces the two strict parsers (`parseRFC3339`, `parseWiseDate`).

## [0.5.0] - 2026-08-08

### Changed

- **BREAKING: `ListTransactionsRequest.Type` is now `DetailType`** (was `string`). The `DetailType*` constants (`DetailTypeCardPayment`, `DetailTypeCardRefund`, etc.) are now typed values of `DetailType` instead of untyped string constants. This prevents sending an invalid type filter to the Wise API at compile time. Use `req.Type = wise.DetailTypeCardPayment` instead of `req.Type = "CARD_PAYMENT"`.

### Added

- `DetailType` typed string enum for Wise `details.type` wire values, replacing untyped `DetailType*` constants.
- Testable godoc examples (`ExampleNewCurrency`, `ExampleMoney_String`, `ExampleMoney_String_negative`) so `go doc` shows usage.
- Tests for `toMoney` currency validation failure path (invalid currency codes in API responses).
- BDD test for zero end-of-statement balance edge case.

### Fixed

- CONTRIBUTING.md stale references updated for v0.4.0+ API: `AmountCents`/`TotalCents` → `Amount.Cents`/`Total.Cents`, `ProfileResult`/`BalanceResult` → `Profile`/`Balance`, raw types location corrected to `internal/raw`, `internal/` removed from "no subpackages" claim.

### CI

- Nix flake check in CI no longer uses `--no-build` — the full sandboxed test now runs.
- gofumpt format check added to the CI lint job.

## [0.4.0] - 2026-08-08

### Changed

- **BREAKING: `Money` + `Currency` value objects.** All paired `XxxCents int64` / `XxxCurrency string` fields are collapsed into single `Money` fields (`Amount`, `Fees`, `Total`, `RunningBalance` on `Transaction`; `From`/`To` on `TransactionExchange`; `Amount`/`Reserved` on `Balance`). `ListTransactionsRequest.Currency` is now `wise.Currency` (not `string`). Mismatched currency/amount is now unrepresentable. Use `wise.Currency("EUR")` for direct construction or `wise.NewCurrency("EUR")` for validated construction.
- **BREAKING: `ProfileResult` renamed to `Profile`; `BalanceResult` renamed to `Balance`.** Raw wire types moved to `internal/raw`, freeing the clean names. The old `wise.Profile` (raw) and `wise.Balance` (raw) types are no longer exported.
- **BREAKING: `BalanceType` enum values normalized to lowercase.** `BalanceTypeStandard` changed from `"STANDARD"` to `"standard"`; `BalanceTypeSavings` from `"SAVINGS"` to `"savings"`. `ProfileType` and `TransactionType` were already lowercase. The parser still accepts Wise's uppercase wire format.
- **BREAKING: `ListTransactionsResponse.HasMore` removed.** The field was always `false`; Wise returns all transactions in a single response.
- **BREAKING: `TransactionTypeUnknown` constant removed.** `classifyTransactionType` never returned it (default falls back to credit/debit based on amount sign).
- **BREAKING: `EndOfStatementBalance` exposed as `Money`** on `ListTransactionsResponse`. Previously decoded from the Wise API then discarded.
- Bump [go-branded-id](https://github.com/larsartmann/go-branded-id) from v0.3.2 to **v0.5.1**.
- Bump [go-error-family](https://github.com/larsartmann/go-error-family) from v0.7.0 to **v0.10.0**.

### Added

- `Money` struct with `Cents int64` + `Currency Currency` and a `String()` method (`"EUR 12.34"`).
- `Currency` type with `NewCurrency(s string) (Currency, error)` ISO 4217 validation (3-letter uppercase ASCII).
- `toMoney(raw.BalanceAmount) (Money, error)` internal helper for raw-to-clean conversion with currency validation.
- `internal/raw` package containing all Wise wire-format types (previously in the public `wise` package).
- `nix:` CI job in `.github/workflows/ci.yml` using `cachix/install-nix-action`.
- README sections: "Mocking the Client", "Request Middleware", UTC timezone note on `Transaction.Date`.
- Unit tests for `NewCurrency` validation and `Money.String()` formatting.
- `currencyCodeLength` constant to replace magic number in currency validation.

### Fixed

- **Depguard configuration** — 14 false-positive errors blocking all third-party imports. Added `failsafe-go`, `go-branded-id`, `go-error-family`, and `onsi` to the allow-list.
- **36 golangci-lint issues resolved** — varnamelen (ignore common short names), makezero (use `make([]T, 0, len(...))`), mnd (extract constants), inamedparam (name `Doer.Do` parameter), err113 (nolint on generic enum parser).
- Updated stale version references in AGENTS.md and `.buildflow.yml` to match go.mod.

### Migration Guide

| Old (v0.3.0)                           | New (v0.4.0)                             |
| -------------------------------------- | ---------------------------------------- |
| `tx.AmountCents`                       | `tx.Amount.Cents`                        |
| `tx.AmountCurrency`                    | `tx.Amount.Currency`                     |
| `tx.TotalCents`                        | `tx.Total.Cents`                         |
| `tx.FeesCents`                         | `tx.Fees.Cents`                          |
| `tx.RunningBalanceCents`               | `tx.RunningBalance.Cents`                |
| `exch.FromCents` / `exch.FromCurrency` | `exch.From.Cents` / `exch.From.Currency` |
| `balance.AmountCents`                  | `balance.Amount.Cents`                   |
| `balance.ReservedCents`                | `balance.Reserved.Cents`                 |
| `wise.ProfileResult`                   | `wise.Profile`                           |
| `wise.BalanceResult`                   | `wise.Balance`                           |
| `req.Currency = "EUR"`                 | `req.Currency = wise.Currency("EUR")`    |
| `resp.HasMore`                         | _(removed — always false)_               |
| `wise.TransactionTypeUnknown`          | _(removed — never returned)_             |
| `BalanceTypeStandard == "STANDARD"`    | `BalanceTypeStandard == "standard"`      |

## [0.3.0] - 2026-07-18

### Changed

- **BREAKING: `GOEXPERIMENT=jsonv2` is now a hard, end-to-end requirement.** `go-branded-id v0.3.2` and `go-error-family v0.7.0` import `encoding/json/v2`, which only builds when the `jsonv2` experiment is enabled. Every consumer must `export GOEXPERIMENT=jsonv2` before any `go build` / `go test` / `go mod tidy`, or the build fails with `build constraints exclude all Go files in encoding/json/v2`. Wired into every build surface: the top-level `env` of `.github/workflows/ci.yml` (inherited by all jobs), both `flake.nix` devShells, the `buildGoModule` checkPhase, and the `.golangci.yml` `run.build-tags`. Documented in README (Installation + Testing), CONTRIBUTING.md (dedicated section), and AGENTS.md. On NixOS, `nix develop` sets it automatically.
- Bump [go-branded-id](https://github.com/larsartmann/go-branded-id) from v0.3.1 to **v0.3.2**.
- Bump [go-error-family](https://github.com/larsartmann/go-error-family) from v0.6.1 to **v0.7.0**.
- Rewrote `CONTRIBUTING.md` to match the actual project (single `package wise`, `flake.nix` workflow, GOEXPERIMENT requirement, real conventions); the previous version described a fictional clean-architecture layout with `just`, `pkg/errors/`, and `cmd/` directories that do not exist.
- Documented UTC timezone assumption on `parseWiseDate` and `Transaction.Date` (Wise sends no timezone; `time.Parse` interprets as UTC).
- README: added coverage/lint/Go badges, sharpened value proposition, moved project status above the fold, linked to FEATURES.md and ROADMAP.md.

### Added

- `flake.nix` — reproducible devShells (default + CI), treefmt (gofumpt + goimports + nixfmt), and a sandboxed test check via `buildGoModule`. Use `nix develop`, `nix fmt`, `nix flake check`.
- `FEATURES.md` — honest feature inventory by status (FULLY_FUNCTIONAL / PARTIALLY_FUNCTIONAL / BROKEN / PLANNED) with code evidence.
- `TODO_LIST.md` — short-term actionable tasks sorted by priority.
- `ROADMAP.md` — long-term direction across completeness, type-safety, and scale axes.
- BDD tests for `ListTransactionsRequest.Type` filter forwarding (previously untested code path).
- Unit tests for positive-amount `CARD_PAYMENT` classification (regression guard).

### Fixed

- **BREAKING: `CARD_PAYMENT` classification corrected.** `classifyTransactionType` grouped `CARD_PAYMENT` and `CARD_REFUND` into a single amount-based case, causing positive-amount card payments (e.g. reversed charges) to be silently classified as `TransactionTypeRefund`. Split into two cases: `CARD_PAYMENT` always returns `TransactionTypeCard`; `CARD_REFUND` retains the amount-dependent dispatch. Consumers switching on `tx.Type` for `CARD_PAYMENT` details will see corrected values.
- **Branded-ID formatting in error messages** — `ListTransactions` passed `req.ProfileID` and `req.BalanceID` directly to `%d` in two `fmt.Errorf` sites, inconsistent with the `.Get()` pattern used everywhere else. Normalized to `.Get()`.

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
