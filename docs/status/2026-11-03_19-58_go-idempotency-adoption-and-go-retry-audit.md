# Status Report: go-idempotency adoption audit + go-retry review session

**Generated:** 2026-11-03 19:58 CET · **Repo:** wise-go (working tree clean @ 903739b) · **Session type:** read-only audits + one cross-repo filing (go-idempotency #13)

## Scope line — what this report VERIFIED vs did NOT verify

**Verified this session (command/artifact-backed):**

- go-idempotency non-adoption in wise-go: `grep` over `go.mod`, all `*.go`, `flake.nix`, docs — zero references.
- Wise's three idempotency mechanisms: read with line numbers — `balances.go:91-143`, `transfers.go:213-235`, `third_party_transfers.go:60-72`, `ids.go:115-118`, `webhooks.go:28`.
- go-idempotency library facts: read its `go.mod` (`go 1.27`, deps go-error-family v0.12.0 + rapid), `store.go` API, `middleware/http.go`, `middleware/middleware.go`, `doc.go` replay recipe, `ROADMAP.md`/`FEATURES.md`, tags (latest v0.3.1).
- Issue filed and landed: `gh issue view 13` → OPEN, title + body intact (3555 bytes).
- Voice gate: `check-draft.go --kind body-issue --ai-drafted` → 0 FAIL / 0 WARN, exit 0.
- go-retry usage: `client.go:61,110-131,151-165,221,303,339,366-452`, `flake.nix:32-33,123`, `go.mod` (v0.7.1).
- Repo states at report time: wise-go clean; go-idempotency has untracked `docs/drafts/` (my draft) + two untracked artifacts from a PARALLEL session (19:43/19:52 brutal-self-review outputs — not mine, not audited).

**NOT verified (and why):**

- No tests run — no code was changed; nothing to test. Audit-only session.
- README/`docs/` coverage of the retry behavior and the balance-key trap — not read this session.
- go-retry upstream versions beyond the pinned v0.7.1 — not checked.
- `nix run .#pin-sync` — not run (no dependency changed).
- Whether the go-idempotency auto-commit daemon will sweep the untracked `docs/drafts/` file — unknown at write time.
- `file:line` anchors cited in issue #13 are on GitHub and rot silently — no mechanical gate watches GitHub-hosted bodies.

## Session work inventory (chronological)

1. **Q: "Do we use/leverage go-idempotency?"** → Audited: **NO**. Zero refs in `go.mod`/tree/flake/docs. Wise's idempotency is wire-native.
2. **Q: "How does wise do idempotency?"** → Documented the three per-endpoint mechanisms (table below), the triple-scope dedup semantics, the balance fresh-UUID trap, webhook consumer-side dedup duty.
3. **Task: file an issue to go-idempotency explaining non-adoption** → loaded github-voice skill, audited the library, drafted in-repo (`docs/drafts/2026-11-03_wise-go-adoption-rejection.md`), voice-checked, filed **[go-idempotency#13](https://github.com/LarsArtmann/go-idempotency/issues/13)**, verified landed.
4. **Q: "How about go-retry?"** → Confirmed genuine use (v0.7.1, ADR 003): `retry.DoWithValue` wraps every request at `client.go:371`.
5. **Q: "Everywhere? Superbly!?"** → Coverage audit: single `httpClient.Do` choke point (`client.go:423`); quality audit found real strengths + one asterisk (keyless-POST hazard, below).

### The Wise idempotency contract (session's reference table)

| Endpoint             | Key                                      | Wire location | Replay behavior                                                                              |
| -------------------- | ---------------------------------------- | ------------- | -------------------------------------------------------------------------------------------- |
| Create transfer      | `customerTransactionId` (mandatory UUID) | body          | same key + `quoteUuid` + `targetAccount` → existing transfer (`transfers.go:215`)            |
| Create balance       | `X-Idempotence-Uuid` (mandatory UUID)    | header        | identical key → the already-created balance (`balances.go:134`, sandbox-verified 2026-10-17) |
| Third-party transfer | `OriginalTransferID`                     | body          | Wise dedupes on it (`third_party_transfers.go:62`)                                           |

### The one asterisk found (keyless-POST hazard)

`isRetryableError` (`client.go:151-165`) retries **untyped (transport) errors by design** — pre-go-retry behavior, deliberately preserved, documented in the doc comment. Consequence: a lost connection on a **keyless POST** (recipients, quotes, webhook subscriptions, OTT trigger) can double-create the resource if Wise processed attempt 1. Keyed POSTs (transfers/balances/third-party) are safe server-side; GETs are safe. No code change was made; the finding was delivered in chat only.

## a) FULLY DONE

| #  | Work                                                                               | Evidence                                                              |
| -- | ---------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| 1  | go-idempotency non-adoption verified with evidence                                 | `go.mod`, tree grep, `flake.nix` — zero hits                          |
| 2  | Wise's three idempotency mechanisms documented with `file:line`                    | table above, `ids.go:115-118`                                         |
| 3  | Triple-scope dedup semantics articulated (key + `quoteUuid` + `targetAccount`)     | `transfers.go:215-218`                                                |
| 4  | go-idempotency issue #13 filed, voice-gated, verified landed                       | `gh issue view 13` → OPEN, 3555 bytes                                 |
| 5  | Draft persisted in the target repo per skill rule                                  | `go-idempotency/docs/drafts/2026-11-03_wise-go-adoption-rejection.md` |
| 6  | go-retry usage confirmed + mechanics mapped                                        | `client.go:61,114-119,371`, `flake.nix:32`, ADR 003                   |
| 7  | Retry coverage proven: one `.Do` call site, all helpers funnel through `doRequest` | `client.go:303,339,423`                                               |
| 8  | Retry hygiene verified: body closed before retry on non-2xx                        | `client.go:445-448`                                                   |
| 9  | `Retry-After` cap + delay-decision logging verified in `DelayFunc`                 | `client.go:119-131`                                                   |
| 10 | Fail-fast mode confirmed to reuse the executor (`MaxAttempts = 1`, no bypass)      | `client.go:221`                                                       |

## b) PARTIALLY DONE

| # | Work                                                                 | State                                                                                                                                                    |
| - | -------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | Issue #13's open ask (README/doc.go "client-side posture" paragraph) | Filed; text not written; repo undecided (go-idempotency docs vs wise-go docs)                                                                            |
| 2 | Keyless-POST hazard analysis                                         | Diagnosed + explained in chat; **not persisted in any repo artifact** — lives only in this session                                                       |
| 3 | Balance `IdempotencyKey` fresh-UUID trap                             | Surfaced; trap is documented in code (`balances.go:92-95`) but consumer-facing docs presence unverified                                                  |
| 4 | go-idempotency draft file                                            | Written but **untracked/uncommitted** in that repo; issue body on GitHub is the durable copy, draft file is now redundant                                |
| 5 | Keyless-POST endpoint inventory                                      | Four families named from memory of the codebase (recipients, quotes, webhook subscriptions, OTT trigger); full list vs FEATURES.md matrix not reconciled |

## c) NOT STARTED

| # | Work                                                                                                  | Why it's listed                                                           |
| - | ----------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| 1 | AGENTS.md memory updates (both repos): non-adoption decision + #13 pointer; go-retry choke-point fact | Aggressive update protocol says write at discovery; not done this session |
| 2 | README caveat for transport-error retry on keyless POSTs (wise-go)                                    | Discovered this session                                                   |
| 3 | `isRetryableError` tightening decision (gate untyped retries to idempotent methods)                   | Product/risk call, flagged verbally only                                  |
| 4 | HARVEST of section (f) into `TODO_LIST.md` / go-idempotency's `TODO_LIST.md`                          | Skill loop: report → harvest; not yet run                                 |
| 5 | Cross-repo traceability link: wise-go TODO item ↔ go-idempotency #13                                  | Both repos affected by the same decision                                  |
| 6 | Confirm whether Wise ever documents server-side dedup TTL; record if published                        | Open question surfaced by the audit                                       |
| 7 | Any test run / lint run this session                                                                  | Out of scope (no code changed) — listed for completeness                  |

## d) TOTALLY FUCKED UP

**Nothing shipped is broken.** No code was modified; both repos' working trees are untouched by me (wise-go clean; go-idempotency has only my untracked draft + a parallel session's artifacts).

One process-level near-miss, worst of the session:

- **A severity-bearing finding (silent duplicate resource creation on keyless POSTs under transport retry) was delivered only as chat prose.** If this session ends without harvest, the finding evaporates while the pinned behavior ships on. That is exactly the "ghost finding" failure mode the memory rules exist to prevent. Mitigation is section (f) items 1-3.

## e) WHAT WE SHOULD IMPROVE

1. **Persist severity-bearing findings at discovery, not at report time** — the keyless-POST hazard should have entered `TODO_LIST.md` the moment it was diagnosed.
2. **GitHub-hosted evidence rots silently** — issue #13 cites `file:line` anchors with no mechanical gate; promote decisions out of issues into living docs (AGENTS/ADR) where anchor-check watches them.
3. **Cross-repo sessions need cross-repo bookkeeping** — one decision (non-adoption) now lives in go-idempotency #13; wise-go's AGENTS/TODO don't know it exists. Link both directions.
4. **Verify consumer-visible documentation before citing "documented intent"** — I leaned on code comments; whether README carries the same warnings was never checked.
5. **Multi-session repo hygiene** — go-idempotency currently holds untracked outputs from two different sessions (mine + the 19:43 brutal-self-review one); the daemon interleave makes untracked files easy to lose or misattribute.
6. **Voice-gate worked; reuse it** — check-draft 0 FAIL/0 WARN first pass; keep drafting→checking→filing→verifying as the standard filing loop.

## f) Up to 50 things we should get done next

Brainstorm inventory (skill rule: most items below the top few are ROADMAP fuel, not commitments). Ordered by impact; ROUTE tags where clear: `[T]` TODO_LIST, `[R]` ROADMAP, `[G#]` go-idempotency repo, `[Q]` needs a question answered first.

1. `[T]` HARVEST this report into wise-go `TODO_LIST.md` (docs-health HARVEST mode) — the loop-closing step.
2. `[T]` Persist the keyless-POST transport-retry hazard with evidence (`client.go:151-165`) as a TODO row.
3. `[Q]` Decide retry posture for keyless POSTs: keep current behavior vs gate untyped retries to idempotent methods in `isRetryableError`.
4. `[T]` If tightening: re-pin with BDD tests per method class (GET/keyed-POST/keyless-POST matrix).
5. `[T]` Document retry behavior in wise-go README: what retries (429/5xx/transport), what doesn't (4xx/SCA), the keyless-POST caveat.
6. `[G#13]` Write the "client-side posture" paragraph (go-idempotency README or doc.go).
7. `[T]` Link wise-go TODO ↔ go-idempotency #13 both directions.
8. `[G#]` Commit or trash `go-idempotency/docs/drafts/2026-11-03_wise-go-adoption-rejection.md` (issue body is canonical).
9. `[T]` wise-go AGENTS.md: add non-adoption decision + #13 pointer + keyless-POST hazard entry.
10. `[G#]` go-idempotency AGENTS/FEATURES: record wise-go as a studied non-adopter (case-study evidence for the library's scope claims).
11. `[T]` Verify README documents the balance `IdempotencyKey` fresh-UUID trap for consumers; add if missing.
12. `[Q]` Decide whether `CreateBalance`'s implicit-key default should change (required key vs per-call UUID vs explicit claim API) — breaking-change decision.
13. `[T]` Reconcile the keyless-POST endpoint list against the FEATURES.md endpoint matrix (46 shipped operations).
14. `[T]` Add a test pin enumerating keyless POSTs so the hazard surface stays visible as endpoints grow.
15. `[R]` Record Wise dedup-TTL as an unknown; sweep for documentation on the ~8-week external-claims cadence.
16. `[T]` Document that webhook consumers own dedup (wise-go ships no handler-side dedup) — confirm deliberate, then document.
17. `[G#]` go-idempotency README: note the `go 1.27` floor is an adoption blocker for `go 1.26` consumers (wise-go case).
18. `[G#]` Evaluate whether go-idempotency could build on `go 1.26` (is the 1.27 directive essential?).
19. `[T]` Promote the session's three-mechanism idempotency table into wise-go living docs (`docs/DOMAIN_LANGUAGE.md` or a docs/ page) — currently only in chat + a GitHub issue.
20. `[T]` When promoting any #13 anchor into living docs, run `nix run .#anchor-check`.
21. `[R]` Track #13's `file:line` citations for rot; re-verify if the issue is ever acted on.
22. `[T]` Verify the `Retry-After` cap specifically is pinned by a test (cap at `MaxDelay`), not just honored-wait tests.
23. `[T]` Verify OTEL meter bridge skips delay-decision log entries (`RetryAfterDelay > 0`) — AGENTS asserts it; not re-verified this session.
24. `[T]` Verify fail-fast mode also disables `Retry-After` honoring (implied by `MaxAttempts = 1`; not explicitly checked).
25. `[R]` Consider surfacing attempt counts / retry outcomes on typed errors for observability consumers.
26. `[T]` Check go-retry upstream for releases > v0.7.1 before the next dependency sweep.
27. `[T]` Run `nix run .#pin-sync` at the next dependency-touching commit (not run this session; no dep changed).
28. `[R]` Investigate whether Wise's keyless POST endpoints accept any undeclared idempotency affordance (spec deep-dive, 2027Q1 spec + live reference).
29. `[T]` Add the keyless-POST caveat to `docs/error-codes.md` if transport retries surface as untyped errors to consumers.
30. `[R]` Explore a request-scoped idempotency-key helper for consumers (generate-once, reuse-across-retries) without making the SDK stateful.
31. `[T]` Confirm `example_test.go` examples don't teach the fresh-UUID-per-restart anti-pattern.
32. `[G#]` go-idempotency: consider a short "when NOT to use this library" section (server-replaying APIs) — the general form of #13's ask.
33. `[T]` wise-go CHANGELOG: no entry needed this session (no code change) — verify the daemon agrees next commit.
34. `[R]` API-diff check at next release: `nix run .#apidiff` (untouched this session; routine pre-release item).
35. `[T]` Re-run `gh issue view 13` after any go-idempotency push to confirm the body survived edits.
36. `[R]` Study how Stripe/Adyen document client retry guidance for keyless endpoints; adopt the best pattern for wise-go docs.
37. `[T]` Ensure the parallel session's go-idempotency artifacts (19:43/19:52) get claimed or trashed by their owner — attribution hygiene.
38. `[G#]` go-idempotency ROADMAP: fold #13's posture paragraph into its next docs pass.
39. `[T]` Add "does not use go-idempotency" one-liner to wise-go README dependencies section (consumer-facing clarity).
40. `[R]` If go-idempotency ever ships a response-replay Store, re-evaluate adoption for a wise-go optional layer (opt-in, not default).
41. `[T]` Sanity-check that `verify-external-claims`-style verification isn't needed for the #13 claims: all were source-read this session — done, none needed. (Negative item, closes the loop.)
42. `[R]` Fuzz/chaos: simulate lost-response-then-retry on a keyless POST against sandbox to demonstrate the duplicate empirically (needs sandbox key + a disposable resource type).
43. `[Q]`–dependent items: 3, 12 resolve first; 4, 5, 13, 14 follow from them.
44. `[T]` Next `go-cqrs-lite`-adjacent sweep: confirm the metaengine `DedupStore` capability and go-idempotency's division of labor is still documented as the README claims (session-adjacent, one grep).
45. `[R]` Reconsider `WithMetrics` hold: retry/attempt counts from this audit's choke point would be the natural first metric — only when a consumer asks.

(45 items; items 1-10 are the actionable core, 11-45 are brainstorm/ROADMAP fuel per the skill's harvest-rigor rule.)

## g) Questions I can NOT figure out myself (max 3)

1. **Retry posture for keyless POSTs** (recipients, quotes, webhook subscriptions, OTT trigger): keep the current always-retry-transport-errors behavior (risk: silent duplicate creation when a response is lost), or gate untyped retries to idempotent methods in `isRetryableError` (behavior change, test rework, safer default)? This is a consumer-risk decision I cannot make for you.
2. **`CreateBalance` key ergonomics**: should the fresh-UUID-per-call default stay (current, documented in code only), or should the API force an explicit stable key (required field / typed claim)? Any change is breaking; the trap is real but rare (process restart mid-retry).
3. **Home for issue #13's open ask**: should the "client-side posture" paragraph (server-replaying APIs degrade under local key-only guards) live in go-idempotency's README/doc.go, in wise-go's docs, or both?

## Then: WAIT FOR INSTRUCTIONS

Report ends here. No harvest, no commits, no further work until instructed. The auto-commit daemon may pick this file up; per the harness contract I am not committing it myself.
