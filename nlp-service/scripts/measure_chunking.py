#!/usr/bin/env python3
"""CHUNK-1: замер прототипа структурного чанкера на корпусе.

Прототип алгоритма из docs/tasks/CHUNK-1-structure-aware-chunker.md:
иерархия границ (структура -> предложения -> клаузы), жёсткий инвариант
len(chunk) <= max_tokens, код/таблицы атомарны предпочтительно.

Замеряет:
- статистику резки: чанков на заметку, токены, forced_split, виды блоков;
- агрегацию doc-вектора: mean vs weighted (первый чанк тяжелее);
- title-инъекцию: сдвиг doc-вектора при подстановке title в каждый чанк;
- потерю текущего пути: cos(aggregated, emb(truncated whole text)).

Датасет: [{"id","title","content"}] — дамп без коммита (work-*/.gitignore).

Запуск с эмбеддингами — внутри kg-test-nlp (модель в кэше):
    docker cp ... && docker exec kg-test-nlp python /tmp/mc.py \
        --dataset /tmp/notes.json --with-embeddings --out /tmp/chunk.md
"""
import argparse
import json
import re
import sys

# --- Прототип чанкера (спека CHUNK-1, v0 для замера) ---

ABBREV = {
    "ru": {"т.е", "т. е", "т.к", "т. к", "т.п", "т. п", "др", "г", "им",
           "ср", "напр", "см", "рис", "ст", "руб", "млн", "млрд", "гг", "в", "вв"},
    "en": {"i.e", "e.g", "etc", "vs", "mr", "mrs", "dr", "st", "no", "fig"},
}
SENT_END = ".!?…"
CLOSERS = "»\"')]"
CLAUSE_SEP = ";:,"


def split_sentences(text):
    """Границы предложений по знакам препинания с охраной от ложных
    срабатываний: сокращения, десятичные, инициалы, URL."""
    out, start = [], 0
    for m in re.finditer(r"[.!?…]+[" + re.escape(CLOSERS) + r"]*\s+", text):
        end = m.end()
        frag = text[start:end].rstrip()
        tail = frag.split()[-1].lower().rstrip(".") if frag.split() else ""
        prev = text[max(0, m.start() - 20):m.start()]
        if (tail in ABBREV["ru"] | ABBREV["en"]
                or re.search(r"\d\.$", frag) and re.match(r"\s*\d", text[end:end + 3])
                or re.search(r"\b[A-ZА-ЯЁ]\.$", frag)        # инициал: "А."
                or re.search(r"(https?|www|ftp)[^\s]*\.$", prev)):
            continue
        out.append(frag)
        start = end
    if text[start:].strip():
        out.append(text[start:].strip())
    return [s for s in out if s]


def parse_blocks(text):
    """Структурные блоки: код-фенсы, таблицы, заголовки, абзацы.
    Возвращает [(kind, heading_path, text)]."""
    blocks = []
    heading_path = []
    pos = 0
    # грубая, но честная разметка: ``` fences и | таблицы атомарны
    pattern = re.compile(
        r"(```.*?```|(?:^\|.*\|$\n?)+|^#{1,6}[^\n]*$)",
        re.S | re.M)
    for m in pattern.finditer(text):
        if text[pos:m.start()].strip():
            for para in re.split(r"\n{2,}", text[pos:m.start()]):
                if para.strip():
                    blocks.append(("prose", list(heading_path), para.strip()))
        chunk = m.group(0)
        if chunk.startswith("```"):
            blocks.append(("code", list(heading_path), chunk))
        elif chunk.lstrip().startswith("|"):
            blocks.append(("table", list(heading_path), chunk))
        else:  # заголовок: обновляет путь, сам не чанк
            level = len(chunk) - len(chunk.lstrip("#"))
            title = chunk.lstrip("#").strip()
            heading_path = heading_path[: level - 1] + [title]
        pos = m.end()
    if text[pos:].strip():
        for para in re.split(r"\n{2,}", text[pos:]):
            if para.strip():
                blocks.append(("prose", list(heading_path), para.strip()))
    return blocks


