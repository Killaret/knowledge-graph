# Cursor Rule: knowledge-graph-nlp

NLP microservice (`nlp-service/`). Python FastAPI + sentence-transformers.
HuggingFace model: `all-MiniLM-L6-v2` (384-dimensional embeddings).

## Manual Found → Automated Covered

If a defect or questionable behavior is discovered during manual testing, create at least one regression test before closing the issue. Choose the test level by severity and scope:

- **unit** — pure logic, validators, or utilities (e.g. `errorMessage.ts`, email validation).
- **integration** — handlers, repositories, or routes (e.g. `PUT /users/me`, `DELETE /users/me` in `router_test.go`).
- **E2E / Playwright** — user-facing scenarios spanning frontend, backend, and data (e.g. public graph, achievements, SSE fallback).

The test should fail before the fix (where safe) and pass after the fix. If the defect depends on manual data or config setup, fix the seed or config script — not only the instructions.

---

## Service Overview

| File | Purpose |
|------|---------|
| `nlp-service/app/main.py` | FastAPI app, endpoint handlers |
| `nlp-service/app/nlp_utils.py` | Model loading, embedding, keyword extraction |
| `nlp-service/app/models.py` | Pydantic request/response models |
| `nlp-service/tests/test_api.py` | FastAPI TestClient tests |
| `nlp-service/tests/test_nlp_utils.py` | Unit tests for NLP utilities |

---

## Lazy Model Loading Pattern

**Never load the model at import time.** The model takes ~15s from cache and
~2-3 minutes to download. Loading at import time blocks the container startup
and fails healthchecks.

```python
# nlp-service/app/nlp_utils.py — actual implementation

_embedding_model: Optional[SentenceTransformer] = None
_model_load_error: Optional[BaseException] = None

def _load_embedding_model() -> SentenceTransformer:
    """Internal loader — called only on demand."""
    global _embedding_model, _model_load_error

    if _embedding_model is not None:
        return _embedding_model      # already loaded — instant return
    if _model_load_error is not None:
        raise _model_load_error      # cached failure — don't retry forever

    _configure_hf_env()
    model_path = _resolve_model_path(local_only=True)  # HF_HUB_OFFLINE=1
    _embedding_model = SentenceTransformer(model_path)
    return _embedding_model

def get_embedding_model() -> SentenceTransformer:
    """Public accessor — lazy loads on first call (~15s from cache)."""
    return _load_embedding_model()

def ensure_model_loaded() -> bool:
    """Returns True if model loaded successfully, False otherwise."""
    try:
        _load_embedding_model()
        return True
    except Exception:
        return False
```

```python
# ❌ WRONG — loads model at import time, blocks container for 15+ seconds
from sentence_transformers import SentenceTransformer
model = SentenceTransformer('all-MiniLM-L6-v2')  # at module level

# ✅ CORRECT — lazy load via get_embedding_model() in endpoint handlers
from .nlp_utils import get_embedding_model, ensure_model_loaded
```

---

## HuggingFace Offline-First Configuration

```python
# nlp-service/app/nlp_utils.py
MODEL_NAME = os.environ.get("MODEL_NAME", "all-MiniLM-L6-v2")
HF_CACHE  = os.environ.get("HF_HOME", "/root/.cache/huggingface")

def _hf_offline_enabled() -> bool:
    return os.environ.get("HF_HUB_OFFLINE", "1").lower() in ("1", "true", "yes")

def _resolve_model_path(local_only: bool) -> str:
    return snapshot_download(
        repo_id=f"sentence-transformers/{MODEL_NAME}",
        cache_dir=HF_CACHE,
        local_files_only=local_only,   # True = never hit network
    )
```

Environment variables (set in docker-compose):
```yaml
environment:
  - MODEL_NAME=all-MiniLM-L6-V2
  - HF_HOME=/root/.cache/huggingface
  - HF_HUB_OFFLINE=1              # offline-first: use cached model only
  - HF_HUB_DISABLE_TELEMETRY=1   # no analytics calls
volumes:
  - ./huggingface_cache:/root/.cache/huggingface  # bind mount pre-populated cache
```

---

## Endpoint Definitions

