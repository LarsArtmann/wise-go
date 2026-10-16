# Open User Asks — 2026-10-16 (post-v0.14.0)

Eight decisions only the user can make, each with the exact unblock step.
Everything else from the 21:45 plan is either done (see the plan's inline
~~annotations~~) or owned by the parallel session's active tiers.

## 1. Enable CI (one command)

`.github/workflows/ci.yml` is `disabled_manually` on GitHub. All local gates
are green at the v0.14.0 tag; the workflow is refreshed and carries the
warn-first art-dub job and the new pin-sync network job.

```sh
gh workflow enable ci -R LarsArtmann/wise-go
```

Then watch the first run and confirm the coverage badge unfreezes on the next
master push. Ratification is the gate; the command itself is the only step.

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
migrated to the typed `OneTimeToken`). What remains is the host-side rollout:
the NixOS host config (outside both repos) consumes bank-sync as a flake
input, so the deploy is a host-config input bump + rebuild. Verify afterwards
via the read-only browser smoke:

```sh
nix run .#smoke   # in bank-sync; expects prod on 127.0.0.1:8097
```

and confirm the FX footer reports the bank-sync build riding wise-go 0.14.0.

## 4. nixpkgs bump window (wise-go)

The scheduled `nix flake update nixpkgs` rehashes every vendorHash (toolchain
shift) — mid-session, with a parallel session active, that is a cascade risk
(the M17 deferral from the 19:45 run). Pick a quiet window and run:

```sh
nix flake update nixpkgs   # then fix vendorHashes one by one until flake check is green
```

## 5. Push master (or leave it to the parallel session)

Local wise-go master is 14+ commits ahead of origin (release authoring, the
v0.14.0 tag's commit, coverage-gap tests, the parallel session's OTT + gate
work). The tag is pushed and the release is live; master itself is untouched
on the remote. Either the parallel session pushes when its current tier
settles, or:

```sh
git push origin master   # after the tree is clean and lint-green
```

Note: at the moment the release was cut, HEAD (f045de1) carried the parallel
session's committed-but-unformatted OTT change — that is why the tag points
at 1271562. If pushing master now, confirm `golangci-lint run ./...` is clean
first.

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

46 exported methods, apidiff-clean against v0.14.0, coverage 93.5% vs the 92
floor. v1.0.0 is a `git tag` away whenever you call it; the pre-release chain
(`nix run .#pre-release -- --shuffle`) is the only gate. Open question only
you can answer: tag now on the 46-method surface, or after the error-system
plan lands its breaking surface (`APIError.Details`, `WiseErrorCode`) so v1.0
consumers get the structured errors from day one.
