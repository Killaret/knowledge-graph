#!/usr/bin/env python3
"""CHUNK-1 acceptance evidence: off-path parity + on-path stats/timing.

Dataset: [{"id","title","content"}] — owner's corpus lives in
work-*/ (gitignored; titles never printed, decision 47).

Measures:
- off parity: `_combined_text(content, title)` must equal the legacy
  `title + " " + content` string for every note — byte-identical model
  input implies cosine 1.0; a sample is also encoded for proof;
- on path: chunk counts, embedding dim, no_content stubs, p50/p95 ms.

Run:  HF_HOME=D:/kg-hf-cache HF_HUB_OFFLINE=1 \
      python nlp-service/scripts/measure_chunk1_flag.py \
      --dataset work-nlp4/notes_dataset.json
"""
import argparse
import json
import os
import sys
import time

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import numpy as np


def cosine(a, b):
    a = np.asarray(a, dtype=np.float64)
    b = np.asarray(b, dtype=np.float64)
    denom = np.linalg.norm(a) * np.linalg.norm(b)
    return float(a @ b / denom) if denom else 0.0


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--dataset", required=True)
    ap.add_argument("--sample", type=int, default=10)
    args = ap.parse_args()

    from app.nlp_utils import (
        _combined_text,
        compute_chunked_embedding,
        get_embedding_model,
    )

    notes = json.load(open(args.dataset, encoding="utf-8"))
    print(f"notes: {len(notes)}")

    # --- off parity: byte-identical input for every note ---
    mismatches = 0
    for n in notes:
        legacy = f"{n['title']} {n['content']}"
        if _combined_text(n["content"], n["title"]) != legacy:
            mismatches += 1
    print(f"off parity (string identity): {len(notes)-mismatches}/{len(notes)} "
          f"identical, mismatches={mismatches}")

    model = get_embedding_model()

    # --- off parity: encoded sample, cosine 1.0 ---
    sample = notes[: args.sample]
    worst = 1.0
    for n in sample:
        legacy = f"{n['title']} {n['content']}"
        a = model.encode(legacy)
        b = model.encode(_combined_text(n["content"], n["title"]))
        worst = min(worst, cosine(a, b))
    print(f"off parity (encoded sample {len(sample)}): min cosine {worst:.6f}")

    # --- on path: stats + timing ---
    counts, times, dims, no_content = [], [], [], 0
    stub_titles = 0
    for n in notes:
        t0 = time.perf_counter()
        vec, n_chunks, nc = compute_chunked_embedding(
            model, n["content"], n["title"]
        )
        times.append((time.perf_counter() - t0) * 1000)
        counts.append(n_chunks)
        dims.append(len(vec))
        if nc:
            no_content += 1
            stub_titles += 1

    counts.sort()
    times_sorted = sorted(times)
    p50 = times_sorted[len(times) // 2]
    p95 = times_sorted[int(len(times) * 0.95)]
    print(f"on: chunks/note min={counts[0]} p50={counts[len(counts)//2]} "
          f"max={counts[-1]}")
    print(f"on: embedding dim set: {sorted(set(dims))}")
    print(f"on: no_content notes: {no_content}")
    print(f"on: latency ms p50={p50:.1f} p95={p95:.1f} "
          f"mean={sum(times)/len(times):.1f}")

    if mismatches or worst < 0.999999 or set(dims) != {384}:
        sys.exit(1)


if __name__ == "__main__":
    main()
