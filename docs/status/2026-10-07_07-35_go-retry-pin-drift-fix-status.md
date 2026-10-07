# Status: go-retry usage audit → flake-pin drift discovered and fixed

> **Scope:** this session only (2026-10-07, ~07:05–07:35 CEST). Trigger question:
> "Do we use `/home/lars/projects/go-retry`? If not, why not? Should we?"
> No other research was done; older open items are listed only where this
> session touched or observed them.

## Session summary

One question produced one answer, one latent defect, and one fix:

1. **Answer:** yes, `go-retry` is the SDK's retry executor (ADR 003), used via
   `go.mod` + a rev-pinned flake input — never as a local path.
2. **Defect found:** `go.mod` was bumped to go-retry v0.7.1 on 2026-09-27 while
   `flake.nix` stayed pinned at the v0.6.0 commit — direct `go test` and the
   sandboxed `nix flake check` verified **different versions for ten days**.
3. **Fix shipped:** flake pin aligned to the v0.7.1 tag, vendorHash repaired via
   buildflow, `nix flake check` green, AGENTS.md updated with the sync rule.

## Self-critique (asked directly: what did I forget / do better / still improve?)

### What I forgot

1. **The other three in-house deps.** I proved the drift pattern exists in this
   repo, then checked exactly ONE of the four rev-pinned inputs. `go-branded-id`
   and `go-error-family` (and `go-nix-helpers`) have the exact same
   rev-vs-go.mod exposure right now and I did not audit them this session. This
   is the biggest miss: the same split-brain may be live for those today.
2. **Root cause, not just symptom.** I fixed the pin but the deeper question —
   *what process* bumps `go.mod` without syncing the flake rev — is only
   partially answered (see "Root cause" under (d)). Without a guard, the next
   in-house dep bump can reopen the same hole.
