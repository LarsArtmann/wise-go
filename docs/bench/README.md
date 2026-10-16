# Benchmarks

Performance baselines for the wise-go hot paths (parsers, mappers, signature
verification). Each baseline file is a raw `go test -bench` output captured at
a release tag, named `<date>_<tag>_baseline.txt`.

| File                                                           | Captured at |
| -------------------------------------------------------------- | ----------- |
| [2026-10-07_v0120_baseline.txt](2026-10-07_v0120_baseline.txt) | v0.12.0     |

## Rerun and compare

`benchstat` is not in the devShell (nixpkgs does not package it); install it
once with `go install golang.org/x/perf/cmd/benchstat@latest` (lands in
`~/go/bin`, which is on PATH).

Capture the current numbers, then let `benchstat` diff them against the last
baseline. Ginkgo rejects `go test -count>1`, so the six samples must come from
six separate invocations (verified 2026-10-11):

```sh
: > /tmp/current.txt
for i in 1 2 3 4 5 6; do
  go test -run '^$' -bench . -benchtime 1s -count 1 >> /tmp/current.txt
done
benchstat docs/bench/2026-10-07_v0120_baseline.txt /tmp/current.txt
```

Six samples per benchmark matter: benchstat needs multiple samples to compute
variance, and single-run comparisons are noise. `-run '^$'` skips the unit
suite so only benchmarks run. Benchmarks added after a baseline's date simply
have no baseline column — benchstat shows them one-sided (expected, not an
error).

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
- Live-fuzz "interesting input" artifacts land in the Go build cache, NOT in
  `testdata/` — the checked-in seeds already pin the regression shapes, and
  harvesting cache corpora bloats the repo (decided 2026-10-11; do not
  relitigate).
