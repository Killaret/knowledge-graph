import pytest
import sys
import os

# Add the parent directory to the path to import app modules
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from app.nlp_utils import extract_keywords, embedding_model
from app.models import ExtractKeywordsRequest, ExtractKeywordsResponse, Keyword, EmbedRequest, EmbedResponse


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
    def test_embedding_model_loaded(self):
        """Test that embedding model is properly loaded"""
        assert embedding_model is not None
        assert hasattr(embedding_model, 'encode')

    def test_embedding_generation(self):
        """Test embedding generation for text"""
        text = "This is a test sentence for embedding generation."
        embedding = embedding_model.encode(text)
        
        assert isinstance(embedding, type(embedding_model.encode("test")))
        assert len(embedding) > 0
        assert all(isinstance(x, (int, float)) for x in embedding)

    def test_embedding_empty_text(self):
        """Test embedding generation with empty text"""
        embedding = embedding_model.encode("")
        assert len(embedding) > 0

    def test_embedding_different_texts(self):
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

    def test_embedding_integration(self):
        """Integration test for embedding workflow"""
        text = "Test sentence for embedding."
        embedding_vector = embedding_model.encode(text).tolist()
        
        response = EmbedResponse(embedding=embedding_vector)
        
        assert isinstance(response, EmbedResponse)
        assert len(response.embedding) > 0
        assert all(isinstance(x, float) for x in response.embedding)


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
