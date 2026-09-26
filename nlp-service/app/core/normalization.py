"""NLP-4: deterministic single-pass note normalization.

Pure module: no FastAPI, no model loading, no I/O. The cosine safety guard
needs embeddings, so the caller injects ``embed_fn`` (e.g. ``model.encode``);
without it only the length guard runs.

Rules are ported from the measurement harness
``nlp-service/scripts/measure_normalization.py`` whose corpus run (108 notes)
showed the ruleset reaches a fixed point in one pass — hence v1 is a single
pass with two guards that roll the result back to the source text:

- ``len(result) < min_chars`` -> rollback "too_short"
- ``cos(emb(result), emb(source)) < min_cosine`` -> rollback "low_cosine"

The source text is sacred: normalization never rewrites ``notes.content``;
the result is a derived artifact stored separately by the backend.
"""

from __future__ import annotations

import re
import unicodedata
from dataclasses import dataclass, field
from typing import Callable, Optional, Sequence

import numpy as np

EmbedFn = Callable[[Sequence[str]], Sequence[Sequence[float]]]

BOILERPLATE_PATTERNS = [
    # cookie / consent
    r"(?i)\b(cookie|cookies)\b.*\b(accept|agree|consent|policy)\b",
    r"(?i)(мы|сайт)?\s*(используем|использует)\s+(cookies|куки)",
    r"(?i)\bпринима(ю|ть)\b.*\b(условия|cookies|куки|политик)",
    # subscribe / newsletter / signup walls
    r"(?i)\b(subscribe|subscription|newsletter|sign\s*up|log\s*in|sign\s*in)\b",
    r"(?i)\b(подпис(аться|ка|ывайтесь)|рассылк|зарегистрир|войти|войдите)\b",
    # share / social
    r"(?i)\b(share|поделиться|поделитесь|расскажите друзьям)\b.*\b(vk|telegram|facebook|twitter|x\.com|ок|одноклассник)?\b",
    # footer / copyright
    r"(?i)^\s*(©|&copy;|\(c\))\s*\d{4}",
    r"(?i)\b(all rights reserved|все права защищены)\b",
    # navigation / read more
    r"(?i)^\s*(read more|see also|related (posts?|articles?)|читайте также|смотрите также|далее)\b",
    r"(?i)^\s*(home|главная)\s*[>/»]",
]
BOILERPLATE_RE = [re.compile(p) for p in BOILERPLATE_PATTERNS]

BARE_URL_RE = re.compile(r"^\s*(https?://|www\.)\S+\s*$")
WORD_RE = re.compile(r"\w+", re.UNICODE)

PIPELINE_VERSION = "norm-v1"
DEFAULT_MIN_CHARS = 100
DEFAULT_MIN_COSINE = 0.7


@dataclass(frozen=True)
class NormalizationParams:
    min_chars: int = DEFAULT_MIN_CHARS
    min_cosine: float = DEFAULT_MIN_COSINE
    embed_fn: Optional[EmbedFn] = None  # None disables the cosine guard
    token_counter: Optional[Callable[[str], int]] = None


@dataclass
class NormalizationResult:
    normalized_text: str
    metrics: dict = field(default_factory=dict)
    rolled_back: bool = False
    rollback_reason: Optional[str] = None
    skipped: bool = False


def normalize(text: str, params: Optional[NormalizationParams] = None) -> NormalizationResult:
    """Run one normalization pass with safety rollback.

    Inputs shorter than ``min_chars`` are passed through untouched
    (``skipped=True``) — per spec open question 7, title-only stubs and
    short notes are not normalized.
    """
    if params is None:
        params = NormalizationParams()
    if not isinstance(text, str):
        raise TypeError("text must be a string")

    source = text
    token_count = params.token_counter or _count_words
    raw_tokens = token_count(source)

    if len(source.strip()) < params.min_chars:
        return NormalizationResult(
            normalized_text=source,
            metrics={
                "raw_tokens": raw_tokens,
                "norm_tokens": raw_tokens,
                "compression": 1.0,
                "iterations": 0,
                "stop_reason": "skipped_short",
            },
            skipped=True,
        )

    result = normalize_pass(source)
    metrics = {
        "raw_tokens": raw_tokens,
        "norm_tokens": token_count(result),
        "compression": (token_count(result) / raw_tokens) if raw_tokens else 1.0,
        "iterations": 1,
        "stop_reason": "single_pass",
    }

    if len(result) < params.min_chars:
        metrics["emb_cosine"] = None
        return NormalizationResult(
            normalized_text=source,
            metrics=metrics,
            rolled_back=True,
            rollback_reason="too_short",
        )

    cosine = _guard_cosine(source, result, params.embed_fn)
    metrics["emb_cosine"] = cosine
    if cosine is not None and cosine < params.min_cosine:
        return NormalizationResult(
            normalized_text=source,
            metrics=metrics,
            rolled_back=True,
            rollback_reason="low_cosine",
        )

    return NormalizationResult(normalized_text=result, metrics=metrics)


def normalize_pass(text: str) -> str:
    """One deterministic rules pass: boilerplate lines, bare URLs,
    navigation runs (>=3 consecutive short unpunctuated lines), exact
    near-duplicate lines, whitespace collapse. Idempotent on the measured
    corpus — a second pass finds nothing new."""
    out_lines: list[str] = []
    seen: set[str] = set()
    nav_run: list[str] = []

    def flush_nav() -> None:
        if len(nav_run) >= 3:
            nav_run.clear()
            return
        out_lines.extend(nav_run)
        nav_run.clear()

    for raw in text.splitlines():
        line = raw.strip()
        if not line:
            flush_nav()
            if out_lines and out_lines[-1] != "":
                out_lines.append("")
            continue
        if BARE_URL_RE.match(line):
            nav_run.clear()
            continue
        is_short_nav = len(line) < 30 and not re.search(r"[.!?…:;]$", line)
        if is_short_nav:
            nav_run.append(line)
            continue
        flush_nav()
        if any(p.search(line) for p in BOILERPLATE_RE):
            continue
        fp = _norm_line(line)
        if fp and fp in seen:
            continue
        if fp:
            seen.add(fp)
        out_lines.append(line)
    flush_nav()

    text = "\n".join(out_lines)
    text = re.sub(r"\n{3,}", "\n\n", text)
    return text.strip()


def _norm_line(line: str) -> str:
    """Line fingerprint for near-duplicate detection."""
    line = unicodedata.normalize("NFKC", line).lower()
    line = re.sub(r"[^\w\s]", " ", line)
    return re.sub(r"\s+", " ", line).strip()


def _count_words(text: str) -> int:
    return len(WORD_RE.findall(text))


def _guard_cosine(
    source: str, result: str, embed_fn: Optional[EmbedFn]
) -> Optional[float]:
    if embed_fn is None:
        return None
    vectors = embed_fn([source, result])
    a, b = np.asarray(vectors[0], dtype=np.float32), np.asarray(
        vectors[1], dtype=np.float32
    )
    na, nb = float(np.linalg.norm(a)), float(np.linalg.norm(b))
    if na == 0 or nb == 0:
        return 0.0
    return float(np.dot(a, b) / (na * nb))
