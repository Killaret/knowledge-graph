import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException

from .models import (
    EmbedRequest,
    EmbedResponse,
    ExtractKeywordsRequest,
    ExtractKeywordsResponse,
    Keyword,
    SimilarityRequest,
    SimilarityResponse,
)
from .nlp_utils import (
    EXTRACTOR_NAME,
    _combined_text,
    _embed_chunking_enabled,
    compute_chunked_embedding,
    compute_similarity,
    ensure_model_loaded,
    extract_keywords,
    get_embedding_model,
    is_model_loaded,
)

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Preload embedding model on startup so /health and endpoints respond quickly."""
    logger.info("Preloading embedding model on startup...")
    if not ensure_model_loaded():
        raise RuntimeError("Failed to preload embedding model")
    logger.info("Embedding model ready")
    yield


app = FastAPI(title="NLP Service for Knowledge Graph", lifespan=lifespan)


@app.get("/health")
async def health():
    """Health check — model is preloaded at startup."""
    if not is_model_loaded():
        raise HTTPException(status_code=503, detail="Embedding model not loaded")
    return {
        "status": "healthy",
        "model_loaded": True,
        "version": "1.0.0",
        "extractor": EXTRACTOR_NAME,
    }


@app.post("/extract_keywords", response_model=ExtractKeywordsResponse)
async def extract_keywords_endpoint(req: ExtractKeywordsRequest):
    try:
        keywords = extract_keywords(req.text, req.top_n, req.title)
        return ExtractKeywordsResponse(
            extractor=EXTRACTOR_NAME,
            keywords=[
                Keyword(keyword=lemma, surface=surface, weight=w)
                for lemma, surface, w in keywords
            ],
        )
    except Exception as e:
        logger.exception("Error extracting keywords")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/embed", response_model=EmbedResponse, response_model_exclude_none=True)
async def embed_endpoint(req: EmbedRequest):
    try:
        model = get_embedding_model()
        if _embed_chunking_enabled():
            vec, chunk_count, no_content = compute_chunked_embedding(
                model, req.text, req.title
            )
            return EmbedResponse(
                embedding=vec.tolist(), chunks=chunk_count, no_content=no_content
            )
        embedding = model.encode(_combined_text(req.text, req.title)).tolist()
        return EmbedResponse(embedding=embedding)
    except Exception as e:
        logger.exception("Error computing embedding")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/similarity", response_model=SimilarityResponse)
async def similarity_endpoint(req: SimilarityRequest):
    try:
        similarity = compute_similarity(req.text_a, req.text_b)
        return SimilarityResponse(similarity=similarity)
    except Exception as e:
        logger.exception("Error computing similarity")
        raise HTTPException(status_code=500, detail=str(e))
