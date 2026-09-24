#!/usr/bin/env python3
"""MODEL-1B supplementary corpus: ~2000 public documents with category
labels, shaped like the local notes corpus ({id,title,content} +
{notes:[{id,folder}]}) so measure_model1b.py can run on it unchanged.

Sources (all public HF datasets, no private data):
  - mteb/mlsum "ru" train   — Russian news articles, folder = topic
  - SetFit/bbc-news         — English news, folder = label_text
  - mteb/GeoreviewClusteringP2P — Russian reviews, folder = category

Overlapping topics are mapped to shared folder names (sport, politics,
business, science/tech, entertainment) so cross-lingual same-topic links
count as correct — multilingual models should get that credit.

Usage:
    python nlp-service/scripts/fetch_public_corpus.py \
        --out-corpus work-model1b/local/public-corpus.json \
        --out-folders work-model1b/local/public-folders.json \
        [--per-source 700] [--seed 20260924]
"""
import argparse
import collections
import json
import random
from pathlib import Path

MAX_CHARS = 4000  # keep documents note-sized

TOPIC_MAP = {
    # mlsum ru topics -> shared folder names
    "sport": "sport", "sports": "sport",
    "politics": "politics",
    "economics": "business", "business": "business",
    "science": "science/tech", "tech": "science/tech",
    "technology": "science/tech",
    "culture": "entertainment", "entertainment": "entertainment",
}


def stratified(records, folder_of, per_source, rng):
    """Sample up to per_source records, round-robin across categories so
    the corpus keeps folder diversity."""
    by_cat = collections.defaultdict(list)
    for r in records:
        by_cat[folder_of(r)].append(r)
    for v in by_cat.values():
        rng.shuffle(v)
    picked = []
    pools = list(by_cat.values())
    while len(picked) < per_source and any(pools):
        for p in pools:
            if p and len(picked) < per_source:
                picked.append(p.pop())
    return picked


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--out-corpus", required=True)
    ap.add_argument("--out-folders", required=True)
    ap.add_argument("--per-source", type=int, default=700)
    ap.add_argument("--seed", type=int, default=20260924)
    args = ap.parse_args()

    from datasets import load_dataset
    rng = random.Random(args.seed)
    corpus, folders = [], []

    def add(recs, folder_of, title_of, text_of, prefix):
        for i, r in enumerate(stratified(recs, folder_of, args.per_source, rng)):
            nid = f"{prefix}{i:04d}"
            corpus.append({"id": nid, "title": title_of(r),
                           "content": text_of(r)[:MAX_CHARS]})
            folders.append({"id": nid, "folder": folder_of(r)})

    mlsum = load_dataset("mteb/mlsum", "ru", split="train")
    add(list(mlsum),
        lambda r: TOPIC_MAP.get(str(r["topic"]).lower(), f"mlsum:{r['topic']}"),
        lambda r: str(r["title"]), lambda r: str(r["text"]), "mlsum-")

    bbc = load_dataset("SetFit/bbc-news", split="train").to_list() + \
        load_dataset("SetFit/bbc-news", split="test").to_list()
    add(bbc, lambda r: str(r["label_text"]), lambda r: "",
        lambda r: str(r["text"]), "bbc-")

    geo = load_dataset("mteb/GeoreviewClusteringP2P", split="test")
    add(list(geo), lambda r: f"geo:{r['labels']}", lambda r: "",
        lambda r: str(r["sentences"]), "geo-")

    Path(args.out_corpus).parent.mkdir(parents=True, exist_ok=True)
    Path(args.out_corpus).write_text(
        json.dumps(corpus, ensure_ascii=False), encoding="utf-8")
    Path(args.out_folders).write_text(
        json.dumps({"notes": folders}, ensure_ascii=False), encoding="utf-8")

    dist = collections.Counter(f["folder"] for f in folders)
    print(f"corpus: {len(corpus)} docs, {len(dist)} folders")
    for k, v in dist.most_common():
        print(f"  {v:5d}  {k}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
