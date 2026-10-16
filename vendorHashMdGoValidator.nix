# Vendor hash for the md-go-validator source-only flake input, consumed by
# checks.md-go-snippets (built with go 1.27 + proxyVendor + jsonv2).
# Extracted from the inline flake.nix literal 2026-10-17 (nix-checker
# hardcoded-hash finding); value verified unchanged by rebuilding the
# md-go-snippets check after the extraction.
#
# Gotcha (AGENTS.md "Markdown Go-snippet check"): this hash is computed
# under go 1.27 and DIFFERS from what a go 1.26 build produces. After
# md-go-validator dependency changes, surface the new hash via the
# fake-hash dance, never by hand-computing.
"sha256-QUgeh99RqCc80oh1UNJDH38Llm8jMW3hQkKmPGZr3NE="
