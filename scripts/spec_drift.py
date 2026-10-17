#!/usr/bin/env python3
"""Check the SDK's API surface against the live Wise API reference.

The SDK vendored an OpenAPI 3.1 snapshot (``docs/reviews/wise-api-openapi.json``)
that goes stale as Wise rolls spec versions forward (2026Q3 -> 2027Q1 happened
without any signal). Two questions drift independently, so this script answers
both:

1. **Does the SDK still match the live API?** Every ``(method, path)`` pair the
   SDK's client helpers call is extracted from the root ``*.go`` files and
   matched (structurally, version-prefix-stripped) against the live spec. A
   missing path or dropped method is a breaking drift for this SDK -> exit 1.

2. **How far has the vendored snapshot drifted?** Added/removed/changed paths
   between the vendored snapshot and the live spec are reported (informational
   by default, ``--fail-on-drift`` to enforce). New paths are the
   demand-gated expansion candidates reviewed against FEATURES.md.

Exit codes: 0 = SDK matches live spec; 1 = SDK endpoint missing/mismatched;
2 = ``--fail-on-drift`` and the vendored snapshot is stale; 3 = fetch/parse
failure or no SDK endpoints extracted (a silent-skip would defeat the check).
"""

from __future__ import annotations

import argparse
import json
import os
import pathlib
import re
import subprocess  # nosec B404 -- shells to the repo's own go/parser extractor only
import sys
import tempfile
import urllib.error
import urllib.request

LIVE_SPEC_URL = (
    "https://docs.wise.com/_bundle/api-reference/@preview/index.json?download"
)
DEFAULT_VENDORED = "docs/reviews/wise-api-openapi.json"

HTTP_VERBS = {"get", "post", "put", "patch", "delete", "head", "options", "trace"}

CLIENT_CALL_RE = re.compile(
    r"(?:\b(?:c|client)\.(?:get|getWithQuery|getWithQueryHeaders|getRaw|post|postWithHeaders|put|patch|delete)\b)"
    r"|\b(?:fetchByID|listTransfersPage)\("
)
SPRINTF_PATH_RE = re.compile(r'fmt\.Sprintf\(\s*"(/[^"]*)"')
SPRINTF_OPEN_RE = re.compile(r"path\s*:?=\s*fmt\.Sprintf\(\s*$")
LITERAL_PATH_RE = re.compile(r'path\s*:?=\s*"(/[^"]*)"')
CALL_WINDOW_PATH_RE = re.compile(r'"(/[^"]*)"')
FORMAT_VERB_RE = re.compile(r"%[-#0 +]*\d*(?:\.\d+)?[a-zA-Z]")
VERSION_PREFIX_RE = re.compile(r"^/(?:v\d+|\d{4}Q[1-4]|\{s\})(?=/)")

KNOWN_EXEMPTIONS = {
    "/profiles/{d}/balance-statements/{d}/statement.{s}": "spec declares only the .json variant (same exemption as the spec-conformance gate)",
}

# The Go extractor lives under testdata/ so the go toolchain ignores it as a
# module package (no build, lint, or coverage impact on the SDK module); it is
# invoked file-mode via `go run`, stdlib-only. In the nix store the script is
# copied alone, so the flake app passes the store path via env.
GO_EXTRACTOR = pathlib.Path(
    os.environ.get("WISE_GO_SPEC_DRIFT_EXTRACTOR")
    or (pathlib.Path(__file__).resolve().parent / "testdata" / "specdriftext" / "main.go")
)

# Extraction floor: a call-shape change (new wrapper idiom the regexes do not
# know) silently shrinks the extracted surface and the check would then pass
# while comparing far fewer endpoints than the SDK really has. 43 endpoints at
# the 46-method surface (2026-10-16); fail well below that, and only ever
# tighter as the surface grows. --min-endpoints overrides it for intentional
# surface changes.
EXTRACTION_FLOOR = 40


