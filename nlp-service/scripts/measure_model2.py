#!/usr/bin/env python3
"""MODEL-2: final model measurement on the *production* pipeline.

Decision 60 (owner, 2026-09-24): three candidates — D (current model with
NLP-4 normalization + CHUNK-1 chunking), e5-small, e5-base — rerun of p@k
and separation AUC on the ready pipeline, not the MODEL-1B prototypes.
"Do not change the model" is a legitimate outcome.

Differences from measure_model1b.py:

* embeddings are built by the real modules — ``app.core.normalization.
  normalize`` (one pass + length/cosine rollback guards) and
  ``app.core.chunking`` (structure-aware chunks, title injection,
  mean-pool + L2) — the same path ``/normalize`` + ``/embed`` serve.
* e5 candidates get ``passage:`` doc / ``query:`` search prefixes, which
  MODEL-2 spec item 2 prescribes for the migration.
* threshold rule is *matching precision*: prod today scores precision
  0.71 at threshold 0.60 (MODEL-1B, variant A); each candidate's
  recommended ``GAMMA_LINK_MIN_SCORE`` is the threshold giving maximal
  coverage at precision >= PROD_PRECISION. t* from MODEL-1B was
  degenerate and is not repeated.
* A (today's prod path) is kept as a reference row only — it is not a
  candidate.

Usage (single reproduction command, also printed in the findings):
    HF_HOME=D:/kg-hf-cache python nlp-service/scripts/measure_model2.py \
        --corpus work-nlp4/notes_dataset.json --folders work-w1/dataset.json \
        --findings docs/tasks/MODEL-2-final-findings.md \
        --local-dir work-model2/local
"""
import argparse
import json
import os
import statistics
import sys
import time
from pathlib import Path

SCRIPT_DIR = Path(__file__).resolve().parent
NLP_ROOT = SCRIPT_DIR.parent
sys.path.insert(0, str(SCRIPT_DIR))
sys.path.insert(0, str(NLP_ROOT))

import measure_model1b as m                                    # noqa: E402
from measure_model1b_separation import auc_rank               # noqa: E402
from measure_models import QUERIES                            # noqa: E402
from eval_link_formula import GRANULARITIES                   # noqa: E402
from app.core.normalization import NormalizationParams, normalize   # noqa: E402
from app.core.chunking import ChunkingParams, chunk, embedding_inputs, aggregate  # noqa: E402
from app.nlp_utils import _chunk_max_tokens, _chunk_token_counter   # noqa: E402

MODELS = {
    "current": "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2",
    "e5-small": "intfloat/multilingual-e5-small",
    "e5-base": "intfloat/multilingual-e5-base",
}

# name -> (model, prod_pipeline, doc_prefix, query_prefix). A is the
# reference production path of today: raw text, truncation at window 128.
VARIANTS = [
    {"name": "A", "model": "current", "prod": False, "pp": "", "qp": "",
     "candidate": False},
    {"name": "D", "model": "current", "prod": True, "pp": "", "qp": "",
     "candidate": True},
    {"name": "E", "model": "e5-small", "prod": True, "pp": "passage: ",
     "qp": "query: ", "candidate": True},
    {"name": "F", "model": "e5-base", "prod": True, "pp": "passage: ",
     "qp": "query: ", "candidate": True},
]

PROD_PRECISION = 0.71   # precision of today's autolink at threshold 0.60
BOOT_SEED = 20260927
BOOT_N = 1000
THRESHOLDS = m.THRESHOLDS
DEGREE = m.DEGREE


