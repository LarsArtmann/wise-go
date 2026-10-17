# Open User Asks — 2026-10-16 (post-v0.14.0; refreshed for v0.15.0 on 2026-10-17)

Eight decisions only the user can make, each with the exact unblock step.
v0.15.0 (the structured error surface) shipped 2026-10-17 since this file was
written — ask #5 is done and #8's open question resolved itself. Everything
else from the 21:45 plan is either done (see the plan's inline
~~annotations~~) or owned by the parallel session's active tiers.

## 1. Enable CI (one command)

`.github/workflows/ci.yml` is `disabled_manually` on GitHub. All local gates
are green at the v0.15.0 tag (full pre-release chain exit 0); the workflow is
refreshed and carries the warn-first art-dupl job and the pin-sync network
job. The first-run triage runbook is written:
`2026-10-17_ci-first-run-watch-package.md`.

```sh
gh workflow enable ci -R LarsArtmann/wise-go
```

Then push any non-docs commit (docs-only pushes are `paths-ignore`d) and
watch the first run; the coverage badge unfreezes on the next master push.

## 2. A Wise sandbox API key

`WISE_SANDBOX_API_KEY` unlocks, in one move: the live-smoke suite's first
recorded run (harness ready since 78a4f35), the 5-assumption live-verification
pass, and the OTT `2026Q3 → 2026Q4` version probe that parks `ottAPIVersion`
today. Export it in the environment (or a direnv-ignored `.env`), then run:

```sh
WISE_API_KEY=$WISE_SANDBOX_API_KEY go test -run TestSandboxLive -v ./...
```

## 3. bank-sync redeploy (v0.14.0 migration is committed)

bank-sync now builds and tests green against wise-go v0.14.0 (9 OTT call sites
migrated to the typed `OneTimeToken`), and — v0.15.0 refresh — compiles
UNMODIFIED against the v0.15.0 tree (replace-compile verified 2026-10-17;
the `*BadRequestError` reclassification breaks no call site). What remains
is the host-side rollout:
the NixOS host config (outside both repos) consumes bank-sync as a flake
input, so the deploy is a host-config input bump + rebuild. When bank-sync
next bumps its wise-go pin, point it at v0.15.0 and the matching recipes in
`docs/error-codes.md`. Verify afterwards
via the read-only browser smoke:

```sh
nix run .#smoke   # in bank-sync; expects prod on 127.0.0.1:8097
```

and confirm the FX footer reports the bank-sync build riding the new
wise-go version.

## 4. nixpkgs bump window (wise-go)

The scheduled `nix flake update nixpkgs` rehashes every vendorHash (toolchain
shift) — mid-session, with a parallel session active, that is a cascade risk
(the M17 deferral from the 19:45 run). Pick a quiet window and run:

```sh
nix flake update nixpkgs   # then fix vendorHashes one by one until flake check is green
```

## 5. ~~Push master (or leave it to the parallel session)~~ DONE

~~Local wise-go master is 14+ commits ahead of origin~~ Resolved 2026-10-17:
the v0.15.0 release train pushed master (through `415cce4`, the tag commit)
and the `v0.15.0` tag; the release is live on GitHub and the module proxy.

## 6. CI token secrets (cachix + erraudit)

Two jobs in `ci.yml` degrade to `continue-on-error` without their secrets:

- `CACHIX_AUTH_TOKEN` — enables the nix cache push (faster CI, no behavior
  change without it). `cachix authtoken <token>` → `gh secret set CACHIX_AUTH_TOKEN`.
- `ERRAUDIT_TOKEN` — the private `erraudit` job (warn-not-fail). With
  `GOPRIVATE=github.com/larsartmann/erraudit`:
  `gh secret set ERRAUDIT_TOKEN <PAT-with-repo-read>`.

Without them the first green run is still meaningful; set them any time before
treating CI as the release gate.

## 7. x86_64-darwin matrix decision

`nix flake check --all-systems` fails EXTERNALLY: nixpkgs-26.11 dropped
x86_64-darwin (2026-10-07). Pick one:

```sh
# Option A: pin a nixpkgs-26.05-darwin input and keep the system
# Option B: drop x86_64-darwin from the check matrix (aarch64-darwin stays)
```

Until decided, `nix flake check` (default systems) stays the gate and
`--all-systems` stays advisory.

## 8. v1.0.0 timing

46 exported methods, apidiff-clean (gorelease suggests the minor v0.15.0,
no blocking deltas), coverage at the 92.8% local measure vs the 92
floor. v1.0.0 is a `git tag` away whenever you call it; the pre-release chain
(`nix run .#pre-release -- --shuffle`) is the only gate. The old open
question — tag before or after the error-system breaking surface — resolved
itself: v0.15.0 shipped the structured errors (`APIError.Details`,
`WiseErrorCode`, sentinels, `*BadRequestError`) on 2026-10-17, so a v1.0.0
tagged from here carries them from day one.
