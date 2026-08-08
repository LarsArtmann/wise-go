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

| Feature                                 | Status           | Evidence                                                           |
| --------------------------------------- | ---------------- | ------------------------------------------------------------------ |
| `wise.New(apiKey, opts...)` constructor | FULLY_FUNCTIONAL | `client.go:37`; functional-options pattern                         |
| Bearer-token authentication             | FULLY_FUNCTIONAL | `client.go:177` `setAuth`                                          |
| Sandbox environment (`WithSandbox`)     | FULLY_FUNCTIONAL | `options.go:26`; `SandboxURL` const in `types.go:20`               |
| Custom base URL (`WithBaseURL`)         | FULLY_FUNCTIONAL | `options.go:33`                                                    |
| Custom HTTP timeout (`WithTimeout`)     | FULLY_FUNCTIONAL | `options.go:40`                                                    |
| Custom retry policy (`WithRetry`)       | FULLY_FUNCTIONAL | `options.go:48`; exponential backoff via failsafe-go               |
| Custom HTTP client (`WithHTTPClient`)   | FULLY_FUNCTIONAL | `options.go:59`; accepts `Doer` interface (`client.go:24`)         |
| Correlation ID (`WithCorrelationID`)    | FULLY_FUNCTIONAL | `options.go:64`; sets `X-External-Correlation-Id` header on all requests |
| Retry with backoff (429, 5xx, network)  | FULLY_FUNCTIONAL | `client.go:91` `isRetryable`; verified by wise_test.go retry suite |
| `Authenticate(ctx)`                     | FULLY_FUNCTIONAL | `client.go:101`; delegates to `ListProfiles`                       |
| `Health(ctx)`                           | FULLY_FUNCTIONAL | `client.go:111`; delegates to `Authenticate`                       |

## Profiles

| Feature                | Status           | Evidence                             |
| ---------------------- | ---------------- | ------------------------------------ |
| `ListProfiles(ctx)`    | FULLY_FUNCTIONAL | `profiles.go:12`; BDD-tested         |
| Personal-profile parse | FULLY_FUNCTIONAL | `profiles.go:33` name construction   |
| Business-profile parse | FULLY_FUNCTIONAL | `profiles.go:33` BusinessName branch |

## Balances

| Feature                                  | Status           | Evidence                                                             |
| ---------------------------------------- | ---------------- | -------------------------------------------------------------------- |
| `ListBalances(ctx, ProfileID)`           | FULLY_FUNCTIONAL | `balances.go:17`                                                     |
| `GetBalance(ctx, ProfileID, BalanceID)`  | FULLY_FUNCTIONAL | `balances.go:45`; linear-scan over `ListBalances`                    |
| Filter visible + non-investment balances | FULLY_FUNCTIONAL | `balances.go:29`; drops `Visible: false` and invested balances       |
| Fetch hidden or invested balances        | BROKEN           | No `WithHiddenBalances` option; Wise exposes no per-balance endpoint |

## Transactions

| Feature                                                  | Status               | Evidence                                                                             |
| -------------------------------------------------------- | -------------------- | ------------------------------------------------------------------------------------ |
| `ListTransactions(ctx, ListTransactionsRequest)`         | FULLY_FUNCTIONAL     | `transactions.go:16`; BDD-tested                                                     |
| `ListTransactionsRequest.Type` filter forwarding         | FULLY_FUNCTIONAL     | `transactions.go:33`; BDD-tested                                                     |
| Request validation (empty currency, inverted date range) | FULLY_FUNCTIONAL     | `transactions.go:214`; returns `wise.transactions.invalid_request` (`:210`)          |
| Transaction type classification                          | FULLY_FUNCTIONAL     | `transactions.go:183`; CARD_PAYMENT / CARD_REFUND split fixed 2026-07-18             |
| Cross-currency transaction mapping                       | FULLY_FUNCTIONAL     | `transactions.go:141` `mapExchange`; uses transaction currency, not request currency |
| `Transaction.Exchange` (`*TransactionExchange`)          | FULLY_FUNCTIONAL     | `types.go:126`; nil for non-conversion transactions                                  |
| `EndOfStatementBalance` exposure                         | FULLY_FUNCTIONAL     | `types.go:188`; surfaced as `Money` on `ListTransactionsResponse`                    |
| `Transaction.Date` UTC semantics                         | PARTIALLY_FUNCTIONAL | `helpers.go:75` parses as UTC; Wise sends no TZ — documented in field comment        |
| Pagination                                               | PLANNED              | Wise returns all transactions in one response; no endpoint requires it yet           |

