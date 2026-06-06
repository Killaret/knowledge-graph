# Инструменты Python NLP Агента

## 🎯 Основные задачи

1. FastAPI endpoints разработка
2. ML модели интеграция (sentence-transformers)
3. Keyword extraction (YAKE)
4. Text preprocessing (NLTK)
5. Pydantic валидация данных
6. Тестирование (pytest)
7. Docker контейнеризация

---

## 🛠️ Разработка

### 1. FastAPI Development

#### Базовое приложение
```python
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="NLP Service for Knowledge Graph",
    description="NLP processing service for text analysis",
    version="1.0.0"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/health")
async def health():
    """Health check endpoint"""
    return {"status": "healthy", "version": "1.0.0"}
```

#### REST API Patterns
```python
from pydantic import BaseModel, Field
from typing import List, Optional

# Request Models
class ExtractKeywordsRequest(BaseModel):
    text: str = Field(..., max_length=10000, description="Text to extract keywords from")
    top_n: int = Field(default=10, ge=1, le=50, description="Number of keywords to return")
    language: Optional[str] = Field(default="ru", description="Language code")

# Response Models
class Keyword(BaseModel):
    keyword: str
    weight: float

class ExtractKeywordsResponse(BaseModel):
    keywords: List[Keyword]
    processing_time: float

# Endpoint
@app.post("/extract_keywords", response_model=ExtractKeywordsResponse)
async def extract_keywords_endpoint(req: ExtractKeywordsRequest):
    import time
    start_time = time.time()
    
    try:
        keywords = extract_keywords(req.text, req.top_n, req.language)
        processing_time = time.time() - start_time
        
        return ExtractKeywordsResponse(
            keywords=[Keyword(keyword=kw, weight=w) for kw, w in keywords],
            processing_time=processing_time
        )
    except Exception as e:
        logger.exception("Error extracting keywords")
        raise HTTPException(status_code=500, detail=str(e))
```

#### Dependency Injection
```python
from fastapi import Depends

class ModelService:
    def __init__(self):
        self.model = None
    
    def load_model(self):
        if self.model is None:
            self.model = SentenceTransformer('all-MiniLM-L6-V2')
        return self.model

model_service = ModelService()

def get_model_service():
    return model_service

@app.post("/embed")
async def embed_endpoint(
    req: EmbedRequest,
    service: ModelService = Depends(get_model_service)
):
    model = service.load_model()
    embedding = model.encode(req.text).tolist()
    return EmbedResponse(embedding=embedding)
```

### 2. ML Models Integration

#### Sentence-Transformers
```python
from sentence_transformers import SentenceTransformer
import logging

logger = logging.getLogger(__name__)

# Model loading
class EmbeddingModel:
    def __init__(self, model_name: str = 'all-MiniLM-L6-V2'):
        self.model_name = model_name
        self.model = None
        self._load_model()
    
    def _load_model(self):
        logger.info(f"Loading embedding model: {self.model_name}")
        self.model = SentenceTransformer(self.model_name)
        logger.info("Embedding model loaded successfully")
    
    def encode(self, text: str) -> List[float]:
        """Encode text to embedding vector"""
        if not text or not text.strip():
            dimension = self.model.get_sentence_embedding_dimension()
            return [0.0] * dimension
        
        embedding = self.model.encode(text)
        return embedding.tolist()
    
    def encode_batch(self, texts: List[str]) -> List[List[float]]:
        """Encode multiple texts efficiently"""
        embeddings = self.model.encode(texts)
        return [emb.tolist() for emb in embeddings]

# Usage
embedding_model = EmbeddingModel()
```

#### Model Configuration
```python
import os
from typing import Optional

class ModelConfig:
    def __init__(self):
        self.model_name = os.getenv("EMBEDDING_MODEL", "all-MiniLM-L6-V2")
        self.device = os.getenv("MODEL_DEVICE", "cpu")
        self.batch_size = int(os.getenv("MODEL_BATCH_SIZE", "32"))
        self.max_length = int(os.getenv("MODEL_MAX_LENGTH", "512"))
    
    def get_model_kwargs(self):
        return {
            "device": self.device,
            "batch_size": self.batch_size
        }

config = ModelConfig()
```

### 3. Keyword Extraction

