#!/usr/bin/env python3
"""Negative fixtures for scripts/pin_sync.py and scripts/spec_drift.py.

Every fixture here encodes a failure mode the scripts once shipped with or
could regress into: the pin-sync trio (flake.lock drifting from flake.nix, a
flake tag lagging go.mod, a go.mod dep with no flake input at all), the
unpinned-input shape, and the spec-drift quartet (unparseable live spec, an
SDK path the spec no longer has, a stale vendored snapshot under
--fail-on-drift, and the same drift in warn-only mode).

Run hermetically (no network): every case passes an explicit --spec-file or
uses --offline, so nothing is fetched. Wired into `nix flake check` as
checks.script-fixtures; runnable directly:

    python3 scripts/negative_fixtures.py
"""

from __future__ import annotations

import argparse
import pathlib
import subprocess
import sys
import tempfile

PIN_SYNC_CASES = [
    ("ok", 0, "2 in-house dependency pins verified"),
    ("stale_lock", 1, "flake.lock rev"),
    ("tag_mismatch", 1, "does not match go.mod"),
    ("missing_input", 1, "no matching input"),
    ("unpinned", 1, "not immutably pinned"),
]

SPEC_DRIFT_CASES = [
    ("broken_json", "@broken@", 3, "could not load live spec"),
    ("missing_endpoint", ["--spec-file", "@testdata@/spec_drift/minimal_live.json"], 1, "problem(s)"),
    ("stale_snapshot_failing", ["--spec-file", "@testdata@/spec_drift/plus_one.json", "--fail-on-drift"], 2, "all endpoints match"),
    ("stale_snapshot_warn_only", ["--spec-file", "@testdata@/spec_drift/plus_one.json"], 0, "all endpoints match"),
]


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument(
        "--repo-root",
        help="SDK repo root (defaults to the parent of this script's directory)",
    )
    parser.add_argument(
        "--scripts-dir",
        help="directory holding pin_sync.py / spec_drift.py (defaults to this script's directory)",
    )
    parser.add_argument(
        "--testdata-dir",
        help="fixture root (defaults to <scripts-dir>/testdata)",
    )
    args = parser.parse_args()

    scripts_dir = (
        pathlib.Path(args.scripts_dir)
        if args.scripts_dir
        else pathlib.Path(__file__).resolve().parent
    )
    testdata = (
        pathlib.Path(args.testdata_dir)
        if args.testdata_dir
        else scripts_dir / "testdata"
    )
    repo_root = (
        pathlib.Path(args.repo_root)
        if args.repo_root
        else scripts_dir.parent
    )

    # The broken-spec fixture is generated, not committed: the check sandbox
    # mounts the source tree read-only, so nothing may be written into it.
    with tempfile.TemporaryDirectory() as tmp:
        broken = pathlib.Path(tmp) / "broken.json"
        broken.write_text("not json {{{")
        return run_all(scripts_dir, testdata, repo_root, broken)


def run_all(
    scripts_dir: pathlib.Path,
    testdata: pathlib.Path,
    repo_root: pathlib.Path,
    broken: pathlib.Path,
) -> int:

    failures: list[str] = []
    ran = 0
    for name, expect_exit, expect_substr in PIN_SYNC_CASES:
        ran += 1
        root = testdata / "pin_sync" / name
        result = subprocess.run(
            [sys.executable, str(scripts_dir / "pin_sync.py"), "--offline", "--root", str(root)],
            capture_output=True,
            text=True,
            check=False,
        )
        output = result.stdout + result.stderr
        if result.returncode != expect_exit:
            failures.append(
                f"pin_sync/{name}: exit {result.returncode}, expected {expect_exit}\n{output.strip()}"
            )
        elif expect_substr not in output:
            failures.append(
                f"pin_sync/{name}: output missing {expect_substr!r}\n{output.strip()}"
            )

    for name, extra, expect_exit, expect_substr in SPEC_DRIFT_CASES:
        ran += 1
        raw = ["--spec-file", str(broken)] if isinstance(extra, str) else extra
        spec_args = [
            a.replace("@testdata@", str(testdata)).replace("@broken@", str(broken))
            for a in raw
        ]
        result = subprocess.run(
            [sys.executable, str(scripts_dir / "spec_drift.py"), "--repo", str(repo_root), *spec_args],
            capture_output=True,
            text=True,
            check=False,
        )
        output = result.stdout + result.stderr
        if result.returncode != expect_exit:
            failures.append(
                f"spec_drift/{name}: exit {result.returncode}, expected {expect_exit}\n{output.strip()}"
            )
        elif expect_substr not in output:
            failures.append(
                f"spec_drift/{name}: output missing {expect_substr!r}\n{output.strip()}"
            )

    for failure in failures:
        print(f"FAIL: {failure}")
    print(f"script fixtures: {ran - len(failures)}/{ran} cases passed")
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())
