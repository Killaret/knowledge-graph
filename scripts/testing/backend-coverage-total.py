#!/usr/bin/env python3
"""Compute backend unit coverage from cover.out, excluding non-unit packages.

The cover.out profile is produced by `go test -coverprofile`. This script
filters out packages that are not part of the unit-testable denominator
(main entrypoints, generated clients, test helpers, scripts) and prints the
result. It exits with a non-zero status when the total is below the threshold.
"""

import os
import re
import sys
from pathlib import Path


def load_excludes(script_dir: Path) -> list[str]:
    """Load the list of package prefixes to exclude from the denominator."""
    excludes_file = script_dir / "backend-coverage-excludes.txt"
    if not excludes_file.exists():
        return []
    excludes = []
    for line in excludes_file.read_text(encoding="utf-8").splitlines():
        line = line.split("#", 1)[0].strip()
        if line:
            excludes.append(line)
    return excludes


def compute_coverage(cover_path: Path, excludes: list[str]) -> float:
    """Return the statement coverage percentage for non-excluded packages."""
    package_pattern = re.compile(r"^(.+)/[^/]+\.go:")
    total_statements = 0
    covered_statements = 0

    with cover_path.open("r", encoding="utf-8") as f:
        for line in f:
            if line.startswith("mode:"):
                continue
            parts = line.rsplit(" ", 2)
            if len(parts) != 3:
                continue
            loc, stmts_str, count_str = parts
            try:
                stmts = int(stmts_str)
                count = int(count_str)
            except ValueError:
                continue

            match = package_pattern.match(loc)
            if not match:
                continue
            package = match.group(1)
            if any(package == e or package.startswith(e + "/") for e in excludes):
                continue

            total_statements += stmts
            if count > 0:
                covered_statements += stmts

    if total_statements == 0:
        return 0.0
    return (covered_statements / total_statements) * 100


def main() -> int:
    if len(sys.argv) < 3:
        print("Usage: backend-coverage-total.py <cover.out> <threshold>", file=sys.stderr)
        return 2

    cover_path = Path(sys.argv[1])
    if not cover_path.is_absolute():
        cover_path = Path.cwd() / cover_path

    try:
        threshold = float(sys.argv[2])
    except ValueError:
        print(f"Invalid threshold: {sys.argv[2]}", file=sys.stderr)
        return 2

    script_dir = Path(__file__).resolve().parent
    excludes = load_excludes(script_dir)

    actual = compute_coverage(cover_path, excludes)
    status = "PASS" if actual >= threshold else "FAIL"
    print(f"Backend unit coverage: {actual:.2f}% (required >= {threshold}%) [{status}]")
    if actual < threshold:
        excluded = ", ".join(excludes) if excludes else "none"
        print(f"Excluded packages: {excluded}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
