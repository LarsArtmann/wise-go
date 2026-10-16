# Decision Memos — 2026-10-16 (post-v0.14.0)

Five pending design decisions, each with context, options, and a
recommendation. Items that need only Lars's word (not analysis) live in
`2026-10-16_open-user-asks.md` instead.

## D1. Live-test harness shape (M18/T69): build tag vs env-gated skip

**Context.** The sandbox-live suite (statement COMPACT/FLAT, idempotence,
OTT Q4 probe) needs `WISE_SANDBOX_API_KEY`. Two shapes exist in the fleet:
a `//go:build live` file tag (excluded from default builds) and an
env-gated skip (`t.Skip` when the key is absent) inside the normal surface.

**Options.**

- (a) `//go:build live` — explicit, zero runtime cost, but the tagged file
  is INVISIBLE to default `go test ./...`, IDE runs, and `nix flake check`;
  it rots unless a dedicated step compiles it (fleet lesson: unverified
  surfaces rot), and `-tags live` forks every invocation.
- (b) env-gated skip (current shape: `TestSandboxLive` skips without the
  key) — one test surface, every gate compiles it, the recorded-run
  workflow dispatches with the secret set; cost: two branches per test.

**Recommendation.** (b), keep the current env-gated suite. The build-tag
variant optimizes for a problem wise-go does not have (live tests slowing
default runs — they skip in ~ms) while giving up the every-gate-compiles
property that caught real drift (v0.14.0's OTT migration was verified by
the default-surface compile). Revisit only if a live test grows a network
cost on EVERY run rather than on the key-present path.

## D2. Health P1b code-half (M20/T71): redocument vs fail-fast alias

**Context.** `Health(ctx)` walks the full retry ladder (~2 min worst case,
rate-budget spend) before reporting; the docs-half (2026-10-16) made the
semantics honest and pinned the typed-error contract. The open question is
whether the CODE should also offer a cheaper probe.

**Options.**

- (a) Status quo — `Health` stays the deep readiness probe; docs tell the
  consumer to wrap it with their own timeout.
- (b) Fail-fast variant — a `VerifyCredentials(ctx)`-style alias that does
  one round-trip (no retry ladder) and returns the same typed errors.

**Recommendation.** (b) as a v1.1 additive method, NOT a behavior change to
`Health` (that would break every existing consumer mid-release-train). The
credential-validating use case (startup config check) is qualitatively
different from readiness polling, the SDK already has every building
block (single-attempt executor + typed errors), and post-v1.0 additive
methods are free. Gate: land it together with the consumer recipe showing
which one to call when.

## D3. spec-drift `--fail-on-drift` posture (M20/T72)

**Context.** Vendored-snapshot drift (upstream 2027Q1 vs vendored 2026Q3)
is currently report-only everywhere; `--fail-on-drift` exists but nothing
wires it.

**Options.**

- (a) Fail the pre-release chain on drift — guarantees the vendored spec
  is fresh at every release; blocks a release on upstream's cadence.
- (b) Keep pre-release report-only; run `--fail-on-drift` on a schedule
  (weekly CI job or manual cadence) — staleness surfaces within days
  without blocking a hotfix.

**Recommendation.** (b). Drift is upstream information, not a defect in
the tree; the pre-release chain already prints the drift report where the
release author reads it, and FEATURES triage of new paths is deliberate
work that must not be forced mid-release. Recurrence home: TODO_LIST P3
row for the scheduled job (lands with CI re-enable; the `schedule:`
trigger is one YAML block).

## D4. Flake input pins: rev vs tag (M20/T73)

**Context.** In-house inputs are rev-pinned to peeled commit SHAs, with
`checks.pin-sync` (hermetic) + `nix run .#pin-sync` (network) asserting
flake rev == go.mod tag commit. The 2026-10-07 audit found tag-object pins
and two-minors-of-drift exactly where this gate was missing.

**Options.**

- (a) Status quo — rev pins + pin-sync gate; every dep bump touches go.mod
  AND flake.nix in one change.
- (b) Tag pins (`github:.../v0.11.0`) — the flake input tracks the tag
  name; the lock pins the resolved commit. Less churn per bump.

**Recommendation.** (a), keep rev pins. With tag pins the failure mode
returns: the lock silently keeps the OLD commit after a tag is re-cut or
moved, and pin-sync's whole job (detecting go.mod↔flake divergence)
degenerates to re-resolving what the lock already hid. The manual
two-place bump is the invariant's cost; `nix run .#pin-sync` verifies it
mechanically. Non-LarsArtmann inputs (nixpkgs, md-go-validator) keep
narHash-locked refs — unchanged.

## D5. x86_64-darwin in the check matrix (M20/T74)

**Context.** nixpkgs-26.11 dropped x86_64-darwin; `nix flake check
--all-systems` fails externally on it (aarch64-linux and aarch64-darwin
evaluate clean). No known consumer builds wise-go on x86_64-darwin.

**Options.**

- (a) Pin a `nixpkgs-26.05-darwin` input for the darwin leg — keeps the
  system, adds a second nixpkgs input to update and vendorHash-shift.
- (b) Drop x86_64-darwin from the matrix — `--all-systems` covers
  aarch64-darwin + linux; re-add when hardware or a consumer exists.

**Recommendation.** (b). A second pinned nixpkgs input is permanent
maintenance for a system nothing consumes; the decision is reversible in
one line if an x86_64-darwin consumer appears. Filed as ask #7 in
`2026-10-16_open-user-asks.md`.
