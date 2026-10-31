# wise-go Architecture Recheck — samber/do v2 + Health Checks

**Date:** 2026-10-31 09:33 CET
**Scope:** Re-verify the 2026-10-16 verdicts on the current tree (post-v0.16.0): Is `samber/do` v2 used — and used superbly — with health checks? Is the architecture service-oriented, composable, resilient, and self-healthy?
**Method:** samber-do-best-practices skill (DO-1 → DO-6 lens) + architecture-review rubric; every claim re-verified against the working tree or live upstream docs **this session** (§8).
**Series:** fourth assessment. Extends and re-verifies `2026-10-16_19-59_samber-do-health-architecture-review.md` (the deep original — this file supersedes it on current-state claims, not on reasoning); which extended `2026-07-18_10-08_modularity-and-composability.html` and `docs/modularization/2026-07-18_ASSESSMENT.html`.

---

## Verdict (TL;DR)

| Question asked                              | Answer                                                                                                                                                                                                                                                                                                          |
| ------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Are we using samber/do v2?                  | **Still no. Zero usage** — not in `go.mod`, zero `.go` references (re-verified 2026-10-31).                                                                                                                                                                                                                     |
| Is that a gap?                              | **Still no — it remains the correct architecture for a library SDK.** Nothing in the two releases since the last review changes the DO-1 → DO-6 counterfactual (§3); ADR 002/003 policy unchanged.                                                                                                              |
| samber/do + health checks "in combination"? | **Belongs to the consumer's composition root — and the recipe now ships in README ("Dependency Injection & Health Checks", README.md:841), compile-verified.** New evidence this round: `samber/do`'s own contract says health checks "should perform a quick check" — see F-1 below.                           |
| Service-oriented, composable, resilient?    | **Yes — 4.0 / 4.5 / 4.5 out of 5, unchanged.** Surface grew 41 → 46 public methods across 15 resource files; seams, retry executor, and typed errors unchanged in shape and still excellent.                                                                                                                    |
| Self-healthy?                               | **Improved: 3.0 → 3.5 out of 5.** The honesty gap is closed (honest godoc + README + FEATURES wording + BDD typed-error pin, all shipped 2026-10-16). The remaining gap is capability, not honesty: `Health` still has no fail-fast path, and that decision is **BLOCKED on the user** (TODO_LIST P3, memo D2). |
| Overall                                     | **Architecture: still sound and now more honest.** Zero NEW findings this round — the meta-finding. The single open item (F-1 code-half) was already correctly scoped, memoized, and parked with its owner.                                                                                                     |

---

## 1. Premise re-audit — is samber/do used?

Re-verified this session (commands in §8):

- `go.mod` requires exactly three runtime dependencies — `go-branded-id v0.7.0`, `go-error-family v0.11.0`, `go-retry v0.7.1` — plus test-only (`kin-openapi`, `ginkgo`, `gomega`). **No `github.com/samber/do`.**
- `rg 'samber/do'` over all Go sources: **0 hits**.
- Upstream currency: **v2.1.0 is still the latest release of `samber/do/v2`** (checked live against pkg.go.dev + GitHub releases this session). The API surface the 2026-10-16 report verified — `Healthchecker` / `HealthcheckerWithContext`, the four `Shutdowner*` variants, `RootScope.HealthCheck() map[string]error` — is unchanged. No drift to react to.

**Conclusion:** the premise still does not hold, and the burden of proof remains on adoption. §2–§3 confirm the burden is not met.

## 2. What changed since the 2026-10-16 review

