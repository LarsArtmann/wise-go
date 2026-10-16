# wise-go Architecture Review — samber/do v2 + Health Checks

**Date:** 2026-10-16 19:59 CEST
**Scope:** Is `samber/do` v2 used — and used superbly — in combination with health checks? Is the architecture service-oriented, composable, resilient, and self-healthy?
**Method:** samber-do-best-practices skill (DO-1 → DO-6 lens) + architecture-review rubric; every claim below verified against the working tree this session (see §8).
**Series:** third architecture assessment; extends `2026-07-18_10-08_modularity-and-composability.html` and `docs/modularization/2026-07-18_ASSESSMENT.html` (both 2026-07-18).

---

## Verdict (TL;DR)

| Question asked | Answer |
| --- | --- |
| Are we using samber/do v2? | **No. Zero usage** — not in `go.mod`, not in any `.go` file. |
| Is that a gap? | **No — it is the correct architecture for a library SDK.** Imposing a DI container on consumers would be an anti-pattern and violates the repo's dependency-light policy (ADR 002/003; `samber/lo` and `samber/oops` already declined with recorded rationale). |
| samber/do + health checks "in combination"? | **Belongs to the consumer application's composition root, not this SDK.** wise-go drops into that pattern cleanly: constructor-based, stateless, and `Health(ctx)` is signature-shaped for a one-line adapter. §5 gives the blueprint. |
| Service-oriented, composable, resilient? | **Yes — 4.0 / 4.5 / 4.5 out of 5** (§4). Resource-oriented surface (41 methods, 16 resources), injection via `Doer`/`Logger`/options/context, retry + typed-error + `Retry-After` resilience, conformance-gated. |
| Self-healthy? | **The weakest of the four: 3/5.** One real finding (F-1): `Health(ctx)` is a **credential-validating readiness probe documented as a "reachability" check** — it conflates auth failure with unreachability, consumes a rate-limited authenticated endpoint, and can block for minutes through the retry executor. |
| Overall | **Architecture: sound and honestly engineered.** Not using samber/do *is* the superb decision. One naming/semantics fix (F-1) and two small doc improvements would make the health story equally superb. |

---

## 1. Premise audit — is samber/do used?

Verified this session:

- `go.mod` requires exactly three runtime dependencies — `go-branded-id v0.7.0`, `go-error-family v0.11.0`, `go-retry v0.7.1` — plus test-only `kin-openapi`, `ginkgo`, `gomega`. **No `github.com/samber/do`.**
- `grep` over all `*.go`: **zero references** to `samber/do`.
- The only samber-adjacent history in this repo is *declines*, both recorded in AGENTS.md:
  - `samber/lo` — DECLINED 2026-10-07 with a full 9-site census (dependency-light policy).
  - `samber/oops` — DECLINED; erraudit's `--enforce-samber-oops` dropped as inapplicable.

**Conclusion:** the question's premise does not hold. The review therefore answers the two questions that do: (a) should it be used? (§2) and (b) is the underlying goal — composable wiring plus health checks — achieved? (§3–§5).

## 2. Should wise-go use samber/do? No — and here is the proof

wise-go is a **library** (an anti-corruption-layer client SDK), not an application. The samber/do model — injector, composition root, scopes, lifecycle — is an **application** pattern. Applying the DO-1 → DO-6 checklist counterfactually to a hypothetical integration shows every rule argues against it:

| Rule | Applied to a samber/do-wise-go | Verdict |
| --- | --- | --- |
| DO-1 (`Must*` never in runtime paths) | An SDK cannot control where consumers resolve services; a container invites `MustInvoke` into consumer handlers. The plain `New(apiKey, opts...)` constructor returning a value is the safe contract. | Container hurts |
| DO-2 (every `do.New` needs `Shutdown`) | `Client` owns **no resources** — verified: no goroutines, no tickers, no mutexes, no finalizers, no open handles outliving a request (bodies closed per-request, `client.go:194-202`). An injector around it would have **nothing to shut down** — a fake lifecycle. | Nothing to manage |
| DO-3 (`Override*` only in tests) | N/A — but the value-injection seam (`WithHTTPClient(Doer)`, `WithLogger`) already provides clean test substitution without any container. | Already solved |
| DO-4 (no package-level injector) | A library mandating a container forces every consumer — including non-container apps — through `do.Injector` plumbing. Go libraries must stay framework-neutral. | Container hurts |
| DO-5 (`Invoke` not in loops) | N/A for the SDK; per-request `Invoke` in consumer loops would be a new anti-pattern surface this SDK would be creating. | Container hurts |
| DO-6 (shutdown stays self-contained) | Nothing to coordinate (see DO-2). | Nothing to manage |

Supporting repo-policy evidence — this exact trade has been made three times before:

