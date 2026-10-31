# `RequestLog.RetryAfterDelay` API freeze — pre-decision memo (2026-10-31 12:28 CEST)

Status: MEMO for the v1.x planning moment (TODO row parked for v1.0 by owner
decision 2026-10-17). No code change; records the recommendation so the
v1.x call is a rubber stamp.

## Context

`RequestLog.RetryAfterDelay` (`options.go:44`) carries the delay-decision
entry shape: positive only on delay decisions, 429 in `Status`, honored
wait in `Duration`/`RetryAfterDelay`, `Method`/`URL` empty — no exchange
happened. Behavior is BDD-pinned (429 3-entry and two-delay 5-entry
sequences, plus the 2026-10-31 deadline-abort contexts: the delay decision
is emitted, then silence).

## Options at v1.x

- **(a) Freeze the entry as-is** — document the shape as the v1.x contract
  in README/CHANGELOG at the v1.x release. Zero code. The pinned tests
  already are the contract; freezing just makes breaking it a major-version
  event.
- **(b) Observer interface** — `type RetryDelayObserver interface{ OnRetryDelay(RetryDelayDecision) }` + a dedicated channel. More surface, a
  second logging path to keep coherent, and NO consumer has asked: the
  `Logger` is data-complete (the AGENTS OTEL convention already notes the
  meter bridge MUST skip delay entries — an observer would not change that
  requirement).

## Recommendation

**(a)** — freeze the entry as-is at v1.x, with one CHANGELOG line declaring
the `RequestLog` entry shapes (request entries + delay-decision entries)
part of the v1 compatibility surface. Revisit (b) only when a consumer
demonstrates a need the `Logger` cannot carry (none known: observability
consumers bridge `LogRequest` directly, incl. the otel-meter recipe,
verified 2026-10-19).

## Execution checklist (at the v1.x release)

1. CHANGELOG v1.x: declare `RequestLog` shapes frozen.
2. README Observability section: one sentence naming the frozen contract.
3. TODO row strike (this memo referenced as the decision record).
