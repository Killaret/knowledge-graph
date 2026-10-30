#!/usr/bin/env python3
"""NLP-4: measure iterative note normalization on a notes dataset.

Answers the open question of the spec — the stopping criterion for the
iterative normalizer. Runs a deterministic rule-based normalize pass up to
N times per note and reports, per iteration:

- tokens/chars, compression vs the original, delta vs previous iteration
- share of notes that reached a fixed point (iteration changed nothing)
- safety hits: result shorter than --min-chars (spec safety: ~100 chars)
- optional embedding drift: cosine(original, iter_k) via the cached model
  (requires --with-embeddings; run inside the nlp container)

The prototype rules are deliberately simple and deterministic — they model
what the spec calls "v1: one deterministic rule pass" (boilerplate from
IMP-8 lessons, near-duplicate lines, nav-like runs, whitespace). The point
of the measurement is not rule quality but iteration behaviour: does a
second pass still change text, where does the token delta fall under the
hypothesized 5%, and does anything degrade below the safety bound.

Usage:
    python measure_normalization.py --dataset notes_dataset.json \
        --out NLP-4-normalization-findings.md --raw raw.json
    # inside kg-test-nlp / kg-nlp (model cached):
    python measure_normalization.py --dataset notes_dataset.json \
        --with-embeddings --out findings.md --raw raw.json
"""
import argparse
import json
import os
import re
import statistics
import unicodedata

# --- normalizer prototype -------------------------------------------------

BOILERPLATE_PATTERNS = [
    # cookie / consent
    r"(?i)\b(cookie|cookies)\b.*\b(accept|agree|consent|policy)\b",
    r"(?i)(мы|сайт)?\s*(используем|использует)\s+(cookies|куки)",
    r"(?i)\bпринима(ю|ть)\b.*\b(условия|cookies|куки|политик)",
    # subscribe / newsletter / signup walls
    r"(?i)\b(subscribe|subscription|newsletter|sign\s*up|log\s*in|sign\s*in)\b",
    r"(?i)\b(подпис(аться|ка|ывайтесь)|рассылк|зарегистрир|войти|войдите)\b",
    # share / social
    r"(?i)\b(share|поделиться|поделитесь|расскажите друзьям)\b.*\b(vk|telegram|facebook|twitter|x\.com|ок|одноклассник)?\b",
    # footer / copyright
    r"(?i)^\s*(©|&copy;|\(c\))\s*\d{4}",
    r"(?i)\b(all rights reserved|все права защищены)\b",
    # navigation / read more
    r"(?i)^\s*(read more|see also|related (posts?|articles?)|читайте также|смотрите также|далее)\b",
    r"(?i)^\s*(home|главная)\s*[>/»]",
]
BOILERPLATE_RE = [re.compile(p) for p in BOILERPLATE_PATTERNS]

BARE_URL_RE = re.compile(r"^\s*(https?://|www\.)\S+\s*$")
WORD_RE = re.compile(r"\w+", re.UNICODE)


def _norm_line(line: str) -> str:
    """Line fingerprint for near-duplicate detection."""
    line = unicodedata.normalize("NFKC", line).lower()
    line = re.sub(r"[^\w\s]", " ", line)
    return re.sub(r"\s+", " ", line).strip()


def normalize_pass(text: str) -> str:
    """One deterministic normalization pass. Should converge to a fixed
    point — if it does not, that is exactly what the measurement reports."""
    out_lines = []
    seen = set()
    nav_run = []

    def flush_nav():
        # >=3 consecutive short unpunctuated lines look like a nav/menu block
        if len(nav_run) >= 3:
            return
        out_lines.extend(nav_run)
        nav_run.clear()

    for raw in text.splitlines():
        line = raw.strip()
        if not line:
            flush_nav()
            if out_lines and out_lines[-1] != "":
                out_lines.append("")
            continue
        if BARE_URL_RE.match(line):
            nav_run.clear()
            continue
        is_short_nav = len(line) < 30 and not re.search(r"[.!?…:;]$", line)
        if is_short_nav:
            nav_run.append(line)
            continue
        flush_nav()
        if any(p.search(line) for p in BOILERPLATE_RE):
            continue
        fp = _norm_line(line)
        if fp and fp in seen:
            continue
        if fp:
            seen.add(fp)
        out_lines.append(line)
    flush_nav()

    text = "\n".join(out_lines)
    text = re.sub(r"\n{3,}", "\n\n", text)
    return text.strip()


def tokens(text: str) -> int:
    return len(WORD_RE.findall(text))


# --- measurement ----------------------------------------------------------

