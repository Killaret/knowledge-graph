#!/usr/bin/env python3
"""W-1 eval harness: measure how well link-weight / recommendation scoring
matches ground truth (bookmark folder_path) on the owner's Personal corpus.

Usage:
    python eval_link_formula.py --dataset dataset.json [--synthetic synth.json] \
        --out report.md --raw raw.json

Scoring variants evaluated:
  - worker   : prod path — undirected BFS over link weights (path product x
               decay per extra hop, max-aggregation) * alpha + keyword * gamma;
               semantic component is a stub (0.0) in this path.
  - server   : composite loader — BFS over link edges*alpha PLUS top-N
               embedding neighbors as virtual edges (weight = cosine * beta).
  - graphsvc : graph-service — candidates only from transitive closure,
               score = alpha * weight * decay^distance + beta * cosine.
  - semantic : pure cosine similarity of doc embeddings (baseline).
  - keyword  : pure set-similarity over keywords (jaccard default).
  - combined : alpha*bfs_links + beta*cosine + gamma*keyword — the full
               three-component formula, grid-searched.

Metrics: precision@{1,3,5} at folder granularities exact / l1 / parent.
Autolink curve: precision@k for k=1..10 under pure semantic ranking.
Train/holdout: folder groups split in half; grid tuned on train, reported
on holdout.
"""

from __future__ import annotations

import argparse
import json
import math
import sys
from collections import defaultdict, deque
from pathlib import Path

DECAY = 0.5
DEPTH = 2
EMB_TOPN = 30
KS = (1, 3, 5)


def load_dataset(path: str) -> dict:
    with open(path, encoding="utf-8") as f:
        return json.load(f)


def parse_vec(text: str) -> list[float]:
    return [float(x) for x in text.strip("[]").split(",")]


def folder_l1(path: str) -> str:
    return path.split(" > ")[0].strip()


def folder_parent(path: str) -> str:
    parts = path.split(" > ")
    return " > ".join(parts[:-1]).strip() if len(parts) > 1 else path.strip()


GRANULARITIES = {
    "exact": lambda p: p.strip(),
    "l1": folder_l1,
    "parent": folder_parent,
}


def cosine(a: list[float], b: list[float]) -> float:
    dot = sum(x * y for x, y in zip(a, b))
    na = math.sqrt(sum(x * x for x in a))
    nb = math.sqrt(sum(y * y for y in b))
    if na == 0 or nb == 0:
        return 0.0
    return dot / (na * nb)


def jaccard(a: set, b: set) -> float:
    if not a and not b:
        return 1.0
    if not a or not b:
        return 0.0
    inter = len(a & b)
    return inter / (len(a) + len(b) - inter)


def overlap_coef(a: set, b: set) -> float:
    if not a and not b:
        return 1.0
    if not a or not b:
        return 0.0
    return len(a & b) / min(len(a), len(b))


def weighted_jaccard(kw_a: dict, kw_b: dict) -> float:
    if not kw_a and not kw_b:
        return 1.0
    if not kw_a or not kw_b:
        return 0.0
    keys = set(kw_a) | set(kw_b)
    mn = sum(min(kw_a.get(k, 0.0), kw_b.get(k, 0.0)) for k in keys)
    mx = sum(max(kw_a.get(k, 0.0), kw_b.get(k, 0.0)) for k in keys)
    return mn / mx if mx else 0.0


KEYWORD_METHODS = {
    "jaccard": lambda ka, kb: jaccard(set(ka), set(kb)),
    "overlap": lambda ka, kb: overlap_coef(set(ka), set(kb)),
    "weighted_jaccard": weighted_jaccard,
}


def bfs_paths(start: str, adj: dict[str, list[tuple[str, float]]],
              depth: int = DEPTH, decay: float = DECAY) -> dict[str, float]:
    """Replica of backend runBFS: path weight = product of edge weights,
    x decay for each hop beyond the first; max-aggregation over paths."""
    best = {start: (1.0, 0)}
    queue = deque([(start, 1.0, 0)])
    while queue:
        node, w, d = queue.popleft()
        if w < best.get(node, (0, 0))[0]:
            continue
        if d >= depth:
            continue
        for to, ew in adj.get(node, []):
            if to == start:
                continue
            nw = w * ew * (decay if d > 0 else 1.0)
            if nw > best.get(to, (0, 0))[0]:
                best[to] = (nw, d + 1)
                queue.append((to, nw, d + 1))
    del best[start]
    return {k: v[0] for k, v in best.items()}


