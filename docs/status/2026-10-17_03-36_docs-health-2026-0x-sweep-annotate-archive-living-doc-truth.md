# Status: Docs-Health Sweep of the 2026-0x Era — Annotate/Archive + Living-Doc Truth — 2026-10-17 03:36 CEST

> **Format note:** the status-report skill's canonical output is a styled HTML
> dashboard; this report is `.md` on the user's explicit path instruction
> (standing dispatch-report exception, same as the 2026-10-16/17 reports).
> Scope discipline: everything below is from THIS session's run and what it
> surfaced — no new research.

**Session scope:** the user ordered "View ALL `**/2026-0*` files! Execute the
docs-health SKILL!" with all six living docs superb and fully-done reports
archived with inline strikethroughs; then this status update. **No Go logic
changed** — one godoc comment added (`New`'s empty-key behavior, a
twelve-factor-review leftover made of honest verification).

**End state:** build ✅ · vet ✅ · `golangci-lint` **0 issues** · markdownlint
**exit 0** (command now execution-tested and documented in CONTRIBUTING) ·
`nix run .#doc-verify` ✅ (46-method claims consistent in all living docs, 0
link errors, 36 Example funcs) · `nix run .#spec-drift` **exit 0 vs live
2027Q1** (188 paths; the AGENTS claim is now evidence-backed, not memory) ·
`nix flake check` **all checks passed** · daemon swept the work in five
batches (`47eb0f2`…`dc605b7`), each spot-verified by content grep; one file
(FEATURES.md anchor re-pin) still uncommitted at report time — daemon will
pick it up.

**Inline health report (AUDIT, printed in-conversation during the run):**
Accuracy **7.25/10 as-found** (10 − 1·1 Critical − 0.5·3 Medium − 0.25·1
Low), Fitness **7.35/10 as-found** (10 − 0.75·3 structural-decay −
2×(0.45−0.25) TODO_LIST done-work ratio). Every finding was fixed in-audit;
first audit against the v0.14.0 release state (no earlier baseline claimed).

---

## a) FULLY DONE

1. **All live `**/2026-0*` files viewed and classified.** 6 markdown reports
   read in full (the 2026-09-16 status report, the 2026-08-19 implementation
   plan, and the 2026-07-23/2026-08-08/2026-08-21 reviews). 7 generated
   HTML/D2 scan artifacts re-classified LEAVE-ALONE per the AGENTS.md gotcha.
   The three `archived/` trees (48 status + 7 planning + 1 review files) were
   NOT opened — closed history per the docs-health model.
2. **Discovery that reframed the living docs: v0.14.0 is already shipped** —
   tag `9d815ca`, GitHub Release "Latest" (published 2026-10-16T20:14Z),
   `CHANGELOG [0.14.0]` complete, `docs/releases/v0.14.0-release-notes.md`
   present, `wise.Version = "0.14.0"`. TODO_LIST still said "Cut v0.14.0" as
   open work and ROADMAP said "staged in [Unreleased]" — the exact split-brain
   class this skill exists to kill.
3. **Surface ground truth re-derived:** 46 `*Client` methods across 17
   resources (`go doc -all`; matches FEATURES/audit; count gate-checked by
   doc-verify), 36 `Example*` funcs, `TransferPayment`/paperwork signatures
   confirmed at source.
4. **ANNOTATE (73 verdicts): the 2026-09-16 status report** — every numbered
   item in §b (5), §c (4), §e (7), §f (50/50), §g (3) now carries an inline
   verdict via the skill's `annotate.go` (dry-run first; shipped items
   verified against code/gates, the rest routed or decided-against; §d left
   bare by design — it is the process-failure record, and the so-what test
   says annotating confessions is noise).
5. **ANNOTATE (27 verdicts): the twelve-factor review** — all factor-level
   recommendation lists (§II.1-3, §III.1-4, §V.1-4, §X.1-4) and the P1-P4
   action plan now carry inline verdicts; used the line-addressed `repair`
   path after learning the prose tool's section-end rule the hard way (§d.4).
6. **ARCHIVE with gates BEFORE the move:** the 2026-09-16 report →
   `docs/status/archived/` (48th file; `check-rows` clean first); the
   twelve-factor review → `docs/reviews/archived/` (new dir) with an archive
   manifest README. `docs/status/` now holds only point-in-time 2026-1x
   snapshots. Manifest entries follow the one-line-per-file rule.
