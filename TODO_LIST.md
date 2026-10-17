# TODO List

Short- and mid-term actionable work for wise-go. Each item is bounded, has clear
ownership, and can be completed in one sitting. For long-term vision and raw ideas
see [ROADMAP.md](ROADMAP.md). For shipped features see [FEATURES.md](FEATURES.md).
Completed work lives in [CHANGELOG.md](CHANGELOG.md), never here.

## P1 — Release readiness

[ ] Bump bank-sync's wise-go pin to v0.15.0 — bank-sync compiles UNMODIFIED
against the v0.15.0 tree (replace-compile verified 2026-10-17); the pin bump
itself is a bank-sync-repo change. Point its error handling at the matching
recipes in `docs/error-codes.md` and sweep its `errors.AsType`-only handling
for the sentinel idioms (AGENTS.md carries the sweep guidance).

[x] ~~Run the sandbox live-verification pass (`WISE_SANDBOX_API_KEY`) against the
five spec-vs-live assumptions settled by the conformance work~~ DONE
2026-10-17: `/v1/rates` VERIFIED, funding-errors VERIFIED SDK-side,
COMPACT/FLAT PARTIAL (server accepts both; content SCA-gated),
`X-idempotence-uuid` PARTIAL (header accepted; sandbox locks balance creation
entirely), `details.type` BLOCKED (same SCA gate). Evidence:
`docs/status/2026-10-17_08-00_sandbox-lane-first-live-pass-ott-q4-flip-and-sca-automation.md`.
Residuals 1+3 parked behind the sandbox SCA web login (single capped attempt
per session).

## P2 — User-gated

[ ] Convert the sandbox lane's evidence into permanent key-gated live tests
(`sandbox_live_test.go`): rates shape (current + historical), `GetOTTStatus`
on 2026Q4 (raise via statement 403), funding-404 typing, SCA-challenge shape,
and the CreateBalance sandbox-lock verdict. First credentialed run recorded
2026-10-17 (see CHANGELOG); keyless skip path re-verified.
Carried since `docs/status/archived/2026-08-08_05-15_wise-sandbox-integration-status.md`.

[ ] Lock the public API at v1.0 — audit is green (re-audit + growth lineage in
`docs/reviews/2026-08-21_v1.0-api-audit.md`, growth lineage maintained through
the current 46-method surface); remaining: tag `v1.0.0`.
User decision 2026-10-17: hold v1.0.0 — v0.16.0 (breaking allowed) ships the
sandbox lane instead; revisit the tag once the v1 blockers in ROADMAP clear.
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

[x] ~~Rewrite the `spec_drift.py` extractor on stdlib `go/parser`~~ DONE
2026-10-17 (v0.15.0 train M16): `scripts/testdata/specdriftext/main.go` (AST
walk incl. the `c.request` method-parameter binding), parity-equal with the
regex extractor at 43 endpoints (`scripts/spec_drift_parity.py`), all 9
negative fixtures green, live spec-drift exit 0.

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

[x] ~~Buildflow/infra micro-batch~~ DONE 2026-10-17 (v0.15.0 train M14):
T56 full-diff confirms only `lll` was lost (no exclusions rule); T57 nested-
fence defect reproduced at source and filed (md-go-validator [#11](https://github.com/LarsArtmann/md-go-validator/issues/11));
T58 bandit/pytest line-fallback reproductions measured mechanically and
commented on BuildFlow [#64](https://github.com/LarsArtmann/BuildFlow/issues/64) (3 issues → 47 lines; 0 tests → 5
lines); T59 legacy `buildflow.db` proven unread (mtime unchanged across a
full run) and deleted; T60 `interrogate` identified (Python docstring-
coverage) and skip_steps'd with rationale.

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