def build_doc_vectors(variant, model, notes):
    """note_id -> L2-normalized doc vector via the PRODUCTION pipeline.

    normalize() runs the real NLP-4 module with its injected embed_fn and
    token counter (same wiring as /normalize); the structural chunker is
    app.core.chunking exactly as compute_chunked_embedding drives it.
    e5 variants prepend the spec'd passage: prefix to every chunk input —
    production will do the same inside the e5 migration (MODEL-2 item 2).
    """
    import numpy as np

    vecs, stats = {}, {"chunks_per_note": [], "zero_chunk": 0,
                     "norm_rollback": 0}
    pp = variant["pp"]

    if not variant["prod"]:
        inputs = [(n["title"] + "\n" + n["content"])[:8000] for n in notes]
        emb = model.encode(inputs, batch_size=16, convert_to_numpy=True,
                           normalize_embeddings=True)
        for n, e in zip(notes, emb):
            vecs[n["id"]] = np.asarray(e, dtype=np.float32)
        return vecs, stats

    token_counter = _chunk_token_counter(model)
    max_tokens = _chunk_max_tokens(model)

    def embed_fn(texts):
        return model.encode(list(texts), convert_to_numpy=True)

    flat, owner = [], []
    for n in notes:
        result = normalize(
            n["content"],
            NormalizationParams(min_cosine=0.7, embed_fn=embed_fn,
                                token_counter=token_counter),
        )
        if result.rolled_back:
            stats["norm_rollback"] += 1
        params = ChunkingParams(
            token_counter=token_counter,
            target_tokens=256,
            max_tokens=max_tokens,
            title=n["title"] or None,
            title_injection=True,
        )
        chunks = chunk(result.normalized_text, params)
        inputs = embedding_inputs(chunks, params)
        if not inputs:
            inputs = [n["title"] or ""]
            stats["zero_chunk"] += 1
        stats["chunks_per_note"].append(len(chunks))
        for text_in in inputs:
            flat.append(pp + text_in)
            owner.append(n["id"])

    emb = model.encode(flat, batch_size=16, convert_to_numpy=True,
                       normalize_embeddings=True)
    buckets = {}
    for nid, e in zip(owner, emb):
        buckets.setdefault(nid, []).append(e)
    for nid, es in buckets.items():
        vecs[nid] = np.asarray(
            m.l2(np.asarray(es).mean(axis=0).tolist()), dtype=np.float32)
    return vecs, stats


def matching_threshold(rows):
    """MODEL-2 rule: max coverage at precision >= prod's 0.71."""
    eligible = [r for r in rows if r["precision"] >= PROD_PRECISION
                and r["links"]]
    if not eligible:
        return None
    best = max(eligible, key=lambda r: (r["linked_share"], -r["t"]))
    return best


def separation(vecs, eval_notes):
    """MODEL-1B separation block: AUC + per-note margin."""
    import numpy as np
    M = np.stack([np.asarray(vecs[n["id"]], dtype=np.float64)
                  for n in eval_notes])
    M /= np.linalg.norm(M, axis=1, keepdims=True)
    sims = M @ M.T
    iu = np.triu_indices(len(eval_notes), k=1)
    same = np.array([eval_notes[a]["folder"] == eval_notes[b]["folder"]
                     for a, b in zip(*iu)])
    intra, inter = sims[iu][same], sims[iu][~same]
    auc = auc_rank(intra, inter)
    margins = []
    for k, n in enumerate(eval_notes):
        others = np.delete(np.arange(len(eval_notes)), k)
        fid = n["folder"]
        same_m = np.array([eval_notes[o]["folder"] == fid
                           for o in others])
        if not same_m.any() or same_m.all():
            continue
        s = sims[k, others]
        margins.append(s[same_m].max() - s[~same_m].max())
    return {
        "auc": float(auc),
        "intra_n": len(intra), "inter_n": len(inter),
        "intra_med": float(np.median(intra)),
        "inter_med": float(np.median(inter)),
        "margin_med": float(np.median(margins)) if margins else None,
        "sims": sims, "same": same, "iu": iu, "intra": intra,
    }


def phantom_bridges(sep, eval_notes, limit=40):
    """Cross-folder pairs above the intra median, for the local listing."""
    import numpy as np
    thr = sep["intra_med"]
    pairs = sep["sims"][sep["iu"]]
    cand = np.where((~sep["same"]) & (pairs > thr))[0]
    cand = cand[np.argsort(-pairs[cand])]
    out = []
    for c in cand[:limit]:
        a, b = sep["iu"][0][c], sep["iu"][1][c]
        na, nb = eval_notes[a], eval_notes[b]
        out.append((float(pairs[c]),
                    na["title"], na["folder"], nb["title"], nb["folder"]))
    return thr, out


