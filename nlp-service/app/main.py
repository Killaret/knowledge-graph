import logging

from fastapi import FastAPI, HTTPException

from .models import (
    EmbedRequest,
    EmbedResponse,
    ExtractKeywordsRequest,
    ExtractKeywordsResponse,
    Keyword,
)
from .nlp_utils import ensure_model_loaded, extract_keywords, get_embedding_model

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="NLP Service for Knowledge Graph")


@app.get("/health")
async def health():
    """Health check — loads model from local cache on first call."""
    if not ensure_model_loaded():
        raise HTTPException(status_code=503, detail="Embedding model not loaded")
    return {
        "status": "healthy",
        "model_loaded": True,
        "version": "1.0.0",
    }


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


@app.post("/embed", response_model=EmbedResponse)
async def embed_endpoint(req: EmbedRequest):
    try:
        embedding = get_embedding_model().encode(req.text).tolist()
        return EmbedResponse(embedding=embedding)
    except Exception as e:
        logger.exception("Error computing embedding")
        raise HTTPException(status_code=500, detail=str(e))
