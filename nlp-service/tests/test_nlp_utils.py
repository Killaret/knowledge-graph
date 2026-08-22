import pytest
import sys
import os
from unittest.mock import patch, MagicMock

# Add the parent directory to the path to import app modules
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import app.nlp_utils as nlp_utils
from app.nlp_utils import extract_keywords, get_embedding_model, is_model_loaded, ensure_model_loaded
from app.models import ExtractKeywordsRequest, ExtractKeywordsResponse, Keyword, EmbedRequest, EmbedResponse


def reset_model_state():
    """Reset global model state for isolated tests."""
    nlp_utils._embedding_model = None
    nlp_utils._model_load_error = None


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


class TestKeywordExtraction:
    def test_extract_keywords_basic(self):
        """Test basic keyword extraction"""
        text = "Machine learning is a subset of artificial intelligence that focuses on neural networks."
        keywords = extract_keywords(text, top_n=5)
        
        assert isinstance(keywords, list)
        assert len(keywords) <= 5
        assert len(keywords) > 0
        
        for kw, weight in keywords:
            assert isinstance(kw, str)
            assert isinstance(weight, float)
            assert 0.0 <= weight <= 1.0
            assert len(kw.strip()) > 0

    def test_extract_keywords_empty_text(self):
        """Test keyword extraction with empty text"""
        keywords = extract_keywords("", top_n=5)
        assert keywords == []
        
        keywords = extract_keywords("   ", top_n=5)
        assert keywords == []

    def test_extract_keywords_none_text(self):
        """Test keyword extraction with None text"""
        keywords = extract_keywords(None, top_n=5)
        assert keywords == []

    def test_extract_keywords_russian_text(self):
        """Test keyword extraction with Russian text"""
        text = "Machine learning - this is subset of artificial intelligence with neural networks"
        keywords = extract_keywords(text, top_n=3)
        
        assert isinstance(keywords, list)
        assert len(keywords) <= 3
        
        for kw, weight in keywords:
            assert isinstance(kw, str)
            assert isinstance(weight, float)

    def test_extract_keywords_top_n_parameter(self):
        """Test that top_n parameter limits results"""
        text = "This is a long text with many different words and concepts to extract from the sentence."
        keywords_3 = extract_keywords(text, top_n=3)
        keywords_10 = extract_keywords(text, top_n=10)
        
        assert len(keywords_3) <= 3
        assert len(keywords_10) <= 10
        assert len(keywords_3) <= len(keywords_10)

    def test_extract_keywords_weight_calculation(self):
        """Test that weights are properly calculated"""
        text = "Machine learning artificial intelligence neural networks"
        keywords = extract_keywords(text, top_n=5)
        
        for kw, weight in keywords:
            assert 0.0 <= weight <= 1.0
            # Higher weight should indicate better keyword
            assert isinstance(weight, float)


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
        # Handle numpy array conversion
        import numpy as np
        if hasattr(embedding, 'tolist'):
            embedding_list = embedding.tolist()
        else:
            embedding_list = list(embedding)
        assert all(isinstance(x, (int, float, np.floating)) for x in embedding_list)

    def test_embedding_empty_text(self, embedding_model):
        """Test embedding generation with empty text"""
        embedding = embedding_model.encode("")
        assert len(embedding) > 0

    def test_embedding_different_texts(self, embedding_model):
        """Test that different texts produce different embeddings"""
        text1 = "Machine learning is great"
        text2 = "Natural language processing is different"

        emb1 = embedding_model.encode(text1)
        emb2 = embedding_model.encode(text2)
        
        assert len(emb1) == len(emb2)
        # Embeddings should be different (not exactly equal)
        assert any(abs(x - y) > 1e-6 for x, y in zip(emb1, emb2))


class TestModels:
    def test_extract_keywords_request_model(self):
        """Test ExtractKeywordsRequest model validation"""
        # Valid request
        req = ExtractKeywordsRequest(text="Test text", top_n=5)
        assert req.text == "Test text"
        assert req.top_n == 5
        
        # Default top_n
        req = ExtractKeywordsRequest(text="Test text")
        assert req.text == "Test text"
        assert req.top_n == 10

    def test_keyword_model(self):
        """Test Keyword model validation"""
        kw = Keyword(keyword="test", weight=0.5)
        assert kw.keyword == "test"
        assert kw.weight == 0.5

    def test_extract_keywords_response_model(self):
        """Test ExtractKeywordsResponse model validation"""
        keywords = [Keyword(keyword="test1", weight=0.8), Keyword(keyword="test2", weight=0.6)]
        response = ExtractKeywordsResponse(keywords=keywords)
        
        assert len(response.keywords) == 2
        assert response.keywords[0].keyword == "test1"
        assert response.keywords[0].weight == 0.8

    def test_embed_request_model(self):
        """Test EmbedRequest model validation"""
        req = EmbedRequest(text="Test text")
        assert req.text == "Test text"

    def test_embed_response_model(self):
        """Test EmbedResponse model validation"""
        embedding = [0.1, 0.2, 0.3, 0.4]
        response = EmbedResponse(embedding=embedding)
        
        assert response.embedding == embedding
        assert len(response.embedding) == 4


class TestIntegration:
    def test_keyword_extraction_integration(self):
        """Integration test for keyword extraction workflow"""
        text = "Machine learning and artificial intelligence are transforming technology."
        keywords = extract_keywords(text, top_n=3)
        
        # Create response model
        keyword_objects = [Keyword(keyword=kw, weight=w) for kw, w in keywords]
        response = ExtractKeywordsResponse(keywords=keyword_objects)
        
        assert isinstance(response, ExtractKeywordsResponse)
        assert len(response.keywords) <= 3

    def test_embedding_integration(self, embedding_model):
        """Integration test for embedding workflow"""
        text = "Test sentence for embedding."
        embedding_vector = embedding_model.encode(text).tolist()
        
        response = EmbedResponse(embedding=embedding_vector)
        
        assert isinstance(response, EmbedResponse)
        assert len(response.embedding) > 0
        assert all(isinstance(x, float) for x in response.embedding)


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
