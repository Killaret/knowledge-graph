from pydantic import BaseModel
from typing import List

class ExtractKeywordsRequest(BaseModel):
    text: str
    top_n: int = 10

class Keyword(BaseModel):
    keyword: str
    weight: float

class ExtractKeywordsResponse(BaseModel):
    keywords: List[Keyword]

class EmbedRequest(BaseModel):
    text: str

class EmbedResponse(BaseModel):
    embedding: List[float]