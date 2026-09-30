from pydantic import BaseModel, Field
from typing import List

class ExtractKeywordsRequest(BaseModel):
    text: str = Field(..., max_length=10000, description="Text to extract keywords from")
    top_n: int = Field(default=10, ge=1, le=50, description="Number of keywords to return (1-50)")
    title: str = Field(default="", max_length=1000, description="Note title (injected into chunk embeddings when EMBED_CHUNKING=on)")

class Keyword(BaseModel):
    keyword: str
    surface: str = ""
    weight: float

class ExtractKeywordsResponse(BaseModel):
    extractor: str = ""
    keywords: List[Keyword]

class EmbedRequest(BaseModel):
    text: str = Field(..., max_length=10000, description="Text to generate embedding for")
    title: str = Field(default="", max_length=1000, description="Note title (injected into chunk embeddings when EMBED_CHUNKING=on)")

class EmbedResponse(BaseModel):
    embedding: List[float]
    chunks: int | None = None
    no_content: bool | None = None


class NormalizeRequest(BaseModel):
    text: str = Field(..., max_length=10000, description="Source note content (never modified upstream)")
    title: str = Field(default="", max_length=1000, description="Note title")


class NormalizedChunk(BaseModel):
    idx: int
    text: str
    heading_path: List[str]
    char_span: List[int]  # [start, end] within normalized_text
    token_count: int
    kind: str
    forced_split: bool = False


class NormalizeMetrics(BaseModel):
    raw_tokens: int
    norm_tokens: int
    compression: float
    iterations: int
    stop_reason: str
    emb_cosine: float | None = None


class NormalizeResponse(BaseModel):
    normalized_text: str
    chunks: List[NormalizedChunk]
    metrics: NormalizeMetrics
    rolled_back: bool
    rollback_reason: str | None = None
    skipped: bool = False
    pipeline_version: str


class SimilarityRequest(BaseModel):
    text_a: str = Field(..., max_length=10000, description="First text")
    text_b: str = Field(..., max_length=10000, description="Second text")


class SimilarityResponse(BaseModel):
    similarity: float = Field(..., ge=0.0, le=1.0, description="Cosine similarity in [0, 1]")