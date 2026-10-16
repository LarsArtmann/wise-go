# Status: docs-health FULL AUDIT — harvest, verify, annotate, archive sweep, living-doc overhaul

> Session 2026-10-16 ~21:05–21:43 CEST. Input: "View ALL **/2026-0* files! Execute
> the docs-health SKILL! PROPERLY! TODO_LIST/CHANGELOG/AGENTS/README/ROADMAP/
> FEATURES must all be SUPERB! Archive FULLY done and UPDATED (inline
> strikethrough) .md files!" Scope: this session's run only. **Format note:**
> `.md` at explicit user demand (status-report skill default is HTML; override
> honored, flagged per skill contract).
>
> **Concurrent-session context (dominates everything):** a parallel session was
> executing the SUPERB Pareto plan in the same tree throughout — M5/M8/M9
> endpoint commits landed 21:16–21:19 (`19ab4e5`, `3c185f7`), M10 typed webhook
> payloads is IN FLIGHT (dirty `webhooks.go`/`types.go`/`internal/raw/webhooks.go`,
> `sandbox_live_test.go`), and the auto-commit daemon interleaved its raw-type
> edits into the same commits as my CHANGELOG/TODO edits (e.g. `3410d5e`). One
> daemon mid-edit commit shipped a duplicate-declaration build break
> (`internal/raw/webhooks.go`); observed, diagnosed, deliberately NOT touched —
> the owning session healed it minutes later (verified: one declaration each,
> `go build ./internal/raw/` exit 0).

## Verdict

All six living docs brought to current truth and gated green (`doc-verify` 0,
`release-notes-check` 0, `nix fmt` clean, 44-method count claims synced). The
harvest debt of seven reports cleared (10 new TODO rows, 3 new Decisions, 7 new
ROADMAP raw ideas). Three historical files fully resolved item-by-item and
archived with gated, manifest-recorded moves; six more kept reports annotated
in part. Remaining gap: the maximalist item-by-item pass on the eight kept
2026-10-07-era reports (deferred, tracked), and the new TODO backlog itself is
unexecuted (by design — routed, not done).

---

## a) FULLY DONE

1. **docs-health AUDIT executed properly** — skill + all five references loaded
   (harvest/verify/resolving/build/health-report), all four modes run; every
   2026-* markdown doc under status/planning/reviews viewed (archived/ dirs
   stayed closed per skill rule; the three 2026-07-18 HTML scan artifacts and
   the sca-auth-data-model HTML left alone per AGENTS policy).
2. **CHANGELOG `[Unreleased]` completed** — entries that were missing for
   already-shipped work: the **breaking** 2027Q1 account-details migration
   (`baba096`), the `spec-drift` gate + live 2027Q1 roll catch, the three new
   transfer endpoints (`19ab4e5`, `3c185f7`), `ExampleClient_GetTransfer`/
   `ListTransfers`, the transfer-category re-audit record, the README
   identifier-level snippet sweep, and the 8-anchor FEATURES fix.
3. **TODO_LIST rebuilt**: completed `[x]` migration item deleted (trophy-case
   removal — it now lives in CHANGELOG); SUPERB-remainder P1 row (M1-M3/M5-M9
   already shipped, hashes cited); v0.14.0 release-chain P1 row; 8 new P3 Open
   items (anchor-check gate, spec-drift hardening, python-gate negative tests,
   `.markdownlint.json`, coverage-floor single-sourcing, linter-allowlist
   check, fresh benchstat baseline, BuildFlow display-defects upstream filing);
   2 new P3 Decisions (spec-drift gate posture, flake pinning strategy) beside
   the extended Health code-half decision; CI item extended with the pin-sync
   job clause; stale "41-method" mentions fixed.
4. **FEATURES.md**: 12 drifted `file.go:NN` anchors re-verified against
   declarations (Health, Authenticate, all 7 OTT rows, WithSCAApprovalToken,
   ParseWebhookPublicKey, account_details ×2); Example count 25→27; missing
   `OneTimeToken`/`OTTAction` type-system rows added; Transfers summary 8/11
   → 11/11; Health row rewritten to readiness-probe semantics.
