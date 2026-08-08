# wise-go — AGENTS.md

## Gotchas

- **Dual date formats** — Wise uses RFC3339 for profile/balance timestamps but `"2006-01-02 15:04:05"` for transaction dates. Two different parsers (`parseRFC3339` vs `parseWiseDate`) — use the right one.
- **Amount vs Total** — `Transaction.Amount.Cents` is absolute value (always positive). `Transaction.Total.Cents` preserves sign (negative for debits). Getting this wrong silently produces wrong data. Both are `Money` fields (cents + currency paired).
- **CARD_PAYMENT vs CARD_REFUND classification** — `CARD_PAYMENT` always maps to `TransactionTypeCard` regardless of amount sign. Only `CARD_REFUND` is amount-dependent (positive → `TransactionTypeRefund`, non-positive → `TransactionTypeCard`). Do not re-group these into one case branch — the README contract requires separate handling.
- **Balance filtering** — `ListBalances` silently drops `Visible: false` and `InvestmentState != "NOT_INVESTED"` balances. `GetBalance` has no direct API endpoint — it calls `ListBalances` then linear-scans, so it also inherits this filtering. There is no `WithHiddenBalances` option (the old doc comment lied).
- **`//nolint:bodyclose`** — `getWithQuery` intentionally defers body close via `responseCloser`. The nolint is required because failsafe-go's executor wraps the response. Don't remove it.
- **Two-layer type design** — Raw API structs (`raw.Profile`, `raw.Balance`, `raw.StatementTransaction` in `internal/raw/types.go`) intentionally use primitives (`int64`/`string`/`float64`) to match Wise's JSON wire format. The parsed result types (`Profile`, `Balance`, `Transaction` in `types.go`) convert these into strong types (branded IDs, enums, `time.Time`, `Money`). Raw types are in `internal/raw` and invisible to consumers.
- **Branded IDs are the public API** — Every client method taking an ID uses branded types (`ListBalances(ctx, ProfileID)`, `GetBalance(ctx, ProfileID, BalanceID)`, `ListTransactionsRequest{ProfileID, BalanceID}`). Construct with `NewProfileID`/`NewBalanceID`/`NewTransactionID`; unwrap with `.Get()`. Mixing entity IDs is a compile-time error.
- **No pagination** — Wise returns all transactions in a single response. `HasMore` was removed in v0.4.0.
- **Retry-After header** — `RateLimitError.RetryAfter` is parsed from the HTTP `Retry-After` header (delta-seconds or HTTP-date), falling back to 1 second.
- **ListTransactions validates before calling the API** — empty currency or `From > To` returns an `errorfamily.Rejection` immediately without a network round-trip.
- **flake.nix must be git-tracked for buildflow** — The buildflow pre-commit hook runs `nix fmt .` which requires `flake.nix` to be in the git index. If you create or modify `flake.nix`, `git add` it before committing or the pre-commit hook fails.
- **GOEXPERIMENT=jsonv2 required** — As of go-branded-id v0.5.1 and go-error-family v0.10.0, both deps import `encoding/json/v2` which requires `GOEXPERIMENT=jsonv2` to build. The flake devShells, buildGoModule checkPhase, and .golangci.yml build-tags are all configured for this. For buildflow: `.buildflow.yml` has an `env:` key that injects `GOEXPERIMENT: jsonv2` into all tool subprocesses (go-fix, test-race, govalid-generate, golangci-lint). `go env -w GOEXPERIMENT=jsonv2` does NOT work here — Nix home-manager symlinks `~/.config/go/env` into the read-only store.
- **Transaction.Date is UTC** — Wise statement dates (`"2006-01-02 15:04:05"`) carry no timezone. `parseWiseDate` interprets them as UTC via `time.Parse`. Callers comparing to local-time values must convert explicitly to avoid off-by-one-day errors at boundaries.
- **buildflow auto-configure is dangerous** — `buildflow --fix` can add 40+ linters (including irrelevant ones like `arangolint`, `clickhouselint`, `depguard`) that produce false positives. If buildflow modifies `.golangci.yml`, review the diff carefully. The project's curated linter list is intentional; do not let buildflow replace it with a generic "enable everything" config.
- **Money + Currency value objects** — All monetary amounts are `Money{Cents int64, Currency Currency}`. `Currency` is a typed string with `NewCurrency(s)` validation (3-letter uppercase ASCII). `ListTransactionsRequest.Currency` is `Currency` (not `string`). Use `wise.Currency("EUR")` for direct construction or `NewCurrency("EUR")` for validated. The `toMoney` helper in `helpers.go` converts `raw.BalanceAmount` to `Money` with currency validation.
- **flake.nix fileset must list all Go files** — The `buildGoModule` check uses an explicit `lib.fileset.unions` list. When adding a new `.go` file (e.g. `internal/raw/types.go`), add it to the fileset in `flake.nix` or the Nix build will fail with "undefined" errors.

## Conventions

- **Error handling** — Domain error types (`APIError`, `RateLimitError`, etc.) implement `go-error-family` interfaces (`ErrorCode`, `ErrorFamily`, `ErrorContext`, `IsRetryable`). Error wrapping at call sites uses `fmt.Errorf("context: %w", err)` — the inner error carries the classification. `errorfamily.WrapCorruption` is used specifically for response decode failures. `errorfamily.NewRejection` is used for "not found" and client-side validation errors.
- **Error types** — `newAPIError()` in `errors.go` handles all status-code-to-type mapping. Never construct `AuthError`/`NotFoundError` etc. directly outside that function.
- **Error families** — `APIError`/`AuthError`/`NotFoundError` → Rejection (not retryable). `RateLimitError`/`ServerError` → Transient (retryable).
- **Every error type has its own `ErrorCode()`** — `wise.api_error`, `wise.rate_limit`, `wise.auth`, `wise.not_found`, `wise.server`.

## Dependencies

- **`go-branded-id v0.5.1`** — branded/phantom types for strongly-typed IDs (ProfileID, BalanceID, TransactionID) that prevent mixing different entity IDs at compile time. v0.5.1 imports `encoding/json/v2` (requires `GOEXPERIMENT=jsonv2`).
- **`go-error-family v0.10.0`** — behavioral error classification with retry decisions, exit codes, and CLI boundary handling. All domain errors implement its interfaces. v0.10.0 imports `encoding/json/v2` (requires `GOEXPERIMENT=jsonv2`).
- **`failsafe-go v0.9.6`** — retry with exponential backoff. `isRetryable` func decides what gets retried (429, 5xx, network errors).

## Build & Dev

- **Test**: `go test ./...`
- **Lint**: `golangci-lint run`
- **Format**: `nix fmt` (gofumpt + goimports + nixfmt via treefmt)
- **Full check**: `nix flake check` (format + sandboxed test via `buildGoModule`)
- **Dev shell**: `nix develop` (provides go, gopls, golangci-lint, go-tools)
