# Status: TODO-List Execution — Cross-Repo Lessons Landed, Fan-Out Guard Verified, Live BuildFlow `lll` Incident Handled

> Session 2026-10-16 08:26–08:52 CEST. Input: the current `TODO_LIST.md`
> (P1/P2 user-gated + P3 + cross-repo items) with the instruction to execute
> everything actionable. Scope: only what this session touched; no new
> research beyond the work itself.

## Verdict

Both Cross-repo items are DONE (one turned out already-resolved, verified
end-to-end). One P3 item is a future-consumption note (no action possible
yet). Everything else in the TODO list is BLOCKED on user input and was
left untouched. Mid-session, a live `buildflow --fix --build-mode=full` run
from another terminal modified the working tree under me; its config damage
was reviewed per the documented gotcha and surgically reverted.

---

## a) FULLY DONE

1. **Cross-repo lesson recording** (TODO_LIST "Cross-repo" item 1) —
   committed to crush-config as `0607ed3`:
   - go-retry drift lesson ("a `go.mod` version bump must update the
     matching flake input rev; pinned revs are immutable under
     `nix flake update`", incl. the annotated-tag-object vs peeled-commit
     trap and the pin-sync-guard remedy).
   - Ginkgo `-count>1` ban (only `-count=1` supported; repetition = separate
     invocations; framework contract is a design input).
   - The `gh issue create --body-file -` quirk was ALREADY recorded
     (`references/lessons.md` line 691, commit `5a50ee1`) — verified before
     writing, so no duplicate entry was created.
   - Daemon raced my append and heuristic-committed it (`0a5cc09`); per the
     repo's own verify-then-amend lesson, HEAD was verified (author, file
     list = only lessons.md, 34 insertions = my append) and the message
     amended to the proper `docs(lessons):` form.
   - Format verification EXECUTED (not assumed): `nix fmt` traversed and my
     appended lessons are format-clean; the one change it wanted was a
     pre-existing one-char code-span nit in an older lesson, fixed and
     committed as `9eb3cab`.
2. **`check-skill-fanout.sh` absence** (Cross-repo item 2) — verified
   ALREADY RESOLVED, no change needed:
   - `scripts/check-skill-fanout.sh` EXISTS in the crush-config repo.
   - The repo AGENTS.md (source of truth) already carries the corrected
     text ("run the guard from this repo; the deployed copy lives in the
     nix store, wired by `crushrc` — there is no `~/.config/crush/scripts/`").
   - `~/.config/crush/scripts/` indeed does not exist — the doc claim is
     accurate.
   - Deployed layer verified: `~/.config/crush/AGENTS.md` (nix-store symlink
     via home-manager) is byte-identical to crush-config HEAD.
   - Guard executed: `bash scripts/check-skill-fanout.sh` →
     "OK: 48 skill symlink(s) resolve cleanly", exit 0.
3. **wise-go TODO_LIST.md updated** — the "Cross-repo" section removed
   (both items done; completed work never lives in TODO_LIST). Daemon
   committed it as `7c13bbb` (13 deletions, exactly my edit — verified).
4. **Blocked-item verification** — `WISE_SANDBOX_API_KEY` is NOT in the
   environment (env grep + no `.env` files beyond the 21-byte GOEXPERIMENT
   `.envrc`), so P1 (sandbox verification pass, OTT Q4 flip) and the P2
   sandbox integration-test item remain genuinely blocked, as documented.
5. **Live BuildFlow incident handled** (see d) for what went wrong first) —
   a `buildflow --fix --build-mode=full --budget 5m` (PID 45951, started
   08:29 from another terminal) was racing this session. After it finished:
   - Its changes were daemon-committed as `1cf83ab` (`.github/workflows/ci.yml`
     install-nix-action v31.11.1→v31.11.2 pin bump, `.golangci.yml`,
     `flake.lock` nixpkgs bump).
   - Diff reviewed hunk-by-hunk per the AGENTS gotcha ("buildflow
     auto-configure is dangerous … the project's curated linter list is
     intentional"):
     - KEPT: `go: 1.26.7` → `go: "1.26"` (matches go-mod-normalize canonical
       form), yaml quote normalization (semantically identical), the action
       pin bump, the nixpkgs bump (the documented vulnix lever).
     - REVERTED: the `- lll` linter addition — 14 findings, EVERY one on
       deliberate lines (art-dupl:accept / nolint directives whose
       placement-and-length is constrained by documented gotchas, verbatim
       documented webhook payload/signature test fixtures) — and the two
       stripped exclusion-rule rationale comments (tagliatelle webhook
       snake_case; nolintlint `//nolint:branching-flow:` source-rule).
     - Committed as `d8e28b5` with the policy rationale in the message.
6. **Gates re-verified after the incident**: `golangci-lint run` → 0 issues;
   `go test ./...` → ok (both packages); `grep '^go ' go.mod` → `go 1.26.0`
   (≤ 1.26.x invariant holds; pre-existing, not from this window).

## b) PARTIALLY DONE

1. **crush-config delivery** — lessons are committed to the SOURCE repo, but
   the deployed `~/.config/crush/references/lessons.md` is a nix-store
   symlink that only updates on the next home-manager activation. The two
   new lessons are in git but NOT yet on disk in the fan-out. Also: local
   master is now 4 commits ahead of origin (2 pre-session + `0607ed3` +
   `9eb3cab`); I do not push without explicit instruction.
2. **Today's BuildFlow run review** — I reviewed the TREE diff it produced,
   but not the run's own findings summary (it ran in another terminal with
   `--log-level warn`; its report was never inspected from here).

