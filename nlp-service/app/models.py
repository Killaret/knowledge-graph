from pydantic import BaseModel, Field
from typing import List

class ExtractKeywordsRequest(BaseModel):
    text: str = Field(..., max_length=10000, description="Text to extract keywords from")
    top_n: int = Field(default=10, ge=1, le=50, description="Number of keywords to return (1-50)")

class Keyword(BaseModel):
    keyword: str
    weight: float

class ExtractKeywordsResponse(BaseModel):
    keywords: List[Keyword]

class EmbedRequest(BaseModel):
    text: str = Field(..., max_length=10000, description="Text to generate embedding for")

class EmbedResponse(BaseModel):
    embedding: List[float]