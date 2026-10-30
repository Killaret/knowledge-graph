#!/usr/bin/env python3
"""MODEL-1: benchmark embedding-model variants on the real notes dataset.

Runs five measurements (search, pair similarity, keywords, speed, memory)
for every variant and appends Markdown tables to the output file.
Same script, same data — numbers stay comparable across variants.

Usage (inside the NLP container):
    python /work/measure_models.py --dataset /work/notes_dataset.json \
        --out /work/MODEL-1-findings.md [--only current@128]
"""
import argparse
import json
import math
import os
import re
import statistics
import time
from collections import Counter

HF_CACHE = os.environ.get("HF_HUB_CACHE") or os.path.join(
    os.environ.get("HF_HOME", "/root/.cache/huggingface"), "hub"
)

VARIANTS = [
    {
        "name": "current@128",
        "repo": "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2",
        "window": 128,
        "qp": "",
        "pp": "",
        "license": "Apache-2.0",
    },
    {
        "name": "current@512",
        "repo": "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2",
        "window": 512,
        "qp": "",
        "pp": "",
        "license": "Apache-2.0",
    },
    {
        "name": "e5-small",
        "repo": "intfloat/multilingual-e5-small",
        "window": 512,
        "qp": "query: ",
        "pp": "passage: ",
        "license": "MIT",
    },
    {
        "name": "rubert-tiny2",
        "repo": "cointegrated/rubert-tiny2",
        "window": 2048,
        "qp": "",
        "pp": "",
        "license": "MIT",
    },
    {
        "name": "e5-base",
        "repo": "intfloat/multilingual-e5-base",
        "window": 512,
        "qp": "query: ",
        "pp": "passage: ",
        "license": "MIT",
    },
]

QUERIES = [
    "книги по архитектуре и микросервисам",
    "английский язык грамматика и времена",
    "вакансии и работа в Европе",
    "манга которую я читаю",
    "keycloak openid authentication",
    "coursera software design courses",
    "криптовалютные биржи и обменники",
    "chrome devtools debugging tips",
    "книга про психологию лжи",
    "english vocabulary dictionaries",
]

STOP_RU = set(
    "и в во не что он на я с со как а то все она так его но да ты к у же вы за бы по "
    "только ее мне было вот от меня еще нет о из ему теперь когда даже ну вдруг ли если "
    "уже или ни быть был него до вас нибудь опять уж вам ведь там потом себя ничего ей "
    "может они тут где есть надо ней для мы тебя их чем была сам чтоб без будто чего раз "
    "тоже себе под будет ж тогда кто этот того потому этого какой совсем ним здесь этом "
    "один почти мой тем чтобы нее сейчас были куда зачем всех никогда можно при наконец "
    "два об другой хоть после над больше тот через эти нас про всего них какая много "
    "разве три эту моя впрочем хорошо свою этой перед иногда лучше чуть том нельзя такой "
    "им более всегда конечно всю между the a an and or of to in for on with at by is are "
    "was were be been it its this that these those from as you your we our he she they them "
    "his her their my your not no do does did have has had will would can could should".split()
)


def rss_mb() -> float:
    with open("/proc/self/status") as f:
        for line in f:
            if line.startswith("VmRSS"):
                return int(line.split()[1]) / 1024.0
    return -1.0


def cosine(a, b):
    dot = sum(x * y for x, y in zip(a, b))
    na = math.sqrt(sum(x * x for x in a))
    nb = math.sqrt(sum(x * x for x in b))
    return dot / (na * nb) if na and nb else 0.0


def percentile(vals, p):
    vals = sorted(vals)
    k = max(0, min(len(vals) - 1, math.ceil(p / 100.0 * len(vals)) - 1))
    return vals[k]


def ngram_candidates(text, top=20):
    words = [w for w in re.findall(r"[A-Za-zА-Яа-яЁё0-9-]{3,}", text.lower())
             if w not in STOP_RU]
    c = Counter()
    for n in (1, 2, 3):
        for i in range(len(words) - n + 1):
            c[" ".join(words[i:i + n])] += 1
    return [w for w, cnt in c.most_common(top) if cnt >= 2 or top <= 10][:top]