def build_adj(links: list[dict], weight_scale: float = 1.0) -> dict[str, list[tuple[str, float]]]:
    """Undirected adjacency like backend neighborLoader (out+incoming reversed)."""
    adj: dict[str, list[tuple[str, float]]] = defaultdict(list)
    for l in links:
        s, t, w = l["src"], l["dst"], float(l["w"]) * weight_scale
        adj[s].append((t, w))
        adj[t].append((s, w))
    return adj


def score_pairs(eval_ids: list[str], all_ids: list[str], fn) -> dict[str, dict[str, float]]:
    """fn(a, b) -> score; returns {eval_id: {other_id: score}}."""
    out = {}
    for a in eval_ids:
        out[a] = {b: fn(a, b) for b in all_ids if b != a}
    return out


def precision_at_k(scores: dict[str, float], truth: set[str], k: int) -> float:
    top = sorted(scores.items(), key=lambda kv: kv[1], reverse=True)[:k]
    if not top:
        return 0.0
    return sum(1 for nid, _ in top if nid in truth) / len(top)


def evaluate(scores_map: dict[str, dict[str, float]],
             truth: dict[str, set[str]], k: int) -> float:
    vals = [precision_at_k(scores_map[a], truth[a], k) for a in scores_map]
    return sum(vals) / len(vals) if vals else 0.0