def build_findings(results, order, notes, eval_ids, boot, timing,
                   thread_info, corpus_note, local_dir):
    """Numbers-only findings text. Tested by test_measure_model2_report."""
    f = ["# MODEL-2 final findings — замер на готовом конвейере\n\n",
         f"Корпус: {len(notes)} заметок, {len(eval_ids)} с непустой папкой. ",
         "Конвейер промышленный: `app.core.normalization.normalize` (один ",
         "проход + откаты по длине и косинусу) и `app.core.chunking` ",
         "(структурные чанки, заголовок в каждом чанке, среднее + L2) — те же ",
         "модули, что у `/normalize` и `/embed`. e5-варианты с префиксами ",
         "`passage:`/`query:` (пункт 2 спеки). A — справочная строка прода ",
         "сегодня, не кандидат.\n",
         f"{corpus_note}\n" if corpus_note else "",
         f"Потоки: torch={thread_info['torch_threads']}, ",
         f"OMP_NUM_THREADS={thread_info['OMP_NUM_THREADS'] or 'не задан'}, ",
         f"cpu_count={thread_info['cpu_count']}.\n\n",
         f"Артефакты с названиями (не в git): `{local_dir}` — `pairs.md`, ",
         "`search.md`, `delta-sheet.md`, `bridges.md`.\n"]

    f.append("\n## Точность по папкам (p@k)\n")
    f.append("| Вариант | p@1 ex | p@3 ex | p@5 ex | p@1 l1 | p@3 l1 | "
             "p@5 l1 | p@1 par | p@3 par | p@5 par |\n"
             "|---|---|---|---|---|---|---|---|---|---|\n")
    for name in order:
        p = results[name]["precision"]
        row = [f"{p[g][k]:.3f}" for g in ("exact", "l1", "parent")
               for k in (1, 3, 5)]
        f.append(f"| {name} | " + " | ".join(row) + " |\n")

    f.append("\n## Сепарация папок\n")
    f.append("| Вариант | пар своих | пар чужих | своих med | чужих med | "
             "AUC | margin med |\n|---|---|---|---|---|---|---|\n")
    for name in order:
        s = results[name]["sep"]
        mg = f"{s['margin_med']:.3f}" if s["margin_med"] is not None else "—"
        f.append(f"| {name} | {s['intra_n']} | {s['inter_n']} | "
                 f"{s['intra_med']:.3f} | {s['inter_med']:.3f} | "
                 f"{s['auc']:.4f} | {mg} |\n")

    f.append("\n## Кривая порога автолинка (степень 2)\n\n")
    f.append("Ячейка: доля связанных заметок / точность связей "
             "(цель в той же папке, exact).\n\n")
    f.append("| Порог | " + " | ".join(order) + " |\n|---|" +
             "---|" * len(order) + "|\n")
    for ti, t in enumerate(THRESHOLDS):
        cells = []
        for name in order:
            r = results[name]["curve"][ti]
            cells.append(f"{r['linked_share']:.2f}/{r['precision']:.2f}")
        f.append(f"| {t:.2f} | " + " | ".join(cells) + " |\n")

    f.append("\n## Порог по совпадающей точности (правило MODEL-2)\n\n")
    f.append(f"Эталон прода: precision {PROD_PRECISION} при пороге 0.60. "
             "Для каждого кандидата — наибольшее покрытие при точности не "
             "ниже эталона; это кандидатное `GAMMA_LINK_MIN_SCORE`.\n\n")
    f.append("| Вариант | порог | покрытие | точность |\n|---|---|---|---|\n")
    for name in order:
        mt = results[name]["match_t"]
        if mt is None:
            f.append(f"| {name} | нет (ни один порог не даёт ≥ "
                     f"{PROD_PRECISION}) | — | — |\n")
        else:
            f.append(f"| {name} | {mt['t']:.2f} | {mt['linked_share']:.2f} | "
                     f"{mt['precision']:.2f} |\n")

    f.append("\n## Распределение близости соседей\n")
    f.append("| Вариант | 1-й сосед med | 1-й mean | 2-й med | 2-й mean "
             "|\n|---|---|---|---|---|\n")
    for name in order:
        d = results[name]["dist"]
        f.append(f"| {name} | {d['first_median']:.3f} | "
                 f"{d['first_mean']:.3f} | {d['second_median']:.3f} | "
                 f"{d['second_mean']:.3f} |\n")

    if boot:
        f.append("\n## Шум (бутстрэп 1000 повторов, 95% CI разности p@3 "
                 "exact)\n")
        for label, (lo, hi) in boot.items():
            f.append(f"- {label}: [{lo:.4f}, {hi:.4f}]\n")

    f.append("\n## Скорость и память\n")
    f.append("| Вариант | заметка p50, мс | заметка p95, мс | чанк p50, мс | "
             "чанков | RSS модели, МБ | дельта RSS, МБ |\n"
             "|---|---|---|---|---|---|---|\n")
    for name in order:
        t = timing.get(name)
        if not t:
            continue
        ch = f"{t['chunk_p50_ms']:.0f}" if t["chunk_p50_ms"] is not None else "—"
        f.append(f"| {name} | {t['note_p50_ms']:.0f} | {t['note_p95_ms']:.0f} | "
                 f"{ch} | {t['chunks_total']} | {t['model_rss_mb']:.0f} | "
                 f"{t['model_delta_mb']:.0f} |\n")

    f.append("\n## Итог правила выбора\n\n")
    f.append("Правило (решение 60): e5-base — только если 95% CI разности "
             "p@3 exact F − E целиком выше нуля; иначе e5-small. Строка "
             "заполняется по таблице выше исполнителем после прогона.\n")
    return "".join(f)


