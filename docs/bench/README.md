# Benchmarks

Performance baselines for the wise-go hot paths (parsers, mappers, signature
verification). Each baseline file is a raw `go test -bench` output captured at
a release tag, named `<date>_<tag>_baseline.txt`.

| File | Captured at |
| --- | --- |
| [2026-10-07_v0120_baseline.txt](2026-10-07_v0120_baseline.txt) | v0.12.0 |

## Rerun and compare

Capture the current numbers, then let `benchstat` diff them against the last
baseline:

```sh
go test -bench . -count 6 -benchtime 1s > /tmp/current.txt
benchstat docs/bench/2026-10-07_v0120_baseline.txt /tmp/current.txt
```

`-count 6` matters: benchstat needs multiple samples per benchmark to compute
variance, and single-run comparisons are noise.

## Conventions

- Consume the latest baseline in any PR that touches a benchmarked path
  (`parseWiseTimestamp`, `classifyTransactionType`, `mapTransaction`,
  `VerifyWebhookSignature`, `decodeExchangeRates`, `ParseWebhookEvent`) —
  a regression above noise should be explained in the PR, not discovered later.
- Cut a new baseline at each release tag (`go test -bench . -count 6`), commit
  it as `<date>_<tag>_baseline.txt`, and add it to the table above. Baselines
  are append-only history; never edit an old one.
- Benchmarks live in the repo root `bench_test.go`; add a benchmark when a
  parser or mapper becomes a hot path for consumers, not speculatively.
