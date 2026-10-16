# SUPERB Pareto Plan: The Error System of Excellence

**Created:** 2026-10-16 22:29 CEST · **Repo:** wise-go (master) · **Origin:** direct
directive ("the most superb error system Go ever saw, for this SDK") after the
same-day central error-catalog work (57 named codes, 4 exported sentinels in
`errors.go`), reconciled against `TODO_LIST.md`, `ROADMAP.md`, the vendored
OpenAPI spec, and the 21:45 release-close plan.

**Input inventory:** live-code gap analysis (the structured-envelope discard in
`newAPIError`, ROADMAP:276 `BadRequestError` idea, spec granular-code census) +
declined-disposition ledger → **25 medium tasks / ~110 fine tasks**. Nothing is
executed by this plan; execution follows after approval.

**North star:** every error the SDK returns is **structured, named, centralized,
wrap-proof machine-matchable, and fully documented** — consumers never regex an
error message again.

---

## Step 0 — Errata from read-before-ranking (verified against the live tree)

1. **The central catalog already shipped today** (2026-10-16, this session):
   57 named error-code constants + 4 exported sentinels in `errors.go`, guard
   tests, anchor fixes. This plan BUILDS ON it; nothing re-does it.
2. **The 51% gap is real and verified:** `newAPIError` (`errors.go:267-296`)
   parses Wise's structured envelope (`raw.ErrorResponse{Errors
   []ErrorDetail{Code, Message}}`) and then **discards the structure** — the
   per-error codes survive only inside a flattened `Message` string. Consumers
   cannot branch on Wise's granular codes (`not.valid`, `parameter_missing`,
   `resource_not_found`, `rate_limit_exceeded`, …) without string matching.
3. **`FundingErrorCode` (types.go:323) already types ONE granular surface** —
   the open-enum pattern (typed string, known constants, unknown passes
   through) is proven in-repo. `WiseErrorCode` generalizes it; this is
   consistency, not duplication.
4. **ROADMAP:276 "Typed `BadRequestError`" is a standing design idea** — adopted
   here as M6 (400s currently fall through to bare `*APIError`).
5. **No `X-Request-Id` response header exists in the vendored spec** (grep 0) —
   request-correlation capture is DECLINED, not forgotten: there is nothing
   documented to capture. The client already SENDS
   `X-External-Correlation-Id`.
6. **Declined dispositions stay declined** (Verschlimmbesserung guards): no
   per-method concrete error-return types (erraudit `--enforce-generic-return`,
   ruled 2026-10-16); no `samber/oops`; no naming the ~130 contextual
   `fmt.Errorf("%w")` wraps (the inner error carries the name); no
   `errors.Is`→`AsType` migrations of true sentinel matches (go-error-
   modernization contract). `APIError.Is` for code sentinels is a DIFFERENT,
   legitimate pattern (structured-data matching, like `grpc/status`) and does
   not touch that contract.
7. **raw.ErrorDetail carries only `code` + `message`** (spec-verified) — there
   is no arguments field to expose; `Details` stops there.

---

## Step 1 — Pareto breakdown (1% → 51% · 4% → 64% · 20% → 80% · rest → 100%)

| Tier     | Share          | Delivers | Medium IDs   | Why                                                                                                                                                                                                                                                                                                                                                                                                                      |
| -------- | -------------- | -------- | ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **1%**   | 3 of 25 (~12%) | **51%**  | M1, M2, M3   | Structure + matching core: expose the discarded envelope as `APIError.Details` (M1), type the granular codes as the `WiseErrorCode` open enum (M2), and make curated sentinels match through ANY wrap layer via `APIError.Is` (M3). This single move converts every non-2xx from "flattened string you regex" to "wrap-proof, machine-matchable data" — the thing every consumer (bank-sync first) actually branches on. |
| **4%**   | 3 more (~12%)  | **64%**  | + M4, M5, M6 | Contract hardening + the roadmap's 400 gap: guard tests pin the new declarations (M4), a full status × envelope-shape mapping matrix pins the whole surface (M5), `BadRequestError` makes 400s first-class (M6).                                                                                                                                                                                                         |
| **20%**  | 7 more (~28%)  | **80%**  | + M7–M13     | Trust surface: `wise_code` in `ErrorContext` for log pipelines (M7), the README error-guide decision tree (M8), compile-verified snippets vs pinned tag + master — the bank-sync lesson made mechanical (M9), godoc examples (M10), the error-code reference doc (M11), fuzzing the envelope parser (M12), spec-conformance fixtures (M13).                                                                              |
| **Rest** | 12 more (~48%) | **100%** | M14–M25      | Second-wave sentinels (decision-gated), optional Ginkgo consolidation (decision-gated), SCA/OTT guide cross-links, FEATURES/ROADMAP/AGENTS/CHANGELOG bookkeeping, the bank-sync consumer recipe, the full gate battery, claim-verification sweep, the family-granularity decision memo, skill feedback upstream, and the final no-regression audit against every declined disposition.                                   |

Reading: the 1% is _structure + matchability_; the 4% makes it _pinned_; the
20% makes it _trusted and documented_; the rest is durability, docs bookkeeping,
decisions, and the closing audit.

---

## Step 2 — Comprehensive plan (medium granularity, 30–100 min, ALL 25 tasks)

Sorted by importance/impact/effort/customer-value. 🟡 = decision-gated.

| ID  | Task                                                                                                                                                                                                                                                                                                                                                                                        | Impact | Effort | Output / definition of done                                                                         |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | --------------------------------------------------------------------------------------------------- |
| M1  | `APIError.Details []WiseErrorDetail`: new `WiseErrorDetail{Code WiseErrorCode, Message string}` type; `newAPIError` populates it; `Message` join format unchanged (backward compat)                                                                                                                                                                                                         | 🔥🔥🔥 | 45m    | Consumers read granular codes structurally; existing tests untouched-green                          |
| M2  | `WiseErrorCode` open enum: typed string + documented constants from the spec census (`not.valid`, `invalid_value`, `parameter_missing`, `parameter_invalid`, `missing_value`, `invalid_request`, `forbidden`, `unauthorized`, `resource_not_found`, `operation_not_supported`, `rate_limit_exceeded`, `system_error`, `invalid_file_extension`, `unexpected.error`); unknown passes through | 🔥🔥🔥 | 40m    | Catalog + `TestErrorCodesUniqueAndWellFormed` extension; godoc states open-enum contract            |
| M3  | Curated wave-1 code sentinels (≈10, e.g. `ErrResourceNotFound`, `ErrOperationNotSupported`, `ErrRateLimitExceeded`…) interface-typed in `errors.go` + `APIError.Is(target)` matching any `Details` code + `HasCode(WiseErrorCode) bool` convenience                                                                                                                                         | 🔥🔥🔥 | 60m    | `errors.Is(err, wise.ErrX)` works through arbitrary `fmt.Errorf("%w")` wrap chains (pinned by test) |
| M4  | Guard-test hardening: sentinel-shape guard covers new sentinels; uniqueness test covers `WiseErrorCode` constants; AsType-vs-Is contract pinned (types via AsType, codes via Is)                                                                                                                                                                                                            | 🔥🔥   | 40m    | New declarations cannot silently regress shape or duplicate                                         |
| M5  | Mapping matrix tests: status × envelope-shape (structured / empty body / SCA headers / plain-text funding) → typed error + `Details` + `Is` match + `ErrorContext` keys                                                                                                                                                                                                                     | 🔥🔥   | 70m    | One table pins the entire error surface end-to-end                                                  |
| M6  | `BadRequestError` (ROADMAP:276): 400 → typed Rejection, code `wise.bad_request`, context promotion pinned like Auth/NotFound/Server                                                                                                                                                                                                                                                         | 🔥🔥   | 45m    | 400s first-class; status table + README rows updated                                                |
| M7  | `ErrorContext` gains `wise_code` (first granular code) on APIError; `TestErrorContexts` extended                                                                                                                                                                                                                                                                                            | 🔥     | 35m    | Log pipelines get the granular code without type asserts                                            |
| M8  | README error-guide rewrite: decision tree (AsType for types / Is for sentinels / `HasCode` for uncurated codes), wrap-safety explanation, updated status table, all snippets in md-go-validator-safe shapes                                                                                                                                                                                 | 🔥🔥   | 80m    | A consumer never guesses how to match an error again                                                |
| M9  | Snippet compile-verification: every error snippet built in a scratch module against v0.13.0 (bank-sync's pin) AND master; prefer version-neutral constructs                                                                                                                                                                                                                                 | 🔥🔥   | 30m    | The 2026-10-16 snippet-drift class is mechanically dead                                             |
| M10 | Godoc examples: `Example_errorHandling`, per-sentinel matching examples, `//nolint:testableexamples` per convention; twin-synced with README story                                                                                                                                                                                                                                          | 🔥     | 50m    | pkg.go.dev teaches the system                                                                       |
| M11 | `docs/error-codes.md` reference: every `WiseErrorCode` with wire example, HTTP status context, retry guidance                                                                                                                                                                                                                                                                               | 🔥     | 45m    | Single lookup table for every code the SDK can surface                                              |
| M12 | Fuzz: `FuzzParseErrorResponse` + `FuzzNewAPIError` invariants (never panic; Details↔Message consistency; classification deterministic) with seed corpus in `testdata/fuzz/` (auto-picked-up fileset)                                                                                                                                                                                        | 🔥     | 50m    | Envelope parser hardened against hostile/odd bodies                                                 |
| M13 | Spec-conformance: granular-code fixtures added; `exemptResponseSchema` audit for intentionally-corrupt error fixtures; floors re-checked                                                                                                                                                                                                                                                    | 🔥     | 45m    | Error fixtures validated against the vendored OpenAPI snapshot                                      |
| M14 | SCA/OTT cross-links in the error guide: challenge flow errors, `SCAChallengeError` → OTT endpoints → retry-with-token path                                                                                                                                                                                                                                                                  | 🔥     | 30m    | The SCA section and error guide tell one story                                                      |
| M15 | 🟡 Wave-2 sentinels: remaining census codes shortlist → declarations + tests + catalog (only codes with a consumer branching story)                                                                                                                                                                                                                                                         | 🔥     | 40m    | Sentinel set complete and justified                                                                 |
| M16 | 🟡 Optional Ginkgo consolidation: `Describe("Error surface")` wrapping M5's matrix (BDD-testing skill contract); skip if table tests read better                                                                                                                                                                                                                                            | ○      | 60m    | Decision recorded either way                                                                        |
| M17 | FEATURES.md error rows (Details, WiseErrorCode, sentinels, BadRequestError) + ROADMAP:276 resolution + anchor-check                                                                                                                                                                                                                                                                         | 🔥     | 30m    | Living docs current                                                                                 |
| M18 | AGENTS.md conventions: error-system contract bullet (Is-sentinels vs AsType split, curated-wave policy, catalog completeness duty)                                                                                                                                                                                                                                                          | 🔥     | 30m    | Next session inherits the design                                                                    |
| M19 | CHANGELOG `[Unreleased]` entry + release-notes-check + doc-verify green                                                                                                                                                                                                                                                                                                                     | 🔥     | 25m    | Release-ready record                                                                                |
| M20 | bank-sync consumer recipe: granular-code matching example (top-up on insufficient funds, no-retry on `not.valid`) as a README subsection + cross-repo note                                                                                                                                                                                                                                  | 🔥🔥   | 35m    | The flagship consumer demonstrates the system                                                       |
| M21 | Full gate battery: `go test -race`, golangci-lint, erraudit canonical + tree, doc-verify, anchor-check, `nix flake check`; pre-release chain dry inventory                                                                                                                                                                                                                                  | 🔥🔥   | 40m    | Every gate green on the finished system                                                             |
| M22 | Verification sweep: grep-based claim checks (counts, anchors, "0 remaining literals"), plan errata updated with what shipped                                                                                                                                                                                                                                                                | 🔥     | 30m    | No stale claims survive                                                                             |
| M23 | Family-granularity decision memo: status (not body code) drives `ErrorFamily`; `system_error` body code inside 5xx stays Transient — documented + pinned                                                                                                                                                                                                                                    | ○      | 25m    | The one subtle semantic is written down                                                             |
| M24 | Skill feedback: verified erraudit flag semantics (`--enforce-generic-return` et al.) upstreamed to go-error-modernization's `cli-and-flags.md` via its source repo                                                                                                                                                                                                                          | ○      | 25m    | Fleet learns from this session                                                                      |
| M25 | Final no-regression audit: run the declined-disposition checklist (per-method types absent, wraps unnamed, sentinels interface-typed, catalog central, pinned tests unweakened) + verdict memo in this file's errata                                                                                                                                                                        | 🔥🔥   | 40m    | Proof the system was superb, not verschlimmbessert                                                  |

