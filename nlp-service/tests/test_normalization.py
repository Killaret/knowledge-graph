"""NLP-4: normalization module and /normalize endpoint tests."""
import os
import sys
import json

import numpy as np
import pytest
from unittest.mock import MagicMock, patch

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from app.core.normalization import (
    NormalizationParams,
    normalize,
    normalize_pass,
)
from app.main import app
from fastapi.testclient import TestClient

client = TestClient(app)

FIXTURE_PATH = os.path.join(
    os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))),
    "work-nlp4",
    "fixture_boilerplate.json",
)


def _long_text(sentence: str, repeat: int = 20) -> str:
    """Enough prose to stay above the 100-char guard."""
    return " ".join([sentence] * repeat)


class TestNormalizePass:
    def test_removes_boilerplate_lines(self):
        text = (
            "We use cookies. Accept our cookie policy to continue.\n"
            "Subscribe to our newsletter!\n"
            "Python decorators wrap functions to add behavior.\n"
            "(c) 2024 Example Corp. All rights reserved.\n"
            "Decorators are used for logging, auth, caching."
        )
        result = normalize_pass(text)
        assert "cookie" not in result.lower()
        assert "newsletter" not in result.lower()
        assert "rights reserved" not in result.lower()
        assert "decorators" in result

    def test_removes_bare_url_lines(self):
        text = "Real content stays.\nhttps://example.com/ads/banner\nMore content."
        result = normalize_pass(text)
        assert "https://" not in result
        assert "Real content stays." in result

    def test_removes_nav_run(self):
        text = (
            "Home\nAbout\nProducts\nPricing\nContact\n"
            + _long_text("Real content starts here and continues.")
        )
        result = normalize_pass(text)
        assert "Pricing" not in result
        assert "Real content starts here" in result

    def test_keeps_two_short_lines_not_nav(self):
        text = "Home\nAbout\n" + _long_text("Real paragraph content.")
        result = normalize_pass(text)
        assert "Home" in result and "About" in result

    def test_deduplicates_repeated_lines(self):
        text = "Line one.\nLine one.\nLine two is different."
        assert normalize_pass(text).count("Line one.") == 1

    def test_idempotent_fixed_point(self):
        text = (
            "Home\nAbout\nProducts\n"
            "We use cookies. Accept our cookie policy.\n"
            + _long_text("Substantive paragraph about graphs.")
        )
        once = normalize_pass(text)
        assert normalize_pass(once) == once


class TestNormalizeGuards:
    def test_rollback_when_result_too_short(self):
        # All boilerplate -> result collapses below min_chars.
        text = (
            "We use cookies. Accept our cookie policy to continue. "
            "Subscribe to our newsletter! " * 1
            + "x" * 120
        )
        text = "We use cookies. Accept our cookie policy.\n" * 10
        result = normalize(text)
        assert result.rolled_back is True
        assert result.rollback_reason == "too_short"
        assert result.normalized_text == text

    def test_rollback_when_cosine_below_threshold(self):
        text = _long_text("Original sentence about neural graphs.")
        embed_fn = MagicMock(return_value=[np.array([1.0, 0.0]), np.array([0.0, 1.0])])
        result = normalize(text, NormalizationParams(embed_fn=embed_fn, min_cosine=0.7))
        assert result.rolled_back is True
        assert result.rollback_reason == "low_cosine"
        assert result.normalized_text == text
        assert result.metrics["emb_cosine"] == 0.0

    def test_passes_when_cosine_above_threshold(self):
        text = _long_text("Original sentence about neural graphs.")
        embed_fn = MagicMock(return_value=[np.array([1.0, 0.0]), np.array([0.99, 0.01])])
        result = normalize(text, NormalizationParams(embed_fn=embed_fn, min_cosine=0.7))
        assert result.rolled_back is False
        assert result.rollback_reason is None
        assert result.metrics["emb_cosine"] > 0.99

    def test_without_embed_fn_only_length_guard_runs(self):
        text = _long_text("Sentence one differs. ") + "Tail."
        result = normalize(text)  # embed_fn=None
        assert result.rolled_back is False
        assert result.metrics["emb_cosine"] is None

    def test_skips_short_input(self):
        text = "Short note."
        result = normalize(text)
        assert result.skipped is True
        assert result.rolled_back is False
        assert result.normalized_text == text
        assert result.metrics["stop_reason"] == "skipped_short"

    def test_empty_input(self):
        result = normalize("")
        assert result.skipped is True
        assert result.normalized_text == ""

    def test_metrics_shape(self):
        text = _long_text("A sentence that survives normalization.")
        result = normalize(text)
        m = result.metrics
        assert m["iterations"] == 1
        assert m["stop_reason"] == "single_pass"
        assert m["raw_tokens"] >= m["norm_tokens"]
        assert 0 < m["compression"] <= 1.0


