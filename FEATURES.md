# Features

Honest inventory of wise-go features by status. Code is the source of truth — every
claim here can be verified against the implementation.

## Status vocabulary

| Status               | When it applies                                              |
| -------------------- | ------------------------------------------------------------ |
| FULLY_FUNCTIONAL     | Code present AND working (tests pass or exercised).          |
| PARTIALLY_FUNCTIONAL | Ships but has known gaps, edge-case bugs, or missing pieces. |
| BROKEN               | Code exists but does not work / is disabled / fails.         |
| PLANNED              | Designed or documented but **no code exists yet**.           |

## Client core

| Feature                                 | Status           | Evidence                                                                  |
| --------------------------------------- | ---------------- | ------------------------------------------------------------------------- |
| `wise.New(apiKey, opts...)` constructor | FULLY_FUNCTIONAL | `client.go:43`; functional-options pattern                                |
| Bearer-token authentication             | FULLY_FUNCTIONAL | `client.go:355` `setHeaders`                                              |
| Sandbox environment (`WithSandbox`)     | FULLY_FUNCTIONAL | `options.go:64`; `SandboxURL` const in `types.go:24`                      |
| Custom base URL (`WithBaseURL`)         | FULLY_FUNCTIONAL | `options.go:71`                                                           |
| Custom HTTP timeout (`WithTimeout`)     | FULLY_FUNCTIONAL | `options.go:78`                                                           |
| Custom retry policy (`WithRetry`)       | FULLY_FUNCTIONAL | `options.go:86`; exponential backoff via failsafe-go                      |
| Custom HTTP client (`WithHTTPClient`)   | FULLY_FUNCTIONAL | `options.go:97`; accepts `Doer` interface (`client.go:27`)                |
| Correlation ID (`WithCorrelationID`)    | FULLY_FUNCTIONAL | `options.go:110`; sets `X-External-Correlation-Id` header on all requests |
| Retry with backoff (429, 5xx, network)  | FULLY_FUNCTIONAL | `client.go:100` `isRetryable`; verified by wise_test.go retry suite       |
| `Authenticate(ctx)`                     | FULLY_FUNCTIONAL | `client.go:110`; delegates to `ListProfiles`                              |
| `Health(ctx)`                           | FULLY_FUNCTIONAL | `client.go:120`; delegates to `Authenticate`                              |
| `WithUserAgent` option (v0.11.0)        | FULLY_FUNCTIONAL | `options.go:81`; custom `User-Agent` on every request, default preserved (BDD) |
| Shared-client concurrency (v0.11.0)     | FULLY_FUNCTIONAL | `wise_test.go:246`; 16 parallel requests, per-request correlation-ID isolation, `-race` |

## Profiles

| Feature                | Status           | Evidence                                                 |
| ---------------------- | ---------------- | -------------------------------------------------------- |
| `ListProfiles(ctx)`    | FULLY_FUNCTIONAL | `profiles.go:31`; BDD-tested                             |
| `GetProfile(ctx, id)`  | FULLY_FUNCTIONAL | `profiles.go:13`; BDD-tested                             |
| `GetMe(ctx)`           | FULLY_FUNCTIONAL | `users.go`; typed `UserID`, 401 + nil-details BDD-tested |
| `GetUser(ctx, id)`     | FULLY_FUNCTIONAL | `users.go`; 404, plain-403, zero-ID BDD-tested           |
| `Profile.UserID` / `Profile.PublicID` (v0.11.0) | FULLY_FUNCTIONAL | `types.go:87-91`; owning user ID + opaque public identifier surfaced from the wire |
| Personal-profile parse | FULLY_FUNCTIONAL | `profiles.go:31` name construction                       |
| Business-profile parse | FULLY_FUNCTIONAL | `profiles.go:31` BusinessName branch                     |

## Balances

| Feature                                    | Status           | Evidence                                                           |
| ------------------------------------------ | ---------------- | ------------------------------------------------------------------ |
| `GetMultiCurrencyAccount(ctx, ProfileID)`  | FULLY_FUNCTIONAL | `account_details.go`; self-RecipientID for top-ups, 404 BDD        |
| `GetBankAccountDetails(ctx, ProfileID)`    | FULLY_FUNCTIONAL | `account_details.go`; LOCAL/INTERNATIONAL receive options, 404 BDD |
| `ListCurrencies(ctx)`                      | FULLY_FUNCTIONAL | `currencies.go`; public reference data, 401 BDD                    |
| `ListBalances(ctx, ProfileID)`             | FULLY_FUNCTIONAL | `balances.go`; sends required `types=STANDARD,SAVINGS` query       |
| `GetBalance(ctx, ProfileID, BalanceID)`    | FULLY_FUNCTIONAL | `balances.go`; direct per-balance endpoint, no filtering           |
| `CreateBalance(ctx, CreateBalanceRequest)` | FULLY_FUNCTIONAL | `balances.go`; savings name enforced client-side                   |
| `GetTotalFunds(ctx, ProfileID, Currency)`  | FULLY_FUNCTIONAL | `balances.go`; Worth + Available as `Money`, 401/404 BDD           |
| Filter visible + non-investment balances   | FULLY_FUNCTIONAL | `balances.go:43`; drops `Visible: false` and invested balances     |
| Fetch hidden or invested balances          | FULLY_FUNCTIONAL | `GetBalance` direct endpoint retrieves them individually           |

