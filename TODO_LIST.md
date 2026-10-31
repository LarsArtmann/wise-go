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
`docs/reviews/archived/2026-08-21_v1.0-api-audit.md`, growth lineage maintained through
the current 47-method surface); remaining: tag `v1.0.0`.
User decision 2026-10-17: hold v1.0.0 — v0.16.0 (breaking allowed) ships the
sandbox lane instead; revisit the tag once the v1 blockers in ROADMAP clear.
**BLOCKED: needs the user's explicit approval (tagging is irreversible).**

[ ] `RequestLog.RetryAfterDelay` API freeze — **PARKED FOR v1.0** (owner
decision 2026-10-17: breaking changes are allowed in v0.x, so freezing early
buys nothing). When v1.0 planning starts: lock the delay-decision-entry shape
(entry riding the `LogRequest` channel, `Method`/`URL` empty, honored wait as a
`time.Duration`), or grow a dedicated observer interface. The behavior is
pinned by the 429 BDD tests either way.
**Recommendation ready 2026-10-31** (spike memo:
`docs/planning/2026-10-31_12-28_retryafterdelay-freeze-memo.md`): freeze the
entry as-is at v1.x; the observer interface waits for a consumer need.
Remaining: the owner's word at v1.x + the 3-line execution checklist in the
memo.

[x] ~~Typed recipient `Details`~~ resolved 2026-10-31 — spike memo
`docs/planning/2026-10-31_12-28_typed-details-spike.md`: keep the map (the
field vocabulary is server-owned and route-polymorphic; typed structs would
promise what the SDK has no authority over), v1.0 freezes the map, and a
`ValidateDetailsAgainstRequirements` helper (types the missing/extra-field
decision, not the vocabulary) sits in the parked demand-gated set. Census:
only the GBP corridor appears in fixtures; a per-corridor census is
implementation work for when a demand signal exists.

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

