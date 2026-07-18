# wise-go — AGENTS.md

## Gotchas

- **Dual date formats** — Wise uses RFC3339 for profile/balance timestamps but `"2006-01-02 15:04:05"` for transaction dates. Two different parsers (`parseRFC3339` vs `parseWiseDate`) — use the right one.
- **AmountCents vs TotalCents** — `Transaction.AmountCents` is absolute value (always positive). `Transaction.TotalCents` preserves sign (negative for debits). Getting this wrong silently produces wrong data.
- **Balance filtering** — `ListBalances` silently drops `Visible: false` and `InvestmentState != "NOT_INVESTED"` balances. `GetBalance` has no direct API endpoint — it calls `ListBalances` then linear-scans, so it also inherits this filtering. There is no `WithHiddenBalances` option (the old doc comment lied).
- **`//nolint:bodyclose`** — `getWithQuery` intentionally defers body close via `responseCloser`. The nolint is required because failsafe-go's executor wraps the response. Don't remove it.
- **Two-layer type design** — Raw API structs (`Profile`, `Balance`, `StatementTransaction` in `types.go`) intentionally use primitives (`int64`/`string`) to match Wise's JSON wire format. The parsed result types (`ProfileResult`, `BalanceResult`, `Transaction`) convert these into strong types (branded IDs, enums, `time.Time`, `int64` cents). Phantom/strong-id linters flag the raw structs as false positives — don't brand the JSON-decode layer.
- **Branded IDs are the public API** — Every client method taking an ID uses branded types (`ListBalances(ctx, ProfileID)`, `GetBalance(ctx, ProfileID, BalanceID)`, `ListTransactionsRequest{ProfileID, BalanceID}`). Construct with `NewProfileID`/`NewBalanceID`/`NewTransactionID`; unwrap with `.Get()`. Mixing entity IDs is a compile-time error.
- **No pagination** — Wise returns all transactions in a single response. `HasMore` exists on the response type but is always `false`.
- **Retry-After header** — `RateLimitError.RetryAfter` is parsed from the HTTP `Retry-After` header (delta-seconds or HTTP-date), falling back to 1 second.
- **ListTransactions validates before calling the API** — empty currency or `From > To` returns an `errorfamily.Rejection` immediately without a network round-trip.

## Conventions

- **Error handling** — Domain error types (`APIError`, `RateLimitError`, etc.) implement `go-error-family` interfaces (`ErrorCode`, `ErrorFamily`, `ErrorContext`, `IsRetryable`). Error wrapping at call sites uses `fmt.Errorf("context: %w", err)` — the inner error carries the classification. `errorfamily.WrapCorruption` is used specifically for response decode failures. `errorfamily.NewRejection` is used for "not found" and client-side validation errors.
- **Error types** — `newAPIError()` in `errors.go` handles all status-code-to-type mapping. Never construct `AuthError`/`NotFoundError` etc. directly outside that function.
- **Error families** — `APIError`/`AuthError`/`NotFoundError` → Rejection (not retryable). `RateLimitError`/`ServerError` → Transient (retryable).
- **Every error type has its own `ErrorCode()`** — `wise.api_error`, `wise.rate_limit`, `wise.auth`, `wise.not_found`, `wise.server`.

## Dependencies

- **`go-branded-id v0.3.1`** — branded/phantom types for strongly-typed IDs (ProfileID, BalanceID, TransactionID) that prevent mixing different entity IDs at compile time.
- **`go-error-family v0.6.0`** — behavioral error classification with retry decisions, exit codes, and CLI boundary handling. All domain errors implement its interfaces.
- **`failsafe-go v0.9.6`** — retry with exponential backoff. `isRetryable` func decides what gets retried (429, 5xx, network errors).

## Build & Dev

- **Test**: `go test ./...`
- **Lint**: `golangci-lint run`
- **Format**: `gofmt -w .`