def chunk_text(text, tok, target=256, max_tokens=512):
    """Пакует предложения в чанки: жадно до target, потолок max.
    Возвращает [{text, heading_path, char_span, token_count, kind,
    forced_split}]. char_span — в исходном тексте."""
    chunks = []
    cursor = 0
    for kind, hpath, block in parse_blocks(text):
        units = [block] if kind in ("code", "table") else split_sentences(block)
        if not units:
            continue
        buf, buf_tok, buf_start = [], 0, None
        for u in units:
            u_tok = tok(u)
            u_pos = text.find(u, cursor)
            if u_pos < 0:
                u_pos = cursor
            if u_tok > max_tokens:
                # блок/предложение само > max — резать внутри
                if buf:
                    chunks.append(_emit(buf, buf_tok, hpath, kind, buf_start, text, u_pos))
                    buf, buf_tok, buf_start = [], 0, None
                chunks.extend(_force_split(u, u_tok, hpath, kind, tok, max_tokens, text, u_pos))
                cursor = u_pos + len(u)
                continue
            if buf and buf_tok + u_tok > target:
                chunks.append(_emit(buf, buf_tok, hpath, kind, buf_start, text, u_pos))
                buf, buf_tok, buf_start = [], 0, None
            if not buf:
                buf_start = u_pos
            buf.append(u)
            buf_tok += u_tok
            cursor = u_pos + len(u)
        if buf:
            end_pos = cursor
            chunks.append(_emit(buf, buf_tok, hpath, kind, buf_start, text, end_pos))
    return chunks


def _emit(units, tok_count, hpath, kind, start, full, end):
    return {"text": " ".join(units), "heading_path": hpath, "kind": kind,
            "char_span": [start, end], "token_count": tok_count,
            "forced_split": False}


def _force_split(unit, u_tok, hpath, kind, tok, max_tokens, full, pos):
    """Единственный юнит > max: код/таблица — по строкам, проза — по клаузам,
    крайний случай — по словам. forced_split=True."""
    sep = "\n" if kind in ("code", "table") else None
    parts = unit.split("\n") if sep else re.split(r"(?<=[;:,\u2026])\s+", unit)
    if all(tok(p) <= max_tokens for p in parts if p.strip()):
        sub = []
        buf, buf_tok = [], 0
        for p in parts:
            pt = tok(p)
            if buf and buf_tok + pt > max_tokens:
                sub.append("\n".join(buf) if sep else " ".join(buf))
                buf, buf_tok = [], 0
            buf.append(p)
            buf_tok += pt
        if buf:
            sub.append("\n".join(buf) if sep else " ".join(buf))
    else:  # крайний случай: жёстко по словам
        words = unit.split()
        sub, cur, cur_tok = [], [], 0
        for w in words:
            wt = tok(w)
            if cur and cur_tok + wt > max_tokens:
                sub.append(" ".join(cur))
                cur, cur_tok = [], 0
            cur.append(w)
            cur_tok += wt
        if cur:
            sub.append(" ".join(cur))
    return [{"text": s, "heading_path": hpath, "kind": kind,
             "char_span": [pos, pos + len(unit)], "token_count": tok(s),
             "forced_split": True} for s in sub if s.strip()]


