# Status Report — Deep Review & Quality Improvements

**Date:** 2026-07-05 21:23
**Session:** Full-code-review execution (16 tasks completed)

---

## Metrics Snapshot

| Metric           | Before           | After                     |
| ---------------- | ---------------- | ------------------------- |
| Test coverage    | 86.1%            | **94.8%**                 |
| Test count       | ~30              | **41**                    |
| Lint issues      | 1 (ginkgolinter) | **0**                     |
| Lines of Go code | 1,681            | 2,126                     |
| Source files     | 10               | 11 (+ `internal_test.go`) |

---

## a) FULLY DONE ✓

### Bugs Fixed

1. **`AmountCurrency` currency source bug** (`transactions.go:99`) — `mapTransaction` used the request `Currency` parameter instead of the transaction's own `t.Amount.Currency`. For cross-currency statements (e.g. a USD transaction on an EUR-requested statement), `AmountCurrency` would silently carry the wrong ISO 4217 code. Now uses `t.Amount.Currency` consistently across `AmountCurrency`, `TotalCurrency`, and `RunningBalanceCurrency`.

2. **`RetryAfter` hardcoded to 1 second** (`helpers.go:30`, `errors.go:113`) — `newAPIError` constructed `RateLimitError` with `RetryAfter: time.Second` regardless of what the Wise API returned. Implemented `parseRetryAfter` which handles both HTTP spec forms: delta-seconds (`"120"`) and HTTP-date (RFC1123), falling back to 1s when missing or unparseable. `checkError` now reads `resp.Header.Get("Retry-After")` and threads it through.

3. **False `WithHiddenBalances()` doc comment** (`balances.go:13`) — `ListBalances` documented a `WithHiddenBalances()` option that never existed in the codebase. Rewrote the doc to accurately describe the filtering behavior and confirm there is no way to retrieve hidden/invested balances.

### Dead Code Removed

4. **`WithNow` option + entire `now` chain** — `WithNow` in `options.go`, `config.now` field, `Client.now` field, and `mapTransaction`'s unused `now func() time.Time` parameter. The entire chain was dead: `mapTransaction` never read `now`. Removed across `options.go`, `client.go`, `transactions.go`. Also removed the obsolete `"now is unused"` lint exclusion in `.golangci.yml`.

5. **`UserID` / `UserBrand` / `NewUserID`** (`ids.go`) — Defined and exported but never used anywhere in the codebase (no Wise endpoint returns a standalone user ID through this SDK).

6. **`joinStrings` function** (`errors.go:138`) — Custom string joiner duplicating `strings.Join`. Replaced with `strings.Join`.

7. **`jsonUnmarshal` wrapper** (`helpers.go:14`) — Trivial one-line wrapper around `json.Unmarshal` used in exactly one place. Inlined directly into `parseErrorResponse`.

### Missing Data Surfaced

8. **`Transaction.RunningBalanceCents` + `RunningBalanceCurrency`** (`types.go:151`) — Wise returns `runningBalance` on every transaction but the SDK silently dropped it. Now mapped into the result type.

9. **`Transaction.Exchange` (`*TransactionExchange`)** (`types.go:155`) — Wise returns `exchangeDetails` on conversion transactions. Previously defined as a raw type but never mapped into the result. Now surfaced via `mapExchange`.

### Type Safety Strengthened

10. **`InvestmentState` constants** (`types.go:182`) — `balances.go:26` compared `b.InvestmentState != "NOT_INVESTED"` against a magic string. Added `InvestmentStateNotInvested` and `InvestmentStateInvested` constants and updated the comparison.

11. **Wise detail-type wire constants** (`transactions.go:137`) — `classifyTransactionType` used repeated string literals (`"CARD_PAYMENT"`, `"CARD_REFUND"`, etc.) which triggered `goconst` (3+ occurrences). Extracted to named constants: `wiseDetailCardPayment`, `wiseDetailCardRefund`, `wiseDetailTransfer`, etc.

12. **Error code constants** (`errors.go:16`) — Error type codes (`"wise.api_error"`, etc.) were inline string literals. Extracted to `errorCodeAPI`, `errorCodeRateLimit`, `errorCodeAuth`, `errorCodeNotFound`, `errorCodeServer`. Also added the missing `RateLimitError.ErrorCode()` override (it was inheriting `"wise.api_error"` from the embedded `APIError`).

