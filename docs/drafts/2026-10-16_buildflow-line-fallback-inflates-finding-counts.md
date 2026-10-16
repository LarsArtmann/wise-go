# Draft issue for LarsArtmann/BuildFlow

## Title

Generic line-fallback parser counts one finding per output line for multi-line diagnostic formats (statix: 2 findings → "27 findings reported")

## Body

```markdown
## Symptom

wise-go (Go 1.26, buildflow `83db580` binary, source read at `f7029b39e`): `statix check .` reports exactly **2 findings** (two "Avoid repeated keys" warnings in `flake.nix`). The same tree through buildflow:
```

⚠ BuildFlow passed with warnings 1/1 3.1s (27 findings)
1 tool(s) reported findings (detect only / no repair aggregation):
⚠ statix 27 findings reported

````
27 = the number of non-empty lines statix's multi-line diagnostic render printed for those 2 findings, not the finding count.

## Root cause (source)

`simpleDetectorFromCheck` (tools/providers/provider_helpers.go, `f7029b39e`) runs statix's plain `check` mode and parses the output with `FindingsFromOutput` (tools/providers/output_parser.go:111). That parser tries `ParseColonFormat` first — statix's default render is a boxed multi-line format with no `file:line:col: message` colon lines, so it matches nothing — and then falls back to `parseLineFormat` (output_parser.go:82), which creates **one finding per non-empty line**:

```go
// skip-validate
// parseLineFormat parses tool output where each non-empty line is a finding.
for line := range strings.SplitSeq(output, "\n") {
````

statix renders each finding as ~13 lines (rule header, source box, underline, label, fix hint), so 2 findings × rendered lines + context = 27 phantom findings. The `usageOrCrashMarkers` guard (gotchas #142/#143/#158/#160/#196) doesn't fire because nothing about the output looks like a usage banner — this is a real scan with real findings, just rendered in a format the fallback counts line-wise. The repo's own gotcha list already names this disease ("one warning per line"); this is a surviving instance of it.

## Fix

- [ ] Give statix a dedicated detector path that parses structured output (`statix check --json`, or `--oneline` which is one `file:line:col` finding per line and parses cleanly with the existing colon parser) instead of the generic colon→line fallback.
- [ ] Defense in depth: when the line-fallback fires for a tool whose output carries known multi-line diagnostic framing (box-drawing characters `╭ ─ │ ╰` are a strong signature), treat the output as unparseable-render rather than N findings — same spirit as the usage guard.

## To verify

On any repo with ≥1 statix finding: `buildflow -s statix` reports the same count as `statix check . --oneline | wc -l`, and the summary's finding count equals the structured finding count.

```
---

💘 Generated with Crush
```
