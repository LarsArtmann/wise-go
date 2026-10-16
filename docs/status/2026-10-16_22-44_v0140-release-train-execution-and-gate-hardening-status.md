# Status Report — v0.14.0 release-train execution, coverage rescue & archive sweep (wise-go)

> **Format note:** `.md` per explicit user instruction (the status-report skill's
> canonical format is styled HTML — same one-off override as the 2026-09-13/16,
> 10-05/10-07 reports; not propagated into the skill).

- **Date:** 2026-10-16 22:44 CEST
- **Scope:** This session only (≈21:58–22:44): execution of the 21:45 SUPERB
  plan (M1–M7, M13, M26-partial) in a tree with MULTIPLE concurrent crush
  sessions and the auto-commit daemon. Everything below is what THIS session
  did or directly observed; parallel-session work is labeled as such.
- **Tree at report time:** `2016e5f` on master, pushed; all local gates green
  (fmt, build, doc-verify, release-notes-check, markdownlint 0 findings).
- **Headline:** **v0.14.0 shipped and verified end-to-end** (tag → push → GitHub
  release → proxy → clean-dir `go get`), a REAL Nix-build breakage was caught
  by the M1 battery and fixed, the coverage floor breach (91.4% < 92%) was
  root-caused and rescued to **93.5%**, all eight kept 2026-10-07 status
  reports were annotated item-by-item and archived, and the on-demand
  markdownlint pass is now repo-wide clean with three real doc bugs fixed.

---

## a) FULLY DONE

