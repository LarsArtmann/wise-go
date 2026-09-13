# ADR 003: Retries via failsafe-go, preserving typed errors at exhaustion

**Status:** Accepted (initial design; codified 2026-09-13)
**Deciders:** Lars Artmann

## Context

A Wise API client must retry transient failures: 429 rate limits (with
`Retry-After`), 5xx, and network errors. It must NOT retry rejections (4xx
validation, auth) — those fail fast by contract.

Two policy constraints applied:

- The project's Go library policy (`how-to-golang` rules) **mandates
  `failsafe-go` for retries** (an order of magnitude lower overhead than
  `avast/retry-go`, which is banned for race conditions).
- The Pareto plan for the 2026-09-13 release suggested an ADR for a "go-retry
  migration … overriding the failsafe-go mandate". **That premise was refuted by
  verification**: the SDK has used `failsafe-go` since the initial commit
  (`9327c5e`), no `retry-go` dependency ever existed, and the mandate already
  points at failsafe-go. There is nothing to migrate; this ADR documents the
  standing decision instead.

## Decision

`failsafe-go` (v0.9.7) with a retry policy driven by an `isRetryable`
classifier:

- Retryable: 429 (honoring `Retry-After`, falling back to exponential backoff),
  5xx, network errors.
- Not retryable: `APIError`/`AuthError`/`NotFoundError` (Rejection family),
  SCA challenges (they need user interaction, not retries).
- **Retry exhaustion preserves typed errors**: when the policy gives up,
  `doRequest` unwraps the `ExceededError`, re-classifies the final response via
  `checkError`, and returns the real typed error (`*RateLimitError` with
  parsed `Retry-After` and `X-Rate-Limited-By`, `*ServerError`, …) — never the
  opaque "retries exceeded" wrapper. Locked in by the 429 BDD tests.

## Consequences

- Consumers see the same error taxonomy whether a failure happened on attempt 1
  or attempt N.
- `retrypolicy.ExceededError` matches by **value**, not pointer (gotcha in
  AGENTS.md) — tests construct `retrypolicy.ExceededError{...}`.
- If the Go policy ever changes its resilience mandate, update this ADR first;
  the `isRetryable` classifier is the only code that would move.
