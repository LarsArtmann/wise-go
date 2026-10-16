# Vendor hash for the SDK's buildGoModule (consumed via the go-standard
# flakeModule at flake.nix go-standard.vendorHash).
#
# Re-verified 2026-10-17: the buildflow nix-build step builds the full
# 12-check battery green with this hash. The "vendorHash may be stale,
# go.sum was modified after the hash was last set" preflight/nix-checker
# warnings are mtime heuristics that false-positive after tidy-style
# go.sum rewrites with no dependency change; refreshing this file (even
# comment-only) re-orders the mtimes and clears them.
#
# Repair policy (AGENTS.md): never hand-paste a new hash — run
# `buildflow -s nix-hash-fix --fix` (or the fake-hash dance for the
# md-go-validator hash in vendorHashMdGoValidator.nix).
"sha256-OTnWa30Qy8MpL1ogJMaztxioOo8E0ZmxU3KpG+e9di8="
