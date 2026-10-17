# md-go-validator: 4-backtick fences should hide inner ```go blocks (CommonMark)

## Problem

The fence scanner toggles on any line starting with ` ``` ` and never records the opening fence's length (`pkg/extractor.go:111`, `startCodeBlock` at `pkg/extractor.go:141`). Under CommonMark, a fence opened with N backticks is only closed by a closing sequence of at least N backticks — so a 4-backtick fence containing a ` ```go ` block is literal text, not a nested fence. The extractor still extracts and validates the inner block, and fails files that are correct Markdown:

```markdown
````            ← 4-backtick fence: literal text per CommonMark
```go
this is (not) valid go {
```
````
```

## Repro (v1.3.0, built from the v1.3.0 tag source)

```sh
printf 'text\n\n````\ninner literal text\n\n```go\nthis is (not) valid go {\n```\n\n````\n\nmore text\n' > nested.md
md-go-validator nested.md   # exit 1: "Go syntax error" at nested.md:6
```

The inner block should be skipped entirely; expected exit 0, 0 blocks.

## Impact

Every Markdown file that uses a 4-backtick fence to SHOW Markdown examples with Go blocks inside (docs about the validator itself, tutorials, test fixtures) fails validation on content that renders as plain text. The current workaround is `// skip-validate` inside the shown snippet, which pollutes displayed examples.

## Fix

Track the opening fence's backtick count at `startCodeBlock` and only accept a closing run at `endCodeBlock` when its length is greater than or equal to the opener's. Openers longer than 3 backticks with a language tag (`` ````go ``) stay valid — CommonMark allows info strings on long fences too.

---
💘 Generated with Crush
