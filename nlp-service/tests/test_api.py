import pytest
import sys
import os
import asyncio
from fastapi.testclient import TestClient
from unittest.mock import patch, MagicMock

# Add the parent directory to the path to import app modules
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from app.main import app
from app.models import ExtractKeywordsRequest, EmbedRequest

# Create test client
client = TestClient(app)


class TestHealthEndpoint:
    @patch("app.main.is_model_loaded", return_value=True)
    @patch("app.main.ensure_model_loaded", return_value=True)
    def test_health_endpoint(self, _mock_loaded, _mock_is_loaded):
        """Test the health check endpoint"""
        response = client.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
        assert data["model_loaded"] is True
        assert "version" in data


class TestKeywordsEndpoint:
    @patch('app.main.extract_keywords')
    def test_extract_keywords_success(self, mock_extract):
        """Test successful keyword extraction"""
        # Mock the extract_keywords function
        mock_extract.return_value = [("machine", 0.8), ("learning", 0.7)]
        
        request_data = {
            "text": "Machine learning is great",
            "top_n": 5
        }
        
        response = client.post("/extract_keywords", json=request_data)
        
        assert response.status_code == 200
        data = response.json()
        assert "keywords" in data
        assert len(data["keywords"]) == 2
        assert data["keywords"][0]["keyword"] == "machine"
        assert data["keywords"][0]["weight"] == 0.8
        
        # Verify the mock was called with correct parameters
        mock_extract.assert_called_once_with("Machine learning is great", 5)

    @patch('app.main.extract_keywords')
    def test_extract_keywords_default_top_n(self, mock_extract):
        """Test keyword extraction with default top_n"""
        mock_extract.return_value = [("test", 0.5)]
        
        request_data = {
            "text": "Test text"
        }
        
        response = client.post("/extract_keywords", json=request_data)
        
        assert response.status_code == 200
        mock_extract.assert_called_once_with("Test text", 10)

    @patch('app.main.extract_keywords')
    def test_extract_keywords_empty_result(self, mock_extract):
        """Test keyword extraction with empty result"""
        mock_extract.return_value = []
        
        request_data = {
            "text": "",
            "top_n": 5
        }
        
        response = client.post("/extract_keywords", json=request_data)
        
        assert response.status_code == 200
        data = response.json()
        assert data["keywords"] == []

    @patch('app.main.extract_keywords')
    def test_extract_keywords_error_handling(self, mock_extract):
        """Test error handling in keyword extraction"""
        mock_extract.side_effect = Exception("Test error")
        
        request_data = {
            "text": "Test text",
            "top_n": 5
        }
        
        response = client.post("/extract_keywords", json=request_data)
        
        assert response.status_code == 500
        assert "Test error" in response.json()["detail"]

    def test_extract_keywords_invalid_request(self):
        """Test keyword extraction with invalid request data"""
        # Missing required field 'text'
        request_data = {
            "top_n": 5
        }
        
        response = client.post("/extract_keywords", json=request_data)
        assert response.status_code == 422  # Validation error

    def test_extract_keywords_invalid_top_n(self):
        """Test keyword extraction with invalid top_n"""
        request_data = {
            "text": "Test text",
            "top_n": -1
        }
        
        response = client.post("/extract_keywords", json=request_data)
        # This might pass validation but should be handled
        assert response.status_code in [422, 200]


class TestEmbedEndpoint:
    @patch("app.main.get_embedding_model")
    def test_embed_success(self, mock_get_model):
        """Test successful embedding generation"""
        mock_model = MagicMock()
        mock_get_model.return_value = mock_model
        mock_embedding = [0.1, 0.2, 0.3, 0.4]
        mock_model.encode.return_value = MagicMock()
        mock_model.encode.return_value.tolist.return_value = mock_embedding

        request_data = {"text": "Test sentence"}

        response = client.post("/embed", json=request_data)

        assert response.status_code == 200
        data = response.json()
        assert "embedding" in data
        assert data["embedding"] == mock_embedding
        mock_model.encode.assert_called_once_with("Test sentence")

    @patch("app.main.get_embedding_model")
    def test_embed_empty_text(self, mock_get_model):
        """Test embedding generation with empty text"""
        mock_model = MagicMock()
        mock_get_model.return_value = mock_model
        mock_embedding = [0.0, 0.0, 0.0]
        mock_model.encode.return_value = MagicMock()
        mock_model.encode.return_value.tolist.return_value = mock_embedding

        request_data = {"text": ""}

        response = client.post("/embed", json=request_data)

        assert response.status_code == 200
        data = response.json()
        assert data["embedding"] == mock_embedding

    @patch("app.main.get_embedding_model")
    def test_embed_error_handling(self, mock_get_model):
        """Test error handling in embedding generation"""
        mock_model = MagicMock()
        mock_get_model.return_value = mock_model
        mock_model.encode.side_effect = Exception("Model error")

        request_data = {"text": "Test text"}

        response = client.post("/embed", json=request_data)

        assert response.status_code == 500
        assert "Model error" in response.json()["detail"]

    def test_embed_invalid_request(self):
        """Test embedding generation with invalid request data"""
        # Missing required field 'text'
        request_data = {}
        
        response = client.post("/embed", json=request_data)
        assert response.status_code == 422  # Validation error