### Input Validation Added

13. **`ListTransactions.validate()`** (`transactions.go:177`) — Empty `Currency` or `From > To` now returns an `errorfamily.NewRejection` immediately, before any network round-trip. Previously, invalid requests would hit the API and produce opaque errors.

### Test Quality Improved

14. **`ginkgolinter` violation fixed** (`wise_test.go:131`) — `defaultListTxReq` was assigned in spec body instead of `BeforeEach`. Moved to the shared `BeforeEach`.

15. **Structural split brain fixed** (`wise_test.go:634`) — The rate-limit retry test was nested inside `Describe("ListTransactions")` but registered a `/v2/profiles` handler and called `ListProfiles`. Moved to its own top-level `Describe("Retry")` block.

16. **Pure-function unit tests added** (`internal_test.go`) — New internal test file covering:
    - `parseRetryAfter`: empty, zero, positive, garbage, negative, HTTP-date (all parallel)
    - `classifyTransactionType`: all 11 type/amount combinations
    - `BalanceAmount.Cents()`: zero, exact, negative, IEEE 754 rounding, half-up
    - Error classification matrix: rate limit, auth, forbidden, not found, server, generic (asserts `ErrorCode()`, `ErrorFamily()`, `IsRetryable()`)
    - `newAPIError` RetryAfter threading
    - Edge cases: `parseBalanceType` error, `parseWiseDate` error, `mapBalance` error, `mapProfile` error, `NewTransactionID`, `WithHTTPClient`, `mapExchange(nil)`

### Documentation Updated

17. **`AGENTS.md`** — Removed `WithNow` gotcha, added Retry-After header and ListTransactions validation gotchas, added error-code convention, updated dependency versions to v0.6.0, removed UserID from branded-ID list, noted WithHiddenBalances was false.

18. **`README.md`** — Documented new `RunningBalanceCents`, `RunningBalanceCurrency`, and `Exchange` fields. Added validation section for ListTransactions.

19. **`docs/DOMAIN_LANGUAGE.md`** — Added `InvestmentState` and `TransactionExchange` to the value objects glossary.

---

## b) PARTIALLY DONE

### API Coverage

The SDK covers **3 of ~12+ Wise API endpoint groups** (profiles, balances, transactions — all read-only). Write operations (transfers, recipients, quotes, payouts) are not started. This was flagged in README "Project Status" already and is expected for the current development phase.

### ExchangeDetails Raw Type

The raw `ExchangeDetails` struct in `types.go` still has `FromCurrency` and `ToCurrency` string fields that duplicate the currency already embedded in `FromAmount.Currency` / `ToAmount.Currency`. The `mapExchange` function ignores these duplicate fields (correctly), but they remain in the raw type for wire-format fidelity. This is acceptable but a minor smell.

---

## c) NOT STARTED

- **Write operations** (transfers, quotes, payouts, recipients)
- **Webhook support** (Wise webhook signature verification, event parsing)
- **Pagination support** (`HasMore` exists but Wise returns all in one response; fine for now)
- **Context-aware request tracing** (no `context` propagation into retry logging)
- **Structured logging integration** (no `slog` or structured debug output)

---

## d) TOTALLY FUCKED UP

**Nothing.** This session was purely additive and corrective. No regressions introduced. All checks pass:

- Build: PASS
- Vet: PASS
- Lint: 0 issues
- Tests: PASS (race-safe)
- Coverage: 94.8%
- `go mod tidy`: CLEAN
- `govulncheck`: No vulnerabilities found

---

## e) WHAT WE SHOULD IMPROVE

### High Priority

1. **`classifyTransactionType` uses `amount float64`** — The amount sign is checked via raw float64 instead of `totalCents int64`. Float comparison for branching is a latent precision risk. Should compare `totalCents > 0` instead.

2. **`isRetryable` doesn't inspect `RateLimitError.RetryAfter`** — The retry policy retries on status code only. It could use `Retry-After` to inform backoff timing (failsafe-go supports custom delay computation).

3. **No `goerrorfamily.RegisterClassification` for domain types** — The error types implement the interfaces, but they aren't registered with the errorfamily classifier for use with `errorfamily.Classify(err)` from external callers. This limits the library's error-family integration story.

