# Status Report — wise-go

**Date:** 2026-05-17 18:48
**Branch:** master
**Commit:** 290d07b (docs: distill AGENTS.md to non-obvious knowledge only)

---

## Executive Summary

wise-go is an **unofficial Go SDK for the Wise (TransferWise) API**, covering read-only endpoints (profiles, balances, transactions). The project is in early development — core functionality is solid with 33 passing tests, but write operations, webhooks, and many API endpoints are not yet implemented.

The most recent work replaced `cockroachdb/errors` with `go-error-family`, adding behavioral error classification (Rejection/Transient/Corruption) to all domain error types.

---

## A) FULLY DONE

### Core SDK
- [x] **Client with Bearer token auth** — `New(apiKey, ...Option)` with functional options
- [x] **`ListProfiles`** — maps raw API types to strongly-typed `ProfileResult` (int64 IDs, time.Time, ProfileType enum)
- [x] **`ListBalances`** — filters to visible, non-investment balances; returns `BalanceResult` with cents
- [x] **`GetBalance`** — delegates to `ListBalances` + linear scan (no single-balance API endpoint)
- [x] **`ListTransactions`** — full date-range query with transaction type classification (9 types)
- [x] **`Authenticate` / `Health`** — validate API key via ListProfiles
- [x] **Automatic retries** — exponential backoff on 429, 5xx, network errors via `failsafe-go`
- [x] **Sandbox support** — `WithSandbox()` one-line switch

### Error Handling (go-error-family migration)
- [x] **Domain error types implement go-error-family interfaces** — `ErrorCode()`, `ErrorFamily()`, `ErrorContext()`, `IsRetryable()`
- [x] **`APIError`** — base type: `Rejection` family, `wise.api_error` code, status code context
- [x] **`RateLimitError`** — `Transient` family, `IsRetryable() = true`, retry_after context
- [x] **`AuthError`** — `Rejection` family (inherits from APIError), `wise.auth` code
- [x] **`NotFoundError`** — `Rejection` family, `wise.not_found` code
- [x] **`ServerError`** — `Transient` family, `IsRetryable() = true`, `wise.server` code
- [x] **`cockroachdb/errors` removed** — replaced with `fmt.Errorf("%w")` for context wrapping and `errorfamily` constructors for new errors
- [x] **Test file updated** — `cockroachdb/errors.As` → stdlib `errors.As`

### Type System
- [x] **Two-layer type system** — Raw API types (mirror JSON) + Result types (clean Go)
- [x] **Monetary amounts as int64 cents** — `BalanceAmount.Cents()` with `math.Round` for IEEE 754 safety
- [x] **Transaction type classification** — 9-type enum mapped from Wise detail types
- [x] **Profile/Balance type enums** — `ProfileType`, `BalanceType`

### Testing
- [x] **33 BDD tests with Ginkgo/Gomega** — all passing
- [x] **`httptest.Server` mocks** — no network access required
- [x] **Coverage of:** client creation, authentication, profiles (valid/error/unknown type), balances (valid/invisible/investment filtered/API error), GetBalance (found/not found), transactions (CRUD, type classification, amount conversion, date parsing, edge cases, API error, rate limit retry)

### Documentation
- [x] **README.md** — comprehensive with quick start, API reference, error handling guide, design decisions
- [x] **CHANGELOG.md** — updated with actual feature list
- [x] **DOMAIN_LANGUAGE.md** — glossary, entities, value objects, raw vs result type system

### Infrastructure
- [x] **`go-error-family v0.1.1`** dependency added
- [x] **8 indirect dependencies removed** — cockroachdb/errors brought in logtags, redact, sentry-go, gogo/protobuf, kr/pretty, kr/text, pkg/errors, rogpeppe/go-internal
- [x] **`golangci-lint`** — 0 issues
- [x] **Build** — clean

---

## B) PARTIALLY DONE