class TestSimilarityEndpoint:
    @patch("app.nlp_utils.get_embedding_model")
    def test_similarity_success(self, mock_get_model):
        """Test successful similarity computation"""
        mock_model = MagicMock()
        mock_get_model.return_value = mock_model
        # Orthogonal unit vectors -> cosine similarity 0 -> mapped to 0.5
        mock_model.encode.return_value = [
            [1.0, 0.0, 0.0],
            [0.0, 1.0, 0.0],
        ]

        request_data = {
            "text_a": "First sentence",
            "text_b": "Second sentence",
        }

        response = client.post("/similarity", json=request_data)

        assert response.status_code == 200
        data = response.json()
        assert "similarity" in data
        assert data["similarity"] == pytest.approx(0.5, abs=0.001)

    @patch("app.nlp_utils.get_embedding_model")
    def test_similarity_empty_text(self, mock_get_model):
        """Test similarity with empty text"""
        mock_model = MagicMock()
        mock_get_model.return_value = mock_model

        request_data = {
            "text_a": "",
            "text_b": "Second sentence",
        }

        response = client.post("/similarity", json=request_data)

        assert response.status_code == 200
        data = response.json()
        assert data["similarity"] == 0.0

    @patch("app.nlp_utils.get_embedding_model")
    def test_similarity_error_handling(self, mock_get_model):
        """Test error handling in similarity computation"""
        mock_model = MagicMock()
        mock_get_model.return_value = mock_model
        mock_model.encode.side_effect = Exception("Model error")

        request_data = {
            "text_a": "First sentence",
            "text_b": "Second sentence",
        }

        response = client.post("/similarity", json=request_data)

        assert response.status_code == 500
        assert "Model error" in response.json()["detail"]

    def test_similarity_invalid_request(self):
        """Test similarity with invalid request data"""
        response = client.post("/similarity", json={})
        assert response.status_code == 422


class TestAPIIntegration:
    """Integration tests for the API endpoints"""
    
    @patch("app.main.is_model_loaded", return_value=True)
    @patch("app.main.ensure_model_loaded", return_value=True)
    def test_api_structure(self, _mock_loaded, _mock_is_loaded):
        """Test that API has correct structure"""
        # Test that endpoints exist and return correct status codes
        health_response = client.get("/health")
        assert health_response.status_code == 200
        
        # Test that endpoints return validation errors for invalid input
        keywords_response = client.post("/extract_keywords", json={})
        assert keywords_response.status_code == 422
        
        embed_response = client.post("/embed", json={})
        assert embed_response.status_code == 422

        similarity_response = client.post("/similarity", json={})
        assert similarity_response.status_code == 422

    def test_cors_headers(self):
        """Test CORS headers if configured"""
        response = client.options("/health")
        # This test depends on CORS configuration
        # Adjust based on your actual CORS setup
        assert response.status_code in [200, 405]  # 405 if OPTIONS not allowed


class TestErrorHandling:
    """Test various error scenarios"""
    
    def test_nonexistent_endpoint(self):
        """Test request to nonexistent endpoint"""
        response = client.get("/nonexistent")
        assert response.status_code == 404

    def test_invalid_method(self):
        """Test invalid HTTP method"""
        response = client.put("/health")
        assert response.status_code == 405

    def test_invalid_content_type(self):
        """Test request with invalid content type"""
        response = client.post(
            "/extract_keywords",
            data="not json",
            headers={"content-type": "text/plain"}
        )
        assert response.status_code == 422


class TestLifespanAndHealth:
    @patch("app.main.ensure_model_loaded", return_value=True)
    def test_lifespan_success(self, _mock_ensure):
        """Test that lifespan context manager completes successfully"""
        from app.main import lifespan
        loop = asyncio.new_event_loop()
        try:
            ctx = lifespan(app)
            loop.run_until_complete(ctx.__aenter__())
            loop.run_until_complete(ctx.__aexit__(None, None, None))
        finally:
            loop.close()

    @patch("app.main.ensure_model_loaded", return_value=False)
    def test_lifespan_failure_raises(self, _mock_ensure):
        """Test that lifespan raises RuntimeError when model fails to load"""
        from app.main import lifespan
        loop = asyncio.new_event_loop()
        try:
            ctx = lifespan(app)
            with pytest.raises(RuntimeError):
                loop.run_until_complete(ctx.__aenter__())
        finally:
            loop.close()

    @patch("app.main.is_model_loaded", return_value=False)
    def test_health_not_loaded(self, _mock_loaded):
        """Test health endpoint returns 503 when model is not loaded"""
        response = client.get("/health")
        assert response.status_code == 503
        assert "not loaded" in response.json()["detail"]


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