| Delta                                             | Evidence (current tree)                                                                                                                                                                                                                                                                                           |
| ------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Surface: 41 → **46 public `*Client` methods**     | `rg '^func \(c \*Client\) [A-Z]'` = 46 across 15 resource files plus `client.go`; FEATURES.md reconciles precisely: "46 of those documented operations through 43 of the 46 public `*Client` methods — the other three are `Authenticate`/`Health`/`ClearSCAChallenge`" (gate-checked by `nix run .#doc-verify`). |
| F-1 docs-half: **shipped**                        | `Health` godoc (`client.go:196-202`) now says exactly what it is: a readiness probe, not a cheap liveness check, with typed-error interpretation and the ~2-min worst case spelled out. FEATURES row updated (`FEATURES.md:315`).                                                                                 |
| F-1 BDD pin: **shipped**                          | `Describe("Health")` (`wise_test.go:372`) pins the typed-error contract: 429 → `*RateLimitError` (with `RateLimitedBy`), 401 → `*AuthError`, retry-exhausted 5xx → `*ServerError`.                                                                                                                                |
| F-3 consumer recipe: **shipped**                  | README "Dependency Injection & Health Checks" (README.md:841): the `do.New()` / `do.Provide` / `healthAdapter` recipe with the `HealthcheckerWithContext` adapter, compile-verified against pinned tag AND master.                                                                                                |
| F-5 spec drift: **resolved**                      | `account-details` migrated in v0.14.0; live spec-drift exits 0 (43 SDK endpoints all match live 2027Q1).                                                                                                                                                                                                          |
| Releases: **v0.14.0, v0.15.0, v0.16.0 shipped**   | `git tag` + CHANGELOG; error-system v2 (`errors.go` catalog), OTT typing, sandbox live-lane passes landed in between.                                                                                                                                                                                             |
| `Health` behavior itself: **unchanged by design** | Still `Health → Authenticate → ListProfiles` through the retry executor (`client.go:186-204`). The behavior change was deliberately NOT made pending the user decision (§4).                                                                                                                                      |

Everything else the 2026-10-16 report verified is unchanged: statelessness holds (zero goroutines/tickers/mutexes/finalizers in non-test sources — re-verified), the seams (`Doer` `client.go:48`, `Logger` `options.go:50`, `WithRetry` `options.go:103`, `WithHTTPClient` `options.go:114`) are intact, and the retry executor wiring (`MaxAttempts = retries + 1` `client.go:115`, `DelayFunc` honoring `Retry-After` `client.go:119`) is as documented.

## 3. Should wise-go adopt samber/do? Still no

The full DO-1 → DO-6 counterfactual lives in the 2026-10-16 report (§2 there); it does not need re-derivation because none of its inputs moved. Condensed:

- **DO-2 is decisive and re-verified:** the `Client` still owns **no resources** (no goroutines, tickers, locks, finalizers, or handles outliving a request). A container around it would manage a fake lifecycle — nothing to `Shutdown`, so no `Shutdowner`, and that remains correct.
- **DO-4 still argues against:** a library mandating a container forces DI plumbing into every consumer's graph, including non-container apps. Go libraries stay framework-neutral.
- **Policy still binds:** ADR 002 (flat package, minimal surface) and ADR 003 (dependency-light; `failsafe-go` dropped, `samber/lo` declined with a 9-site census, `samber/oops` declined). A DI container is the same trade at larger blast radius: one dependency in every consumer's graph for zero SDK-side benefit.
- **The idiomatic equivalent is already shipped:** constructor + functional options, `Doer`/`Logger` consumer-implementable seams, context-carried per-request correlation, branded IDs. The README recipe proves by construction that a `do`-based consumer needs exactly a two-line adapter — composability without imposition.

## 4. Health-check audit — current state, one sharpened finding

### 4.1 What exists (verified today)

```
Client.Health(ctx)                client.go:203   readiness probe (honest godoc)
  └─ Client.Authenticate(ctx)     client.go:186
       └─ Client.ListProfiles(ctx) profiles.go    authenticated GET /v2/profiles
            └─ retry executor      client.go:115-119, 348   4 attempts, backoff, Retry-After honored
```

- Godoc (client.go:196-202) now states: readiness probe; `*RateLimitError` proves reachability under throttling; `*AuthError` proves reachability with bad credentials; worst case ~4 × 30s + capped delays; "consumers needing a fail-fast probe should call an endpoint directly with a minimal retry policy."
- The typed-error contract is BDD-pinned (§2), so the honesty is not just prose — it is regression-guarded.
- README carries the consumer-side `do.HealthcheckerWithContext` adapter recipe plus a naming note about the optional `HealthCheck(ctx)` alias.

### 4.2 F-1 status — docs-half done; code-half BLOCKED on the user; new upstream evidence sharpens it

**New this round (verified live against pkg.go.dev v2.1.0):** the `samber/do` documentation for `Healthchecker` says the method "should perform a **quick check** to verify the service is healthy," and v2 ships `ErrHealthCheckTimeout` with parallel scope-wide fan-out (`RootScope.HealthCheck()` / `HealthCheckWithContext`). In other words: the container's own contract expectation is a quick probe. wise-go's `Health`, plugged in via the README adapter, is a **deep** probe — worst case ~2 minutes and one authenticated, rate-budget-consuming call per probe. The adapter pattern works (it compiles, it types), but the semantics it forwards are heavier than the interface contract implies. That is not a wise-go defect — it is exactly why the fail-fast question deserves its decision.