def cosine(a, b) -> float:
    dot = sum(x * y for x, y in zip(a, b))
    na = sum(x * x for x in a) ** 0.5
    nb = sum(x * x for x in b) ** 0.5
    return dot / (na * nb) if na and nb else 0.0


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--dataset", required=True,
                    help="JSON: list of {id,title,content} (or {clen})")
    ap.add_argument("--out", required=True)
    ap.add_argument("--raw", default="")
    ap.add_argument("--iterations", type=int, default=4,
                    help="max passes (spec hypothesis: useful max ~3)")
    ap.add_argument("--min-chars", type=int, default=100,
                    help="safety bound from the spec: shorter result = rollback")
    ap.add_argument("--stop-delta", type=float, default=5.0,
                    help="hypothesized stop: token delta below this %%")
    ap.add_argument("--with-embeddings", action="store_true")
    args = ap.parse_args()

    notes = json.load(open(args.dataset, encoding="utf-8"))
    notes.sort(key=lambda n: n.get("id", ""))  # deterministic order
    texts = {
        n["id"]: (n.get("title", "") + "\n" + n.get("content", "")).strip()
        for n in notes
    }

    model = None
    if args.with_embeddings:
        os.environ.setdefault("HF_HUB_OFFLINE", "1")
        from sentence_transformers import SentenceTransformer
        model = SentenceTransformer(
            "paraphrase-multilingual-MiniLM-L12-v2"
        )

    # per-note iteration trace
    traces = {}
    for nid, text in texts.items():
        iters = [{"text": text, "chars": len(text), "tok": tokens(text)}]
        cur = text
        for _ in range(args.iterations):
            cur = normalize_pass(cur)
            iters.append({"text": cur, "chars": len(cur), "tok": tokens(cur)})
        traces[nid] = iters

    # embedding drift
    drift = {}
    if model is not None:
        for nid, iters in traces.items():
            embs = [model.encode(i["text"] or " ", normalize_embeddings=True)
                    for i in iters]
            drift[nid] = [float(cosine(embs[0], e)) for e in embs]

    # aggregate per iteration
    n_notes = len(traces)
    rows = []
    for k in range(args.iterations + 1):
        toks = [t[k]["tok"] for t in traces.values()]
        chars = [t[k]["chars"] for t in traces.values()]
        comp = [t[k]["tok"] / max(1, t[0]["tok"]) for t in traces.values()]
        unchanged = sum(
            1 for t in traces.values() if t[k]["text"] == t[0]["text"]
        ) if k == 0 else sum(
            1 for t in traces.values() if t[k]["text"] == t[k - 1]["text"]
        )
        safety = sum(1 for t in traces.values() if 0 < t[k]["chars"] < args.min_chars)
        empty = sum(1 for t in traces.values() if t[k]["chars"] == 0)
        row = {
            "iter": k,
            "tok_median": int(statistics.median(toks)),
            "tok_mean": int(statistics.mean(toks)),
            "chars_median": int(statistics.median(chars)),
            "compression_median": round(statistics.median(comp), 3),
            "unchanged_pct": round(100 * unchanged / n_notes, 1),
            "safety_hits": safety,
            "empty": empty,
        }
        if drift:
            ds = [d[k] for d in drift.values()]
            row["emb_cos_median"] = round(statistics.median(ds), 4)
            row["emb_cos_min"] = round(min(ds), 4)
        rows.append(row)

    # where would the hypothesized stop rule land? (delta < stop% or i>=3,
    # rollback if < min_chars)
    stop_iter = []
    degradations = []
    for nid, iters in traces.items():
        stopped_at = args.iterations
        for k in range(1, min(args.iterations, 3) + 1):
            prev, cur = iters[k - 1]["tok"], iters[k]["tok"]
            if iters[k]["chars"] < args.min_chars and iters[0]["chars"] >= args.min_chars:
                degradations.append(nid)
                break
            if prev == 0 or 100 * (prev - cur) / prev < args.stop_delta:
                stopped_at = k
                break
        stop_iter.append(stopped_at)

    md = open(args.out, "a", encoding="utf-8")
    md.write(f"\n## Нормализация: поведение итераций ({n_notes} заметок)\n\n")
    hdr = "| iter | tok med | tok mean | chars med | compr med | unchanged % | safety<{} | empty |".format(args.min_chars)
    if drift:
        hdr += " emb cos med | emb cos min |"
    md.write(hdr + "\n" + "|" + "---|" * (8 + (2 if drift else 0)) + "\n")
    for r in rows:
        line = (f"| {r['iter']} | {r['tok_median']} | {r['tok_mean']} | "
                f"{r['chars_median']} | {r['compression_median']} | "
                f"{r['unchanged_pct']} | {r['safety_hits']} | {r['empty']} |")
        if drift:
            line += f" {r['emb_cos_median']} | {r['emb_cos_min']} |"
        md.write(line + "\n")

    from collections import Counter
    sc = Counter(stop_iter)
    md.write("\n**Стоп-итерация при гипотезе «бюджет 3 + дельта <"
             f"{args.stop_delta}%»:** " +
             ", ".join(f"iter {k}: {v} заметок" for k, v in sorted(sc.items())))
    md.write(f"\n\n**Деградации ниже {args.min_chars} символов (откат бы сработал):** "
             f"{len(degradations)}\n")
    md.flush()

    if args.raw:
        json.dump({
            "rows": rows,
            "stop_iter": dict(sc),
            "degradations": degradations,
            "per_note": {
                nid: [{"chars": i["chars"], "tok": i["tok"]} for i in t]
                for nid, t in traces.items()
            },
            "drift": drift,
        }, open(args.raw, "w", encoding="utf-8"), ensure_ascii=False)

    print(f"done: {n_notes} notes, {args.iterations} iterations -> {args.out}")


if __name__ == "__main__":
    main()
