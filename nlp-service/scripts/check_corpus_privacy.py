#!/usr/bin/env python3
"""MODEL-1B privacy gate: verify that files about to be committed contain
no titles from the local notes corpus (decision 47 — private note titles
and texts must not enter the repository).

Usage:
    python nlp-service/scripts/check_corpus_privacy.py \
        --corpus work-nlp4/notes_dataset.json --changed
    python nlp-service/scripts/check_corpus_privacy.py \
        --corpus work-nlp4/notes_dataset.json --files a.md b.py

--changed scans files that git reports as modified/untracked
(`git status --porcelain`). --files scans an explicit list. Exit code is
non-zero when any corpus title (>= 12 chars after whitespace
normalization) is found in a scanned file.
"""
import argparse
import re
import subprocess
import sys
import json
from pathlib import Path

MIN_TITLE = 12


def norm(s: str) -> str:
    return re.sub(r"\s+", " ", s).strip().lower()


def changed_files(repo: Path) -> list[str]:
    out = subprocess.run(
        ["git", "status", "--porcelain"], cwd=repo,
        capture_output=True, text=True, encoding="utf-8").stdout
    files = []
    for line in out.splitlines():
        path = line[3:].strip().strip('"')
        if " -> " in path:
            path = path.split(" -> ")[1]
        files.append(path)
    return files


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--corpus", required=True)
    ap.add_argument("--files", nargs="*", default=[])
    ap.add_argument("--changed", action="store_true")
    args = ap.parse_args()

    repo = Path(__file__).resolve().parents[2]
    corpus = json.load(open(args.corpus, encoding="utf-8"))
    titles = [norm(n["title"]) for n in corpus]
    short = [t for t in titles if len(t) < MIN_TITLE]
    needles = [t for t in titles if len(t) >= MIN_TITLE]

    files = list(args.files)
    if args.changed:
        files += changed_files(repo)
    files = sorted(set(files))
    if not files:
        print("no files to scan", file=sys.stderr)
        return 2

    hits, scanned = [], 0
    for rel in files:
        p = repo / rel
        if not p.is_file() or p.stat().st_size > 4 * 1024 * 1024:
            continue
        try:
            text = norm(p.read_text(encoding="utf-8", errors="ignore"))
        except OSError:
            continue
        scanned += 1
        for t in needles:
            if t in text:
                hits.append((rel, t[:60]))

    print(f"corpus titles: {len(titles)} (skipped {len(short)} shorter "
          f"than {MIN_TITLE} chars); files scanned: {scanned}")
    if hits:
        print(f"FAIL: {len(hits)} title match(es) in tracked files:")
        for rel, t in hits:
            print(f"  {rel}: «{t}»")
        return 1
    print("OK: zero corpus-title matches")
    return 0


if __name__ == "__main__":
    sys.exit(main())