def extract_sdk_endpoints_go(repo: pathlib.Path) -> list[dict]:
    """Run the go/parser extractor; a failed run is extraction-broken (exit 3)."""
    if not GO_EXTRACTOR.exists():
        raise RuntimeError(f"go extractor missing: {GO_EXTRACTOR}")
    with tempfile.TemporaryDirectory() as tmp:
        env = {
            "GOCACHE": str(pathlib.Path(tmp) / "gocache"),
            "GOPATH": str(pathlib.Path(tmp) / "gopath"),
            "GOTOOLCHAIN": "local",
            "GOFLAGS": "",
            "HOME": str(pathlib.Path(tmp)),
            "PATH": os.environ.get("PATH", "/usr/bin:/bin"),
        }
        proc = subprocess.run(  # nosec B603 -- fixed argv, repo-owned tool, stdlib-only
            ["go", "run", str(GO_EXTRACTOR), str(repo)],
            capture_output=True,
            text=True,
            check=False,
            env=env,
            timeout=300,
        )
        if proc.returncode != 0:
            raise RuntimeError(
                f"go extractor failed (exit {proc.returncode}): {proc.stderr.strip()[:500]}"
            )
    endpoints = json.loads(proc.stdout)
    if not isinstance(endpoints, list):
        raise RuntimeError("go extractor returned non-list output")
    return endpoints


def load_spec(spec_file: str | None, url: str) -> dict:
    if spec_file:
        with open(spec_file, encoding="utf-8") as f:
            return json.load(f)
    req = urllib.request.Request(
        url,
        headers={
            "Accept": "application/json",
            "User-Agent": "wise-go-spec-drift/0.1",
        },
    )
    with urllib.request.urlopen(req, timeout=60) as resp:  # nosec B310 -- URL is the repo-pinned https Wise OpenAPI endpoint
        return json.loads(resp.read().decode("utf-8"))


def extract_sdk_endpoints(repo: pathlib.Path) -> list[dict]:
    """Extract (method, raw path template) pairs from root *.go client files."""
    endpoints: list[dict] = []
    for go_file in sorted(repo.glob("*.go")):
        if go_file.name.endswith("_test.go"):
            continue
        lines = go_file.read_text(encoding="utf-8").splitlines()
        last_path: str | None = None
        enclosing = ""
        pending: list[dict] = []
        for lineno, line in enumerate(lines, 1):
            func_def = FUNC_DEF_RE.match(line)
            if func_def:
                enclosing = func_def.group(1)
            m = SPRINTF_PATH_RE.search(line) or LITERAL_PATH_RE.search(line)
            if m:
                last_path = m.group(1)
            elif SPRINTF_OPEN_RE.search(line):
                ahead = CALL_WINDOW_PATH_RE.search(
                    "\n".join(lines[lineno : lineno + 2])
                )
                if ahead:
                    last_path = ahead.group(1)
            method = request_method(line)
            if method is None:
                if REQUEST_CALL_RE.search(line):
                    window = "\n".join(lines[lineno - 1 : lineno + 3])
                    inline = CALL_WINDOW_PATH_RE.search(window)
                    template = inline.group(1) if inline else last_path
                    if template:
                        literal = HTTP_METHOD_RE.search(window)
                        entry = {
                            "method": literal.group(1).upper() if literal else None,
                            "template": template,
                            "file": go_file.name,
                            "line": lineno,
                            "enclosing": enclosing,
                        }
                        if entry["method"]:
                            endpoints.append(entry)
                        else:
                            pending.append(entry)
                continue
            window = "\n".join(lines[lineno - 1 : lineno + 3])
            inline = CALL_WINDOW_PATH_RE.search(window)
            if inline:
                template = inline.group(1)
            elif last_path is not None:
                template = last_path
            else:
                template = None
            if template:
                endpoints.append(
                    {
                        "method": method,
                        "template": template,
                        "file": go_file.name,
                        "line": lineno,
                    }
                )
        for entry in pending:
            methods: set[str] = set()
            for i, line in enumerate(lines):
                if re.search(r"\b" + re.escape(entry["enclosing"]) + r"\(", line):
                    window = "\n".join(lines[i : i + WRAPPER_WINDOW_LINES])
                    methods.update(
                        mm.group(1).upper() for mm in HTTP_METHOD_RE.finditer(window)
                    )
            for resolved in sorted(methods) or ["UNRESOLVED"]:
                endpoints.append({**entry, "method": resolved})
    return endpoints


def request_method(line: str) -> str | None:
    if CLIENT_CALL_RE.search(line) is None:
        return None
    if "fetchByID" in line or "listTransfersPage" in line:
        return "GET"
    for fragment, method in (
        ("getWithQueryHeaders", "GET"),
        ("getWithQuery", "GET"),
        ("getRaw", "GET"),
        ("get", "GET"),
        ("postWithHeaders", "POST"),
        ("post", "POST"),
        ("put", "PUT"),
        ("patch", "PATCH"),
        ("delete", "DELETE"),
    ):
        if re.search(r"\b(?:c|client)\." + fragment + r"\b", line):
            return method
    return None