---

## Step 3 — Fine breakdown (≤12 min per task, ALL ~110 tasks)

Sorted by importance/impact/effort/customer-value within each workstream; IDs
map to their medium task.

### Workstream A — Structure + matching core (the 1%)

| ID   | Fine task                                                           | Est |
| ---- | ------------------------------------------------------------------- | --- |
| F1.1 | `WiseErrorDetail` type + godoc in `errors.go`                       | 10m |
| F1.2 | `Details []WiseErrorDetail` field on `APIError` + godoc             | 10m |
| F1.3 | `newAPIError` populates `Details` (Message join format untouched)   | 12m |
| F1.4 | Build + vet + existing suite green                                  | 8m  |
| F1.5 | Table test: envelope fixture → `Details` content pinned             | 12m |
| F2.1 | Spec census extraction: confirm code list vs live preview reference | 12m |
| F2.2 | `WiseErrorCode` type + constants + open-enum godoc                  | 12m |
| F2.3 | Catalog uniqueness test extension for `WiseErrorCode` values        | 8m  |
| F2.4 | Unknown-code passthrough note + example in godoc                    | 8m  |
| F3.1 | `APIError.Is(target)` implementation (sentinel-code comparison)     | 12m |
| F3.2 | Wave-1 sentinels, interface-typed, central in `errors.go`           | 12m |
| F3.3 | `HasCode(WiseErrorCode) bool` convenience + godoc                   | 8m  |
| F3.4 | Catalog + `TestSentinelsStayInterfaceTyped` extension               | 12m |
| F3.5 | Is-through-wraps tests: `fmt.Errorf`, `WrapCorruption`, subtypes    | 12m |
| F3.6 | Lint + full test battery green                                      | 4m  |

