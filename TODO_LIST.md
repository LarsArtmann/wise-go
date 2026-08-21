# TODO List

Short- and mid-term actionable work for wise-go. Each item is bounded, has clear
ownership, and can be completed in one sitting. For long-term vision and raw ideas
see [ROADMAP.md](ROADMAP.md). For shipped features see [FEATURES.md](FEATURES.md).

## P2 — v1.0 release (API lock)

[ ] Add credentialed Wise sandbox integration tests — create a protected or
manually triggered live-test workflow using sandbox credentials, cover the supported
read-only endpoints, and keep it separate from fork-safe mock tests. The sandbox URL
was updated to V2 (`api.wise-sandbox.com`); tests must verify the new endpoint works.
This is required to verify that the SDK actually interoperates with Wise's sandbox
rather than only matching mocked responses.
**BLOCKED: needs a sandbox API key from the user.**

[ ] Lock the public API at v1.0 — the formal audit is DONE
(`docs/reviews/2026-08-21_v1.0-api-audit.md`: 31 client methods, 74 types,
godoc reviewed, breaking-change risk register — nothing blocks the tag).
v0.9.0 shipped 2026-08-21; remaining: tag `v1.0.0`.
**BLOCKED: needs the user's explicit approval (tagging is irreversible).**

## P3 — Tier-2 API surface (plan doc tier 2 + near-term reports)

(shipped 2026-08-21: mTLS pattern documented in README — `WithBaseURL` +
`WithHTTPClient` compose; no dedicated option needed.)

## P4 — Tooling & quality

(shipped 2026-08-21, execution session:)

- **govulncheck triaged** — all 4 reachable findings (GO-2026-6218, -6090,
  -5972, -5026) are Go standard library at go1.26.5, each fixed in go1.26.6.
  Zero SDK-code remediation possible; the fix is a toolchain bump. CI's
  `go-version: "1.26"` picks up the patch automatically once setup-go serves
  it; the nix devShell follows nixpkgs' `go_1_26`. No action left except
  re-running `govulncheck` after the toolchain moves.
- **Cachix** — `cachix/cachix-action` (pinned to the verified v15 commit
  ad2ddac) pushes build outputs from the `nix:` job on master. Needs the
  `CACHIX_AUTH_TOKEN` secret set once by the user.
- **Coverage badge automated** — CI commits `.github/badges/coverage.json`;
  README reads it via the shields.io dynamic-json endpoint (86.9% at ship
  time). Measurement basis documented under the badge.
- **Lychee link check** — offline `links` gate in `nix flake check` over the
  living docs; first run caught and fixed one ghost LICENSE reference.
- **Housekeeping bundle** — 68-vs-77 resolved (both counting bases documented
  inline in report 18-15; canonical: 121 `It(` specs today),
  `wise-api-core-schemas.json` mentioned in AGENTS.md beside the main spec.