FUNC_DEF_RE = re.compile(r"^func (?:\([^)]*\) )?([A-Za-z_][A-Za-z0-9_]*)")
HTTP_METHOD_RE = re.compile(r"http\.Method([A-Za-z]+)\b")
REQUEST_CALL_RE = re.compile(r"\bc\.request\(")
WRAPPER_WINDOW_LINES = 9


def format_verb_marker(m: re.Match) -> str:
    return "{d}" if m.group(0).endswith("d") else "{s}"


def normalize_sdk_path(template: str) -> str:
    path = template.split("?", 1)[0]
    path = FORMAT_VERB_RE.sub(format_verb_marker, path)
    return VERSION_PREFIX_RE.sub("", path)


def normalize_spec_path(path: str) -> str:
    return re.sub(r"\{[^}]*\}", "{p}", path)


def segments_match(a: str, b: str) -> bool:
    if a == b:
        return True
    if a == "{p}" or b == "{p}":
        return "{d}" in (a, b) or "{s}" in (a, b) or a == b
    return a == "{s}" or b == "{s}"


def build_spec_index(spec: dict) -> dict[str, set[str]]:
    index: dict[str, set[str]] = {}
    for path, item in spec.get("paths", {}).items():
        if not isinstance(item, dict):
            continue
        index[path] = {m.upper() for m in item if m in HTTP_VERBS}
    return index


def match_endpoint(
    index: dict[str, set[str]], method: str, sdk_norm: str
) -> tuple[str, list[str]]:
    """Return (verdict, matched spec paths): 'ok' | 'method' | 'missing'.

    A numeric SDK param ({d}) only matches a spec parameter, never a literal
    segment, so a removed ID endpoint is not silently absorbed by a literal
    sibling path like /balances/hold-limit-breach.
    """
    exact = index.get(sdk_norm)
    if exact is not None:
        return ("ok", [sdk_norm]) if method in exact else ("method", [sdk_norm])
    aligned: list[str] = []
    loose: list[str] = []
    for path in index:
        spec_norm = normalize_spec_path(path)
        a, b = sdk_norm.split("/"), spec_norm.split("/")
        if len(a) != len(b):
            continue
        pairs = list(zip(a, b))
        if not all(segments_match(x, y) for x, y in pairs):
            continue
        if all(x == y or "{p}" in (x, y) for x, y in pairs):
            aligned.append(path)
        else:
            loose.append(path)
    structural = aligned or loose
    if not structural:
        return "missing", []
    if any(method in index[path] for path in structural):
        return "ok", structural
    return "method", structural


def spec_label(spec: dict) -> str:
    info = spec.get("info", {})
    version = info.get("version") or "(no version)"
    return f"OpenAPI {spec.get('openapi', '?')}, {version}, {len(spec.get('paths', {}))} paths"


