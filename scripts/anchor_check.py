#!/usr/bin/env python3
"""Verify every `file.go:NN` line anchor and quoted title pointer in the
living docs still lands.

The repo's conventions carry three checkable claim shapes:

1. `` `Symbol` (`file.go:NN`) `` -- the cited line must contain Symbol as a
   word; if the symbol moved, the error reports the line where it now lives.
2. `` `file.go:NN` `` (bare) -- the file must exist, NN must be in range,
   and the cited line must be non-empty.
3. `` "Title" gotcha `` (quoted text + pointer noun, or the noun inside the
   quotes) -- the quoted text must resolve against the registry of `**bold**`
   titles from the living docs: verbatim, or as a UNIQUE prefix of exactly
   one title (titles carry inner quotes sometimes, so a stable prefix is the
   shortest enforceable pointer form). Pointer-shaped quotes that only
   paraphrase a title (high token overlap, no prefix hit) or match nothing
   are failures: a title rename would otherwise break the reference silently
   (the 2026-10-23 07:22 sweep found 5 such paraphrased pointers). Quotes
   inside code fences or inline code are prose-blind; genuinely external
   quoted phrases can be exempted in ALLOWLISTED_QUOTED with a rationale.

Point-in-time docs under docs/status/, docs/planning/, docs/reviews/,
docs/architecture-understanding/ and every archived/ tree are deliberately
NOT swept: they are historical records (docs-health annotates/archives them,
never keeps them line-true). doc-verify gates counts; this gate owns line
anchors and title pointers, which rot silently (8 dead anchors 2026-10-16,
more re-drifted the same evening).

Usage: anchor_check.py [--repo DIR] [--self-test]. Exit 0 clean, 1 drifts,
2 environment.
"""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

LIVING_DOCS = [
    "README.md",
    "FEATURES.md",
    "TODO_LIST.md",
    "ROADMAP.md",
    "AGENTS.md",
    "CONTRIBUTING.md",
]

ANCHOR_RE = re.compile(r"`((?:[A-Za-z0-9_.-]+/)*[A-Za-z0-9_-]+\.go):(\d+)`")
SYMBOL_BEFORE_RE = re.compile(r"`([A-Za-z_][A-Za-z0-9_]*)`\s*\(\s*$")

TITLE_RE = re.compile(r"\*\*(.+?)\*\*")
QUOTED_RE = re.compile(r'"([^"\n]+)"')
INLINE_CODE_RE = re.compile(r"`[^`]*`")
FENCE_RE = re.compile(r"^\s*```")

POINTER_NOUNS = (
    "gotchas",
    "gotcha",
    "entries",
    "entry",
    "bullets",
    "bullet",
    "sections",
    "section",
    "rows",
    "row",
    "memos",
    "memo",
    "notes",
    "note",
    "claims",
    "claim",
    "conventions",
    "convention",
    "items",
    "item",
    "requirements",
    "requirement",
    "warnings",
    "warning",
)
NOUN_AFTER_RE = re.compile(rf"^(?:{'|'.join(POINTER_NOUNS)})\b", re.IGNORECASE)
NOUN_TAIL_RE = re.compile(
    rf"^(?:the\s+)?(.*\S)\s+(?:{'|'.join(POINTER_NOUNS)})\s*$",
    re.IGNORECASE,
)

STOPWORDS = {
    "the",
    "a",
    "an",
    "of",
    "on",
    "in",
    "to",
    "for",
    "and",
    "or",
    "is",
    "are",
    "be",
    "with",
    "at",
    "as",
    "it",
    "its",
    "not",
    "no",
    "vs",
    "per",
    "by",
}

ALLOWLISTED_QUOTED: set[str] = set()


def prose_lines(text: str):
    """Yield (lineno, prose-line) outside fenced code, with inline-code spans
    blanked (length preserved) so wire strings inside backticks are never
    mistaken for prose quotes or titles."""
    in_fence = False
    for lineno, line in enumerate(text.splitlines(), start=1):
        if FENCE_RE.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        yield lineno, INLINE_CODE_RE.sub(lambda m: " " * len(m.group(0)), line)


def normalize_title(raw: str) -> str:
    return re.sub(r"\s+", " ", raw).strip()


def build_title_registry(docs) -> dict[str, str]:
    """Map every bold title in the living docs to the doc defining it."""
    registry: dict[str, str] = {}
    for doc, text in docs:
        for _, line in prose_lines(text):
            for match in TITLE_RE.finditer(line):
                registry.setdefault(normalize_title(match.group(1)), str(doc))
    return registry


def title_tokens(text: str) -> set[str]:
    words = re.findall(r"[a-z0-9]+", text.lower())
    return {w for w in words if len(w) > 1 and w not in STOPWORDS}


def near_miss_title(candidate: str, registry: dict[str, str]) -> str | None:
    """Return the registered title the candidate most likely paraphrases."""
    cand = title_tokens(candidate)
    if len(cand) < 2:
        return None
    best_n, best_t = 0, ""
    for title in registry:
        overlap = len(cand & title_tokens(title))
        if overlap > best_n:
            best_n, best_t = overlap, title
    if best_n >= 2 and best_n >= 0.5 * len(cand):
        return best_t
    return None


