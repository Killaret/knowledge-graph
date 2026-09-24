#!/usr/bin/env python3
"""MODEL-1B supplement: does the larger model blur folder boundaries?

Owner's hypothesis: e5-base's wider window/capacity may find similarity
where none meaningfully exists — cross-folder "phantom bridges" whose
weight dilutes real clusters. Measured two ways on the same corpus:

1. Separation AUC — P(a random same-folder pair scores higher than a
   random cross-folder pair). Lower AUC = more boundary blur.
   Also the per-note margin: best same-folder sim minus best
   cross-folder sim.
2. Phantom bridges — cross-folder pairs whose similarity sits above the
   median of the same-folder distribution. Listed with titles into
   --local-dir/bridges.md (private titles stay out of git).

Numbers-only report -> --findings; title-bearing list -> --local-dir.

Usage:
    HF_HOME=D:/kg-hf-cache python nlp-service/scripts/measure_model1b_separation.py \
        --corpus work-nlp4/notes_dataset.json --folders work-w1/dataset.json \
        --findings docs/tasks/MODEL-1B-separation-findings.md \
        --local-dir work-model1b/local
"""
import argparse
import json
import os
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import numpy as np                                        # noqa: E402
import measure_model1b as m                               # noqa: E402


def auc_rank(intra, inter):
    """Mann-Whitney AUC: P(intra > inter) + 0.5*P(tie), average ranks."""
    allv = np.concatenate([intra, inter])
    order = allv.argsort(kind="mergesort")
    ranks = np.empty(len(allv))
    ranks[order] = np.arange(1, len(allv) + 1)
    # average ranks for ties
    uniq, inv, counts = np.unique(allv, return_inverse=True,
                                  return_counts=True)
    sums = np.zeros(len(uniq))
    np.add.at(sums, inv, ranks)
    ranks = sums[inv] / counts[inv]
    ri = ranks[: len(intra)].sum()
    n_i, n_j = len(intra), len(inter)
    return (ri - n_i * (n_i + 1) / 2) / (n_i * n_j)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--corpus", required=True)
    ap.add_argument("--folders", required=True)
    ap.add_argument("--findings", required=True)
    ap.add_argument("--local-dir", required=True)
    ap.add_argument("--only", default="",
                    help="comma-separated variant names (default: all 8)")
    ap.add_argument("--bridges", type=int, default=40,
                    help="top phantom bridges to list per variant")
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
    eval_set = set(eval_ids)
    print(f"corpus {len(notes)} notes, {len(eval_ids)} with folder",
          flush=True)

    os.environ.setdefault("TOKENIZERS_PARALLELISM", "false")
    from sentence_transformers import SentenceTransformer

    want = set(args.only.split(",")) if args.only else None
    loaded, rows, bridges = {}, [], {}
    for v in m.VARIANTS:
        if want and v["name"] not in want:
            continue
        print(f"=== {v['name']} ===", flush=True)
        if v["model"] not in loaded:
            loaded[v["model"]] = SentenceTransformer(m.MODELS[v["model"]])
        vecs, _stats = m.build_doc_vectors(v, loaded[v["model"]], notes)

        eval_notes = [n for n in notes if n["id"] in eval_set
                      and n["id"] in vecs]
        M = np.stack([np.asarray(vecs[n["id"]], dtype=np.float64)
                      for n in eval_notes])
        M /= np.linalg.norm(M, axis=1, keepdims=True)
        sims = M @ M.T
        iu = np.triu_indices(len(eval_notes), k=1)
        same = np.array([eval_notes[a]["folder"] == eval_notes[b]["folder"]
                         for a, b in zip(*iu)])
        intra, inter = sims[iu][same], sims[iu][~same]
        auc = auc_rank(intra, inter)

        # per-note margin: best same-folder neighbour minus best
        # cross-folder neighbour (notes having at least one of each)
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
        med_intra, med_inter = float(np.median(intra)), float(np.median(inter))
        rows.append((v["name"], len(intra), len(inter), med_intra,
                     med_inter, auc,
                     float(np.median(margins)) if margins else float("nan")))
        print(f"    auc={auc:.3f} intra-med={med_intra:.3f} "
              f"inter-med={med_inter:.3f} margin-med="
              f"{np.median(margins):.3f}", flush=True)

        # phantom bridges: cross-folder pairs above the intra median
        thr = med_intra
        cand = np.where((~same) & (sims[iu] > thr))[0]
        cand = cand[np.argsort(-sims[iu][cand])]
        bl = []
        for c in cand[: args.bridges]:
            a, b = iu[0][c], iu[1][c]
            na, nb = eval_notes[a], eval_notes[b]
            bl.append((float(sims[iu][c]), na["title"], na["folder"],
                       nb["title"], nb["folder"]))
        bridges[v["name"]] = (thr, bl)

    f = ["# MODEL-1B — разделимость папок (сепарация)\n\n",
         "Гипотеза владельца: e5-base за счёт большего окна/ёмкости может ",
         "находить сходство там, где связи нет — «мосты» между чужими ",
         "темами размывают границы. Проверка на тех же 76 заметках с ",
         "папками.\n\n",
         "Метрика AUC: вероятность, что случайная пара из одной папки ",
         "ближе случайной пары из разных папок (Mann-Whitney по всем ",
         "парам оценочного множества). Выше — чище границы. Margin — ",
         "медиана по заметкам разности «лучший сосед своей папки − ",
         "лучший сосед чужой».\n\n",
         "| Вариант | пар своих | пар чужих | своих med | чужих med | ",
         "AUC | margin med |\n|---|---|---|---|---|---|---|\n"]
    for name, ni, nj, mi, mj, auc, mg in rows:
        f.append(f"| {name} | {ni} | {nj} | {mi:.3f} | {mj:.3f} | "
                 f"{auc:.4f} | {mg:.3f} |\n")
    f.append("\nМосты-фантомы (чужие пары выше медианы своих) — локально, "
             f"с названиями: `{local_dir.resolve()}/bridges.md`.\n")
    Path(args.findings).write_text("".join(f), encoding="utf-8")

    bl_out = ["# MODEL-1B: мосты-фантомы — чужие пары с близостью выше "
              "медианы своих\n"]
    for name, (thr, bl) in bridges.items():
        bl_out.append(f"\n## {name} (порог = медиана своих {thr:.3f})\n\n"
                      "| # | Близость | Заметка A | Папка A | Заметка B | "
                      "Папка B |\n|---|---|---|---|---|---|\n")
        for i, (s, ta, fa, tb, fb) in enumerate(bl, 1):
            bl_out.append(f"| {i} | {s:.3f} | {m.cell(ta, 60)} | "
                          f"{m.cell(fa, 25)} | {m.cell(tb, 60)} | "
                          f"{m.cell(fb, 25)} |\n")
    (local_dir / "bridges.md").write_text("".join(bl_out), encoding="utf-8")
    print("written:", args.findings, "and", local_dir / "bridges.md",
          flush=True)
    return 0


if __name__ == "__main__":
    sys.exit(main())
