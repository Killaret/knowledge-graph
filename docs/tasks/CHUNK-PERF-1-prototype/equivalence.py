"""CHUNK-PERF-1: compare a candidate chunker with the reference one.

    git show HEAD:nlp-service/app/core/chunking.py > /tmp/chunking_ref.py
    python docs/tasks/CHUNK-PERF-1-prototype/equivalence.py \
        --reference /tmp/chunking_ref.py --candidate nlp-service/app/core/chunking.py \
        --texts 6000 --seed 7 --counter words

Checks, on random texts with headings, code, tables, abbreviations, initials, numbers, URLs,
closers, "İ", "Σ" and "\r\n":
1. chunk() output is identical (text, heading_path, char_span, token_count, kind, forced_split)
   for three window sizes;
2. if the candidate exposes _ParagraphIndex and _is_false_boundary_at (prototype names), every
   sentence-end point gets the same verdict as the reference, with points in shuffled order.
The counters are monotone (words, chars/3): with them the result must not change at all.
"""
import argparse
import importlib.util
import random
import re
import sys
import time


def load(name, path):
    spec = importlib.util.spec_from_file_location(name, path)
    module = importlib.util.module_from_spec(spec)
    sys.modules[name] = module
    spec.loader.exec_module(module)
    return module


PIECES = [
    "т.е.", "т. е.", "и т.д.", "др.", "см.", "рис.", "e.g.", "i.e.", "etc.", "Dr.", "Mr.", "Fig.",
    "А.С.", "A.B.C.", "а. б.", "x.", "3.14", "1,5", "1.2.3", "12.", "1..2", ",1", "www.", "WWW.",
    "http://a.b", "HTTPS://X.Y/", "hTtP://", "İ", "Σ", "ς", "К", "»", "”", ")", "\"", "'", "]",
    "}", "...", "…", "!?", "?", "!", ".", ".", ".", ". ", "  ", "\n", "\n\n", "\r\n", "\t", "# H",
    "## Sub ##", "```\ncode\n```", "| a | b |", "    indented", "слово", "Word", "word", "x",
    "Y", "1", "a.a.a", "www.www.www.",
    # tokens exactly at the length bounds of the string rules, bare and before a closer
    "a.b.c.d.e", "a.b.c.d.e)", "А.Б.В.Г.Д»", "a.b.c.d.)", "a.b.c.d.e.f", "approx", "approx)",
    "Approx\"", "т. е)", "x)",
]


def random_text(rnd):
    out = []
    for _ in range(rnd.randint(1, 60)):
        piece = rnd.choice(PIECES)
        if rnd.random() < 0.3:
            piece *= rnd.randint(2, 30)  # long tokens without spaces
        out.append(piece)
        out.append(rnd.choice(["", " ", " ", "\n", ""]))
    return "".join(out)


def as_tuples(chunks):
    return [(c.text, tuple(c.heading_path), c.char_span, c.token_count, c.kind, c.forced_split) for c in chunks]


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--reference", required=True)
    ap.add_argument("--candidate", required=True)
    ap.add_argument("--texts", type=int, default=6000)
    ap.add_argument("--seed", type=int, default=7)
    ap.add_argument("--counter", choices=["words", "chars"], default="words")
    args = ap.parse_args()

    ref = load("chunking_reference", args.reference)
    new = load("chunking_candidate", args.candidate)
    if args.counter == "words":
        count = lambda s: len(re.findall(r"\w+", s))
    else:
        count = lambda s: (len(s) + 2) // 3
    per_point = hasattr(new, "_ParagraphIndex") and hasattr(new, "_is_false_boundary_at")

    rnd = random.Random(args.seed)
    mismatches = points = 0
    started = time.time()
    for _ in range(args.texts):
        text = random_text(rnd)
        title = rnd.choice([None, "Заметка"])
        for target, window in ((8, 12), (126, 126), (3, 5)):
            a = ref.chunk(text, ref.ChunkingParams(token_counter=count, target_tokens=target, max_tokens=window, title=title))
            b = new.chunk(text, new.ChunkingParams(token_counter=count, target_tokens=target, max_tokens=window, title=title))
            if as_tuples(a) != as_tuples(b):
                mismatches += 1
                print("chunk() differs:", ascii(text)[:300])
                break
        if per_point:
            index = new._ParagraphIndex(text, 0, len(text))
            ends = [i for i, ch in enumerate(text) if ch in ref._SENTENCE_END]
            rnd.shuffle(ends)
            for i in ends:
                points += 1
                if ref._is_false_boundary(text, 0, i) != new._is_false_boundary_at(text, 0, i, index):
                    mismatches += 1
                    print("verdict differs at", i, ascii(text[max(0, i - 40) : i + 5]))
                    break
        if mismatches > 5:
            break
    print(f"texts={args.texts} points={points} mismatches={mismatches} counter={args.counter} in {time.time() - started:.1f}s")
    sys.exit(1 if mismatches else 0)


if __name__ == "__main__":
    main()
