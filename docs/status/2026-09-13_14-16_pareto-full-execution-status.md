# Status: Pareto Full-Execution Run — v0.10 Release Prep, Webhooks, Quality Debt — 2026-09-13 14:16 CEST

> Session scope: the user approved **Full Execution Mode** on the Pareto plan
> (`docs/planning/2026-09-13_12-32_pareto-plan-v0.10-release-webhooks-and-quality-debt.md`,
> 106 fine-grained tasks in 3 tiers). This report covers exactly that run —
> nothing outside it. ~25 daemon commits + one plan commit (`fa315be`) landed
> during the session. Branch: `master`, not pushed (harness rule: no push
> without explicit ask).

---

## a) FULLY DONE

1. **Tier 1% — v1.0 API re-audit (tasks 1.1–1.6).** `go doc -all` inventory:
   33 methods / 92 exported types / 22 const groups at re-audit time. Exact
   lineage pinned via a temporary worktree at the v0.9.0 tag (`e508572`): 31
   methods / 91 types there; the audit doc's own list named 30 of its claimed
   31 (the missing entry was `RefreshQuoteAccountRequirements`); v0.10.0 added
   exactly `GetTransferReceipt` + `GetTransferPayoutInfo` (+ `TransferPayoutInfo`).
   Godoc pass over the 3 post-audit methods: clean. **One stale godoc claim
   found and fixed on sight** (`ListBalances` still said Wise exposes no
   per-balance endpoint — false since v0.9.0's direct-endpoint `GetBalance`).
   Risk register extended with items 8–10 (receipt-404 semantics, MT103
   nilability, two-pass requirements flow). Audit doc carries a re-audit
   section; TODO_LIST/ROADMAP/AGENTS drift references updated.
2. **Tier 1% — v0.10.0 release prep (2.1–2.2).** `[0.10.0]` CHANGELOG verified
   complete against the `e508572..fe896a8` diff; receipt + payout-info BDD
   coverage confirmed (happy/404/zero-ID each). Release notes drafted at
   `docs/releases/v0.10.0-release-notes.md` (v0.9.0 house style), ready for
   `gh release create` once the tag is approved.
3. **Tier 1% — CI auth SOLVED BETTER THAN PLANNED (3.1–3.5).** Discovered all
   three `git+ssh` flake inputs (`go-nix-helpers`, `go-branded-id`,
   `go-error-family`) are **PUBLIC** repos → migrated `flake.nix` to
   `github:` URLs pinned at the exact locked revs. `nix flake check` passed
   with **narHash-identical** inputs (zero content drift) and locked type now
   `github`. Consequence: the nix CI job needs **no deploy keys, no
   `GITHUB_TOKEN`+`insteadOf`, no secrets at all** — plan task 3.2 (collect
   secrets) dissolved. Decision documented in `ci.yml` (nix job comment) and
   AGENTS.md (Build & Dev).
4. **Tier 1% — govulncheck evidence (5.1–5.2).** govulncheck 1.8.0 on
   go1.26.7: **"No vulnerabilities found."** The 4 historical reachable
   findings (GO-2026-6218/-6090/-5972/-5026) are cleared by the 1.26.6+
   stdlib fixes; recorded in TODO_LIST.
5. **Tier 1% — exhaustruct_v5 migration (6.1–6.3).** All 5 `.golangci.yml`
   references + the `transfers.go` nolint renamed; CI lint pin bumped
   v2.12 → v2.13 (v2.12 does not know the v5 linter). **Lint: 0 issues.**
   Full gates green.
6. **Tier 4% — webhook subscription types (7.1–7.7).** Spec-verified:
   subscription IDs are **UUID strings** (→ `WebhookSubscriptionID =
   id.ID[WebhookSubscriptionBrand, string]`), request = `name` + `trigger_on`
   - `delivery{version,url}`. Shipped: `internal/raw` wire types
     (Subscription/Delivery/Creator/Scope), public `WebhookSubscription`,
     `WebhookDelivery`, `WebhookCreator(+Type)`, `WebhookScope(+Domain)`,
     `CreateWebhookSubscriptionRequest` with client-side validation
     (HTTPS-only URL, all required fields) and `toWire()`, plus the
     `WebhookEventType` **open enum with 33 documented constants** (union of
     the OpenAPI spec's references and the live webhook-event page). Flake
     fileset updated; build green.
7. **Tier 4% — profile webhook CRUD (8.1–8.5).** Four client methods —
   `CreateProfileWebhookSubscription`, `ListProfileWebhookSubscriptions`,
   `GetProfileWebhookSubscription`, `DeleteProfileWebhookSubscription` (204)
   — against the **quarterly versioned surface
   `/2026Q3/profiles/{profileId}/subscriptions`** (path verified against both
   the spec's server URL and the live docs; first non-`/vN` prefix in the
   SDK, constant `webhookAPIVersion`). A `delete` client helper was added
   (did not exist). Mapper `toWebhookSubscription` classifies unparseable
   timestamps as corruption.
8. **Tier 4% — BDD suite for the four methods (9.1–9.5).** Ginkgo specs
   covering: create happy path **with wire-body assertions** (snake_case keys)
   - 400 + 401; the five-case client-side validation matrix (missing
     name/triggerOn/version, non-HTTPS URL, zero profile ID — no network call);
     list (two subscriptions ordered + empty + corrupt-`created_at` corruption +
     401 + zero ID); get (happy + 404 + both zero-ID rejections); delete (204 +
     404 + zero-ID).
9. **Tier 4% — typed webhook event decoding (12.1–12.6, 14.1–14.3).**
   `WebhookEvent` envelope (schema version, subscription ID, event type,
   tolerant `sent_at` via `parseWiseTimestamp`, `jsontext.Value` raw data),
   `ParseWebhookEvent` with corruption classification (undecodable JSON,
   missing `event_type`, unparseable `sent_at`), and per-type accessors.
   **Unknown event types parse successfully** (forward compatibility, tested);
   malformed envelopes/payloads classify as corruption (tested); tolerant
   timestamp layouts tested. 14.3 decided: **composition, not a
   `VerifyAndParse` helper** — documented in `ParseWebhookEvent` godoc
   (verification failure is a transport-level 401/403, not a parse error).
10. **Tier 4% — typed payloads with a verified substitution (13.1–13.5).**
    The plan's `deposits#completed`/`deposits#top-up-failed` are **not
    documented** on the live webhook-event reference (verified by fetch) —
    dropped per verify-external-claims and replaced with
    `transfers#payout-failure` (schema 5.0.0, fetched field-by-field; Wise
    explicitly recommends processing it alongside state-change) and
    `balances#credit` (verified). Shipped: `TransferStateChangeData`
    (reuses `TransferStatus`), `TransferPayoutFailureData` (branded
    `TransferID`/`ProfileID`), `BalanceCreditData` (Money conversion via the
    existing `toMoney`), shared `WebhookResource`. Per-type fixtures from the
    live docs.
11. **Tier 4% — docs + corrections (15.1–15.4, 16.1–16.4).** Two godoc
    examples (subscription create/delete round-trip; verify → parse →
    switch-dispatch with BLOCKED-style handling), both compile-only with the
    required nolint. README: "Webhook subscriptions" + "Typed event
    decoding" sections + TOC entries; the "decode the event JSON and process
    it" hand-wave replaced. Plan item #49 annotated (~~8~~ 9 operations).
    FEATURES rows flipped to FULLY_FUNCTIONAL; CHANGELOG `[Unreleased]`
    Added entries; TODO_LIST P3 checked off; AGENTS webhook gotcha added
    (2026Q3 surface, snake_case wire, UUID IDs, tagliatelle exclusion
    mechanics, deposits#-events-not-documented finding). Full gate set green
    (race + lint 0 + `nix flake check`, which includes the README links
    check).
12. **Tier 20% — dedup follow-ups (17.1–17.4, 17.6 verdict, 17.7).**
    `TestRequireID` pins the contract exactly
    (`[rejection:wise.<domain>.invalid_request] <field> is required`) across
    int64 + string IDs; `fetchByID` now takes the branded ID and routes
    through `requireID` (4 callers updated; messages unchanged for int64).
    **17.5 verification REFUTED the plan**: the create-transfer schema's
    `details` accept exactly the five keys `CreateTransferRequest` already
    carries — `sourceOfFundsOther` is requirements-only, `transferNature`
    lives in quote paymentMetadata. No fields added; the nolint at
    `transfers.go:205` stays as the spec-verified marker (comment sharpened).
13. **Tier 20% — art-dupl verdicts (18.1–18.3).** Re-inspection with
    `--no-accept-directives --explain`: **ONE** suppressed clone group today
    (raw/public `TransferRequirement` mirror, type-1, 44 lines) — the plan's
    "6 groups" figure was stale (the dedup refactor already eliminated the
    rest). Verdict: ACCEPT (two-layer boundary; never merge via aliases).
    The in-source `// art-dupl:accept` directives ARE the suppression
    config; no separate baseline file (would duplicate the same information).
    AGENTS' stale "permanently flags TransferRequirement*/UserAddress/QuoteFee"
    claim corrected.
14. **Tier 20% — error-context polish (19.1–19.6).** `checkError` now
    surfaces unreadable response bodies (`"(response body could not be read:
    …)"`) instead of silently empty ones; `map*` errors carry raw wire values
    (`total amount %v %q`, `amount %v %q`, `reserved amount %v %q`);
    `SCAChallengeError.ErrorContext` exposes `approval_result` +
    `approval_token_issued` (bool — never the raw OTT); `AuthError`/
    `NotFoundError` keep their **promoted** APIError contexts — verified and
    pinned by tests rather than overridden with nothing new (an empty
    override is API surface without information). `TestErrorContexts`
    extended; new `TestIsRetryableContract` pins that ONLY RateLimit/Server
    implement `IsRetryable` (absence = non-retryable, by convention).
    AGENTS error-context convention entry written.
15. **Tier 20% — micro-batch A (21.1–21.5).** `classifyTransactionType`
    signature changed to cents (`totalCents int64`, test matrix updated with
    a documented sub-cent edge); `wiseDateFormat` constant extracted in
    `users.go`; `Profile.UserID` (branded) and `Profile.PublicID` (documented
    opaque string) now mapped from the wire.
16. **Tier 20% — micro-batch B (22.1, 22.5 + two recorded declines).**
    `WithUserAgent` option shipped end-to-end (config → Client →
    `setHeaders` → BDD tests: custom agent forwarded, default preserved) +
    README features bullets. Declined with rationale in TODO_LIST:
    `fmt.Stringer` on the string enums (no-op methods) and
    `errorfamily.RegisterClassification` (third-party sentinels only; wise-go's
    six types implement the Classified interface directly).
17. **Tier 20% — test infrastructure (23.1–23.4).** `bench_test.go` with four
    benchmarks (Cents 25.9ns/0 allocs; classify 11.4ns; mapTransaction 1.69µs
    /12 allocs; ParseWebhookEvent+accessor 34.5µs) with the benchstat
    workflow documented in-file; `FuzzParseWiseTimestamp` (never-panic +
    RFC3339 round-trip + zoneless-means-UTC invariants) and
    `FuzzNewCurrency` (accepts exactly 3 uppercase ASCII letters or errors)
    — seed runs + a 5s live fuzz (375k execs) pass;
    `internal/raw/types_test.go` with four wire-JSON round-trip tests
    (camelCase core keys, snake_case webhook keys, unknown-field tolerance,
    verbatim raw data). Flake fileset extended for both new test files.
18. **AGENTS maintenance.** Gotcha list pruned honestly (34 → 32 rows): the
    duplicated X-Rate-Limited-By rows merged, the no-action "int64 IDs
    already compliant" row removed, the stale art-dupl claim fixed. New
    entries: webhook 2026Q3/snake_case/UUID gotcha, error-context
    convention, the deliberate two-form validation idiom (requireID +
    inline; `requireNonEmpty` declined with written rationale).
19. **Living-doc hygiene during the run.** TODO_LIST/FEATURES/CHANGELOG/
    ROADMAP rows updated in the same tasks that changed code (never
    deferred), per the plan's verification gates.

## b) PARTIALLY DONE

1. **Tier 20% test infrastructure (23.5 missing).** Benchmarks, fuzz, and
   raw round-trips are in; the **client concurrent-safety test** (shared
   client, parallel requests, `-race` verification) is NOT yet written.
2. **v0.10.0 release (2.3–2.4).** Everything prepared (changelog verified,
   release notes drafted, gates green); the annotated tag + push + proxy
   verification + GitHub release are **gated on user approval** (tagging is
   irreversible through the module proxy).
3. **CI re-enable (4.1–4.3).** The workflow is fully prepared and locally
   proven (`nix flake check` green with the new no-auth inputs; YAML
   validated; lint pin bumped) — but enabling requires **push + the user's
   CI decision**. Nothing was pushed this session (harness rule).
4. **Method/example count claims have drifted.** The webhook work grew the
   surface to **37 client methods** (was 33) and **22 godoc examples** (was
   20), but `AGENTS.md` ("33 endpoint methods across 14 resources") and
   `FEATURES.md` ("20 `Example*` funcs") still carry the pre-webhook counts,
   and the audit doc's re-audit section describes the 33-method snapshot.
   The webhook resource is also a 15th resource. Counts need one refresh
   pass (see f).
5. **CHANGELOG `[Unreleased]` is webhook-only.** The micro-batch items
   (`WithUserAgent`, `Profile.UserID`/`PublicID`, classifier signature,
   fetchByID/refactor internals) are shipped but not yet listed under
   Unreleased.
6. **DOMAIN_LANGUAGE.md** has not gained the webhook vocabulary
   (`WebhookSubscription`, envelope, event type, 2026Q3 surface).
7. **Two declined plan items need no code but stay recorded**: enum
   `String()` methods and `RegisterClassification` (rationale in TODO_LIST);
   if the user disagrees, both are cheap to reverse.

## c) NOT STARTED

1. **4.1–4.3** — enable `ci.yml` on GitHub, watch the first run, confirm the
   coverage-badge job unfreezes the 86.9% badge (needs push + approval).
2. **2.3–2.4** — annotated `v0.10.0` tag + push + proxy `@v/list` +
   clean-dir `go get` consumer test + `gh release create` with the drafted
   notes (user-gated).
3. **23.5** — client concurrent-safety test.
4. **24.1–24.2** — split `errors_test.go` / `helpers_test.go` out of the
   `internal_test.go` monolith.
5. **24.3** — coverage-threshold step in `ci.yml` (fail under N%, keep badge).
6. **24.4** — gorelease/apidiff breaking-change check (CI job or flake app).
7. **24.5** — full gate run after the test-file surgery.
8. **25.1–25.2** — issue templates (bug + feature) and PR template.
9. **25.3–25.5** — ADR 001 (Money, no arithmetic), ADR 002 (flat package +
   `internal/raw` boundary), ADR 003 (go-retry migration rationale).
10. **25.6** — `doc-verify` flake app (lychee + godoc freshness) +
    `reports/jscpd-report.json` cleanup.
11. **Gated bundle** (untouched, per plan): sandbox key run, v1.0.0 tag,
    typed recipient `Details`, Cachix token, direnv GOEXPERIMENT pin.
12. **ROADMAP long-tail**: WithMetrics, `Page[T]`, property tests, quote
    residuals, service-client refactor, tier-3/4 APIs, 2026Q4 versioning,
    blog post, `nix flake check --all-systems`.
13. **App-level webhook scope** — deferred to ROADMAP with options +
    consequences written; `TestWebhookSubscription` ships only if that
    decision pulls client-credentials into scope.

## d) TOTALLY FUCKED UP (honest accounting)

1. **Four failed edit round-trips from daemon/fmt races.** Twice the
   auto-commit daemon changed files between my read and my edit
   ("modified since read"), and twice a python script asserted against
   patterns that `nix fmt` (gofumpt/golines) had just re-aligned — the
   scripts either died or, worse, wrote stale-shaped code. Root cause:
   I chained script edits after formatting without re-reading. Fix applied
   mentally and should be permanent: **after any `nix fmt`, re-read before
   ANY scripted edit**; prefer the edit tool with fresh views for
   post-format code. No damage survived — all failures were caught by
   build/lint before commit.
2. **The same syntax mistake twice.** Both python tail-splices into
   `wise_test.go` produced a stray `)` after a raw-string literal
   (`` `}) `` instead of `` ` ``) — I copy-pasted a func-closer pattern onto
   an assignment. Same class, repeated. Caught immediately by the compiler,
   fixed in one edit each; the lesson (unique-anchor splicing needs a
   `gofmt` sanity check before `go test`) is in (e).
3. **Three attempts to suppress tagliatelle before the right mechanism.**
   Field-level `//nolint` on the first field was "unused" (nolintlint),
   a comment block above `package` became a second package doc (godoclint),
   and a trailing nolint on the package clause did not suppress at all.
   The working answer was the repo's existing path-exclusion rule. The LSP
   lint cache mislead me twice during this dance; the CLI arbitrated.
   The full lesson is now recorded in the AGENTS webhook gotcha so no
   future session repeats it.
4. **`bench_test.go` briefly contained garbage import-keeper lines**
   (`var _ = json.Unmarshal` etc.) from drafting with unused imports —
   self-caught and removed before any commit landed, but it should never
   have been written.
5. **A placeholder survived briefly in CHANGELOG** ("- Nothing yet below."
   under the new Added entries) — self-caught on re-read, removed.
6. **Living-doc counts went stale mid-session** (see b4): I updated
   FEATURES/CHANGELOG/TODO for the webhook _features_ but did not refresh
   the _counts_ (33→37 methods, 20→22 examples, 14→15 resources) that the
   same changes invalidates. This is exactly the drift class the docs-health
   pass hunts; it must be closed in the next pass (f1–f3).
7. **Not fucked up but worth stating:** nothing was committed by hand and
   nothing was pushed — the harness forbids both without an explicit user
   ask; the daemon's ~25 auto-commits carry the session, and per AGENTS I
   spot-verified several of them contain what I wrote (they did).

## e) WHAT WE SHOULD IMPROVE

1. **Re-read after formatting, always.** Any `nix fmt` invalidates
   scripted-edit patterns; make "fmt → view → edit" the mechanical loop.
2. **Unique-anchor splicing needs a `gofmt -l` sanity check** before
   running tests; the stray-`)` class is invisible to grep.
3. **Trust the CLI, not the LSP, for lint semantics** — the LSP lint cache
   was wrong twice this session (stale findings for minutes). When they
   disagree, re-run `golangci-lint run` and move on.
4. **Suppress lint at the mechanism the repo already uses** (path
   exclusions in `.golangci.yml`) instead of inventing nolint placements;
   field-level/package-level nolint semantics differ per linter and cost
   three attempts to learn.
5. **Verify plan premises at task start, not mid-task.** Three plan
   premises failed verification this session (deposits# events, "6 art-dupl
   groups", "create accepts sourceOfFundsOther/transferNature") — all were
   caught, but each cost investigation. The plan could carry an
   explicit "verify-before-execute" line per risky task.
6. **Batch the expensive gate.** `nix flake check` is ~4–5 minutes; run it
   at task-group boundaries (done mostly) rather than per task.
7. **Pin what CI installs ad hoc**: `govulncheck@latest` and
   `gofumpt@latest` in `ci.yml` are unpinned (the gofumpt pin was in the
   micro-batch list and I missed it — see f13/f14).
8. **Daemon-authored history is opaque.** ~25 "chore: auto-commit N file(s)"
   commits this session make bisection and review harder than necessary;
   when the user wants, logical commits per task group (as the plan
   intended) would give real history.
9. **The audit doc should note its own snapshot semantics** — it describes
   the 33-method re-audit surface; after the webhook landing it needs either
   a delta note or the next re-audit (v1.0.0 will force one anyway).
10. **ParseWebhookEvent benchmark shows 34µs** dominated by the
    multi-layout timestamp trial loop; fine for webhook rates, but a
    fast-path for RFC3339 (the dominant wire format) is a one-line win if
    webhook volume ever matters.

## f) Next things (up to 50, sorted by impact)

1. **Refresh the living-doc counts** — 37 methods / 15 resources
   (AGENTS.md:58), 22 examples (FEATURES.md:115), + the audit doc delta
   note. (Drift introduced this session; ~20 min.)
2. **Add the micro-batch items to CHANGELOG `[Unreleased]`** (WithUserAgent,
   Profile.UserID/PublicID, classifier cents signature, fetchByID refactor).
3. **Update `docs/DOMAIN_LANGUAGE.md`** with the webhook vocabulary
   (subscription, envelope, event type, 2026Q3 surface).
4. **23.5 — client concurrent-safety test** (shared client, parallel
   requests under `-race`).
5. **24.1–24.2 — split `errors_test.go` / `helpers_test.go`** out of
   `internal_test.go`; then 24.5 full gates.
6. **24.3 — coverage-threshold CI step** (fail under N%; badge keeps
   working).
7. **24.4 — gorelease/apidiff breaking-change check** (CI job or flake app).
8. **25.1–25.2 — issue + PR templates.**
9. **25.3–25.5 — ADR 001/002/003** (Money, flat package, go-retry).
10. **25.6 — `doc-verify` flake app** (lychee + godoc freshness) + jscpd
    report cleanup.
11. **4.1–4.3 — push, enable CI, watch the first run green**, confirm the
    badge job unfreezes (needs user approval to push/enable).
12. **2.3–2.4 — tag `v0.10.0`, push, proxy `@v/list` + clean-dir
    `go get`, `gh release create` with the drafted notes** (user-gated).
13. **Pin `gofumpt` in `ci.yml`** (`go install mvdan.cc/gofumpt@latest` —
    missed micro-batch item).
14. **Pin `govulncheck` in `ci.yml`** (`go install ...@latest` today).
15. **Answer-folding: webhook app-level scope** — implement
    client-credentials CRUD + `test-notifications`, or keep the ROADMAP
    deferral (user decision).
16. **Cachix token (g2)** — set the secret, verify the cache name, flip
    `continue-on-error: false` in the nix job.
17. **Sandbox key (g1)** — first recorded sandbox run + CHANGELOG entry.
18. **v1.0.0 tag** — after 1–11 + explicit approval.
19. **erraudit pass over the new code** (`go-error-modernization` skill)
    — webhook code uses `errors.New` + typed errors; confirm zero findings.
20. **Re-run the full gate after the count-refresh docs pass** (flake
    links check included).
21. **Fuzz smoke job in CI** (short `-fuzztime 30s` for the two fuzz
    targets) — optional.
22. **art-dupl baseline + CI check step** — clone regression guard beyond
    the in-source directives.
23. **Fast-path RFC3339 in `parseWiseTimestamp`** if webhook latency ever
    matters (benchmark shows it dominates `ParseWebhookEvent`).
24. **Typed payloads for more event types on demand** (cards#*,
    swift-in#credit, batch-payment-initiations#state-change — all
    spec-verified constants already exist).
25. **Godoc example for `ListProfileWebhookSubscriptions`** (only create has
    one today).
26. **Verify `FEATURES.md` webhook rows' line numbers** after further
    `webhooks.go` edits (line-cited evidence goes stale fast).
27. **`docs/DOMAIN_LANGUAGE.md` enum audit** — 33 webhook event constants
    deserve a one-line mention under terms.
28. **README "Project Status" section refresh** if it carries method
    counts (same drift class as f1).
29. **Consider `nix flake check --all-systems` in CI** (aarch64-darwin
    coverage for the consumer fleet).
30. **Sandbox test growth**: quote → recipient → transfer credentialed
    scenario after the first live run.
31. **Property tests for Money cents conversion** (gopter) — ROADMAP.
32. **`WithMetrics` option** — ROADMAP.
33. **`Page[T]` pagination generalization** — ROADMAP (third paginated
    endpoint trigger).
34. **Service-client refactor** — post-v1.0 sequencing stands.
35. **Tier-3/4 APIs per demand** (addresses, balance-capacity, batch
    groups, OAuth, 2026Q4 versioning) — ROADMAP.
36. **Blog post / launch** — website-launch skill when v1.0 lands.
37. **Direnv GOEXPERIMENT pin** — user-machine ergonomics (gated).
38. **Typed recipient `Details` decision** — five-reports-old open design
    question (gated).
39. **Cross-check `wise-api-core-schemas.json` against the webhook
    schemas** — the small spec predates the webhook work.
40. **Re-verify the re-audit's "const groups stable at 22"** after the
    webhook enums (33 event constants + 2 creator + 2 scope values may have
    added const GROUPS — count drift to confirm).
41. **Sweep for remaining `@latest` installs in CI** beyond gofumpt/
    govulncheck.
42. **Add `TestWebhookSubscription`** only if f15 pulls app-level in scope.
43. **Consider a `VerifyAndParseWebhookEvent` helper** — decided against
    (composition); revisit only with real user demand.
44. **benchmark baselines**: store a first benchstat baseline for future
    comparisons (bench file documents the workflow).
45. **TODO_LIST: add the f1–f10 items as actionable rows** (docs-health
    HARVEST) so this report's section (f) does not get entombed.
46. **AGENTS: note the jsontext.Value choice** for raw webhook data (v2-native
    raw type; RawMessage avoided) — hard to rediscover.
47. **Check `sandbox_live.yml` still aligns** with the renamed workflow
    state after any CI changes.
48. **Consider naming the 2026Q3 constant upgrade path** — when Wise moves
    the surface (2026Q4), `webhookAPIVersion` is the single lever; document
    the expected bump ritual in AGENTS.
49. **Review daemon commits for the session** — verify each auto-commit
    contains what the session intended (spot-checks done; full sweep
    pending).
50. **Close out the todos tool** — mark 23.x partial state and 24/25 as
    pending so the next session inherits an accurate list.

## g) Questions I cannot figure out myself (top 3)

1. **Approve the `v0.10.0` tag now?** Changelog verified, release notes
   drafted, all gates green. Tagging is irreversible (module proxy caches
   it forever). Alternative: fold the webhook work into the tag and ship
   one bigger release — but then the tag needs another changelog pass.
2. **Approve push to `origin/master` + CI enable?** The workflow is
   prepared and locally proven; enabling unfreezes the coverage badge and
   restores govulncheck/lychee/cachix. This is the "CI re-enable strategy"
   question from the plan — the strategy turned out to be "no auth needed
   at all", so the only decision left is go/no-go on pushing.
3. **Webhook app-level scope**: implement the client-credentials token flow
   (pulls OAuth `POST /oauth/token` into scope, grows the auth surface
   before v1.0) — or keep the documented ROADMAP deferral and stay
   user-token-only?

---

_Report written per the status-report skill with one explicit override: the
user requested `.md`, so this is Markdown instead of the skill's HTML
dashboard default. No manual commit (harness rule) — the auto-commit daemon
will pick this file up. WAITING FOR INSTRUCTIONS._