### Workstream B — Contract hardening (the 4% addendum)

| ID   | Fine task                                                 | Est |
| ---- | --------------------------------------------------------- | --- |
| F4.1 | Shape-guard regex covers wave-1 sentinels                 | 8m  |
| F4.2 | Uniqueness/format test for code constants extended        | 8m  |
| F4.3 | AsType-vs-Is contract pin (types ≠ codes, both idiomatic) | 12m |
| F4.4 | No-context-loss / no-secret-in-message assertions         | 12m |
| F5.1 | Matrix fixture design: statuses × body shapes grid        | 12m |
| F5.2 | 2xx/empty-body/nil-header edge rows                       | 10m |
| F5.3 | Per-status typed-mapping rows (401/403/404/429/5xx/SCA)   | 12m |
| F5.4 | Envelope-with-details rows incl. multi-error bodies       | 12m |
| F5.5 | Legacy/malformed-shape rows (plain text, non-JSON)        | 12m |
| F5.6 | Assertion helpers (typed, details, context in one check)  | 12m |
| F6.1 | `BadRequestError` type + `ErrorCode` + family (Rejection) | 10m |
| F6.2 | `newAPIError` 400 case + spec fixture check               | 12m |
| F6.3 | Tests incl. context-promotion pin + matrix row            | 12m |
| F6.4 | README status-table + FEATURES rows                       | 11m |

