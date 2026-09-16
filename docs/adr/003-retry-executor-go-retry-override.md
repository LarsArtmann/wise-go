# ADR 003: Retry executor — `go-retry` override of the failsafe-go mandate

**Status:** Proposed (migration pending — see TODO_LIST P2 "Adopt go-retry")
**Deciders:** Lars Artmann
**Context date:** 2026-09-13
**Re-verified:** 2026-09-16 against go-retry v0.6.0 (go-doc API diff + runnable probe)

## Context

The SDK's retry executor is `failsafe-go` (v0.9.7, present since the initial
commit `9327c5e`). The project's general Go library policy (`how-to-golang`
rules) mandates `failsafe-go` for retries — an order of magnitude lower overhead
than the banned `avast/retry-go` — and this SDK complied.

Two policy-legal friction points exist with failsafe-go for THIS repo:

1. **`Retry-After` is not expressible in the policy.** Wise's 429s carry
   `Retry-After` (delta-seconds or HTTP-date), which the SDK parses into
   `RateLimitError.RetryAfter` — but failsafe-go's backoff policy cannot consume
   a per-response delay, so the parsed hint is unused by the executor.
2. **Exhaustion needs a hand-rolled bridge.** failsafe-go's
   `ExceededError` wraps the last response; `doRequest` must unwrap it,
   re-classify via `checkError`, and re-emit the typed error
   (`classifyExhaustedRetries`). The matching is by value (AGENTS.md gotcha),
   which is easy to get wrong in tests.

`github.com/larsartmann/go-retry` (in-house, v0.6.0) addresses both: its
exhaustion error carries the final typed error via `WithCause` (probed
2026-08-21 on v0.4.0, re-probed 2026-09-16 on v0.6.0:
`errors.Is`/`errors.AsType` traverse `ErrExhausted` to the final attempt's
error, so `classifyExhaustedRetries` deletes entirely), and `Config.DelayFunc`
accepts Wise's `Retry-After` directly (a >0 return overrides the backoff for
that attempt; 0 falls through to exponential backoff — probe-confirmed on
v0.6.0).

The v0.4.0 → v0.6.0 delta is purely additive (go-doc API diff, 2026-09-16):
v0.5.0 adds `DoWithValue[T]`/`ResultFunc[T]`, v0.6.0 is test/CI hardening.
`DoWithValue` is a direct fit for `doRequest`'s `(*http.Response, error)` — it
removes the closure-plus-variable dance failsafe's `GetWithExecution` requires.
Context end during a backoff delay returns an error wrapping
`ErrCanceled`/`ErrDeadlineExceeded` that unwraps to the stdlib sentinel and
keeps the last attempt error in the chain.

## Decision (proposed)

Migrate the retry executor from `failsafe-go` to `go-retry`:

- Swap the executor in `client.go`; delete `classifyExhaustedRetries`.
- Feed `Retry-After` (both forms) through `Config.DelayFunc`.
- Keep the `isRetryable` classification and the typed-error contract unchanged —
  the 429 BDD tests are the regression gate.

**Why this overrides the general failsafe-go mandate for this repo:** the
executor is maintained by the owner of this repo, is zero-dependency, and is
the only candidate that expresses Wise's `Retry-After` as first-class policy.
The mandate's rationale (overhead, `avast/retry-go` race conditions) does not
apply against `go-retry`, and the mandate itself names failsafe-go as the
alternative to banned libraries, not as a prohibition on owned tooling.

## Consequences

- One fewer third-party dependency (failsafe-go drops from go.mod).
- `RateLimitError.RetryAfter` becomes operational (honored by the executor), not
  informational.
- The value-vs-pointer `ExceededError` gotcha disappears with
  `classifyExhaustedRetries`.
- If the migration is rejected, this ADR is rejected with it and the status
  flips to Rejected — the failsafe-go executor stays, and `Retry-After` remains
  informational until a failsafe-go-native mechanism exists.
