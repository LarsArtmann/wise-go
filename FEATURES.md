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

| Feature                                 | Status               | Evidence                                                           |
| --------------------------------------- | -------------------- | ------------------------------------------------------------------ |
| `wise.New(apiKey, opts...)` constructor | FULLY_FUNCTIONAL     | `client.go:30`; functional-options pattern                         |
| Bearer-token authentication             | FULLY_FUNCTIONAL     | `client.go:169` `setAuth`                                          |
| Sandbox environment (`WithSandbox`)     | FULLY_FUNCTIONAL     | `options.go:27`; `SandboxURL` const in `types.go:20`               |
| Custom base URL (`WithBaseURL`)         | FULLY_FUNCTIONAL     | `options.go:34`                                                    |
| Custom HTTP timeout (`WithTimeout`)     | FULLY_FUNCTIONAL     | `options.go:41`                                                    |
| Custom retry policy (`WithRetry`)       | FULLY_FUNCTIONAL     | `options.go:49`; exponential backoff via failsafe-go               |
| Custom HTTP client (`WithHTTPClient`)   | PARTIALLY_FUNCTIONAL | `options.go:58`; accepts `*http.Client` not a `Doer` interface     |
| Retry with backoff (429, 5xx, network)  | FULLY_FUNCTIONAL     | `client.go:84` `isRetryable`; verified by wise_test.go retry suite |
| `Authenticate(ctx)`                     | FULLY_FUNCTIONAL     | `client.go:94`; delegates to `ListProfiles`                        |
| `Health(ctx)`                           | FULLY_FUNCTIONAL     | `client.go:104`; delegates to `Authenticate`                       |

## Profiles

| Feature                | Status           | Evidence                             |
| ---------------------- | ---------------- | ------------------------------------ |
| `ListProfiles(ctx)`    | FULLY_FUNCTIONAL | `profiles.go:11`; BDD-tested         |
| Personal-profile parse | FULLY_FUNCTIONAL | `profiles.go:33` name construction   |
| Business-profile parse | FULLY_FUNCTIONAL | `profiles.go:34` BusinessName branch |

## Balances

| Feature                                              | Status           | Evidence                                                             |
| ---------------------------------------------------- | ---------------- | -------------------------------------------------------------------- |
| `ListBalances(ctx, ProfileID)`                       | FULLY_FUNCTIONAL | `balances.go:16`                                                     |
| `GetBalance(ctx, ProfileID, BalanceID)`              | FULLY_FUNCTIONAL | `balances.go:44`; linear-scan over `ListBalances`                    |
| Filter visible + non-investment balances             | FULLY_FUNCTIONAL | `balances.go:28`                                                     |
| Amount conversion to cents (`BalanceAmount.Cents()`) | FULLY_FUNCTIONAL | `types.go:62`; uses `math.Round`                                     |
| Fetch hidden or invested balances                    | BROKEN           | No `WithHiddenBalances` option; Wise exposes no per-balance endpoint |

## Transactions

| Feature                                                  | Status               | Evidence                                                                                 |
| -------------------------------------------------------- | -------------------- | ---------------------------------------------------------------------------------------- |
| `ListTransactions(ctx, ListTransactionsRequest)`         | FULLY_FUNCTIONAL     | `transactions.go:15`; BDD-tested                                                         |
| `ListTransactionsRequest.Type` filter forwarding         | FULLY_FUNCTIONAL     | `transactions.go:32`; BDD-tested (added 2026-07-18)                                      |
| Request validation (empty currency, inverted date range) | FULLY_FUNCTIONAL     | `transactions.go:177`; returns `wise.transactions.invalid_request`                       |
| Transaction type classification                          | FULLY_FUNCTIONAL     | `transactions.go:148`; CARD_PAYMENT / CARD_REFUND split fixed 2026-07-18                 |
| Cross-currency transaction mapping                       | FULLY_FUNCTIONAL     | `transactions.go:122` `mapExchange`; uses transaction currency, not request currency     |
| `Transaction.Exchange` (`*TransactionExchange`)          | FULLY_FUNCTIONAL     | `types.go:154`; nil for non-conversion transactions                                      |
| `Transaction.Date` UTC semantics                         | PARTIALLY_FUNCTIONAL | `helpers.go:44` parses as UTC; Wise sends no TZ — documented in field comment            |
| `ListTransactionsResponse.HasMore`                       | PARTIALLY_FUNCTIONAL | `types.go:227`; always `false`; Wise returns all in one response                         |
| `EndOfStatementBalance` exposure on response             | BROKEN               | Parsed at `types.go:69` but never surfaced; consumer must call `ListBalances` separately |
| Pagination                                               | PLANNED              | No Wise endpoint requires it yet                                                         |

