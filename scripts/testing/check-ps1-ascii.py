#!/usr/bin/env python3
"""Fail if any .ps1 under scripts/ contains a non-ASCII byte without a UTF-8 BOM."""

import os
import sys
from pathlib import Path


def main() -> int:
    root = Path(__file__).resolve().parents[2] / "scripts"
    failures = []
    for p in root.rglob("*.ps1"):
        data = p.read_bytes()
        has_bom = data[:3] == b"\xef\xbb\xbf"
        body = data[3:] if has_bom else data
        for i, b in enumerate(body):
            if b > 127:
                failures.append((p, i, has_bom))
                break

    if failures:
        print("Non-ASCII bytes found in .ps1 files without UTF-8 BOM:")
        for p, offset, has_bom in failures:
            print(f"  {p} at byte {offset} (BOM={has_bom})")
        return 1

    print("All .ps1 files under scripts/ are ASCII or have a UTF-8 BOM.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
