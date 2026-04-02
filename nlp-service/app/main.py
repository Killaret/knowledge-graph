import logging
from fastapi import FastAPI, HTTPException
from .models import ExtractKeywordsRequest, ExtractKeywordsResponse, Keyword, EmbedRequest, EmbedResponse
from .nlp_utils import extract_keywords, embedding_model

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="NLP Service for Knowledge Graph")

@app.get("/health")
async def health():
    return {"status": "ok"}

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
        embedding = embedding_model.encode(req.text).tolist()
        return EmbedResponse(embedding=embedding)
    except Exception as e:
        logger.exception("Error computing embedding")
        raise HTTPException(status_code=500, detail=str(e))