1. **ADR 002** — flat package, one public surface, deliberate dependency minimization.
2. **ADR 003** — dropped `failsafe-go` (third-party) for in-house `go-retry`; "the SDK is deliberately dependency-light … a third-party dep in every consumer's graph … is the wrong trade" (AGENTS.md, re: samber/lo). A DI container would add a dependency to **every consumer's graph for zero SDK-side benefit** — a strictly worse version of the trade ADR 003 rejected.
3. **Modularization assessment (2026-07-18)** — "Why Staying Single-Module Is Correct"; same logic one level up: stay plain-Go.

**What the SDK does instead (and it is the idiomatic Go equivalent):** constructor + functional options (`options.go`), small consumer-implementable interfaces (`Doer` `client.go:41`, `Logger` `options.go:50`), context-carried per-request data (`WithRequestCorrelationID` `options.go:141`), and compile-time-mismatch-proof branded IDs (`ids.go`). That is composition over container wiring — the skill's own "canonical patterns" reduce to exactly these seams once you remove the container.

## 3. Health-check audit

### 3.1 What exists

Single method, three hops deep:

```
Client.Health(ctx)                client.go:184   "checks if the Wise API is reachable"
  └─ Client.Authenticate(ctx)     client.go:174
       └─ Client.ListProfiles(ctx) profiles.go    authenticated GET /v1/profiles
            └─ retry executor      client.go:329  4 attempts, 100ms→5s backoff, Retry-After honored
```

`Health` is one of the 41 exported methods (39 endpoint + `Authenticate` + `Health`; count gate-checked by `nix run .#doc-verify`). README:245 documents it as "Health check (delegates to Authenticate)".

### 3.2 Finding F-1 — the name and the behavior disagree (P1)

The doc comment says **"reachable"**. The behavior is **credential validation + a full rate-limited API call**. Four concrete consequences:

1. **False negative under throttling.** A 429 (or 5xx-after-retries) makes `Health` return an error although the API is provably reachable — an orchestrator or `samber/do` scope health check would flap a perfectly healthy dependency.
2. **Conflated failure modes.** Bad key (`*AuthError`), unreachable host (transport error), and throttling (`*RateLimitError`) all collapse into "Health failed". The typed errors *carry* the distinction — but the method's name, doc, and README framing never tell the consumer to use it.
3. **Slow probe.** `Health` inherits the retry executor: against a 429-ing endpoint it makes up to 4 attempts (each up to the 30 s HTTP timeout) with up to 3 capped delays — **a "health check" that can block for ~2 minutes**. Health probes should fail fast.
4. **Rate-budget cost.** Every probe spends authenticated, rate-limited quota on `ListProfiles`.

None of this makes `Health` *wrong to ship* — as a "verify my credentials end-to-end" readiness probe it is useful. The defect is that it is **named and documented as reachability**, which it is not.

**Recommended resolution (smallest honest change first):**
- **P1a (docs-only):** rewrite the `Health` doc comment + README line to say what it is: *"validates the API key end-to-end (readiness probe); errors are typed — `*RateLimitError` implies reachability with throttling; not a cheap liveness check (retries up to ~2 min worst case)."*
- **P1b (code, optional):** consider `Health` skipping retries (a fail-fast probe path) — one-line-ish via a `retryConfig` with `MaxAttempts=1` — or a separate `VerifyCredentials` alias if consumers already depend on the retried semantics. Deciding between P1a and P1b is a one-decision task, tracked in TODO_LIST.
- **Liveness probe note:** the vendored OpenAPI snapshot documents **no** dedicated ping/health endpoint (verified by search); a true cheap liveness probe would need an endpoint choice verified against the live API (candidates exist but are unverified in-repo) — demand-gated like the rest of the expansion surface.

### 3.3 The samber/do v2 contract, precisely

Verified against pkg.go.dev `github.com/samber/do/v2` v2.1.0 (not from memory):

- `do.Healthchecker` — `HealthCheck() error`
- `do.HealthcheckerWithContext` — `HealthCheck(context.Context) error`
- Shutdown family — `Shutdowner`, `ShutdownerWithError`, `ShutdownerWithContext`, `ShutdownerWithContextAndError`
- Scope fan-out — `RootScope.HealthCheck() map[string]error`, `RootScope.Shutdown() *ShutdownReport`

**Precision matters here:** `wise.Client.Health(ctx context.Context) error` has the **right signature shape but the wrong method name** to satisfy `do.HealthcheckerWithContext` directly (`HealthCheck` ≠ `Health`). Integration therefore needs a 3-line adapter — shown in §5. There is **no** shutdown story because there is nothing to shut down (§2, DO-2): a consumer registering the client in a `do` scope simply never implements a `Shutdowner`, and that is correct.

### 3.4 Self-health beyond `Health` — the strongest form

An SDK's best self-health strategy is to **have no internal state that can rot**, and wise-go has exactly that:

