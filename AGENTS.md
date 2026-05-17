# wise-go — AGENTS.md

**Updated:** 2025-05-17

Go SDK for the Wise (TransferWise) API. Single-package library (`package wise`) with no internal sub-packages.

## Commands

```bash
go test ./...                    # Run all tests
go test -v ./...                 # Verbose test output
go build ./...                   # Verify compilation
```

No Makefile, justfile, flake.nix, or CI config exists. No linter config (`.golangci*`) present.

## Architecture

**Two-layer type system** — this is the core design pattern:

1. **Raw API types** (`types.go`) — mirror the Wise JSON wire format exactly. Fields like `CreatedAt string`, `Amount.Value float64`. Never exposed to callers.
2. **Result types** (`types.go`, prefixed with `*Result`) — strongly-typed public types with `time.Time`, `int64` cents, and enum types. These are what SDK consumers receive.
3. **Mapping functions** (`mapProfile`, `mapBalance`, `mapTransaction`) — convert raw → result, handling date parsing, currency conversion to cents, and enum classification.

**Package structure** (flat, all in `package wise`):

| File | Purpose |
|---|---|
| `client.go` | `Client` struct, HTTP transport, retry via failsafe-go, auth header injection |
| `types.go` | All types: raw API structs, result structs, enums, request/response types, constants |
| `options.go` | Functional options (`Option` func) and internal `config` struct |
| `errors.go` | Typed error hierarchy: `APIError` → `RateLimitError`, `AuthError`, `NotFoundError`, `ServerError` |
| `helpers.go` | JSON helpers, date parsers, body reader |
| `profiles.go` | `ListProfiles()` |
| `balances.go` | `ListBalances()`, `GetBalance()` |
| `transactions.go` | `ListTransactions()`, transaction type classifier |
| `wise_test.go` | All tests (Ginkgo/Gomega BDD style) |

`reports/` directory exists but is empty.

## Key Design Decisions

- **Monetary amounts** — Wise API returns `float64` values. SDK converts to `int64` cents via `BalanceAmount.Cents()` using `math.Round` to handle IEEE 754 precision errors. All result types use `int64` for money.
- **Retry** — `failsafe-go` with exponential backoff (100ms–5s, max 3 retries). Retries on network errors, HTTP 429, and 5xx. Body close is handled via `responseCloser` with `//nolint:bodyclose` on the caller.
- **Date formats** — Wise uses two formats: RFC3339 for profile/balance timestamps (`parseRFC3339`), and `"2006-01-02 15:04:05"` for transaction dates (`parseWiseDate`).
- **Balance filtering** — `ListBalances` only returns `Visible: true && InvestmentState == "NOT_INVESTED"`. `GetBalance` delegates to `ListBalances` + linear scan (no single-balance API endpoint).
- **Authentication** — Bearer token in `Authorization` header. `Authenticate()` and `Health()` both call `ListProfiles()` as a validation check.
- **No pagination** — Wise returns all transactions in a single response. `HasMore` field exists on the response but is always `false`.

## Dependencies

| Dependency | Purpose |
|---|---|
| `cockroachdb/errors` | Error wrapping with `errors.Wrap`/`errors.Newf` |
| `failsafe-go/failsafe-go` | Retry policy with backoff |
| `onsi/ginkgo/v2` + `onsi/gomega` | BDD test framework |

Go version: 1.26.2

## Testing

- Framework: Ginkgo v2 + Gomega
- Pattern: `httptest.Server` with `http.ServeMux` for mock API responses
- Client constructed with `WithBaseURL(server.URL)` to point at test server
- All tests in single file `wise_test.go`, package `wise_test` (external test package)
- Tests import `"github.com/larsartmann/wise-go"` (not the local package path)

## Gotchas

- `withNow` option in `options.go` is unexported and currently unused — it's for injecting a clock in tests but nothing uses it yet.
- `classifyTransactionType` uses the raw `amount.Value` (float64) not cents for the `> 0` check — this is intentional for sign detection before cents conversion.
- `AmountCents` in `Transaction` is the **absolute value** (always positive). `TotalCents` preserves the sign (negative for debits).
- Wise sandbox URL is `api.sandbox.transferwise.tech` (not `sandbox.wise.com`) — see `SandboxURL` constant.
- No `ListTransactionsRequest.Type` validation — the `Type` field is passed as a raw query parameter to the Wise API.