def write_report(path, results, names, notes):
    """Regenerate the whole findings report for every variant in `results`.

    `names` fixes section/column order (VARIANTS order for measured names).
    The file is rewritten, not appended: MODEL-1 lost table 2 for all
    variants except the first because each `--only` run appended its own
    one-variant report block and the assembly kept the first table-2
    section only. Accumulating results in a state file and rewriting the
    report makes every run produce complete tables for all variants
    measured so far."""
    out = open(path, "w", encoding="utf-8")

    out.write("\n## Таблица 0. Параметры моделей (сверено по файлам)\n\n")
    out.write("| Вариант | hidden | max_position | окно в файле | окно замера | лицензия |\n|---|---|---|---|---|---|\n")
    lic = {v["name"]: v["license"] for v in VARIANTS}
    for n in names:
        p = results[n]["params"]
        out.write(f"| {n} | {p['hidden']} | {p['max_pos']} | {p['file_window']} | {p['window_used']} | {lic.get(n, '?')} |\n")

    out.write("\n## Таблица 1. Поиск — top-10 по каждому запросу\n")
    for qi, q in enumerate(QUERIES):
        out.write(f"\n### Запрос {qi + 1}: «{q}»\n\n| # | " + " | ".join(names) + " |\n|---|" + "---|" * len(names) + "\n")
        for r in range(10):
            cells = " | ".join(results[n]["search"][qi][r][:60] for n in names)
            out.write(f"| {r + 1} | {cells} |\n")

    out.write("\n## Таблица 2. Близость пар — top-15 по 20 типичным заметкам\n\n")
    for n in names:
        out.write(f"\n### {n}\n\n| # | Пара | cos |\n|---|---|---|\n")
        for i, (s, a, b) in enumerate(results[n]["pairs"]):
            out.write(f"| {i + 1} | {a[:45]} ↔ {b[:45]} | {s:.3f} |\n")

    out.write("\n## Таблица 3. Ключевые слова — top-10 (гибрид: частотный шорт-лист → модель; колонка yake — базовая линия)\n")
    yake_kw = None
    try:
        import yake
        nltk_stop = set()
        try:
            import nltk
            nltk_stop = set(nltk.corpus.stopwords.words("russian") + nltk.corpus.stopwords.words("english"))
        except Exception:
            pass
        yake_kw = yake.KeywordExtractor(lan="ru", top=10, stopwords=nltk_stop or None)
    except ImportError:
        pass
    kw_notes = list(results[names[0]]["kw"].keys())
    note_by_title = {n["title"]: n for n in notes}
    for t in kw_notes:
        if yake_kw is not None:
            yake_top = [w for w, _ in yake_kw.extract_keywords(note_by_title[t]["content"])[:10]]
        else:
            yake_top = []
        out.write(f"\n### {t[:70]}\n\n| # | " + " | ".join(names) + " | yake |\n|---|" + "---|" * (len(names) + 1) + "\n")
        for r in range(10):
            cells = " | ".join(
                (results[n]["kw"][t][r] if r < len(results[n]["kw"][t]) else "—") for n in names)
            cells += " | " + (yake_top[r] if r < len(yake_top) else "—")
            out.write(f"| {r + 1} | {cells} |\n")

    out.write("\n## Таблица 4. Скорость (CPU, внутри контейнера)\n\n")
    out.write("| Вариант | embed p50, мс | embed p95, мс | весь корпус, с | keybert p50, мс | keybert p95, мс |\n|---|---|---|---|---|---|\n")
    for n in names:
        r = results[n]
        out.write(f"| {n} | {r['p50_ms']:.0f} | {r['p95_ms']:.0f} | {r['encode_all_s']:.1f} | {r['kw_p50_ms']:.0f} | {r['kw_p95_ms']:.0f} |\n")

    out.write("\n## Таблица 5. Память контейнера с загруженной моделью\n\n")
    out.write("| Вариант | RSS процесса, МБ | прирост над baseline, МБ |\n|---|---|---|\n")
    for n in names:
        r = results[n]
        out.write(f"| {n} | {r['proc_mb']:.0f} | {r['model_mb']:.0f} |\n")

    out.close()


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--dataset", required=True)
    ap.add_argument("--out", required=True)
    ap.add_argument("--only", default="")
    ap.add_argument("--state", default="",
                    help="JSON-файл накопленных результатов; по умолчанию "
                         "<out>.state.json. Прогоны --only дополняют его, "
                         "отчёт перегенерируется по всем накопленным вариантам.")
    args = ap.parse_args()

    from huggingface_hub import snapshot_download
    from sentence_transformers import SentenceTransformer

    notes = json.load(open(args.dataset, encoding="utf-8"))
    notes.sort(key=lambda n: n["id"])  # deterministic order
    texts = [(n["title"] + "\n" + n["content"])[:8000] for n in notes]

    # subsets
    by_len = sorted(notes, key=lambda n: n["clen"])
    longest = by_len[-20:]
    mid = len(by_len) // 2
    typical = by_len[mid - 10:mid + 10]
    sample_kw = (by_len[-5:] + by_len[mid - 3:mid + 2])[:8]  # 8 notes for keywords table

    results = {}

    for v in VARIANTS:
        if args.only and v["name"] != args.only:
            continue
        print(f"\n=== {v['name']} ===", flush=True)
        path = snapshot_download(repo_id=v["repo"], cache_dir=HF_CACHE)

        # verify params by files
        cfg = json.load(open(os.path.join(path, "config.json")))
        sb_cfg = {}
        sbc_path = os.path.join(path, "sentence_bert_config.json")
        if os.path.exists(sbc_path):
            sb_cfg = json.load(open(sbc_path))
        params = {
            "hidden": cfg.get("hidden_size"),
            "max_pos": cfg.get("max_position_embeddings"),
            "file_window": sb_cfg.get("max_seq_length"),
            "window_used": v["window"],
        }

        base_rss = rss_mb()
        model = SentenceTransformer(path)
        model.max_seq_length = v["window"]
        model_rss = rss_mb()

        # encode corpus
        docs_in = [v["pp"] + t for t in texts]
        t0 = time.time()
        doc_emb = model.encode(docs_in, batch_size=16, convert_to_numpy=False)
        encode_all_s = time.time() - t0
        doc_emb = [e.tolist() if hasattr(e, "tolist") else list(e) for e in doc_emb]

        # per-note latency
        lat = []
        for t in texts:
            t0 = time.time()
            model.encode([v["pp"] + t], convert_to_numpy=False)
            lat.append(time.time() - t0)

        # search
        q_emb = model.encode([v["qp"] + q for q in QUERIES], convert_to_numpy=False)
        q_emb = [e.tolist() if hasattr(e, "tolist") else list(e) for e in q_emb]
        search = []
        for qe in q_emb:
            scored = sorted(range(len(notes)), key=lambda i: -cosine(qe, doc_emb[i]))[:10]
            search.append([notes[i]["title"] for i in scored])

        # pair similarity on 20 typical
        t_idx = [notes.index(n) for n in typical]
        pairs = []
        for a in range(len(t_idx)):
            for b in range(a + 1, len(t_idx)):
                s = cosine(doc_emb[t_idx[a]], doc_emb[t_idx[b]])
                pairs.append((s, notes[t_idx[a]]["title"], notes[t_idx[b]]["title"]))
        pairs.sort(key=lambda x: -x[0])
        top_pairs = pairs[:15]

        # keywords: hybrid freq-shortlist -> model rank, 8 sample notes
        kw_lat = []
        kw_out = {}
        for n in sample_kw:
            i = notes.index(n)
            cands = ngram_candidates(n["title"] + " " + n["content"], 20)
            if not cands:
                kw_out[n["title"]] = []
                continue
            t0 = time.time()
            c_emb = model.encode([v["pp"] + c for c in cands], convert_to_numpy=False)
            kw_lat.append(time.time() - t0)
            c_emb = [e.tolist() if hasattr(e, "tolist") else list(e) for e in c_emb]
            scored = sorted(zip(cands, (cosine(doc_emb[i], e) for e in c_emb)),
                            key=lambda x: -x[1])[:10]
            kw_out[n["title"]] = [w for w, _ in scored]

        results[v["name"]] = {
            "params": params, "model_mb": model_rss - base_rss, "proc_mb": model_rss,
            "encode_all_s": encode_all_s,
            "p50_ms": percentile(lat, 50) * 1000, "p95_ms": percentile(lat, 95) * 1000,
            "kw_p50_ms": percentile(kw_lat, 50) * 1000 if kw_lat else 0,
            "kw_p95_ms": percentile(kw_lat, 95) * 1000 if kw_lat else 0,
            "search": search, "pairs": top_pairs, "kw": kw_out,
        }
        del model, doc_emb
        print(f"    rss +{model_rss - base_rss:.0f}MB  p50 {results[v['name']]['p50_ms']:.0f}ms", flush=True)

    # ---- merge into state, regenerate the whole report ----
    state_path = args.state or (args.out + ".state.json")
    state = {}
    if os.path.exists(state_path):
        with open(state_path, encoding="utf-8") as f:
            state = json.load(f)
    state.update(results)
    with open(state_path, "w", encoding="utf-8") as f:
        json.dump(state, f, ensure_ascii=False)

    names = [v["name"] for v in VARIANTS if v["name"] in state]
    write_report(args.out, state, names, notes)
    print("\nwritten:", args.out)


if __name__ == "__main__":
    main()