### Error Handling
- [ ] **`errorfamily.WrapCorruption` only used in one place** — `getWithQuery` for JSON decode. Other error wrapping sites use plain `fmt.Errorf`. Should consider using go-error-family constructors more consistently (or decide it's intentional — inner errors carry classification).
- [ ] **`Authenticate` doesn't use error-family wrapping** — uses `fmt.Errorf("authenticate: %w", err)`. Could use `WrapRejection` for semantic clarity.
- [ ] **No `RegisterClassification` for third-party errors** — `failsafe-go` errors, `net/http` errors are not registered. `Classify(err)` would fall back to `Transient` (default), which is correct for network errors but could be explicit.

### Documentation
- [ ] **AGENTS.md is stale** — still references `cockroachdb/errors` in conventions section, still mentions `withNow` as "unexported and unused" (now exported as `WithNow`). Needs update.
- [ ] **README.md mentions `cockroachdb/errors`** — "Zero dependencies beyond retry — Only failsafe-go for production, cockroachdb/errors for error wrapping" — this is now wrong, it uses go-error-family.

---

## C) NOT STARTED

### API Endpoints (Write Operations)
- [ ] **CreateTransfer** — initiate a transfer between currencies
- [ ] **CreateQuote** — get exchange rate quotes
- [ ] **CreateRecipient** — add recipient bank accounts
- [ ] **CancelTransfer** — cancel a pending transfer
- [ ] **Webhooks** — receive and validate Wise webhook events
- [ ] **Multi-currency account management** — create/close currency balances

### API Endpoints (Read Operations)
- [ ] **GetTransfer** — get transfer by ID
- [ ] **ListTransfers** — list all transfers
- [ ] **GetProfile** — get single profile by ID
- [ ] **Exchange rates** — get current exchange rates
- [ ] **Borderless account details** — get account numbers for a balance

### SDK Features
- [ ] **Request logging/middleware** — observability hook for request/response
- [ ] **Rate limit header parsing** — extract `Retry-After` from 429 responses (currently hardcoded to 1s)
- [ ] **Context timeout propagation** — pass context deadlines to failsafe-go executor
- [ ] **Pagination support** — framework for endpoints that paginate (transactions don't, but others might)
- [ ] **Response caching** — optional cache layer for idempotent reads
- [ ] **Metrics/telemetry** — request count, latency, error rate instrumentation
- [ ] **Circuit breaker** — failsafe-go supports it, not wired up

### Testing
- [ ] **Integration tests** — against real Wise sandbox
- [ ] **Error classification tests** — verify `errorfamily.Classify(err)` returns correct family for each error type
- [ ] **Retry policy tests** — verify backoff timing, max retry enforcement
- [ ] **Edge case tests** — empty responses, malformed JSON, unexpected status codes
- [ ] **Race condition tests** — concurrent client usage
- [ ] **Benchmarks** — measure allocation and latency overhead

### Infrastructure
- [ ] **CI/CD pipeline** — GitHub Actions for build, lint, test
- [ ] **GoReleaser** — automated release with semver tags
- [ ] **golangci-lint config file** — `.golangci.yml` with project-specific rules
- [ ] **Examples directory** — runnable example programs
- [ ] **Go doc** — package-level examples for go.dev documentation

---

## D) TOTALLY FUCKED UP

### Critical Issues
- **None** — Build clean, lint clean, all 33 tests pass. No broken functionality.

### Stale/Incorrect
- **AGENTS.md** — references `cockroachdb/errors` and `withNow` (now `WithNow`). Will be fixed in this commit.
- **README.md line** — mentions `cockroachdb/errors` as a dependency. Will be fixed in this commit.

### Technical Debt
- **`now` parameter unused in `mapTransaction`** — `mapTransaction` takes a `now func() time.Time` parameter but never uses it (gopls warning: `unused parameter: now`). Dead code from an earlier design.
- **`RetryAfter` hardcoded to 1s** — `RateLimitError.RetryAfter` is always `time.Second`. Should parse `Retry-After` header from the 429 response.
- **No typed error for 400 Bad Request** — falls through to generic `APIError`. Could add `BadRequestError` for validation errors.

---

## E) WHAT WE SHOULD IMPROVE

### High Priority
1. **Fix AGENTS.md** — stale references to cockroachdb/errors and withNow (done in this commit)
2. **Fix README.md** — remove cockroachdb/errors mention (done in this commit)
3. **Remove unused `now` parameter** from `mapTransaction` — dead code warning
4. **Parse Retry-After header** from 429 responses instead of hardcoding 1s
5. **Add error classification tests** — verify go-error-family integration works for each error type
6. **Consistent error wrapping strategy** — decide: use errorfamily everywhere or only at domain boundaries

### Medium Priority
7. **Add `BadRequestError`** for 400 responses with validation error details
8. **Add CI/CD** — GitHub Actions for automated build/lint/test on push
9. **Add GoReleaser** — automated versioned releases
10. **Wire up circuit breaker** — failsafe-go supports it, just needs configuration
11. **Add request/response logging hook** — for debugging in production
12. **Examples directory** — runnable example programs for common use cases

### Low Priority
13. **Integration tests against sandbox** — real API validation
14. **Race condition tests** — `go test -race`
15. **Benchmarks** — measure SDK overhead
16. **golangci.yml** — project-specific lint config
17. **Go doc examples** — for go.dev documentation

---

## F) TOP 25 THINGS TO DO NEXT

| # | Priority | Task | Impact | Effort |
|---|----------|------|--------|--------|
| 1 | P0 | Update AGENTS.md (stale references) | High | 5min |
| 2 | P0 | Fix README.md (remove cockroachdb/errors mention) | High | 2min |
| 3 | P0 | Remove unused `now` param from `mapTransaction` | Medium | 5min |
| 4 | P1 | Add error classification tests (go-error-family integration) | High | 30min |
| 5 | P1 | Parse `Retry-After` header from 429 responses | Medium | 15min |
| 6 | P1 | Add `BadRequestError` for 400 responses | Medium | 20min |
| 7 | P1 | Add CI/CD (GitHub Actions) | High | 30min |
| 8 | P1 | Consistent error wrapping strategy (errorfamily everywhere?) | Medium | 30min |
| 9 | P2 | Add GoReleaser config | Medium | 20min |
| 10 | P2 | Wire up failsafe-go circuit breaker | Medium | 30min |
| 11 | P2 | Add request/response logging hook (Option) | Medium | 45min |
| 12 | P2 | Examples directory with runnable programs | Medium | 1hr |
| 13 | P2 | Add `GetProfile(ctx, id)` endpoint | Low | 20min |
| 14 | P2 | Add `ListTransfers` endpoint | High | 2hr |
| 15 | P2 | Add `GetTransfer` endpoint | Medium | 30min |
| 16 | P2 | Add `CreateQuote` endpoint | High | 2hr |
| 17 | P2 | Add `CreateRecipient` endpoint | High | 2hr |
| 18 | P3 | Add `CreateTransfer` endpoint | High | 3hr |
| 19 | P3 | Add webhook validation and parsing | High | 3hr |
| 20 | P3 | Integration tests against Wise sandbox | High | 2hr |
| 21 | P3 | Race condition tests (`go test -race`) | Medium | 30min |
| 22 | P3 | Benchmarks for SDK overhead | Low | 1hr |
| 23 | P3 | Context timeout propagation to failsafe-go | Medium | 30min |
| 24 | P4 | golangci.yml project-specific config | Low | 20min |
| 25 | P4 | Go doc examples for go.dev | Low | 1hr |

---

## G) TOP #1 QUESTION I CANNOT FIGURE OUT MYSELF

**What is the intended scope of this SDK?**

Wise has a large API surface — transfers, quotes, recipients, webhooks, multi-currency accounts, direct debits, etc. The current SDK covers only read-only endpoints (profiles, balances, transactions). I cannot determine:

1. **Is this meant to be a comprehensive SDK** covering all Wise API endpoints, or a focused library for read-only financial data aggregation (e.g., for accounting/budgeting tools)?
2. **Who is the target consumer?** Internal tooling? Open-source community? A specific product (like bank-sync or accounting-pitch-deck)?
3. **Should write operations (transfers, recipients) be added at all?** They require significantly more error handling (2FA, anti-fraud checks, compliance requirements) and change the trust model.

This affects all prioritization decisions — whether to focus on breadth (more endpoints) or depth (observability, robustness, edge cases).

---

## Metrics

| Metric | Value |
|--------|-------|
| Go source files | 8 (+ 1 test file) |
| Lines of Go code | 1,548 total (932 production, 616 test) |
| Test count | 33 specs, all passing |
| Test framework | Ginkgo v2 / Gomega |
| Direct dependencies | 4 (failsafe-go, go-error-family, ginkgo, gomega) |
| Indirect dependencies | 15 |
| Lint issues | 0 |
| Build errors | 0 |
| Go version | 1.26.2 |

---

## File Inventory

| File | Lines | Purpose |
|------|-------|---------|
| `client.go` | 191 | Client struct, HTTP helpers, retry logic |
| `errors.go` | 153 | Domain error types with go-error-family interfaces |
| `types.go` | 209 | Raw API types, result types, enums, request/response types |
| `transactions.go` | 120 | ListTransactions, mapping, type classification |
| `wise_test.go` | 616 | Full BDD test suite |
| `options.go` | 70 | Functional options (WithSandbox, WithBaseURL, etc.) |
| `profiles.go` | 64 | ListProfiles, mapping |
| `balances.go` | 90 | ListBalances, GetBalance, mapping |
| `helpers.go` | 35 | JSON helpers, date parsing, body reading |

---

_Generated by Crush — 2026-05-17_