### Workstream C — Trust surface: context, docs, verification (into the 20%)

| ID    | Fine task                                                  | Est |
| ----- | ---------------------------------------------------------- | --- |
| F7.1  | `wise_code` in `APIError.ErrorContext` (when present)      | 10m |
| F7.2  | `TestErrorContexts` extension rows                         | 12m |
| F7.3  | README context-keys note                                   | 8m  |
| F8.1  | Decision-tree section draft (AsType / Is / HasCode)        | 12m |
| F8.2  | AsType snippet block refresh                               | 12m |
| F8.3  | Sentinel-list + `errors.Is` snippet                        | 12m |
| F8.4  | `WiseErrorCode`/`HasCode` snippet                          | 12m |
| F8.5  | Wrap-safety explanation (`%w` chain preserves `Is`)        | 10m |
| F8.6  | Status table + BadRequest row update                       | 10m |
| F8.7  | All blocks in md-go-validator-safe shapes                  | 12m |
| F9.1  | Scratch module: snippets vs v0.13.0 build                  | 12m |
| F9.2  | Scratch module: snippets vs master build                   | 12m |
| F9.3  | Version-neutral wording fixes where pinned-tag builds fail | 6m  |
| F10.1 | `Example_errorHandling` skeleton + nolint                  | 12m |
| F10.2 | Per-sentinel matching examples                             | 12m |
| F10.3 | `HasCode`/Details examples                                 | 10m |
| F10.4 | Example/README twin story check                            | 12m |
| F11.1 | `docs/error-codes.md` skeleton + census table              | 10m |
| F11.2 | Wire example per code (from spec fixtures)                 | 12m |
| F11.3 | doc-verify link coverage for new doc                       | 10m |
| F11.4 | FEATURES reference row                                     | 8m  |
| F12.1 | `FuzzParseErrorResponse` + seeds                           | 12m |
| F12.2 | Corpus fixtures under `testdata/fuzz/`                     | 12m |
| F12.3 | `FuzzNewAPIError` invariants (no panic; determinism)       | 12m |
| F12.4 | Gate wiring check (flake fuzz picks corpus)                | 10m |
| F13.1 | Error-fixture inventory vs conformance harness             | 12m |
| F13.2 | `exemptResponseSchema` audit for corrupt fixtures          | 12m |
| F13.3 | Granular-code conformance fixtures added                   | 12m |
| F13.4 | Coverage floors re-run + floors note                       | 8m  |

