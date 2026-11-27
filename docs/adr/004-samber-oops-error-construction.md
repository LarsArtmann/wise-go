# ADR 004: Error construction with `samber/oops` next to `go-error-family`

**Status:** Accepted (decided and executed 2026-11-27)
**Deciders:** Lars Artmann
**Context date:** 2026-11-27
**Re-verified:** 2026-11-27 against samber/oops v1.23.2 (source read, scratch consumer run)

## Context

Every error the SDK adds context to was built with the standard library:
`fmt.Errorf("context: %w", err)` at call sites and `errors.New` for package
sentinels. `erraudit lint ./... --max` reports each of those as a violation
("project enforces samber/oops"), and that advice had been declined several
times on one ground: keeping the library dependency-light.

The owner reversed that ground on 2026-11-27. Minimizing dependencies was
never a goal for this project; `samber/oops` and `go-error-family` are both
wanted, and they answer different questions:

- `go-error-family` says what an error **means**: its family (Rejection,
  Transient, Corruption), its code, whether it is retryable.
- `samber/oops` says where an error **happened**: the wrap layer, a stack
  trace, key-value context, available to any consumer without extra setup.

## Decision

1. `github.com/samber/oops` is a direct production dependency.
2. Contextual wraps use `oops.Wrapf(err, "message", args...)`. It renders as
   `"message: <wrapped>"`, exactly the text `fmt.Errorf("message: %w", err)`
   produced, so no pinned message changes.
3. New errors with a dynamic message use `oops.Errorf`; fixed-message errors
   use `oops.New`.
4. Package sentinels stay plain, comparable values, built only through
   `newSentinel` (`errors.go`).
5. The erraudit gate adds `--enforce-samber-oops`.

### Why sentinels do not use oops

`oops.OopsError` holds maps and slices, so it is not comparable.
`APIError.Is` looks its target up in `codeSentinels`, a map keyed by the
sentinel value; an `OopsError` key would panic at match time. The sentinel
is an identity token, not a diagnostic record, and a stack trace of a
package-level variable carries no information. One suppression
(`newSentinel`, with this rationale inline) is the whole cost.

## Consequences

- **Consumers gain diagnostics.** The context layers are `oops.OopsError`:
  `%+v` prints the stack and context, `oops.AsOops(err)` exposes the layers.
  `%v`, `%s` and `error.Error()` are unchanged.
- **Matching is unaffected.** `errors.Is`, `errors.AsType` and `APIError.Is`
  traverse `OopsError.Unwrap` exactly as they traversed the `fmt.Errorf`
  chain (pinned by the existing sentinel and mapping tests, extended with an
  oops wrap layer).
- **Error creation is slower.** An `oops.Wrapf` captures a stack: about
  1.3 µs against 0.14 µs for `fmt.Errorf` (microbenchmark, 2026-11-27). It
  runs only on failure paths, orders of magnitude below one HTTP round trip.
- **Dependency graph grows.** `oops` brings `samber/lo`, `oklog/ulid` and
  `go.opentelemetry.io/otel`. Rewriting loops into `samber/lo` calls stays
  declined; its presence as an indirect dependency is not an endorsement.
- **`erraudit --max` still exits non-zero**, on the `generic_return` rule
  only: every method returns `error` because its error surface is a union.
  That advisory stays accepted (AGENTS.md, "erraudit `--enforce-generic-return`
  findings are deliberate non-fixes").
- **`wrapcheck.ignore-sigs` gains `.New(`**, matching the existing `.Wrap(`,
  `.Wrapf(` and `.Errorf(` entries: `oops.New` is an error constructor, not
  an unwrapped pass-through.