## Transactions

| Feature                                                   | Status               | Evidence                                                                             |
| --------------------------------------------------------- | -------------------- | ------------------------------------------------------------------------------------ |
| `ListTransactions(ctx, ListTransactionsRequest)`          | FULLY_FUNCTIONAL     | `transactions.go:15`; BDD-tested                                                     |
| `GetStatement` file download (csv/pdf/xlsx/xml/mt940/qif) | FULLY_FUNCTIONAL     | `transactions.go`; raw bytes, client-side format validation                          |
| `ListTransactionsRequest.Type` filter forwarding          | FULLY_FUNCTIONAL     | `transactions.go:33`; BDD-tested                                                     |
| Request validation (empty currency, inverted date range)  | FULLY_FUNCTIONAL     | `transactions.go:208`; returns `wise.transactions.invalid_request` (`:204`)          |
| Transaction type classification                           | FULLY_FUNCTIONAL     | `transactions.go:177`; CARD_PAYMENT / CARD_REFUND split fixed 2026-07-18             |
| Cross-currency transaction mapping                        | FULLY_FUNCTIONAL     | `transactions.go:135` `mapExchange`; uses transaction currency, not request currency |
| `Transaction.Exchange` (`*TransactionExchange`)           | FULLY_FUNCTIONAL     | `types.go:119`; nil for non-conversion transactions                                  |
| `EndOfStatementBalance` exposure                          | FULLY_FUNCTIONAL     | `types.go:192`; surfaced as `Money` on `ListTransactionsResponse`                    |
| `Transaction.Date` UTC semantics                          | PARTIALLY_FUNCTIONAL | `helpers.go:75` parses as UTC; Wise sends no TZ — documented in field comment        |
| Pagination                                                | PLANNED              | Wise returns all transactions in one response; no endpoint requires it yet           |

## Error handling

| Feature                                           | Status           | Evidence                                                            |
| ------------------------------------------------- | ---------------- | ------------------------------------------------------------------- |
| `APIError` base type                              | FULLY_FUNCTIONAL | `errors.go:42`; embeds into all subtypes; carries `Headers` (`:49`) |
| `RateLimitError` (HTTP 429) with `RetryAfter`     | FULLY_FUNCTIONAL | `errors.go:71`; parses delta-seconds + HTTP-date                    |
| `RateLimitError.RateLimitedBy` (429 header)       | FULLY_FUNCTIONAL | `errors.go:75`; captures `X-Rate-Limited-By` header                 |
| `AuthError` (HTTP 401, 403)                       | FULLY_FUNCTIONAL | `errors.go:103`                                                     |
| `SCAChallengeError` (403 + 2FA headers, v0.6.0)   | FULLY_FUNCTIONAL | `errors.go:116`; `TwoFAApprovalToken()` returns the OTT             |
| `WithSCAApprovalToken` option (v0.6.0)            | FULLY_FUNCTIONAL | `options.go:142`; sends cleared OTT as `x-2fa-approval`             |
| `NotFoundError` (HTTP 404)                        | FULLY_FUNCTIONAL | `errors.go:142`                                                     |
| `ServerError` (HTTP 5xx)                          | FULLY_FUNCTIONAL | `errors.go:151`                                                     |
| `ErrorCode()` / `ErrorFamily()` / `IsRetryable()` | FULLY_FUNCTIONAL | All implement go-error-family interfaces                            |
| `errors.As` matching                              | FULLY_FUNCTIONAL | Demonstrated in README; tested                                      |

## Type system