#### YAKE Integration
```python
import yake
import nltk
from typing import List, Tuple

class KeywordExtractor:
    def __init__(self, language: str = "ru"):
        self.language = language
        self._setup_nltk()
        self.stopwords = self._load_stopwords()
        self.extractor = yake.KeywordExtractor(
            lan=language,
            n=3,  # max n-gram size
            top=20,  # top keywords to extract
            stopwords=self.stopwords
        )
    
    def _setup_nltk(self):
        """Download required NLTK data"""
        try:
            nltk.data.find('tokenizers/punkt')
        except LookupError:
            nltk.download('punkt')
        
        try:
            nltk.data.find('corpora/stopwords')
        except LookupError:
            nltk.download('stopwords')
    
    def _load_stopwords(self):
        """Load stopwords for supported languages"""
        stopwords = set()
        
        # Russian stopwords
        if self.language == "ru":
            stopwords.update(nltk.corpus.stopwords.words('russian'))
        
        # English stopwords
        stopwords.update(nltk.corpus.stopwords.words('english'))
        
        return stopwords
    
    def extract(self, text: str, top_n: int = 10) -> List[Tuple[str, float]]:
        """Extract keywords with weights"""
        if not text or not text.strip():
            return []
        
        keywords = self.extractor.extract_keywords(text)
        
        # Convert YAKE scores to weights (lower score = better keyword)
        result = []
        for kw, score in keywords[:top_n]:
            weight = max(0.0, min(1.0, 1.0 - score))
            result.append((kw, weight))
        
        return result

# Usage
keyword_extractor = KeywordExtractor(language="ru")
```

#### Alternative Methods
```python
from sklearn.feature_extraction.text import TfidfVectorizer
import numpy as np

class TfidfKeywordExtractor:
    def __init__(self, max_features: int = 1000):
        self.vectorizer = TfidfVectorizer(
            max_features=max_features,
            stop_words='english'
        )
    
    def extract(self, text: str, top_n: int = 10) -> List[Tuple[str, float]]:
        """Extract keywords using TF-IDF"""
        # Fit and transform
        tfidf_matrix = self.vectorizer.fit_transform([text])
        feature_names = self.vectorizer.get_feature_names_out()
        
        # Get TF-IDF scores
        tfidf_scores = tfidf_matrix.toarray()[0]
        
        # Get top keywords
        top_indices = np.argsort(tfidf_scores)[-top_n:][::-1]
        
        keywords = []
        for idx in top_indices:
            keywords.append((feature_names[idx], float(tfidf_scores[idx])))
        
        return keywords
```

### 4. Text Preprocessing

#### NLTK Preprocessing
```python
import nltk
import re
from typing import List

class TextPreprocessor:
    def __init__(self, language: str = "ru"):
        self.language = language
        self._setup_nltk()
    
    def _setup_nltk(self):
        """Download required NLTK data"""
        required_packages = ['punkt', 'stopwords', 'wordnet']
        
        for package in required_packages:
            try:
                nltk.data.find(f'tokenizers/{package}' if package == 'punkt' else f'corpora/{package}')
            except LookupError:
                nltk.download(package)
    
    def clean_text(self, text: str) -> str:
        """Basic text cleaning"""
        if not text:
            return ""
        
        # Convert to lowercase
        text = text.lower()
        
        # Remove special characters but keep alphanumeric and spaces
        text = re.sub(r'[^a-zA-Zа-яА-Я0-9\s]', '', text)
        
        # Remove extra whitespace
        text = ' '.join(text.split())
        
        return text
    
    def tokenize(self, text: str) -> List[str]:
        """Tokenize text"""
        from nltk.tokenize import word_tokenize
        tokens = word_tokenize(text)
        return [token for token in tokens if token.isalnum()]
    
    def remove_stopwords(self, tokens: List[str]) -> List[str]:
        """Remove stopwords from tokens"""
        stop_words = set(nltk.corpus.stopwords.words('english'))
        
        if self.language == "ru":
            stop_words.update(nltk.corpus.stopwords.words('russian'))
        
        return [token for token in tokens if token.lower() not in stop_words]
    
    def preprocess_pipeline(self, text: str) -> List[str]:
        """Full preprocessing pipeline"""
        text = self.clean_text(text)
        tokens = self.tokenize(text)
        tokens = self.remove_stopwords(tokens)
        return tokens
```

### 5. Pydantic Data Validation

