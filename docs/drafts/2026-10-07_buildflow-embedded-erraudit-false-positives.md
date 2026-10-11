# Draft issue for LarsArtmann/buildflow

## Title

embedded erraudit step reports 29 false positives the standalone binary does not

## Body

```markdown
## Symptom

On `github.com/larsartmann/wise-go` (root-package Go SDK, go 1.26, GOEXPERIMENT=jsonv2), BuildFlow's embedded `erraudit` step reports 29 error-severity findings. The canonical standalone `erraudit` (`github.com/larsartmann/erraudit@v0.5.1-0.20260922174106-1c6809adf02e`, invocation `erraudit lint ./... --type-aware --enforce-coded-errors`) is exit-0 clean on the identical tree.

## What breaks

- Consumers must `skip_steps` the embedded step to get a green run, which also mutes it for repos where it would be right.
- The two tools disagree, so "which erraudit is authoritative" is unanswerable for consumers.

Verified identical at buildflow `3bb229e` and `202b114` (repo HEAD `0e9f5e301`, 2026-10-07). Findings breakdown:

- 8× `[ignored]`: two shapes, both misreads:
  - `_, ok := errors.AsType[T](err)` — the discarded value is a TYPE, not an error; there is nothing to echo.
  - `_ = resp.Body.Close()` — deliberate cleanup discard, not an unhandled error.
- 21× `[context_loss]`: demands every in-scope variable be echoed into every `fmt.Errorf` wrap message. That contradicts wise-go's error-context convention (the inner typed error carries classification; wrap messages add call-site context, not a variable dump).

## Root cause (source)

The embedded analyzer and the standalone `erraudit` binary diverge — same rule names, different verdicts on the same input. The standalone is the reference implementation the CI gate (`erraudit` job, warn-not-fail) runs; the embedded duplicate disagrees with it on:

- two-value `errors.AsType[T](err)` assignment semantics (`[ignored]` fires on a discarded type, not a discarded error)
- wrap-message context policy (`[context_loss]` wants full variable echo; standalone accepts the convention)

## Operator steps (human)

- [ ] Decide which implementation is authoritative (standalone binary is the one pinned in CI contracts and the go-error-modernization skill).
- [ ] Fix the embedded analyzer's `AsType` two-value handling (a discarded TYPE is not an ignorable error).
- [ ] Align `context_loss` with the standalone's policy, or gate it behind a flag.
- [ ] Re-run the embedded step against wise-go and compare verdicts byte-for-byte with the standalone before lifting the consumer-side skip.

## To verify

`buildflow` on wise-go with the `erraudit` skip removed reports 0 error-severity findings, matching the standalone's exit-0 on the same commit.
```

---

💘 Generated with Crush
