# Draft: crush-config issue — hermetic dprint check pattern for the skill flows

Proposed title: `Ship the hermetic dprint check pattern so agent-written markdown is always gated`

```markdown
> [!NOTE]
> This filing was drafted by GLM-5.3 via Crush from an AI-run investigation, not at my request. When the failure traces to an external report, that source is linked in the body.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

**TL;DR:** Agent sessions following the status-report/docs-health skill flows can write markdown that nothing formats or verifies until a later buildflow full run fails — wise-go shipped exactly that on 2026-10-23. The proven fix is a hermetic `dprint check` inside `nix flake check`; the recipe should live in the skill flows so every LarsArtmann repo gets it, not just the one that had the incident.

## Why

Two wise-go incidents (2026-10-23), same root: `nix fmt`/treefmt does not cover dprint file types, the auto-git daemon commits with hooks bypassed, and the pre-commit dprint hook only guards manual commits. An unformatted status report landed and was only caught by a buildflow full run later (evidence: wise-go `docs/status/2026-10-23_07-48_buildflow-recovery.md`).

## The pattern, already proven

wise-go now runs `dprint check` hermetically in `nix flake check` (`checks.dprint-format`, wise-go `flake.nix:355`, verified 2026-10-31): the four plugin `.wasm` files are `fetchurl`-pinned, a wrapper config rewrites `dprint.json`'s plugin URLs to store paths, plugins compile into a throwaway `DPRINT_CACHE_DIR`, and a built-in negative probe self-fails the derivation if the check would accept an unformatted file. No network, no daemon cooperation required.

## Proposal

- Add the hermetic-check recipe (or a link to the wise-go implementation) to the skill flows that write markdown: status-report, docs-health, and any report-producing skill — as the "gate" step next to the existing scoped `nix develop -c dprint fmt <file>` habit.
- Keep the scoped-fmt-first rule; the check is the backstop that catches whatever slipped past it.

## To verify

A repo carrying the check fails `nix flake check` after an unformatted markdown file is committed by any path, including hook-bypassing daemons; the negative probe in the derivation proves the check itself is not a silent no-op.

💘 Generated with Crush

Assisted-by: Crush:glm-5.3
```