- **Stateless value:** zero goroutines, timers, locks, finalizers, or background workers in non-test sources (verified by search). Concurrency safety is pinned by the shared-client BDD test (16 parallel requests, per-request correlation-ID isolation, `-race` — `wise_test.go:246`, FEATURES table).
- **Diagnosability:** `RequestLog` observes every attempt including retries and delay decisions (`RetryAfterDelay` entries, `options.go:31-45`); correlation IDs client-wide and per-request (`client.go:414-440`); every error type exposes `ErrorContext()` with pinned promotion semantics (`TestErrorContexts`).
- **Truth gate:** the spec-conformance suite validates every recorded exchange against the vendored OpenAPI 3.1 snapshot (202/202 green at v0.12.0; floors 25 templates / 60 exchanges / 4 statement variants, shuffle-proof `TestMain` guard) — the SDK cannot silently drift from the wire contract it claims.
- **Known adjacent drift (pre-existing, tracked):** the live 2027Q1 spec moved `account-details`; `nix run .#spec-drift` currently exits 1 MISSING until `GetBankAccountDetails` migrates (already in TODO_LIST/AGENTS — cited here because it is exactly the "self-health of the architecture" this review is about).

## 4. The four asked qualities — scored (1–5 rubric)

| Dimension | Score | Evidence |
| --- | --- | --- |
| **Service-oriented** | **4.0** | 41 exported methods across 16 resources, uniform shape (raw type → mapper → client method, ADR 002); resource files do not cross-depend. Half point withheld: the single-package service-client split is deferred with its trigger already reached (15+ resources) — documented, deliberate, but unresolved. |
| **Composable** | **4.5** | Constructor + options; `Doer` and `Logger` seams; context-carried correlation; branded IDs make ID mixing a compile error; two-layer raw/public types keep Wise's wire format from leaking into consumer serialization. No container imposed — the highest form of library composability. |
| **Resilient** | **4.5** | go-retry executor with exponential backoff + jitter; exhaustive typed retryability classification (`client.go:145-171`); `Retry-After` honored via `DelayFunc`, capped at the caller's ceiling; exhaustion preserves typed errors (`errors.AsType` reaches through, ADR 003); `X-Rate-Limited-By` scope surfaced; SCA 403 challenge solved programmatically via OTT endpoints. No circuit breaker — correct for a library and already demand-gated in ROADMAP (`Circuit breaker — only with a consumer demand signal`). |
| **Self-healthy** | **3.0** | Stateless (nothing to rot) + diagnosability + conformance gate = excellent foundations. `Health` exists but its semantics lie (F-1): conflated failure modes, throttle false-negatives, minutes-long worst case, rate-budget spend. One focused fix closes most of the gap. |

**Aggregate: 4.0 / 5** — architecture sound; health-check semantics is the one real finding.

## 5. The correct layering: consumer-side composition root (blueprint)

This is what "samber/do v2 in combination with health checks" should look like around an SDK like wise-go — in the **consumer application**, not in the library:

```go
// Consumer's composition root (their main package) — wise-go stays plain-Go.
injector := do.New()

do.Provide(injector, func(i do.Injector) (*wise.Client, error) {
    return wise.New(os.Getenv("WISE_API_KEY"),
        wise.WithSandbox(),              // or WithBaseURL for tests
        wise.WithLogger(logAdapter),     // RequestLog → their logging stack
        wise.WithHTTPClient(tracedClient), // tracing middleware via Doer
    ), nil
})

// do.HealthcheckerWithContext adapter — required because the SDK method is
// named Health, not HealthCheck (deliberate: no samber/do import in the SDK).
do.ProvideNamed(injector, "wise.health", func(i do.Injector) (do.HealthcheckerWithContext, error) {
    return healthAdapter{client: do.MustInvoke[*wise.Client](i)}, nil // composition root: Must is correct here (DO-1)
})
```

The adapter itself is a two-line type (declared in the same file as the composition root):

```go
type healthAdapter struct{ client *wise.Client }

func (h healthAdapter) HealthCheck(ctx context.Context) error { return h.client.Health(ctx) }
```

Notes:
- This pattern is **already achievable today** — nothing in wise-go blocks it. That is the composability claim of §4 proven by construction.
- Registering the adapter as a *transient* or wrapping retries at the adapter level lets a consumer turn F-1's slow-probe behavior into fail-fast without any SDK change.
- A `Shutdowner` is intentionally absent: the client owns no resources (§2).

**Optional follow-up (P2):** a short "Wire into a DI container / orchestrator health checks" recipe in README (sales-page fit: it sells composability). The snippet above minus the commentary is the content.

## 6. Findings & action roadmap