## Error handling

| Feature                                           | Status           | Evidence                                         |
| ------------------------------------------------- | ---------------- | ------------------------------------------------ |
| `APIError` base type                              | FULLY_FUNCTIONAL | `errors.go:27`; embeds into all subtypes         |
| `RateLimitError` (HTTP 429) with `RetryAfter`     | FULLY_FUNCTIONAL | `errors.go:52`; parses delta-seconds + HTTP-date |
| `RateLimitError.RateLimitedBy` (429 header)        | FULLY_FUNCTIONAL | `errors.go:55`; captures `X-Rate-Limited-By` header |
| `AuthError` (HTTP 401, 403)                       | FULLY_FUNCTIONAL | `errors.go:78`                                   |
| `NotFoundError` (HTTP 404)                        | FULLY_FUNCTIONAL | `errors.go:87`                                   |
| `ServerError` (HTTP 5xx)                          | FULLY_FUNCTIONAL | `errors.go:96`                                   |
| `ErrorCode()` / `ErrorFamily()` / `IsRetryable()` | FULLY_FUNCTIONAL | All implement go-error-family interfaces         |
| `errors.As` matching                              | FULLY_FUNCTIONAL | Demonstrated in README; tested                   |

## Type system

| Feature                                               | Status           | Evidence                                                                                  |
| ----------------------------------------------------- | ---------------- | ----------------------------------------------------------------------------------------- |
| Branded `ProfileID` / `BalanceID`                     | FULLY_FUNCTIONAL | `ids.go:17,20`; phantom types prevent entity-ID mixing at compile time                    |
| Branded `TransactionID`                               | FULLY_FUNCTIONAL | `ids.go:23`                                                                               |
| `Money` value object (`Cents` + `Currency`)           | FULLY_FUNCTIONAL | `types.go:55`; paired cents/currency makes mismatch unrepresentable                       |
| `Currency` branded type with ISO 4217 validation      | FULLY_FUNCTIONAL | `types.go:33`; `NewCurrency` validates 3-letter uppercase ASCII                           |
| Two-layer raw/result split (`internal/raw` boundary)  | FULLY_FUNCTIONAL | Wire types in `internal/raw/types.go`; parsed types in `types.go`; `helpers.go:17` bridge |
| `ProfileType`, `BalanceType`, `TransactionType` enums | FULLY_FUNCTIONAL | `types.go:135,143,160`                                                                    |
| `InvestmentState` typed enum                          | FULLY_FUNCTIONAL | `types.go:151`; used for balance filtering (`balances.go:29`)                             |
| `DetailType` typed enum + constants                   | FULLY_FUNCTIONAL | `transactions.go:165`; typed filter for `ListTransactionsRequest.Type`                    |
| Enum casing normalization (lowercase SDK values)      | FULLY_FUNCTIONAL | `BalanceType` normalized; `ProfileType`/`TransactionType` already lowercase               |

## Build & tooling

| Feature                                                | Status           | Evidence                                     |
| ------------------------------------------------------ | ---------------- | -------------------------------------------- |
| `go test ./...`                                        | FULLY_FUNCTIONAL | httptest mocks; no network required          |
| `golangci-lint run`                                    | FULLY_FUNCTIONAL | 0 issues                                     |
| `.github/workflows/ci.yml` (build/test/lint/vulncheck) | FULLY_FUNCTIONAL | Three jobs + `nix:` job, 15-min timeout each |
| `nix flake check`                                      | FULLY_FUNCTIONAL | Format + sandboxed test via `buildGoModule`  |
| `nix fmt` (gofumpt + goimports + nixfmt)               | FULLY_FUNCTIONAL | `flake.nix` treefmt config                   |
| BDD tests via Ginkgo + httptest                        | FULLY_FUNCTIONAL | `wise_test.go`                               |

## Out of scope (not yet started)

| Feature                      | Status  | Notes                                          |
| ---------------------------- | ------- | ---------------------------------------------- |
| Write operations (transfers) | PLANNED | No POST/PATCH/DELETE helpers yet               |
| Recipients API               | PLANNED | No code                                        |
| Quotes API                   | PLANNED | No code                                        |
| Exchange rates (`GET /v1/rates`) | PLANNED | Self-contained, high-value; see [API docs study](docs/reviews/2026-08-08_api-docs-study.md) |
| Webhooks                     | PLANNED | No code                                        |
| Statements (CSV/PDF)         | PLANNED | SDK consumes `statement.json` only             |
| Service-client sub-structure | PLANNED | Trigger: resource count > 6-8 (see ROADMAP.md) |