3. **Harvest + lesson capture.** I did not add the pin-sync task to
   `TODO_LIST.md` (this report's section (f) would otherwise die entombed), and
   I did not record the cross-project lesson
   ("go.mod version bumps must update flake input revs") in the crush-config
   `references/lessons.md`, although the pattern generalizes to every
   multi-repo flake project in the fleet.
4. **Stale-memory trap.** AGENTS.md said "go-retry v0.6.0" and my first reply
   nearly asserted it as fact; only the `go.mod` grep corrected it. The AGENTS.md
   entry was wrong for ten days — memory hygiene caught it by luck of the
   question asked, not by process.

### What I could have done better

1. **Cheaper verification loop.** After the first `nix flake check` FOD failure,
   I ran `buildflow -s nix-hash-fix --fix` (which tripped its findings gate in a
   confusing way), then a second full detect run (5m50s) to confirm. A `git diff`
   of `vendorHash.nix` against the FOD's `got:` hash would have confirmed the
   repair in seconds. Roughly six minutes burned on confirmation theater.
2. **Read the 09-27 commit before finishing.** The root-cause evidence
   (`949f629` = the gomod-restore session) was one `git show` away; I initially
   framed the cause as "unknown" and only dug when writing this report.
3. **Skill scoping transparency.** I loaded `library-deep-dive` (the trigger
   matched) but deliberately executed only Phase 1 and skipped its HTML
   deep-dive deliverable — correct for a binary in-house question, but the
   deviation should have been stated up front, not discovered by omission.

### What could still improve (process-level)

1. **Automated rev↔go.mod sync guard** so this class of drift is impossible to
   miss (see (e) and (f)).
2. **A "binary question fast path"** in the library-deep-dive skill: when the
   question is "do we use X" and X is in-house, Phase 1 + a verdict is the right
   scope; the full research audit is overkill.
3. **AGENTS.md dependency entries should carry a "last verified" date** so a
   ten-day-stale version claim is visibly stale instead of authoritative.

---

## a) FULLY DONE

| # | Item | Evidence |
| - | ---- | -------- |
| 1 | go-retry usage audit answered with evidence | `go.mod:13` (v0.7.1); import `client.go:14`; `retry.DoWithValue` `client.go:329`; `isRetryableError` `client.go:139`; `DelayFunc` (Retry-After honored, capped) `client.go:107`; ADR 003 `docs/adr/003-retry-executor-go-retry-override.md`; flake input `flake.nix:32-35`, module override `flake.nix:112` |
| 2 | Drift discovered and root-caused | `go.mod` v0.6.0→v0.7.1 in `949f629` (2026-09-27, the gomod-restore session; 5 files incl. a `...gomod-restore-status.md` doc); flake pin `61058487` proven to be the v0.6.0 release commit (git log/tag in `/home/lars/projects/go-retry`) |
| 3 | Flake pin aligned to v0.7.1 tag | `flake.nix:33` → `a2d063a421f368c9e293f50f256e6ba0035da8b1` (= v0.7.1); `flake.lock` refreshed via `nix flake update go-retry`; committed by daemon `3ed4fb4` (contents verified: flake.nix, flake.lock, AGENTS.md) |
| 4 | vendorHash repaired by the owned tool, not by hand | `buildflow -s nix-hash-fix --fix` → `vendorHash.nix` = `sha256-PHGvRfUbcuanO9pQMVjAV2/LvF0d+/ouooZe17kq7ZQ=`, byte-matching the FOD's `got:` hash |
| 5 | Verification green on v0.7.1 | nix-hash-fix diagnose: 9/9 targets built incl. sandboxed `wise-go-test` derivation; final `nix flake check` **EXIT=0**, all 7 checks pass (build, format, links, md-go-snippets, pre-commit, test, treefmt) |
| 6 | AGENTS.md corrected | go-retry entry v0.6.0→v0.7.1 + "keep flake rev in sync with the go.mod tag" rule + drift history (committed in `3ed4fb4`) |
| 7 | Local go-retry HEAD triaged | `v0.7.1-19-g09bbfce`; non-chore delta = only "Restore go directive to 1.26" → the tag remains the correct pin, no bump warranted |
| 8 | "running 0 flake checks" oddity investigated | nix counts *unbuilt* derivations on re-run; verbose run explicitly checked `checks.x86_64-linux.test`; exit 0. Benign, undocumented quirk |

## b) PARTIALLY DONE

| # | Item | Works now | Remains open | Blocker | Effort |
| - | ---- | --------- | ------------ | ------- | ------ |
| 1 | In-house pin-sync audit | go-retry audited + fixed | go-branded-id, go-error-family, go-nix-helpers pins NOT compared to their go.mod/tag versions | none — out of this session's trigger scope | S |
| 2 | `vendorHash.nix` commit | content verified correct + green build | still uncommitted at session end (daemon expected to sweep) | none | S |
| 3 | library-deep-dive skill execution | Phase 1 usage discovery done | Phases 2–7 (capability research, HTML report) deliberately skipped | deviation, not blocker — in-house lib, binary question | M if wanted |
| 4 | Drift guard | fully specified idea (see (e)1) | zero implementation | not started | S–M in-repo, M–L upstream |

## c) NOT STARTED

| # | Item | Why not started | Still wanted? |
| - | ---- | --------------- | ------------- |
| 1 | Automated pin-sync guard (repo flake check attr or buildflow step) | discovered this session; needs the (g) policy answer | yes — top priority |
| 2 | Cross-project lesson in crush-config `references/lessons.md` | needs a commit in the crush-config repo (out of repo scope) | yes |
| 3 | HARVEST of section (f) into `TODO_LIST.md` | report written last; harvest is the immediate next step | yes |
| 4 | go-retry upstream: Retry-After-aware recipe/example (archived ADR-003 TODO #33) | inherited open item, untouched | demand-gated |
| 5 | go-retry upstream: record capped-honoring design decision (archived TODO #34) | inherited, untouched | demand-gated |
| 6 | Dependabot pickup check for go-retry bumps (archived TODO #31) | inherited, untouched | low |

## d) TOTALLY FUCKED UP

**1. The verification split-brain (found, fixed, but the hole class remains open).**

- **What was broken:** from 2026-09-27 (`949f629`) to 2026-10-07, `go test`
  executed against go-retry **v0.7.1** while the release-gating sandboxed
  `nix flake check` built against go-retry **v0.6.0**. Two gates, two versions,
  ten days, nobody noticed — the pre-release chain reported green throughout.
- **Severity:** silent. No user-facing bug this time (v0.6.0→v0.7.1 added only
  options + docs, and the used API surface is identical), but the same drift on
  a behavioral release would ship code the sandbox never tested.
- **Root cause:** the 2026-09-27 gomod-restore session (recovery from the Go
  1.27 sweep incident) re-added `go.mod` deps at latest versions without
  syncing the flake input revs. The flake rev is pinned explicitly in
  `flake.nix`, so `nix flake update` can never heal it — only a human/flake.nix
  edit can.
- **Mitigation now:** pin aligned + AGENTS.md rule recorded. **Not mitigated
  structurally:** no automated guard exists; any future go.mod bump (buildflow
  update, daemon, manual) can recreate this silently.
- **Residual risk rating:** the pattern is unfixed even though this instance
  is fixed.

**2. Nothing else.** No test failures, no data loss, no reverts. The expected
vendorHash FOD on the first check is normal churn after an input change, not a
fuckup. The one process blemish (six minutes of redundant verification) is in
the self-critique, not here — it cost time, nothing else.

## e) WHAT WE SHOULD IMPROVE

1. **Automate pin-sync.** Concrete: add a flake check attr (`checks.pin-sync`)
   that, for each `github:LarsArtmann/*` input with a `rev=`, asserts the rev
   equals a tag in that repo AND that the tag's module version satisfies
   `go.mod`. Every future drift becomes a red check instead of a lucky question.
2. **Make dep-bump processes flake-aware.** Whatever bumped go.mod on 09-27
   (recovery session, buildflow update, daemon) must either update flake revs or
   emit a loud "flake pin now stale" warning. Candidate: a buildflow upstream
   step, or a line in the gomod-check step.
3. **Faster input-pin iteration loop.** Use targeted
   `nix build .#checks.x86_64-linux.test` while iterating on pins; run the full
   `nix flake check` once at the end. (This session: full check, FOD, fix, full
   check again.)
4. **Verify repairs by diff before re-running detect.** After any
   `buildflow -s nix-hash-fix --fix`, compare the written hash to the FOD's
   `got:` value via `git diff` before spending another detect cycle.
5. **Document the nix quirk:** `nix flake check` prints "running N flake
   checks" where N = derivations needing building; "0" on a cached re-run reads
   like "checked nothing". Worth one AGENTS.md line to save the next session's
   investigation.
6. **"Last verified" dates on AGENTS.md dependency entries** so stale claims
   (like v0.6.0 living ten days past its truth) are visibly stale.
7. **Library-deep-dive fast path:** record (in the skill or as a lesson) that
   in-house "do we use X" questions resolve at Phase 1; the HTML deep-dive is
   for third-party utilization audits.

## f) Up to 50 things we should get done next

> Ranked by impact; this is a brainstorm per the skill (user asked for up to
> 50, not a commitment list). Items 1–12 are the real queue; 13+ are ROADMAP
> fuel. HARVEST should pull the actionable ones into `TODO_LIST.md`.

| # | Task | Impact | Effort | Category |
|---|------|--------|--------|----------|
| 1 | Audit go-error-family flake pin vs go.mod (v0.10.0) for the same drift | High | S | Cleanup |
| 2 | Audit go-branded-id flake pin vs go.mod (v0.5.1) for the same drift | High | S | Cleanup |
| 3 | Audit go-nix-helpers input currency (not a go.mod dep — check rev freshness vs its releases) | High | S | Cleanup |
| 4 | Commit `vendorHash.nix` (or verify the daemon swept it) | High | S | Cleanup |
| 5 | Add `checks.pin-sync` flake check attr asserting rev==tag and version==go.mod for all in-house inputs | High | M | Quality |
| 6 | Re-run the full pre-release chain end-to-end after the pin change | High | M | Quality |
| 7 | HARVEST this report's actionable items into `TODO_LIST.md` | High | S | Documentation |
| 8 | Identify which process bumped go.mod deps on 2026-09-27 and decide how it must sync flake revs going forward | High | S | Quality |
| 9 | Record the cross-project lesson (go.mod bumps must update flake revs) in crush-config `references/lessons.md` | Medium | S | Documentation |
| 10 | Propose the pin-sync guard upstream to BuildFlow (fleet-wide value: every flake+go repo) | Medium | M | Feature |
| 11 | When CI is re-enabled: add the pin-sync check as a ci.yml job | Medium | S | Feature |
| 12 | Direct `go test -race ./...` sanity run to re-confirm the non-sandbox baseline (go.mod unchanged, cheap insurance) | Medium | S | Quality |
| 13 | Document the "0 flake checks" nix message quirk in AGENTS.md | Low | S | Documentation |
| 14 | Add "last verified" dates to AGENTS.md dependency entries (go-branded-id, go-error-family, go-retry) | Medium | S | Documentation |
| 15 | Ask go-retry upstream for a Retry-After-aware recipe/example (archived ADR-003 TODO #33) | Low | M | Feature |
| 16 | Record the capped-honoring design decision in go-retry's docs (archived TODO #34) | Low | S | Documentation |
| 17 | Check whether the Dependabot config picks up go-retry bumps (archived TODO #31) | Low | S | Quality |
| 18 | Sweep the other fleet repos (md-go, go-output, dynamic-markdown-site, …) for the same rev↔go.mod drift pattern | Medium | M | Cleanup |
| 19 | Decide the pin policy question in (g)2 (tag-equality vs version-match) | Medium | S | Decision |
| 20 | Add a tiny script/one-liner doc: `git ls-remote --tags` rev-vs-tag verification for all inputs | Low | S | Cleanup |
| 21 | Consider making nix-hash-fix detect run before fix to avoid gate-error confusion (buildflow UX feedback) | Low | S | Quality |
| 22 | Verify the sandbox-live workflow still passes post-pin (hits real API; unaffected, but cheap) | Low | S | Quality |
| 23 | Annotate the 2026-10-07_07-09 status doc if any of its "pre-verify in-house deps" items advanced | Low | S | Documentation |
| 24 | Review whether buildflow's gomod-check could warn when a required module version is flake-pinned at a different rev | Medium | M | Quality |
| 25 | Clean up go-retry's local branch: 19 auto-commits since v0.7.1 — tag or squash discussion belongs in that repo | Low | S | Cleanup |
| 26 | Extract the drift story into a short AGENTS.md gotcha cross-link from the flake.nix comment | Low | S | Documentation |
| 27 | Consider a `nix flake check --all-systems` x86_64-darwin decision (pre-existing blocker: nixpkgs-26.11 dropped it) | Low | S | Decision |
| 28 | go-retry upstream: cut v0.7.2/v0.8.0 if the go-directive restore in HEAD matters to consumers | Low | S | Release |
| 29 | Add the pin-sync invariant to CONTRIBUTING.md so humans bumping deps see it | Medium | S | Documentation |
| 30 | Evaluate whether `nix flake update` should be forbidden when go.mod and flake revs disagree (pre-commit guard) | Medium | M | Quality |
| 31 | Backfill: confirm the 09-27 session's other go.mod changes (5 files) are all pin-consistent now | Medium | S | Cleanup |
| 32 | Check md-go-validator v1.3.0 input is still current vs its releases | Low | S | Cleanup |
| 33 | File the "0 flake checks" message confusion upstream to nix (optional; low value) | Low | S | Cleanup |
| 34 | Add a session-start check: diff every `github:` input rev against `git ls-remote` tags once per week (cron/daemon) | Low | M | Quality |
| 35 | AGENTS.md: note that flake-pinned revs are IMMUTABLE under `nix flake update` (only flake.nix edits move them) | Medium | S | Documentation |

## g) Questions I can NOT figure out myself

1. **Was the 2026-09-27 go.mod dependency bump (`949f629`, the gomod-restore
   session) an intentional choice of latest versions, or a side effect you
   did not notice?** I tried: `git show` on the commit, the session's status
   doc name, and AGENTS.md incident notes — they show *what* happened but not
   *intent*. The answer decides whether the fix is "add a guard" (accident) or
   "also change the recovery runbook" (policy).
2. **What is the pin policy: must every `github:LarsArtmann/*` flake input rev
   equal a release tag commit, or is any rev acceptable while the go.mod
   version matches?** I tried: CONTRIBUTING.md, AGENTS.md, the archived ADR-003
   status — no policy recorded. This decides what the `checks.pin-sync` guard
   asserts (tag-equality is strict; version-match is looser but still catches
   this drift).
3. **Should the permanent guard live upstream in BuildFlow (fleet-wide step,
   new provider) or repo-local (flake check attr + CI job)?** I tried: the
   buildflow step catalog and `.buildflow.yml` — no existing step covers it.
   Fleet scope is your call; it determines whether I file a BuildFlow feature
   or just implement it here.

---

*Report ends. Next step per the status-report skill: HARVEST section (f) into
`TODO_LIST.md`, then wait for instructions.*
