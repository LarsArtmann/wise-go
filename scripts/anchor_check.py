#!/usr/bin/env python3
"""Verify every `file.go:NN` line anchor in the living docs still lands.

The repo's convention is `Symbol` (`file.go:NN`). Two claim shapes are
checked:

1. `` `Symbol` (`file.go:NN`) `` -- the cited line must contain Symbol as a
   word; if the symbol moved, the error reports the line where it now lives.
2. `` `file.go:NN` `` (bare) -- the file must exist, NN must be in range,
   and the cited line must be non-empty.

Point-in-time docs under docs/status/, docs/planning/, docs/reviews/,
docs/architecture-understanding/ and every archived/ tree are deliberately
NOT swept: they are historical records (docs-health annotates/archives them,
never keeps them line-true). doc-verify gates counts; this gate owns line
anchors, which rot silently (8 dead anchors 2026-10-16, more re-drifted the
same evening).

Usage: anchor_check.py [--repo DIR]. Exit 0 clean, 1 drifts, 2 environment.
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


def check_anchor(repo: Path, doc: Path, line: str, match: re.Match, line_no: int) -> str | None:
    rel, num = match.group(1), int(match.group(2))
    go_file = find_go_file(repo, rel)
    if go_file is None:
        return f"{doc}:{line_no}: anchor `{rel}:{num}` -- file not found in repo"
    text = go_file.read_text(encoding="utf-8")
    lines = text.splitlines()
    if not 1 <= num <= len(lines):
        return f"{doc}:{line_no}: anchor `{rel}:{num}` -- file has only {len(lines)} lines"

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


SELF_TEST_GO = (
    "package fixture\n\n"
    "func Alpha() {}\n\n"
    "func Beta() {}\n"
)

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


def self_test() -> int:
    import tempfile

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
        missing = [exp for exp in SELF_TEST_EXPECTED if not any(exp in d for d in drifts)]
        unexpected = len(drifts) - len(SELF_TEST_EXPECTED)
        if missing or unexpected != 0:
            for miss in missing:
                print(f"self-test FAIL: expected drift class not raised: {miss}")
            if unexpected != 0:
                print(f"self-test FAIL: {unexpected} unexpected drift(s): {drifts}")
            return 1
        print(f"self-test OK: all {len(SELF_TEST_EXPECTED)} drift classes raised")
        return 0


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
        print(f"warning: living docs not found, skipped: {', '.join(missing)}", file=sys.stderr)

    drifts: list[str] = []
    anchors = 0
    for doc in docs:
        text = doc.read_text(encoding="utf-8")
        for lineno, line in enumerate(text.splitlines(), start=1):
            if line.lstrip().startswith(">"):
                continue
            for match in ANCHOR_RE.finditer(line):
                anchors += 1
                problem = check_anchor(repo, doc, line, match, lineno)
                if problem:
                    drifts.append(problem)

    for drift in drifts:
        print(f"FAIL: {drift}")
    if drifts:
        print(
            f"anchor-check: {len(drifts)} drifted of {anchors} anchors "
            f"across {len(docs)} living docs"
        )
        return 1
    print(f"anchor-check: all {anchors} anchors verified across {len(docs)} living docs")
    return 0


if __name__ == "__main__":
    sys.exit(main())