## c) NOT STARTED (and why)

Everything user-gated, correctly untouched:

1. P1 sandbox live-verification pass (5 spec-vs-live assumptions) — needs
   `WISE_SANDBOX_API_KEY`.
2. P1 OTT `2026Q3 → 2026Q4` flip — needs the key (probe-blind surface).
3. P2 credentialed sandbox integration tests (first recorded run +
   CHANGELOG entry) — needs the key.
4. P2 v1.0.0 tag — needs explicit approval (irreversible).
5. P2 `RequestLog.RetryAfterDelay` API freeze — needs design decision.
6. P2 typed recipient `Details` — needs design decision.
7. P2 `CACHIX_AUTH_TOKEN` secret — needs the token.
8. P2 `ERRAUDIT_TOKEN` secret — needs the PAT.
9. P2 CI re-enable (push master + `gh workflow enable ci` + watch first
   run) — needs approval to push.
10. P3 decision: nixpkgs-26.05-darwin pin vs drop x86_64-darwin — user
    decision.
11. P3 Go 1.27 migration — blocked on nixpkgs stable 1.27 (last probe:
    1.26.8) + 3 other unlock conditions.
12. P3 benchstat baseline consumption — deliberately untouched: it is a
    note for the NEXT perf-touching PR, and none is in flight.

## d) TOTALLY FUCKED UP (caught, no lasting damage)

1. **Ran verification gates against a moving tree.** I launched
   `golangci-lint run` in the background at ~08:31 while a live buildflow
   was still modifying `.golangci.yml` (finished ~08:34). My 14-finding
   "verification" could have been stale or incomplete — it happened to hold
   because the config didn't change again in that window, but that is luck,
   not rigor. Root cause: I never checked `ps` for repo-modifying tools
   before running gates, despite this repo's documented history of
   concurrent buildflow runs. The correct sequence: detect the process
   FIRST, wait, then verify.
2. **Claimed formatter compliance by reasoning, not execution.** In the
   moment I justified lessons.md style with "dprint's default textWrap is
   maintain" from memory. That is assumption dressed as verification. I
   closed this later in the session (`nix fmt` executed — clean), but the
   initial claim should never have been made unexecuted.
3. Minor: one transient misread of an empty-looking `git status --short`
   output (display truncation) — immediately re-checked, no wrong action
   taken; and the daemon heuristic-committed my TODO_LIST.md edit
   (`7c13bbb`) while I hand-committed `.golangci.yml` (`d8e28b5`) —
   inconsistent commit hygiene between two of my own changes.

## e) WHAT WE SHOULD IMPROVE

1. **Gate-run preconditions**: in this repo (and fleet-wide under the
   auto-commit daemon), `ps aux | grep buildflow` before running any gate —
   a moving tree invalidates verification. Consider a crush-config lesson.
2. **Never claim tool compliance without running the tool** — formatter,
   linter, validator: execute or say "unverified".
3. **Structural prevention for the `lll` class**: the AGENTS gotcha is
   prose-only; nothing stops the next `buildflow --fix` from re-adding
   `lll` (or other linters) tomorrow. A small check comparing the enabled
   linters set against a pinned allowlist would make the curated-list
   policy a gate instead of a memory.
4. **Commit-message consistency**: author EVERY meaningful change's commit
   myself where the policy matters, or accept daemon heuristic messages
   uniformly — not a mix decided by mood.
5. **BuildFlow run provenance**: when a foreign buildflow run lands changes,
   review its findings summary too, not just the tree diff (warning-class =
   triage queue per AGENTS).

## f) NEXT — up to 50 (realistic, deduplicated; ~25 real items)

**Unblocks (user input, from the TODO list):**

1. Provide `WISE_SANDBOX_API_KEY` (env or GH secret) — unblocks items 2–4.
2. Run the sandbox live-verification pass (5 assumptions: statement
   COMPACT/FLAT, `X-idempotence-uuid`, `details.type` enum, `/v1/rates`
   array shape, text/plain funding errors).