**Mitigation available today (verified):** the retry ladder is context-aware — `go-retry` v0.7.1's delay is a `select` on `timer.C` (retry.go:175-181) with cancel/timeout behavior pinned by its own test suite, and `Health` threads the caller's ctx end-to-end. A consumer can therefore bound the probe with a deadline:

```go
import (
	"context"
	"time"
)

probeCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
if err := client.Health(probeCtx); err != nil {
	// *wise.RateLimitError: reachable, throttled. *wise.AuthError: reachable, bad key.
}
```

What remains impossible without an SDK change: a **single-attempt** credential check that does not burn the retry budget and surfaces the same typed errors. That is F-1's code-half, status unchanged since 2026-10-16:

- **BLOCKED: user design decision** — TODO_LIST P3 "Decisions" (`TODO_LIST.md:140`).
- Options + recommendation already memoized in `docs/planning/2026-10-16_decision-memos.md` D2: **(a)** status quo (docs tell consumers to wrap with their own timeout) vs **(b)** additive `VerifyCredentials(ctx)` — one round-trip, same typed errors, no `Health` behavior change. **Memo recommends (b) as a v1.1 additive method**, landing together with the recipe update saying which probe to call when.
- A true cheap liveness probe additionally needs an endpoint choice verified against the live API — still demand-gated (no ping/health endpoint exists in the vendored snapshot).

**This recheck endorses D2-(b) unchanged** and adds one nuance: since `samber/do` scopes run health checks in parallel with an overall timeout sentinel, a fail-fast single-attempt probe is also the shape that composes best with the container the question asks about.

### 4.3 Self-health beyond `Health` — re-verified

- **Stateless value:** zero goroutines/tickers/mutexes/finalizers in non-test sources (search re-run; only test files match). Concurrency safety pinned by the shared-client BDD test (16 parallel requests, `-race`).
- **Diagnosability:** `RequestLog` observes every attempt including delay decisions (`options.go:31`); correlation IDs client-wide + per-request; every error type exposes `ErrorContext()`.
- **Truth gates:** spec-conformance suite validates every recorded exchange against the vendored OpenAPI 3.1 snapshot (floors ≥25 templates / ≥60 exchanges / ≥4 statement variants; actuals 25/180/6 per AGENTS.md); live spec-drift exits 0 since v0.14.0.

## 5. The four asked qualities — re-scored (1–5 rubric)

| Dimension            | 2026-10-16 | Now     | Delta reasoning                                                                                                                                                                                                                                                             |
| -------------------- | ---------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Service-oriented** | 4.0        | **4.0** | Surface grew 41 → 46 methods across 15 resource files, uniform shape (raw type → mapper → client method), resource files still do not cross-depend. The single-package service-client split is still deferred with its trigger reached — documented, deliberate, unchanged. |
| **Composable**       | 4.5        | **4.5** | Same seams, now also proven by the shipped + compile-verified README DI recipe. A consumer composition root needs two lines of adapter — composability demonstrated, not just claimed.                                                                                      |
| **Resilient**        | 4.5        | **4.5** | Retry executor unchanged; typed retryability classification exhaustive; `Retry-After`/`X-Rate-Limited-By` surfaced; exhaustion preserves typed errors; SCA solved programmatically. Circuit breaker still correctly demand-gated (ROADMAP.md:305).                          |
| **Self-healthy**     | 3.0        | **3.5** | +0.5: the honesty gap is closed and regression-pinned (godoc + README + FEATURES + `Describe("Health")`). Still not 4+: the fail-fast capability gap remains open and user-gated (F-1 code-half, §4.2), and no liveness endpoint exists to verify.                          |

**Aggregate: 4.1 / 5** (was 4.0). The improvement is real but earned by documentation and test pinning, not by capability — which is precisely what the remaining decision releases.

## 6. Findings & action roadmap

**This recheck raises zero new findings.** Every item below pre-exists, is already tracked in its right home, and none changed status except by shipping:

