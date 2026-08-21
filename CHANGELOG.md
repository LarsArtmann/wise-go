# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- `FundTransfer` (`POST /v1/profiles/{profileId}/transfers/{transferId}/payments`):
  fund a created transfer from the profile's Wise balance, the final step of
  the core transfer flow. Returns `*FundTransferResult` with typed
  `FundingType`, `FundingStatus`, and `FundingErrorCode` (open enum, all 17
  documented codes as constants). A declined funding (e.g. insufficient
  balance) is a rejected *result*, not an error; `BalanceTransactionID`
  identifies the debit applied to the balance. SCA-protected for UK/EEA
  profiles.
- Error-path BDD coverage for every write endpoint plus `GetTransfer` and
  `ListRecipients`: 400 validation, 401, 404, 409 conflict,
  SCA-403 (`*SCAChallengeError` with one-time token), and
  429-with-`Retry-After`/`X-Rate-Limited-By` exhaustion.
- `MissingTransferDetails(requirements, req)`: cross-reference a
  transfer-requirements response against a prepared `CreateTransferRequest` to
  learn which required fields are still unsatisfied before spending a
  `customerTransactionId`. Select fields are checked against their allowed
  values; keys outside the typed request surface are reported explicitly.
- `GetQuoteAccountRequirements` (`GET /v1/quotes/{quoteId}/account-requirements`):
  the recipient fields required for an authenticated quote's currency corridor,
  one dynamic form per payout route (`AccountRequirement`). Bridges quotes to
  `CreateRecipientRequest` (route Type + Details keys). Sends
  `Accept-Minor-Version: 1` per Wise's recommendation for new integrations.
- Table-driven unit tests for all three `validate()` functions
  (transfer, quote, transfer-requirements) covering the missing-field and
  amount/currency-mismatch matrices, asserting Rejection classification.

### Fixed

- Exhausted retries now surface the typed error of the final attempt. When
  the retry policy gave up on 429/5xx responses, failsafe-go's opaque
  `retries exceeded` wrapper replaced the classification — consumers never
  saw `*RateLimitError` (with `Retry-After`) or `*ServerError`, losing
  `IsRetryable`/`ErrorFamily` semantics. `doRequest` now classifies the last
  response carried by the exceeded-retries error.

## [0.8.1] - 2026-08-21

### Fixed

- Outgoing query timestamps are normalized to UTC `Z` format. `ListTransfers`
  (`createdDateStart`/`createdDateEnd`), `ListTransactions`
  (`intervalStart`/`intervalEnd`), and exchange-rate lookups (`time`) previously
  formatted caller-local `time.Time` values with their zone offset (e.g.
  `+02:00`), which Wise rejects with HTTP 422 `wrong.date.format`. Live
  regression 2026-08-19..21: every bank-sync Wise transfers call failed for
  ~2.5 days because `To: time.Now()` carried CEST. New `formatWiseTimestamp`
  mirrors `parseWiseTimestamp` for the wire-out side.

## [0.8.0] - 2026-08-20

### Added

- `Transfer.SourceAccount` (`*BalanceID`): the balance a transfer debits, when
  Wise reports `sourceAccount`. `mapTransfer` previously parsed the wire field
  and discarded it, leaving consumers to attribute transfers by source
  currency — ambiguous whenever a profile holds multiple same-currency
  balances. Nil preserves an omitted wire value.
- `CancelTransfer` (`PUT /v1/transfers/{id}/cancel`): cancel a transfer before it
  is processed. Cancellation is final; the API rejects cancellations of
  transfers in `funds_converted` or later states with 409.
- `GetDeliveryEstimate` (`GET /v1/delivery-estimates/{id}`): live expected
  delivery time for a transfer, with an optional IANA `timezone` query
  parameter for the formatted estimate text.
- `ValidateTransferRequirements` (`POST /v1/transfer-requirements`): discover
  the dynamic transfer-detail fields required for a quote + recipient
  combination before creating the transfer. Introduces
  `ValidateTransferRequirementsRequest`, `TransferRequirement` (dynamic form),
  `TransferRequirementForm`, `TransferRequirementField`, and
  `TransferRequirementValue`. Fields flagged `RefreshRequirementsOnChange` mean
  the validation must be repeated once populated.
- `Quote` expansion: `paymentOptions` (per pay-in/pay-out combination fee
  breakdown, source/target amounts, estimated delivery, `PayInProduct`,
  `FeePercentage`), `notices` (user-facing messages; a BLOCKED notice means the
  quote must not be used), `RateType` (FIXED/FLOATING), `ProvidedAmountType`,
  `GuaranteedTargetAmountAllowed`, `GuaranteedTargetAmount`.