3. Probe `GET /2026Q4/one-time-token/status`; flip `ottAPIVersion` if
   non-404; `go test ./...`; commit.
4. First recorded sandbox-live workflow run + CHANGELOG entry.
5. Approve + tag `v1.0.0`.
6. Decide `RequestLog.RetryAfterDelay` freeze (freeze shape vs observer
   interface) — apidiff says it is the only compatible delta vs v0.12.0.
7. Decide typed recipient `Details` (v1.0 freezes the map; typed accessors
   later?).
8. Set `CACHIX_AUTH_TOKEN` secret; confirm the `larsartmann` cache exists.
9. Set `ERRAUDIT_TOKEN` secret; then delete the `continue-on-error`/`if`
   guard lines in the erraudit CI job to make it blocking.
10. Approve push + `gh workflow enable ci` + watch the first run (badge
    unfreezes after).

**Decisions / watches:**

11. Decide nixpkgs-26.05-darwin pin vs dropping x86_64-darwin from the
    check matrix.
12. Re-probe `nix eval nixpkgs#go.version` periodically for Go 1.27
    condition 1 (plan: `docs/planning/2026-10-07_go-1.27-migration-plan.md`).
13. Track upstream fixes: BuildFlow #34/#35/#36, md-go-validator #8 (then
    drop the source-only build override in wise-go's flake).

**Unblocked quality work:**

14. Deploy crush-config (home-manager activation) so the two new lessons +
    `9eb3cab` reach the live fan-out; then re-verify through
    `~/.config/crush/references/lessons.md`.
15. Push crush-config master (4 commits ahead; user call).
16. Review TODAY's buildflow run findings summary (it ran ~08:29–08:34;
    warning-class = triage queue).
17. Add a recurrence note to the wise-go AGENTS buildflow-auto-configure
    gotcha: `lll` re-added 2026-10-16, reverted in `d8e28b5` — the class
    recurs, consider the allowlist gate (item 18).
18. Build a linter-allowlist check (enabled-linters set must match a pinned
    list) to make the curated-`.golangci.yml` policy structural.
19. Audit `git log 93601d1..HEAD` in full (daemon-committed `1cf83ab` /
    `7c13bbb` / `d8e28b5` — verify each daemon commit contains exactly what
    was reviewed; the AGENTS daemon-race gotcha demands it).
20. Run `nix flake check` once on wise-go after today's `flake.lock`
    nixpkgs bump (buildflow likely ran its own, but nobody verified from
    here; also exercises `checks.pin-sync` hermetically).
21. Fleet-wide rev↔go.mod drift sweep (from the 07-35 backlog: md-go,
    go-output, dynamic-markdown-site, …) — the drift class was proven
    fleet-wide; consider porting the pin-sync check pattern.
22. Next perf-touching PR: install benchstat
    (`go install golang.org/x/perf/cmd/benchstat@latest`, NOT in devShell)
    and consume `docs/bench/2026-10-07_v0120_baseline.txt`.
23. Re-run the manual art-dupl pass
    (`/home/lars/go/bin/art-dupl -t 2 --type-aware .`) if any
    cloned-adjacent code changes (baseline: 0 actionable, 16 groups).
24. Consider a crush-config lesson for "check ps for live buildflow before
    running gates in daemon-managed repos" (improvement e.1).
25. After CI re-enables: confirm the coverage-badge job rewrites README to
    current coverage (frozen at last CI-measured value; local 92.8% vs
    92.0 floor as of 2026-10-11).

## g) QUESTIONS ONLY YOU CAN ANSWER (max 3)

1. **`WISE_SANDBOX_API_KEY`**: can you provide it (now, or as a GitHub
   secret)? It single-handedly unblocks the two P1 items AND the P2
   sandbox-test first run.
2. **v1.0.0**: approve tagging now, or hold until the
   `RequestLog.RetryAfterDelay` freeze decision (question of freezing the
   delay-entry shape vs growing an observer interface first)?
3. **crush-config delivery**: should I push the 4 unpushed master commits
   and trigger the home-manager activation myself, or do you own
   deploy/push for that repo? (The new lessons reach future sessions only
   after activation.)

---

## Session commit ledger

| Repo | Commit | Content |
| --- | --- | --- |
| crush-config | `0607ed3` | two lessons (pin-drift + Ginkgo count ban), message amended over daemon sweep `0a5cc09` |
| crush-config | `9eb3cab` | dprint code-span nit fix in an older lesson (from the executed format verification) |
| wise-go | `7c13bbb` | TODO_LIST.md cross-repo section removed (daemon commit of my edit) |
| wise-go | `d8e28b5` | revert `lll` + restore two rationale comments after buildflow auto-configure |
| wise-go | `1cf83ab` | daemon commit of buildflow's ci.yml/.golangci.yml/flake.lock (kept, partially reverted by `d8e28b5`) |