| ID       | Status                                    | Finding / next action                                                                                                                                                                                                                                                          |
| -------- | ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| F-1b     | **OPEN — blocked on user** (TODO_LIST P3) | `Health` code-half: adopt memo D2 option (b) — additive `VerifyCredentials(ctx)` (single attempt, same typed errors) in v1.1, plus recipe guidance on which probe to call when. Recheck endorsement: unchanged, slightly sharpened by DO's "quick check" contract note (§4.2). |
| Liveness | OPEN — demand-gated                       | A true cheap liveness probe needs a verified endpoint choice; nothing documented in the vendored spec.                                                                                                                                                                         |
| Split    | OPEN — deliberate deferral (ADR 002)      | Service-client split inside the single package; trigger (15+ resources) reached, sequencing v1.x. Unchanged.                                                                                                                                                                   |
| CB       | CLOSED as demand-gated (ROADMAP.md:305)   | Circuit breaker: only with a consumer demand signal. Concur, again.                                                                                                                                                                                                            |

Per repo convention this file is a point-in-time snapshot: annotate inline (docs-health ANNOTATE) as F-1b resolves; the 2026-10-16 file keeps the full derivation and stays the deep original.

## 7. Cross-references (series context)

- `docs/architecture-understanding/2026-10-16_19-59_samber-do-health-architecture-review.md` — the deep original this recheck re-verifies (DO-1..DO-6 table, full health-chain analysis).
- `docs/architecture-understanding/2026-07-18_10-08_modularity-and-composability.html` + `docs/modularization/2026-07-18_ASSESSMENT.html` — first structural reviews.
- `docs/planning/2026-10-16_decision-memos.md` D2 — the F-1 options memo this recheck endorses.
- `docs/adr/002-flat-package-raw-boundary.md`, `docs/adr/003-retry-executor-go-retry-override.md` — the packaging and dependency-light policies §3 rests on.
- README "Dependency Injection & Health Checks" (README.md:841); FEATURES resilience table; AGENTS.md (samber/lo + samber/oops declines).

## 8. Verification appendix (all commands run this session, 2026-10-31)

| Check                      | Command                                                                | Result                                                                                                                                              |
| -------------------------- | ---------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| samber/do absent           | `cat go.mod` + `rg 'samber/do'` (all Go sources)                       | 0 hits; 3 runtime deps (all in-house LarsArtmann) + test-only                                                                                       |
| samber/do version currency | pkg.go.dev `github.com/samber/do/v2` + GitHub releases (live fetch)    | v2.1.0 still latest; health/shutdown interfaces unchanged; docs quote: health checks "should perform a quick check"; `ErrHealthCheckTimeout` exists |
| Method surface             | `rg -o '^func \(c \*Client\) [A-Z][A-Za-z]*' --no-filename` (non-test) | 46 public methods across 15 resource files plus client.go (matches FEATURES' gate-checked 43-carrying-46-operations reconciliation)                 |
| `Health` chain + godoc     | read `client.go:150-215`                                               | Health (client.go:203) → Authenticate (client.go:186) → ListProfiles; godoc honest (client.go:196-202)                                              |
| Health BDD pin             | `rg 'Describe\("Health'` + read block                                  | wise_test.go:372; typed-error contexts pinned                                                                                                       |
| Statelessness              | `rg 'go func                                                           | Ticker                                                                                                                                              |
| Seams + retry wiring       | `rg 'type (Doer                                                        | Logger                                                                                                                                              |
| Retry honors ctx           | read `go-retry@v0.7.1/retry.go:175-181` + its cancel/timeout tests     | delay is a `select` on `timer.C` (+ ctx cases); cancellation pinned upstream                                                                        |
| README recipe present      | `sed -n '841,940p' README.md`                                          | full `do.New`/`Provide`/`healthAdapter` recipe + naming note                                                                                        |
| F-1 open-item home         | `sed -n '130,160p' TODO_LIST.md` + decision memo D2                    | P3 Decisions, BLOCKED: user decision; D2 recommends additive `VerifyCredentials` v1.1                                                               |
| Circuit-breaker stance     | `rg -i 'circuit' ROADMAP.md`                                           | ROADMAP.md:305 — demand-gated, concur                                                                                                               |

---

_Point-in-time recheck, 2026-10-31 09:33 CEST. Scores follow the architecture-review skill's 1–5 rubric; the samber/do v2 surface was re-verified live this session, not recalled. The standing question this series tracks is F-1b — everything else is green by evidence._