5. **ROADMAP.md**: vision paragraph current (v0.13.0 + staged v0.14.0, 44
   methods, transfer category complete, spec-drift gate); Axis-1 "Today"
   rewritten; release state + release strategy extended through v0.13.0 with
   the v0.14.0 staging note; 7 new raw-idea bullets (fleet pin-drift sweep,
   release-process automation, spec-drift growth, default UA from
   `wise.Version`, BuildFlow upstream proposals, crush-config lessons).
6. **Health F-1 docs-half EXECUTED** (was a harvested TODO): `client.go` godoc
   comment now states what `Health` IS (credential-validating readiness probe,
   typed-error meanings, ~2-min worst case), README snippet rewritten, FEATURES
   row synced. Code-half remains the user decision it is.
7. **AGENTS.md updated**: anchor-drift gotcha (doc-verify gates counts, not
   line anchors — with the same-evening re-drift as evidence); Ginkgo-Equal
   newtype lesson; `lll` recurrence clause on the buildflow gotcha; the nix
   "running 0 flake checks" quirk; gitleaks/codespell on-demand cadence added
   to the full-buildflow bullet; **erraudit entry rewritten for the truth**
   (BuildFlow #34 CLOSED 2026-10-07 — verified via `gh issue view`; the skip
   removal in `.buildflow.yml` was correct and full runs are green since) —
   closing a policy-vs-repo split brain.
8. **ANNOTATE, maximalist, on 2 reports**: `2026-10-07_07-35` (§a–§g, all 35
   §f rows) and `2026-10-07_08-22` (§a–§g, all 50 §f rows across 7 restarting
   subsections) — every item carries done-at-hash / routed / Won't-implement /
   NOT-DO with reasons; resolution banners added; `check-rows` COMPLETE;
   formatter-damage triage clean (`repair-mangle` 0, re-wraps stable);
   archived via `git mv` + manifest lines in both `archived/README.md`s.
9. **2026-07-18 comprehensive improvement plan fully resolved + ARCHIVED** —
   all ~100 tier/coarse/fine rows struck per the verified 2026-10-07 resolution
   banner (E8/F4/F5 marked declined non-goals), banner updated, grep-verified
   no living doc references it, `check-rows` complete, moved to
   `docs/planning/archived/`.
10. **transfer-category reaudit corrected inline** — the two "UNTRACKED GAP"
    rows proven FALSE (bankingpartner IS `GetTransferPayoutInfo`,
    `transfers.go:165`; deposit-details already DEMAND_GATED in FEATURES);
    M5/M8/M9 rows marked done with hashes; M15 rows routed; conformance note
    resolved (188-path snapshot). Kept in place (actively referenced).
11. **Light annotations on the kept set**: SUPERB plan (M1-M3/M5-M9 struck with
    hashes), 10-11 pair (§b resolved via 19:20 evidence; §f dones/routeds),
    19-48 (§f P1 1–3 done, 2/4–6 routed), 20-41 (§f 1/13/16/17/18/23 done),
    21-09 (§f 13–17 routed/done), 08-52 (§f 16–25 dispositioned). Open items
    left bare by design.
12. **HARVEST ledger complete** (see §f routing below) — park-landed pre-check
    applied against CHANGELOG/AGENTS throughout.
13. **Gates green, exit codes captured unpiped**: `nix run .#doc-verify` 0
    (44-method claims, 27 examples, 63 links), `nix run .#release-notes-check`
    0, `nix fmt` clean, `annotate.go self-test` passed, `check-rows` complete
    on all archived files, completeness grep (`grep -L '~~'`) clean on every
    moved file.

## b) PARTIALLY DONE

1. **Maximalist ANNOTATE on the kept 2026-10-07 reports** (04-57, 06-07, 06-34,
   07-09, 03-45, 04-20) + the two docs-health audit records (05-14, 07-51) —
   they remain at banner granularity from the 2026-10-07 passes. All their
   open items ARE routed in TODO_LIST (verified), so no signal is lost; the
   per-row strike pass is deferred. Tracked by 07-51 §f1.
2. **Light-annotation coverage** on kept reports covers only items with
   this-pass evidence; several still-open items were deliberately left bare
   (correct per canon, but it means those files mix struck and bare rows —
   `check-rows` reports them PARTIAL, which is the honest state).
3. **DOMAIN_LANGUAGE.md** existence confirmed but terms NOT re-verified against
   code this pass. Same for the 12-factor/api-docs-study/open-banking bodies —
   skimmed for currency/claims only.