### `GET /health`
```python
# nlp-service/app/main.py
@app.get("/health")
async def health():
    """Health check — triggers model load on first call."""
    if not ensure_model_loaded():
        raise HTTPException(status_code=503, detail="Embedding model not loaded")
    return {
        "status": "healthy",
        "model_loaded": True,
        "version": "1.0.0",
    }
```

### `POST /embed`
```python
@app.post("/embed", response_model=EmbedResponse)
async def embed_endpoint(req: EmbedRequest):
    try:
        embedding = get_embedding_model().encode(req.text).tolist()
        return EmbedResponse(embedding=embedding)
    except Exception as e:
        logger.exception("Error computing embedding")
        raise HTTPException(status_code=500, detail=str(e))
```

### `POST /extract_keywords`
```python
@app.post("/extract_keywords", response_model=ExtractKeywordsResponse)
async def extract_keywords_endpoint(req: ExtractKeywordsRequest):
    keywords = extract_keywords(req.text, req.top_n)
    return ExtractKeywordsResponse(
        keywords=[Keyword(keyword=kw, weight=w) for kw, w in keywords]
    )
```

---

## Pydantic Models

```python
# nlp-service/app/models.py
from pydantic import BaseModel

class EmbedRequest(BaseModel):
    text: str

class EmbedResponse(BaseModel):
    embedding: list[float]   # 384-dimensional vector for all-MiniLM-L6-v2

class ExtractKeywordsRequest(BaseModel):
    text: str
    top_n: int = 10

class Keyword(BaseModel):
    keyword: str
    weight: float            # 0.0 to 1.0

class ExtractKeywordsResponse(BaseModel):
    keywords: list[Keyword]
```

Always use `BaseModel` for all request/response types. Never return raw dicts
from endpoints.

---

## Startup Timing

```
t=0s    container starts, uvicorn launches
t~1s    FastAPI app is ready (uvicorn bound to port 5000)
t=1s    Docker healthcheck begins pinging /health
t=1s    /health call triggers _load_embedding_model()
t~16s   model loaded from ./huggingface_cache (SSD: ~5s, HDD: ~15s)
t=16s   /health returns 200 → container marked healthy
```

Healthcheck configuration (allows up to 600s × 30 retries):
```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:5000/health"]
  interval: 30s
  timeout: 10s
  retries: 30
  start_period: 600s
```

---

## pytest Patterns

```python
# nlp-service/tests/test_api.py
from fastapi.testclient import TestClient
from unittest.mock import patch, MagicMock
from app.main import app

client = TestClient(app)

def test_health_model_loaded():
    with patch("app.main.ensure_model_loaded", return_value=True):
        response = client.get("/health")
    assert response.status_code == 200
    assert response.json()["status"] == "healthy"

def test_health_model_not_loaded():
    with patch("app.main.ensure_model_loaded", return_value=False):
        response = client.get("/health")
    assert response.status_code == 503

def test_embed_returns_vector():
    mock_model = MagicMock()
    mock_model.encode.return_value = [0.1] * 384
    with patch("app.main.get_embedding_model", return_value=mock_model):
        response = client.post("/embed", json={"text": "hello world"})
    assert response.status_code == 200
    assert len(response.json()["embedding"]) == 384
```

Run tests:
```bash
cd nlp-service
pip install -r requirements.txt
pytest tests/ -v
```

---

## Anti-Patterns

```python
# ❌ Model loaded at import / module level
from sentence_transformers import SentenceTransformer
MODEL = SentenceTransformer("all-MiniLM-L6-v2")  # blocks startup

# ❌ No type hints on endpoint functions
@app.post("/embed")
def embed(req):                   # missing type annotation
    return {"embedding": []}

# ❌ Returning raw dict instead of Pydantic model
@app.post("/embed")
async def embed(req: EmbedRequest):
    return {"embedding": model.encode(req.text).tolist()}  # not EmbedResponse

# ❌ Hardcoded model name (ignores MODEL_NAME env)
model = SentenceTransformer("all-MiniLM-L6-v2")  # use os.environ["MODEL_NAME"]

# ❌ Catching all exceptions silently
try:
    result = model.encode(text)
except:
    pass  # hides failures — always log and re-raise or return 500
```
