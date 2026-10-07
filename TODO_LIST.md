# TODO List

Short- and mid-term actionable work for wise-go. Each item is bounded, has clear
ownership, and can be completed in one sitting. For long-term vision and raw ideas
see [ROADMAP.md](ROADMAP.md). For shipped features see [FEATURES.md](FEATURES.md).
Completed work lives in [CHANGELOG.md](CHANGELOG.md), never here.

## P1 — Release readiness

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
`docs/reviews/2026-08-21_v1.0-api-audit.md`, now covering the 41-method
v0.12.0 surface); remaining: tag `v1.0.0`.
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
Until then the coverage badge stays frozen at its last CI-measured value.
Source: `docs/status/archived/2026-10-07_03-57_session-status-unblocked-prep-release-readiness.md`.
**BLOCKED: needs the user's approval to push and enable.**

## P3 — Quality & tooling (unblocked)

### Open

[ ] Wire `md-go-snippets` + `release-notes-check` into `.github/workflows/ci.yml`
when CI re-enables — both are flake-local today and ran green inside
`nix flake check` twice on 2026-10-07.
Source: `docs/status/2026-10-07_07-09_p3-resume-execution-prerelease-green-status.md` §f18.

[ ] Type the webhook label surface — `decodeWebhookEvent` takes `WebhookEventType`
instead of `string`, and the `"get"`/`"refresh"` verbs become typed values
(carried since 2026-09-27; `webhooks.go:360` still takes `eventType string`).
Source: `docs/status/archived/2026-09-27_23-42_dedup-pass2-gomod-flipflop-rootcause-status.md` §b3/§f5.

[ ] Promote the enforced coverage floor 90 → 92 — actual is 93.1%; pair with the
CI re-enable so the gate also runs off-machine.
Source: `docs/status/2026-10-07_06-07_p3-quality-tooling-sweep-status.md` §f13.

[ ] Close the remaining 06-07 test gaps: checked-in `testdata/` seeds for the
three fuzz targets; `FuzzParseWiseDate`; and a `docs/bench/README.md` one-pager.
(The `Retry-After` delay-entry cases are covered — `Retry-After: 0` emits no
entry, and two consecutive 429s emit two, both pinned 2026-10-07.)
Source: `docs/status/2026-10-07_06-07_p3-quality-tooling-sweep-status.md` §f26–27, §f42–45.

[ ] Consume the committed `benchstat` baseline
(`docs/bench/2026-10-07_v0120_baseline.txt`) in the next perf-touching PR.

[ ] De-duplicate the LSP-resolution narrative — the struck ROADMAP raw idea and
the AGENTS.md gotcha should have one canonical home.
Source: `docs/status/2026-10-07_07-09_p3-resume-execution-prerelease-green-status.md` §f20.

### Decisions (P3)

[ ] `nix flake check --all-systems` — verified 2026-10-07: aarch64-linux and
aarch64-darwin evaluate clean; x86_64-darwin FAILS EXTERNALLY (nixpkgs-26.11
dropped that system, not a flake bug). Decide: pin a nixpkgs-26.05-darwin
input for darwin checks, or drop x86_64-darwin from the check matrix.
**BLOCKED: user decision.**

[ ] Go 1.27 migration — plan written at
`docs/planning/2026-10-07_go-1.27-migration-plan.md` (4 unlock conditions,
single-commit flip/rollback). Condition-1 probe 2026-10-07: `nix eval
nixpkgs#go.version` = 1.26.8 → not met. Execution is blocked on those
conditions (nixpkgs stable 1.27, in-house deps verified on 1.27, erraudit local install,
buildflow skip_steps removal). Source:
`docs/status/2026-10-07_03-45_buildflow-red-to-green-repair-status.md` §g1.

## Cross-repo (crush-config, not wise-go)

[ ] Record the go-retry drift lesson ("a `go.mod` version bump must update the
matching flake input rev; pinned revs are immutable under `nix flake update`") in
crush-config `references/lessons.md`, plus the Ginkgo `-count>1` ban and the
`gh issue create --body-file -` silent-failure quirk.
Sources: `docs/status/2026-10-07_07-35_go-retry-pin-drift-fix-status.md` §f9; `docs/status/2026-10-07_07-09_p3-resume-execution-prerelease-green-status.md` §f25.

[ ] Resolve the `check-skill-fanout.sh` absence — global AGENTS.md documents
`bash ~/.config/crush/scripts/check-skill-fanout.sh`, but the path does not exist;
restore the script or fix the doc.
Source: `docs/status/2026-10-07_07-09_p3-resume-execution-prerelease-green-status.md` §b6.
