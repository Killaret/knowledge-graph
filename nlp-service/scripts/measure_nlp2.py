"""NLP-2 measurement harness: yake vs keybert-naive vs hybrid.

Runs over the exported note dataset (work-w1/dataset.json — real notes,
content field), times each extractor for p50/p95 latency, prints top-10
side-by-side for the 20 longest notes, and counts notes that end up with
zero keywords under each extractor.

Usage: python scripts/measure_nlp2.py [dataset.json] [out_dir]
"""

import json
import os
import statistics
import sys
import time

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import nltk  # noqa: E402

for res in ("tokenizers/punkt", "corpora/stopwords", "corpora/wordnet"):
    try:
        nltk.data.find(res)
    except LookupError:
        nltk.download(res.split("/")[1])

STOP = set(
    nltk.corpus.stopwords.words("russian") + nltk.corpus.stopwords.words("english")
)


def pct(values, q):
    if not values:
        return 0.0
    values = sorted(values)
    idx = min(len(values) - 1, int(round(q / 100 * (len(values) - 1))))
    return values[idx]


def main():
    dataset_path = sys.argv[1] if len(sys.argv) > 1 else "../work-w1/dataset.json"
    out_dir = sys.argv[2] if len(sys.argv) > 2 else "../work-nlp2"
    os.makedirs(out_dir, exist_ok=True)

    data = json.load(open(dataset_path, encoding="utf-8"))
    if isinstance(data, dict):
        data = data.get("notes", [])
    notes = [n for n in data if (n.get("content") or "").strip()]
    print(f"notes with content: {len(notes)}")

    import yake
    from keybert import KeyBERT

    from app.nlp_utils import get_embedding_model, extract_keywords

    model = get_embedding_model()
    kb = KeyBERT(model=model)

    def run_yake(text, top=10):
        kw = yake.KeywordExtractor(lan="ru", top=top, stopwords=STOP)
        return [(w, max(0.0, min(1.0, 1.0 - s))) for w, s in kw.extract_keywords(text)]

    def run_keybert_naive(text, top=10):
        pairs = kb.extract_keywords(
            text, keyphrase_ngram_range=(1, 2), stop_words=list(STOP), top_n=top
        )
        return [(w, max(0.0, min(1.0, s))) for w, s in pairs]

    def run_hybrid(text, top=10):
        return [(lemma, w) for lemma, _surface, w in extract_keywords(text, top)]

    runners = [
        ("yake", run_yake),
        ("keybert-naive", run_keybert_naive),
        ("hybrid", run_hybrid),
    ]

    # Warm-up once so model init doesn't skew the first sample.
    run_hybrid(notes[0]["content"])

    timings = {name: [] for name, _ in runners}
    empty_counts = {name: 0 for name, _ in runners}
    longest = sorted(notes, key=lambda n: len(n["content"]), reverse=True)[:20]
    table = []

    for note in notes:
        text = note["content"]
        for name, fn in runners:
            t0 = time.perf_counter()
            res = fn(text)
            timings[name].append(time.perf_counter() - t0)
            if not res:
                empty_counts[name] += 1
            if note in longest:
                table_entry = None
                for e in table:
                    if e["title"] == note.get("title", ""):
                        table_entry = e
                        break
                if table_entry is None:
                    table_entry = {"title": note.get("title", ""), "cols": {}}
                    table.append(table_entry)
                table_entry["cols"][name] = [w for w, _ in res[:10]]

    raw = {
        "notes": len(notes),
        "timings_ms": {k: [round(t * 1000, 1) for t in v] for k, v in timings.items()},
        "empty_counts": empty_counts,
        "top10": table,
    }
    with open(os.path.join(out_dir, "raw.json"), "w", encoding="utf-8") as f:
        json.dump(raw, f, ensure_ascii=False, indent=1)

    out = open(os.path.join(out_dir, "measurements.md"), "w", encoding="utf-8")
    out.write("# NLP-2 замеры экстракторов\n\n")
    out.write(f"Датасет: {dataset_path}, заметок с текстом: {len(notes)}\n\n")

    out.write("## Латентность (мс, полный корпус)\n\n")
    out.write("| экстрактор | p50 | p95 | mean | max |\n|---|---|---|---|---|\n")
    for name, _ in runners:
        ts = [t * 1000 for t in timings[name]]
        out.write(
            f"| {name} | {pct(ts, 50):.0f} | {pct(ts, 95):.0f} | "
            f"{statistics.mean(ts):.0f} | {max(ts):.0f} |\n"
        )

    out.write("\n## Заметки без ключевых слов\n\n")
    out.write("| экстрактор | пустых | доля |\n|---|---|---|\n")
    for name, _ in runners:
        c = empty_counts[name]
        out.write(f"| {name} | {c} | {c / len(notes):.1%} |\n")

    out.write("\n## Top-10 рядом (20 самых длинных заметок)\n")
    for e in table:
        out.write(f"\n### {e['title'][:70]}\n\n")
        out.write("| # | yake | keybert-naive | hybrid |\n|---|---|---|---|\n")
        for i in range(10):
            cells = []
            for name, _ in runners:
                col = e["cols"].get(name, [])
                cells.append(col[i] if i < len(col) else "—")
            out.write(f"| {i + 1} | " + " | ".join(cells) + " |\n")
    out.close()

    print(f"written: {out_dir}/measurements.md, raw.json")
    for name, _ in runners:
        ts = [t * 1000 for t in timings[name]]
        print(
            f"{name}: p50={pct(ts, 50):.0f}ms p95={pct(ts, 95):.0f}ms "
            f"empty={empty_counts[name]}"
        )


if __name__ == "__main__":
    main()
