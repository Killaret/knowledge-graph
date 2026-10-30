import pytest
import sys
import os
from unittest.mock import patch, MagicMock

import numpy as np

# Add the parent directory to the path to import app modules
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import app.nlp_utils as nlp_utils
from app.nlp_utils import (
    extract_keywords,
    lemmatize_phrase,
    get_embedding_model,
    is_model_loaded,
    ensure_model_loaded,
)
from app.models import (
    ExtractKeywordsRequest,
    ExtractKeywordsResponse,
    Keyword,
    EmbedRequest,
    EmbedResponse,
)


def reset_model_state():
    """Reset global model state for isolated tests."""
    nlp_utils._embedding_model = None
    nlp_utils._model_load_error = None


class UniformModel:
    """Deterministic fake embedding model: every text gets the same unit
    vector, so all candidates score similarity 1.0 and ranking falls back to
    the (deterministic) statistical short-list order. Lets unit tests verify
    lemmatization, dedup and contract without downloading a real model."""

    DIM = 8

    def encode(self, texts, convert_to_numpy=True):
        if isinstance(texts, str):
            texts = [texts]
        v = np.ones(self.DIM, dtype=np.float32)
        return np.tile(v, (len(texts), 1))


@pytest.fixture()
def fake_model():
    """Patch the singleton so extract_keywords runs hermetically."""
    with patch("app.nlp_utils.get_embedding_model", return_value=UniformModel()):
        yield


@pytest.fixture(scope="module")
def embedding_model():
    """Load model once for integration tests; skip if unavailable."""
    nlp_utils._embedding_model = None
    nlp_utils._model_load_error = None
    prev_offline = os.environ.get("HF_HUB_OFFLINE")
    os.environ["HF_HUB_OFFLINE"] = "0"
    try:
        model = get_embedding_model()
    except Exception as exc:
        pytest.skip(f"Embedding model unavailable: {exc}")
    finally:
        if prev_offline is None:
            os.environ.pop("HF_HUB_OFFLINE", None)
        else:
            os.environ["HF_HUB_OFFLINE"] = prev_offline
    return model


class TestLemmatization:
    def test_english_verb_lemma(self):
        # WordNet: verb form resolves to infinitive
        assert lemmatize_phrase("running") == "run"

    def test_english_plural_lemma(self):
        assert lemmatize_phrase("networks") == "network"

    def test_russian_lemma(self):
        assert lemmatize_phrase("был") == "быть"
        assert lemmatize_phrase("деревья") == "дерево"

    def test_russian_phrase_lemma(self):
        lemma = lemmatize_phrase("баз данных")
        assert lemma.split()[0] == "база"
        assert lemma.split()[1] == "данные"

    def test_mixed_language_token_survives(self):
        # Hyphenated latin+cyrillic token is routed by cyrillic and must not
        # crash the lemmatizer.
        lemma = lemmatize_phrase("gpt-модель")
        assert isinstance(lemma, str)
        assert len(lemma) > 0

    def test_punctuation_and_digits_pass_through(self):
        lemma = lemmatize_phrase("2024")
        assert lemma == "2024"


