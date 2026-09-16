# Status: Docs-Health Full Audit (v0.11.0 Living-Doc Overhaul + Annotate/Archive Sweep) — 2026-09-16 15:30 CEST

**Session scope:** the user ordered "View ALL `**/2026-0*` files! Execute the docs-health SKILL!" with all six living docs superb and fully-done reports archived with inline strikethroughs. This report covers ONLY that run. **No Go source was modified.**
**Format note:** written as `.md` per explicit user instruction (the status-report skill's canonical format is HTML — one-off override, not propagated; same override as the 2026-09-13 reports).
**End state:** `go build` ✅ · `go vet` ✅ · `go test ./...` ✅ (README drift guard incl.) · `golangci-lint` **0 issues** · `nix run .#doc-verify` ✅ (62 links OK, count-claims match) · `nix flake check` **all checks passed** · daemon committed everything as `57766d6` (52 files, renames R099–R100 verified).
**Health report (printed inline in-conversation):** Accuracy **6.75 → 10** post-fix, Fitness **8.55 → 10** post-fix (baseline: the 2026-09-13 pass scored 9.4/9.5 — v0.11.0 shipped the day after and re-accumulated drift).

---

## a) FULLY DONE

1. **All `**/2026-0*` files viewed and classified** — 35 markdown files read (20 unarchived status reports, 3 planning docs, 4 reviews, 8 already-archived reports); 7 generated HTML scan artifacts + 4 D2/SVG diagrams re-classified LEAVE-ALONE per the "so what" test — and that verdict is now **codified as an AGENTS.md gotcha** so future runs stop re-litigating it (closes the 12-29 report's e.5).
2. **Deep code verification before every doc claim** — surface is verifiably **41 `*Client` methods** (`go doc -all`; 39 endpoint methods + `Authenticate`/`Health`) across **16 resources**; **24** `Example*` funcs; **7** OTT types (`OTTChannel`, `OTTChallengeType`, `OTTChallengeView`, `OTTChallenge`, `OTTStatus`, `OTTTriggerResult`, `OTTCodeProvider`); webhook line refs re-derived (`webhooks.go:33,64,90,117,149,178,283`); tags `v0.10.0=fe896a8`, `v0.11.0=5a6448c` on origin; `gh release list` still shows **v0.9.0 as Latest** (no Release objects for v0.10.0/v0.11.0); `ci.yml:92,113` still install gofumpt/govulncheck with `@latest`; no `.github/SECURITY.md`; no coverage floor in the flake `checks.test`; no `wise.Version` constant.
3. **All six living docs brought to v0.11.0 truth:**
   - **README** — status callout v0.10.0 → v0.11.0 with the full capability list (webhook subscriptions + typed events, SCA one-time-token clearing, `WithUserAgent`); the broken `paymentOptions` blockquote glitch fixed.
   - **FEATURES** — new **"SCA one-time tokens (v0.11.0)"** section (4 endpoints + `OTTChannel` + the OTT-secrecy contract, all line-cited); new rows for `WithUserAgent` (`options.go:81`), `Profile.UserID`/`PublicID` (`types.go:87-91`), shared-client concurrency (`wise_test.go:246`); **7 stale webhook line refs re-derived**; "15 resources" → 16 in the deferred row.
   - **ROADMAP** — Vision "33 endpoint methods" and Axis-1/Axis-4 "37 methods / 15 resources" → **41 across 16**; release state now records v0.9.0 + v0.10.0 + v0.11.0 as shipped (tags on origin, proxy-served, Release objects pending); "re-audit pending" note corrected; release-strategy line extended through v0.11.0; new **"Raw ideas (harvested 2026-09-16)"** section with 12 routed items (BadRequestError, circuit breaker, WithMetrics, GoReleaser automation, webhook quickstart, ParseWebhookEvent fast path, benchstat baseline, gopls skew, `--all-systems`, `Authenticate()` future, 2026Q4 bump ritual, wishlist leftovers).
   - **AGENTS** — fused bullet repaired (the `Retry-After` and `//nolint:bodyclose` gotchas had merged into one line), exact-duplicate Retry-After bullet deleted, "(v0.10.0)" mislabel on the 41-method claim → v0.11.0 with a pointer to the `doc-verify` count gate; **new gotchas**: historical-report annotate-then-archive policy (incl. the HTML/D2 LEAVE-ALONE rule) and the `jsontext.Value` (not `json.RawMessage`) webhook raw-data choice.
   - **TODO_LIST** — fully rebuilt: **12 done `[x]` items deleted** (completed work lives in CHANGELOG, never here); open items re-tiered P1/P2/P3 with per-item citations to code AND source reports; a "harvested-and-closed pointer" replaces the trophy case.
   - **DOMAIN_LANGUAGE** — OTT and challenge-channel glossary rows added (uppercase-challenge-type vs lowercase-wire-path mapping noted).
   - **CHANGELOG** — verified correct as-is (v0.11.0 entry complete; `[Unreleased]` legitimately empty); untouched.
4. **v1.0 API audit updated** (`docs/reviews/2026-08-21_v1.0-api-audit.md`) — new v0.11.0 delta blockquote: +4 `*Client` methods (OTT surface on `/2026Q3`) / +7 exported types → **41 methods across 16 resources**, verified 2026-09-16; risk-register extension: the OTT value must never appear in an error string or log (extends the item-3 freeze list). The audit is now current through the full shipped surface — nothing between it and the v1.0.0 tag but the user's approval.
5. **~250 inline annotations applied across 25 historical files** using the skill's `annotate-rows.py`/`annotate-prose.py` (dry-run first, atomic writes, read-back shape checks): the 2026-09-13 trio (12-29: 33 verdicts, 14-16: 20, 15-41: 13), the 08-21 sextet (21-46: 30, 17-14: 23, 18-15: 24, 03-32-era batches, 20-57, 23-10 dedup, nix-migration), the 08-08 quintet, both 08-19 reports, 08-28 webhooks (all 22 f-items struck as shipped in v0.11.0), the three May/July reports, and backfill on 6 of the 7 pre-existing archives that failed the gate. Stale resolution banners corrected where prior passes' "still open" verdicts had since shipped or been declined (webhook CRUD, typed events, CI SSH auth, SourceOfFundsOther, requireNonEmpty, art-dupl suppression).
6. **21 fully-resolved files ARCHIVED via `git mv`** (history preserved): `docs/status/archived/` ← 20 reports (all of 05-17 through 09-13; `docs/status/` now holds only `archived/`), `docs/planning/archived/` ← the 2026-09-13 Pareto plan (banner: EXECUTED, with the 17.5/17.6 refutation and 22.2–22.4 declines recorded). Pre-archive gate fixed: **all 28 archived files now carry at least one inline strikethrough** (`grep -rLn '~~' docs/{status,planning}/archived/` prints nothing). Living docs reference zero unarchived paths (grep-verified post-move).
7. **HARVEST executed, not just harvested-into-a-report** — every surviving forward item from the archived f-lists is now either done (struck with hash/decline evidence), user-gated in TODO_LIST P2, or routed: fresh P1 items (Release objects, CI tool pins), new P3 items (coverage-floor parity, erraudit gate, SECURITY.md, doc-verify hardening, CONTRIBUTING notes, quality micro-batch incl. the revived `wise.Version`), and 12 raw ideas in ROADMAP.
8. **Quality gates green after the full sweep** — build/vet/test (incl. the README drift guard), `golangci-lint` 0 issues, `nix run .#doc-verify` (godoc renders; AGENTS' "41 endpoint methods" and FEATURES' "24 `Example*` funcs" both match the real surface; 62 links OK), `nix flake check` all checks passed (format + sandboxed race/coverage test + links).
9. **Daemon-commit verified** — `57766d6` contains all 52 changed files with R099–R100 rename detection; no living-doc reference to a moved path survives; working tree clean.
10. **Inline health report printed** with visible math and the 2026-09-13 baseline cited (not invented).

## b) PARTIALLY DONE

1. **Inline-annotation depth on the largest f-tables.** For 50-row wishlists (05-15, 12-16, 02-53, 03-32, 17-14, 18-15, 15-41) I struck every row whose verdict I could determine from code (15–24 rows each); the remaining bare rows are genuinely open/routed ideas (sandbox scenarios, tier-3/4 APIs, property tests) whose "open" status is the correct signal — the corrected banners carry the routing. A maximalist pass could verdict literally every row.
2. **HTML/D2 artifact annotation** — classified and justified as LEAVE (now written into AGENTS), but not banner-annotated per file; same residual as the previous pass left it.
3. **TODO_LIST micro-batch compression** — six ≤30-min quality items (statement content-types, webhook spec-example fixture, webhook pagination-shape check, `ListProfileWebhookSubscriptions` godoc example, `wise.Version`, plus related) are bundled into ONE row to keep the list scannable; splitting into six rows would make progress tracking finer-grained.
4. **15-41 §b items** (b.1–b.6) were resolved via the corrected banner rather than per-item inline strikes (they are prose paragraphs, not numbered items).
5. **CHANGELOG conventions question** (should contributor-infra changes get a standing "Internal" subsection? — 15-41 §f.42) noticed, still undecided, not routed.

## c) NOT STARTED (open work observed this session, none of it executable docs-only)

1. **All user-gated items** (TODO_LIST P2): sandbox API key + first live run; `v1.0.0` tag; typed recipient `Details` decision (six reports old); `CACHIX_AUTH_TOKEN`; CI push + `gh workflow enable ci`; ADR 003 acceptance → go-retry migration; direnv/home-manager GOEXPERIMENT pin.
2. **P1 release work**: GitHub Release objects for v0.10.0 + v0.11.0 (gated); gofumpt/govulncheck/gorelease version pins (unblocked, not started this session).
3. **P3 quality/tooling batch** (just harvested): coverage-floor parity in `checks.test`, erraudit pass over the v0.11.0 webhook/OTT code + curated config + CI gate, SECURITY.md, doc-verify fail-on-empty hardening, CONTRIBUTING updates, the six-item micro-batch.
4. **ROADMAP long-tail** (WithMetrics, `Page[T]`, service-client refactor post-v1.0, tier-3/4 APIs, OAuth/app-level webhooks, blog) — demand-gated, untouched.
5. **No Go source was modified this session** — by design (docs-health run); the codebase state is exactly as v0.11.0 shipped it.

## d) TOTALLY FUCKED UP (honest accounting — nothing shipped is broken; these are my process failures)

1. **The first archive batch failed 20-for-20.** I built the destination as `docs/status/archived/$f` while `$f` already contained the `docs/status/` prefix — a nonexistent intermediate directory, so every `git mv` fatal'd and the `&&` chain silently skipped the planning move. Root cause: pattern-composed the loop without dry-checking a single move. The redo with `basename` worked, but a one-file smoke test would have saved the whole round trip.
2. **I raced MY OWN annotate scripts, twice, on the dedup report.** The scripts modified the file; my next multiedit used the stale read and got "modified since last read" — twice (the second attempt failed identically because I re-applied instead of re-viewing). Two wasted round trips on a lesson that is written down in THREE prior reports (re-read after any scripted/automated write). Bash `sed` does not reset the edit tool's read-tracking either — only `view` does.
3. **Three wrong-tool annotation attempts.** 05-23's Top-25 is a prose list (I fed annotate-rows), 15-41's f-list is a table (I fed annotate-prose), and 20-57/23-10 items 14/7/8/11 were already struck by the previous pass (I re-spec'd them and hit the already-annotated refusal). Each was visible in my own earlier views; I didn't check the list shape before writing specs. Atomicity meant zero damage — just four dead calls.
4. **Duplicate row-id surprise on the 05-21 archives** — `row 1: expected 1 match, found 3` (multiple tables in one file share row numbers); needed `--section "## F)"` scoping on the retry. The script's loud failure caught it; I should have scoped proactively for files with more than one table.
5. **Claim-then-verify ordering on the audit delta note** — I wrote "+7 exported types" into the v1.0 audit from inference BEFORE counting; the immediate `rg '^type OTT'` confirmed 7, so no lie shipped, but the order was wrong. Verify-then-write, always — this repo's own history has a fabricated-SHA incident from exactly this ordering.
6. **Daemon-commit verification was spot-checks, not a sweep.** `57766d6` carried 52 files; I verified the rename detection, the stale-path grep, and the final gates, but did not individually re-read all 52 files post-commit. The daemon's heuristic messages ("52 changed file(s)") remain opaque; the final state is verified green, but the review was sampled, not exhaustive.

## e) WHAT WE SHOULD IMPROVE

1. **Bulk destructive/structural commands get a one-unit smoke test first** — one `git mv`, one file, verify, then loop. The 20-fatal batch was entirely avoidable.
2. **"Script touched it → view before edit" must be mechanical, not remembered.** This session paid the tax twice on the same file; the rule is in AGENTS-adjacent lessons three times over. Consider: never issue an `edit` in the same minute-window as an annotate-script run on the same path.
3. **Shape-check before spec** — one `sed -n` on the target section (prose vs table vs duplicate row-ids) before writing annotate specs would have saved four failed calls.
4. **Verify-then-write for every number in NEW content**, not just for claims inherited from old docs — the `+7 types` near-miss is the same class as the fabricated cachix SHA (2026-08-21).
5. **Run the archive ~~ gate BEFORE the move, not after** — the 7 pre-existing failures were known from the session's first gate check; fixing them in place first would have made the move a single clean step.
6. **Standard daemon-commit audit step for doc-heavy runs**: `git show --stat` + rename-status + a full stale-reference grep (done) PLUS a bounded re-read of the N largest diffs (not done — add it).
7. **The 41 vs 39 counting basis should be stated once, canonically** — "41 `*Client` methods (39 endpoint methods + `Authenticate`/`Health`)" appears in the audit doc now; AGENTS/ROADMAP/FEATURES should keep quoting the `go doc -all` number and pointing at the audit for the derivation, so the 31/32/33/37/41 drift class never returns.

## f) Up to 50 things we should get done next (impact-ordered; ⏳ = user-gated)

> Top ~14 are TODO_LIST-grade (already routed this session); the rest are ROADMAP fuel. Sources: this session's verification + the archived 15-41/14-16/12-29 f-lists.

1. ⏳ **Publish GitHub Release objects for v0.10.0 + v0.11.0** — tags proxied, notes drafted for v0.10.0, cut v0.11.0 notes from CHANGELOG (TODO P1).
2. **Pin gofumpt (`ci.yml:92`), govulncheck (`ci.yml:113`), gorelease (apidiff app)** — kill `@latest` (TODO P1).
3. ⏳ **CI go-live** — push master + `gh workflow enable ci` + watch the first green run; unfreezes the coverage badge and exercises the 90% gate (TODO P2).
4. ⏳ **`v1.0.0` tag** — audit is green and now current through the 41-method v0.11.0 surface; only approval remains (TODO P2).
5. ⏳ **Sandbox key → first live run** — validates the FundTransfer empty-body semantics, still the highest-risk unverified behavior (TODO P2).
6. ⏳ **ADR 003 acceptance → go-retry migration** (executor swap, delete `classifyExhaustedRetries`, `Retry-After` via `DelayFunc`) (TODO P2).
7. ⏳ **Typed recipient `Details` decision** — six reports old; the v1.0 audit says the map can be frozen (TODO P2).
8. ⏳ **`CACHIX_AUTH_TOKEN`** secret + confirm the cache (TODO P2).
9. **Mirror the 90% coverage floor into `checks.test`** — local parity with the CI gate (TODO P3).
10. **erraudit pass over the v0.11.0 webhook/OTT code + curated config + CI gate** (TODO P3).
11. **`.github/SECURITY.md`** — vulnerability reporting path for a published SDK (TODO P3).
12. **`doc-verify` hardening** — fail loudly on empty extraction; cover ROADMAP/audit count claims (TODO P3).
13. **CONTRIBUTING** — apidiff/doc-verify apps, 90% gate, Ginkgo `-count` note (TODO P3).
14. **Quality micro-batch** — statement PDF/XLSX content-type assertions; webhook spec-example fixture; `ListProfileWebhookSubscriptions` pagination-shape verify; its godoc example; `wise.Version` (TODO P3).
15. Decide the CHANGELOG "Internal" subsection convention (15-41 §f.42, noticed this session, unrouted until now — fold into 13).
16. ⏳ **App-level webhook scope** (client-credentials token model + `TestWebhookSubscription`) — deferred decision (ROADMAP).
17. **Webhook end-to-end README quickstart** — subscribe → verify → parse in one block (ROADMAP raw idea).
18. **`ParseWebhookEvent` RFC3339 fast path** — 34µs bench dominated by timestamp trials (ROADMAP).
19. **Commit a benchstat baseline file** (ROADMAP).
20. **gopls/LSP config-skew fix** — stop the ~70-warnings/session CLI-vs-LSP divergence tax (ROADMAP).
21. **`nix flake check --all-systems`** — aarch64/darwin coverage (ROADMAP).
22. **Decide `Authenticate()`'s future** now that `GetMe` exists (ROADMAP).
23. **2026Q4 bump ritual note** — `webhookAPIVersion`/`quarterlyAPIVersion` upgrade path (ROADMAP).
24. **Typed `BadRequestError`** for 400s (ROADMAP design idea).
25. **Circuit breaker** — only on a consumer demand signal (ROADMAP).
26. **GoReleaser-style release automation** (ROADMAP).
27. **`WithMetrics` hook** — Prometheus/OTel (ROADMAP Axis 3).
28. **Fuzz smoke job in CI** (30s for the two fuzz targets) — optional (archived 14-16 §f.21).
29. **Fast-path sweep for remaining `@latest` in CI beyond the three known** (archived 14-16 §f.41; partially covered by item 2).
30. **Verify the v0.11.0 tag object** points at the intended commit (done for v0.10.0 in 15-41 §f.10; v0.11.0 spot-verified via `git rev-parse` this session — a one-line formal check remains).
31. **Re-check `sandbox-live.yml` alignment** after any CI change (archived 14-16 §f.47; verified consistent this session — recurring guard).
32. **`gitignore`/residue sweep** — confirm no dangling `reports/` references in tooling configs (archived 15-41 §f.43).
33. **`.github/ISSUE_TEMPLATE/config.yml`** — blank-issues toggle + discussion link (archived 15-41 §f.34).
34. **Review `dependabot.yml` coverage** (workflows + Go modules) (archived 15-41 §f.36).
35. **Cross-link ADRs from AGENTS.md** ("Money" bullet → ADR 001, retry bullets → ADR 003) (archived 15-41 §f.26).
36. **README docs table**: `docs/adr/` index + status-report pointers (archived 15-41 §f.25).
37. **Godoc example for `VerifyWebhookSignature` + `ParseWebhookEvent` composition** (archived 15-41 §f.30).
38. **`VerifyAndParseWebhookEvent` helper** — revisit ONLY on user demand (decided against; archived 14-16 §f.43).
39. **Cross-check `wise-api-core-schemas.json` against the webhook schemas** — the small spec predates the webhook work (archived 14-16 §f.39).
40. **Property tests for Money cents conversion** (ROADMAP).
41. **`Quote` residual fields** (`rateExpirationTime`, `targetAmountAllowed`, `user`) (ROADMAP).
42. **`Page[T]` generalization** — third paginated endpoint trigger (ROADMAP).
43. **Service-client refactor** — post-v1.0 sequencing stands (ROADMAP Axis 4).
44. **Tier-3/4 APIs per demand** (addresses, balance-capacity, batch groups, OAuth, simulations) (ROADMAP).
45. **Blog/design-story update** — the retry-typed-error + OTT-secrecy finds are the material (ROADMAP).
46. **Daemon-commit full-sweep review** — bounded re-read of the largest diffs in each auto-commit batch (this session's e.6).
47. **Ginkgo `-count` note verification** — confirm CONTRIBUTING lands it (tracks item 13).
48. **Raise the coverage gate 90% → 91%** once webhook payload tests mature (archived 15-41 §f.33).
49. **Mark the 3 runnable examples in README** (archived 15-41 §f.49).
50. **HARVEST this report's f-list** — items 1–14 are pre-routed; 15 and 30 are the only new TODO-grade items; the rest are ROADMAP (this session's successor pass).

## g) Three questions I cannot answer myself

1. **Release strategy for the two unpaginated tags:** publish GitHub Release objects for BOTH v0.10.0 and v0.11.0 (both tags are already on the module proxy; v0.10.0 notes are drafted, v0.11.0 notes cut from CHANGELOG), or skip the v0.10.0 page entirely and publish only a v0.11.0 release? Backfilling a Release for an already-proxied tag is safe but unusual — your call.
2. **CI go-live timing:** approve push + `gh workflow enable ci` now, so the first-ever run sees the full 41-method surface and unfreezes the coverage badge — or hold until after the `v1.0.0` tag so CI's first run is the frozen API? Everything on my side is prepared and locally green; the push/enable itself is gated on you.
3. **Typed recipient `Details` (six reports old):** do you want to (a) freeze `map[string]string` at v1.0 and add typed accessors additively later (the audit's recommendation), (b) commission typed per-corridor structs before the tag, or (c) the compromise layer — `DetailsKey*` constants for the top corridors? It is the last open design decision between the current tree and the v1.0.0 tag.

---

_Point-in-time snapshot. Status reports go stale; re-verify before acting on them. The auto-git daemon will pick up this file._