def timed_pipeline(variant, model, notes):
    """Per-note wall time through the production pipeline; per-chunk encode."""
    import numpy as np  # noqa: F401
    note_ms, chunk_ms = [], []
    pp = variant["pp"]
    if not variant["prod"]:
        for n in notes:
            t0 = time.perf_counter()
            model.encode([(n["title"] + "\n" + n["content"])[:8000]],
                         convert_to_numpy=True)
            note_ms.append((time.perf_counter() - t0) * 1000)
        return note_ms, chunk_ms

    token_counter = _chunk_token_counter(model)
    max_tokens = _chunk_max_tokens(model)

    def embed_fn(texts):
        return model.encode(list(texts), convert_to_numpy=True)

    for n in notes:
        t0 = time.perf_counter()
        result = normalize(
            n["content"],
            NormalizationParams(min_cosine=0.7, embed_fn=embed_fn,
                                token_counter=token_counter))
        params = ChunkingParams(token_counter=token_counter,
                                target_tokens=256, max_tokens=max_tokens,
                                title=n["title"] or None,
                                title_injection=True)
        chunks = chunk(result.normalized_text, params)
        inputs = embedding_inputs(chunks, params) or [n["title"] or ""]
        for text_in in inputs:
            c0 = time.perf_counter()
            model.encode([pp + text_in], convert_to_numpy=True)
            chunk_ms.append((time.perf_counter() - c0) * 1000)
        note_ms.append((time.perf_counter() - t0) * 1000)
    return note_ms, chunk_ms


