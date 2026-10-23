# Archived reviews

Closed history. A review lands here only when every item is resolved — marked
inline as `~~item~~ done at <hash>`, `Won't implement`, `NOT-DO`, or routed to
TODO_LIST/ROADMAP. Do not open an archived file unless you need the original
wording for provenance.

## 2026-10-17 docs-health archive sweep

- `2026-07-23_twelve-factor-compliance.md` — RESOLVED + ARCHIVE: every
  factor-level recommendation list and the P1-P4 action plan carry inline
  verdicts (nix-canonical CI configured but execution user-gated; releases
  ship via the pre-release chain with Release objects; sandbox workflow
  shipped `ecdc738`; env-var-free library preserved; empty-key behavior now
  documented on `New`). The two routed README notes (retry composition,
  transport redaction) live in ROADMAP raw ideas.

## 2026-10-23 docs-health 2026-0x sweep (owner-ordered archive of fully-resolved reports)

- `2026-08-08_api-docs-study.md` — RESOLVED + ARCHIVE: all §8 next-steps and
  every tier-1/2/3 row carry inline outcomes (shipped v0.8.0–v0.14.0, or
  routed to the FEATURES matrix); the 2026-10-07/M22 "keep as canonical gap
  map" disposition is superseded — `FEATURES.md` owns the live 46/49/123
  matrix.
- `2026-08-21_v1.0-api-audit.md` — RESOLVED + ARCHIVE: every §2 godoc finding
  fixed or clean, §4 struck-and-routed, risk register frozen; surface
  verified 46 methods / 136 types / 31 const groups through v0.16.0. The
  flake count gate no longer parses this file (living docs + release notes
  anchor instead).
- `2026-08-08_open-banking-comparison.md` — RESOLVED + ARCHIVE: settled
  comparative analysis, zero open items; note refreshed to v0.16.0 and the
  two superseded snapshot claims (failsafe-go retry, read-only) corrected
  inline.
