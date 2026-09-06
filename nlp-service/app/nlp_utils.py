import logging
import os
import threading
from typing import Optional

import nltk
import yake
from huggingface_hub import snapshot_download
from sentence_transformers import SentenceTransformer

logger = logging.getLogger(__name__)

MODEL_NAME = os.environ.get("NLP_MODEL_NAME") or os.environ.get("MODEL_NAME", "all-MiniLM-L6-v2")
HF_HOME = os.environ.get("HF_HOME", "/root/.cache/huggingface")
HF_CACHE = os.environ.get("HF_HUB_CACHE") or os.path.join(HF_HOME, "hub")

_embedding_model: Optional[SentenceTransformer] = None
_model_load_error: Optional[BaseException] = None
_model_lock = threading.Lock()

# NLTK data (lightweight — safe at import time)
try:
    nltk.data.find("tokenizers/punkt")
except LookupError:
    nltk.download("punkt")
try:
    nltk.data.find("corpora/stopwords")
except LookupError:
    nltk.download("stopwords")

stop_words = set(
    nltk.corpus.stopwords.words("russian") + nltk.corpus.stopwords.words("english")
)
kw_extractor = yake.KeywordExtractor(lan="ru", top=20, stopwords=stop_words)


def _hf_offline_enabled() -> bool:
    return os.environ.get("HF_HUB_OFFLINE", "1").lower() in ("1", "true", "yes")


def _configure_hf_env() -> None:
    os.environ.setdefault("HF_HUB_DISABLE_TELEMETRY", "1")
    os.environ.setdefault("HF_HUB_DISABLE_SYMLINKS", "1")


def _resolve_model_path(local_only: bool) -> str:
    """Resolve model directory from HuggingFace cache."""
    return snapshot_download(
        repo_id=f"sentence-transformers/{MODEL_NAME}",
        cache_dir=HF_CACHE,
        local_files_only=local_only,
    )


def _load_embedding_model() -> SentenceTransformer:
    global _embedding_model, _model_load_error

    with _model_lock:
        if _embedding_model is not None:
            return _embedding_model
        if _model_load_error is not None:
            raise _model_load_error

        _configure_hf_env()
        attempts: list[tuple[str, bool]] = [("cache (offline)", True)]
        if not _hf_offline_enabled():
            attempts.append(("network", False))

        last_error: Optional[BaseException] = None
        for label, local_only in attempts:
            try:
                model_path = _resolve_model_path(local_only)
                model = SentenceTransformer(model_path)
                _embedding_model = model
                logger.info("Embedding model loaded via %s from %s", label, model_path)
                return model
            except Exception as exc:
                last_error = exc
                logger.warning("Embedding model load (%s) failed: %s", label, exc)

        _model_load_error = last_error or RuntimeError("Failed to load embedding model")
        logger.error("All embedding model load attempts failed")
        raise _model_load_error


def get_embedding_model() -> SentenceTransformer:
    """Return the embedding model, loading from local cache on first use."""
    return _load_embedding_model()


def is_model_loaded() -> bool:
    with _model_lock:
        return _embedding_model is not None


def ensure_model_loaded() -> bool:
    try:
        _load_embedding_model()
        return True
    except Exception:
        return False


def extract_keywords(text: str, top_n: int = 10) -> list:
    if not text or not str(text).strip():
        return []
    keywords = kw_extractor.extract_keywords(text)
    result = []
    for kw, score in keywords[:top_n]:
        weight = max(0.0, min(1.0, 1.0 - score))
        result.append((kw, weight))
    return result


def _cosine_similarity(a: list[float], b: list[float]) -> float:
    if len(a) != len(b):
        raise ValueError("embedding dimensions do not match")
    dot = sum(x * y for x, y in zip(a, b))
    norm_a = sum(x * x for x in a) ** 0.5
    norm_b = sum(x * x for x in b) ** 0.5
    if norm_a == 0 or norm_b == 0:
        return 0.0
    similarity = dot / (norm_a * norm_b)
    # SentenceTransformer cosine similarity typically lives in [-1, 1].
    # Clamp to [0, 1] and map to a non-negative weight usable for link strength.
    return max(0.0, min(1.0, (similarity + 1.0) / 2.0))


def _as_list(embedding):
    if hasattr(embedding, "tolist"):
        return embedding.tolist()
    return list(embedding)


def compute_similarity(text_a: str, text_b: str) -> float:
    """Compute cosine similarity between two texts using the embedding model."""
    if not text_a or not text_b:
        return 0.0
    model = get_embedding_model()
    embeddings = model.encode([text_a, text_b], convert_to_numpy=False)
    return _cosine_similarity(_as_list(embeddings[0]), _as_list(embeddings[1]))
