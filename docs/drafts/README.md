# docs/drafts/

Voice-checked filing drafts for GitHub issues in Lars's own repos (BuildFlow,
md-go-validator). Kept per skill policy: never stage issue bodies in /tmp —
each draft is a real file that doubles as the paper trail.

## Convention

- One file per issue, named `<date>_<slug>.md`.
- Body format: `# Draft issue` heading, `## Title`, then ONE markdown code
  fence (info string `markdown`) containing the exact body posted to GitHub.
- Validate with check-draft.py before posting (`--kind body-issue --ai-drafted`).
- Post with `gh issue create --body-file <real-file>` — NEVER `--body-file -`.
  The piped form fails silently under this shell (no stdout, no stderr, no
  issue; verified twice on 2026-10-07 before the working form was found).
- Extract bodies with python `str.split` on the fence marker (open on the
  `markdown`-info-string fence, close on the LAST fence in the file), never
  sed range-matches (sed stops at the FIRST inner fence and truncates
  nested-fence bodies — near-miss 2026-10-07). Byte-count the result before
  posting.
- Draft bodies contain no Go code fences today; if one ever does, the block
  must parse-check (md-go-validator shape rules) before filing.
