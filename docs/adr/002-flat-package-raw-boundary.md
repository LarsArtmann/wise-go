# ADR 002: Flat public package with an `internal/raw` wire boundary

**Status:** Accepted (2026-08-21 audit; codified 2026-09-13)
**Deciders:** Lars Artmann

## Context

Go SDKs for large APIs face a packaging fork: sub-packages per resource
(`wise/transfers`, `wise/quotes`, …) promise modularity but create import cycles
around shared types (`Money`, branded IDs), fragment the docs, and make
cross-resource consistency hard. Conversely, one flat package risks god-package
entropy as the surface grows (37 `*Client` methods across 15 resources as of
2026-09-13).

Wise's JSON wire format also leaks into any naive type design: snake_case keys,
polymorphic polymorph IDs (int64 for most entities, UUID strings for quotes and
webhook subscriptions), and timestamps in four different layouts.

## Decision

1. **One public package `wise`** with a flat `client.X` surface. All entities,
   value objects, enums, and error types live there.
2. **`internal/raw` is the only place that knows the wire.** Raw structs use
   primitives (`int64`/`string`/`float64`) with JSON tags matching Wise exactly.
   They are invisible to consumers (Go `internal/` rule). Mapper functions
   (`mapProfile`, `toWebhookSubscription`, …) convert raw → typed; decode
   failures are corruption-classified.
3. **Public types are tag-free.** Raw and public structs intentionally mirror
   each other; the duplication is accepted (art-dupl suppression) because type
   aliases would leak Wise's wire JSON tags into the public serialization
   surface.
4. **The service-client split stays deferred.** The ROADMAP trigger (resource
   count crossed ~6–8; now 15) is reached; the recommended sequencing remains
   v1.x on the flat surface, then a sub-structure refactor in one release cycle.

## Consequences

- Consumers import exactly one package; godoc and README stay coherent.
- Adding an endpoint = raw type + mapper + client method, always in that shape
  (see AGENTS.md conventions).
- The flat surface is bounded by discipline, not the compiler: when the
  sub-structure refactor lands, `internal/raw` and the mapper layer carry over
  unchanged.
- `go doc -all` is the authoritative surface inventory (the 2026-08-21 audit and
  2026-09-13 re-audit both used it).