7. **Banner honesty restored on the two kept reviews:** api-docs-study got a
   "Surface update (2026-10-17)" blockquote (46/17, paperwork trio,
   originator pair, 2027Q1 account-details); the implementation plan's banner
   now says v0.14.0 and tier-1 row 5 is struck `~~PLANNED~~ shipped v0.9.0`
   in the table itself. The v1.0 audit and open-banking comparison verified
   current (SKIP), the HTML artifacts LEAVE-ALONE.
8. **Living docs brought to v0.14.0+ truth:**
   - **TODO_LIST** — deleted the done rows (the `[x]` error-system epic, the
     stale "Cut v0.14.0", three struck-done P3 items); added a **"Cut
     v0.15.0"** P1 row (the structured error surface sits in `[Unreleased]`
     and is breaking); fixed "44-method" → 46 in the v1.0 row; split the
     accidentally-glued all-systems/go-1.27 items into two proper rows;
     compressed the spec_drift row to its open remainder (go/parser rewrite);
     pointed benchstat-consume at the v0.14.0 baseline; added the buildflow
     micro-batch row and the `BUILDFLOW_NO_RESULT_CACHE` decision row
     harvested from the 01-55 §f; fixed the MD018 line-start `#15/#16`.
   - **ROADMAP** — vision + Axis 1 + Axis 4 now say 46 methods / 17 resources
     (was 44/16 in three places); release state through v0.14.0 with the
     error surface correctly "staged"; "Shipped (unreleased)" → "Shipped in
     v0.14.0"; release-strategy line extended; six new raw ideas (retry-
     composition + transport-redaction README notes, Quote residual fields,
     Money property tests, CI fuzz smoke job, repo polish incl. the
     ISSUE_TEMPLATE config.yml and the session-claim convention).
   - **README** — status callout now sells the v0.14.0 additions (payments
     list, US combined receipt, NOC, third-party originator transfers); new
     "Transfer receipts & payout paperwork" section with the three new
     snippets; third-party create example; Project Status is now a docs
     table (FEATURES/ROADMAP/CONTRIBUTING/CHANGELOG/`docs/adr/`/
     `docs/error-codes.md`). **Every new snippet scratch-compiled against
     master** per the AGENTS compile rule.
   - **AGENTS.md** — Money bullet now cites ADR 001; the stale "2027Q1 spec
     moved account-details" gotcha rewritten (it still claimed the migration
     was pending — the worst lie found this session).
   - **CHANGELOG** — `[Unreleased]` gained `checks.script-fixtures` and the
     `docs/bench/2026-10-16_v0140_baseline.txt` entries (the TODO deletions
     required the CHANGELOG to hold them; append-only respected).
   - **FEATURES** — verified current (no upsert needed beyond anchor
     re-pins); **CONTRIBUTING** — markdownlint on-demand section added with
     an execution-tested command.
9. **Fix-on-sight batch:** the stale linter-allowlist error message in
   `flake.nix` (still citing the pre-skip auto-configure re-adds); two drafts'
   stray tails (bare fence + Crush attribution at EOF); the `New` godoc now
   documents the empty-key no-Authorization behavior (twelve-factor §III.4's
   "consider" item resolved by writing the missing doc, not by claiming it
   existed).
10. **All gates green** (list in the header). The one gate failure during the
    run was **anchor-check catching my own godoc edit** drifting two
    FEATURES.md anchors (`client.go:180/190` → `185/202`) — fixed, re-run,
    all green. That is the gate working; the failure was mine (§d.1).

## b) PARTIALLY DONE

1. **api-docs-study body tables** — the banner is refreshed, but §2's tier
   tables still carry v0.8-era "wise-go Status: PLANNED" columns. Left as a
   historical snapshot under a current banner (the FEATURES matrix owns live
   status); a maximalist per-row strike pass could archive-grade it.
2. **Implementation-plan tier tables** — only row 5 is struck; tier-2/3 rows
   have no per-row verdicts (the banner carries the truth). Same deliberate
   line as (1).
3. **AGENTS.md size** — the prune did not advance this session (my edits are
   net-positive bytes). Still ~56 KB: over the 30 KB flag and the checklist's
   > 50 KB fail line. The dedicated compression pass (T87) remains queued.
4. **TODO_LIST P1 composition** — "P1 — Release readiness" now holds one
   executable row (Cut v0.15.0) and two key-gated rows (sandbox pass, OTT
   flip). Arguably the gated pair belongs under P2; left in P1 because they
   gate the release train. Needs a ruling (§g.2).
5. **09-16 report §c.5 and §d** — deliberately bare (factual statement /
   confessions). A uniform-marker reader will notice; the banner explains.
