# knowledge-graph-python-nlp

**Version:** 1.0  
**Purpose:** Python NLP development - FastAPI, ML models, text processing  
**Status:** Active  
**Priority:** 🟢 High

---

## Overview

`knowledge-graph-python-nlp` specializes in Python NLP development for the Knowledge Graph project.

**Key Areas:**
- FastAPI web framework
- ML models (sentence-transformers)
- Keyword extraction (YAKE)
- Text preprocessing (NLTK)
- Pydantic data validation
- API testing
- Docker containerization

---

## Development Patterns

### 1. FastAPI Development

#### Endpoint Structure
```python
from fastapi import FastAPI, HTTPException
from .models import ExtractKeywordsRequest, ExtractKeywordsResponse
from .nlp_utils import extract_keywords
import logging

logger = logging.getLogger(__name__)
app = FastAPI(title="NLP Service for Knowledge Graph")

@app.post("/extract_keywords", response_model=ExtractKeywordsResponse)
async def extract_keywords_endpoint(req: ExtractKeywordsRequest):
    try:
        keywords = extract_keywords(req.text, req.top_n)
        return ExtractKeywordsResponse(
            keywords=[Keyword(keyword=kw, weight=w) for kw, w in keywords]
        )
    except Exception as e:
        logger.exception("Error extracting keywords")
        raise HTTPException(status_code=500, detail=str(e))
```

#### Health Check Endpoint
```python
@app.get("/health")
async def health():
    """Health check endpoint with model verification"""
    try:
        if embedding_model is None:
            raise HTTPException(status_code=503, detail="Embedding model not loaded")
        return {
            "status": "healthy",
            "model_loaded": True,
            "version": "1.0.0"
        }
    except Exception as e:
        logger.error(f"Health check failed: {e}")
        raise HTTPException(status_code=503, detail=f"Service unhealthy: {str(e)}")
```

### 2. Pydantic Models

#### Request Models
```python
from pydantic import BaseModel, Field
from typing import List

class ExtractKeywordsRequest(BaseModel):
    text: str = Field(..., max_length=10000, description="Text to extract keywords from")
    top_n: int = Field(default=10, ge=1, le=50, description="Number of keywords to return (1-50)")

class EmbedRequest(BaseModel):
    text: str = Field(..., max_length=10000, description="Text to generate embedding for")
```

#### Response Models
```python
class Keyword(BaseModel):
    keyword: str
    weight: float

class ExtractKeywordsResponse(BaseModel):
    keywords: List[Keyword]

class EmbedResponse(BaseModel):
    embedding: List[float]
```

### 3. NLP Utilities

#### Embedding Model
```python
from sentence_transformers import SentenceTransformer
import logging

logger = logging.getLogger(__name__)

# Load embedding model (CPU)
embedding_model = SentenceTransformer('all-MiniLM-L6-V2')
logger.info("Embedding model loaded")

def generate_embedding(text: str) -> List[float]:
    """Generate embedding for text using sentence-transformers"""
    if not text or not text.strip():
        return [0.0] * embedding_model.get_sentence_embedding_dimension()
    embedding = embedding_model.encode(text).tolist()
    return embedding
```

#### Keyword Extraction
```python
import nltk
import yake

# Download NLTK data
try:
    nltk.data.find('tokenizers/punkt')
except LookupError:
    nltk.download('punkt')
try:
    nltk.data.find('corpora/stopwords')
except LookupError:
    nltk.download('stopwords')

# Load stopwords for Russian and English
stop_words = set(nltk.corpus.stopwords.words('russian') + 
                 nltk.corpus.stopwords.words('english'))

# Initialize YAKE extractor
kw_extractor = yake.KeywordExtractor(lan="ru", top=20, stopwords=stop_words)

def extract_keywords(text: str, top_n: int = 10) -> List[tuple]:
    """Extract keywords using YAKE algorithm"""
    if not text or not text.strip():
        return []
    keywords = kw_extractor.extract_keywords(text)
    result = []
    for kw, score in keywords[:top_n]:
        # Convert YAKE score to weight (score in [0..1], lower is better)
        weight = max(0.0, min(1.0, 1.0 - score))
        result.append((kw, weight))
    return result
```

#### Text Preprocessing
```python
import re
from typing import List

def preprocess_text(text: str) -> str:
    """Basic text preprocessing"""
    if not text:
        return ""
    
    # Convert to lowercase
    text = text.lower()
    
    # Remove special characters but keep alphanumeric and spaces
    text = re.sub(r'[^a-zA-Zа-яА-Я0-9\s]', '', text)
    
    # Remove extra whitespace
    text = ' '.join(text.split())
    
    return text

def tokenize_text(text: str) -> List[str]:
    """Tokenize text using NLTK"""
    from nltk.tokenize import word_tokenize
    tokens = word_tokenize(text)
    return [token for token in tokens if token.isalnum()]
```

---

## Testing