def quoted_title_drifts(doc: Path, text: str, registry: dict[str, str]):
    """Check pointer-shaped quoted strings against the title registry.

    Returns (drifts, checked): drift messages and the number of
    pointer-shaped quotes verified."""
    drifts: list[str] = []
    checked = 0
    for line_no, line in prose_lines(text):
        if line.lstrip().startswith(">"):
            continue
        for match in QUOTED_RE.finditer(line):
            quoted = match.group(1)
            after = line[match.end() :].lstrip()
            tail = NOUN_TAIL_RE.match(quoted)
            if not (NOUN_AFTER_RE.match(after) or tail):
                continue
            checked += 1

            def stripped(raw: str) -> str:
                plain = re.sub(r"^the\s+", "", raw, flags=re.IGNORECASE)
                return normalize_title(plain).strip(" \t.:;,!")

            candidates = [stripped(quoted)]
            if tail:
                candidates.append(stripped(tail.group(1)))
            candidates = [c for c in candidates if c]
            if not candidates:
                continue
            resolved = False
            for cand in candidates:
                if cand in registry or cand in ALLOWLISTED_QUOTED:
                    resolved = True
                    break
                prefix_hits = [t for t in registry if t.startswith(cand)]
                if len(prefix_hits) == 1:
                    resolved = True
                    break
            if resolved:
                continue
            near = next(
                (
                    hit
                    for hit in (near_miss_title(c, registry) for c in candidates)
                    if hit
                ),
                None,
            )
            if near:
                drifts.append(
                    f'{doc}:{line_no}: quoted pointer "{quoted}" is not a '
                    f'registered title -- paraphrase of "{near}"?'
                )
            else:
                drifts.append(
                    f'{doc}:{line_no}: quoted pointer "{quoted}" matches no '
                    f"registered bold title in the living docs"
                )
    return drifts, checked


def find_go_file(repo: Path, rel: str) -> Path | None:
    direct = repo / rel
    if direct.is_file():
        return direct
    matches = [p for p in repo.rglob(rel.split("/")[-1]) if ".git" not in p.parts]
    hits = [p for p in matches if p.relative_to(repo).as_posix().endswith(rel)]
    return hits[0] if len(hits) == 1 else None


def locate_symbol(text: str, symbol: str) -> int | None:
    pattern = re.compile(rf"\b{re.escape(symbol)}\b")
    func_pattern = re.compile(rf"\bfunc\b[^=]*\b{re.escape(symbol)}\b")
    func_hit = None
    for lineno, line in enumerate(text.splitlines(), start=1):
        if func_hit is None and func_pattern.search(line):
            func_hit = lineno
        if func_hit is not None:
            break
        if pattern.search(line):
            return lineno
    return func_hit


def check_anchor(
    repo: Path, doc: Path, line: str, match: re.Match, line_no: int
) -> str | None:
    rel, num = match.group(1), int(match.group(2))
    go_file = find_go_file(repo, rel)
    if go_file is None:
        return f"{doc}:{line_no}: anchor `{rel}:{num}` -- file not found in repo"
    text = go_file.read_text(encoding="utf-8")
    lines = text.splitlines()
    if not 1 <= num <= len(lines):
        return (
            f"{doc}:{line_no}: anchor `{rel}:{num}` -- file has only {len(lines)} lines"
        )

    tail = line[: match.start()][-120:]
    symbol = SYMBOL_BEFORE_RE.search(tail)
    if symbol is None:
        cited = lines[num - 1]
        if not cited.strip():
            return (
                f"{doc}:{line_no}: anchor `{rel}:{num}` -- cited line is blank "
                f"(the anchor drifted)"
            )
        return None

    name = symbol.group(1)
    cited = lines[num - 1]
    if re.search(rf"\b{re.escape(name)}\b", cited):
        return None
    actual = locate_symbol(text, name)
    where = f", now at {go_file.name}:{actual}" if actual else ", not found in file"
    return (
        f"{doc}:{line_no}: anchor `{rel}:{num}` -- expected `{name}` on the "
        f"cited line{where}"
    )


SELF_TEST_GO = "package fixture\n\nfunc Alpha() {}\n\nfunc Beta() {}\n"

SELF_TEST_DOC = (
    "- good: `Alpha` (`fixture.go:3`)\n"
    "- moved symbol: `Beta` (`fixture.go:3`)\n"
    "- out of range: `fixture.go:99`\n"
    "- missing file: `ghost.go:1`\n"
    "- blank line: `fixture.go:4`\n"
    "- bare ok: `fixture.go:5`\n"
)

SELF_TEST_EXPECTED = [
    "expected `Beta` on the cited line, now at fixture.go:5",
    "file has only 5 lines",
    "file not found in repo",
    "cited line is blank",
]

SELF_TEST_TITLES_DOC = (
    "# Fixture\n"
    "\n"
    "- **Real Gotcha Title** -- body text.\n"
    "- **Another Entry Name** -- body.\n"
)

