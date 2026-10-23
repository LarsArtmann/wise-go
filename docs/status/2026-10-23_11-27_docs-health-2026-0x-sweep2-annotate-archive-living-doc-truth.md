# Status Report — Docs-Health 2026-0x Sweep #2: Annotate/Archive + Living-Doc Truth + Harvest (2026-10-23 11:27 CEST)

_Point-in-time snapshot #4 of the day. Prior reports this day: `07-22` (split-brain sweep), `07-48` (buildflow recovery), `08-25` (continuation backlog execution — its §c1 HARVEST gate is satisfied by THIS session). Scope discipline: everything below is from THIS session's run and what it surfaced — no new research. Format: `.md` at explicit user path (skill's HTML default overridden, standing dispatch-report exception)._

**Session scope:** the user ordered "View ALL `**/2026-0*` files! Execute the docs-health SKILL! … Archive FULLY done and UPDATED (inline strikethrough) .md files!" — then this report. **No Go source changed**; edits are markdown + one flake.nix gate line.

**End state at report time:** dprint 0 · `nix fmt` 0 · `nix run .#anchor-check` **101/101 across 8 living docs** (first try this time — the 03:36 lesson applied) · `nix run .#doc-verify` **all checks passed** (46-method count claims green on the four remaining anchors) · `nix flake check` **exit 0** (also closes the 08:25 report's §c8: the prior session's config-only edits verified inert) · 13 modified/moved paths unswept (daemon idle since `47b18aa`).

---

## Brutal Self-Review (the three questions asked)

### What did you forget?

1. **Cells whose text starts with a literal `~`.** The AGENTS gotcha says strikethrough needs EXACTLY two tildes — I had just re-read it — and I still wrapped `"~15 endpoints"` as `~~~~~15 endpoints~~~` (five tildes) in the study's tier-3 table, then did it again with `~~~~` + backtick on two tier-4 rows of the plan. The exact corruption class the gotcha documents, twice in one pass. Caught both by my own `grep '~~~~'` check immediately after each edit, never shipped to a gate red.
2. **Scripted table rewrites are the damage class.** My python strike pass for the plan's tier-1 table built lines with `'|' + '|'.join(cells)` — double-piping the row start. dprint's realignment then silently ATE the verdict cell and shifted the row (empty first cell). I then misdiagnosed the repair TWICE: ran a second scripted pass whose `s.index()` failed on a dprint-reformatted header ("substring not found") and mistook the stale check-rows output for a live failure; and submitted an identical no-op edit ("all edits resulted in identical content"). The fix that worked first try was a hand-written whole-block replacement — the scripted approach was wrong from the second it touched a dprint-aligned table.
3. **Formatter invalidates reads.** Two edit-tool rejections ("file modified since read") because I kept editing after dprint runs without re-reading. Mechanical rule broken twice: format LAST, or re-read after every format.
4. **Wrong-file edit.** I aimed the six-item "Out of Scope" strike at the open-banking comparison; that list lives in the implementation plan. 1 of 3 multiedits failed harmlessly — but it was a read-the-file-first failure, not a typo.
5. **The check-rows "struck table" rule.** I learned mid-pass that ONE struck row makes every CLEAN sibling row an offender — meaning half-struck tables must go fully uniform. My first repair round struck only the flagged rows and the gate stayed red. The invariant is now recorded in §e so the next sweep starts with it.

### What could I have done better?

- Dry-run the strike shape on ONE tier-3 row (the `~`-leading one) before batch-editing — the exact annotate-tool lesson (`--dry-run` the first spec against a new file shape) applied to hand edits.
- Hand-write table blocks from the start. Every hand-written block (tier-2/3/4, phases, banners) passed check-rows immediately; both scripted passes damaged shape. The tool exists for bulk; tables under dprint alignment are not bulk.
- Confirm the working-tree state BEFORE planning edits: the 08:25 session's 5 modified files + probe deletion + untracked report were still unswept when I started (conversation-start "clean" was stale). I discovered this only when compiling the report — earlier would have changed nothing, but "check git status first" is a rule I cite and skimmed.
- The v1.0 audit archive changes a flake gate's input. I ran `nix flake check` at the end anyway, but repointing/removing a gate line deserved the check in the same breath as the edit, not as part of the final battery.

### What could I still improve?

