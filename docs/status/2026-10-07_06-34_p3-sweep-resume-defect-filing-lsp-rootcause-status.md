# P3 Sweep Resume — Defect Filing, LSP Root Cause, Honest Gaps

**Date:** 2026-10-07 06:34 CEST
**Session:** resume of the P3 quality/tooling sweep (`docs/status/2026-10-07_06-07_p3-quality-tooling-sweep-status.md`), executing its leftover next-steps.
**Tree:** clean at report time (auto-commit daemon swept everything); HEAD `47bf3a7`.

---

## a) FULLY DONE

1. **All four tool-repo defects FILED upstream** (drafted → voice-checked 0 FAIL → posted → links verified):
   - BuildFlow [#34](https://github.com/LarsArtmann/BuildFlow/issues/34) — embedded `erraudit` reports 29 false positives the standalone binary does not (8× `[ignored]` misreads of `_, ok := errors.AsType[T](err)` + `_ = resp.Body.Close()`; 21× `[context_loss]` full-variable-echo demands contradicting the error-context convention).
   - BuildFlow [#35](https://github.com/LarsArtmann/BuildFlow/issues/35) — `nix run .#reinstall` prints REINSTALL-OK but never switches `~/.nix-profile`.
   - BuildFlow [#36](https://github.com/LarsArtmann/BuildFlow/issues/36) — `file-size-check` scans 0 files in root-package Go layouts; `max_file_size` inert.
   - md-go-validator [#8](https://github.com/LarsArtmann/md-go-validator/issues/8) — flake cannot build its own package (go.mod 1.27 vs buildGoModule default go 1.26.7); evidence freshly reproduced (`nix build .#default` fails verbatim) with `package.nix:21` quoted.
   - Voice-checked drafts kept in `docs/drafts/` (skill policy: never /tmp).
2. **LSP phantom ROOT-CAUSED with A/B proof** — both gopls and golangci_lint_ls are launched by Crush OUTSIDE the nix devShell and lacked `GOEXPERIMENT=jsonv2`; the repo's `encoding/json/v2` imports are build-constraint-excluded without it, so package load FAILED and both servers produced phantoms from degraded analysis. Proof: `env -u GOEXPERIMENT gopls check internal_test.go` → `build constraints exclude all Go files` + `undefined: json`; with the var set the package loads. This unifies the `bytes unused` false positive (internal_test.go:4 — `bytes` IS used at :1162/:1166), the `webhooks.go:142` "File is not properly formatted" phantom, and likely every prior mystery LSP diagnostic.
3. **Fix deployed for the phantom** — project-local `.crushrc` re-declares gopls + golangci_lint_ls with `--env GOEXPERIMENT jsonv2` (bash -n validated; activates at next session start; `stdversion` deliberately stays disabled — with the experiment ON, `gopls check` flags json/v2 usage as "requires go1.27" under the 1.26 pin while the suite compiles and passes). Poisoned golangci analysis cache trashed. Recorded as an AGENTS.md gotcha with a reopen condition.
4. **AGENTS.md memory sweep (11 edits total)** — shuffle-proof conformance guard paragraph (floors 25/60/4, actuals ≥25/180/6, `conformanceSuiteCompleted` semantics); art-dupl enforced-gate DECISION recorded (baseline is policy, no gate, revisit warn-first at CI re-enable); erraudit CLASS-WIDE policy recorded (skip as a class, standalone gate governs, never per-finding); md-go-snippets wiring + source-only-input/vendorHash/go-override gotchas; release-notes-check + pre-release apps; `nix flake check --all-systems` x86_64-darwin blocker; `RequestLog.RetryAfterDelay` delay-decision entry semantics; coverage 93.1% vs frozen badge; defect bullet updated with filed issue links; LSP-phantom gotcha.
5. **TODO_LIST.md harvested** — P3 section rewritten: 14 items closed with resolutions + pointers, 3 genuinely open (all-systems decision, Go 1.27 execution, plus phantom item closed with reopen condition); defect-filing item closed with the four issue links.
6. **Go 1.27 migration plan written** — `docs/planning/2026-10-07_go-1.27-migration-plan.md`: 4 unlock conditions (nixpkgs stable 1.27; in-house deps verified on 1.27; erraudit local install; buildflow go-version skip_steps removable), single-commit flip procedure, single-commit rollback, residual risks (jsonv2 retirement tracked separately), re-check cadence (`nix eval nixpkgs#go.version`).
7. **P3 sweep carry-over from the prior session confirmed intact** — full suite green, golangci-lint 0 issues, coverage 93.1%, conformance actuals 180/6/≥25 (from prior session's verification; unchanged — docs-only edits this session).

## b) PARTIALLY DONE

1. **LSP env fix is deployed but NOT yet verified end-to-end** — `.crushrc` loads at session START, and `lsp_restart` re-forks with the session-loaded (old) env, so the fix is provably-correct-by-A/B but not yet observed killing the live phantoms. Verification belongs to the next session.
2. **The 06:07 report's three user questions** — 1 of 3 resolved this session (art-dupl policy: decided + recorded); 2 remain open for the user (x86_64-darwin strategy; `RequestLog.RetryAfterDelay` v1.0 API lock) — see §g.
3. **ROADMAP consistency** — AGENTS.md previously said the three BuildFlow defects were "filed as ROADMAP raw ideas until upstream fixes land"; they are now FILED upstream (#34–36), but I did not check/update the ROADMAP raw-idea entries to point at the issues.
4. **Pre-existing dirty docs at session start** (`docs/planning/2026-07-18_19-59…`, `docs/status/2026-10-07_04-57…`, `docs/status/2026-10-07_05-14…`) — committed by the daemon mid-session; the diffs were authored before this resume and never reviewed by me.

## c) NOT STARTED

1. **`nix run .#pre-release` end-to-end** — the first full-chain run (incl. apidiff, needs network) still has never happened; it was queued when the tree was dirty and deferred behind the defect filing.
2. **x86_64-darwin decision execution** — blocked on the user's §g answer (pin nixpkgs-26.05-darwin vs drop the system).
3. **GOEXPERIMENT ergonomics TODO item** (direnv/home-manager pin) — untouched; `.crushrc` covers only Crush-launched LSPs, not arbitrary shells.

## d) TOTALLY FUCKED UP (honest)

1. **`gh issue create` failed silently TWICE before any issue landed** — first attempt: "no output", no issue created (only caught because I listed issues afterward); second attempt (batched 3-in-a-row via piped `--body-file -`): all three silently created NOTHING. Working form: `--body-file <real-file>` (exit 0, URL on stdout). The pipe form prints no error to stdout or stderr under this shell. Cost: the "file the defects" step appeared briefly stuck; only systematic verify-after-each saved it.
2. **The extraction almost filed TRUNCATED issue bodies** — the sed-based fence extraction (`/^```markdown$/,/^```$/`) stops at the FIRST closing fence, so drafts containing nested code fences extracted at 784 and 135 bytes (of ~1.5 KB). Caught by byte-count check before posting, but the first create attempt ran against an unverified pipeline — the failure that followed was lucky, not designed.
3. **Two TODO_LIST.md multiedit accidents** — the phantom-item edit ALSO deleted the defects item; the restore edit then clobbered the md-go-validator closed item. Both caught by post-edit greps and restored, but three consecutive editing mistakes in one file is sloppy discipline — the exact failure mode "include more context in old_string" exists to prevent.
4. **Exit codes masked by pipes** — `env -u GOEXPERIMENT go build ./... 2>&1 | head -5; echo "exit: $?"` reported head's exit (0), not go's. The error TEXT made the verdict obvious anyway, but the printed "build exit: 0" was false evidence I nearly quoted.
5. **From the prior session (still unfixed): the pre-release chain has never run end-to-end** — every component has run individually, the composed app has not.

## e) WHAT WE SHOULD IMPROVE

1. **Verify-after-every-mutating-call, mechanically** — the gh silent failures and the truncation near-miss were both caught only because I re-checked state; make "list/verify after create" a reflex, not a salvage.
2. **Never pipe bodies/secrets through `gh --body-file -`** — write a real file first (also keeps a paper trail in docs/drafts/).
3. **Byte-count any extracted content before posting it** — 135 bytes for a 1.6 KB issue was the tell.
4. **Multiedit discipline: never replace an old_string with new content that doesn't retain it** — when closing a TODO item, target ONLY the item's exact block.
5. **Exit-code checks must not share a pipeline with text filters** — capture exit before `| head`/`| tail`.
6. **Env-parity checks for editor tooling** — any tool launched outside the devShell (LSPs, daemons, hooks) must be A/B-tested against the devShell env once per repo-level toolchain requirement; jsonv2 will bite again in the Go 1.27 flip if the wrapper path is forgotten.
7. **Record the gh-pipe silent-failure quirk as a cross-project lesson** (crush-config `references/lessons.md`) — it will bite every repo that files issues from scripts.

## f) NEXT — up to 50 things to get done

**Verify the fixes landed (this week)**
1. Next session: confirm `.crushrc` env fix kills both phantoms (gopls `bytes unused`, golangci webhooks.go:142); reopen if not.
2. Run `nix run .#pre-release` end-to-end on the clean tree (first full-chain validation incl. apidiff).
3. Update ROADMAP raw-idea entries for the three BuildFlow defects to point at #34–36.
4. Watch BuildFlow #34–36 + md-go-validator #8; when #34 fixes, re-run the embedded step byte-for-byte vs standalone, then lift the `skip_steps`.
5. When md-go-validator #8 fixes: switch wise-go's flake input from source-only to upstream flake; drop the `go_1_27` override + vendorHash dance; re-lock.
6. Review the three pre-session dirty docs (planning 2026-07-18, status 04-57, status 05-14) that the daemon committed unreviewed.
7. Re-run `golangci-lint run` + `go test .` once after the doc-only commit burst (cheap green confirmation).
8. Record the gh-pipe silent-failure lesson in crush-config `references/lessons.md` (needs a commit there).
9. Decide whether `.crushrc` GOEXPERIMENT fix should ALSO land in the home-manager wrapper for fleet-wide coverage (ties to §g Q3).
10. Add the four filed issues to the docs-health watch list — they gate wise-go's own gates when fixed.

**Blocked on user (execute immediately after answers)**
11. x86_64-darwin: pin a nixpkgs-26.05-darwin input OR drop the system from `nix flake check --all-systems` (§g Q1).
12. `RequestLog.RetryAfterDelay`: lock the Logger-channel shape for v1.0 OR design a dedicated observer interface (§g Q2).

**Open P3 (from TODO_LIST)**
13. Root-cause any phantom RECURRENCE after 2026-10-07 (reopen condition documented in AGENTS.md).
14. GOEXPERIMENT ergonomics: direnv/home-manager pin so `jsonv2` is set without `.buildflow.yml` env injection.
15. Keep coverage ≥ 93% through the next feature PR (floor 90).
16. Consume the benchstat baseline in the next perf-touching PR (`benchstat docs/bench/2026-10-07_v0120_baseline.txt new.txt`).
17. Wire `md-go-snippets` + `release-notes-check` into CI's job list when CI re-enables (they are flake-local today).

**Open P1/P2 (user-blocked, listed for completeness)**
18. Sandbox live-verification pass (`WISE_SANDBOX_API_KEY`).
19. Flip `ottAPIVersion` 2026Q3 → 2026Q4 once the OTT surface verifies live (probe ritual in CONTRIBUTING).
20. Credentialed Wise sandbox integration tests.
21. v1.0 API lock audit (re-audit + growth lineage).
22. Typed recipient `Details` (per-corridor structs vs `map[string]string`) — design decision.
23. Set `CACHIX_AUTH_TOKEN` secret.
24. Set `ERRAUDIT_TOKEN` secret (PAT read access to private erraudit).
25. Re-enable CI on GitHub (push → enable → first green run).

**Go 1.27 migration (per the plan; condition-triggered)**
26. Check `nix eval nixpkgs#go.version` on toolchain-touching sessions; start at condition 2 when it prints 1.27.x.
27. Pre-verify in-house deps (go-branded-id, go-error-family, go-retry) on a 1.27 toolchain opportunistically.
28. Test erraudit local install the day the toolchain clears 1.27.
29. On flip day: remove buildflow go-version skip_steps LAST (condition 4).
30. After flip: adopt the modernizer's 1.27 syntax (promoted-field literals) deliberately, in one commit.
31. Track json/v2 retirement separately (when in-house deps drop jsonv2 imports, remove the env in one commit).

**Tooling/polish**
32. Consider a `docs/drafts/README.md` convention note (drafts live here, bodies are fenced, never /tmp).
33. Add a `--body-file <file>` (never `-`) note to the github-voice skill's filing steps if Lars agrees (crush-config commit).
34. Evaluate whether the `md-go-snippets` check should also scan `docs/drafts/` (draft bodies contain Go snippets? currently no).
35. Bench: add `BenchmarkParseWiseTimestamp` covering the four known layouts (guards the tolerant parser's hot path).
36. Fuzz: seed `parseWiseTimestamp` corpus with the five live timestamp shapes from AGENTS.md.
37. Consider a tiny `make`-free `justfile`-style alias in flake apps for "run the exact P3 verification battery" (test, lint, fmt, dupl, fuzz smoke).
38. Confirm the daemon committed `.crushrc`, TODO_LIST, AGENTS.md, plan doc, drafts intact (spot-check `git show` on the next auto-commit).

**Documentation**
39. Fold the LSP-phantom root cause into CONTRIBUTING's troubleshooting section (one paragraph, links to AGENTS.md).
40. Annotate `docs/status/2026-10-07_06-07_p3-quality-tooling-sweep-status.md` §g question ① as RESOLVED (art-dupl policy recorded).
41. When issues #34–36 close, annotate this report inline (docs-health ANNOTATE pattern).
42. Refresh FEATURES.md if the v1.0 audit changes the 41-method count (unchanged this session).

**Nice-to-have (raw ideas, not commitments)**
43. Baseline alerting: a flake check that fails if coverage drops >1pt below 93.1% (hair-trigger insurance beyond the 90 floor).
44. golangci-lint cache: consider project-local cache dir to avoid cross-repo analysis-cache poisoning (today: trashed by hand).
45. `release-notes-check`: extend the odd-backtick detector to catch unbalanced fences globally (currently per-file).
46. pre-release app: add `-shuffle` test loop (×6) as an optional flag.
47. pre-release app: add the art-dupl manual gate as a WARN step when CI enables.
48. Sketch the v1.0 CHANGELOG structure now (the API-lock audit will need it).
49. Extract the "quarterly rollover probe" into a flake app so the probe is one command, not a CONTRIBUTING transcription.
50. PRUNE this list at the next docs-health pass — 50 items is a ceiling, not a quota.

## g) QUESTIONS FOR LARS (cannot self-answer)

1. **x86_64-darwin in `nix flake check --all-systems`:** nixpkgs-26.11 dropped the system (verified 2026-10-07; aarch64-linux + aarch64-darwin evaluate clean). Pin a `nixpkgs-26.05-darwin` input just for darwin checks, or drop x86_64-darwin from the check matrix? This decides the flake-input shape and any future CI matrix.
2. **`RequestLog.RetryAfterDelay` API freeze:** v1.0 is approaching — do you want the current shape (delay-decision entries riding the `LogRequest` channel, `Method`/`URL` empty, honored wait as `time.Duration`) LOCKED for v1.0, or should the SDK grow a dedicated observer/callback interface (bigger surface, cleaner long-term)? Locking now avoids a breaking change later.
3. **Scope of the GOEXPERIMENT LSP fix:** keep it project-local in wise-go's `.crushrc` (as shipped), or move it into the home-manager wrapper/`crush-config` so every jsonv2 repo on every machine gets it automatically? Fleet policy call — I can't derive which repos need it or how many machines repeat this bug.