SELF_TEST_QUOTED_DOC = (
    'ok exact: see the "Real Gotcha Title" gotcha.\n'
    'ok noun-inside: see "Another Entry Name" entry.\n'
    'ok unique prefix: see the "Real Gotcha" gotcha.\n'
    'paraphrase: see the "Gotcha Title for Real" gotcha.\n'
    'unmatched: see "wholly unrelated phrase here" entry.\n'
    "non-pointer quotes are ignored:\n"
    'value "111111" stays a wire literal.\n'
    "fenced code is prose-blind:\n"
    "```go\n"
    'x := "invented gotcha inside fence"\n'
    "```\n"
)


def self_test() -> int:
    import tempfile

    failures = 0
    with tempfile.TemporaryDirectory(prefix="anchor-check-selftest-") as tmp:
        root = Path(tmp)
        (root / "go.mod").write_text("module fixture\n\ngo 1.26\n", encoding="utf-8")
        (root / "fixture.go").write_text(SELF_TEST_GO, encoding="utf-8")
        (root / "README.md").write_text(SELF_TEST_DOC, encoding="utf-8")
        drifts: list[str] = []
        for lineno, line in enumerate(SELF_TEST_DOC.splitlines(), start=1):
            for match in ANCHOR_RE.finditer(line):
                problem = check_anchor(root, root / "README.md", line, match, lineno)
                if problem:
                    drifts.append(problem)
        missing = [
            exp for exp in SELF_TEST_EXPECTED if not any(exp in d for d in drifts)
        ]
        unexpected = len(drifts) - len(SELF_TEST_EXPECTED)
        if missing or unexpected != 0:
            for miss in missing:
                print(f"self-test FAIL: expected drift class not raised: {miss}")
            if unexpected != 0:
                print(f"self-test FAIL: {unexpected} unexpected drift(s): {drifts}")
            failures += 1
        else:
            print(f"anchor self-test OK: all {len(SELF_TEST_EXPECTED)} classes raised")

        registry = build_title_registry([("A.md", SELF_TEST_TITLES_DOC)])
        qdrifts, checked = quoted_title_drifts(
            Path("B.md"), SELF_TEST_QUOTED_DOC, registry
        )
        expected_quotes = [
            ("paraphrase of", 'non-prefix paraphrase "Gotcha Title for Real"'),
            ("matches no registered", "unmatched phrase must be flagged"),
        ]
        qmissing = [
            hint
            for fragment, hint in expected_quotes
            if not any(fragment in d for d in qdrifts)
        ]
        if qmissing or len(qdrifts) != len(expected_quotes) or checked != 5:
            for miss in qmissing:
                print(f"self-test FAIL: quoted-pointer class not raised: {miss}")
            if len(qdrifts) != len(expected_quotes):
                print(f"self-test FAIL: unexpected quoted drifts: {qdrifts}")
            if checked != 5:
                print(f"self-test FAIL: expected 5 checked pointers, got {checked}")
            failures += 1
        else:
            print(
                "quoted-title self-test OK: exact/prefix/noun-inside pass, "
                "paraphrase+unmatched flagged, fence+non-pointer ignored"
            )
    return 1 if failures else 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--self-test",
        action="store_true",
        help="run the built-in negative-scenario fixture and exit",
    )
    parser.add_argument(
        "--repo",
        default=".",
        help="repo root (default: current directory)",
    )
    args = parser.parse_args()
    if args.self_test:
        return self_test()
    repo = Path(args.repo).resolve()
    if not (repo / "go.mod").is_file():
        print(f"error: {repo} is not the repo root (no go.mod)", file=sys.stderr)
        return 2

    docs = [repo / name for name in LIVING_DOCS]
    docs += sorted((repo / "docs").glob("*.md"))
    missing = [p.name for p in docs if not p.is_file()]
    docs = [p for p in docs if p.is_file()]
    if missing:
        print(
            f"warning: living docs not found, skipped: {', '.join(missing)}",
            file=sys.stderr,
        )

    texts = [(doc, doc.read_text(encoding="utf-8")) for doc in docs]

    drifts: list[str] = []
    anchors = 0
    for doc, text in texts:
        for lineno, line in enumerate(text.splitlines(), start=1):
            if line.lstrip().startswith(">"):
                continue
            for match in ANCHOR_RE.finditer(line):
                anchors += 1
                problem = check_anchor(repo, doc, line, match, lineno)
                if problem:
                    drifts.append(problem)

    registry = build_title_registry(texts)
    pointers = 0
    for doc, text in texts:
        qdrifts, checked = quoted_title_drifts(doc, text, registry)
        pointers += checked
        drifts.extend(qdrifts)

    for drift in drifts:
        print(f"FAIL: {drift}")
    if drifts:
        print(
            f"anchor-check: {len(drifts)} drifted of {anchors} anchors + "
            f"{pointers} quoted pointers across {len(docs)} living docs"
        )
        return 1
    print(
        f"anchor-check: all {anchors} anchors + {pointers} quoted pointers "
        f"verified across {len(docs)} living docs"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