4. **`nix flake check` and `go test ./...` NOT run** — deliberate: the
   parallel session owns the in-flight tree (M10 + `sandbox_live_test.go`
   dirty). Running repo-wide gates against a moving tree is the documented
   anti-pattern; doc-scoped gates only. A fresh battery belongs to whoever
   finishes M10.
5. **CONTRIBUTING peeled-commit nuance** — decided mid-flight to skip the
   one-liner (AGENTS carries the trap in full); the 08-22 §f.8 verdict marks
   it done-elsewhere.
6. **The TODO anchors I wrote** (e.g. `client.go:190`) will drift like all
   anchors — the anchor-check item exists precisely for this; nothing guards
   them yet.

## c) NOT STARTED (routed this pass, zero execution)

1. The 8 new P3 tooling items themselves: anchor-check gate, spec-drift
   hardening (floor guard, PYTHONDONTWRITEBYTECODE, 2027Q1 path triage),
   python-gate negative tests, `.markdownlint.json`, coverage-floor
   single-sourcing, linter-allowlist check, fresh benchstat baseline,
   BuildFlow display-defects upstream filing.
2. Health code-half decision; spec-drift posture decision; flake pinning
   strategy decision (all BLOCKED: user).
3. v0.14.0 release chain (P1 row written).
4. Cross-repo: crush-config lessons (pipe-masking, ps-before-gates) — routed
   to ROADMAP ideas only; writing them belongs to a crush-config session.
5. Annotating the 19-59 architecture-understanding review as F-1/F-3 resolve
   (left for its own cycle).
6. AGENTS.md size reduction (48.6 KB, flagged — not attempted).

## d) TOTALLY FUCKED UP

1. **The pipe-masking sin, committed in the very session that re-canonized
   it:** `go build ./... 2>&1 | head; echo "build-exit=$?"` printed
   `build-exit=0` while the build had FAILED (the parallel session's
   duplicate-decl break). Caught only because I read the compiler errors in
   the output text. Every subsequent gate ran unpiped — but the reflex fired
   once, exactly as the AGENTS gotcha predicts.
2. **A self-canceling multiedit on AGENTS.md:** my insert-after trick ran two
   edits that replaced each other's output (Money bullet → anchor bullet →
   Money bullet), silently deleting the new gotcha. Caught by an immediate
   `grep` verify and re-inserted properly. Lesson: never compose insertions
   as swap pairs.
