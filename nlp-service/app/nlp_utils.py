import logging
import os
import re
import threading
from typing import Optional

import nltk
from huggingface_hub import snapshot_download
from sentence_transformers import SentenceTransformer

logger = logging.getLogger(__name__)

MODEL_NAME = os.environ.get("NLP_MODEL_NAME", "paraphrase-multilingual-MiniLM-L12-v2")
HF_HOME = os.environ.get("HF_HOME", "/root/.cache/huggingface")
HF_CACHE = os.environ.get("HF_HUB_CACHE") or os.path.join(HF_HOME, "hub")
MODEL_ALLOW_PATTERNS = ["*.json", "*.txt", "*.model", "*.safetensors"]

# Extractor version reported in /extract_keywords and stored in
# note_keywords.extractor — bump when the algorithm changes so stale rows are
# distinguishable from fresh ones.
EXTRACTOR_NAME = "keybert-hybrid-0.9"

_CANDIDATE_LIMIT = 100  # statistical short-list size over the whole text
_DOC_CHUNK_WORDS = 100  # ~within the model's 128-token window

_embedding_model: Optional[SentenceTransformer] = None
_model_load_error: Optional[BaseException] = None
_model_lock = threading.Lock()

_morph_ru = None   # lazy pymorphy3.MorphAnalyzer singleton
_morph_en = None   # lazy nltk WordNetLemmatizer singleton
_morph_lock = threading.Lock()

# NLTK data (lightweight — safe at import time)
try:
    nltk.data.find("tokenizers/punkt")
except LookupError:
    nltk.download("punkt")
try:
    nltk.data.find("corpora/stopwords")
except LookupError:
    nltk.download("stopwords")
try:
    nltk.data.find("corpora/wordnet")
except LookupError:
    nltk.download("wordnet")

stop_words = set(
    nltk.corpus.stopwords.words("russian") + nltk.corpus.stopwords.words("english")
)

_TOKEN_RE = re.compile(r"[0-9A-Za-zА-Яа-яЁё]+(?:-[0-9A-Za-zА-Яа-яЁё]+)*")
_CYRILLIC_RE = re.compile(r"[А-Яа-яЁё]")
_LATIN_RE = re.compile(r"[A-Za-z]")


def _hf_offline_enabled() -> bool:
    return os.environ.get("HF_HUB_OFFLINE", "1").lower() in ("1", "true", "yes")


def _embed_chunking_enabled() -> bool:
    return os.environ.get("EMBED_CHUNKING", "0").strip().lower() in (
        "1",
        "true",
        "yes",
        "on",
    )


def _normalization_min_cosine() -> float:
    """NLP-4 cosine guard threshold (nlp.normalization.min_cosine).
    Depends on the model's similarity scale — 0.7 was measured on e5-base;
    recalibration for the chosen model belongs to MODEL-2."""
    try:
        return float(os.environ.get("NLP_NORMALIZATION_MIN_COSINE", "0.7"))
    except ValueError:
        return 0.7


def _combined_text(text: str, title: Optional[str] = None) -> str:
    """Legacy embedding input: content plus title joined by a single space,
    matching what the worker used to concatenate before CHUNK-1."""
    if title:
        return f"{title} {text}"
    return text


def _configure_hf_env() -> None:
    os.environ.setdefault("HF_HUB_DISABLE_TELEMETRY", "1")
    os.environ.setdefault("HF_HUB_DISABLE_SYMLINKS", "1")