### Workstream D — Durability, bookkeeping, decisions (the rest → 100%)

| ID    | Fine task                                                | Est |
| ----- | -------------------------------------------------------- | --- |
| F14.1 | OTT error-path section in guide                          | 12m |
| F14.2 | SCA section cross-link from error guide                  | 10m |
| F14.3 | anchor-check green after doc moves                       | 8m  |
| F15.1 | 🟡 Wave-2 shortlist decision (consumer-branching test)   | 10m |
| F15.2 | Wave-2 declarations + tests                              | 12m |
| F15.3 | Catalog + docs extension                                 | 12m |
| F16.1 | 🟡 Ginkgo feasibility spike (BDD skill contract)         | 12m |
| F16.2 | 🟡 `Describe("Error surface")` migration I               | 12m |
| F16.3 | 🟡 Migration II + suite integration                      | 12m |
| F17.1 | FEATURES rows (4 new capabilities)                       | 12m |
| F17.2 | ROADMAP:276 resolution edit                              | 10m |
| F17.3 | anchor-check                                             | 8m  |
| F18.1 | AGENTS error-system contract bullet                      | 12m |
| F18.2 | Curated-wave policy sentence                             | 10m |
| F18.3 | Errata cross-links                                       | 8m  |
| F19.1 | CHANGELOG entry                                          | 12m |
| F19.2 | release-notes-check + doc-verify                         | 8m  |
| F20.1 | bank-sync recipe snippet (insufficient funds / no-retry) | 12m |
| F20.2 | Cross-repo note in AGENTS/consumer docs                  | 10m |
| F21.1 | test -race + lint + erraudit battery                     | 10m |
| F21.2 | `nix flake check`                                        | 12m |
| F21.3 | doc-verify + anchor-check                                | 10m |
| F21.4 | Pre-release chain dry inventory                          | 8m  |
| F22.1 | Claim greps (counts, anchors, zero-literals)             | 12m |
| F22.2 | Plan errata updated with shipped hashes                  | 10m |
| F23.1 | Family-granularity decision memo                         | 12m |
| F23.2 | Semantic pin test (body code does not flip family)       | 10m |
| F24.1 | Skill `cli-and-flags.md` verified-rows draft             | 12m |
| F24.2 | Upstream commit via skill source repo                    | 12m |
| F25.1 | Declined-disposition checklist run                       | 12m |
| F25.2 | Verdict memo appended to this plan                       | 12m |

