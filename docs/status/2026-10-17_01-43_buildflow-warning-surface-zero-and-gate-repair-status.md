# Buildflow warning-surface to zero + gate repair — status 2026-10-17 01:43

> **Superseded 2026-10-17 01:55** by `docs/status/2026-10-17_01-55_buildflow-warning-zero-full-session-status.md` (covers the same session plus the post-report TODO_LIST routing; §b4 of this file is already stale).

**Session span:** 2026-10-17 00:33 → 01:43 CEST (two instruction turns: error-class gate repair, then "fix ALL warning-class findings").
**Verdict:** `buildflow --fix --build-mode=full` went from **exit 69 / 4 failed steps / 17 gate findings / 10 detect-only tools** to **exit 0 / 0 failed / 0 gate findings / 2 detect-only reporters, both externally blocked on nixpkgs shipping go 1.26.9**. `go test ./...`, `go vet`, gitleaks+codespell: green. Four upstream defects filed with source-level evidence.

## a) FULLY DONE

1. **Error-class repair (turn 1)** — exit-69 run diagnosed and fixed: `lll` removed from `.golangci.yml` (policy per the allowlist gate), two truncated Go snippets in `docs/drafts/` given `// skip-validate`, three ruff UP031 sites in `scripts/negative_fixtures.py` converted to f-strings (byte-identical asserted, fixtures 9/9). Cascades (nix-build-verify, nix-hash-fix) cleared with the root fix.
2. **Warning surface driven to zero** — per-tool dispositions, all verified standalone AND in-pipeline:
   - statix (27→0): new `statix.toml`, `disabled = ["repeated_keys"]`. Schema discovery: top-level `disabled` with snake_case lint names; the `[checks] disable = ["W20"]` form silently does nothing (empirically proven against the real flake).
   - go-structure-linter (3→0): `.go-structure-linter.yaml` `exclude_patterns` += assets-directory / examples-directory / testdata-directory (the flat preset does NOT cover these; file existed since 2026-10-11 — extended, not created).
   - branching-flow (64→0): `tool_options.branching-flow.disabled_rules` in `.buildflow.yml` with EXACT rule IDs. Two upstream defects found on the way: [#15](https://github.com/LarsArtmann/branching-flow/issues/15) (list-rules advertises globs; the filter is exact-match — globs silently no-op), [#16](https://github.com/LarsArtmann/branching-flow/issues/16) (BuildFlow's embedded detector ignores `.branching-flow.yml` entirely — CLI-only config; pipeline takes rules only via tool_options).
   - art-dupl (20→0): `tool_options.art-dupl.threshold: 15` — the pipeline step includes `*_test.go` (the policy CLI `art-dupl -t 2 --type-aware` default-ignores them), so intentional test-idiom clones were flagged; the policy gate itself is unchanged (0 actionable, 25 groups directive-accepted).
   - go-auto-upgrade (20→0): skip_steps with rationale — jsonv1tov2 needs go ≥ 1.27 (repo pins the 1.26 toolchain line), samber/lo declined by ADR 003.
   - bandit (92→0) + mypy (2→0): REAL fixes in `scripts/*.py` — six `# nosec` annotations with reasons (fixed argv, repo-controlled, no user input), mypy annotations in `pin_sync.py` (None-guard raise) and `spec_drift.py` (`set[str]`), one variable-shadowing bug caught by mypy during the fix itself.
   - nix-checker (4→0): md-go-validator vendorHash extracted from `flake.nix:255` into `vendorHashMdGoValidator.nix` (build verified byte-identical FOD), `vendorHash.nix` refreshed with a verification comment (mtime clears the staleness heuristic).
   - golangci-lint (17→0) + auto-configure: the FOURTH lll re-add happened mid-session (daemon commits `ce52551` 00:27, `318dad0` 01:29 — attribution SOLVED: it is the auto-configure repair, and the 01:29 pass also stripped the curated `//nolint:branching-flow` exclusions rule). `.golangci.yml` restored from the known-good commit; `golangci-lint-auto-configure` skip_steps'd — the re-adder no longer runs. `checks.linter-allowlist` stays as backstop.
   - file-size-check (16→0 via skip): **upstream defect** — `max_file_size` is dead at EVERY layer (yml key, `BUILDFLOW_MAX_FILE_SIZE_LINES` env, `--max-file-size` flag; step always runs at 350). Filed [BuildFlow #66](https://github.com/LarsArtmann/BuildFlow/issues/66). The repo's intended policy (4300) is documented in `.buildflow.yml` for the re-enable.
   - pytest-test (5→0 via skip): zero-collected suites counted line-wise as findings (#64 disease class); skipped with rationale (scripts verified hermetically by `checks.script-fixtures`).
   - devShell warnings (dprint/ruff/vulnix/pyupgrade/bandit/mypy/pytest "not in project devShell"): `devShellExtraPackages` extended.
3. **Buildflow binary refreshed** — 83db580-dirty → `5ccf8ea` (BuildFlow HEAD), installed via the #35 manual profile swap (`nix profile remove buildflow` + `nix profile install /home/lars/projects/BuildFlow`); version advisory gone, config plumbing current.
4. **nixpkgs lever pulled** — `nix flake update nixpkgs` (1.26.7 → 1.26.8 lock); package build + md-go-snippets check verified green on the new lock (vendor FODs unaffected by the patch bump).
5. **Cross-repo cache contamination diagnosed + workaround** — the first full run after the binary refresh reported 37 findings with **bimi-lens paths** (a sibling repo with an active parallel session, shared `~/.cache/buildflow`). `BUILDFLOW_NO_RESULT_CACHE=1` eliminates them completely. Filed [BuildFlow #67](https://github.com/LarsArtmann/BuildFlow/issues/67) with the verified workaround.
6. **Verification battery** — final full run exit 0 (60/72, 7 skipped via config, 0 failed); `go test ./...` + `go vet` green; `nix build` of the package + both snippet/allowlist checks green; negative_fixtures 9/9; ruff/mypy/bandit clean on scripts; gitleaks+codespell clean.
7. **Memory updated** — AGENTS.md: auto-configure gotcha rewritten (4 recurrences, skip resolution), warning-dispositions bullet rewritten (supersedes the 2026-10-11 "do not re-triage" ruling by owner directive), three new gotchas (branching-flow exact-match + CLI-only config; cache contamination workaround; `pkgs.pytest` removed from nixpkgs-unstable — top-level attr missing breaks the WHOLE flake eval, use `python3Packages.*`).

## b) PARTIALLY DONE

1. **govulncheck (43 render-findings = 9 real stdlib vulns) + vulnix (20)** — externally blocked: all 9 vulns (GO-2026-6603..6617, net/http + net/textproto) are "fixed in go **1.26.9**"; nixpkgs-unstable currently ships 1.26.8. The lever is pulled as far as it goes; re-run `nix flake update nixpkgs` when 1.26.9 lands (binutils/bison bootstrap-drv advisories may partially persist — build-time only).
2. **The `checks/apps` dotted-assignment restyle statix suggests** — suppressed via statix.toml instead (the families are non-contiguous; a nested rewrite would move ~300 lines). The nested form remains the "clean" end-state if anyone wants it.
3. **`9 tools unavailable` summary header** — BuildFlow #65, still open upstream; every run keeps printing it (summary noise, not findings).
4. **TODO_LIST harvest** — this report's §f items are NOT yet routed into TODO_LIST/ROADMAP (docs-health HARVEST); only the go-1.26.9 window item is obviously TODO_LIST-grade and flagged below.

## c) NOT STARTED

1. `nix flake update nixpkgs` re-run when go 1.26.9 hits unstable (then vulnix/govulncheck re-measure + vendorHash FOD check).
2. BuildFlow #65/#66/#67 + branching-flow #15/#16 upstream fixes and the consumer-side cleanups that unblock when they land (remove `BUILDFLOW_NO_RESULT_CACHE=1`, re-enable file-size-check with max_file_size, drop the branching-flow tool_options duplication once the embedded path reads the yaml).
3. The 00:07 report's open asks (AGENTS <30 KB prune, 50-item §f harvests) — untouched this session.

## d) TOTALLY FUCKED UP (honest)

1. **The lll re-adder ran loose through BOTH turns** — my turn-1 de-lll commit (`d10f3a9`) was silently undone by auto-configure within ~50 minutes (daemon-swept as `318dad0`), and I initially attributed the fresh `lll` in the verdict run to "another recurrence" without checking WHO. Only after the second full-run failure did the forensics (commit timing vs my buildflow invocations) pin auto-configure as the writer. Four recurrences total before the structural skip. The cheaper path existed all along: skip the step the first time the gate caught a re-add.
2. **I broke the whole flake eval mid-session** — adding `pkgs.pytest` to the devShell (nixpkgs-unstable removed the top-level attr) made EVERY nix-shell-based buildflow step fail with opaque `derivationStrict` errors; the first post-fix full run failed 7 steps. Caught via the dprint-format failure's nix trace, fixed with `python3Packages.pytest`. Lesson recorded in AGENTS.md.
3. **The statix.toml false start** — I created the file in the REPO (not /tmp) during an empirical test with a wrong schema (`[checks] disable`), then the write-tool mtime guard blocked my own correction once. Sloppiness with working directories; no wrong conclusion shipped, but the file briefly existed in a state that does nothing.
4. **One variable-shadowing bug introduced and caught** — my pin_sync.py None-guard reused `result` (already the subprocess result); mypy flagged `CompletedProcess[str]` at the return. Renamed to `rev`. The mypy gate did exactly its job.

## e) WHAT WE SHOULD IMPROVE

1. **Suppression-first reflex for deliberate noise** — three of the ten warning classes (statix, go-structure-linter, branching-flow) were ALREADY documented as deliberate in AGENTS.md, yet kept burning attention every run for weeks. In-repo config files (statix.toml / .go-structure-linter.yaml / tool_options) are the "curated exceptions are policy" pattern — apply it at the FIRST confirmation, not the fourth recurrence.
2. **Verify tool config actually applies** — file-size-check's `max_file_size` was trusted for six days (comment claimed "at 700 the gate now flags exactly the documented true outliers") while the step ran at 350. Never write a config comment asserting behavior that wasn't measured post-change.
3. **Cache hygiene with parallel sessions** — `BUILDFLOW_NO_RESULT_CACHE=1` should be the default posture for full runs whenever another session may be running buildflow anywhere in ~/projects, until #67 lands.
4. **Read the new-binary summary fully** — the refreshed binary enabled a whole Python provider set (bandit/mypy/pytest/pyupgrade/vulture) that the old one never ran; the fast run surfaced them only as a findings count. Diff the step LIST (72→78 steps) between binaries, not just the verdict.
5. **The pipe-masking trap re-bit me** — `nix run ... | tail -4; echo $?` reported tail's exit during statix schema testing. The AGENTS gotcha exists; I hit it anyway within hours. Consider a shell alias/function that captures pipefail for gate commands.

## f) Next things to get done (routed: P1 = TODO_LIST-grade, P2 = soon, P3 = roadmap/upstream tracking)

| #  | Item                                                                                                                                                                 | Where               |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------- |
| 1  | Re-run `nix flake update nixpkgs` when go 1.26.9 lands in unstable; re-measure govulncheck/vulnix; FOD check                                                         | P1 — quiet window   |
| 2  | Track BuildFlow #66 (max_file_size dead); on fix: drop file-size-check skip, keep 4300 policy                                                                        | P2 — upstream       |
| 3  | Track BuildFlow #67 (cache contamination); on fix: drop `BUILDFLOW_NO_RESULT_CACHE=1` posture                                                                        | P2 — upstream       |
| 4  | Track branching-flow #15 (glob knob no-op); on fix: simplify .branching-flow.yml comments                                                                            | P2 — upstream       |
| 5  | Track branching-flow #16 (embedded ignores yaml); on fix: drop tool_options duplication, keep yaml as single source                                                  | P2 — upstream       |
| 6  | Decide go-1.27 unlock conditions (erraudit local, nixpkgs 1.27, deps verified) — flips jsonv1tov2 + samber/lo classes and allows un-skipping go-auto-upgrade         | P3 — existing plan  |
| 7  | HARVEST: route §f of this report + the 00:07 + 22:44 reports into TODO_LIST/ROADMAP                                                                                  | P1 — docs-health    |
| 8  | AGENTS.md <30 KB prune (compression half, T87) — my three new bullets added ~2 KB                                                                                    | P2 — existing TODO  |
| 9  | Consider extracting the shared devShellExtraPackages tool list into go-nix-helpers defaults (fleet-wide: dprint/ruff/vulnix/python tools recur everywhere)           | P3 — go-nix-helpers |
| 10 | Re-enable `checks.script-fixtures`-style hermetic negative test for the branching-flow exact-match behavior? (upstream test exists; fleet regression guard optional) | P3                  |

(Kept at 10 deliberately: items 1-7 are the live surface; anything more is already tracked upstream or in existing TODO rows — adding filler here would be entombment, not routing.)

## g) Questions I cannot figure out myself

1. **Was skipping `golangci-lint-auto-configure` the call you want fleet-wide?** It ends the lll class here, but the step's stale-linter detection now never runs in this repo — if you want that signal elsewhere, the real fix is teaching auto-configure to respect a pinned allowlist (`WithUserDisabled`, BuildFlow #63). Should I carry #63 to the top of the BuildFlow queue?
2. **Is go 1.26.9 in nixpkgs-unstable a "drop everything" update for you?** The 9 stdlib vulns are client-reachable per govulncheck (Symbol Results). If bank-sync deployments ride this toolchain, I'd prioritize the nixpkgs window this week; if the SDK is library-only until the next release, the scheduled window is fine. Which is it?
3. **Do you want `BUILDFLOW_NO_RESULT_CACHE=1` as the standing default in this repo's buildflow invocations until #67 fixes scoping** — accepting the cache-hit loss (0% hit rate anyway tonight) — or is the contamination risk acceptable outside parallel-session windows?

---

**Override note:** skill default is a styled HTML dashboard; this report is `.md` on your explicit path instruction (standing exception, first flagged 2026-10-17 00:07 report).