class TestKeywordExtraction:
    def test_extract_returns_triples(self, fake_model):
        text = "Machine learning is a subset of artificial intelligence that focuses on neural networks."
        result = extract_keywords(text, top_n=5)

        assert isinstance(result, list)
        assert 0 < len(result) <= 5
        for lemma, surface, weight in result:
            assert isinstance(lemma, str) and lemma.strip()
            assert isinstance(surface, str) and surface.strip()
            assert isinstance(weight, float)
            assert 0.0 <= weight <= 1.0

    def test_empty_text(self, fake_model):
        assert extract_keywords("", top_n=5) == []
        assert extract_keywords("   ", top_n=5) == []
        assert extract_keywords(None, top_n=5) == []

    def test_no_keyword_text(self, fake_model):
        """Text consisting only of stopwords yields no candidates."""
        assert extract_keywords("the and of to in", top_n=5) == []

    def test_very_short_text(self, fake_model):
        result = extract_keywords("кот", top_n=5)
        assert isinstance(result, list)
        assert len(result) <= 5

    def test_english_normalization_in_output(self, fake_model):
        text = "Running networks. The runner runs through connected networks daily."
        result = extract_keywords(text, top_n=10)
        lemmas = {lemma for lemma, _, _ in result}
        # Inflected forms collapse to lemmas
        assert "run" in lemmas or "network" in lemmas
        assert "networks" not in lemmas
        assert "running" not in lemmas

    def test_russian_normalization_in_output(self, fake_model):
        text = "Деревья были зелёными. Дерево росло у дороги, деревья шумели."
        result = extract_keywords(text, top_n=10)
        lemmas = {lemma for lemma, _, _ in result}
        assert "дерево" in lemmas
        assert "зелёный" in lemmas
        assert "деревья" not in lemmas
        # "были" is in the NLTK russian stopword list and must be dropped
        assert "быть" not in lemmas
        assert "были" not in lemmas

    def test_surface_distinct_from_lemma(self, fake_model):
        """The surface form keeps the spelling from the text while the lemma
        is normalized — 'деревья' must surface as written but store lemma
        'дерево'."""
        text = "Деревья были зелёными, деревья шумели на ветру."
        result = extract_keywords(text, top_n=10)
        pairs = {lemma: surface for lemma, surface, _ in result}
        assert "дерево" in pairs
        assert pairs["дерево"] == "деревья"

    def test_dedup_by_lemma(self, fake_model):
        """Two surface forms of the same lemma produce a single row."""
        text = "Cats are running. The cat runs fast, cats everywhere."
        result = extract_keywords(text, top_n=10)
        lemmas = [lemma for lemma, _, _ in result]
        assert len(lemmas) == len(set(lemmas))

    def test_mixed_language_text(self, fake_model):
        text = "Модель gpt-модель обучается на данных, transformer architecture works."
        result = extract_keywords(text, top_n=10)
        assert isinstance(result, list)
        for lemma, surface, _ in result:
            assert lemma.strip()
            assert surface.strip()

    def test_top_n_limits(self, fake_model):
        text = (
            "Alpha beta gamma delta epsilon zeta eta theta iota kappa. "
            "Alpha beta gamma delta epsilon zeta. Alpha beta gamma."
        )
        assert len(extract_keywords(text, top_n=3)) <= 3
        assert len(extract_keywords(text, top_n=10)) <= 10

    def test_weights_sorted_desc(self, fake_model):
        text = "Machine learning neural networks deep learning machine learning networks."
        result = extract_keywords(text, top_n=10)
        weights = [w for _, _, w in result]
        assert weights == sorted(weights, reverse=True)

    def test_closer_candidate_has_higher_weight(self):
        """Direction of the score: the candidate aligned with the document
        vector must get a HIGHER weight — guards the `weight = 1 - score`
        mutation (yake semantics) that a range-only test would miss."""

        class DirectionalModel:
            def encode(self, texts, convert_to_numpy=True):
                import numpy as np

                if isinstance(texts, str):
                    texts = [texts]
                rows = []
                for t in texts:
                    rows.append([1.0, 0.0] if "alpha" in t.lower() else [0.0, 1.0])
                return np.asarray(rows, dtype=np.float32)

        with patch(
            "app.nlp_utils.get_embedding_model", return_value=DirectionalModel()
        ):
            result = extract_keywords(
                "alpha alpha alpha beta gamma delta epsilon", top_n=10
            )
        by_lemma = {lemma: w for lemma, _, w in result}
        assert by_lemma["alpha"] == max(by_lemma.values())
        assert by_lemma["alpha"] > by_lemma["beta"]


class TestApiContract:
    def test_extract_keywords_endpoint_contract(self, fake_model):
        from fastapi.testclient import TestClient
        from app.main import app

        client = TestClient(app)
        resp = client.post(
            "/extract_keywords",
            json={"text": "Нейронные сети обучаются на данных.", "top_n": 5},
        )
        assert resp.status_code == 200
        body = resp.json()
        assert body["extractor"] == nlp_utils.EXTRACTOR_NAME
        assert isinstance(body["keywords"], list)
        for item in body["keywords"]:
            assert set(item.keys()) == {"keyword", "surface", "weight"}
            assert 0.0 <= item["weight"] <= 1.0

    def test_health_reports_extractor(self):
        from fastapi.testclient import TestClient
        from app.main import app

        with patch("app.main.is_model_loaded", return_value=True):
            client = TestClient(app)
            resp = client.get("/health")
        assert resp.status_code == 200
        assert resp.json()["extractor"] == nlp_utils.EXTRACTOR_NAME


class TestEmbeddingModel:
    def test_embedding_model_loaded(self, embedding_model):
        """Test that embedding model is properly loaded"""
        assert embedding_model is not None
        assert hasattr(embedding_model, "encode")

    def test_embedding_generation(self, embedding_model):
        """Test embedding generation for text"""
        text = "This is a test sentence for embedding generation."
        embedding = embedding_model.encode(text)

        assert isinstance(embedding, type(embedding_model.encode("test")))
        assert len(embedding) > 0
        import numpy as np

        if hasattr(embedding, "tolist"):
            embedding_list = embedding.tolist()
        else:
            embedding_list = list(embedding)
        assert all(isinstance(x, (int, float, np.floating)) for x in embedding_list)

    def test_embedding_different_texts(self, embedding_model):
        """Test that different texts produce different embeddings"""
        emb1 = embedding_model.encode("Machine learning is great")
        emb2 = embedding_model.encode("Natural language processing is different")

        assert len(emb1) == len(emb2)
        assert any(abs(x - y) > 1e-6 for x, y in zip(emb1, emb2))


