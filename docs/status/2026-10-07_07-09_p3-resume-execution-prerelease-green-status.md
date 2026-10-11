# P3 Resume Execution — Pre-Release Green, Shuffle Flag Caught & Fixed, Cross-Repo Lessons

> **Shipped (2026-10-07 docs-health pass):** the pre-release chain is green
> end-to-end and `--shuffle` is verified; the cross-repo lessons and the
> `meta.description`/CI-wiring follow-ups are harvested to `TODO_LIST.md` (P3 +
> Cross-repo). Open items: the three §g user questions (x86_64-darwin,
> `RetryAfterDelay` freeze, GOEXPERIMENT scope) and the P1/P2 user-blocked set.

**Date:** 2026-10-07 07:09 CEST
**Session:** continuation of the P3 sweep resume (`docs/status/2026-10-07_06-34_p3-sweep-resume-defect-filing-lsp-rootcause-status.md`), executing its §f backlog end-to-end under the user's blanket "do the whole list" directive.
**Tree:** clean at report time; HEAD `0dc97d6`. All 13 session todo items completed. Go tests + golangci-lint (0 issues) green after every mutation; `nix run .#pre-release` green twice.

---

## a) FULLY DONE (with evidence)

1. **LSP phantom fix VERIFIED LIVE** (first action of the session) — zero diagnostics on `internal_test.go` and `webhooks.go`; the `bytes unused` (internal_test.go:4) and "not properly formatted" (webhooks.go:142) phantoms are GONE. The `.crushrc` GOEXPERIMENT env fix works at session start. Reopen condition stays in AGENTS.md.
2. **FIRST-EVER end-to-end pre-release run: ALL GATES PASSED, exit 0** — dirty-tree check → build → vet → race → golangci-lint 0 issues → `nix flake check` (9/9 checks) → doc-verify (41-method count claims on all 5 files, 25 Example funcs, 64 links OK / 33 excluded) → release-notes-check → apidiff (network). Every component of the chain verified composed, not just individually.
3. **apidiff baseline established** — vs v0.12.0: exactly ONE compatible change, `RequestLog.RetryAfterDelay: added`; gorelease suggests v0.13.0. This is hard evidence for the §g v1.0 API-lock question.
4. **`--shuffle` flag on the pre-release app: implemented, caught broken by its own verification, fixed, re-verified** — first implementation used `go test -shuffle=on -count=3`; the verification run FAILED because Ginkgo bans `-count>1` (only `-count=1` supported). Fixed to three separate `-shuffle=on` passes; re-verified end-to-end (exit 0) plus standalone 3/3 passes green (6× ok). The failure and fix recorded in AGENTS.md and the 06-34 report.
5. **Pre-session daemon-committed doc diffs REVIEWED (§f item 6)** — `docs/status/2026-10-07_04-57…` and `…05-14…` are harmless dprint table re-alignment + emphasis normalization. `docs/planning/2026-07-18_19-59…` was **semantic corruption**: the resolution banner's `~~~90% shipped~~~` (three tildes = CommonMark tilde code-fence opener) had been reformatted into a fenced code block, destroying the strikethrough. Fixed to `~~90% shipped~~` (2-tildes, formatter-stable); trap recorded in the AGENTS.md annotate bullet.
6. **GOEXPERIMENT ergonomics TODO item CLOSED with live proof** — the `.envrc` (`use flake` + `use_go_env`) already existed locally (2026-08-03) with exact fleet convention but was untracked: the buildflow-managed `.gitignore` block ignores `.envrc`. Added a `!.envrc` negation in the non-managed section (last-match-wins overrides; buildflow regenerates only inside its markers). **direnv verified live**: `direnv exec .` injects `GOEXPERIMENT=jsonv2` and go 1.26.8. `.envrc` is now committed (fleet-matching go-output/md-go-validator behavior).
7. **ROADMAP raw ideas made truthful** — struck with resolutions: gopls/LSP config-skew (resolved, root cause linked), art-dupl enforced-gate (DECIDED 2026-10-07), erraudit class-wide policy (RESOLVED, BuildFlow #34 link); ADDED entries for BuildFlow #35 (`reinstall` no-op) and #36 (`file-size-check` inert) with issue links. §f item 3 closed.
8. **`BenchmarkParseWiseTimestamp` added (§f item 35)** — bench_test.go, five sub-benchmarks covering every live Wise wire shape (RFC3339-Z, zoneless-T, space-separated, millis+numeric-zone, no-millis+numeric-zone), file's `b.Loop()`/`ReportAllocs` idiom. Smoke-run: compiles and passes all five (benchtime 1x — correctness proof, not perf numbers).
9. **Missing fuzz seed added (§f item 36)** — `FuzzParseWiseTimestamp` seed corpus gained `2026-10-07T00:17:01+0000`, the no-millis numeric-zone shape (the live `/v1/rates` regression shape from 2026-10-07); the other four live shapes were already seeded. Full suite green with seeds exercised.
10. **Cross-project lessons COMMITTED to crush-config `references/lessons.md`** (commit `5a50ee1`) — (1) `gh issue create --body-file -` can fail silently; write real files, verify after create; (2) sed range-extraction truncates nested-fenced bodies at the first inner fence; split/rsplit + byte-count before posting.
11. **github-voice skill gained the filing-mechanics step** (SKILLS commit `5df2e91`) — new step 8 "File mechanically, then verify it landed": `--body-file <real-file>` never `-`, verify with `gh issue list`/`view`, python split/rsplit extraction, byte-count. Fan-out verified by content through `~/.config/crush/skills/github-voice/SKILL.md`.
12. **Daemon commit spot-checks (§f item 38)** — all four commits verified to contain exactly my edits and nothing else: `3aad8e1` (7 files: bench, fuzz seed, ROADMAP, CONTRIBUTING, drafts README, 06-07 annotation, TODO_LIST), `7a743ac` (flake flag), `491b30e` (6 files incl. `.envrc` now tracked), `0dc97d6` (flag fix + annotations).
13. **Docs wiring** — CONTRIBUTING: new "Editor/LSP environments need it too" subsection + `.envrc` in the wired-in list; TODO_LIST: WATCH note on the filed-defects item (lift skip_steps on #34 fix; switch flake input on #8 fix), GOEXPERIMENT item closed, Go 1.27 item now carries the condition-1 probe result; 06-07 report §g Q1 annotated RESOLVED; 06-34 report §b3/§b4/§c1/§c3 struck with resolutions + §f resolution banner; new `docs/drafts/README.md` (convention: real files never /tmp, `--body-file <file>`, python extraction, no Go fences today — §f items 32+34 in one file).
14. **Go 1.27 condition-1 probe (§f item 26)** — `nix eval nixpkgs#go.version` = **1.26.8**, NOT met; recorded in TODO_LIST. Plan remains condition-gated.

## b) PARTIALLY DONE

1. **§f item 46 (shuffle) shipped as ×3, spec said ×6** — the original idea asked for a ×6 shuffle loop; I shipped ×3 (~11 s total). Deliberate: each pass is ~3.5 s, ×3 already randomizes order three times pre-CI; ×6 is one-line. Needs a one-word decision (see §f item 25).
2. **06-34 report §f annotated at BANNER granularity, not item-by-item** — the docs-health ANNOTATE precedent (2026-09-16) strikes items individually; I wrote a dated resolution banner listing executed item numbers instead. Shallower, faster; flagged for the next docs-health pass (§f item 23).
3. **apidiff evidence not yet attached to the TODO_LIST §g Q2 item** — the v0.13.0/only-delta fact lives in this report and the 06-34 annotation, but the question text the user reads in TODO_LIST was not extended (it edits the question the user must answer; I deferred — arguably over-deferred, it is a 2-min additive edit; see §f item 22).
4. **github-voice skill bundle partially swept** — SKILL.md updated, but the bundle's `references/` (voice-profile.md, revision-lessons.md) were not checked for a second filing-mechanics mention that should stay in sync.
5. **Coverage number is morning-stale** — 93.1% is the prior session's measurement; both pre-release runs enforce the 90 floor (passed), so ≥90 is proven, but §f item 15's ≥93 claim has no fresh number from this session.
6. **check-skill-fanout.sh is MISSING** — global AGENTS.md documents `bash ~/.config/crush/scripts/check-skill-fanout.sh`; the path does not exist. I verified the fan-out directly by content instead, but the documented guard is absent (restore script or fix the doc — crush-config territory).

## c) NOT STARTED (blocked or trigger-gated, listed for completeness)

1. **The three §g user questions** — x86_64-darwin strategy; `RequestLog.RetryAfterDelay` v1.0 lock; GOEXPERIMENT fleet scope. All execution-blocked on answers.
2. **Upstream watch items** — BuildFlow #34 (lift `skip_steps` after embedded-vs-standalone re-verification), #35, #36, md-go-validator #8 (flake input switch). All gated on issue closure.
3. **Phantom-recurrence watch** — nothing to do unless diagnostics reappear (reopen condition in AGENTS.md).
4. **P1/P2 user-blocked set** — sandbox live pass (`WISE_SANDBOX_API_KEY`), ottAPIVersion 2026Q3→2026Q4 flip, credentialed integration tests, v1.0 API-lock audit, typed recipient `Details` design, CACHIX/ERRAUDIT tokens, CI re-enable.
5. **CI-enable work** — wire `md-go-snippets` + `release-notes-check` into ci.yml (they are flake-local today; both ran green inside `nix flake check` twice this session).
6. **Go 1.27 execution** — conditions 2–4 all gated on condition 1 (nixpkgs stable 1.27; currently 1.26.8).
7. **Raw ideas pruned consciously** (§f 50) — rollover-probe app, global fence detector, golangci cache dir, coverage alerting, art-dupl WARN step, v1.0 CHANGELOG sketch, P3-battery alias app. ROADMAP fuel, not commitments.

## d) TOTALLY FUCKED UP (honest)

1. **Shipped the `--shuffle` flag BROKEN on first implementation — in the release-gate app.** `-count=3` instantly fails the suite (Ginkgo supports only `-count=1`). The constraint was knowable BEFORE writing the flag: AGENTS.md documents the Ginkgo suite and the shuffle-guard semantics; I implemented without consulting the test-stack contracts my flag sits on top of. The flag's own verification caught it within minutes — but a broken release-gate flag that ships unnoticed would have failed on tag day, the worst possible moment. Root cause: wrote code against the shell contract, not against the framework contract.
2. **Trashed the verification log BEFORE inspecting it** — `trash /tmp/pr-shuffle.log` ran in the same command as a grep that returned 0 hits, destroying the evidence I still needed. The exit-0 verdict stood (set -euo pipefail), but I had to re-run the 3-pass battery to produce honest evidence instead of trusting an inference. Cleanup must never precede evidence extraction.
3. **Heredoc append to `references/lessons.md` executed in the WRONG working directory** — ran in wise-go, failed with "no such file" (failed safe, nothing corrupted). The absolute path was in the very next command; the first construction was sloppy.
4. **Attempted to WRITE `.envrc` without an existence check** — collided with the "file modified since read" guard because the file already existed (since August 3). The failed write was the discovery mechanism for a fact an `ls` would have shown instantly, and it briefly mis-framed the task (create vs. track).
5. **Edited `.gitignore` after only `cat`-ing it** — tool rejected the edit (View-before-Edit is a hard rule I know); one wasted round trip.
6. **Background exit-code capture printed EMPTY** — `nix run … | tail; echo ${PIPESTATUS[0]}` in the background shell produced nothing (undiagnosed interaction of pipes + background capture). I switched to `cmd > logfile 2>&1; echo EXIT:$?` which worked, but the PIPESTATUS failure itself remains unexplained.
7. **Two of my own written-in-this-session artifacts needed immediate self-correction** — the drafts README originally contained three literal unbalanced ``` fence hazards (unclosed `markdown` fence, triple-backticks inside inline code); my own md-go-validator training kicked in only at re-read. Write-then-verify caught it pre-commit, but the first draft should never have contained them.

## e) WHAT WE SHOULD IMPROVE

1. **Existence-check before any write** — `glob`/`ls` the target path first whenever I didn't create it this session. The `.envrc` collision and the mis-framed task were both avoidable with one cheap call.
2. **Run logs are evidence until the close-out** — never trash them in the same command as inspection; keep until the final report is written.
3. **Read the framework contract before adding options on top of it** — Ginkgo's `-count` ban, `writeShellApplication` escaping, treefmt behavior: the touchpoint's constraints belong in the design step, not the debug step.
4. **Standardize background verification capture** — `cmd > logfile 2>&1; echo EXIT:$?` (worked twice); no pipes, no PIPESTATUS in background contexts until the empty-capture behavior is understood.
5. **Absolute paths for every cross-repo operation** — cwd assumptions cost one failed command this session.
6. **Attach decision evidence to the decision point** — the apidiff fact belongs on the TODO_LIST question text the user actually reads, at discovery time, not only in a timestamped report.
7. **Skill updates sweep the whole bundle** — grep `references/` too, not just SKILL.md.
8. **Cheap warnings noticed, not fixed (twice seen in logs)** — all 8 flake apps warn `lacks attribute 'meta.description'` in `nix flake check`; trivial polish item (§f item 24).

## f) NEXT — up to 50 things to get done next

**Verify / watch (immediate, unblocked)**

1. Watch BuildFlow #34 → on fix: re-run embedded vs standalone byte-for-byte, lift the `skip_steps` (TODO_LIST WATCH).
2. Watch BuildFlow #35 (reinstall) and #36 (file-size-check) → verify profile switch / non-zero scan on fix.
3. Watch md-go-validator #8 → switch wise-go's flake input to upstream flake; drop `go_1_27` override + vendorHash; re-lock (TODO_LIST WATCH).
4. Next session start: `lsp_diagnostics` tripwire on `internal_test.go` + `webhooks.go` (AGENTS.md reopen condition).
5. Next tag day: use `nix run .#pre-release -- --shuffle` (now the verified invocation).
6. Fresh coverage measurement at next code-touching session (§f 15; 93.1% is morning-stale, floor 90 held twice today).
7. One-probe check: does `gh release create --notes-file -` share the silent-pipe failure class of `gh issue create --body-file -`?

**Blocked on user (execute immediately after answers)**
8. x86_64-darwin: pin nixpkgs-26.05-darwin OR drop from `--all-systems` (§g Q1).
9. `RequestLog.RetryAfterDelay`: LOCK current shape for v1.0 OR dedicated observer interface (§g Q2) — evidence: only compatible delta since v0.12.0, gorelease suggests v0.13.0.
10. GOEXPERIMENT scope: project-local (`.crushrc` + tracked `.envrc`) vs home-manager fleet rollout (§g Q3).
11. Sandbox live-verification pass (`WISE_SANDBOX_API_KEY`).
12. Flip `ottAPIVersion` 2026Q3 → 2026Q4 after the live probe.
13. Credentialed sandbox integration tests.
14. v1.0 API-lock audit re-run + growth lineage.
15. Typed recipient `Details` design decision (per-corridor structs vs `map[string]string`).
16. Set `CACHIX_AUTH_TOKEN` + `ERRAUDIT_TOKEN` secrets.
17. Re-enable CI on GitHub (push → enable → first green run).

**Open P3 (unblocked, small)**
18. Wire `md-go-snippets` + `release-notes-check` into ci.yml when CI enables.
19. Resolve `check-skill-fanout.sh` absence: restore the script or fix global AGENTS.md (crush-config commit).
20. De-duplicate the LSP-resolution narrative (ROADMAP struck idea vs AGENTS.md gotcha) — one canonical home.
21. Attach the apidiff v0.13.0 evidence line to TODO_LIST's §g Q2 item (2-min additive edit).
22. Item-level strikethrough pass on the 06-34 §f list at next docs-health (banner summarizes today).
23. Add `meta.description` to the 8 flake apps (silences every `nix flake check` warning seen today).
24. Decide shuffle depth: ×3 (shipped) vs ×6 (§f 46 original) — my unopposed call: ×3 pre-CI, ×6 if CI measures flake rate.
25. Record the Ginkgo `-count>1` ban as a cross-project lesson (crush-config) — Ginkgo fleet repos hit this identically.
26. Consume the benchstat baseline in the next perf-touching PR (`docs/bench/2026-10-07_v0120_baseline.txt`).
27. Hold coverage ≥93 through the next feature PR (floor 90 enforced by flake).

**Go 1.27 (condition-gated)**
28. Re-probe `nix eval nixpkgs#go.version` at toolchain-touching sessions (2026-10-07: 1.26.8, not met).
29. Pre-verify in-house deps (go-branded-id, go-error-family, go-retry) on 1.27 opportunistically.
30. Test erraudit local install the day the toolchain clears 1.27.
31. Flip-day: remove buildflow go-version skip_steps LAST (condition 4); single-commit flip + rollback per the plan doc.
32. After flip: adopt modernizer 1.27 syntax deliberately in one commit; track json/v2 retirement separately.

**Docs**
33. CONTRIBUTING: one line on `--shuffle` for tag day (AGENTS.md has it; CONTRIBUTING doesn't yet).
34. When upstream issues close: ANNOTATE the 06-34 report and this report inline (docs-health pattern).
35. FEATURES.md: surface unchanged this session — doc-verify count claims green twice today; re-verify at next docs-health.

**Raw ideas (pruned as not-commitments; ROADMAP fuel)**
36. Rollover-probe flake app (§f 49 of 06-34).
37. `release-notes-check`: global unbalanced-fence detector (§f 45).
38. Project-local golangci analysis cache (§f 44).
39. Coverage alerting check: fail >1pt below 93.1 (§f 43).
40. art-dupl WARN step when CI enables (§f 47).
41. Sketch the v1.0 CHANGELOG structure (§f 48).
42. Scan `docs/drafts/` for Go fences only if any appear (convention recorded in drafts README).
43. "P3 battery" alias app (§f 37).
44. Expand webhook fuzz corpus (older plans).
45. Prune TODO_LIST + ROADMAP at the next docs-health pass (§f 50 of 06-34).

## g) QUESTIONS FOR LARS (cannot self-answer)

1. **x86_64-darwin in `nix flake check --all-systems`:** nixpkgs-26.11 dropped the system (aarch64-linux + aarch64-darwin evaluate clean). Pin a `nixpkgs-26.05-darwin` input just for darwin checks, or drop x86_64-darwin from the matrix? Decides the flake-input shape and any future CI matrix.
2. **`RequestLog.RetryAfterDelay` API freeze:** LOCK the current shape (delay-decision entries riding the `LogRequest` channel, `Method`/`URL` empty, honored wait as `time.Duration`) for v1.0, or grow a dedicated observer/callback interface? **New hard evidence today:** apidiff shows this is the ONLY compatible change since v0.12.0 (gorelease suggests v0.13.0) — whichever way you decide shapes the next tag.
3. **GOEXPERIMENT fix scope:** keep it project-local (wise-go `.crushrc` + tracked `.envrc`, both now verified working) or move it into home-manager/crush-config so every jsonv2 repo on every machine gets it automatically? Fleet policy call — I cannot enumerate which repos need it or how many machines repeat this bug.