- The annotation shape for outcome-bearing historical tables (my `~~cell~~ verdict` form) is now consistent across four files but exists only in this session's output — extracting it into a short convention note (or the annotate tool's docs) would keep sweep #3 from re-deriving it.
- "Archive-grade" still has a gray zone I resolved by judgment (informational tables left bare under an archive verdict — §b1). A one-line ruling in the archive manifests would remove the judgment call.
- The daemon remains unobservable from inside a session; I benefited from its sweep (`9eb0167` carried the 08:25 changes + probe deletion + my first batch, `--stat`-verified) but can only infer liveness from `git log` timestamps.

### Did I lie / oversell?

No. Every verdict in the archived files cites a version or a FEATURES routing I checked against CHANGELOG or Go source this session (ship versions read from CHANGELOG section heads; `cashAccountType` absence verified by grep, not assumed; 46/136/31 re-measured via `go doc -all` before writing the audit delta). The one soft spot: "all gates green" at the top of this report refers to the tree as it stood pre-sweep — the daemon may commit the 13 pending paths after this line; nothing in them was unverified at write time.

---

## a) FULLY DONE

1. **All live `**/2026-0*` files viewed and classified.** 4 markdown reports read in full (api-docs-study, v1.0-api-audit, open-banking-comparison, the 2026-08-19 implementation plan); 7 generated HTML/D2 artifacts (2026-07-18 era) re-confirmed LEAVE-ALONE per the AGENTS gotcha; the three `archived/` trees (48 status + 8 planning + 2 review files) NOT opened — closed history per the docs-health model.
2. **ANNOTATE: api-docs-study** — every §8 next-step struck with its ship version (1→v0.8.0, 2→**PARTIALLY, see #4**, 3/4/5/6/7/8/9/10→v0.8.0–v0.14.0, 11–13→routed with destinations); tier-1/2/3 tables fully struck (18 rows, whole-cell strikes, per-row verdicts: shipped vX / PLANNED / DEMAND_GATED / OUT_OF_SCOPE per the FEATURES matrix); banner extended to v0.16.0; the 2026-10-07 + M22 "keep as canonical gap map" dispositions superseded by an explicit archive verdict citing the owner's order.
3. **ANNOTATE: implementation plan** — banner updated to v0.16.0 + archive verdict; tier-2 rows 14–22 struck (shipped, with version + note), row 23 struck-routed (comparisons stays FEATURES-PLANNED); tier-3 rows 24–35 struck (row 24 superseded — `/v2` GetProfile; 30/31 shipped-with-2027Q1-migration; rest DEMAND_GATED); tier-4 rows 36–51 struck (42: OTT shipped v0.11.0 with sessions/factors OUT_OF_SCOPE; 49: profile-level CRUD shipped v0.10.0, app-level ON_HOLD; rest OUT_OF_SCOPE); all four phase bullets struck; the six-item out-of-scope list struck with a still-true confirmation note.
4. **HARVEST FIND (genuine lost work):** api-docs-study §8.2 ("expand `Balance` struct") never shipped AND never routed — `cashAccountType` appears NOWHERE in the Go source (not even raw), `InvestmentState` drives list filtering but is absent from the public `Balance` struct, `bankDetails` unexposed. Item 2 carries a PARTIALLY verdict and the work is now routed: ROADMAP raw idea "`Balance` residual wire fields" (alongside the Quote sibling that shipped v0.16.0 — now struck SHIPPED in ROADMAP).
5. **ANNOTATE: v1.0-audit** — new delta blockquote: surface RE-MEASURED via `go doc -all` (46 `*Client` methods / 136 types / 31 const groups across 17 resources — unchanged through v0.16.0), archive verdict added; §2 godoc findings (all FIXED/Clean at authoring), §3 freeze register, and §4 (already struck) verified; archived.
6. **ANNOTATE: open-banking-comparison** — point-in-time note refreshed to v0.16.0 with the two superseded snapshot claims named explicitly (failsafe-go → go-retry per ADR 003 v0.11.0; "read-only" → the own-account transfer flow is the SDK core since v0.8.0); the false "no payment capabilities" line struck-and-corrected inline; zero open items; archived.
7. **ARCHIVE with the gate BEFORE the move:** `check-rows` exit 0 on all four files (after repairing three pre-existing PARTIAL rows — the 2026-09-13-era strike-the-fragment-inside-the-cell style — into uniform whole-cell strikes), THEN `git mv` to the sibling `archived/` dirs. Completeness gate `grep -L '~~'` over the four moved files prints nothing.
8. **Archive manifests updated** (bulk-archive manifest rule): `docs/reviews/archived/README.md` gained the 2026-10-23 sweep section with one-line-per-file classifications (3 files); `docs/planning/archived/README.md` gained the plan's entry. Both formatted.
9. **Flake count-gate de-anchored from the archived audit** — the `check_count` line reading `docs/reviews/2026-08-21_v1.0-api-audit.md` removed with an in-file rationale (an archived file is frozen; a future method-count change could never satisfy it). The remaining anchors (AGENTS, FEATURES, ROADMAP, latest release notes) still enforce 46 — verified by doc-verify.
10. **Citation sweep: zero dangling references.** All living-doc pointers repointed: AGENTS "API docs study" gotcha (3 paths), FEATURES plan link, ROADMAP ×4 (Axis-1 plan path, Vision audit path, release-state audit path, Axis-2 audit path), TODO_LIST v1.0 row, `2026-10-16_transfer-category-reaudit.md` baseline pointer (inline-corrected with an archive note). The one remaining old-path hit is CHANGELOG's v0.8.0-era historical entry — append-only, frozen by design. Verified by grep over all 9 living docs + flake.nix.
11. **TODO_LIST rebuilt as the open-work surface:** the stale "security window" row DELETED — its premise was false: the nixpkgs lock is at `02c087f` (go 1.26.9) since the 07:48 session's repair (verified in `flake.lock`, not trusted from prose), the checklist ran organically, and the stdlib advisories were recharacterized as directive-keyed (48 findings / 10 advisories, warning-class). Ten rows harvested from the three 2026-10-23 reports, each with source pointers: BuildFlow binary rebuild + daemon observability (P2 user-gated); dprint daemon-gap closure, quoted-title gate, buildflow count-inflation reconcile, upstream filings (P3); vulnix replacement, go-directive policy, AGENTS structural split (P3 Decisions). Nothing dropped — every open f-item from all three reports is either routed here or confirmed closed with evidence.
12. **ROADMAP brought to v0.16.0 truth:** Vision header ("As of 2026-10-23"), release-state paragraph (v0.15.0 + v0.16.0 sentences), release-strategy chain extended through v0.16.0, the audit citations repointed; the stale "Sandbox verification — blocked on a sandbox API key" Medium-term bullet REWRITTEN as SHIPPED (the lane went live 2026-10-17 — the old text was a split-brain against FEATURES/CHANGELOG); Quote-residual-fields idea struck SHIPPED v0.16.0; new Balance-residual-fields idea added (see #4).
13. **CHANGELOG `[Unreleased]`** now carries the 2026-10-23 infra surface (nixpkgs lock bump + directive-vs-toolchain advisories explanation, vulnix skip, dprint excludes, allowlist extensions, B607 fix) — append-only respected, prior sections untouched.
14. **AGENTS.md:** the API-docs gotcha repointed with the endpoint-count timeline COMPRESSED to current-count + lineage pointer (executes f#14 of the 03:36 report — the "grown to 37, then 41, then 46" narrative is gone); the daemon gotcha extended with the never-stage-scratch-probes rule (evidence: `eb62699`); two new Build & Dev entries — the living-doc edit gate (scoped dprint → anchor-check → doc-verify → flake check) and the write-time one-fact-one-place discipline (both born from the 07:22/07:48 sessions' findings).
15. **Cross-file fact sweep** (07:22 f#4): the 8 consolidated facts (erraudit gate command, quarterly body-shape probing, DelayFunc mechanism, md-go-validator #8 story, auto-configure skip, anonymous `%w` idiom, Money contract, CI-disabled state) — grepped across README/FEATURES/TODO_LIST/ROADMAP/CONTRIBUTING: NO stale copies; only legitimate own-job references (FEATURES' DISABLED ci.yml row, CONTRIBUTING's aligned auto-configure warning). Closes f#4 of the 07:22 report.
16. **Money entry re-verified against code** (07:22 f#5): `Money{Cents int64, Currency Currency}` (`types.go:65`), `toMoney` (`helpers.go:61`), `NewCurrency` (`types.go:49`), `ListTransactionsRequest.Currency Currency` (`types.go:194`), ADR 001 present. The merged AGENTS entry is TRUE.
17. **State-claim sweep** (07:22 f#24): "CI is disabled"-class claims checked across living docs — AGENTS owns the gotcha, FEATURES' DISABLED row and TODO's conditional usage are consistent. Clean.
18. **All gates green** (see header) — including the first-try anchor-check pass and the full `nix flake check` that also retroactively verifies the 08:25 session's config edits.
19. **Daemon-sweep verification:** `9eb0167` (7 files) contains exactly the 08:25 session's verified changes + the intentional probe deletion + my first annotation batch (audit delta) — `--stat`-inspected; `b558525`/`38b857f`/`47b18aa` carried the subsequent batches. The docs-integrity "gutted file" warning's cause (pending probe deletion) has landed.

## b) PARTIALLY DONE

1. **"Archive-grade" uniformity in the four archived files** — every WORK item is struck, but outcome-bearing tables (study §1 changelog/§7 actions tables, §3/§4 findings tables beyond the repaired rows) and the plan's quality-gate checklist + execution-order list remain bare under their archive verdicts. Deliberate (they record what was done, they are not open items), but a uniform-marker reader will notice bare tables in `archived/`. §g1 asks for the ruling.
2. **TODO_LIST P1/P2 composition ruling** (carried from the 03:36 report §g.2, never answered): key-gated items (sandbox browser attempt, CI tokens) sit in P2 "User-gated" while the release-train framing lives in row text. I preserved the status quo; the question is re-asked in §g2.
3. **AGENTS.md size** — my additions are net-positive bytes (~+1.5 KB; now ~54 KB): the <30 KB prune (T87 part 2) did not advance this session. Still over the verify-checklist's >50 KB fail line.
4. **Daemon observability** — improved only by inference: liveness confirmed via `git log` timestamps + `--stat` verification; the process/heartbeat/hook-bypass-intent questions remain (P2 row).
5. **Working tree at report time** — 13 modified/moved paths (6 living docs, flake.nix, 2 manifests, reaudit pointer, 4 renames) gate-green but unswept; daemon idle since `47b18aa`. The daemon's docs-integrity preflight will see renames + edits until it wakes.
6. **Pointer-title exactness** — my new AGENTS cross-references use exact bold titles; the pre-existing informal pointers ("its own gotcha", "the dedicated gotcha above") left alone per the 08:25 ruling, still without a mechanical gate.

## c) NOT STARTED (observed, none executed — docs-only session by design)

1. **All user-gated rows** (TODO_LIST P2): bank-sync v0.16.0 pin push; sandbox assumptions 1+3 browser attempt; v1.0.0 tag; CACHIX/ERRAUDIT tokens; CI enable + first-run watch; BuildFlow binary rebuild; daemon observability.
2. **All P3 user-decision rows**: vulnix replacement; go-directive advisory policy; AGENTS structural split; Health code-half; spec-drift posture; flake pinning; BUILDFLOW_NO_RESULT_CACHE default; x86_64-darwin drop; Go 1.27 (conditions unmet).
3. **Full buildflow run after THIS session's flake.nix touch** — the AGENTS rule fires after LANDING commits that touch `flake.nix`; the daemon hasn't swept yet, so the run is queued, not skipped. Same pass: on-demand gitleaks/codespell (last verified clean 2026-10-17; the 08:25 session re-verified after ITS changes, mine are newer).
4. **The routed implementation tasks**: quoted-title gate, dprint vendored-WASM sandbox check, count-inflation reconciliation, upstream filings (nixpkgs/vulnix NVD-404 with verify-before-filing legwork; crush-config scoped-dprint step) — routed to TODO_LIST, zero code.
5. **No Go source changed** — the codebase is exactly as v0.16.0 left it.

## d) TOTALLY FUCKED UP

1. **I introduced the documented corruption class I had just re-read.** Five-tildes-in-a-cell (study tier-3, 4 cells) and four-tildes-before-a-backtick (plan tier-4, 2 cells) — both from naively wrapping text that starts with a literal `~`. The AGENTS gotcha exists since the 2026-10-07 formatter incident. Both caught by my own `grep '~~~~'` sweeps within minutes; zero shipped damage; the rule cost me nothing but should have cost me nothing AT ALL.
2. **A scripted rewrite + formatter interplay ate table cells, and my first two "fixes" were misdiagnoses.** The double-pipe join bug was mine; dprint's cell-drop under a mismatched row was the interplay; but the real failure was confirmation-biasing the repair: I re-ran check-rows, saw the same PARTIAL list, and assumed my second script had fixed it (it had failed to find its target — "substring not found" — leaving the file untouched); then I burned an edit on an identical no-op. Three attempts for what a hand-written block replacement solved first try. The lesson is §e2, and it is now paid for.
3. **Stale-read edits.** Two edit rejections (mod-time newer than last read) because dprint ran between my read and my edit. Mechanical negligence, not judgment — but in a session whose subject is documentation discipline.
4. **Wrong-file edit target.** The out-of-scope strike aimed at the open-banking file; the list lives in the plan. Harmless (the edit failed loudly), but it was a failure of read-before-edit, the rule at the top of every session.
5. **A conversation-stale assumption I nearly planned around.** The session-start snapshot said the working tree was clean; it held the 08:25 session's unswept work. I found this mid-session only incidentally (via `git status` for the probe check). No damage done — but planning edits in a repo whose state I hadn't verified is how daemon races bite.

## e) WHAT WE SHOULD IMPROVE

1. **Tilde rule, operational form:** before striking any cell, transform leading tildes into words — `~15` → `approx. 15` INSIDE the strike; never nest or neighbor `~~` with content tildes. Candidate one-line extension of the AGENTS strikethrough gotcha (owner endorsement welcome).
2. **Tables are hand-written territory under dprint.** Scripted multi-cell rewrites produced every damage instance this session; hand-written whole-block replacement produced zero. Corollary: after ANY formatter run, re-read before editing (two rejected edits this session).
3. **check-rows uniformity invariant, now explicit:** one struck row obligates the WHOLE table ("CLEAN row in a struck table" is an offender class); repair order is format → check-rows → move. The 03:36 sweep's placement rule (gate BEFORE `git mv`) held and caught all damage pre-move — keep that ordering sacred.
4. **Archive-grade definition needs the gray zone settled** (§g1): work items struck + verdict-bearing/log tables bare = the shape all four archived files now have. If the owner confirms, note it in the archive manifests and the judgment call dies.
5. **State checks before plan checks:** `git status` at session start is the first command in daemon-managed repos — the conversation-start snapshot is a snapshot, not the tree.
6. **Anchor discipline held and is worth keeping:** `nix run .#anchor-check` ran right after the .go-adjacent batch (101/101 first try) — the 03:36 §e1 rule applied, zero wasted flake cycles this time.

## f) Next tasks (up to 50; ⏳ = user-gated; most are PRE-ROUTED to TODO_LIST this session — marked)

1. ⏳ **Rebuild the BuildFlow binary + global-profile swap** (#35 workaround; stale-binary verdicts are advisory-grade) (routed: TODO P2).
2. ⏳ **Daemon observability** — locate, heartbeat/log, hook-bypass intent (routed: TODO P2; also §g3).
3. **Close the daemon-commit format gap** — daemon-side post-commit `dprint check` OR vendored-WASM sandboxed flake check (routed: TODO P3).
4. ⏳ **Vulnix replacement decision** — grype/syft over `./result` vs govulncheck-only gap (routed: TODO P3 Decisions).
5. ⏳ **Go-directive stdlib-advisory policy** — patch-pin 1.26.9 vs wait for 1.27 (routed: TODO P3 Decisions).
6. **Quoted-title gate** — extend `anchor_check.py` to verify quoted entry titles (routed: TODO P3).
7. ⏳ **CI resurrection** — `gh workflow enable ci` + first-run watch (routed: TODO P2; runbook ready).
8. ⏳ **v1.0.0 tag** — audit green, growth lineage current; one command on your word (routed: TODO P2).
9. ⏳ **Bank-sync v0.16.0 pin push** + error-codes recipe follow-through (routed: TODO P2).
10. ⏳ **Sandbox assumptions 1+3** — one capped browser SCA attempt or bank the verdicts permanently (routed: TODO P2).
11. ⏳ **CACHIX_AUTH_TOKEN / ERRAUDIT_TOKEN secrets** (routed: TODO P2).
12. ⏳ **AGENTS.md structural split** — single file vs per-area `context_paths` (routed: TODO P3 Decisions).
13. **AGENTS.md prune part 2** — ~54 KB → <30 KB target, carry-every-concept (routed: TODO P3).
14. **Full buildflow run + on-demand gitleaks/codespell after the daemon sweeps this session's flake.nix touch** (new this session; the AGENTS infra rule fires on landing).
15. **Reconcile buildflow count inflation** — bandit 78-vs-1 characterization; "9 tools unavailable" PATH audit (routed: TODO P3).
16. **Upstream filings** — nixpkgs/vulnix NVD-404 issue (verify-before-filing first); crush-config scoped-dprint step (routed: TODO P3).
17. ⏳ **Typed recipient `Details` design decision** (in TODO P2).
18. ⏳ **`Health` code-half decision** (D2 memo written; TODO P3 Decisions).
19. ⏳ **Spec-drift `--fail-on-drift` posture** (D3; TODO P3 Decisions).
20. ⏳ **Flake pinning strategy rev-vs-tag** (D4; TODO P3 Decisions).
21. ⏳ **`BUILDFLOW_NO_RESULT_CACHE=1` standing default** (D6; TODO P3 Decisions).
22. ⏳ **x86_64-darwin drop** (D5 one-liner; TODO P3 Decisions).
23. ⏳ **Go 1.27 migration** — 4 unlock conditions (TODO P3; nixpkgs lock still on the 1.26 line for the module).
24. **Supersedes-chain compression in AGENTS** — 2026-10-11→17 rulings, auto-configure history (07:22 f#7–8).
25. **AGENTS taxonomy + mini-index** (07:22 f#11–12).
26. **Gotchas-vs-Conventions placement audit** (07:22 f#13).
27. **Build & Dev vs CONTRIBUTING command dedupe** (07:22 f#25).
28. **Tight-list markdown lint** (07:22 f#26).
29. **README OTEL section vs AGENTS entry consistency check** (07:22 f#27).
30. **incident-narratives pointer verification** (6 claims) + the "two runtime deps" vs three annotation (07:22 f#23, f#31).
31. **Schedule the FEATURES external-claims re-sweep** — the 215-op live-preview census is dated 2026-09-16; due ~2026-12-16 (03:36 f#24).
32. **Spec-refresh item: OTT 2027Q1 canonicalization + vendored snapshot refresh** (07:22 f#21–22).
33. **Wave-2 error-code sentinels** (conscious pin update; ROADMAP fuel, 07:22 f#29).
34. **`WithMetrics` trigger watch** (hold until consumer need; 07:22 f#30).
35. **Upstream watches**: md-go-validator #11 (nested fences), BuildFlow #65 ("9 tools"), BuildFlow #35 (routed #1) (07:22 f#35–37).
36. **Consume the v0.15.0 benchstat baseline in the first perf-touching PR** (convention armed; TODO P3).
37. **Balance residual wire fields** — parse when a consumer needs them (new ROADMAP raw idea, born §a4).
38. **docs-integrity warning disambiguation** — make "deleted intentionally, sweep pending" visible (08:25 e#8).
39. **Confirm the docs-integrity warning is gone now that the probe deletion landed** (08:25 f#10; verify after the next sweep).
40. **Stale `./result` guard or deliberate drop** (deprioritized 08:25; vulnix skipped).
41. **Fleet pin-drift sweep** — port `checks.pin-sync` to sibling flakes (standing ROADMAP idea).
42. **Release-process automation** — checklist app / RELEASE_CHECKLIST.md (standing ROADMAP idea).
43. **Coverage badge unfreeze after CI enable** (07:22 f#18).
44. **erraudit CI posture on re-enable** — blocking vs warn-not-fail (07:22 f#17).
45. **README webhook end-to-end quickstart** (standing ROADMAP raw idea).
46. **gitleaks allowlist shape** — commit-pin vs rule-scoped (08:25 f#13; low).
47. **Quarterly anchor re-verify + CONTRIBUTING titles into the title gate** (07:22 f#32–33; folds into #6).
48. **VERIFY-mode sweep convention** — every N weeks (07:22 f#34; today's run is the first datapoint).
49. **Wishlist residue** — Postman collection, currency-conversion helpers, batch APIs, mock server (demand-gated; ROADMAP).
50. **Post-hoc audit of THIS f-list against the worktree before the next session acts on it** (ritual, self-applied last two sessions).

## g) Three questions I cannot figure out myself

1. **Archive-grade uniformity ruling:** the four files I archived today strike every WORK item but leave outcome-bearing tables bare under their archive verdicts (§b1). Is that the archive-grade definition going forward (I'll note it in the manifests), or do you want literally every table struck so no bare row survives in `archived/`?
2. **TODO_LIST P-composition ruling (carried since the 03:36 report §g.2):** key-gated items currently sit in P2 "User-gated" while the release-train framing lives in row text. Keep P2 as "user-gated regardless of train impact", or introduce a "P1 — release-train blockers" bucket that may contain gated rows?
3. **Daemon contract:** where does the daemon live (process/unit/script), is the hook bypass intentional, and do you want a heartbeat file? This is the third report asking; it gates #3's design (a daemon-side check needs to know where to hook) and I cannot find it from inside a session.

---

_Point-in-time snapshot. Status reports go stale; re-verify before acting on them. All gates green at report time (dprint 0 · anchor-check 101/101 · doc-verify pass · `nix flake check` exit 0). The auto-git daemon will pick up this file._