class TestModels:
    def test_extract_keywords_request_model(self):
        req = ExtractKeywordsRequest(text="Test text", top_n=5)
        assert req.text == "Test text"
        assert req.top_n == 5

        req = ExtractKeywordsRequest(text="Test text")
        assert req.top_n == 10

    def test_keyword_model(self):
        kw = Keyword(keyword="test", surface="tests", weight=0.5)
        assert kw.keyword == "test"
        assert kw.surface == "tests"
        assert kw.weight == 0.5

    def test_keyword_model_surface_default(self):
        kw = Keyword(keyword="test", weight=0.5)
        assert kw.surface == ""

    def test_extract_keywords_response_model(self):
        keywords = [
            Keyword(keyword="test1", surface="test1", weight=0.8),
            Keyword(keyword="test2", surface="tests2", weight=0.6),
        ]
        response = ExtractKeywordsResponse(
            extractor="keybert-hybrid-0.9", keywords=keywords
        )
        assert response.extractor == "keybert-hybrid-0.9"
        assert len(response.keywords) == 2

    def test_embed_request_model(self):
        req = EmbedRequest(text="Test text")
        assert req.text == "Test text"

    def test_embed_response_model(self):
        embedding = [0.1, 0.2, 0.3, 0.4]
        response = EmbedResponse(embedding=embedding)
        assert response.embedding == embedding
        assert len(response.embedding) == 4


class TestEmbeddingModelHelpers:
    def test_hf_offline_enabled(self):
        prev = os.environ.get("HF_HUB_OFFLINE")
        try:
            os.environ["HF_HUB_OFFLINE"] = "1"
            assert nlp_utils._hf_offline_enabled() is True
            os.environ["HF_HUB_OFFLINE"] = "true"
            assert nlp_utils._hf_offline_enabled() is True
            os.environ["HF_HUB_OFFLINE"] = "0"
            assert nlp_utils._hf_offline_enabled() is False
        finally:
            if prev is None:
                os.environ.pop("HF_HUB_OFFLINE", None)
            else:
                os.environ["HF_HUB_OFFLINE"] = prev

    def test_configure_hf_env(self):
        nlp_utils._configure_hf_env()
        assert os.environ.get("HF_HUB_DISABLE_TELEMETRY") == "1"
        assert os.environ.get("HF_HUB_DISABLE_SYMLINKS") == "1"

    def test_resolve_model_path(self):
        with patch("app.nlp_utils.snapshot_download") as mock_download:
            mock_download.return_value = "/fake/model/path"
            path = nlp_utils._resolve_model_path(local_only=True)
            assert path == "/fake/model/path"
            mock_download.assert_called_once()

    def test_get_embedding_model_loads_from_cache(self):
        reset_model_state()
        fake_model = MagicMock()
        with patch("app.nlp_utils.snapshot_download", return_value="/fake/path"):
            with patch("app.nlp_utils.SentenceTransformer", return_value=fake_model) as mock_st:
                model = get_embedding_model()
                assert model is fake_model
                mock_st.assert_called_once_with("/fake/path")
        assert is_model_loaded() is True

    def test_get_embedding_model_returns_cached_instance(self):
        reset_model_state()
        fake_model = MagicMock()
        with patch("app.nlp_utils.snapshot_download", return_value="/fake/path"):
            with patch("app.nlp_utils.SentenceTransformer", return_value=fake_model):
                first = get_embedding_model()
                second = get_embedding_model()
                assert first is second

    def test_get_embedding_model_network_fallback(self):
        reset_model_state()
        prev = os.environ.get("HF_HUB_OFFLINE")
        os.environ["HF_HUB_OFFLINE"] = "0"
        try:
            with patch("app.nlp_utils.snapshot_download") as mock_download:
                mock_download.side_effect = [Exception("offline fail"), "/online/path"]
                fake_model = MagicMock()
                with patch("app.nlp_utils.SentenceTransformer", return_value=fake_model) as mock_st:
                    model = get_embedding_model()
                    assert model is fake_model
                    assert mock_download.call_count == 2
                    mock_st.assert_called_once_with("/online/path")
        finally:
            if prev is None:
                os.environ.pop("HF_HUB_OFFLINE", None)
            else:
                os.environ["HF_HUB_OFFLINE"] = prev

    def test_get_embedding_model_all_attempts_fail(self):
        reset_model_state()
        with patch("app.nlp_utils.snapshot_download", side_effect=Exception("no model")):
            with patch("app.nlp_utils.SentenceTransformer") as mock_st:
                with pytest.raises(Exception, match="no model"):
                    get_embedding_model()
                mock_st.assert_not_called()
        assert ensure_model_loaded() is False

    def test_is_model_loaded(self):
        reset_model_state()
        assert is_model_loaded() is False
        nlp_utils._embedding_model = MagicMock()
        assert is_model_loaded() is True

    def test_ensure_model_loaded(self):
        reset_model_state()
        with patch("app.nlp_utils.snapshot_download", return_value="/fake/path"):
            with patch("app.nlp_utils.SentenceTransformer", return_value=MagicMock()):
                assert ensure_model_loaded() is True

    def test_ensure_model_loaded_failure(self):
        reset_model_state()
        with patch("app.nlp_utils.snapshot_download", side_effect=Exception("fail")):
            with patch("app.nlp_utils.SentenceTransformer"):
                assert ensure_model_loaded() is False


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
