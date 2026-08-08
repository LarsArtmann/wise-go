# TODO List

Short- and mid-term actionable work for wise-go. Each item is bounded, has clear
ownership, and can be completed in one sitting. For long-term vision and raw ideas
see [ROADMAP.md](ROADMAP.md). For shipped features see [FEATURES.md](FEATURES.md).

## P1 — v1.0 release (API lock)

[ ] Lock the public API at v1.0 — formal audit of every exported symbol, godoc
review pass, then tag `v1.0.0`. All v0.5.0 breaking changes are shipped and tested;
the lock is the logical next milestone. Requires explicit approval (tagging is
irreversible). The API surface is now stable: `Money`/`Currency` value objects,
branded IDs, typed enums (`DetailType`, `TransactionType`, `ProfileType`,
`BalanceType`, `InvestmentState`), and the two-layer raw/result split are all
finalized.

## P2 — CI speed

[ ] Add Cachix binary cache to the `nix:` CI job — `nix flake check` no longer
uses `--no-build` (v0.5.0 change), so the full sandboxed test runs. Without a
binary cache, building `go_1_26` from nixpkgs source takes 15+ minutes. Add
`cachix/cachix-action` with a public cache for `nixpkgs` to keep CI under 5 minutes.