4. **`ProfileType` enum values are lowercase** (`"personal"`, `"business"`) — The wire format from Wise uses uppercase (`"PERSONAL"`, `"BUSINESS"`). The `parseProfileType` function correctly maps these, but the exported enum values don't match either Wise's format or a consistent Go convention. This is a minor API design smell.

5. **`TransactionTypeUnknown` is exported but never returned** — `classifyTransactionType` has a `default` branch that returns `Credit` or `Debit`, never `Unknown`. The constant exists but is unreachable through the SDK.

### Medium Priority

6. **No `GetTransaction` method** — Only `ListTransactions`. Wise may not have a single-transaction endpoint, but this should be explicitly documented.

7. **`ExchangeDetails.Rate` is `float64`** — Exchange rates benefit from arbitrary precision. A `string` or `decimal` type would be safer for rate-sensitive consumers. (Acceptable for now given the SDK's scope.)

8. **`parseWiseDate` uses a hardcoded layout** — The `"2006-01-02 15:04:05"` format string appears inline. Could be a named constant (`wiseDateFormat`).

9. **`ListBalances` makes a network call per `GetBalance`** — `GetBalance` calls `ListBalances` then linear-scans. For callers fetching many balances, this is O(n²). Acceptable given no direct API endpoint, but could be cached or batched by callers.

10. **`Health` is a trivial alias for `Authenticate`** — Not wrong, but `Health` implies a different concern (API reachability vs. credential validity). Consider documenting the semantics or adding a dedicated health endpoint if Wise provides one.

---

## f) Up to 25 Things to Get Done Next

| #  | Priority | Task                                                                                    |
| -- | -------- | --------------------------------------------------------------------------------------- |
| 1  | HIGH     | Replace `classifyTransactionType`'s `amount float64` param with `totalCents int64`      |
| 2  | HIGH     | Wire `Retry-After` into failsafe-go's backoff policy (custom delay computation)         |
| 3  | HIGH     | Register domain error types with `errorfamily.RegisterClassification`                   |
| 4  | MED      | Add a `GetProfile` method (single-profile endpoint) if Wise supports it                 |
| 5  | MED      | Remove or reach `TransactionTypeUnknown` (currently unreachable)                        |
| 6  | MED      | Extract `wiseDateFormat` constant from `parseWiseDate`                                  |
| 7  | MED      | Add integration test with a real Wise sandbox (gated behind build tag or env var)       |
| 8  | MED      | Document `GetBalance` O(n) cost and consider optional in-memory caching                 |
| 9  | LOW      | Add `fmt.Stringer` implementations for `ProfileType`, `BalanceType`, `TransactionType`  |
| 10 | LOW      | Consider `ExchangeDetails.Rate` as a string or decimal type for precision               |
| 11 | LOW      | Add `Profile.UserID` to `ProfileResult` (the raw `Profile.UserID` is currently dropped) |
| 12 | LOW      | Add `Profile.PublicID` to `ProfileResult` (currently dropped)                           |
| 13 | LOW      | Add a `Roundtripper` interface for request/response logging or debugging                |
| 14 | LOW      | Add `context.Context` propagation into failsafe-go retry logging                        |
| 15 | LOW      | Add `Exchange` field to the transaction type classification docs in README              |
| 16 | LOW      | Add benchmarks for hot paths (`Cents()`, `classifyTransactionType`, `mapTransaction`)   |
| 17 | LOW      | Add `Example_` test functions for godoc                                                 |
| 18 | LOW      | Consider `go:generate` for enum string-to-value maps (parseXType functions)             |
| 19 | LOW      | Add `CHANGELOG.md` entry for this session's changes                                     |
| 20 | LOW      | Consider splitting `internal_test.go` by domain (errors_test.go, helpers_test.go)       |
| 21 | LOW      | Add `WithUserAgent` option (currently no custom User-Agent header)                      |
| 22 | LOW      | Add `WithLogger` option for structured debug logging                                    |
| 23 | LOW      | Consider a `Money` type (cents + currency) instead of separate fields                   |
| 24 | LOW      | Add `StatementResponse.EndOfStatementBalance` to `ListTransactionsResponse`             |
| 25 | LOW      | Write operations: transfers, quotes, payouts (major feature work)                       |

---

## g) Top #1 Question

**Should `ProfileType` enum values match Wise's wire format (`"PERSONAL"`/`"BUSINESS"`) or stay lowercase (`"personal"`/`"business"`)?**

The wire format is uppercase. `parseProfileType` maps `"PERSONAL"` → `ProfileTypePersonal` (lowercase). This means the exported enum values don't match what Wise sends. `BalanceType` correctly uses uppercase (`BalanceTypeStandard = "STANDARD"`). This inconsistency is a naming split-brain between the two enum families. I can't determine the right call without knowing whether external consumers depend on the lowercase `ProfileType` values via `errors.As` or API responses — changing them is a breaking change.

---

> **Bottom line:** The codebase is in strong shape after this session. 3 bugs fixed, 4 dead-code paths removed, 2 missing data fields surfaced, type safety strengthened across 3 areas, input validation added, and test coverage pushed to 94.8%. All 7 quality gates green.

---

## Resolution (2026-07-18)

A follow-up review pass on 2026-07-18 (post-v0.2.0, commit `218c2d3` + same-day fixes) re-tracked the action items above into the new living docs. Items not mentioned here remain open as written.

**Moved to `TODO_LIST.md` (short-term, actionable):**

- Item **#5** (`TransactionTypeUnknown` is unreachable) — TODO_LIST P2 / v0.3.0; also called out in `docs/brainstorming/2026-07-18_data-model-review.html`.
- Item **#13** (`Roundtripper` interface for logging) — TODO_LIST P1; reframed in `docs/architecture-understanding/2026-07-18_10-08_*.html` as "accept a `Doer` interface in `WithHTTPClient`," backward-compatible.
- Item **#23** (consider a `Money` type) — TODO_LIST P2 / v0.3.0; the centerpiece of the data-model review. The proposed `Money{Cents int64; Currency Currency}` collapses `Transaction` from 20 fields to 14 and makes cents/currency mismatch unrepresentable.
- Item **#24** (surface `EndOfStatementBalance`) — TODO_LIST P2; will land paired with `Money`.

**Moved to `ROADMAP.md` (long-term):**

- Item **#25** (write operations: transfers, recipients, quotes, payouts) — ROADMAP "Axis 1: Completeness." `client.post` / `patch` / `delete` helpers will mirror the existing `get` / `getWithQuery` shape.

**Resolved elsewhere:**

- Item **#19** (CHANGELOG entry) — `CHANGELOG.md` now contains the full v0.2.0 entry covering this session's changes (currency-source fix, Retry-After parsing, WithNow removal, etc.).
- Item **#15** (Exchange field in README) — present in README's "Amount semantics" subsection.

**New defects found and fixed on 2026-07-18** (discovered during the follow-up full-code-review, not anticipated by this report):

- `classifyTransactionType` grouped `CARD_PAYMENT` and `CARD_REFUND` into one amount-based case branch, contradicting the README contract. A positive-amount `CARD_PAYMENT` was silently classified as `TransactionTypeRefund`. Split into two cases; 2 regression unit tests added.
- `ListTransactions` formatted branded IDs with `%d` directly (not `.Get()`) in two error paths, inconsistent with the rest of the codebase. Normalized.
- `ListTransactionsRequest.Type` filter was forwarded in the URL but never covered by a BDD test. Two tests added (forwarded when set, omitted when unset).
- `Transaction.Date` UTC assumption was undocumented. Doc comment added to `parseWiseDate` and to the `Transaction` type itself.

**Confirmed still open and not yet re-tracked:**

- Item **#1** (replace `amount float64` with `totalCents int64` in classifier) — moot once `Money` lands; the classifier signature will change with it.
- Item **#2** (wire `Retry-After` into failsafe-go backoff) — still open.
- Item **#3** (`errorfamily.RegisterClassification`) — `go-error-family v0.6.1` exposes the API; wise-go does not yet call it. Still open.
- Items **#4, #6, #7, #8, #9, #10, #11, #12, #14, #16, #17, #18, #20, #21, #22** — still open as written; no progress since this report.

**Question (g) resolution path:** the `ProfileType` lowercase vs wire-uppercase casing question is now one instance of the broader enum-casing inconsistency documented in `docs/brainstorming/2026-07-18_data-model-review.html` (Step 4 of its migration roadmap). Decision deferred to v0.3.0 where all enum casings will be normalized together.
