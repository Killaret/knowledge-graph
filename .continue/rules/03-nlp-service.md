---
name: NLP Service Rules
alwaysApply: false
globs: ["nlp-service/**/*.py"]
description: Python FastAPI NLP service - sentence-transformers, lazy loading, HuggingFace offline-first
---

# NLP Service Rules

## Manual Found → Automated Covered

If a defect or questionable behavior is discovered during manual testing, create at least one regression test before closing the issue. Choose the test level by severity and scope:

- **unit** — pure logic, validators, or utilities (e.g. `errorMessage.ts`, email validation).
- **integration** — handlers, repositories, or routes (e.g. `PUT /users/me`, `DELETE /users/me` in `router_test.go`).
- **E2E / Playwright** — user-facing scenarios spanning frontend, backend, and data (e.g. public graph, achievements, SSE fallback).

The test should fail before the fix (where safe) and pass after the fix. If the defect depends on manual data or config setup, fix the seed or config script — not only the instructions.

## Overview

The NLP service provides embedding generation and keyword extraction via FastAPI.
It uses sentence-transformers with HuggingFace models, loaded lazily from local cache.

## Architecture

```
nlp-service/
├── app/
│   ├── __init__.py
│   ├── main.py         # FastAPI app, endpoints
│   ├── models.py       # Pydantic request/response models
│   └── nlp_utils.py    # Model loading, embedding, keyword extraction
└── tests/
    ├── test_api.py     # API endpoint tests
    └── test_nlp_utils.py  # Unit tests for NLP utilities
```

## Lazy Loading Pattern — CRITICAL

The embedding model (~90MB) MUST be loaded lazily on first request, NOT at import time.
This allows uvicorn to start fast (~1s) while the model loads on first health check (~15s).

```python
# ✅ Good — lazy loading with singleton pattern
_embedding_model: Optional[SentenceTransformer] = None

def get_embedding_model() -> SentenceTransformer:
    """Return the embedding model, loading from local cache on first use."""
    global _embedding_model
    if _embedding_model is not None:
        return _embedding_model
    # Load from cache
    model_path = snapshot_download(
        repo_id=f"sentence-transformers/{MODEL_NAME}",
        cache_dir=HF_CACHE,
        local_files_only=True,
    )
    _embedding_model = SentenceTransformer(model_path)
    return _embedding_model


def ensure_model_loaded() -> bool:
    """Attempt to load model, return True if successful."""
    try:
        get_embedding_model()
        return True
    except Exception:
        return False
```

```python
# ❌ Bad — loading at import time (blocks startup)
from sentence_transformers import SentenceTransformer
model = SentenceTransformer("all-MiniLM-L6-v2")  # BAD: 15s startup delay
```

## HuggingFace Offline-First

```python
# Environment variables (set in docker-compose.yml)
MODEL_NAME = os.environ.get("MODEL_NAME", "all-MiniLM-L6-v2")
HF_CACHE = os.environ.get("HF_HOME", "/root/.cache/huggingface")

def _hf_offline_enabled() -> bool:
    """Check if offline mode is enabled (default: yes)."""
    return os.environ.get("HF_HUB_OFFLINE", "1").lower() in ("1", "true", "yes")

def _configure_hf_env() -> None:
    """Disable telemetry and symlinks for containerized usage."""
    os.environ.setdefault("HF_HUB_DISABLE_TELEMETRY", "1")
    os.environ.setdefault("HF_HUB_DISABLE_SYMLINKS", "1")
```

**Rules:**
- `HF_HUB_OFFLINE=1` — NEVER download models from internet in production
- Models must be pre-cached in `huggingface_cache/` volume
- `HF_HOME` points to the cache directory inside the container
- First attempt always uses `local_files_only=True`

## FastAPI Endpoint Pattern

```python
from fastapi import FastAPI, HTTPException

app = FastAPI(title="NLP Service for Knowledge Graph")

@app.get("/health")
async def health():
    """Health check — triggers model loading on first call."""
    if not ensure_model_loaded():
        raise HTTPException(status_code=503, detail="Embedding model not loaded")
    return {"status": "healthy", "model_loaded": True, "version": "1.0.0"}

@app.post("/embed", response_model=EmbedResponse)
async def embed_endpoint(req: EmbedRequest):
    try:
        embedding = get_embedding_model().encode(req.text).tolist()
        return EmbedResponse(embedding=embedding)
    except Exception as e:
        logger.exception("Error computing embedding")
        raise HTTPException(status_code=500, detail=str(e))
```

## Pydantic Models

```python
from pydantic import BaseModel, Field

class EmbedRequest(BaseModel):
    text: str = Field(..., min_length=1, max_length=10000)

class EmbedResponse(BaseModel):
    embedding: list[float]

class ExtractKeywordsRequest(BaseModel):
    text: str = Field(..., min_length=1)
    top_n: int = Field(default=10, ge=1, le=50)

class Keyword(BaseModel):
    keyword: str
    weight: float

class ExtractKeywordsResponse(BaseModel):
    keywords: list[Keyword]
```

## Type Hints — Required

```python
# ✅ Good — full type annotations
def extract_keywords(text: str, top_n: int = 10) -> list[tuple[str, float]]:
    if not text or not str(text).strip():
        return []
    keywords = kw_extractor.extract_keywords(text)
    return [(kw, max(0.0, min(1.0, 1.0 - score))) for kw, score in keywords[:top_n]]

# ❌ Bad — no type hints
def extract_keywords(text, top_n=10):
    ...
```

## Pytest Test Pattern

```python
import pytest
from httpx import AsyncClient, ASGITransport
from app.main import app

@pytest.fixture
def client():
    transport = ASGITransport(app=app)
    return AsyncClient(transport=transport, base_url="http://test")

@pytest.mark.asyncio
async def test_health_returns_healthy(client):
    response = await client.get("/health")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "healthy"
    assert data["model_loaded"] is True

@pytest.mark.asyncio
async def test_embed_returns_vector(client):
    response = await client.post("/embed", json={"text": "Hello world"})
    assert response.status_code == 200
    data = response.json()
    assert len(data["embedding"]) == 384  # all-MiniLM-L6-v2 dimension

@pytest.mark.asyncio
async def test_embed_empty_text_returns_422(client):
    response = await client.post("/embed", json={"text": ""})
    assert response.status_code == 422
```

## Anti-Patterns

```python
# ❌ Bad — loading model at module level
model = SentenceTransformer("all-MiniLM-L6-v2")

# ❌ Bad — downloading from network in production
snapshot_download(repo_id="...", local_files_only=False)

# ❌ Bad — catching and silently swallowing errors
try:
    embedding = model.encode(text)
except:
    pass

# ✅ Good — proper error handling with logging
try:
    embedding = get_embedding_model().encode(text)
except Exception as e:
    logger.exception("Error computing embedding")
    raise HTTPException(status_code=500, detail=str(e))
```

## Docker Configuration

```yaml
nlp:
  build: ./nlp-service
  environment:
    - MODEL_NAME=all-MiniLM-L6-V2
    - HF_HOME=/root/.cache/huggingface
    - HF_HUB_DISABLE_TELEMETRY=1
    - HF_HUB_OFFLINE=1
  volumes:
    - ./huggingface_cache:/root/.cache/huggingface
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:5000/health"]
    interval: 30s
    timeout: 10s
    retries: 30
    start_period: 600s  # Model loading can take time on cold start
```
