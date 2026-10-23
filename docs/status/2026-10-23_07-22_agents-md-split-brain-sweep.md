# Status Report — AGENTS.md Split-Brain Sweep (2026-10-23 07:22)

Session scope: user report "AGENTS.md is fucked up, fix all split-brain issues". One session, commit `e42cafe` (auto-git daemon, contents verified). This report covers THAT session only — no fresh project audit.

---

## Brutal Self-Review (the three questions asked)

### What did you forget?

1. **Full `nix flake check` was never run after the edits.** I ran anchor-check, doc-verify, and `nix fmt` — the gates directly touching the two edited markdown files — but announced them under the umbrella word "Gates", which oversold completeness. links-check, md-go-snippets, script-fixtures, linter-allowlist never ran. Probability they'd fail on a markdown-only change: low. But "low" was assumed, not verified.
2. **I introduced new micro-duplications while deduplicating.** My erraudit rewrite states "currently dormant while CI is disabled" — the CI-disabled fact now lives in TWO entries (mine + the canonical "CI workflow is disabled on GitHub" entry). Same for "skip lifted" (still two sites). I fixed 8 split brains and minted 2 new small ones.
3. **Cross-references I added are paraphrases, not exact titles.** "the Conventions quarterly versioned surface entry", "the BuildFlow-defects gotcha", "the dedicated gotcha above" — none are the literal bold titles, and no gate verifies prose quotes. A future title rename breaks these pointers silently. The exact irony: the split-brain fixer shipped unguarded references.
4. **Cross-file sweep not done.** I fixed AGENTS-internal split brains plus two mechanically-anchored pointers (CONTRIBUTING, ADR 001, ci.yml). The same 8 consolidated facts may have stale copies in README / FEATURES.md / TODO_LIST.md / ROADMAP.md — never checked.
5. **The merged Money entry was preserved, not re-verified against code.** Both source entries could have been jointly stale; I verified structure and anchors, not `types.go` truth (`Money` field shape, `NewCurrency` rules, `toMoney` location).
6. **HARVEST loop intentionally left open** — this report's section (f) must feed TODO_LIST/ROADMAP, but user ordered "wait for instructions", so no harvest yet.

### What could I have done better?

- Run `nix flake check` as the final gate, or explicitly say it was skipped and why.
- Use exact entry titles in every cross-ref I authored (5 sites).
- Before editing, grep the 8 target facts across ALL living docs, not just AGENTS.md — docs-health's cross-file consistency leg, applied retroactively.
- The blank-line normalization was done blind by pattern; a formatter-level rule would have been the durable fix.

### What could I still improve?

- The root cause is unaddressed: AGENTS.md accretes ~60 mega-bullet entries with no write-time one-fact-one-place discipline and no structural index. Today's 8 dupes will regrow. Systemic options in section (e)/(f).
- A quoted-title checker (anchor_check extension) would make prose pointers gate-verified instead of trust-me.

### Did I lie / oversell?

