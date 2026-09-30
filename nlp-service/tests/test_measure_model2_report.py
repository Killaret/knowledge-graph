"""MODEL-2: findings generator keeps every variant in every table and
carries the matching-precision threshold row — regression for the class
of assembly bugs MODEL-1 had (per-variant runs silently dropping rows)."""
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent.parent / "scripts"))

import measure_model2 as m2


def _fake_result(name):
    return {
        "variant": {"name": name, "model": "fake", "prod": True,
                    "pp": "", "qp": "", "candidate": name != "A"},
        "precision": {g: {1: 0.5, 3: 0.4, 5: 0.3}
                      for g in ("exact", "l1", "parent")},
        "per_note_p": {"exact": [0.0, 1.0]},
        "curve": [{"t": t, "linked_share": 0.5, "links": 10,
                   "precision": 0.8} for t in m2.THRESHOLDS],
        "match_t": {"t": 0.75, "linked_share": 0.5, "precision": 0.8},
        "dist": {"first_median": 0.7, "first_mean": 0.7,
                 "second_median": 0.6, "second_mean": 0.6},
        "sep": {"auc": 0.9, "intra_n": 10, "inter_n": 90,
                "intra_med": 0.7, "inter_med": 0.3, "margin_med": 0.4},
        "stats": {"chunks_per_note": [1], "zero_chunk": 0,
                  "norm_rollback": 0},
        "top2": {},
    }


def _build(names):
    results = {n: _fake_result(n) for n in names}
    timing = {n: {"note_p50_ms": 1.0, "note_p95_ms": 2.0,
                  "chunk_p50_ms": 0.5, "chunks_total": 3,
                  "model_rss_mb": 100.0, "model_delta_mb": 50.0}
              for n in names}
    boot = {"F − E": (0.01, 0.05)} if "E" in results and "F" in results else {}
    notes = [{"id": f"id-{i}", "title": "t", "content": "c", "folder": "f"}
             for i in range(2)]
    return m2.build_findings(results, names, notes, ["id-0"], boot, timing,
                             {"torch_threads": 1, "OMP_NUM_THREADS": "",
                              "cpu_count": 4}, "", "local-dir")


def test_all_variants_present_in_all_tables():
    names = ["A", "D", "E", "F"]
    text = _build(names)
    for section in ("Точность по папкам", "Сепарация папок",
                    "Кривая порога", "Порог по совпадающей точности",
                    "Распределение близости", "Скорость и память"):
        body = text.split(section)[1]
        for name in names:
            assert f"| {name} |" in body, f"{name} missing in {section}"


def test_matching_precision_rule_row():
    text = _build(["D", "E", "F"])
    assert "0.75" in text.split("Порог по совпадающей точности")[1]
    assert "0.71" in text  # prod anchor named in the rule


def test_partial_run_lists_only_measured_variants():
    text = _build(["D", "F"])
    assert "| E |" not in text
    assert "| D |" in text and "| F |" in text
