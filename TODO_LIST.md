# TODO List

Short- and mid-term actionable work for wise-go. Each item is bounded, has clear
ownership, and can be completed in one sitting. For long-term vision and raw ideas
see [ROADMAP.md](ROADMAP.md). For shipped features see [FEATURES.md](FEATURES.md).
Completed work lives in [CHANGELOG.md](CHANGELOG.md), never here.

## P2 — User-gated

[ ] Push bank-sync's v0.16.0 pin bump — committed locally 2026-10-17 (go.mod +
flake input/lock + vendorHash dance + SCA mock paths flipped Q3→Q4 + sentinel-sweep
verdict in its AGENTS/07); the parallel `internal/server` refactor landing green
is the other input. **User-gated: pushes to other repos need the explicit word.**

[ ] Sandbox assumptions 1+3 — one capped (~30 min) browser SCA attempt
(existing `/tmp` tooling; DOM probe with a 20s timeout first, sandbox 2FA
catch-all `111111`) would close the COMPACT-vs-FLAT content diff and the
statement `details.type` census; otherwise bank the PARTIAL/BLOCKED verdicts
permanently and delete the browser tooling. **User-gated (Q1 of the
2026-10-17 self-review).**

[ ] Lock the public API at v1.0 — audit is green (re-audit + growth lineage in
`docs/reviews/2026-08-21_v1.0-api-audit.md`, growth lineage maintained through
the current 46-method surface); remaining: tag `v1.0.0`.
User decision 2026-10-17: hold v1.0.0 — v0.16.0 (breaking allowed) ships the
sandbox lane instead; revisit the tag once the v1 blockers in ROADMAP clear.
**BLOCKED: needs the user's explicit approval (tagging is irreversible).**

[ ] `RequestLog.RetryAfterDelay` API freeze — **PARKED FOR v1.0** (owner
decision 2026-10-17: breaking changes are allowed in v0.x, so freezing early
buys nothing). When v1.0 planning starts: lock the delay-decision-entry shape
(entry riding the `LogRequest` channel, `Method`/`URL` empty, honored wait as a
`time.Duration`), or grow a dedicated observer interface. The behavior is
pinned by the 429 BDD tests either way.

[ ] Typed recipient `Details` — typed per-corridor structs vs `map[string]string`
key constants. Carried unanswered through seven status reports (2026-08-19_17-14
g.2 through 2026-10-05_14-53 g2 — all under `docs/status/`, several now in `archived/`). The v1.0 audit confirms
the map is the only shape consumers depend on today, so v1.0 can freeze the map
and add typed accessors later.
**BLOCKED: needs the user's design decision.**

[ ] Set the `CACHIX_AUTH_TOKEN` secret (and confirm the `larsartmann` cache
exists) — the CI cachix step (pinned to verified v15 commit `ad2ddac`) is
`continue-on-error: true` until then.
**BLOCKED: needs the user's cachix token.**

[ ] Set the `ERRAUDIT_TOKEN` secret (PAT with read access to the private
`github.com/larsartmann/erraudit` repo — the module is not on the public proxy)
so the erraudit CI job can install the tool; until then that job's install step
fails, the gate is skipped, and the job stays green (`continue-on-error` pattern,
2026-10-05). When the token lands, delete the two `continue-on-error` / `if`
lines in the `erraudit` job to make the gate blocking.
**BLOCKED: needs the user's PAT.**

[ ] Re-enable CI on GitHub — the auth blocker is RESOLVED (2026-09-13): all
flake inputs moved from `git+ssh` to public `github:` URLs pinned by rev, so CI
needs no secrets; the workflow file is refreshed (8 jobs incl. the network
`pin-sync` job, 92.0 coverage gate single-sourced, tool pins).
Remaining: `gh workflow enable ci` (🟡 user word) and the first-run watch —
triage runbook + badge checklist + art-dupl promotion criteria are written in
`docs/planning/2026-10-17_ci-first-run-watch-package.md`.
Until then the coverage badge stays frozen at its last CI-measured value.
**BLOCKED: needs the user's approval to enable.**

## P3 — Quality & tooling (unblocked)

### Open

[ ] Security window: re-run `nix flake update nixpkgs` when go **1.26.9**
lands in nixpkgs-unstable (locked nixpkgs `55bfb4d` still ships **1.26.8**,
re-probed 2026-10-17 on the v0.15.0 train — window NOT open) — clears
all 9 govulncheck stdlib vulns (GO-2026-6603..6617, fixed in 1.26.9) and
shrinks the vulnix advisory set. The update procedure is pre-staged as a
click-through checklist:
`docs/planning/2026-10-17_security-window-checklist.md` (vendor FOD →
flake check → full buildflow → re-measure govulncheck/vulnix). Then
track the BuildFlow #66/#67 and
branching-flow #15/#16 consumer-side cleanups. Source:
`docs/status/2026-10-17_01-43_buildflow-warning-surface-zero-and-gate-repair-status.md` §b1/§f.

[ ] Continue the AGENTS.md prune — part 1 done 2026-10-17 (v0.15.0 train M15):
57,097 → 47,354 bytes with every concept carried (narratives compressed to
their evidence pointers; one erraudit split-brain fixed). Remaining: the
<30 KB target needs a second pass over the remaining mid-size bullets;
carry-every-concept rule stays. Source: 21:45 plan M24/T87.