One presentation-level oversell: "Gates: anchor-check, doc-verify, nix fmt" implied complete gating (see forget-#1). Every individual claim made was mechanically true and evidence-backed; the framing was wider than the verification.

---

## a) FULLY DONE

| What                                                                                                                                                                                                                         | Evidence                                                                                                   |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| erraudit contradiction fixed ("run via CI only" vs CI-disabled; verified CI is disabled via `gh workflow list`)                                                                                                              | commit `e42cafe`, entry rewritten to two real paths (dormant ci.yml job + local BuildFlow-embedded engine) |
| 7 duplicated facts consolidated to one canonical home + cross-refs: erraudit gate command, quarterly body-shape probing, retry DelayFunc mechanism, md-go-validator #8 story, auto-configure skip, anonymous `%w` wrap idiom | commit `e42cafe`; mechanical greps: 1 occurrence each                                                      |
| Two Money contract entries merged under the ADR-001-anchored title, all concepts from both carried over                                                                                                                      | commit `e42cafe`; concept-by-concept checklist in session                                                  |
| CONTRIBUTING.md stale title quote aligned ("LSP phantoms were an env problem" → current title)                                                                                                                               | commit `e42cafe`                                                                                           |
| 8 stray blank lines normalizing the Gotchas/Conventions lists                                                                                                                                                                | commit `e42cafe`; only heading-adjacent blanks remain                                                      |
| Externally-anchored entry titles preserved (ci.yml ×2, ADR 001 ×1, CONTRIBUTING ×1)                                                                                                                                          | verified pre-edit, intact post-edit                                                                        |
| False-positive split brains ruled out with evidence: ADR 003 dual citation (ADR text confirms dependency-light rationale), 43-extractor-endpoints vs 46-client-methods (different measures)                                  | ADR 003 read; both entries reconciled                                                                      |
| Verification: anchor-check 101/101, doc-verify all passed, `nix fmt` 0 changed, daemon commit `e42cafe` diff-inspected                                                                                                       | all output captured in session                                                                             |

## b) PARTIALLY DONE

| Item                                 | Done                                                 | Missing                                                                                                                    | Effort |
| ------------------------------------ | ---------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- | ------ |
| Split-brain elimination in AGENTS.md | 8 duplications + 1 contradiction fixed, verified     | 2 self-introduced micro-dupes remain ("CI is disabled" restated; "skip lifted" ×2); 5 paraphrased pointer titles unguarded | S      |
| Doc-consistency verification         | Direct gates for edited files green                  | Full `nix flake check` not run; merged Money entry not re-verified against `types.go`/`helpers.go`                         | S      |
| Cross-file split-brain coverage      | AGENTS-internal + 3 anchored external pointers fixed | README/FEATURES/TODO_LIST/ROADMAP never swept for stale copies of the 8 facts                                              | S-M    |

## c) NOT STARTED

- HARVEST of this report's section (f) into TODO_LIST.md / ROADMAP.md (blocked on user review — "wait for instructions").
- Systemic AGENTS.md restructure (index, section taxonomy, entry-size discipline) — needs owner decision (question g-2).
- Quoted-title gate in `scripts/anchor_check.py` — idea born from this session's gap, no code.
- Any README/FEATURES dedupe work (nothing scanned yet).

## d) TOTALLY FUCKED UP

Nothing user-facing broke; no gates red. Honest defects self-inflicted this session, all Low severity:

1. **New micro-split-brains introduced** — "CI is disabled" now stated in two entries; "skip lifted" still two sites. Severity: low (no numbers/commands involved). Root cause: rewriting for correctness without re-applying the one-fact-one-place test to my own output. Mitigation: S-effort pointer swap (f #1).
2. **Unguarded paraphrased pointers** — 5 cross-refs quote non-literal titles; silent breakage on rename. Root cause: no gate exists for prose quotes. Mitigation: exact-title pass now (f #2), gate later (f #3).
3. **Verification framing oversold** — "Gates:" headline without `nix flake check`. Root cause: equating "gates for the touched surface" with "gates". Mitigation: doc-change gate checklist (f #25).

## e) WHAT WE SHOULD IMPROVE

1. **Write-time discipline for AGENTS.md** — every append should grep for the fact first (one-fact-one-place). Today's dupes all entered as appends that didn't check. Impact: high — this is the recurrence mechanism. Fix: a one-line rule in AGENTS.md itself + habit.
2. **Quoted-title checking** — extend anchor_check.py to verify `"quoted entry title"` strings in living docs against actual bold titles. Turns my paraphrase defect class into a gate failure.
3. **Doc-change gate checklist** — living-doc edits get anchor-check + doc-verify + `nix flake check`, explicitly listed in Build & Dev.
4. **Entry titles as stable IDs** — externally-anchored titles (ci.yml, ADRs) already behave as IDs; formalize: titles don't churn casually, or update anchoring files in the same commit (the AGENTS.md gotcha about `.golangci.yml` + allowlist is the same pattern).
5. **AGENTS.md structure** — 135 lines × ~60 mega-bullets is at the edge of what a session reliably reads; duplicates HID in that density. Sections/index or per-area files would make both reading and dedup easier.

## f) Next tasks (session-derived, up to 50; #1-10 are this session's own debt)

| #  | Task                                                                                                                              | Impact | Effort | Category       |
| -- | --------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | -------------- |
| 1  | Swap my 2 micro-dupes for cross-refs ("CI is disabled", "skip lifted")                                                            | Medium | S      | Quality        |
| 2  | Replace 5 paraphrased pointer titles with exact entry titles                                                                      | Medium | S      | Quality        |
| 3  | Extend anchor_check.py: quoted-title verification for living docs                                                                 | High   | M      | Quality        |
| 4  | Grep the 8 consolidated facts across README/FEATURES/TODO_LIST/ROADMAP/CONTRIBUTING; fix stale copies                             | High   | S      | Documentation  |
| 5  | Re-verify merged Money entry against `types.go`/`helpers.go` (shape, `NewCurrency`, `toMoney`)                                    | Medium | S      | Quality        |
| 6  | Run full `nix flake check` on current master                                                                                      | Medium | S      | Quality        |
| 7  | Compress the 2026-10-11→2026-10-17 "supersedes" chain to the final ruling (history → link)                                        | Low    | S      | Documentation  |
| 8  | Same compression pass for other resolved-narrative entries (auto-configure history, BuildFlow #34)                                | Low    | M      | Documentation  |
| 9  | Add write-time one-fact-one-place rule to AGENTS.md Conventions                                                                   | High   | S      | Documentation  |
| 10 | Add doc-change gate checklist (anchor-check + doc-verify + flake check) to Build & Dev                                            | Medium | S      | Documentation  |
| 11 | AGENTS.md section taxonomy: group the 6 error entries, split Gotchas into Wire/Errors/Build/Linters/Process                       | Medium | M      | Documentation  |
| 12 | Mini-TOC or index at top of AGENTS.md                                                                                             | Low    | S      | Documentation  |
| 13 | Audit Gotchas vs Conventions placement (e.g. Accept-Minor-Version is a convention; "Historical reports" is process)               | Low    | M      | Documentation  |
| 14 | Compress endpoint-count timeline (33→37→41→46) to current count + history link                                                    | Low    | S      | Documentation  |
| 15 | Decide + implement AGENTS.md structural split (single file vs per-area files via context_paths)                                   | High   | M-L    | Decision       |
| 16 | `gh workflow enable ci` + first-run watch (existing plan doc `2026-10-17_ci-first-run-watch-package.md`)                          | High   | M      | Infrastructure |
| 17 | Decide erraudit job posture on CI re-enable (blocking vs warn-not-fail; needs ERRAUDIT_TOKEN secret)                              | Medium | S      | Decision       |
| 18 | Unfreeze coverage-badge job after CI re-enable; verify 93.7% local vs 92.0 floor                                                  | Low    | S      | Infrastructure |
| 19 | `nix flake update nixpkgs` to reach go 1.26.9 and clear the 63 dispositioned vuln findings (govulncheck 43 + vulnix 20)           | High   | M      | Security       |
| 20 | Drop x86_64-darwin system per decision memo D5 (one-liner, reversible)                                                            | Low    | S      | Cleanup        |
| 21 | Spec-refresh item: canonicalize OTT to 2027Q1 (not piecemeal) + refresh vendored snapshot                                         | Medium | M      | Maintenance    |
| 22 | `--fail-on-drift` posture for spec-drift once snapshot refreshed                                                                  | Low    | S      | Maintenance    |
| 23 | Verify the 6 AGENTS claims citing `2026-10-16_agents-incident-narratives.md` still match that doc                                 | Low    | S      | Quality        |
| 24 | Sweep living docs for other stale "CI is disabled"-class state claims (one grep, fix all)                                         | Medium | S      | Documentation  |
| 25 | Check Build & Dev command list vs CONTRIBUTING for duplicated command documentation                                               | Low    | S      | Documentation  |
| 26 | Markdown lint for tight-list consistency (dprint doesn't enforce; blank-line drift recurs)                                        | Low    | M      | Quality        |
| 27 | README "Observability → OpenTelemetry" section vs AGENTS OTEL entry consistency check                                             | Low    | S      | Documentation  |
| 28 | bank-sync consumer: when it bumps past v0.14.0, sweep its error handling against the recipe (standing instruction, AGENTS line 5) | Medium | M      | External       |
| 29 | Wave-2 error-code sentinels (conscious `TestCodeSentinelsMapCuratedCodes` pin update) — ROADMAP fuel                              | Low    | M      | Feature        |
| 30 | `WithMetrics` second OTEL hook — hold until consumer need — ROADMAP fuel                                                          | Low    | M      | Feature        |
| 31 | Historical doc check: incident-narratives' "two runtime deps" claim vs actual three in go.mod (annotate, don't rewrite)           | Low    | S      | Documentation  |
| 32 | Quarterly re-verify of external title anchors (ci.yml ×2, ADR 001, CONTRIBUTING) — fold into anchor-check scope if #3 lands       | Medium | M      | Quality        |
| 33 | Add CONTRIBUTING quoted titles to whatever gate #3 becomes                                                                        | Low    | S      | Quality        |
| 34 | Consider an AGENTS.md freshness pass convention: every N weeks, VERIFY-mode sweep (docs-health)                                   | Medium | S      | Process        |
| 35 | md-go-validator #11 (nested-fence extraction) — upstream, watch                                                                   | Low    | S      | External       |
| 36 | BuildFlow #65 ("9 tools unavailable" header) — upstream, watch                                                                    | Low    | S      | External       |
| 37 | BuildFlow #35 (`nix run .#reinstall` no-op) — manual profile swap remains the workaround                                          | Low    | S      | External       |
| 38 | Run on-demand `buildflow -s gitleaks -s codespell` after this doc-heavy session (per AGENTS infra rule)                           | Low    | S      | Quality        |
| 39 | Confirm daemon commit hygiene: no temp files swept into `e42cafe` (spot-check full file list)                                     | Low    | S      | Quality        |
| 40 | Self-check: re-run the 10-item todo list of this session against the worktree as a post-hoc audit                                 | Low    | S      | Process        |

(40 items — everything else I noticed is already gate-covered or upstream-tracked.)

## g) Questions I cannot answer myself

1. **AGENTS.md end-state:** keep it one file with taxonomy/index, or split into per-area memory files loaded via `context_paths`? This is your memory-architecture call — both work, but the recurrence of split brains is a direct symptom of the append-to-one-file pattern.
2. **Resolved-ruling narrative:** should superseded rulings (the 2026-10-11 vs 2026-10-17 chain, auto-configure history) stay as inline narrative for context, or compress to final-state + history link? I chose not to touch them without your preference — they're working, just heavy.
3. **erraudit CI posture:** when CI is re-enabled, should the erraudit job become blocking (ci.yml comments say: delete the continue-on-error lines once ERRAUDIT_TOKEN exists) or stay warn-not-fail? Decides whether I fold "gate is covered by CI" language back into the AGENTS entry.

---

_Point-in-time snapshot. Section (f) awaits HARVEST into TODO_LIST/ROADMAP on user instruction (held per "wait for instructions"). Format: `.md` at explicit user path (overrides this skill's HTML default)._
