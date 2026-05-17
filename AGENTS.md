# wise-go — AGENTS.md

**Updated:** 2025-05-17

## Gotchas

- **Dual date formats** — Wise uses RFC3339 for profile/balance timestamps but `"2006-01-02 15:04:05"` for transaction dates. Two different parsers (`parseRFC3339` vs `parseWiseDate`) — use the right one.
- **AmountCents vs TotalCents** — `Transaction.AmountCents` is absolute value (always positive). `Transaction.TotalCents` preserves sign (negative for debits). Getting this wrong silently produces wrong data.
- **Balance filtering** — `ListBalances` silently drops `Visible: false` and `InvestmentState != "NOT_INVESTED"` balances. `GetBalance` has no direct API endpoint — it calls `ListBalances` then linear-scans, so it also inherits this filtering.
- **`//nolint:bodyclose`** — `getWithQuery` intentionally defers body close via `responseCloser`. The nolint is required because failsafe-go's executor wraps the response. Don't remove it.
- **`withNow` in options.go** — unexported, currently unused, but exists for clock injection in tests. Don't delete it.
- **No pagination** — Wise returns all transactions in a single response. `HasMore` exists on the response type but is always `false`.

## Conventions

- **Error wrapping** — Use `cockroachdb/errors.Wrap` for wrapping errors from API calls, `fmt.Errorf("context: %w", err)` for mapping/parsing errors. The codebase mixes both; follow the pattern of whichever file you're editing.
- **Error types** — `newAPIError()` in `errors.go` handles all status-code-to-type mapping. Never construct `AuthError`/`NotFoundError` etc. directly outside that function.