def _resolve_model_path(local_only: bool) -> str:
    """Resolve model directory from HuggingFace cache."""
    return snapshot_download(
        repo_id=f"sentence-transformers/{MODEL_NAME}",
        cache_dir=HF_CACHE,
        local_files_only=local_only,
        allow_patterns=MODEL_ALLOW_PATTERNS,
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


def _get_morph_ru():
    global _morph_ru
    with _morph_lock:
        if _morph_ru is None:
            import pymorphy3

            _morph_ru = pymorphy3.MorphAnalyzer()
        return _morph_ru


def _get_morph_en():
    global _morph_en
    with _morph_lock:
        if _morph_en is None:
            from nltk.stem import WordNetLemmatizer

            _morph_en = WordNetLemmatizer()
        return _morph_en


def _lemmatize_token(token: str) -> str:
    """Route by alphabet: cyrillic -> pymorphy3, latin -> WordNet, else as-is."""
    if _CYRILLIC_RE.search(token):
        parsed = _get_morph_ru().parse(token)
        return parsed[0].normal_form if parsed else token
    if _LATIN_RE.search(token):
        lem = _get_morph_en()
        # Try noun first, then verb/adjective/adverb — pick the first form
        # that differs, so "networks" -> "network" and "running" -> "run".
        for pos in ("n", "v", "a", "r"):
            lemma = lem.lemmatize(token, pos=pos)
            if lemma != token:
                return lemma
        return token
    return token


def lemmatize_phrase(phrase: str) -> str:
    """Lemmatize each token of a phrase and join with spaces (lowercase in)."""
    return " ".join(_lemmatize_token(t) for t in phrase.split())


def _statistical_candidates(text: str, limit: int = _CANDIDATE_LIMIT) -> list[str]:
    """Short-list of candidate keyphrases from token statistics over the
    whole document: unigrams and bigrams of adjacent non-stopword tokens,
    scored by frequency plus an early-position bonus. A bigram is only
    formed when the gap between tokens is pure whitespace — never across
    sentence punctuation."""
    matches = list(_TOKEN_RE.finditer(text))
    tokens = [m.group(0).lower() for m in matches]
    if not tokens:
        return []

    is_content = [t not in stop_words for t in tokens]
    freq: dict[str, int] = {}
    first_pos: dict[str, int] = {}
    content_idx = [i for i, ok in enumerate(is_content) if ok]

    def add(phrase: str, pos: int) -> None:
        freq[phrase] = freq.get(phrase, 0) + 1
        if phrase not in first_pos:
            first_pos[phrase] = pos

    for seq, i in enumerate(content_idx):
        unigram = tokens[i]
        add(unigram, seq)
        if seq + 1 < len(content_idx) and content_idx[seq + 1] == i + 1:
            gap = text[matches[i].end() : matches[i + 1].start()]
            if gap.strip() == "":
                add(unigram + " " + tokens[i + 1], seq)

    scored = sorted(
        freq,
        key=lambda p: (freq[p] + 1.0 / (1.0 + first_pos[p]), len(p.split())),
        reverse=True,
    )
    return scored[:limit]


def _chunk_token_counter(model: SentenceTransformer):
    """Token counter for the structural chunker: model tokenizer when
    available, word+punctuation count otherwise (keeps Java-portable)."""
    tokenizer = getattr(model, "tokenizer", None)
    if tokenizer is not None:

        def count_tokens(text: str) -> int:
            return len(tokenizer.encode(text, add_special_tokens=False))

        return count_tokens

    def count_words(text: str) -> int:
        return len(re.findall(r"\w+|[^\w\s]", text, flags=re.UNICODE))

    return count_words


def compute_chunked_embedding(
    model: SentenceTransformer, text: str, title: Optional[str] = None
):
    """CHUNK-1 structural embedding: structure-aware chunks, one batched
    encode, mean vector + L2 normalization. Title is injected into every
    chunk's model input (never into the stored chunk text). Empty/stub
    notes produce zero content chunks and embed the title alone.
    Returns (vector, chunk_count, no_content)."""
    import numpy as np

    from .core.chunking import ChunkingParams, aggregate, chunk, embedding_inputs

    params = ChunkingParams(
        token_counter=_chunk_token_counter(model),
        target_tokens=256,
        max_tokens=int(getattr(model, "max_seq_length", 512) or 512),
        title=title or None,
        title_injection=True,
    )
    chunks = chunk(text, params)
    inputs = embedding_inputs(chunks, params)
    if not inputs:
        inputs = [params.title or ""]
    vectors = np.asarray(model.encode(inputs), dtype=np.float32)
    if vectors.ndim == 1:
        vectors = vectors.reshape(1, -1)
    return aggregate(vectors), len(chunks), len(chunks) == 0


def _doc_vector(text: str, title: Optional[str] = None):
    """Document embedding as the mean of per-chunk embeddings, so the whole
    text contributes instead of only the model's first-window tokens.
    With EMBED_CHUNKING on, the structural chunker replaces word slices."""
    import numpy as np

    model = get_embedding_model()
    if _embed_chunking_enabled():
        vec, _chunk_count, _no_content = compute_chunked_embedding(
            model, text, title
        )
        return vec
    words = _combined_text(text, title).split()
    if not words:
        return None
    chunks = [
        " ".join(words[i : i + _DOC_CHUNK_WORDS])
        for i in range(0, len(words), _DOC_CHUNK_WORDS)
    ]
    vectors = model.encode(chunks, convert_to_numpy=True)
    vec = np.asarray(vectors, dtype=np.float32).mean(axis=0)
    norm = float(np.linalg.norm(vec))
    return vec / norm if norm else vec


def extract_keywords(
    text: str, top_n: int = 10, title: Optional[str] = None
) -> list:
    """Hybrid keyphrase extraction (NLP-2): a statistical short-list covers
    the whole document; the embedding model only *ranks* it against a
    chunked document vector. Returns [(lemma, surface, weight)] sorted by
    weight desc — weight is the cosine similarity clamped to [0, 1]."""
    if not text or not str(_combined_text(text, title)).strip():
        return []

    full_text = _combined_text(text, title)

    candidates = _statistical_candidates(full_text)
    if not candidates:
        return []

    import numpy as np

    model = get_embedding_model()
    doc_vec = _doc_vector(text, title)
    if doc_vec is None:
        return []

    cand_vecs = np.asarray(
        model.encode(candidates, convert_to_numpy=True), dtype=np.float32
    )
    norms = np.linalg.norm(cand_vecs, axis=1, keepdims=True)
    norms[norms == 0] = 1.0
    sims = (cand_vecs / norms) @ doc_vec

    ranked = sorted(
        zip(candidates, sims.tolist()), key=lambda kv: kv[1], reverse=True
    )

    # Lemmatize, drop stopword lemmas, deduplicate by lemma keeping the
    # highest-scoring surface form.
    seen: dict[str, tuple[str, float]] = {}
    for surface, score in ranked:
        weight = max(0.0, min(1.0, float(score)))
        lemma = lemmatize_phrase(surface)
        if not lemma or all(t in stop_words for t in lemma.split()):
            continue
        if lemma not in seen or weight > seen[lemma][1]:
            seen[lemma] = (surface, weight)

    result = [(lemma, surface, w) for lemma, (surface, w) in seen.items()]
    result.sort(key=lambda r: r[2], reverse=True)
    return result[:top_n]


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