#### Advanced Validation
```python
from pydantic import BaseModel, Field, validator
from typing import List, Optional
import re

class TextProcessingRequest(BaseModel):
    text: str = Field(..., min_length=1, max_length=10000)
    language: str = Field(default="ru", regex="^(ru|en)$")
    options: Optional[dict] = None
    
    @validator('text')
    def validate_text(cls, v):
        """Validate text content"""
        if not v or not v.strip():
            raise ValueError("Text cannot be empty")
        
        # Check for excessive special characters
        special_char_ratio = len(re.findall(r'[^a-zA-Zа-яА-Я0-9\s]', v)) / len(v)
        if special_char_ratio > 0.5:
            raise ValueError("Text contains too many special characters")
        
        return v.strip()
    
    @validator('options')
    def validate_options(cls, v):
        """Validate options dictionary"""
        if v is None:
            return {}
        
        valid_keys = {'top_n', 'min_length', 'max_length'}
        invalid_keys = set(v.keys()) - valid_keys
        
        if invalid_keys:
            raise ValueError(f"Invalid options: {invalid_keys}")
        
        return v
```

#### Custom Validators
```python
from pydantic import BaseModel, validator

class EmbeddingRequest(BaseModel):
    texts: List[str] = Field(..., min_items=1, max_items=100)
    
    @validator('texts')
    def validate_texts(cls, v):
        """Validate all texts in the list"""
        validated_texts = []
        
        for text in v:
            if not text or not text.strip():
                continue  # Skip empty texts
            
            if len(text) > 10000:
                raise ValueError(f"Text too long: {len(text)} characters")
            
            validated_texts.append(text.strip())
        
        if not validated_texts:
            raise ValueError("At least one valid text is required")
        
        return validated_texts
```

---

## 🧪 Тестирование

### Unit Tests
```python
import pytest
from app.nlp_utils import extract_keywords, generate_embedding

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
        """Test with empty text"""
        keywords = extract_keywords("", top_n=5)
        assert keywords == []
    
    def test_extract_keywords_russian_text(self):
        """Test with Russian text"""
        text = "Машинное обучение это подраздел искусственного интеллекта"
        keywords = extract_keywords(text, top_n=3)
        assert isinstance(keywords, list)

class TestEmbeddingGeneration:
    def test_generate_embedding(self):
        """Test embedding generation"""
        text = "Test sentence for embedding"
        embedding = generate_embedding(text)
        
        assert isinstance(embedding, list)
        assert len(embedding) > 0
        assert all(isinstance(x, float) for x in embedding)
    
    def test_generate_embedding_empty_text(self):
        """Test with empty text"""
        embedding = generate_embedding("")
        assert isinstance(embedding, list)
        assert len(embedding) > 0  # Should return zero vector
```

### API Integration Tests
```python
from fastapi.testclient import TestClient
from unittest.mock import patch, MagicMock
from app.main import app

client = TestClient(app)

class TestAPIEndpoints:
    def test_health_endpoint(self):
        """Test health check"""
        response = client.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
    
    @patch('app.main.extract_keywords')
    def test_extract_keywords_endpoint(self, mock_extract):
        """Test keyword extraction endpoint"""
        mock_extract.return_value = [("machine", 0.8), ("learning", 0.7)]
        
        response = client.post("/extract_keywords", json={
            "text": "Machine learning is great",
            "top_n": 5
        })
        
        assert response.status_code == 200
        data = response.json()
        assert "keywords" in data
        assert len(data["keywords"]) == 2
    
    def test_extract_keywords_validation_error(self):
        """Test validation error handling"""
        response = client.post("/extract_keywords", json={
            "top_n": 5  # Missing required 'text' field
        })
        
        assert response.status_code == 422  # Validation error
```

### Test Configuration
```python
# conftest.py
import pytest
from app.main import app
from fastapi.testclient import TestClient

@pytest.fixture
def client():
    """Create test client"""
    return TestClient(app)

@pytest.fixture
def sample_text():
    """Sample text for testing"""
    return "Machine learning is a subset of artificial intelligence that focuses on neural networks."

@pytest.fixture
def mock_embedding_model():
    """Mock embedding model"""
    with patch('app.nlp_utils.embedding_model') as mock:
        mock.encode.return_value.tolist.return_value = [0.1, 0.2, 0.3, 0.4]
        yield mock
```

### Running Tests
```bash
# Run all tests
pytest tests/ -v

# Run with coverage
pytest tests/ --cov=app --cov-report=html

# Run specific test file
pytest tests/test_nlp_utils.py -v

# Run specific test
pytest tests/test_nlp_utils.py::TestKeywordExtraction::test_extract_keywords_basic -v

# Run with verbose output
pytest tests/ -vv -s

# Run and stop on first failure
pytest tests/ -x
```

---

## 🐳 Docker

### Dockerfile
```dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements first for better caching
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY app/ ./app/

# Create non-root user
RUN useradd -m -u 1000 appuser && chown -R appuser:appuser /app
USER appuser

# Expose port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import requests; requests.get('http://localhost:8000/health')"

# Run application
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

### Docker Compose
```yaml
version: '3.8'

