# TODO List

Short- and mid-term actionable work for wise-go. Each item is bounded, has clear
ownership, and can be completed in one sitting. For long-term vision and raw ideas
see [ROADMAP.md](ROADMAP.md). For shipped features see [FEATURES.md](FEATURES.md).
Completed work lives in [CHANGELOG.md](CHANGELOG.md), never here.

## P1 — Release readiness

[ ] Cut v0.15.0 — the structured error surface staged in `[Unreleased]` is
breaking (HTTP 400 now surfaces `*BadRequestError`; `APIError.Details` is new
public surface; 11 code-sentinels). Run `nix run .#pre-release`, cut release
notes, and point bank-sync at the matching recipe in `docs/error-codes.md`
when it bumps its pin (AGENTS.md carries the sweep guidance).

[ ] Run the sandbox live-verification pass (`WISE_SANDBOX_API_KEY`) against the
five spec-vs-live assumptions settled by the conformance work: statement
`type` COMPACT/FLAT, `X-idempotence-uuid` on balance creation, statement
`details.type` enum, `/v1/rates` array shape, text/plain funding errors.
Source: docs/planning/archived/2026-09-16_16-54_pareto-plan-openapi-conformance-green.md, 2026-09-16.

[ ] Flip `ottAPIVersion` 2026Q3 → 2026Q4 once the OTT surface verifies live.
The OTT surface is **probe-blind**: unauthenticated requests to `/one-time-token/*`
404 on every prefix including the known-good 2026Q3, so Q4 cannot be verified
without credentials. Procedure: with `WISE_SANDBOX_API_KEY` (pairs with the sandbox
pass above), probe `GET /2026Q4/one-time-token/status` for a non-404, change
`ottAPIVersion` in `client.go`, run `go test ./...`, commit. Probes 2026-10-05.

## P2 — User-gated

[ ] Add credentialed Wise sandbox integration tests — the workflow
(`.github/workflows/sandbox-live.yml`, active, dispatch-gated) and test skeleton
(`sandbox_live_test.go`) are key-drop-ready; only a first recorded run against
`api.wise-sandbox.com` is missing, plus a CHANGELOG entry once a run succeeds.
Skip path re-verified 2026-10-07 (no key → clean SKIP, suite green).
Carried since `docs/status/archived/2026-08-08_05-15_wise-sandbox-integration-status.md`.
**BLOCKED: needs a sandbox API key (`WISE_SANDBOX_API_KEY`) from the user.**

[ ] Lock the public API at v1.0 — audit is green (re-audit + growth lineage in
`docs/reviews/2026-08-21_v1.0-api-audit.md`, growth lineage maintained through
the current 46-method surface); remaining: tag `v1.0.0`.
**BLOCKED: needs the user's explicit approval (tagging is irreversible).**

[ ] `RequestLog.RetryAfterDelay` API freeze — lock the delay-decision-entry shape
(entry riding the `LogRequest` channel, `Method`/`URL` empty, honored wait as a
`time.Duration`) for v1.0, or grow a dedicated observer interface. Evidence:
`nix run .#apidiff` vs v0.12.0 shows it is the ONLY compatible delta (gorelease
suggests v0.13.0). Source: `docs/status/2026-10-07_07-09_p3-resume-execution-prerelease-green-status.md` §g2.
**BLOCKED: needs the user's design decision.**

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
needs no secrets; the workflow file is refreshed (`GOLANGCI_LINT_VERSION`
v2.13, no-auth nix job, 90% coverage gate). Remaining: push master, re-enable
the workflow server-side (`gh workflow enable ci`), and watch the first run.
Add a network `pin-sync` job to ci.yml at the same time (mirroring the
sandbox-live dispatch pattern).
Until then the coverage badge stays frozen at its last CI-measured value.
Source: `docs/status/archived/2026-10-07_03-57_session-status-unblocked-prep-release-readiness.md`.
**BLOCKED: needs the user's approval to push and enable.**

## P3 — Quality & tooling (unblocked)

### Open

[ ] Security window: re-run `nix flake update nixpkgs` when go **1.26.9**
lands in nixpkgs-unstable (lock pulled to 1.26.8 on 2026-10-17) — clears
all 9 govulncheck stdlib vulns (GO-2026-6603..6617, fixed in 1.26.9) and
shrinks the vulnix advisory set; then re-verify vendor FOD hashes and the
full buildflow run (govulncheck/vulnix are the ONLY remaining
warning-class findings). Track the BuildFlow #66/#67 and
branching-flow #15/#16 consumer-side cleanups. Source:
`docs/status/2026-10-17_01-43_buildflow-warning-surface-zero-and-gate-repair-status.md` §b1/§f.

[ ] Rewrite the `spec_drift.py` extractor on stdlib `go/parser` — kills the
call-shape-bug class (the regex extractor already needed a manual fix for the
`c.request` method-variable indirection). The extraction-floor guard and the
2027Q1 FEATURES triage shipped 2026-10-16 (CHANGELOG). Source:
`docs/status/2026-10-16_19-48_wise-api-spec-drift-check-session.md` §f.

[ ] Continue the AGENTS.md prune — 2026-10-16 pass moved the five worst
incident narratives to `docs/reviews/2026-10-16_agents-incident-narratives.md`
(57,987 → 53,216 bytes, zero constraint loss) but the plan's <30 KB target
needs a dedicated compression pass over the ~40 remaining 400-900 B constraint
bullets; do it bullet-by-bullet with the carry-every-concept rule, not in bulk.
Source: 21:45 plan M24/T87.

[ ] Consume the committed `benchstat` baseline
(`docs/bench/2026-10-16_v0140_baseline.txt`) in the next perf-touching PR.
Note: the README rerun commands are execution-verified (six separate
`-count 1` invocations; Ginkgo bans `-count>1`), and benchstat is NOT in the
devShell — install with `go install golang.org/x/perf/cmd/benchstat@latest`.

[ ] Buildflow/infra micro-batch (01-55 warning-zero session §f): full-diff
`.golangci.yml` against the pre-`ce52551` state (confirm only `lll` + the
exclusions rule were lost); verify + file the md-go-validator nested-fence
extraction defect upstream; comment the bandit/pytest line-fallback
reproductions on BuildFlow #64; delete the migrated
`~/.cache/buildflow/buildflow.db` once a fresh run confirms it is no longer
read; disposition the lingering `interrogate` WARN (install to devShell or
exclude). Source:
`docs/status/2026-10-17_01-55_buildflow-warning-zero-full-session-status.md`.

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
