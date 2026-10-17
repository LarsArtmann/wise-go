# Status — v0.16.0 shipped, assumption #2 corrected, bank-sync bumped

**Date:** 2026-10-17 ~09:15 CEST · **Session:** post-plan execution (SUPERB plan M1→M14)
**Repo state at close:** master clean, all gates green, tag `v0.16.0` on origin + GitHub prerelease; local coverage of the plan's unblocked items complete.

## What shipped

| Item | Outcome |
| --- | --- |
| **v0.16.0 released** | Pre-release chain green incl. `--shuffle` + apidiff (only the four additive `Quote` fields; gorelease itself suggested v0.16.0). Annotated tag on `8f2cdb7`, pushed with master; GitHub prerelease created; proxy indexed; clean-dir `go get` + consumer build verified. pkg.go.dev lags normally. |
| **Assumption #2 corrected: PARTIAL → VERIFIED** | The 08:00 session recorded "sandbox locks balance creation entirely (403 `non.existing.balance`)". Within hours the lock stopped reproducing: creation round-trips for standard AND savings balances, and an identical `X-idempotence-uuid` replay returns the already-created balance. Caught by the new permanent test initially asserting the lock. Lesson recorded in AGENTS: single-probe 403 verdicts are provisional. |
| **Five permanent key-gated live tests** | `sandbox_live_test.go`: rates (current+historical), funding-404 typing, SCA challenge shape (OTT never leaks into the error string), `GetOTTStatus` on 2026Q4 (statement action, ≤1h validity, pending SIGNATURE challenge, zero alternatives), CreateBalance idempotency (fixed key ⇒ non-mutating replays). Credentialed run 8/8 PASS; keyless SKIP. |
| **Sandbox lane recorded** | CHANGELOG entries (first credentialed run, five verdicts, two wire findings, tests), TODO strikes, AGENTS sandbox-lane gotcha (SIGNATURE shape, body-shape probing, served window 2026Q2..2027Q1). FEATURES row flipped DISABLED → FULLY_FUNCTIONAL; README sandbox section updated. |
| **Probe procedure fixed** | CONTRIBUTING rollover step 1 rewritten: body-shape is the discriminator (app-level `{"code":...}` vs router `{"error":"Not Found"}`), not status; unauthenticated probes are blind. `docs/error-codes.md` gained the envelope-variants section. |
| **bank-sync bumped to v0.16.0** | Straight jump over v0.15.0: go.mod + flake input/lock + vendorHash dance + check-dep-alignment green; wise package tests green after flipping its SCA mocks Q3→Q4; sentinel-sweep verdict recorded in bank-sync AGENTS/07 (keep `errors.AsType` — sentinels match only the granular envelope, switching would narrow router-404 handling). Committed locally; **push stays user-gated**. |
| **Gate battery** | `nix flake check` (incl. the carried flake.lock-nixpkgs-bump question), spec-drift (43 endpoints, zero drift), lint 0, race green, doc-verify, release-notes-check, anchor-check (in flake check). |
| **Daemon audit `93601d1..HEAD`** | Every heuristic commit's file set matches surrounding session work; the only damage was the known sweep `92bab25` (repaired `594a23a`). Config integrity verified: `go 1.26.0`, lint caps 50/10, `.buildflow.yml` 106 lines. |

## Notable events

- **The permanent test falsified a recorded verdict within one hour** — the CreateBalance lock test failed with `err == nil`, live probing showed creation works, and the recorded PARTIAL verdict was corrected to VERIFIED across CHANGELOG/TODO/AGENTS before the tag. This is the lane-recording bundle working as designed.
- **bank-sync commit raced two daemons and a parallel session**: the vendor-hash dance died once on `.git/index.lock`, the pre-commit hook failed once on a leaked `GOTOOLCHAIN=local` (wise-go's direnv), and the daemon mixed one of my files into a commit with the parallel session's `default_selection.go`. Final state verified three-way-consistent (HEAD = index = work: `go 1.27.1`, wise-go v0.16.0). The parallel session's `internal/server` mid-refactor breakage (`defaultSelection` type change) is theirs and untouched.
- **08:29 buildflow findings artifact expired** (surviving /tmp logs belong to another repo); disposition per the standing policy — warning classes pre-dispositioned, no error severity, and today's commits touched no infra paths requiring a full run.

## Open / user-gated

- **Q1**: one capped (~30 min) browser SCA attempt for assumptions 1+3, or bank the partial verdicts (current state).
- **Q3**: crush-config push + home-manager activation (other repo).
- **CI enable** (M6), **RetryAfterDelay freeze** (M5 — moot for v0.x; breaking changes allowed until v1.0), typed `Details` (M16), and the remaining parked decisions (M17–M27) all still await the user's word.
- bank-sync push (bump commit is local only).
