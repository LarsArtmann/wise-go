# Security-Window Checklist — go 1.26.9 in nixpkgs (pre-staged 2026-10-17)

Probe at staging: locked nixpkgs `55bfb4d` ships go **1.26.8**; the window
(1.26.9, clearing all 9 stdlib vulns GO-2026-6603..6617) is NOT open. Probe
with:

```sh
nix flake metadata . --json | jq -r '.locks.nodes.nixpkgs.locked.rev'
# then
nix eval github:NixOS/nixpkgs/<rev>#go.version --raw
```

When it prints **1.26.9**, execute in order (each step gates the next):

1. `nix flake update nixpkgs` — commit the lock change alone
   (message: `chore: update nixpkgs for go 1.26.9 security window`).
2. Vendor FOD re-verify: `nix build .#check-x86_64-linux` (or let
   `nix flake check` do it) — if `vendorHash` shifts, the build names it;
   fix in the same commit series, never hand-edit the hash blindly.
3. `nix flake check` — hermetic gates (race tests, coverage floor, pin-sync
   shape, linter allowlist, script fixtures, md-go-snippets) all green.
4. Full buildflow run: `GOTOOLCHAIN=auto buildflow --fix --build-mode=full`
   (infra-touching commit rule from AGENTS.md), plus on-demand
   `buildflow -s gitleaks -s codespell`.
5. Re-measure and record:
   - `govulncheck ./...` — expect the 9 stdlib findings gone.
   - `vulnix` in the buildflow run — expect the advisory set to shrink;
     update the warning-class census note if the count changes.
6. Update the TODO_LIST security-window row with the landing date + new
   counts; fold the BuildFlow #66/#67 and branching-flow #15/#16
   consumer-side cleanups into the same sweep if they landed upstream.

Failure note: if 1.26.9 introduces a regression (jsonv2 experiment, go
directive), roll back the lock commit — the single-commit rule keeps the
revert atomic.