services:
  nlp-service:
    build: ./nlp-service
    ports:
      - "8000:8000"
    environment:
      - EMBEDDING_MODEL=all-MiniLM-L6-V2
      - LOG_LEVEL=INFO
    volumes:
      - ./nlp-service:/app
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### Docker Commands
```bash
# Build image
docker build -t knowledge-graph-nlp-service ./nlp-service

# Run container
docker run -p 8000:8000 knowledge-graph-nlp-service

# Run with environment variables
docker run -p 8000:8000 \
  -e EMBEDDING_MODEL=all-MiniLM-L6-V2 \
  -e LOG_LEVEL=DEBUG \
  knowledge-graph-nlp-service

# Run with volume mount for development
docker run -p 8000:8000 \
  -v $(pwd)/nlp-service:/app \
  knowledge-graph-nlp-service

# View logs
docker logs -f nlp-service

# Execute command in container
docker exec -it nlp-service python -c "import app; print(app.__version__)"
```

---

## 🔧 Команды разработки

### Установка зависимостей
```bash
# Create virtual environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Install development dependencies
pip install pytest pytest-cov black flake8 mypy
```

### Запуск сервера
```bash
# Development server with hot reload
uvicorn app.main:app --reload --host 0.0.0.0 --port 8000

# Production server
uvicorn app.main:app --host 0.0.0.0 --port 8000 --workers 4

# With specific configuration
uvicorn app.main:app \
  --host 0.0.0.0 \
  --port 8000 \
  --workers 4 \
  --log-level info \
  --access-log
```

### Linting и форматирование
```bash
# Format code with black
black app/ tests/

# Check formatting
black --check app/ tests/

# Lint with flake8
flake8 app/ tests/

# Type checking with mypy
mypy app/

# Run all checks
black app/ tests/ && flake8 app/ tests/ && mypy app/
```

---

## 📊 Мониторинг и логирование

### Структурированное логирование
```python
import logging
import json
from datetime import datetime

class JSONFormatter(logging.Formatter):
    """JSON formatter for structured logging"""
    
    def format(self, record):
        log_data = {
            "timestamp": datetime.utcnow().isoformat(),
            "level": record.levelname,
            "message": record.getMessage(),
            "module": record.module,
            "function": record.funcName,
            "line": record.lineno
        }
        
        if hasattr(record, 'extra_data'):
            log_data.update(record.extra_data)
        
        return json.dumps(log_data)

# Configure logging
logger = logging.getLogger(__name__)
handler = logging.StreamHandler()
handler.setFormatter(JSONFormatter())
logger.addHandler(handler)
logger.setLevel(logging.INFO)

# Usage
logger.info("Processing request", extra={
    'extra_data': {
        'endpoint': '/extract_keywords',
        'text_length': len(text),
        'top_n': top_n
    }
})
```

### Performance мониторинг
```python
import time
from functools import wraps

def log_performance(func):
    """Decorator to log function performance"""
    @wraps(func)
    def wrapper(*args, **kwargs):
        start_time = time.time()
        result = func(*args, **kwargs)
        end_time = time.time()
        
        logger.info(f"Function {func.__name__} executed in {end_time - start_time:.2f}s")
        return result
    return wrapper

@log_performance
def extract_keywords(text: str, top_n: int):
    # Your implementation
    pass
```

---

## 🔒 Безопасность

### Input Validation
```python
from pydantic import BaseModel, Field, validator
import re

class SecureTextRequest(BaseModel):
    text: str = Field(..., min_length=1, max_length=10000)
    
    @validator('text')
    def validate_text_security(cls, v):
        """Validate text for security issues"""
        # Check for potential injection attempts
        dangerous_patterns = [
            r'<script.*?>.*?</script>',
            r'javascript:',
            r'on\w+\s*=',
        ]
        
        for pattern in dangerous_patterns:
            if re.search(pattern, v, re.IGNORECASE):
                raise ValueError("Text contains potentially dangerous content")
        
        return v
```

### Rate Limiting
```python
from slowapi import Limiter
from slowapi.util import get_remote_address
from fastapi import Request

limiter = Limiter(key_func=get_remote_address)

@app.post("/extract_keywords")
@limiter.limit("10/minute")
async def extract_keywords_endpoint(
    req: SecureTextRequest,
    request: Request
):
    # Your implementation
    pass
```

---

**Coverage Target:** > 70%  
**Response Time Target:** p95 < 2s  
**Memory Limit:** < 2GB per container