def word_tokens(text):
    return max(1, len(re.findall(r"[\w]+", text, re.U)))


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--dataset", required=True)
    ap.add_argument("--out", required=True)
    ap.add_argument("--raw", default=None)
    ap.add_argument("--target", type=int, default=256)
    ap.add_argument("--max-tokens", type=int, default=None,
                    help="потолок чанка; по умолчанию окно модели")
    ap.add_argument("--model", default="sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2")
    ap.add_argument("--with-embeddings", action="store_true")
    args = ap.parse_args()

    notes = json.load(open(args.dataset, encoding="utf-8"))
    model = None
    if args.with_embeddings:
        from sentence_transformers import SentenceTransformer
        model = SentenceTransformer(args.model)
        tok = lambda t: len(model.tokenizer.encode(t, add_special_tokens=False))
    else:
        tok = word_tokens
    max_tokens = args.max_tokens or (model.max_seq_length if model else 10**9)
    if args.target > max_tokens:
        sys.exit(f"target {args.target} > max {max_tokens} — чанки не влезут в окно")

    import statistics as st
    n_chunks, tok_counts, forced, kinds = [], [], 0, {}
    per_note = {}
    for n in notes:
        chs = chunk_text(n["content"], tok, args.target, max_tokens)
        per_note[n["id"]] = chs
        n_chunks.append(len(chs))
        tok_counts += [c["token_count"] for c in chs]
        forced += sum(1 for c in chs if c["forced_split"])
        for c in chs:
            kinds[c["kind"]] = kinds.get(c["kind"], 0) + 1

    lines = [f"\n## Чанкер: статистика резки ({len(notes)} заметок, "
             f"target={args.target}, max={max_tokens}, "
             f"model={args.model if model else 'word-count'})",
             f"- чанков/заметку: med {st.median(n_chunks):.0f}, mean {st.mean(n_chunks):.1f}, "
             f"max {max(n_chunks)}",
             f"- токенов/чанк: med {st.median(tok_counts):.0f}, "
             f"p90 {sorted(tok_counts)[int(len(tok_counts)*0.9)]:.0f}, max {max(tok_counts)}",
             f"- forced_split: {forced} чанков",
             f"- виды блоков: {kinds}"]

    if model:
        import numpy as np
        nrm = lambda v: v / max(np.linalg.norm(v), 1e-9)
        drift = []          # (id, cos_mean, cos_weighted)
        title_shift = []
        docs_mean, docs_trunc, docs_title_all = {}, {}, {}
        for n in notes:
            chs = per_note[n["id"]]
            if not chs:
                continue
            vecs = np.asarray(model.encode(
                [c["text"] for c in chs], convert_to_numpy=True), dtype=np.float32)
            vecs_t = np.asarray(model.encode(
                [n["title"] + "\n\n" + c["text"] for c in chs],
                convert_to_numpy=True), dtype=np.float32)
            doc_mean = nrm(vecs.mean(axis=0))
            w = np.array([2.0 if i == 0 else 1.0 for i in range(len(chs))])
            doc_w = nrm((vecs * w[:, None]).sum(axis=0) / w.sum())
            doc_title = nrm(vecs_t.mean(axis=0))
            trunc = nrm(np.asarray(
                model.encode(n["content"], convert_to_numpy=True), dtype=np.float32))
            drift.append((n["id"], float(doc_mean @ trunc), float(doc_w @ trunc)))
            title_shift.append(float(doc_title @ doc_mean))
            docs_mean[n["id"]] = doc_mean
            docs_trunc[n["id"]] = trunc
            docs_title_all[n["id"]] = doc_title

        dm = [d[1] for d in drift]
        dw = [d[2] for d in drift]
        lines += ["",
                  "### Doc-вектор vs текущий усечённый /embed",
                  f"- cos(mean_agg, truncated): med {st.median(dm):.3f}, "
                  f"min {min(dm):.3f} — доля смысла, которую теряет усечение",
                  f"- cos(weighted_agg, truncated): med {st.median(dw):.3f}, "
                  f"min {min(dw):.3f}",
                  f"- cos(doc с title-инъекцией, doc без): med {st.median(title_shift):.3f}, "
                  f"min {min(title_shift):.3f}"]

        # Парное сходство заметок: меняются ли top-2 соседа (семантика gamma-связей)
        ids = [n["id"] for n in notes if per_note[n["id"]]]
        titles = {n["id"]: n["title"] for n in notes}
        def top2(docs):
            M = np.stack([docs[i] for i in ids])
            S = M @ M.T
            np.fill_diagonal(S, -1)
            top = np.argsort(-S, axis=1)[:, :2]
            return ({ids[i]: [float(S[i, j]) for j in top[i]]
                     for i in range(len(ids))},
                    {ids[i]: set(ids[j] for j in top[i])
                     for i in range(len(ids))})
        (t_trunc, n_trunc), (t_mean, n_mean) = top2(docs_trunc), top2(docs_mean)
        changed = sum(1 for i in ids if n_trunc[i] != n_mean[i])
        sims_t = [s for v in t_trunc.values() for s in v]
        sims_m = [s for v in t_mean.values() for s in v]
        lines += ["",
                  "### Соседи top-2 (прокси gamma-связей): усечение vs чанки",
                  f"- top-2 score: truncated med {st.median(sims_t):.3f} → "
                  f"chunked med {st.median(sims_m):.3f}",
                  f"- заметок, у которых сменился хотя бы один top-2 сосед: "
                  f"{changed}/{len(ids)} ({changed/len(ids)*100:.0f}%)"]

        # title-инъекция: влияние на соседей
        if docs_title_all:
            (t_ti, n_ti) = top2(docs_title_all)
            changed_ti = sum(1 for i in ids if n_mean[i] != n_ti[i])
            sims_ti = [s for v in t_ti.values() for s in v]
            lines += ["",
                      "### Title-инъекция: влияние на top-2",
                      f"- сменился top-2 сосед: {changed_ti}/{len(ids)} "
                      f"({changed_ti/len(ids)*100:.0f}%)",
                      f"- top-2 score: без title med {st.median(sims_m):.3f} → "
                      f"с title med {st.median(sims_ti):.3f}"]

        lines += ["", "### Хвост дрейфа (худшие 5 по mean_agg)"]
        for nid, cm, _ in sorted(drift, key=lambda d: d[1])[:5]:
            lines.append(f"- {cm:.3f} | {titles.get(nid, '?')[:60]} "
                         f"| {len(per_note[nid])} чанков")

    report = "\n".join(lines) + "\n"
    with open(args.out, "a", encoding="utf-8") as f:
        f.write(report)
    if args.raw:
        json.dump({"per_note": {k: [{"chars": len(c["text"]),
                                     "tok": c["token_count"],
                                     "kind": c["kind"],
                                     "forced": c["forced_split"]}
                                    for c in v] for k, v in per_note.items()}},
                  open(args.raw, "w", encoding="utf-8"), ensure_ascii=False)
    print(f"done: {len(notes)} notes -> {args.out}")
    print(report)


if __name__ == "__main__":
    main()