def _fake_model(emb_dim: int = 4):
    """Mock embedding model: identical texts get identical unit vectors —
    normalization of unchanged text then scores cosine 1.0."""
    model = MagicMock()
    model.max_seq_length = 128
    model.tokenizer = None

    def encode(texts, convert_to_numpy=True):
        rows = []
        for t in texts:
            v = np.full(emb_dim, 1.0 / np.sqrt(emb_dim), dtype=np.float32)
            v[0] = float(len(str(t)) % 7) / 10.0  # slight variation by length
            rows.append(v)
        return np.asarray(rows)

    model.encode.side_effect = encode
    return model


class TestNormalizeEndpoint:
    @patch("app.main.get_embedding_model")
    def test_normalize_returns_artifact(self, mock_get_model):
        mock_get_model.return_value = _fake_model()
        text = (
            "Home\nAbout\nProducts\nPricing\nContact\n"
            + _long_text("This note explains graph traversal in depth.")
        )
        response = client.post("/normalize", json={"text": text, "title": "Graphs"})
        assert response.status_code == 200
        data = response.json()
        assert data["pipeline_version"] == "norm-v1"
        assert data["rolled_back"] is False
        assert "Pricing" not in data["normalized_text"]
        assert data["metrics"]["stop_reason"] == "single_pass"
        assert data["metrics"]["iterations"] == 1
        assert isinstance(data["chunks"], list)
        if data["chunks"]:
            c = data["chunks"][0]
            assert set(c) >= {"idx", "text", "heading_path", "char_span", "token_count", "kind"}
            s, e = c["char_span"]
            assert data["normalized_text"][s:e] == c["text"]

    @patch("app.main.get_embedding_model")
    def test_normalize_rolled_back_returns_source(self, mock_get_model):
        model = _fake_model()
        # orthogonal vectors force the cosine guard to fire
        def ortho(texts, convert_to_numpy=True):
            rows = []
            for i, _ in enumerate(texts):
                v = np.zeros(4, dtype=np.float32)
                v[i % 4] = 1.0
                rows.append(v)
            return np.asarray(rows)
        model.encode.side_effect = ortho
        mock_get_model.return_value = model

        text = _long_text("Legitimate paragraph about recommender systems.")
        response = client.post("/normalize", json={"text": text, "title": ""})
        data = response.json()
        assert data["rolled_back"] is True
        assert data["rollback_reason"] == "low_cosine"
        assert data["normalized_text"] == text  # source returned unchanged

    @patch("app.main.get_embedding_model")
    def test_normalize_short_note_skipped(self, mock_get_model):
        mock_get_model.return_value = _fake_model()
        response = client.post("/normalize", json={"text": "tiny", "title": "t"})
        data = response.json()
        assert data["skipped"] is True
        assert data["normalized_text"] == "tiny"
        assert data["metrics"]["stop_reason"] == "skipped_short"
        # skipped input still chunks its (unchanged) text — artifact for Mongo
        assert len(data["chunks"]) == 1
        assert data["chunks"][0]["text"] == "tiny"

    @pytest.mark.skipif(
        not os.path.exists(FIXTURE_PATH), reason="local corpus fixture absent"
    )
    @patch("app.main.get_embedding_model")
    def test_normalize_on_boilerplate_fixture(self, mock_get_model):
        mock_get_model.return_value = _fake_model()
        fixture = json.load(open(FIXTURE_PATH, encoding="utf-8"))
        by_id = {n["id"]: n for n in fixture}
        r = client.post(
            "/normalize",
            json={"text": by_id["f1"]["content"], "title": by_id["f1"]["title"]},
        )
        data = r.json()
        assert r.status_code == 200
        out = data["normalized_text"]
        assert "cookies" not in out.lower()
        assert "newsletter" not in out.lower()
        assert "decorators" in out.lower()