def build_truth(note_groups: dict[str, str], eval_ids: list[str]) -> dict[str, set[str]]:
    """truth[a] = set of note ids sharing a's group label."""
    by_group: dict[str, set[str]] = defaultdict(set)
    for nid, g in note_groups.items():
        if g:
            by_group[g].add(nid)
    return {a: (by_group.get(note_groups.get(a, ""), set()) - {a}) for a in eval_ids}


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--dataset", required=True)
    ap.add_argument("--synthetic", default=None)
    ap.add_argument("--out", required=True)
    ap.add_argument("--raw", default=None)
    args = ap.parse_args()

    data = load_dataset(args.dataset)
    notes = {n["id"]: n for n in data["notes"]}
    emb = {e["note_id"]: parse_vec(e["vec"]) for e in data["embeddings"]}
    kw = defaultdict(dict)
    for k in data["keywords"]:
        kw[k["note_id"]][k["kw"]] = float(k["w"])
    links = data["links"]

    synthetic = {}
    if args.synthetic:
        synthetic = load_dataset(args.synthetic)

    # real eval rows: notes with non-empty folder
    eval_ids = [nid for nid, n in notes.items() if (n.get("folder") or "").strip()]
    all_ids = list(notes.keys())

    lines = []
    raw = {}

    def report(title, fn_text):
        lines.append(f"## {title}\n")
        lines.append(fn_text)

    # --- semantic similarity matrix (lazy pairwise) ---
    def sem(a, b):
        if a not in emb or b not in emb:
            return 0.0
        return cosine(emb[a], emb[b])

    adj_links = build_adj(links)
    bfs_cache: dict[str, dict[str, float]] = {}

    def bfs_score(a, b):
        if a not in bfs_cache:
            bfs_cache[a] = bfs_paths(a, adj_links)
        return bfs_cache[a].get(b, 0.0)

    kw_methods_scores = {}
    for name, m in KEYWORD_METHODS.items():
        kw_methods_scores[name] = score_pairs(eval_ids, all_ids, lambda a, b: m(kw.get(a, {}), kw.get(b, {})))

    sem_scores = score_pairs(eval_ids, all_ids, sem)
    bfs_scores = score_pairs(eval_ids, all_ids, bfs_score)

    # composite server path: virtual embedding edges (top-N) * beta + link edges * alpha
    def composite_scores(alpha, beta):
        def score(a, b):
            # build virtual edges for a: top-N semantic neighbors
            emb_edges = sorted(
                ((nid, s) for nid, s in sem_scores[a].items()),
                key=lambda kv: kv[1], reverse=True)[:EMB_TOPN]
            comp_adj = defaultdict(list)
            for t, w in adj_links.get(a, []):
                comp_adj[a].append((t, w * alpha))
            for t, s in emb_edges:
                comp_adj[a].append((t, s * beta))
            # single-level expansion only needs direct edges plus one hop
            paths = bfs_paths(a, comp_adj)
            return paths.get(b, 0.0)
        return score_pairs(eval_ids, all_ids, score)

    # --- baseline table: single-component variants ---
    header = "| вариант | p@1 exact | p@3 exact | p@5 exact | p@3 l1 | p@3 parent |\n|---|---|---|---|---|---|\n"
    rows = []

    # groups per granularity
    results = {}
    for gname, gfn in GRANULARITIES.items():
        groups = {nid: gfn(n.get("folder") or "") for nid, n in notes.items()}
        truth = build_truth(groups, eval_ids)
        for vname, smap in [
            ("semantic", sem_scores),
            ("bfs_links (worker α)", bfs_scores),
            *[(f"keyword/{m}", kw_methods_scores[m]) for m in KEYWORD_METHODS],
        ]:
            for k in KS:
                results.setdefault(vname, {})[f"{gname}@{k}"] = evaluate(smap, truth, k)

    for vname, metrics in results.items():
        rows.append("| " + " | ".join([
            vname,
            f"{metrics.get('exact@1', 0):.3f}",
            f"{metrics.get('exact@3', 0):.3f}",
            f"{metrics.get('exact@5', 0):.3f}",
            f"{metrics.get('l1@3', 0):.3f}",
            f"{metrics.get('parent@3', 0):.3f}",
        ]) + " |")

    report("Базовые варианты (реальные заметки)", header + "\n".join(rows) + "\n")

    # --- grid search combined: alpha*bfs + beta*sem + gamma*kw(jaccard) ---
    grid_vals = [0.0, 0.25, 0.5, 0.75, 1.0]
    groups_exact = {nid: (n.get("folder") or "").strip() for nid, n in notes.items()}
    truth_exact = build_truth(groups_exact, eval_ids)

    # split folders for train/holdout: even/odd by sorted folder name
    folders_sorted = sorted({g for g in groups_exact.values() if g})
    train_folders = set(folders_sorted[::2])
    holdout_folders = set(folders_sorted[1::2])
    train_ids = [nid for nid in eval_ids if groups_exact[nid] in train_folders]
    holdout_ids = [nid for nid in eval_ids if groups_exact[nid] in holdout_folders]
    truth_train = build_truth(groups_exact, train_ids)
    truth_hold = build_truth(groups_exact, holdout_ids)

    grid_rows = []
    best = None
    for a in grid_vals:
        for b in grid_vals:
            for g in grid_vals:
                if a + b + g == 0:
                    continue
                na, nb, ng = a / (a + b + g), b / (a + b + g), g / (a + b + g)
                smap = {}
                for nid in train_ids:
                    smap[nid] = {
                        o: na * bfs_scores[nid].get(o, 0.0)
                        + nb * sem_scores[nid].get(o, 0.0)
                        + ng * kw_methods_scores["jaccard"][nid].get(o, 0.0)
                        for o in all_ids if o != nid
                    }
                p = evaluate(smap, truth_train, 3)
                grid_rows.append((p, na, nb, ng))
                if best is None or p > best[0]:
                    best = (p, na, nb, ng)
    grid_rows.sort(reverse=True)
    grid_tbl = "| α граф | β семантика | γ keywords | p@3 exact (train) |\n|---|---|---|---|\n"
    grid_tbl += "\n".join(
        f"| {a:.2f} | {b:.2f} | {g:.2f} | {p:.3f} |" for p, a, b, g in grid_rows[:10]
    ) + "\n"
    report("Grid-search α/β/γ (train-половина папок)", grid_tbl)

    # holdout check for best weights
    p, na, nb, ng = best
    smap_hold = {}
    for nid in holdout_ids:
        smap_hold[nid] = {
            o: na * bfs_scores[nid].get(o, 0.0)
            + nb * sem_scores[nid].get(o, 0.0)
            + ng * kw_methods_scores["jaccard"][nid].get(o, 0.0)
            for o in all_ids if o != nid
        }
    hold_p = evaluate(smap_hold, truth_hold, 3)
    # apples-to-apples: single components on the same holdout subset
    sem_hold = evaluate({n: sem_scores[n] for n in holdout_ids}, truth_hold, 3)
    kw_hold = evaluate({n: kw_methods_scores["jaccard"][n] for n in holdout_ids}, truth_hold, 3)
    bfs_hold = evaluate({n: bfs_scores[n] for n in holdout_ids}, truth_hold, 3)
    report("Holdout-проверка лучших весов",
           f"Лучшие веса на train: α={na:.2f} β={nb:.2f} γ={ng:.2f} (p@3={p:.3f}). "
           f"На holdout-папках: **p@3={hold_p:.3f}** "
           f"(train {len(train_ids)} заметок, holdout {len(holdout_ids)}).\n\n"
           f"На той же holdout-выборке: semantic p@3={sem_hold:.3f}, "
           f"keyword/jaccard p@3={kw_hold:.3f}, bfs_links p@3={bfs_hold:.3f}.\n")
    raw["best_weights"] = {"alpha": na, "beta": nb, "gamma": ng,
                           "train_p3": p, "holdout_p3": hold_p,
                           "holdout_baselines": {"semantic": sem_hold,
                                                 "keyword": kw_hold,
                                                 "bfs": bfs_hold}}

    # graph coverage: how many eval notes are reachable at all
    covered = [nid for nid in eval_ids if bfs_cache.get(nid) or bfs_paths(nid, adj_links)]
    neigh_counts = [len(bfs_cache.get(nid) or bfs_paths(nid, adj_links)) for nid in eval_ids]
    report("Покрытие графа",
           f"Заметок с хотя бы одним достижимым соседом по связям: "
           f"{len(covered)}/{len(eval_ids)} ({100*len(covered)/len(eval_ids):.0f}%). "
           f"Среднее число достижимых соседей: {sum(neigh_counts)/len(neigh_counts):.1f}.\n")
    raw["graph_coverage"] = {"covered": len(covered), "total": len(eval_ids),
                             "mean_neighbors": sum(neigh_counts)/len(neigh_counts)}

    # --- autolink curve: semantic top-k, precision per k + score ---
    curve = []
    for k in range(1, 11):
        pv = evaluate(sem_scores, truth_exact, k)
        # mean score of k-th neighbor
        kth = []
        for nid in eval_ids:
            top = sorted(sem_scores[nid].items(), key=lambda kv: kv[1], reverse=True)
            if len(top) >= k:
                kth.append(top[k - 1][1])
        curve.append((k, pv, sum(kth) / len(kth) if kth else 0.0))
    ctbl = "| k | p@k exact | средний score k-го соседа |\n|---|---|---|\n"
    ctbl += "\n".join(f"| {k} | {p:.3f} | {s:.3f} |" for k, p, s in curve) + "\n"
    report("Кривая автолинка (чистая семантика)", ctbl)
    raw["autolink_curve"] = [{"k": k, "p": p, "score": s} for k, p, s in curve]

    # --- server composite variant ---
    comp = composite_scores(0.5, 0.5)
    comp_p = {k: evaluate(comp, truth_exact, k) for k in KS}
    report("Server composite (BFS: links*0.5 + emb-top30*0.5)",
           f"p@1={comp_p[1]:.3f} p@3={comp_p[3]:.3f} p@5={comp_p[5]:.3f}\n")
    raw["composite_p"] = comp_p

    # --- synthetic run ---
    if synthetic:
        sn = synthetic["notes"]
        semb = {e["note_id"]: parse_vec(e["vec"]) for e in synthetic["embeddings"]}
        sids = [n["id"] for n in sn]
        sgroups = {n["id"]: n["cluster"] for n in sn}
        struth = build_truth(sgroups, sids)
        ssem = score_pairs(sids, sids, lambda a, b: cosine(semb[a], semb[b]))
        synth_p = {k: evaluate(ssem, struth, k) for k in KS}
        report("Синтетика (чистая семантика)",
               f"{len(sids)} заметок, {len(set(sgroups.values()))} кластеров. "
               f"p@1={synth_p[1]:.3f} p@3={synth_p[3]:.3f} p@5={synth_p[5]:.3f}\n")
        raw["synthetic_p"] = synth_p

    raw["baseline"] = results
    raw["grid_top10"] = [
        {"alpha": a, "beta": b, "gamma": g, "train_p3": p}
        for p, a, b, g in grid_rows[:10]
    ]

    meta = (f"# W-1 eval findings\n\nКорпус: {len(notes)} заметок, "
            f"{len(eval_ids)} с папкой, {len(emb)} векторов, {len(links)} связей.\n\n")
    Path(args.out).write_text(meta + "\n".join(lines), encoding="utf-8")
    if args.raw:
        Path(args.raw).write_text(json.dumps(raw, ensure_ascii=False, indent=2), encoding="utf-8")
    print(f"written: {args.out}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
