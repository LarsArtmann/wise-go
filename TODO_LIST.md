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

[ ] Typed recipient `Details` — typed per-corridor structs vs `map[string]string`
key constants. Carried unanswered through seven status reports (2026-08-19_17-14
g.2 through 2026-10-05_14-53 g2 — all in `docs/status/`). The v1.0 audit confirms
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
**BLOCKED: needs the user's approval to push and enable.**

[ ] GOEXPERIMENT ergonomics — pin direnv/home-manager setup so `jsonv2` is set
without relying on `.buildflow.yml` env injection (user-machine work; the
workaround is documented in CONTRIBUTING.md). Carried since
`docs/status/archived/2026-07-23_03-49_buildflow-env-fix-and-golangci-restore.md`.
**BLOCKED: user-machine change.**

## P3 — Quality & tooling (unblocked)

### Open

[x] File the four evidenced tool-repo defects — FILED 2026-10-07: BuildFlow
[#34](https://github.com/LarsArtmann/BuildFlow/issues/34) (embedded `erraudit`
false positives), [#35](https://github.com/LarsArtmann/BuildFlow/issues/35)
(`reinstall` not switching the profile),
[#36](https://github.com/LarsArtmann/BuildFlow/issues/36) (`file-size-check`
scanning 0 files in root-package layouts); md-go-validator
[#8](https://github.com/LarsArtmann/md-go-validator/issues/8) (own flake
cannot build its package). Drafts (voice-checked) kept in `docs/drafts/`.
WATCH: when #34 fixes, re-run the embedded step vs standalone byte-for-byte
and lift the `skip_steps`; when #8 fixes, switch the flake input to upstream
flake and drop the `go_1_27` override + vendorHash.
Source: `docs/status/2026-10-07_04-20_dual-phase-repair-and-tool-reevaluation-status.md` §b1.

[ ] `nix flake check --all-systems` — verified 2026-10-07: aarch64-linux and
aarch64-darwin evaluate clean; x86_64-darwin FAILS EXTERNALLY (nixpkgs-26.11
dropped that system, not a flake bug). Decide: pin a nixpkgs-26.05-darwin
input for darwin checks, or drop x86_64-darwin from the check matrix.
**BLOCKED: user decision.**

[ ] Go 1.27 migration — plan written at
`docs/planning/2026-10-07_go-1.27-migration-plan.md` (4 unlock conditions,
single-commit flip/rollback). Execution is blocked on those conditions
(nixpkgs stable 1.27, in-house deps verified on 1.27, erraudit local install,
buildflow skip_steps removal). Source:
`docs/status/2026-10-07_03-45_buildflow-red-to-green-repair-status.md` §g1.

### Closed 2026-10-07 (P3 sweep — `docs/status/2026-10-07_06-07_p3-quality-tooling-sweep-status.md`)

[x] Wire `md-go-validator` into `nix flake check` — done as
`checks.md-go-snippets`, consuming the tool as a source-only flake input
(`v1.3.0`, `flake = false`) because upstream's flake cannot build it (see open
defect-filing item); scans README + `docs/**`.

[x] Root-cause the `golangci_lint_ls` phantom on `webhooks.go:142` — RESOLVED
2026-10-07: the LSP servers run outside the nix devShell without
`GOEXPERIMENT=jsonv2`, so the json/v2-dependent package failed to load and
gopls/golangci emitted phantoms from degraded analysis. Fixed via project
`.crushrc` (`--env GOEXPERIMENT jsonv2` on gopls + golangci_lint_ls;
activates at session start) + poisoned golangci analysis cache cleared.
Recorded in AGENTS.md; reopen only if phantoms recur after 2026-10-07.
Source: `docs/status/2026-10-06_14-16_q4-flip-execution-and-gates-status.md` §d4.

[x] Extend `doc-verify` count-claims to `docs/releases/*.md` — done; the
latest `*-release-notes.md` client-method count is now gate-checked.

[x] Build coverage headroom — done: 93.1% total against the 90.0 floor
(was 90.1%); added brand-name, corruption-path, validate-path, and
unreadable-body tests.

[x] Write the quarterly-surface rollover ritual into CONTRIBUTING — done
("Quarterly API-surface rollover ritual" section: probe → flip → straggler
grep → changelog → verify, with the OTT probe-blind caveat).

[x] `art-dupl` enforced-gate decision — RESOLVED: accept the suppressed
baseline as policy, NO enforced gate; revisit at CI re-enable as a warn-first
job. Recorded in AGENTS.md (art-dupl bullet).

[x] erraudit class-wide policy — RESOLVED: embedded-analyzer findings are
ignored AS A CLASS via `skip_steps`; the canonical standalone gate governs;
never per-finding. Recorded in AGENTS.md (erraudit bullet).

[x] Add a CAMT (`.xml`) `GetStatement` test — done in `wise_test.go`
(`application/xml`, camt.053.001.02 raw bytes asserted).

[x] Commit a `benchstat` baseline — done: `docs/bench/2026-10-07_v0120_baseline.txt`
(6 benchmarks × count=6).

[x] Fuzz + bench webhook funcs — done: `FuzzVerifyWebhookSignature`,
`FuzzParseWebhookPublicKey` (8s live fuzz clean), `BenchmarkVerifyWebhookSignature`
(~22.8µs/op) in `internal_test.go`/`bench_test.go`.

[x] Add a `nix run .#pre-release` flake app — done: dirty-tree guard → build →
vet → race → golangci-lint → `nix flake check` → doc-verify → release-notes-check
→ apidiff.

[x] Release-notes check for split code spans and relative links — done as
`nix run .#release-notes-check` (odd-backtick-outside-fence + repo-relative
link detectors, both negative-tested).

[x] Make the spec-conformance coverage guard shuffle-proof — done: floors now
assert in `TestMain` post-run, gated by `conformanceSuiteCompleted`; statement
floor raised 3 → 4 (actual 6), exchange floor hit at 180 recorded.

[x] Add a go-retry `Retry-After`-honored log line — done: `RequestLog.RetryAfterDelay`
delay-decision entry (Status=429, honored wait as Duration), pinned by a
Ginkgo Context asserting the 3-entry sequence.

[x] Fuzz + bench `decodeExchangeRates` — done: `FuzzDecodeExchangeRates`
(8s live fuzz clean) + `BenchmarkDecodeExchangeRates` (~1.2–1.7µs/op, 4 allocs).

## Harvested-and-closed pointer

Everything this list used to carry — including the v0.12.0 cycle's webhook
subscription CRUD, typed event decoding, OTT endpoints, `WithUserAgent`,
`Profile.UserID`/`PublicID`, `TestRequireID`, `fetchByID` routing, exhaustruct_v5,
govulncheck zero-findings, benchmarks/fuzz/raw round-trips, test-file split,
concurrency test, coverage gate, templates, ADRs, `doc-verify`/`apidiff` apps,
CI tool pinning, the `erraudit` CI gate, the 90% coverage floor, the Q4 webhook
rollover, the go-retry v0.6.0 executor, the v0.10.0/v0.11.0/v0.12.0 GitHub
Release objects (published 2026-10-07), and the declined ideas — is recorded in
[CHANGELOG.md](CHANGELOG.md) (`[0.12.0]` and the v0.10.0/v0.11.0 / 2026-09-13
sections) and annotated inline in the corresponding `docs/status/` reports.