| Feature                                               | Status           | Evidence                                                                                  |
| ----------------------------------------------------- | ---------------- | ----------------------------------------------------------------------------------------- |
| Branded `ProfileID` / `BalanceID`                     | FULLY_FUNCTIONAL | `ids.go:32,38`; phantom types prevent entity-ID mixing at compile time                    |
| Branded `TransactionID`                               | FULLY_FUNCTIONAL | `ids.go:41`                                                                               |
| `Money` value object (`Cents` + `Currency`)           | FULLY_FUNCTIONAL | `types.go:59`; paired cents/currency makes mismatch unrepresentable                       |
| `Currency` branded type with ISO 4217 validation      | FULLY_FUNCTIONAL | `types.go:39` `NewCurrency`; validates 3-letter uppercase ASCII                           |
| Two-layer raw/result split (`internal/raw` boundary)  | FULLY_FUNCTIONAL | Wire types in `internal/raw/types.go`; parsed types in `types.go`; `helpers.go:47` bridge |
| `ProfileType`, `BalanceType`, `TransactionType` enums | FULLY_FUNCTIONAL | `types.go:139,147,164`                                                                    |
| `InvestmentState` typed enum                          | FULLY_FUNCTIONAL | `types.go:155`; used for balance filtering (`balances.go:43`)                             |
| `DetailType` typed enum + constants                   | FULLY_FUNCTIONAL | `transactions.go:159`; typed filter for `ListTransactionsRequest.Type`                    |
| Enum casing normalization (lowercase SDK values)      | FULLY_FUNCTIONAL | `BalanceType` normalized; `ProfileType`/`TransactionType` already lowercase               |

## Build & tooling

| Feature                                                | Status               | Evidence                                                                                                                                                                                 |
| ------------------------------------------------------ | -------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `go test ./...`                                        | FULLY_FUNCTIONAL     | httptest mocks; no network required                                                                                                                                                      |
| `golangci-lint run`                                    | FULLY_FUNCTIONAL     | 0 issues                                                                                                                                                                                 |
| `.github/workflows/ci.yml` (build/test/lint/vulncheck) | PARTIALLY_FUNCTIONAL | Workflow file maintained (refreshed 2026-09-12, commit `14523ae`) but still `disabled_manually` on GitHub since 2026-07-05; no runs since. `sandbox-live.yml` is active (dispatch-gated) |
| `nix flake check`                                      | FULLY_FUNCTIONAL     | Format + sandboxed test via the `go-standard` module (`checks.test` race + coverage)                                                                                                     |
| `nix fmt` (gofumpt + goimports + nixfmt)               | FULLY_FUNCTIONAL     | `flake.nix` treefmt config                                                                                                                                                               |
| BDD tests via Ginkgo + httptest                        | FULLY_FUNCTIONAL     | `wise_test.go`                                                                                                                                                                           |

## Documentation

| Feature                           | Status           | Evidence                                                                                                      |
| --------------------------------- | ---------------- | ------------------------------------------------------------------------------------------------------------- |
| Godoc examples for the public API | FULLY_FUNCTIONAL | `example_test.go`; 24 `Example*` funcs (compile-only doc examples + 3 runnable) covering every resource group |
| README API reference              | FULLY_FUNCTIONAL | All 16 resources documented with runnable snippets + TOC                                                      |

## API surface expansion (v0.8.0)

| Feature                                 | Status           | Evidence                                                        |
| --------------------------------------- | ---------------- | --------------------------------------------------------------- |
| `GetProfile`                            | FULLY_FUNCTIONAL | `profiles.go:13`                                                |
| `GetExchangeRate`                       | FULLY_FUNCTIONAL | `rates.go:16`                                                   |
| `GetTransfer`                           | FULLY_FUNCTIONAL | `transfers.go:116`                                              |
| Quotes API (create + get)               | FULLY_FUNCTIONAL | `quotes.go`                                                     |
| `Quote.paymentOptions` + `notices`      | FULLY_FUNCTIONAL | `quotes.go` mapper; BDD-tested with fees/delivery/notices       |
| Recipients API (list + get + create)    | FULLY_FUNCTIONAL | `recipients.go`                                                 |
| `CreateTransfer`                        | FULLY_FUNCTIONAL | `transfers.go:190`                                              |
| `Transfer.SourceAccount` (`*BalanceID`) | FULLY_FUNCTIONAL | `types.go:219`; debited balance, nil = omitted on wire          |
| `CancelTransfer`                        | FULLY_FUNCTIONAL | `transfers.go`                                                  |
| `FundTransfer` (balance funding)        | FULLY_FUNCTIONAL | `transfers.go`; typed `FundTransferResult`, corruption mapper   |
| `GetDeliveryEstimate`                   | FULLY_FUNCTIONAL | `delivery_estimates.go`                                         |
| `ValidateTransferRequirements`          | FULLY_FUNCTIONAL | `transfer_requirements.go`                                      |
| `GetQuoteAccountRequirements`           | FULLY_FUNCTIONAL | `quotes.go`; route forms, `Accept-Minor-Version: 1`             |
| `RefreshQuoteAccountRequirements`       | FULLY_FUNCTIONAL | `quotes.go`; two-pass flow, omit-empties partial form, BDD+unit |