### Unit Tests for NLP Utilities
```python
import pytest
from app.nlp_utils import extract_keywords, embedding_model

class TestKeywordExtraction:
    def test_extract_keywords_basic(self):
        """Test basic keyword extraction"""
        text = "Machine learning is a subset of artificial intelligence"
        keywords = extract_keywords(text, top_n=5)
        
        assert isinstance(keywords, list)
        assert len(keywords) <= 5
        assert len(keywords) > 0
        
        for kw, weight in keywords:
            assert isinstance(kw, str)
            assert isinstance(weight, float)
            assert 0.0 <= weight <= 1.0

    def test_extract_keywords_empty_text(self):
        """Test keyword extraction with empty text"""
        keywords = extract_keywords("", top_n=5)
        assert keywords == []
```

### API Tests
```python
from fastapi.testclient import TestClient
from unittest.mock import patch
from app.main import app

client = TestClient(app)

class TestKeywordsEndpoint:
    @patch('app.main.extract_keywords')
    def test_extract_keywords_success(self, mock_extract):
        """Test successful keyword extraction"""
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
```

### Running Tests
```bash
# Unit tests
cd nlp-service
pytest tests/test_nlp_utils.py -v

# API tests
pytest tests/test_api.py -v

# All tests with coverage
pytest tests/ -v --cov=app --cov-report=html
```

---

## Commands

### Development
```bash
# Install dependencies
cd nlp-service
pip install -r requirements.txt

# Run development server
uvicorn app.main:app --reload --host 0.0.0.0 --port 8000

# Run with specific port
uvicorn app.main:app --reload --port 8001
```

### Testing
```bash
# Run all tests
pytest tests/ -v

# Run specific test file
pytest tests/test_nlp_utils.py -v

# Run with coverage
pytest tests/ --cov=app --cov-report=html

# Run specific test
pytest tests/test_nlp_utils.py::TestKeywordExtraction::test_extract_keywords_basic -v
```

### Docker
```bash
# Build image
docker build -t knowledge-graph-nlp-service .

# Run container
docker run -p 8000:8000 knowledge-graph-nlp-service

# Run with environment variables
docker run -p 8000:8000 -e MODEL_NAME=all-MiniLM-L6-V2 knowledge-graph-nlp-service
```

### Linting
```bash
# Run black formatter
black app/ tests/

# Run flake8 linter
flake8 app/ tests/

# Run mypy type checker
mypy app/
```

---

## Best Practices

### Error Handling
```python
from fastapi import HTTPException
import logging

logger = logging.getLogger(__name__)

@app.post("/process")
async def process_text(req: ProcessRequest):
    try:
        result = process_text_internal(req.text)
        return ProcessResponse(result=result)
    except ValueError as e:
        logger.warning(f"Validation error: {e}")
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        logger.exception("Unexpected error in process_text")
        raise HTTPException(status_code=500, detail="Internal server error")
```

### Logging
```python
import logging

logger = logging.getLogger(__name__)

# Log at appropriate levels
logger.debug("Detailed debug information")
logger.info("General information about execution")
logger.warning("Warning for potential issues")
logger.error("Error occurred", exc_info=True)
logger.critical("Critical error")
```

### Model Loading
```python
# Lazy loading to reduce startup time
_embedding_model = None

def get_embedding_model():
    global _embedding_model
    if _embedding_model is None:
        logger.info("Loading embedding model...")
        _embedding_model = SentenceTransformer('all-MiniLM-L6-V2')
        logger.info("Embedding model loaded")
    return _embedding_model
```

### Resource Management
```python
# Add request size limits
from pydantic import Field

class ProcessRequest(BaseModel):
    text: str = Field(..., max_length=10000, description="Text to process")
    
# Add timeout for long-running operations
from asyncio import TimeoutError

@app.post("/process")
async def process_text(req: ProcessRequest):
    try:
        result = await asyncio.wait_for(
            process_text_async(req.text), 
            timeout=30.0
        )
        return ProcessResponse(result=result)
    except TimeoutError:
        raise HTTPException(status_code=504, detail="Processing timeout")
```

---

## Performance Considerations

### Model Caching
```python
from functools import lru_cache

@lru_cache(maxsize=1000)
def get_cached_embedding(text: str) -> List[float]:
    """Cache embeddings to avoid recomputation"""
    return generate_embedding(text)
```

### Batch Processing
```python
def generate_embeddings_batch(texts: List[str]) -> List[List[float]]:
    """Generate embeddings for multiple texts efficiently"""
    embeddings = embedding_model.encode(texts)
    return [emb.tolist() for emb in embeddings]
```

### Memory Management
```python
# Process large texts in chunks
def process_large_text(text: str, chunk_size: int = 1000) -> List[dict]:
    """Process large text in chunks to manage memory"""
    chunks = [text[i:i+chunk_size] for i in range(0, len(text), chunk_size)]
    results = []
    for chunk in chunks:
        result = process_chunk(chunk)
        results.append(result)
    return results
```

---

## API Documentation

### OpenAPI Specification
- Location: `nlp-service/openapi.yaml`
- Auto-generated by FastAPI
- Access at: `http://localhost:8000/docs` (Swagger UI)
- Alternative: `http://localhost:8000/redoc` (ReDoc)

### Endpoints
- `GET /health` - Health check
- `POST /extract_keywords` - Extract keywords from text
- `POST /embed` - Generate text embedding

---

**Tools:** `python-nlp-tools.md`  
**Coverage Target:** > 70%  
**Response Time Target:** p95 < 2s (for embeddings)