## Error handling

| Feature                                           | Status           | Evidence                                         |
| ------------------------------------------------- | ---------------- | ------------------------------------------------ |
| `APIError` base type                              | FULLY_FUNCTIONAL | `errors.go:26`; embeds into all subtypes         |
| `RateLimitError` (HTTP 429) with `RetryAfter`     | FULLY_FUNCTIONAL | `errors.go:51`; parses delta-seconds + HTTP-date |
| `AuthError` (HTTP 401, 403)                       | FULLY_FUNCTIONAL | `errors.go:76`                                   |
| `NotFoundError` (HTTP 404)                        | FULLY_FUNCTIONAL | `errors.go:85`                                   |
| `ServerError` (HTTP 5xx)                          | FULLY_FUNCTIONAL | `errors.go:94`                                   |
| `ErrorCode()` / `ErrorFamily()` / `IsRetryable()` | FULLY_FUNCTIONAL | All implement go-error-family interfaces         |
| `errors.As` matching                              | FULLY_FUNCTIONAL | Demonstrated in README; tested                   |

## Type system

| Feature                                               | Status           | Evidence                                                                           |
| ----------------------------------------------------- | ---------------- | ---------------------------------------------------------------------------------- |
| Branded `ProfileID` / `BalanceID`                     | FULLY_FUNCTIONAL | `ids.go`; phantom types prevent entity-ID mixing at compile time                   |
| Branded `TransactionID`                               | FULLY_FUNCTIONAL | `ids.go:23`                                                                        |
| Two-layer raw/result split                            | FULLY_FUNCTIONAL | Raw `Profile`/`Balance`/`StatementTransaction` vs parsed result types              |
| `ProfileType`, `BalanceType`, `TransactionType` enums | FULLY_FUNCTIONAL | `types.go:176,184,198`                                                             |
| `InvestmentState` typed enum                          | PLANNED          | Bare string constants at `types.go:192`; data-model review Step 1                  |
| `Money` value object                                  | PLANNED          | Currently paired `XxxCents int64` + `XxxCurrency string`; data-model review Step 3 |
| `Currency` branded type                               | PLANNED          | Currently raw `string`; data-model review Step 3                                   |
| Exported `DetailType` constants                       | PLANNED          | `wiseDetail*` are unexported; data-model review Step 2                             |

## Build & tooling

| Feature                                                | Status               | Evidence                                                          |
| ------------------------------------------------------ | -------------------- | ----------------------------------------------------------------- |
| `go test ./...` (94.8% coverage)                       | FULLY_FUNCTIONAL     | 1.42s; uses httptest; no network required                         |
| `golangci-lint run` (63 linters)                       | FULLY_FUNCTIONAL     | 0 issues at commit 218c2d3 + 2026-07-18 fixes                     |
| `.github/workflows/ci.yml` (build/test/lint/vulncheck) | FULLY_FUNCTIONAL     | Three jobs, 15-min timeout each                                   |
| `nix flake check --no-build`                           | FULLY_FUNCTIONAL     | `flake.nix` added 2026-07-18                                      |
| `nix fmt` (gofumpt + goimports + nixfmt)               | FULLY_FUNCTIONAL     | `flake.nix` treefmt config                                        |
| `nix flake check` (full, with builds)                  | PARTIALLY_FUNCTIONAL | test/lint derivations evaluate; full build not yet verified in CI |
| BDD tests via Ginkgo + httptest                        | FULLY_FUNCTIONAL     | `wise_test.go`; 19 `It` blocks                                    |

## Out of scope (not yet started)

| Feature                          | Status  | Notes                                               |
| -------------------------------- | ------- | --------------------------------------------------- |
| Write operations (transfers)     | PLANNED | No POST/PATCH/DELETE helpers yet                    |
| Recipients API                   | PLANNED | No code                                             |
| Quotes API                       | PLANNED | No code                                             |
| Webhooks                         | PLANNED | No code                                             |
| Service-client sub-structure     | PLANNED | Trigger: resource count > 6-8 (architecture review) |
| Move raw types to `internal/raw` | PLANNED | v1.0; pairs with data-model + naming redesign       |