[ ] Rebuild the BuildFlow binary + swap the global profile — the green full-run
verdicts ride a stale binary (built at `e547e2b`; BuildFlow repo HEAD is
`952cf3e`, preflight warns). Rebuild requires the manual
`nix profile remove/add` dance (the `reinstall` app is a no-op, BuildFlow
[#35](https://github.com/LarsArtmann/BuildFlow/issues/35)) on the user's
global profile. **BLOCKED: touches the user's global profile.** Source:
`docs/status/2026-10-23_07-48_buildflow-recovery.md` §f4.

[ ] Daemon observability — locate the auto-git daemon (no process/unit/log
findable from inside a session), document a heartbeat/log location, and
record whether the hook bypass (`--no-verify`) is intentional. Until then
"is the daemon alive?" is answerable only from `git log` timestamps.
**BLOCKED: the owner knows where the daemon lives.** Source:
`docs/status/2026-10-23_08-25_continuation-backlog-execution.md` §f4.

[ ] Re-enable CI on GitHub — the auth blocker is RESOLVED (2026-09-13): all
flake inputs moved from `git+ssh` to public `github:` URLs pinned by rev, so CI
needs no secrets; the workflow file is refreshed (8 jobs incl. the network
`pin-sync` job, 92.0 coverage gate single-sourced, tool pins).
**ENABLED 2026-10-31** (`gh workflow enable ci`, blanket-directive default —
disable again with `gh workflow disable ci` if unwanted). The workflow also
gained `workflow_dispatch` + a weekly `spec-drift` scheduled job (D3-b,
`--fail-on-drift`, Mondays 06:00 UTC) — these activate when the local ci.yml
reaches GitHub. **Remaining: the first-run watch** — the REMOTE ci.yml (at
origin/master) has no `workflow_dispatch`, so the first run fires on the next
CODE push (paths-ignore skips md-only); NO push was made from the session
(never-push rule) — the owner's next push or explicit word triggers it.
Triage runbook + badge checklist + art-dupl promotion criteria:
`docs/planning/2026-10-17_ci-first-run-watch-package.md`. The coverage badge
unfreezes on that first run (coverage-badge job, master pushes).

## P3 — Quality & tooling (unblocked)

### Open

[x] ~~Close the daemon-commit format gap~~ closed at the gate level
2026-10-31 — `checks.dprint-format` in `flake.nix`: the four plugin .wasm
files are fetchurl-pinned, a wrapper config rewrites `dprint.json`'s
plugins to store paths, plugins compile into a throwaway `DPRINT_CACHE_DIR`,
and `dprint check` runs hermetically in `nix flake check` (negative probe
built in: the derivation self-fails if the check would accept an unformatted
file). Any daemon-committed unformatted file now fails the next full check.
A daemon-side post-commit hook remains available via the P2 observability
row but is no longer load-bearing. Source:
`docs/status/2026-10-23_08-25_continuation-backlog-execution.md` §f5/f6.

[ ] Reconcile buildflow finding counts against direct tool output — bandit
claimed 78 vs direct 1 (fixed); characterize the inflation
(per-occurrence? scope?) and document or file upstream; same pass: verify
which of buildflow's "9 tools unavailable" are actually missing vs
PATH-scoping. Source:
`docs/status/2026-10-23_08-25_continuation-backlog-execution.md` §f7/f8.

[x] ~~Upstream filings/proposals~~ done 2026-10-31 — (1) nixpkgs/vulnix:
**nothing to file — the premise healed**: NIST serves the
`nvdcve-2.0-modified.json.gz` archive again (HEAD 200, ~14.5 MB, verified
2026-10-31) and vulnix 1.12.5 completes (on-demand run over `./result`:
22 derivations with advisories, all nixpkgs-closure level, 0
SDK-actionable); the skip rationale in `.buildflow.yml` is updated to the
healed state (verify-before-filing did its job — filing today would have
been a false report). (2) crush-config scoped-dprint: **filed as
[LarsArtmann/crush-config#3](https://github.com/LarsArtmann/crush-config/issues/3)**
(propagate the hermetic `checks.dprint-format` pattern into the
report-producing skill flows; voice-checked, draft kept in
`docs/drafts/2026-10-31_crush-config-hermetic-dprint.md`). Sources:
`docs/status/2026-10-23_07-48_buildflow-recovery.md` §f7,
`docs/status/2026-10-23_08-25_continuation-backlog-execution.md` §f19.

[x] ~~Continue the AGENTS.md prune~~ part 3 done 2026-10-31 (full-rewrite pass):
56,812 → 40,634 bytes with every title preserved (69→69, pointer registry
intact — anchor-check + doc-verify green) and every removed clause verified
greppable at its evidence owner (narratives doc, `.buildflow.yml`, dprint
exclusions, bank-sync AGENTS). **The <30 KB target is retired as
unreachable-without-loss**: the remaining largest bullets (~1.0–1.2 KB each —
API-docs study, error-system contract, SCA gotcha, spec-conformance gate,
sandbox lane facts) are normative rule density, not narrative; cutting
further would drop operative rules (Verschlimmbessern guard). History: part 1
2026-10-17 (57,097→47,354), part 2 2026-10-23 (→52,687), part 3 2026-10-31
(→40,634). The remaining size lever is the per-area split decision (row
below). Source: 21:45 plan M24/T87; 2026-10-23 plan T11; 2026-10-31 plan M6.

[x] ~~Extend the anchor gate to `###`-heading pointers~~ done 2026-10-31 —
the shape-3 title registry in `scripts/anchor_check.py` now also registers
`##`-and-deeper markdown headings from the living docs (generic headings —
under two tokens or eight chars — are skipped for prefix safety; registry
keys strip emphasis/backticks), so a quoted `"Heading" section` pointer
resolves and a renamed heading with a dangling pointer fails the gate.
Self-test extended (heading-pass + renamed-heading-drift classes; 7 checked
pointers); census over the living docs: 0 findings — no dangling heading
pointers existed. Remaining prose-blind surface: backtick-form titles
(``the `Gotcha` gotcha``) — deliberate, inline code is wire-string territory.
Source: `docs/status/2026-10-23_16-07_pareto-unblocked-set-execution-title-gate-prune-battery.md` §e3.

[ ] Consume the committed `benchstat` baseline
(`docs/bench/2026-10-17_v0150_baseline.txt`) in the next perf-touching PR.
Verdict 2026-10-17 (M25): NO benchmarked-path change shipped post-v0.15.0
(Quote mapper + the money guard are not in the benchmark set), so there was
nothing to compare — the convention is armed for the first real one. The
capture-hygiene rule (idle machine, capture at the tag) is written in
`docs/bench/README.md`; a polluted parallel-load capture showed +73% geomean
with identical B/op and was discarded. benchstat is NOT in the devShell —
install with `go install golang.org/x/perf/cmd/benchstat@latest`.

[x] ~~`Health` deadline-mitigation BDD test~~ done 2026-10-31 — two
contexts in the `Health` describe: a deadline firing mid-backoff aborts
inside the delay (context deadline exceeded, one request-log entry, &lt;1 s)
and a deadline firing during an honored Retry-After wait emits the delay
decision then stops (2 entries, no post-abort noise). Landed with
`VerifyCredentials`.

[x] ~~Health operability docs bundle~~ done 2026-10-31 — README DI recipe
carries the ctx-timeout bounding snippet (delay aborts mid-backoff because
the retry delay is context-aware); `ExampleClient_Health` +
`ExampleClient_VerifyCredentials` are compile-only godoc examples; `Health`
godoc states the exact default budget (4 attempts x 30 s + up to 3 x 5 s
honored Retry-After = 2 min 15 s).

[x] ~~Harmonize surface-count denominators in AGENTS.md~~ done 2026-10-31 —
convention: **47 public `*Client` methods / 46 Wise operations shipped /
15 resource files** everywhere ("17 resource families" survives only as
ROADMAP's enumerated trigger list; README's `16 resources` claims dropped).
Cross-reference + derivation command live in the AGENTS "API docs study"
bullet; the doc-verify gate pins the method count (FEATURES pattern updated
in the same change — methods and operations diverged at `VerifyCredentials`).
Source: 2026-10-31 status report §c5.

[x] ~~Series index for `docs/architecture-understanding/`~~ done 2026-10-31 —
`docs/architecture-understanding/README.md`: table of all five artifacts with
verdict one-liners, supersession chain (2026-10-31 recheck supersedes
2026-10-16 on current-state claims), leave-alone note for generated
HTML/D2/SVG, and the recheck-cadence trigger. Source: 2026-10-31 status
report §e6.

### Decisions (P3)

[x] ~~`Health` code-half~~ done 2026-10-31 — executed memo D2-b's
recommendation under the 2026-10-31 blanket execute directive: additive
`VerifyCredentials(ctx)` (single authenticated `ListProfiles` round-trip,
retry ladder disabled via a stateless-client shallow copy with
`MaxAttempts = 1`), same typed errors as `Health`, BDD-pinned
(one-server-hit contexts + deadline honoring), godoc cross-links and the
README "which probe when" recipe (scratch-compiled against master; needs
the next release tag for pinned consumers). `HealthCheck` alias declined
—the two-line `do.HealthcheckerWithContext` adapter is the whole cost.
Semantics of `Health` unchanged. Flip: revert `client.go` + tests + docs
rows if the owner rules otherwise.
~~A true cheap liveness probe additionally needs
an endpoint choice verified against the live API (no ping/health endpoint is
documented in the vendored snapshot).~~ still true: no ping endpoint in the
spec; liveness remains out of scope. SHIPPED 2026-10-16: the docs-half
(godoc + README + FEATURES wording), the `Describe("Health")` typed-error
contract tests (429 → `*RateLimitError`, bad key → `*AuthError`, retry-
exhausted 5xx → `*ServerError`), and the consumer-side DI recipe (README
"Dependency Injection & Health Checks", compile-verified vs v0.13.0 +
master).

[ ] Spec-drift gate posture — decide whether `--fail-on-drift` should be
wired into `nix run .#pre-release` (releases then require a fresh vendored
spec; today vendored-snapshot staleness is report-only) and where spec-drift
recurs while CI is disabled (scheduled local run, buildflow step, or
manual-before-release). Source:
`docs/status/2026-10-16_19-48_wise-api-spec-drift-check-session.md` §g2/§g3.
**Standing default applied 2026-10-31 (memo D3-b, blanket directive):**
pre-release stays report-only; the snapshot was re-vendored 2026-10-31
(`--fail-on-drift` exit 0 verified); the scheduled report-only CI job is the
remaining wire-up — lands with the CI re-enable (row in P2). Flip = wire
`--fail-on-drift` into pre-release instead.

[x] ~~Flake pinning strategy~~ resolved 2026-10-31 — (a) kept: rev pins +
pin-sync gates stay the invariant (memo D4 annotated; the recommendation
describes the standing state, no flake change needed).

[x] ~~`BUILDFLOW_NO_RESULT_CACHE=1` standing default~~ resolved 2026-10-31 —
(a) applied (blanket-directive default, reversible): every buildflow run in
this repo uses the bypass invocation shape until BuildFlow #67 lands
(documented in the AGENTS result-cache gotcha; memo D6 annotated). Source:
`docs/status/2026-10-17_01-55_buildflow-warning-zero-full-session-status.md` §g3.

[x] ~~`nix flake check --all-systems`~~ resolved 2026-10-31 — (b) applied:
x86_64-darwin dropped from the `flake.nix` `systems` list (reversible
one-liner; memo D5 annotated; AGENTS all-systems gotcha updated to shipped
state).

[x] ~~Vulnix replacement~~ resolved 2026-10-31 — **no replacement needed:
NIST healed the feed** (archive HEAD 200 verified; vulnix completes). The
step stays skip'd (unmaintained tool + closure-noise rationales, updated
in `.buildflow.yml`), and non-Go closure scanning is available ON DEMAND:
`nix develop -c vulnix ./result` (run 2026-10-31: 22 advisories, all
nixpkgs-closure level, 0 SDK-actionable — e.g. binutils CVEs nothing in
this repo can fix). grype 0.120.0 is in nixpkgs if a maintained scanner is
ever wanted; the third-party NVD mirror stays declined. Re-enable trigger:
a triage-worthy advisory class. Source:
`docs/status/2026-10-23_07-48_buildflow-recovery.md` §g1.

[x] ~~Stdlib-advisory go-directive policy~~ resolved 2026-10-31 (standing
posture applied, blanket directive): the go-mod-normalize policy wins — the
directive stays `1.26.0`, the 48 govulncheck findings / 10 advisories stay
warning-class until the Go 1.27 migration (M25/TODO row below retires
them). Clean findings output was never worth a policy violation that
go-mod-normalize would revert anyway. Flip = the owner explicitly ordering
a patch pin. Source: `docs/status/2026-10-23_07-48_buildflow-recovery.md` §g2.

[ ] AGENTS.md structural split — keep the single file (with taxonomy/index,
see the split-brain sweep's systemic analysis) or split into per-area files
loaded via `context_paths`. The 2026-10-23 split-brain recurrence is a
direct symptom of the append-to-one-file pattern. Part-3 rewrite landed
2026-10-31 (40.6 KB — see the prune row above); the single-file taxonomy
(sections index + placement audit re-verified during the rewrite) is the
standing shape. Remaining open: ONLY the per-area-split verdict —
recommended default: keep single-file (the write-time memory discipline
entry + the title-registry pointer gate address the split-brain class; a
split multiplies sync surface for 40 KB of content). **BLOCKED: user
memory-architecture call.** Source:
`docs/status/2026-10-23_07-22_agents-md-split-brain-sweep.md` §g1/§f15.

[ ] Go 1.27 toolchain migration — execute
`docs/planning/2026-10-07_go-1.27-migration-plan.md` (4 unlock conditions,
single-commit flip/rollback) when its conditions hold. Latest probe:
nixpkgs lock at go 1.26.8 (2026-10-17) → condition 1 not met. Blocked on:
nixpkgs stable 1.27, in-house deps verified on 1.27, erraudit local install,
buildflow skip_steps removal. Source:
`docs/status/2026-10-07_03-45_buildflow-red-to-green-repair-status.md` §g1.

[x] ~~samber/do + health recheck cadence~~ resolved 2026-10-31 (blanket-directive
default, reversible) — joins the ~8-week seasonal sweep (next ~2026-12-16)
AND re-triggers on `samber/do` lifecycle releases (>v2.1.0). Wired in the
AGENTS "Seasonal conventions" bullet + the series index
(`docs/architecture-understanding/README.md`). Source: 2026-10-31 recheck
report §g2; status report §g2.
