# Draft issue for LarsArtmann/buildflow

## Title

file-size-check scans 0 files in root-package Go layouts (max_file_size inert)

## Body

```markdown
## Symptom

On `github.com/larsartmann/wise-go` — a Go library whose package lives at the repo root (`wise.go`, `client.go`, … beside `go.mod`, no `/internal`, no `/pkg`) — BuildFlow's `file-size-check` tool scans **0 files**. `max_file_size` at both 350 and 700 lines produces the same result: the gate never sees a single file, so it can never fail.

## What breaks

- `max_file_size` is config-only theater in root-package layouts: accepted, reported, inert.
- A 2,000-line file at the repo root would pass a 350-line gate.

## Root cause (source)

The scanner's file discovery walks conventional Go layout roots (`/internal`, `/pkg`, …) and misses the root package itself. wise-go is root-package BY DESIGN (SDK library layout; documented non-fix in its AGENTS.md), so the correct behavior is: include root-level `*.go` files, not only conventional subdirs.

## Operator steps (repo)

- [ ] Include root-package `*.go` files in the scan set (a root package is still a package).
- [ ] Report the scanned-file count in the tool's summary so "0 files scanned" is visible instead of a silent pass.
- [ ] Optionally: fail (or warn) when the scan set is empty — a size gate that saw no files proved nothing.

## To verify

Run buildflow on wise-go at ~HEAD: summary shows a non-zero scanned-file count and `max_file_size: 350` evaluates against the real files (several exceed 350 lines today, so the gate should produce findings, not silence).
```

---

💘 Generated with Crush
