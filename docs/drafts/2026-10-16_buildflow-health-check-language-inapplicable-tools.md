# Draft issue for LarsArtmann/BuildFlow

## Title

Health-check phase runs language-inapplicable tools and reports them as "N tools unavailable (health check failed)" (Go-only repo: 9 JS/TS + Python tools)

## Body

```markdown
## Symptom

wise-go — a Go-only repo, zero JS/TS/Python files, no `package.json` — every full run ends its summary with:
```

⚠ 9 tools unavailable (health check failed):
[JS/TS] (7):
jest-test: tool jest-test is not installed or not in PATH (install: pnpm add -D jest)
knip: ... (install: pnpm add -D knip)
madge: ... (install: pnpm add -g madge)
publint: ... (install: pnpm dlx publint)
svelte-check: ... (install: pnpm add -D svelte-check)
vitest-test: ... (install: pnpm add -D vitest)
vue-tsc: ... (install: pnpm add -D vue-tsc)
[Other] (1):
js-coverage: ... (install: pnpm add -D c8)
[Python] (1):
interrogate: ... (install: pip install interrogate)

````
None of these tools' steps are applicable here (the same summary reports "48 not applicable"), `buildflow doctor` exits 0, and nothing about the Go toolchain is actually unhealthy — but the header reads as an environment failure on every run, and the count is dominated by tools the project can never use.

## Root cause (source)

`RunHealthChecks` (execution/pipeline.go:90 at `f7029b39e`) health-checks every provider in the post-filter provider list. jest-test's own trigger is language-gated —

```go
// skip-validate
// tools/providers/js_tools.go:37
Language: dtool.OnlyLanguage(config.LanguageJavaScript),
````

— yet jest-test still gets health-checked and lands in the failure list, so the health-check phase does not apply the same Language/Requires applicability its steps use. The summary renderer (internal/cli/workflow_summary_output_5.go:29) then labels the accumulated list "tools unavailable (health check failed)", which conflates "binary absent for a tool this repo will never run" with "environment broken".

## Fix

- [ ] Apply the step-applicability filter (Language + Requires, the same gate that produces "N not applicable") to the health-check list, so only tools whose steps could actually run get checked.
- [ ] For tools that stay in the list but whose steps never ran, either drop them from the header count or render them under a separate non-alarming group ("not installed, not applicable") — "health check failed" should mean a tool the pipeline needed was missing.

## To verify

Full run on wise-go (or any Go-only repo): the "unavailable" list contains only Go-ecosystem tools actually referenced by the pipeline (in wise-go's case: zero), and the JS/TS + Python entries no longer appear.