| ID | Priority | Finding | Action | Effort |
| --- | --- | --- | --- | --- |
| F-1 | **P1** | `Health` semantics: documented as "reachable", behaves as credential-validating readiness probe; throttle false-negatives; ~2 min worst case through retries; rate-budget spend | ~~Decide P1a (redocument doc comment + README, typed-error guidance) vs P1b (fail-fast `Health` or `VerifyCredentials` alias); harvest in TODO_LIST~~ Docs-half done 2026-10-16: honest godoc + README + FEATURES wording shipped, and `Describe("Health")` now pins the typed-error contract (429 → `*RateLimitError`, bad key → `*AuthError`, retry-exhausted 5xx → `*ServerError`). P1b code-half stays a user decision (TODO_LIST P3 Decisions). | S (docs-half done) |
| F-2 | **P2 — FIXED this session** | FEATURES.md "Resilience & operability" evidence anchors drifted +17 lines (client.go comment block) / +7 (WithUserAgent); 8 rows pointed at wrong lines | All 8 anchors corrected and grep-verified against declarations (§8) | S (done) |
| F-3 | **P2 (optional) — DONE 2026-10-16** | The consumer-side samber/do + health-check recipe exists only in this report | ~~Add a short README recipe (§5 snippet), if README scope agrees~~ Shipped as README "Dependency Injection & Health Checks" — compile-verified against v0.13.0 AND master (a first draft with an unparseable if-statement was caught by the scratch-module build, vindicating the snippet rule) | S (done) |
| F-4 | **P3** | No circuit breaker | Concur with existing ROADMAP stance: demand-gated, belongs to consumer middleware for a stateless library — do not build speculatively | — |
| F-5 | **P0 (pre-existing, tracked) — RESOLVED 2026-10-16** | Live 2027Q1 spec moved `account-details`; `spec-drift` exits 1 MISSING | ~~Already TODO_LIST-tracked (endpoint migration)~~ Migrated in `baba096` (`GET /{version}/balances/account-details/{id}`, new ID space, conformance fixtures); live spec-drift re-run 2026-10-16 exits 0 (live 2027Q1, 188 paths, 43 SDK endpoints all match) | M (done) |

**docs-health note:** F-1 and F-3 are the only new actionable items from this review. Both were harvested into `TODO_LIST.md` this session (P3 "Open" for the docs-half + README recipe, P3 "Decisions" for the fail-fast/alias design decision) and are now RESOLVED except F-1's user-gated code-half. Per repo convention this file is a point-in-time snapshot: annotate inline (docs-health ANNOTATE) as items resolve, archive when fully superseded.

## 7. Cross-references (series context)

- `docs/architecture-understanding/2026-07-18_10-08_modularity-and-composability.html` — first structural review (structure/coupling/scalability/composability/service/roadmap).
- `docs/modularization/2026-07-18_ASSESSMENT.html` — single-module verdict, "What Would Change the Verdict", anti-recommendations (sub-packages-per-resource, service interfaces without consumers).
- `docs/adr/002-flat-package-raw-boundary.md` — packaging decision + deferred service-client split (trigger reached, sequencing v1.x).
- `docs/adr/003-retry-executor-go-retry-override.md` — the dependency-light trade this review's §2 extends to DI containers.
- `ROADMAP.md` (circuit breaker), `FEATURES.md` (method count, resilience table), AGENTS.md (samber/lo + samber/oops declines).

## 8. Verification appendix (all commands run this session)

| Check | Command | Result |
| --- | --- | --- |
| samber/do absent | `grep samber/do` (all files) + `go.mod` | 0 hits; deps = 3 runtime (all in-house) + test-only |
| Statelessness | `rg 'go func|Ticker|AfterFunc|sync.(Mutex|WaitGroup|Once)|SetFinalizer|x/time/rate'` over non-test `*.go` | no matches |
| Surface count | `rg '^func \(c \*Client\) [A-Z]'` (non-test) | 41 (matches AGENTS/FEATURES/doc-verify gate) |
| `Health` chain | read `client.go:173-186`, `322-367` | Health → Authenticate → ListProfiles → retry executor |
| samber/do v2 API | pkg.go.dev/github.com/samber/do/v2 (v2.1.0) | `Healthchecker{HealthCheck() error}`, `HealthcheckerWithContext{HealthCheck(context.Context) error}`, 4 Shutdowner variants, `Provide`/`Invoke` families |
| Ping endpoint existence | search vendored `docs/reviews/wise-api-openapi.json` + in-repo docs for ping/health | none documented |
| Prior reviews | read headings/verdicts of both 2026-07-18 HTML reports | cross-referenced §7 |
| F-2 fix | `sed -n` each corrected line + `rg 'func With'`/`func` | all 8 anchors hit their declarations |

---

_Point-in-time review, 2026-10-16 19:59 CEST. Scores reflect the rubric of the architecture-review skill; the samber/do v2 API surface was verified live (§3.3), not recalled._