def main() -> int:
    # Windows console is cp1251 — the report labels carry '−' (U+2212).
    if hasattr(sys.stdout, "reconfigure"):
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
    ap = argparse.ArgumentParser()
    ap.add_argument("--corpus", required=True)
    ap.add_argument("--folders", required=True)
    ap.add_argument("--findings", required=True)
    ap.add_argument("--local-dir", required=True)
    ap.add_argument("--only", default="", help="single variant name")
    ap.add_argument("--no-cache", action="store_true",
                    help="recompute vectors even if vecs-*.npz exists")
    ap.add_argument("--corpus-note", default="")
    args = ap.parse_args()

    local_dir = Path(args.local_dir)
    local_dir.mkdir(parents=True, exist_ok=True)

    corpus = json.load(open(args.corpus, encoding="utf-8"))
    corpus.sort(key=lambda n: n["id"])
    w1 = json.load(open(args.folders, encoding="utf-8"))
    folder_of = {n["id"]: (n.get("folder") or "").strip()
                 for n in w1["notes"]}
    notes = [dict(n, folder=folder_of.get(n["id"], "")) for n in corpus]
    eval_ids = [n["id"] for n in notes if n["folder"]]
    all_ids = [n["id"] for n in notes]
    eval_notes = [n for n in notes if n["id"] in set(eval_ids)]
    print(f"corpus {len(notes)} notes, {len(eval_ids)} with folder", flush=True)

    os.environ.setdefault("TOKENIZERS_PARALLELISM", "false")
    from sentence_transformers import SentenceTransformer
    import torch
    thread_info = {"torch_threads": torch.get_num_threads(),
                   "OMP_NUM_THREADS": os.environ.get("OMP_NUM_THREADS", ""),
                   "cpu_count": os.cpu_count()}

    loaded, rss, results = {}, {}, {}
    for v in VARIANTS:
        if args.only and v["name"] != args.only:
            continue
        print(f"=== {v['name']} ({v['model']}, prod={v['prod']}) ===",
              flush=True)
        if v["model"] not in loaded:
            base = m.rss_mb()
            loaded[v["model"]] = SentenceTransformer(MODELS[v["model"]])
            rss[v["model"]] = {"rss_mb": m.rss_mb(),
                               "delta_mb": m.rss_mb() - base}
        model = loaded[v["model"]]
        if not v["prod"]:
            model.max_seq_length = 128
        cache = local_dir / f"vecs-{v['name']}.npz"
        t0 = time.time()
        if cache.exists() and not args.no_cache:
            import numpy as np
            z = np.load(cache, allow_pickle=False)
            vecs = {nid: z["M"][i] for i, nid in
                    enumerate(z["ids"].tolist())}
            stats = {"encode_all_s": 0.0, "from_cache": True}
            print(f"    vectors from {cache.name}", flush=True)
        else:
            vecs, stats = build_doc_vectors(v, model, notes)
            stats["encode_all_s"] = time.time() - t0
            import numpy as np
            np.savez(cache, ids=np.array(all_ids),
                     M=np.stack([vecs[i] for i in all_ids]))
        prec, per_note_p = m.precision_maps(vecs, notes, eval_ids)
        top2 = m.top_neighbors(vecs, all_ids, DEGREE)
        curve, _tstar = m.threshold_curve(
            {e: top2[e] for e in eval_ids}, folder_of, eval_ids)
        sep = separation(vecs, eval_notes)
        results[v["name"]] = {
            "variant": v, "vecs": vecs, "stats": stats,
            "precision": prec, "per_note_p": per_note_p,
            "curve": curve, "match_t": matching_threshold(curve),
            "dist": m.neighbor_distribution(top2, eval_ids),
            "top2": top2, "sep": sep,
        }
        print(f"    p@3 exact={prec['exact'][3]:.3f} "
              f"auc={sep['auc']:.3f} "
              f"match_t={results[v['name']]['match_t']}", flush=True)

    boot = {}
    for label, a, b in (("F − E", "E", "F"), ("E − D", "D", "E"),
                        ("F − D", "D", "F")):
        if a in results and b in results:
            boot[label] = m.bootstrap_diff(
                results[a]["per_note_p"]["exact"],
                results[b]["per_note_p"]["exact"])
            print(f"bootstrap {label} p@3 exact 95% CI: "
                  f"[{boot[label][0]:.4f}, {boot[label][1]:.4f}]", flush=True)

    timing = {}
    for name in ("A", "D", "E", "F"):
        if name not in results:
            continue
        v = results[name]["variant"]
        model = loaded[v["model"]]
        print(f"timing {name}...", flush=True)
        note_ms, chunk_ms = timed_pipeline(v, model, notes)
        timing[name] = {
            "note_p50_ms": m.percentile(note_ms, 50),
            "note_p95_ms": m.percentile(note_ms, 95),
            "chunk_p50_ms": m.percentile(chunk_ms, 50) if chunk_ms else None,
            "chunks_total": len(chunk_ms),
            "model_rss_mb": rss[v["model"]]["rss_mb"],
            "model_delta_mb": rss[v["model"]]["delta_mb"],
        }

    # ---- local artifacts (titles; dir must stay gitignored) ----
    order = [v["name"] for v in VARIANTS if v["name"] in results]
    by_len = sorted(notes, key=lambda n: len(n["content"]))
    mid = len(by_len) // 2
    t_ids = [n["id"] for n in by_len[mid - 10:mid + 10]]
    title_of = {n["id"]: n["title"] for n in notes}

    pairs_md = ["# MODEL-2: близость пар — top-15 на 20 типичных заметках\n"]
    for name in order:
        pairs = []
        for a in range(len(t_ids)):
            for b in range(a + 1, len(t_ids)):
                s = m.cosine(results[name]["vecs"][t_ids[a]],
                             results[name]["vecs"][t_ids[b]])
                pairs.append((s, title_of[t_ids[a]], title_of[t_ids[b]]))
        pairs.sort(key=lambda x: -x[0])
        pairs_md.append(f"\n## {name}\n\n| # | Пара | cos |\n|---|---|---|\n")
        for i, (s, a, b) in enumerate(pairs[:15]):
            pairs_md.append(f"| {i + 1} | {m.cell(a, 45)} ↔ {m.cell(b, 45)} "
                            f"| {s:.3f} |")
    (local_dir / "pairs.md").write_text("\n".join(pairs_md), encoding="utf-8")

    search_md = ["# MODEL-2: выдача поиска top-10\n"]
    for name in order:
        v = results[name]["variant"]
        model = loaded[v["model"]]
        q_emb = model.encode([v["qp"] + q for q in QUERIES],
                             convert_to_numpy=True,
                             normalize_embeddings=True)
        search_md.append(f"\n## {name}\n")
        for qi, q in enumerate(QUERIES):
            scored = sorted(all_ids, key=lambda i: -m.cosine(
                q_emb[qi].tolist(), results[name]["vecs"][i].tolist()))[:10]
            search_md.append(f"\n### {qi + 1}. {q}\n")
            for r, nid in enumerate(scored):
                search_md.append(f"{r + 1}. {m.cell(title_of[nid], 70)}")
    (local_dir / "search.md").write_text("\n".join(search_md), encoding="utf-8")

    bridges_md = ["# MODEL-2: мосты-фантомы (чужие пары выше медианы своих)\n"]
    for name in order:
        thr, bl = phantom_bridges(results[name]["sep"], eval_notes)
        bridges_md.append(f"\n## {name} (порог = медиана своих {thr:.3f})\n\n"
                          "| # | Близость | Заметка A | Папка A | Заметка B |"
                          " Папка B |\n|---|---|---|---|---|---|\n")
        for i, (s, ta, fa, tb, fb) in enumerate(bl, 1):
            bridges_md.append(f"| {i} | {s:.3f} | {m.cell(ta, 60)} | "
                              f"{m.cell(fa, 25)} | {m.cell(tb, 60)} | "
                              f"{m.cell(fb, 25)} |\n")
    (local_dir / "bridges.md").write_text("".join(bridges_md), encoding="utf-8")

    candidates = [n for n in order if results[n]["variant"]["candidate"]]
    if len(candidates) >= 2 and "A" in results:
        leader = max(candidates,
                     key=lambda c: results[c]["precision"]["exact"][3])
        top1_a = {i: results["A"]["top2"][i][0][0] for i in all_ids}
        top1_l = {i: results[leader]["top2"][i][0][0] for i in all_ids}
        diff = [i for i in all_ids if top1_a[i] != top1_l[i]][:25]
        lines = [f"# MODEL-2: лист дельта-разметки (лидер {leader} vs A)\n",
                 "Отметка: «было лучше / стало лучше / оба мимо»\n",
                 "| # | Заметка | Сосед A | Сосед " + leader + " | Оценка |",
                 "|---|---|---|---|---|"]
        for i, nid in enumerate(diff):
            lines.append(f"| {i + 1} | {m.cell(title_of[nid], 50)} | "
                         f"{m.cell(title_of[top1_a[nid]], 45)} | "
                         f"{m.cell(title_of[top1_l[nid]], 45)} | |")
        (local_dir / "delta-sheet.md").write_text("\n".join(lines) + "\n",
                                                 encoding="utf-8")

    findings = build_findings(results, order, notes, eval_ids, boot, timing,
                              thread_info, args.corpus_note,
                              local_dir.resolve())
    Path(args.findings).write_text(findings, encoding="utf-8")
    print("written:", args.findings, flush=True)
    return 0


if __name__ == "__main__":
    sys.exit(main())
