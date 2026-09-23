"""MODEL-1B: the report generator must keep the pair table (таблица 2)
for every measured variant — regression test for the MODEL-1 assembly bug
where per-variant `--only` runs each appended a one-variant report and the
merged document kept only the first variant's pair table."""
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent.parent / "scripts"))

import measure_models as mm


def _fake_result(tag):
    return {
        "params": {"hidden": 384, "max_pos": 512,
                   "file_window": 128, "window_used": 128},
        "model_mb": 100.0, "proc_mb": 500.0, "encode_all_s": 1.5,
        "p50_ms": 10.0, "p95_ms": 20.0, "kw_p50_ms": 5.0, "kw_p95_ms": 9.0,
        "search": [[f"{tag}-q{q}-r{r}" for r in range(10)]
                   for q in range(len(mm.QUERIES))],
        "pairs": [(0.9 - i * 0.01, f"{tag}-note-a-{i}", f"{tag}-note-b-{i}")
                  for i in range(15)],
        "kw": {},
    }


def test_pair_tables_present_for_all_variants(tmp_path):
    results = {"fake-a": _fake_result("fake-a"),
               "fake-b": _fake_result("fake-b")}
    out = tmp_path / "report.md"
    mm.write_report(str(out), results, ["fake-a", "fake-b"], notes=[])

    text = out.read_text(encoding="utf-8")
    table2 = text.split("## Таблица 2.")[1].split("## Таблица 3.")[0]
    for name in ("fake-a", "fake-b"):
        assert f"### {name}" in table2, f"pair section missing for {name}"
        assert f"{name}-note-a-0" in table2
        assert f"{name}-note-b-14" in table2


def test_state_merge_regenerates_all_pair_sections(tmp_path):
    """Two --only-style runs accumulate in state; the report rewritten from
    the merged state still carries both pair tables."""
    state = {}
    for name in ("fake-a", "fake-b"):
        state.update({name: _fake_result(name)})
        out = tmp_path / "report.md"
        mm.write_report(str(out), state,
                        [n for n in ("fake-a", "fake-b") if n in state],
                        notes=[])
    text = out.read_text(encoding="utf-8")
    table2 = text.split("## Таблица 2.")[1].split("## Таблица 3.")[0]
    assert table2.count("| # | Пара | cos |") == 2
    for name in ("fake-a", "fake-b"):
        assert f"### {name}" in table2
