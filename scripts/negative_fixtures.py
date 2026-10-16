#!/usr/bin/env python3
"""Negative fixtures for scripts/pin_sync.py and scripts/spec_drift.py.

Every fixture here encodes a failure mode the scripts once shipped with or
could regress into: the pin-sync quartet (flake.lock drifting from flake.nix,
a flake tag lagging go.mod, a go.mod dep with no flake input at all, an
unpinned input) plus the consistent baseline, and the spec-drift quartet
(unparseable live spec, an SDK path the spec no longer has, a stale vendored
snapshot under --fail-on-drift, and the same drift in warn-only mode).

All fixtures are generated into a tempdir at run time, never committed: the
repo-scan tooling treats a committed fixture go.mod as a real module (go mod
tidy wants to "fix" it), and the check sandbox mounts the source tree
read-only. Wired into `nix flake check` as checks.script-fixtures; runnable
directly:

    python3 scripts/negative_fixtures.py
"""

from __future__ import annotations

import argparse
import pathlib
import subprocess
import sys
import tempfile

REV_A = "a" * 40
REV_B = "b" * 40

GO_MOD = """module fixture

go 1.26

require (
\tgithub.com/larsartmann/go-retry v0.7.1
)
"""

FLAKE_LOCK_FMT = (
    '{{"nodes":{{"root":{{"inputs":{{}}}},"go-retry":{{"locked":{{"type":"github",'
    '"owner":"LarsArtmann","repo":"go-retry","rev":"{rev}"}}}}}},"version":7}}\n'
)

# name -> (go.mod, flake.nix, flake.lock rev, expected exit, expected substring)
PIN_SYNC_CASES = {
    "ok": (
        GO_MOD,
        '{\n  input-retry = "github:LarsArtmann/go-retry?rev=%s";\n}\n' % REV_A,
        FLAKE_LOCK_FMT.format(rev=REV_A),
        0,
        "1 in-house dependency pins verified",
    ),
    "stale_lock": (
        GO_MOD,
        '{\n  input-retry = "github:LarsArtmann/go-retry?rev=%s";\n}\n' % REV_B,
        FLAKE_LOCK_FMT.format(rev=REV_A),
        1,
        "flake.lock rev",
    ),
    "tag_mismatch": (
        GO_MOD,
        '{\n  input-retry = "github:LarsArtmann/go-retry/v0.6.0";\n}\n',
        FLAKE_LOCK_FMT.format(rev=REV_A),
        1,
        "does not match go.mod",
    ),
    "missing_input": (
        GO_MOD,
        '{\n  input-unrelated = "github:LarsArtmann/some-other-repo?rev=%s";\n}\n'
        % REV_A,
        FLAKE_LOCK_FMT.format(rev=REV_A),
        1,
        "no matching input",
    ),
    "unpinned": (
        GO_MOD,
        '{\n  input-retry = "github:LarsArtmann/go-retry";\n}\n',
        FLAKE_LOCK_FMT.format(rev=REV_A),
        1,
        "not immutably pinned",
    ),
}

# name -> (inline spec content or None, extra args, expected exit, expected substring)
SPEC_DRIFT_CASES = [
    (
        "broken_json",
        "not json {{{",
        ["--spec-file", "@broken@"],
        3,
        "could not load live spec",
    ),
    ("missing_endpoint", None, ["--spec-file", "@minimal@"], 1, "problem(s)"),
    (
        "stale_snapshot_failing",
        None,
        [
            "--spec-file",
            "@plus_one@",
            "--fail-on-drift",
        ],
        2,
        "all endpoints match",
    ),
    (
        "stale_snapshot_warn_only",
        None,
        ["--spec-file", "@plus_one@"],
        0,
        "all endpoints match",
    ),
]


def write_pin_sync_fixtures(root: pathlib.Path) -> None:
    for name, (go_mod, flake, lock, _exit, _substr) in PIN_SYNC_CASES.items():
        case = root / name
        case.mkdir(parents=True)
        (case / "go.mod").write_text(go_mod)
        (case / "flake.nix").write_text(flake)
        (case / "flake.lock").write_text(lock)


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
        help="spec-drift fixture root (defaults to <scripts-dir>/testdata)",
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
    repo_root = pathlib.Path(args.repo_root) if args.repo_root else scripts_dir.parent

    with tempfile.TemporaryDirectory() as tmp:
        tmpdir = pathlib.Path(tmp)
        pin_root = tmpdir / "pin_sync"
        write_pin_sync_fixtures(pin_root)
        minimal = tmpdir / "minimal_live.json"
        minimal.write_text((testdata / "spec_drift" / "minimal_live.json").read_text())
        return run_all(scripts_dir, pin_root, minimal, testdata, repo_root)


def run_all(
    scripts_dir: pathlib.Path,
    pin_root: pathlib.Path,
    minimal: pathlib.Path,
    testdata: pathlib.Path,
    repo_root: pathlib.Path,
) -> int:
    failures: list[str] = []
    ran = 0
    for name, (
        go_mod,
        flake,
        lock,
        expect_exit,
        expect_substr,
    ) in PIN_SYNC_CASES.items():
        ran += 1
        result = subprocess.run(
            [
                sys.executable,
                str(scripts_dir / "pin_sync.py"),
                "--offline",
                "--root",
                str(pin_root / name),
            ],
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

    for name, inline_spec, extra, expect_exit, expect_substr in SPEC_DRIFT_CASES:
        ran += 1
        spec_file = minimal
        if inline_spec is not None:
            spec_file = minimal.with_name(f"{name}.json")
            spec_file.write_text(inline_spec)
        spec_args = [
            a.replace("@minimal@", str(minimal))
            .replace("@broken@", str(spec_file))
            .replace("@plus_one@", str(testdata / "spec_drift" / "plus_one.json"))
            for a in extra
        ]
        result = subprocess.run(
            [
                sys.executable,
                str(scripts_dir / "spec_drift.py"),
                "--repo",
                str(repo_root),
                *spec_args,
            ],
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