[ ] Consume the committed `benchstat` baseline
(`docs/bench/2026-10-17_v0150_baseline.txt`) in the next perf-touching PR.
Verdict 2026-10-17 (M25): NO benchmarked-path change shipped post-v0.15.0
(Quote mapper + the money guard are not in the benchmark set), so there was
nothing to compare — the convention is armed for the first real one. The
capture-hygiene rule (idle machine, capture at the tag) is written in
`docs/bench/README.md`; a polluted parallel-load capture showed +73% geomean
with identical B/op and was discarded. benchstat is NOT in the devShell —
install with `go install golang.org/x/perf/cmd/benchstat@latest`.

### Decisions (P3)

[ ] `Health` code-half — decide: keep
`Health` as the retried readiness probe as-is, give it a fail-fast path
(`MaxAttempts=1`-style, health probes should not spend ~2 min in the retry
executor), and/or add a `VerifyCredentials` alias if any consumer depends on
the current retried semantics; consider a `HealthCheck(ctx)` name-alias so
consumer `do.HealthcheckerWithContext` containers satisfy the interface
directly. A true cheap liveness probe additionally needs
an endpoint choice verified against the live API (no ping/health endpoint is
documented in the vendored snapshot). SHIPPED 2026-10-16: the docs-half
(godoc + README + FEATURES wording), the `Describe("Health")` typed-error
contract tests (429 → `*RateLimitError`, bad key → `*AuthError`, retry-
exhausted 5xx → `*ServerError`), and the consumer-side DI recipe (README
"Dependency Injection & Health Checks", compile-verified vs v0.13.0 +
master). Remaining is the code-half ONLY.
**BLOCKED: user design decision.** Recommendation + options written:
`docs/planning/2026-10-16_decision-memos.md` D2 (suggests additive
`VerifyCredentials` in v1.1, no `Health` behavior change). Source: 2026-10-16 samber/do + health
architecture review, finding F-1; `docs/architecture-understanding/2026-10-16_19-59_samber-do-health-architecture-review.md`.

[ ] Spec-drift gate posture — decide whether `--fail-on-drift` should be
wired into `nix run .#pre-release` (releases then require a fresh vendored
spec; today vendored-snapshot staleness is report-only) and where spec-drift
recurs while CI is disabled (scheduled local run, buildflow step, or
manual-before-release). Source:
`docs/status/2026-10-16_19-48_wise-api-spec-drift-check-session.md` §g2/§g3.
**BLOCKED: user decision.** Recommendation + options written:
`docs/planning/2026-10-16_decision-memos.md` D3 (keep pre-release
report-only; wire `--fail-on-drift` into a scheduled job once CI runs).

[ ] Flake pinning strategy — keep `?rev=<commit-sha>` inputs (the hermetic
`checks.pin-sync` asserts structural invariants only; the network
`apps.pin-sync` is the real version detector, firing at pre-release time) or
switch to version-tag refs (`github:LarsArtmann/go-retry/v0.7.1`, flake.lock
still pinning the immutable commit) so the hermetic check can catch version
drift offline. Reverses the documented rev-pinned decision — user call.
Recommendation + options written: `docs/planning/2026-10-16_decision-memos.md` D4
(keep rev pins; tag pins hide drift behind the lock).
Source: `docs/status/2026-10-07_08-22_v0.13.0-release-pin-sync-drift-discovery-status.md` §g1.
**BLOCKED: user decision.**

[ ] `BUILDFLOW_NO_RESULT_CACHE=1` standing default — make the result-cache
bypass the standing buildflow invocation shape until BuildFlow #67 fixes the
cache scoping (the 01-55 session measured ~0% hit rate, so the cost is
near-zero), or keep it opt-in for parallel-session nights.
**BLOCKED: user decision.** Source:
`docs/status/2026-10-17_01-55_buildflow-warning-zero-full-session-status.md` §g3.

[ ] `nix flake check --all-systems` — verified 2026-10-07: aarch64-linux and
aarch64-darwin evaluate clean; x86_64-darwin FAILS EXTERNALLY (nixpkgs-26.11
dropped that system, not a flake bug). Decide: pin a nixpkgs-26.05-darwin
input for darwin checks, or drop x86_64-darwin from the check matrix.
**BLOCKED: user decision.** Recommendation + options written:
`docs/planning/2026-10-16_decision-memos.md` D5 (drop x86_64-darwin;
reversible one-liner if a consumer appears) — also ask #7 in
`docs/planning/2026-10-16_open-user-asks.md`.

[ ] Go 1.27 toolchain migration — execute
`docs/planning/2026-10-07_go-1.27-migration-plan.md` (4 unlock conditions,
single-commit flip/rollback) when its conditions hold. Latest probe:
nixpkgs lock at go 1.26.8 (2026-10-17) → condition 1 not met. Blocked on:
nixpkgs stable 1.27, in-house deps verified on 1.27, erraudit local install,
buildflow skip_steps removal. Source:
`docs/status/2026-10-07_03-45_buildflow-red-to-green-repair-status.md` §g1.