## API surface expansion (v0.10.0)

| Feature                           | Status           | Evidence                                                           |
| --------------------------------- | ---------------- | ------------------------------------------------------------------ |
| `GetTransferReceipt` PDF download | FULLY_FUNCTIONAL | `transfers.go`; raw bytes, 404 = `*NotFoundError` (no receipt yet) |
| `GetTransferPayoutInfo` (MT103)   | FULLY_FUNCTIONAL | `transfers.go`; `TransferPayoutInfo`, nil `MT103` for non-SWIFT    |

## Wire format hardening

| Feature                                        | Status           | Evidence                                                  |
| ---------------------------------------------- | ---------------- | --------------------------------------------------------- |
| Tolerant timestamp parsing (4 layouts)         | FULLY_FUNCTIONAL | `helpers.go` `parseWiseTimestamp`; zoneless = UTC         |
| Outgoing query timestamps as UTC `Z` (v0.8.1)  | FULLY_FUNCTIONAL | `helpers.go:148` `formatWiseTimestamp`; Wise 422s offsets |
| `ListBalances` `types` query param (v0.5.3)    | FULLY_FUNCTIONAL | `balances.go:18,23`; live API 400s without it             |
| Mapper errors classified `Corruption` (v0.5.2) | FULLY_FUNCTIONAL | `internal_test.go`; fail-fast instead of retry loops      |

## Webhooks

| Feature                                                                                                         | Status           | Evidence                                                                                                        |
| --------------------------------------------------------------------------------------------------------------- | ---------------- | --------------------------------------------------------------------------------------------------------------- |
| `ParseWebhookPublicKey` (PKIX + PKCS#1 PEM, RSA-only)                                                           | FULLY_FUNCTIONAL | `webhooks.go:33`; garbage/non-RSA input rejected with clear errors                                              |
| `VerifyWebhookSignature` (RSA-SHA256 over raw body)                                                             | FULLY_FUNCTIONAL | `webhooks.go:64`; valid/tampered/wrong-key/malformed/empty/5 MiB tested                                         |
| `HeaderWebhookSignature` / `HeaderDeliveryID` constants                                                         | FULLY_FUNCTIONAL | `webhooks.go:22,24`; delivery-dedup guidance in README Webhooks section                                         |
| Profile webhook subscription CRUD (`Create`/`List`/`Get`/`Delete`, v0.11.0)                                     | FULLY_FUNCTIONAL | `webhooks.go:90,117,149,178`; BDD-tested (happy/400/401/404/204/validation); app-level scope deferred (ROADMAP) |
| `WebhookEventType` open enum (33 documented event constants)                                                    | FULLY_FUNCTIONAL | `types.go`; unknown event types pass through by design                                                          |
| `ParseWebhookEvent` envelope + typed payloads (`TransferStateChange`, `TransferPayoutFailure`, `BalanceCredit`) | FULLY_FUNCTIONAL | `webhooks.go:283+`; unknown-event passthrough + corruption-classified malformed-envelope tests                  |

## SCA one-time tokens (v0.11.0)

| Feature                                                        | Status           | Evidence                                                                                     |
| -------------------------------------------------------------- | ---------------- | -------------------------------------------------------------------------------------------- |
| `GetOTTStatus` (challenge list, validity, action type)         | FULLY_FUNCTIONAL | `ott.go:188`; `/2026Q3/one-time-token` surface, personal API token                           |
| `TriggerOTT` / `VerifyOTT` per phone channel                   | FULLY_FUNCTIONAL | `ott.go:209,235`; sms/whatsapp/voice channels, UPPERCASE challenge types mapped once         |
| `ClearSCAChallenge` convenience loop                           | FULLY_FUNCTIONAL | `ott.go:284`; triggers, prompts via callback, verifies, repeats until every challenge passed |
| `OTTChannel` typed channel enum                                | FULLY_FUNCTIONAL | `ott.go:24`; single lowercase-wire → UPPERCASE-challenge-type mapping                        |
| OTT secrecy: token never in error strings                      | FULLY_FUNCTIONAL | `SCAChallengeError.Error()` reports `token issued: true/false` only; value via `TwoFAApprovalToken()` |

## Deferred (demand-gated, not started)

| Feature                      | Status  | Notes                                          |
| ---------------------------- | ------- | ---------------------------------------------- |
| Service-client sub-structure | PLANNED | Trigger reached: 16 resources (see ROADMAP.md) |