(Fine total: 33 + 15 + 29 + 31 ≈ **108 tasks**; every task ≤ 12 min.)

---

## Step 4 — Execution graph (mermaid)

```mermaid
flowchart TD
    subgraph T1["1% — Structure + matching core (51%)"]
        M2[M2 WiseErrorCode open enum]
        M1[M1 APIError.Details]
        M3[M3 Is-sentinels + HasCode]
        M2 --> M1 --> M3
    end
    subgraph T2["4% — Contract hardening (→64%)"]
        M4[M4 guard tests]
        M5[M5 mapping matrix]
        M6[M6 BadRequestError]
    end
    subgraph T3["20% — Trust surface (→80%)"]
        M7[M7 wise_code context]
        M8[M8 README guide]
        M9[M9 snippet compile-verify]
        M10[M10 godoc examples]
        M11[M11 error-code reference]
        M12[M12 fuzz]
        M13[M13 conformance fixtures]
    end
    subgraph T4["Rest — Durability + decisions (→100%)"]
        M14[M14 SCA/OTT links]
        M15🟡[M15 wave-2 sentinels]
        M16🟡[M16 Ginkgo consolidation]
        M17[M17 FEATURES/ROADMAP]
        M18[M18 AGENTS contract]
        M19[M19 CHANGELOG]
        M20[M20 bank-sync recipe]
        M21[M21 gate battery]
        M22[M22 verification sweep]
        M23[M23 family memo]
        M24[M24 skill feedback]
        M25[M25 no-regression audit]
    end
    M3 --> M4 --> M5
    M6 --> M5
    M5 --> M7
    M1 --> M8 --> M9 --> M10
    M2 --> M11
    M5 --> M12 --> M13
    M8 --> M14
    M2 --> M15🟡
    M5 --> M16🟡
    M6 --> M17 --> M18 --> M19
    M3 --> M20
    M19 --> M21 --> M22 --> M25
    M23 --> M25
    M24 --> M25
```

Critical path: **M2 → M1 → M3 → M4 → M5 → M8 → M21 → M25.**

---

## Verschlimmbesserung guards (binding for every task)

1. **Additive only**: no existing exported signature changes; `Message` join
   format byte-identical; pinned tests may be EXTENDED, never weakened.
2. **Sentinels interface-typed** (`var ErrX error = ...`) per the verified
   contract; wave-1 goes in `errors.go` beside the catalog.
3. **`APIError.Is` matches only curated code sentinels** — it never claims
   type-equivalence, never matches on message text, and stays wrap-delegating
   (the `errors.Is` chain does the walking).
4. **Contextual `fmt.Errorf("%w")` wraps stay anonymous** — inner errors carry
   names; do not "complete" them.
5. **The catalog completeness duty**: any new code/sentinel lands in `errors.go`
   AND `TestErrorCodesUniqueAndWellFormed`/shape guards in the same change.
6. **Snippets**: compiled against v0.13.0 AND master before pasting (the
   bank-sync lesson); version-neutral constructs preferred.
7. **Every task ends green**: `go test ./...` + `golangci-lint run` minimum;
   docs tasks add doc-verify + anchor-check.
8. **Decisions stay decisions**: 🟡 tasks record a verdict, not a drift.

---

## Verification protocol (per medium task)

| Gate     | Command                                                            | Required for      |
| -------- | ------------------------------------------------------------------ | ----------------- |
| Tests    | `go test ./...` (log-redirected exit capture — never piped)        | all code tasks    |
| Lint     | `golangci-lint run`                                                | all code tasks    |
| erraudit | `erraudit lint ./... --type-aware --enforce-coded-errors` + `tree` | error-path tasks  |
| Docs     | `nix run .#doc-verify`, `nix run .#anchor-check`                   | docs tasks        |
| Full     | `nix flake check`                                                  | M1-M3, M5-M6, M21 |
| Snippets | scratch-module `go build` vs v0.13.0 + master                      | M8-M10, M20       |

_Plan is a snapshot; living rows go to `TODO_LIST.md` (added as P1 error-system
epic). Annotate, never rewrite, when bringing this plan current._