1. **M1 release-inventory verify** — webhook payload decoders wired
   (`TransferActiveCases` webhooks.go:450, `TransferRefund` :471), originator
   create/get present (`third_party_transfers.go:95/:145`), 46-method surface
   confirmed, CHANGELOG gap list produced (4 missing entries, later closed by
   the parallel session's `[0.14.0]` cut). Gate battery: `go test -race ./...`
   exit 0, `golangci-lint` 0 issues, doc-verify / release-notes-check /
   spec-drift / pin-sync all exit 0.
2. **Nix fileset breakage caught + fixed (`47c301a`)** — the M1 battery's
   `nix flake check` was RED (`GetThirdPartyTransfer undefined`) while plain
   `go test` passed: the new `third_party_transfers.go` was never registered in
   flake.nix's fileset union (the documented AGENTS gotcha, recurring). One
   line fixed it; full audit confirmed it was the only missing file.
3. **Coverage-floor breach root-caused + rescued (91.4% → 93.5%)** — the flake
   gate failed at `coverage 91.4% is below the 92% floor`. Root cause: Go's
   per-package instrumentation means `internal/raw`'s own test binary measured
   0% (its JSON tests exercise no functions) and the two NEW brand
   `Name()` methods (AccountDetails, TransferPayment) were untested. Fixed
   with white-box tests in `internal/raw/types_test.go` (all four
   `webhookOccurredAt` carriers + generic `WebhookOccurredAt` +
   `BalanceAmount.Cents`) and `internal_test.go` (brand rows,
   `TestOTTChannelChallengeType`, `TestParseOTTAlternatives`,
   originator wire trio, `TestMapAccountIdentification`). A concurrent session
   fixed the same gap from the Ginkgo side (`1271562`) — the two layers are
   complementary, no duplication.
4. **A real docs bug fixed in ott.go** — `parseOTTAlternatives`' doc comment
   claimed malformed input "decodes to no alternatives"; the code returns an
   `errOTTAlternativesShape` error that surfaces as corruption. Comment
   corrected; the new test pins the actual behavior.
5. **M2 contributions** — `wise.Version = "0.14.0"` (types.go:21) and the
   README status line bumped (the parallel session authored the CHANGELOG
   `[0.14.0]` cut and `docs/releases/v0.14.0-release-notes.md`; both landed
   with my bumps inside).
6. **M3 verified end-to-end** — annotated signed tag `v0.14.0` at `1271562`,
   pushed; GitHub Release published 22:14 CEST; proxy `.info` confirms the
   tag→commit pin; clean-dir `go get github.com/larsartmann/wise-go@v0.14.0`
   exit 0. Tagged tree content-verified (`const Version = "0.14.0"`, README
   v0.14.0).
7. **M4 non-gated parts (`13bad72`)** — pin-sync network job added to ci.yml
   (mirrors the nix job; actionlint exit 0); the three `continue-on-error`
   sites (erraudit, cachix, art-dupl) dispositioned as token-gated warns and
   routed into the consolidated user ask. `gh workflow enable ci` remains the
   user's one-command step.
8. **M5 benchstat baseline (`7b211fc`)** — six `-count 1` captures at the
   v0.14.0 tag in an isolated worktree (7 benchmarks × 6 = 42 result lines);
   committed as `docs/bench/2026-10-16_v0140_baseline.txt` + table row;
   benchstat vs the v0.12.0 baseline: no regression above noise (geomean
   −1.88%), the four new `ParseWiseTimestamp` shape subcases enter the set.
9. **M6 + M7: the kept-report tail is CLOSED** — all eight 2026-10-07 status
   reports (03-45, 04-20, 04-57, 05-14, 06-07, 06-34, 07-09, 07-51) now carry
   item-by-item verdicts (done/wont/routed with evidence, dry-run-first,
   `check-rows` gate green) and are `git mv`-archived with manifest lines
   (`6fa82c5`, `f6a6777`). `docs/status/` now holds only point-in-time
   2026-09-16/10-11/10-16 snapshots.
10. **M13 markdownlint (`ee25ca3` via daemon)** — repo-wide pass is CLEAN
    (0 findings). Config tuned to repo conventions (MD013/MD028/MD049/MD060
    off with rationale: dprint owns tables/long lines, frozen history keeps
    mixed emphasis; tabs legal in code fences; MD024 siblings-only);
    `.markdownlintignore` extended (superset of the parallel session's
    version — no loss). Real fixes: two merged-cell rows in the
    transfer-category re-audit (data was rendering into wrong columns), an
    unescaped pipe inside inline code breaking a plan-table column, four
    language-less CONTRIBUTING fences → `text`.
11. **M26 (partial) HARVEST (`2016e5f`)** — plan rows M5/M6/M7/M13 struck with
    commit hashes, completion checklist ticked for everything verifiably
    shipped; the parallel session had already struck M4/M5 (crediting this
    session's commits by hash).
12. **Worktree hygiene** — the apidiff and bench worktrees were created for
    isolation and removed after use; no stray checkouts left.

## b) PARTIALLY DONE

1. **M4 CI resurrection** — jobs and dispositions are committed, but `gh
   workflow enable ci`, the first green run, the badge unfreeze, and the
   art-dupl baseline confirmation (now 22 clone groups after tonight's new
   code, up from the 16-group note) all remain — the enable is user-gated.
2. **M26 HARVEST** — plan M-rows and the checklist are done, but the
   fine-grain T-rows for M6/M7 (T23–T30, T56–T62) are not individually struck,
   and TODO_LIST.md was DIRTY with a parallel session's edits all session, so
   this report's (f) section is NOT yet harvested into it (the standing rule:
   never touch another session's in-flight file).
3. **M15 BuildFlow filings** — evidence exists in AGENTS (the "9 tools
   unavailable" persistence; findings-counter counting rendered lines), but
   nothing was filed this session; the local art-dupl verification (T18) was
   skipped because `go install` is sandbox-blocked in this harness.
4. **apidiff vs v0.14.0** — ran clean in the isolated worktree and now
   correctly reports "Base v0.14.0, suggest v0.14.1" (master is ahead by
   test/doc commits) — but the delta set was not itemized into a review.
5. **Shuffle coverage** — the race suite ran green after every test batch, but
   the three-pass `-shuffle=on` loop was not re-run after the new tests
   landed (the Ginkgo `-count>1` ban makes it a separate invocation set).

## c) NOT STARTED (this session's lane never touched them)

1. M9 linter-allowlist gate + negative test (flake lane).
2. M10 spec-drift hardening (extraction-floor guard, PYTHONDONTWRITEBYTECODE,
   16 added 2027Q1 paths → FEATURES triage).
3. M11 negative-scenario tests for pin_sync.py + spec_drift.py.
4. M12 coverage-floor single-sourcing (flake.nix ↔ ci.yml).
5. M16 docs-truth tail (DOMAIN_LANGUAGE terms, api-docs-study counts,
   19-59 architecture-review annotation).
6. M17 crush-config lessons (pipe-masking, ps-before-gates) — needs a commit
   in the crush-config repo.
7. M18 live-smoke harness decision + M20 decision memos (4 memos).
8. M19 consolidated user ask — ingredients gathered (tokens, CI enable,
   sandbox key, darwin matrix, RetryAfterDelay lock, GOEXPERIMENT fleet
   scope), not yet written as one document.
9. M21 bank-sync migration to v0.14.0 (other repo, gated on M3 which is done).
10. M22 sandbox-key lane; M23 Go 1.27 condition watch; M24 AGENTS prune
    (now ~50 KB vs the <30 KB target); M25 baseline consumption (PR-coupled);
    M27 v1.0.0 path.

## d) TOTALLY FUCKED UP (this session's own failures, honestly)

1. **The worktree isolation was defeated by my own flake URL — twice.** First
   attempt used an invalid `path.#app` syntax; second attempt pointed
   `nix run` at the MAIN TREE while a parallel session was mid-edit on
   flake.nix, so nix parsed their half-landed dirty copy and reported a bogus
   syntax error, which I briefly treated as a HEAD-state fact. The worktree
   only isolated the checkout, not the flake being evaluated. Fixed by
   pointing the URL at `/tmp/wise-apidiff-wt#apidiff` — and apidiff then
   passed immediately. Lesson: the isolation boundary is the flake URL, not
   the cwd.
2. **Shipped a broken test file on first write** — `if got :=
   BalanceAmount{...}.Cents(); ...` is a compile error (composite literal in
   an if-init needs parens). The gate caught it one run later; a known Go
   gotcha I should have applied at write time.
3. **Misread an edit refusal as success.** The first multiedit on the plan doc
   was REFUSED ("modified since last read" — parallel edits landed between my
   read and write) and I proceeded as if it had applied; only the verification
   grep exposed it. Re-read → re-applied cleanly. Reading a tool refusal as a
   no-op is exactly the daemon-era discipline AGENTS warns about.
4. **Overwrote a pre-existing file I believed was new.** `.markdownlint.json`
   and `.markdownlintignore` had been created by a parallel session minutes
   earlier; my write treated them as fresh creations. The write-tool's
   modified-since-read guard caught the config, and the final ignore list
   happened to be a superset (no loss) — but only by luck, not by checking.
5. **Coverage-profile parsing by hand-rolled regex failed twice** (summed
   hit-counts as statements → a nonsense 1703%) before I read the raw format
   (`count` is hits, not boolean). `go tool cover -func` was the right tool
   from the start.
6. **My explicit commit lost the race with the daemon** — `git commit` failed
   because the daemon swept my staged M13 files into `ee25ca3` first. Content
   verified intact afterward, but the "verify a daemon commit contains what
   you wrote" step was mandatory cleanup for a race I could have avoided by
   committing faster after staging.
7. **A wasted background slot on a sandbox-blocked command** — `go install
   art-dupl` is not permitted in this harness; T18 (local art-dupl run) was
   silently dropped instead of re-planned. The CI job owns that verification
   once enabled.

## e) WHAT WE SHOULD IMPROVE

1. **Isolation boundary = the flake URL.** Any nix app run intended to see a
   frozen tree must pass the WORKTREE's flake path (`/tmp/wt#app`), never the
   main repo path — otherwise dirty-tree races leak into verdicts.
2. **Never treat a tool refusal as success.** Every edit/write refusal means
   the mutation did not happen; verify with a grep before moving on (this
   session caught it only because a later check happened to look).
3. **Check `git log` for files that "don't exist"** before writing them —
   under a multi-session fleet, `ls` absence is not creation-history absence.
4. **Use `go tool cover -func` for coverage verdicts**; hand parsing the
   profile wastes cycles and produces confident nonsense.
5. **The full battery ordering paid for itself**: `nix flake check` (the
   expensive gate) is what caught the fileset omission and the coverage
   breach that plain `go test` hid. Keep running the heavy gates FIRST on
   release-verify lanes, not last.
6. **Per-package coverage semantics should be checked before writing tests** —
   internal packages need in-package tests for their own coverage number;
   root-package tests calling into them count for nothing.
7. **Cross-session credit works** — the parallel session's plan annotations
   cite this session's commit hashes ("done at 13bad72"). Keep that
   convention; it makes HARVEST reconciliation mechanical.
8. **Commit immediately after staging** under an active daemon; staging then
   composing a long message is a race window.

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

_Brainstorm, impact-ordered; not a commitment list. Items 1–17 are the 21:45
plan's own remainder; 18–50 are session fallout._

| #  | Task                                                                                                                                                                      | Source           |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------- |
| 1  | Write the consolidated user ask (CI enable, WISE_SANDBOX_API_KEY, CACHIX/ERRAUDIT tokens, darwin matrix, RetryAfterDelay lock, GOEXPERIMENT fleet scope, v1.0 timing)     | M19              |
| 2  | `gh workflow enable ci` (on your word) → watch first green → badge unfreeze → confirm art-dupl job output                                                                 | M4               |
| 3  | Confirm the grown art-dupl baseline (16 → 22 clone groups after tonight's new code) is fully accepted before CI locks it                                                  | fallout          |
| 4  | M9: linter-allowlist check (`.golangci.yml` vs pinned list) + negative test                                                                                               | plan             |
| 5  | M10: spec-drift hardening (floor guard, `PYTHONDONTWRITEBYTECODE=1`, 16 added 2027Q1 paths → FEATURES triage)                                                             | plan             |
| 6  | M11: negative fixtures for pin_sync.py + spec_drift.py wired into flake checks                                                                                            | plan             |
| 7  | M12: single-source the coverage floor (flake.nix ↔ ci.yml)                                                                                                                | plan             |
| 8  | M15: file the two BuildFlow display defects (9-tools-unavailable persistence; findings-counter counting rendered lines) with repros                                       | plan             |
| 9  | M21: bank-sync → v0.14.0 (typed `OneTimeToken` call sites), scratch-verify, redeploy, confirm FX footer                                                                   | plan             |
| 10 | M22 (on key): 5-assumption sandbox pass + OTT `2026Q3→2026Q4` probe + flip + first recorded sandbox-live run                                                              | plan             |
| 11 | M23: Go 1.27 condition re-probe (`nix eval nixpkgs#go.version`) and execute the migration plan when its 4 conditions hold                                                 | plan             |
| 12 | M24: AGENTS.md size pass (≈50 KB → <30 KB; extract incident narratives, keep every live gotcha)                                                                           | plan             |
| 13 | M17: commit the pipe-masking + ps-before-gates lessons to crush-config `references/lessons.md`                                                                            | plan             |
| 14 | M16: docs-truth tail (DOMAIN_LANGUAGE terms vs code, api-docs-study counts vs the 188-path spec, 19-59 review annotation)                                                 | plan             |
| 15 | M18: live-smoke harness decision (build-tag `_live_test.go` vs sandbox-live fold) + implement the winner                                                                  | plan             |
| 16 | M20: four decision memos (Health code-half, spec-drift posture, flake tag-pins, darwin matrix)                                                                            | plan             |
| 17 | M27: v1.0.0 path (growth lineage at 46 methods; tag on your word)                                                                                                         | plan             |
| 18 | HARVEST this report's (f) into TODO_LIST.md once the parallel session's TODO edits land                                                                                   | skill            |
| 19 | Strike the plan's fine-grain T-rows (T23–T30, T56–T62) for the archived batches                                                                                           | fallout          |
| 20 | Itemize the current apidiff v0.14.0→HEAD delta (test/doc commits only expected)                                                                                           | fallout          |
| 21 | Verify the tag message's "32 godoc examples" claim against the actual `Example*` count (doc-verify gates FEATURES, not the tag)                                           | fallout          |
| 22 | Verify the GitHub release page renders correctly (code spans, tables) via `gh release view`                                                                               | fallout          |
| 23 | pkg.go.dev render check for v0.14.0 (lags the proxy by up to ~1h; non-signal 404 on /fetch)                                                                               | fallout          |
| 24 | Re-run the three-pass `-shuffle=on` loop after tonight's test additions                                                                                                   | fallout          |
| 25 | Re-run `nix flake check` once the tree settles (fmt/build/doc-verify/markdownlint are green; the flake gate predates M13's commits)                                       | fallout          |
| 26 | Confirm `checks.pin-sync` still green after the parallel session's flake.nix anchor-check wiring                                                                          | fallout          |
| 27 | Confirm anchor-check's negative test (T35) exists and red-lines on a wrong anchor                                                                                         | M8 tail          |
| 28 | Confirm anchor-check covers AGENTS.md anchors or document its scope                                                                                                       | M8 tail          |
| 29 | LSP tripwire: confirm the gopls duplicate-decl phantoms (visible ALL session in editor diagnostics) are gone in a FRESH session; if not, the `.crushrc` env fix regressed | reopen condition |
| 30 | codespell + gitleaks re-run after tonight's doc churn (CONTRIBUTING/README changed; configs skip status/reviews trees)                                                    | fallout          |
| 31 | Keep `internal/raw` at 100% function coverage — new raw payload types need carrier tests in the same commit                                                               | fallout          |
| 32 | Next baseline at v0.15.0 per the bench README convention                                                                                                                  | cadence          |
| 33 | Decide whether markdownlint joins the pre-release chain or stays on-demand (2026-10-11 decision says on-demand)                                                           | decision         |
| 34 | CONTRIBUTING: one line documenting the on-demand markdownlint invocation                                                                                                  | fallout          |
| 35 | Dependabot actions-group re-check after ci.yml changes (pin-sync job uses pinned SHAs — expected clean)                                                                   | fallout          |
| 36 | Record the status-report `.md`-override recurrence (skill's standing exception covers dispatch reports only)                                                              | process          |
| 37 | Fleet pin-drift sweep (ROADMAP raw idea) — tonight proved the class again via the flake fileset miss                                                                      | ROADMAP          |
| 38 | Daemon-hygiene upstream ask (heuristic commits sweeping staged work mid-commit) already on ROADMAP — keep evidence accumulating                                           | ROADMAP          |
| 39 | Multi-session claim convention proposal (see g3)                                                                                                                          | process          |
| 40 | Coordinate with the 22:29 "error-system excellence" plan from the other session — dedupe before executing                                                                 | fleet            |
| 41 | Verify no `.envrc`/direnv regression in fresh shells (GOEXPERIMENT injection)                                                                                             | standing         |
| 42 | `release-notes-check` on any future release-notes edit before `gh release edit`                                                                                           | ritual           |
| 43 | Consider a `docs/releases/` template with the count-claim slot pre-positioned                                                                                             | M19-adjacent     |
| 44 | Consider extracting the M1 gate battery into a scripted release-inventory lane (it caught two real breakages tonight)                                                     | improvement      |
| 45 | Keep `docs/status/` holding only point-in-time snapshots (achieved tonight — maintain on each sweep)                                                                      | cadence          |
| 46 | Re-check `README.md` install snippet for a version pin that should say v0.14.0                                                                                            | fallout          |
| 47 | Confirm the coverage-badge job's floor constant still matches 92 after M12 (or single-source it first — item 7)                                                           | M12-adjacent     |
| 48 | When #34-adjacent BuildFlow fixes land, re-diff embedded vs standalone before lifting any skip (none currently skipped)                                                   | upstream watch   |
| 49 | Watch md-go-validator #8 for the flake-input switch that deletes the go_1_27/vendorHash dance                                                                             | upstream watch   |
| 50 | Next docs-health pass: annotate THIS report inline when its (f) items ship (annotate-then-archive)                                                                        | skill            |

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF (3)

1. **CI enable, now?** Everything short of `gh workflow enable ci` is
   committed and linted (pin-sync job included). Say the word and I enable +
   babysit the first green run. Related: after the first green, do you want
   art-dupl promoted to blocking after its 22-group baseline is confirmed, or
   keep the documented warn-first-for-two-weeks policy?
2. **The art-dupl baseline grew 16 → 22 clone groups** with tonight's
   originator/webhook/OTT code (per the parallel session's baseline-note
   commit). Should I verify those six new groups are all
   `// art-dupl:accept`-suppressed BEFORE CI's first run locks the baseline,
   or do you accept the note's claim as-is?
3. **Multi-session orchestration:** three sessions executed the same plan
   tonight with real collisions (a half-edited flake.nix parsed by my nix
   run, duplicate parallel coverage fixes, TODO_LIST contention, a daemon
   sweep of my staged commit). Do you want a lightweight file-claim
   convention (e.g. an OWNERSHIP.md or a TODO_LIST owner column that
   sessions check and set), or is watch-and-dodge acceptable as the standing
   protocol?

---

_Point-in-time snapshot, 2026-10-16 22:44 CEST. Feed section (f) into
`docs-health` HARVEST; annotate (don't rewrite) when bringing this report
current. HEAD at write time: `2016e5f`, master pushed, all local gates green._