def spec_drift(vendored: dict, live: dict) -> dict:
    v = build_spec_index(vendored)
    l = build_spec_index(live)
    return {
        "added": sorted(set(l) - set(v)),
        "removed": sorted(set(v) - set(l)),
        "changed": sorted(p for p in set(v) & set(l) if v[p] != l[p]),
        "vendored_version": vendored.get("info", {}).get("version") or "(no version)",
        "live_version": live.get("info", {}).get("version") or "(no version)",
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--url", default=LIVE_SPEC_URL, help="live spec URL to fetch")
    parser.add_argument(
        "--spec-file", help="use a local spec JSON instead of downloading"
    )
    parser.add_argument(
        "--repo",
        help="repo root (default: the git toplevel containing this script)",
    )
    parser.add_argument(
        "--fail-on-drift",
        action="store_true",
        help="exit 2 when the vendored snapshot is stale",
    )
    parser.add_argument(
        "--min-endpoints",
        type=int,
        default=None,
        metavar="N",
        help="override the extraction floor (intentional surface changes)",
    )
    parser.add_argument(
        "--extractor",
        choices=("go", "regex"),
        default="go",
        help="endpoint extractor: go (AST-based, default) or regex (legacy, parity tests)",
    )
    parser.add_argument(
        "--verbose",
        action="store_true",
        help="list every extracted endpoint and its match",
    )
    args = parser.parse_args()

    repo = (
        pathlib.Path(args.repo)
        if args.repo
        else pathlib.Path(__file__).resolve().parent.parent
    )
    try:
        live = load_spec(args.spec_file, args.url)
    except (OSError, json.JSONDecodeError, urllib.error.URLError) as err:
        print(f"FAIL: could not load live spec: {err}", file=sys.stderr)
        return 3
    if not isinstance(live.get("paths"), dict) or not live["paths"]:
        print(
            "FAIL: live spec has no paths; refusing to judge on a broken document",
            file=sys.stderr,
        )
        return 3
    index = build_spec_index(live)

    try:
        if args.extractor == "go":
            endpoints = extract_sdk_endpoints_go(repo)
        else:
            endpoints = extract_sdk_endpoints(repo)
    except (RuntimeError, json.JSONDecodeError, OSError) as err:
        print(f"FAIL: SDK endpoint extraction broken: {err}", file=sys.stderr)
        return 3
    unique = {(e["method"], normalize_sdk_path(e["template"])) for e in endpoints}
    if not unique:
        print(
            f"FAIL: extracted 0 SDK endpoints from {repo}/*.go; extraction is broken",
            file=sys.stderr,
        )
        return 3

    print(f"Live spec:  {spec_label(live)}")
    print(
        f"SDK surface: {len(unique)} unique (method, path) endpoints extracted from root *.go files"
    )

    unresolved = sorted(norm for method, norm in unique if method == "UNRESOLVED")
    unique = {(method, norm) for method, norm in unique if method != "UNRESOLVED"}

    # Extraction-floor guard: see EXTRACTION_FLOOR above the module constants.
    floor = args.min_endpoints if args.min_endpoints is not None else EXTRACTION_FLOOR
    if len(unique) < floor:
        print(
            f"FAIL: extraction floor broken: {len(unique)} resolved endpoints "
            f"extracted, expected >= {floor}; the SDK grew or the "
            f"extractor lost call shapes (check new request-helper idioms)",
            file=sys.stderr,
        )
        return 3

    failures: list[str] = []
    for method, norm in sorted(unique):
        verdict, matched = match_endpoint(index, method, norm)
        if verdict == "missing" and norm in KNOWN_EXEMPTIONS:
            if args.verbose:
                print(f"  [EXEMPT] {method} {norm} - {KNOWN_EXEMPTIONS[norm]}")
            continue
        if args.verbose:
            tag = "OK" if verdict == "ok" else "PROBLEM"
            print(
                f"  [{tag}] {method} {norm}"
                + (f"  ->  {', '.join(matched)}" if verdict != "ok" else "")
            )
        if verdict == "missing":
            failures.append(
                f"MISSING: {method} {norm} matches no path in the live spec"
            )
        elif verdict == "method":
            offered = (
                sorted(set().union(*(index[p] for p in matched))) if matched else []
            )
            failures.append(
                f"METHOD GONE: {method} {norm} - live spec offers {offered} on {matched}"
            )
    if unresolved:
        print(
            f"\nWARNING: {len(unresolved)} endpoint(s) with unresolved HTTP method (wrapper indirection):"
        )
        for norm in unresolved:
            print(f"  UNRESOLVED: {norm} - not verified against the live spec")
    if failures:
        print(f"\nSDK vs live spec: {len(failures)} problem(s)")
        for failure in failures:
            print(f"  {failure}")
        status = 1
    else:
        print("SDK vs live spec: all endpoints match")
        status = 0

    vendored_file = repo / DEFAULT_VENDORED
    if vendored_file.exists():
        vendored = load_spec(str(vendored_file), args.url)
        drift = spec_drift(vendored, live)
        print(f"\nVendored snapshot: {spec_label(vendored)}")
        print(
            f"Drift vs live: +{len(drift['added'])} added / -{len(drift['removed'])} removed / ~{len(drift['changed'])} changed paths"
        )
        if args.verbose:
            for label, entries in (
                ("added", drift["added"]),
                ("removed", drift["removed"]),
                ("changed methods", drift["changed"]),
            ):
                for entry in entries:
                    print(f"  {label}: {entry}")
        if drift["added"] or drift["removed"] or drift["changed"]:
            print(
                "Re-vendor docs/reviews/wise-api-openapi.json when adopting; new paths are expansion candidates for FEATURES.md"
            )
            if args.fail_on_drift and status == 0:
                status = 2
    else:
        print(
            f"\nVendored snapshot not found at {vendored_file}; skipping drift report"
        )

    return status


if __name__ == "__main__":
    sys.exit(main())