6. **Daemon-commit verification** — spot-checks (content greps per batch),
   not a full diff read of all ~30 swept files.

## c) NOT STARTED (observed, none executed — docs-only session by design)

1. **All user-gated items** (TODO_LIST P2): CI enable + first green; sandbox
   key lane (5-assumption pass + OTT 2026Q4 probe/flip); CACHIX/ERRAUDIT
   tokens; v1.0.0 tag; typed recipient `Details`; RetryAfterDelay freeze;
   darwin matrix; flake pinning strategy; `BUILDFLOW_NO_RESULT_CACHE`
   default.
2. **The new buildflow/infra micro-batch** (routed this session): golangci
   full-diff vs pre-`ce52551`; md-go-validator nested-fence defect; BuildFlow
   #64 reproduction comments; legacy `buildflow.db` cleanup; `interrogate`
   disposition.
3. **AGENTS.md <30 KB compression pass**; **go/parser extractor rewrite**;
   **nixpkgs 1.26.9 security window** (lock at 1.26.8 as of today).
4. **No Go source changed** except the one godoc comment — the codebase is
   exactly as v0.14.0 + [Unreleased] error system left it.

## d) TOTALLY FUCKED UP (honest accounting — nothing shipped is broken; these are my process failures)

1. **I drifted two FEATURES.md anchors with my own godoc edit and only found
   out via the full `nix flake check`.** The AGENTS gotcha documents this
   exact class ("file.go:NN evidence anchors rot silently… the SCA-typing
   session shifted declarations") and `nix run .#anchor-check` takes seconds —
   running it right after touching `client.go` was the obvious move, and I
   ran the heavy battery first instead. One wasted multi-minute flake cycle;
   zero shipped damage.
2. **My archive-manifest edit deleted a heading.** I used the section heading
   itself as `old_string` without re-including it in `new_string`, silently
   removing "## 2026-10-16 docs-health archive sweep" from the archived
   README. Caught on the post-scripted-write re-view (the rule exists because
   of incidents like this) — but the edit spec was sloppy.
3. **A TODO multiedit consumed a still-open item.** Replacing the
   benchstat-consume row with the new micro-batch row DELETED open work; I
   caught it on immediate reflection and re-added it. The carry-every-concept
   rule applies to TODO rows, not just AGENTS bullets — I knew that and
   didn't apply it in the moment.
4. **Three wasted annotate calls on a scoping misunderstanding.** I assumed
   `### P1:` scopes to the next heading of any level; the tool scans to the
   next same-or-higher level, so "found 3" (P1+P2+P3 all have a "1." item).
   P3 then "succeeded" for an accidental reason (P4 has no numbered list) —
   I verified the diff AFTER applying, not before. The line-addressed
   `repair` path was the right tool from the second attempt onward.
5. **A wrong diagnosis almost became a bulk rewrite.** I "found" leading
   spaces on the P1/P2 list lines by misreading awk's `NR:` separator as
   indentation, then ran a python pass to strip them — the pass was a no-op
   (lines were already flush; the real problem was the scoping rule). A bulk
   scripted rewrite based on an unverified byte-level diagnosis is the
   doc-damage triage anti-pattern (`cat -A` first) that I documented for
   others this very session.
6. **Two fabrication near-misses, caught by verify-before-writing:** I first
   wrote a markdownlint command with markdownlint-cli2's `#`-ignore syntax I
   had never run (caught by deciding to test it — the real CLI takes
   `--ignore` flags, and testing surfaced a real MD018 bug); and I drafted a
   CHANGELOG-adjacent verdict "empty-key documented in README Authentication"
   before checking — the README did NOT document it, so I wrote the missing
   godoc instead of the false verdict. Both catches worked; both drafts
   should never have existed.

## e) WHAT WE SHOULD IMPROVE

1. **"Touched a .go file → run `nix run .#anchor-check` immediately" must be
   mechanical.** It is seconds-cheap, catches the drift class the same minute,
   and would have saved the wasted flake cycle. Consider a git hook or a
   pre-release-chain reorder (anchor-check earlier) if discipline won't hold.
2. **Edit specs: never consume a heading/row you intend to keep.** When the
   edit target IS the anchor text, re-include it in `new_string` (or use
   add-before/add-after). Applies to headings (§d.2) and TODO rows (§d.3) —
   same root cause: replace-without-carry.
3. **Learn the tool's actual scoping before batch specs** — one `--dry-run`
   per SECTION shape, not per file; and prefer `repair` (line-addressed)
   whenever a file has several numbered lists with restarting ids.
4. **Byte-level proof (`cat -A`) before any scripted bulk rewrite** — the
   awk-misread in §d.5 is exactly the triage rule I cited in the same pass.
5. **Verify CLI syntax before it enters docs** — the tested markdownlint
   command is now execution-verified AND caught a real defect; keep that
   ordering mandatory.
6. **Carry-every-concept applies to TODO deletion too:** before deleting a
   TODO row, grep its nouns in CHANGELOG/AGENTS/ROADMAP to prove the knowledge
   survives somewhere (the park-landed pre-check, applied to myself).
7. **Status-report timing honored** (report written after the tree was final,
   per the 01-55 lesson d7) — keep it.

## f) Up to 50 things we should get done next (impact-ordered; ⏳ = user-gated; most are PRE-ROUTED to TODO_LIST/ROADMAP this session — marked)

> Sources: this session's verification + the routed queues it left behind.
> Items without a "routed" mark are the residue THIS report adds.

1. ✅ ~~**Cut v0.15.0** — the structured error surface in `[Unreleased]` is~~ SHIPPED 2026-10-17: tag `415cce4` (pre-release chain exit 0, GitHub Release live, proxy + clean-dir `go get` verified); the original
   breaking (400s → `*BadRequestError`, 11 sentinels, `Details`); pre-release
   chain + release notes + bank-sync recipe pointer (routed: TODO P1).
2. ⏳ **Security window: nixpkgs → go 1.26.9** when it lands (9 stdlib vulns
   GO-2026-6603..6617); re-verify vendor FODs + full buildflow (routed:
   TODO P3).
3. ⏳ **CI resurrection** — `gh workflow enable ci`, first green, badge
   unfreeze, cachix/erraudit disposition, art-dupl blocking decision (routed:
   TODO P2).
4. ⏳ **Sandbox key lane** — 5-assumption live pass, OTT `2026Q3→2026Q4`
   probe + flip, first recorded sandbox-live run (routed: TODO P1/P2).
5. ⏳ **v1.0.0 tag on approval** — audit lineage verified at 46 this session;
   one command on your word (routed: TODO P2).
6. ⏳ **Typed recipient `Details` decision** (routed: TODO P2).
7. ⏳ **`Health` code-half decision** (routed: TODO P3 Decisions).
8. ⏳ **Spec-drift `--fail-on-drift` posture** (routed: TODO P3 Decisions).
9. ⏳ **Flake pinning strategy (rev vs tag)** (routed: TODO P3 Decisions).
10. ✅ ~~**Darwin check matrix**~~ APPLIED 2026-10-17 (D5 option b under the train's blanket execution order, reversible): `go-standard.systems` 3-system list, `--all-systems` exit 0 (`03ad312`).
11. ⏳ **`BUILDFLOW_NO_RESULT_CACHE` standing default** (routed: TODO P3
    Decisions, new this session).
12. ✅ ~~**Buildflow/infra micro-batch**~~ DONE 2026-10-17 (T56–T60: only `lll` lost; nested-fence filed as md-go-validator#11; #64 reproductions measured + commented; legacy db deleted after mtime proof; interrogate skip_steps'd) —
    md-go-validator nested-fence defect (verify + file); #64 reproduction
    comments; legacy `buildflow.db` cleanup; `interrogate` disposition
    (routed: TODO P3, new this session).
13. 🟡 **AGENTS.md <30 KB compression pass** — PART 1 done 2026-10-17 (`a84bc35`: 57,097 → 47,354 bytes, zero constraint loss, one split-brain fixed); the second pass to <30 KB stays open — bullet-by-bullet under
    carry-every-concept; now also carries a rubric FAIL (>50 KB) on the
    verify-checklist (routed: TODO P3; the rubric line is new).
14. ✅ ~~**Rewrite `spec_drift.py` extractor on `go/parser`**~~ DONE 2026-10-17 (`92b8f2f`: AST extractor parity-equal at 43 endpoints, fixtures 9/9, live drift exit 0).
15. ✅ ~~**Retry-composition + transport-redaction README notes**~~ DONE 2026-10-17 (`bef6e2a`; routed: ROADMAP
    raw ideas, new this session).
16. ✅ ~~**`Quote` residual wire fields**~~ DONE 2026-10-17 (all four incl. `preferredPayIn`; corruption-pinned; `bd584ca`) — original:
    `targetAmountAllowed`, `user` (routed: ROADMAP, new this session).
17. ✅ ~~**Property tests for `Money` cents conversion**~~ DONE 2026-10-17 — the fuzz found a REAL boundary bug (non-representable amounts mapped to garbage cents) and `toMoney` now rejects them as Corruption (`2e64a4d`).
18. ✅ ~~**CI fuzz smoke job**~~ SPEC'D 2026-10-17 (`4d3343d`: auto-discovering 3s-per-target job landed in ci.yml; activates with the workflow enable).
19. ✅ ~~**Repo polish** — `.github/ISSUE_TEMPLATE/config.yml`~~ DONE 2026-10-17 (`7e51ede`, discussions enabled repo-side) — original:
    toggle); session-claim convention for parallel sessions (routed: ROADMAP,
    new).
20. ✅ ~~**api-docs-study §2 tier-table maximalist strike pass**~~ RESOLVED 2026-10-17 (`963d094`): banner-grade re-affirmed with a disposition footer — per-row strikes would duplicate FEATURES/TODO maintenance (§g1 was unanswered; conservative default taken).
21. ✅ ~~**Implementation-plan tier-2/3 per-row verdicts**~~ RESOLVED with #20 (same disposition, `963d094`).
22. ⏳ **TODO P1 retitle/ruling** — key-gated rows under P1 vs P2 (new; §g.2) — v0.15.0 train kept the ruling untouched; still awaiting.
23. ✅ ~~**DOMAIN_LANGUAGE: error-system terms**~~ DONE 2026-10-17 (`73003c5`: WiseErrorCode, WiseErrorDetail, code sentinel, status-driven typed errors). Original observation:
    `Details` envelope semantics are new domain vocabulary since the 00-07
    session verified 22/22 terms (new observation this session).
24. **FEATURES external-claims cadence** — the 215-operation live-preview
    census is dated 2026-09-16 (within quarterly cadence, but schedule the
    re-sweep; new observation).
25. ✅ ~~**Verify the daemon swept the FEATURES anchor re-pin**~~ VERIFIED (daemon commits carried it; anchors green through the session, incl. re-pinning four MORE drifted anchors after the Quote/money edits — caught by the buildflow full run, `846b3bb`). Original:
    uncommitted at report time).
26. ✅ ~~**Run the pre-release chain before the v0.15.0 tag**~~ DONE (exit 0 at the tag, T15). Original:
    network `spec-drift` + `apidiff` legs (ritual).
27. **Bank-sync pin bump → `docs/error-codes.md` recipe follow-through**
    (their repo; AGENTS gotcha carries the pointer).
28. ✅ ~~**Annotate THIS report inline when its f-items ship**~~ DONE — this annotation pass IS that ritual (2026-10-17, v0.15.0 train M27). Original:
    annotate-then-archive ritual this session just performed on its own
    predecessor (ritual).
29. **Fleet pin-drift sweep** — port `checks.pin-sync` to the sibling flakes
    (pre-existing ROADMAP raw idea; untouched).
30. ✅ ~~**Consume the v0.14.0 benchstat baseline in the next perf-touching PR**~~ RESOLVED 2026-10-17: the v0.15.0 baseline was CUT (`e9d1f41`, geomean −1.85% vs v0.14.0, quiet-machine hygiene rule added) and the M25 verdict recorded — no benchmarked-path change shipped, nothing to compare; the convention is armed. Original:
    (routed: TODO P3).

## g) Three questions I cannot figure out myself

1. **Archive-grade or banner-grade for the two kept reviews?** The
   api-docs-study and implementation plan now have current banners but
   v0.8-era body tables (§b.1/§b.2). A maximalist per-row pass would let them
   be archived like their twelve-factor sibling; leaving them banner-grade
   keeps zero-risk canonical gap maps in place. Which do you want the next
   sweep to do?
2. **TODO_LIST P1 composition:** the sandbox pass and OTT-flip rows are
   key-gated (really P2) but gate the release train; "Cut v0.15.0" is the
   only truly executable P1 row. Move the gated pair under P2, or keep P1 as
   "release-train blockers regardless of gate"?
3. **AGENTS.md size policy:** the verify-checklist flags >30 KB and fails
   > 50 KB; the file is ~56 KB and the <30 KB target needs a dedicated
   > compression session with real constraint-loss risk. Prioritize that pass
   > now, or after v0.15.0 ships? (I will not bulk-compress without your word —
   > every bullet there was paid for.)

---

_Point-in-time snapshot. Status reports go stale; re-verify before acting on
them. The auto-git daemon will pick up this file._