- `parseWiseTimestamp` now also accepts Wise's delivery-estimate layout
  (`2006-01-02T15:04:05.000+0000`).
- `GetProfile` (`GET /v2/profiles/{id}`): retrieve a single profile by ID.
- `GetExchangeRate` (`GET /v1/rates`): current and historical exchange rates between two currencies.
- `GetTransfer` (`GET /v1/transfers/{id}`): retrieve a single transfer by ID.
- Quotes API: `CreateUnauthenticatedQuote`, `CreateQuote`, and `GetQuote` (`/v3/quotes`).
  Introduces `QuoteID` (UUID string), `Quote`, `CreateQuoteRequest`, `PayIn`/`PayOut` enums,
  and `QuoteStatus`.
- Recipients API: `ListRecipients`, `GetRecipient`, and `CreateRecipient`
  (`/v1/accounts`, `/v2/accounts`). Introduces `Recipient`, `CreateRecipientRequest`,
  and `ListRecipientsRequest`.
- Transfers API: `CreateTransfer` (`POST /v1/transfers`). Introduces
  `CreateTransferRequest` with idempotency key and optional transfer details.
- Generic HTTP request helper (`Client.request`/`Client.doRequest`) supporting POST bodies,
  enabling all write operations while preserving retry, header, and error handling behaviour.
- `fetchByID` helper to eliminate duplicated get-by-ID boilerplate.
- Comprehensive API implementation plan at `docs/planning/2026-08-19_wise-api-full-implementation-plan.md`
  with Pareto prioritisation and dependency graph derived from the full Wise Platform API reference.

## [0.7.0] - 2026-08-19

### Added

- `ListTransfers` (`GET /v1/transfers`): transfer history with automatic pagination (100/page until a short page), `createdDateStart`/`createdDateEnd` and status filters. Not SCA-protected and available to personal API tokens in all regions — the reliable source for outgoing transfer history, unlike balance statements.
- `Transfer` result type with branded `TransferID`/`RecipientID` identifiers, `TransferStatus` open string enum with documented lifecycle constants, source/target `Money` amounts, exchange `Rate`, `Reference`, `CustomerTransactionID`, and `HasActiveIssues`.
- Tolerant `Created` timestamp parsing (RFC3339 and Wise's space-separated format; zoneless values interpreted as UTC).

## [0.6.1] - 2026-08-19

### Fixed

- **Retracted v0.6.0.** It was tagged on a pre-merge lineage missing the v0.5.2/v0.5.3 fixes (balances `types` parameter, Corruption classification); v0.6.1 is the same SCA feature on the integrated lineage.

## [0.6.0] - 2026-08-19

### Added

- **SCA (Strong Customer Authentication) support.** Wise answers SCA-protected endpoints (balance statements for UK/EEA profiles among them) with HTTP 403 and an EMPTY body — the verdict and the one-time token live in the `x-2fa-approval-result` / `x-2fa-approval` response headers, which the client previously discarded. Now:
  - `APIError.Headers` carries the response headers on every API error.
  - A 403 carrying Wise's two-factor headers is classified as the new `*SCAChallengeError` (error code `wise.sca_challenge`, family Rejection) instead of a bare `*AuthError`. Its `Error()` names the header values so logs finally show WHY the request was denied; `TwoFAApprovalToken()` returns the one-time token (OTT).
  - `WithSCAApprovalToken(token)` option sends the cleared OTT as `x-2fa-approval` on every request, completing the challenge flow after user approval.
  - Header name constants `HeaderTwoFAApproval` / `HeaderTwoFAApprovalResult`.

## [0.5.3] - 2026-08-18

### Fixed

- **`ListBalances` sends the required `types` query parameter.** The live v4 endpoint rejects a bare `/v4/profiles/{id}/balances` with `400 query.types: NotNull`, which made every balance listing fail after authentication succeeded. The SDK now always requests all mappable balance types (`types=STANDARD,SAVINGS`, kept in sync with the `parseBalanceType` table); visibility and investment-state filtering remain client-side, so returned data is unchanged.

## [0.5.2] - 2026-08-18

### Fixed

- **Response-shape parse failures are now classified as `Corruption`.** Mapper errors (unparseable timestamps, unknown enum types, invalid currency codes) were plain errors, so consumers that blanket-wrap SDK failures as `Transient` retried permanent failures with exponential backoff — a bad `createdAt` once silenced a sync for hours. All mapper and money-parse errors now carry `errorfamily` Corruption classification with dot-notation codes (`wise.profile.parse_created_at`, `wise.balance.parse_creation_time`, `wise.transaction.parse_date`, `wise.money.parse`) so consumers can fail fast.

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