3. **Two edit failures from read-state/wrap drift** (TODO_LIST "Optionally add
   a" wrap, ROADMAP vision paragraph) plus one FEATURES mod-time refusal
   (parallel session edited it at 21:18 between my read and my edit). All
   recovered by re-reading; each cost a round trip.
4. **Hand-typed annotate specs hit shape errors three times** (§c/§f are prose
   with restarting numbers, not tables; the tool refused atomically — dry-run
   discipline + atomicity meant zero damage, but the specs should have been
   probed with one dry-run per NEW shape first, per the skill's own rule).
5. **My first TODO_LIST P1 row was stale within minutes** — I wrote "remaining:
   M5/M8/M9..." while the parallel session was shipping exactly those; had to
   rewrite the row twice. In a multi-session repo, `git log` must be re-checked
   between composing and landing living-doc edits.
6. **Initially trusted stale LSP diagnostics** (duplicate-decl errors persisted
   in diagnostics after the working tree was fixed); ground-truthed with
   grep + targeted build instead of restarting the LSP once.

## e) WHAT WE SHOULD IMPROVE

1. **Serialize the docs lane across sessions** — tonight three sessions wrote
   TODO/CHANGELOG/FEATURES/AGENTS concurrently and the daemon interleaved them
   into shared commits. A "living-docs lock" convention (check `git log` for
   foreign doc commits immediately before AND after each living-doc edit
   batch) would remove the stale-in-minutes class.
2. **Build the anchor-check gate before writing more anchors** — I added a
   dozen hand-verified `file.go:NN` anchors tonight; they rot by the next
   code-touching session. The TODO item is the real fix; until then, anchors
   are debt.
3. **Pipefail posture**: the correct template is `cmd > log; echo $?` — I know
   it, taught it in AGENTS this session, and still typed a pipe first. The
   habit needs to be mechanical, not recalled.
4. **Dry-run-first for every NEW annotate shape** — one `--dry-run` invocation
   per unfamiliar section shape (table vs restarting prose) before any write
   batch; my three shape errors would have been zero.
5. **docs-health maximalist passes should be scheduled with the harvest** —
   the kept-report backlog survives because each pass defers it; a standing
   "annotate N reports per docs pass" quota would finish the tail in ~2 passes.
6. **AGENTS.md needs a size plan** — 48.6 KB with battle-tested content; the
   right move is extracting incident NARRATIVES to docs/reviews and keeping
   the constraint statements, but that deserves its own careful session, not a
   drive-by prune.

## f) Up to 50 things we should get done next (honest count: 21)

1. Maximalist item-by-item annotate pass on `2026-10-07_04-57` → archive candidate.
2. Same for `2026-10-07_06-07`.
3. Same for `2026-10-07_06-34` (partially marked already).
4. Same for `2026-10-07_07-09`.
5. Same for `2026-10-07_03-45` + `2026-10-07_04-20` (buildflow repair pair).
6. Same for the two docs-health audit records (`05-14`, `07-51`) — their §f
   maximalist item closes when the tail above is done.
7. Build `nix run .#anchor-check` + repo-wide anchor sweep (TODO P3).
8. spec-drift hardening: extraction-floor guard, `PYTHONDONTWRITEBYTECODE`,
   2027Q1 16-path FEATURES triage (TODO P3).
9. Persist negative-scenario tests for `pin_sync.py` + `spec_drift.py` (TODO P3).
10. Add `.markdownlint.json` (MD013 off) (TODO P3).
11. Single-source the coverage floor (TODO P3).
12. Build the linter-allowlist check (TODO P3).
13. Cut a fresh benchstat baseline at the v0.14.0 tag (TODO P3).
14. File the two BuildFlow display defects upstream after verify-before-filing (TODO P3).
15. Verify the parallel session's M10 lands completely (typed payloads + webhooks
    count claims + CHANGELOG entry) and gate it.
16. M12: README SCA/OTP endpoint matrix.
17. M14: env-gated live-smoke harness; M15: originator transfers.
18. Cut v0.14.0 (pre-release chain, tag, apidiff, proxy verify) — after M10+.
19. Annotate the 19-59 architecture-understanding review as F-1/F-3 resolve.
20. Re-verify DOMAIN_LANGUAGE terms; re-audit the api-docs-study 215/51 counts
    post-2027Q1; add a point-in-time note to the open-banking comparison.
21. Cross-repo: pipe-masking + ps-before-gates lessons into crush-config
    `references/lessons.md`; AGENTS.md size-reduction session.

## g) Questions I can NOT figure out myself (3)

1. **v0.14.0 timing and push:** master carries a large unpushed commit train
   (today's parallel sessions + mine; `origin/master` is behind). Cut and push
   v0.14.0 now with what's staged (breaking OTT typing + 2027Q1 migration +
   transfer-category completion), or hold until M10/M12-M14 land? And do you
   approve pushing master at all?
2. **Parallel-session arbitration:** another session is mid-flight on M10
   (typed webhook payloads, dirty `webhooks.go`/`types.go`/`sandbox_live_test.go`).
   Should every other session (including future me) treat the Go-code lane as
   exclusively theirs until their report lands, with only docs/status work
   allowed in parallel?
3. **Annotation depth policy:** do you want the remaining eight kept reports
   item-by-item resolved over the next ~2 docs passes (the maximalist
   standard), or is banner-granularity acceptable now that every open item is
   verifiably routed into TODO_LIST/ROADMAP? This decides whether §f items
   1–6 above stay queued.

---

*Point-in-time snapshot 2026-10-16 21:43 CEST. Verification appendix: doc-verify
0 · release-notes-check 0 · nix fmt clean · annotate self-test + check-rows
complete on all archived files · completeness grep clean · BuildFlow #34 state
verified via gh (CLOSED 2026-10-07T06:39:22Z) · spec snapshot 188 paths ·
`go build ./internal/raw/` 0 · method count 44 (doc-verify-gated). Not run this
session (deliberate): go test ./..., nix flake check, apidiff — in-flight tree
owned by the parallel session. WAITING FOR INSTRUCTIONS.*
