from __future__ import annotations

import re
from dataclasses import dataclass
from typing import Callable, Sequence

import numpy as np


ChunkKind = str
TokenCounter = Callable[[str], int]


@dataclass(frozen=True)
class ChunkingParams:
    token_counter: TokenCounter
    target_tokens: int = 256
    max_tokens: int = 512
    overlap: int = 0
    title: str | None = None
    title_injection: bool = True


@dataclass
class Chunk:
    idx: int
    text: str
    heading_path: list[str]
    char_span: tuple[int, int]
    token_count: int
    kind: ChunkKind
    forced_split: bool = False


@dataclass(frozen=True)
class _Block:
    kind: ChunkKind
    start: int
    end: int
    heading_path: list[str]


_SENTENCE_END = frozenset(".!?…")
_CLOSERS = frozenset('"»”)]}')
_ABBREVIATIONS = frozenset(
    {
        "т.е",
        "т. е",
        "т.к",
        "т. к",
        "т.п",
        "т. п",
        "т.д",
        "т. д",
        "т.н",
        "т. н",
        "т.о",
        "т. о",
        "и т.д",
        "и т. п",
        "и т. е",
        "и т. к",
        "др",
        "г",
        "с",
        "стр",
        "рис",
        "см",
        "напр",
        "им",
        "св",
        "e.g",
        "i.e",
        "etc",
        "vs",
        "viz",
        "cf",
        "ca",
        "approx",
        "dr",
        "mr",
        "mrs",
        "ms",
        "prof",
        "sr",
        "jr",
        "st",
        "no",
        "fig",
        "inc",
        "ltd",
        "jan",
        "feb",
        "mar",
        "apr",
        "jun",
        "jul",
        "aug",
        "sep",
        "sept",
        "oct",
        "nov",
        "dec",
    }
)
_HEADING_RE = re.compile(r"^(#{1,6})\s+(.+?)\s*#*\s*$")
_WORD_RE = re.compile(r"\S+")
_HEADING_LINK_STUB_RE = re.compile(
    r"\A\s*#{1,6}\s*\[[^\]\n]+\]\([^)\n]+\)\s*\Z",
    re.UNICODE,
)


def chunk(text: str, params: ChunkingParams) -> list[Chunk]:
    if not isinstance(text, str):
        raise TypeError("text must be a string")
    if params.target_tokens <= 0:
        raise ValueError("target_tokens must be positive")
    if params.max_tokens <= 0:
        raise ValueError("max_tokens must be positive")
    if params.overlap != 0:
        raise ValueError("overlap other than zero is not supported in v1")
    if _is_heading_link_stub(text):
        return []

    max_tokens = _effective_max_tokens(params)
    target_tokens = min(params.target_tokens, max_tokens)
    result: list[Chunk] = []

    for block in _parse_blocks(text, params.token_counter):
        units = _block_units(text, block)
        result.extend(
            _pack_units(text, units, block, target_tokens, max_tokens, params)
        )

    for idx, item in enumerate(result):
        item.idx = idx
    return result


def embedding_inputs(
    chunks: Sequence[Chunk], params: ChunkingParams
) -> list[str]:
    title = (params.title or "").strip()
    if not title or not params.title_injection:
        return [item.text for item in chunks]
    return [f"{title}\n\n{item.text}" for item in chunks]


def aggregate(vectors: Sequence[Sequence[float]]) -> np.ndarray | None:
    if vectors is None or len(vectors) == 0:
        return None
    matrix = np.asarray(vectors, dtype=np.float32)
    if matrix.size == 0:
        return None
    if matrix.ndim == 1:
        matrix = matrix.reshape(1, -1)
    mean = matrix.mean(axis=0)
    norm = np.linalg.norm(mean)
    if norm == 0:
        return mean
    return mean / norm


def _is_heading_link_stub(text: str) -> bool:
    return bool(_HEADING_LINK_STUB_RE.match(text))


def _title_token_overhead(params: ChunkingParams) -> int:
    title = (params.title or "").strip()
    if not title or not params.title_injection:
        return 0
    return params.token_counter(f"{title}\n\n")


def _effective_max_tokens(params: ChunkingParams) -> int:
    overhead = _title_token_overhead(params)
    if overhead >= params.max_tokens:
        return params.max_tokens
    return max(1, params.max_tokens - overhead)


def _trim_span(text: str, start: int, end: int) -> tuple[int, int]:
    while start < end and text[start].isspace():
        start += 1
    while end > start and text[end - 1].isspace():
        end -= 1
    return start, end


def _line_spans(text: str, start: int, end: int) -> list[tuple[int, int]]:
    spans: list[tuple[int, int]] = []
    cursor = start
    while cursor < end:
        newline = text.find("\n", cursor, end)
        line_end = end if newline == -1 else newline + 1
        span = _trim_span(text, cursor, line_end)
        if span[0] < span[1]:
            spans.append(span)
        cursor = line_end
    return spans


def _is_prose_interrupt(line: str) -> bool:
    stripped = line.lstrip()
    if not stripped:
        return False
    return bool(
        stripped.startswith("```")
        or stripped.startswith("|")
        or _HEADING_RE.match(line)
        or line.startswith("    ")
        or line.startswith("\t")
    )


def _parse_blocks(
    text: str, token_counter: TokenCounter
) -> list[_Block]:
    raw_lines = text.splitlines(keepends=True)
    if not raw_lines:
        return []

    line_starts: list[int] = []
    offset = 0
    for line in raw_lines:
        line_starts.append(offset)
        offset += len(line)

    blocks: list[_Block] = []
    heading_path: list[str] = []
    i = 0

    def add_block(kind: ChunkKind, start: int, end: int, path: list[str]) -> None:
        start, end = _trim_span(text, start, end)
        if start < end:
            blocks.append(_Block(kind=kind, start=start, end=end, heading_path=path.copy()))

    while i < len(raw_lines):
        line = raw_lines[i]
        content = line.rstrip("\r\n")
        stripped = content.lstrip()
        if not stripped:
            i += 1
            continue

        start = line_starts[i]
        path = heading_path

        if stripped.startswith("```"):
            end = line_starts[i] + len(line)
            i += 1
            while i < len(raw_lines):
                end = line_starts[i] + len(raw_lines[i])
                is_closing = raw_lines[i].lstrip().startswith("```")
                i += 1
                if is_closing:
                    break
            add_block("code", start, end, path)
            continue

        if content.startswith("    ") or content.startswith("\t"):
            end = line_starts[i] + len(line)
            i += 1
            while i < len(raw_lines):
                next_content = raw_lines[i].rstrip("\r\n")
                if not next_content.startswith("    ") and not next_content.startswith("\t"):
                    break
                end = line_starts[i] + len(raw_lines[i])
                i += 1
            add_block("code", start, end, path)
            continue

        if stripped.startswith("|"):
            end = line_starts[i] + len(line)
            i += 1
            while i < len(raw_lines):
                next_content = raw_lines[i].rstrip("\r\n")
                if not next_content.lstrip().startswith("|"):
                    break
                end = line_starts[i] + len(raw_lines[i])
                i += 1
            add_block("table", start, end, path)
            continue

        heading = _HEADING_RE.match(content)
        if heading:
            level = len(heading.group(1))
            title = heading.group(2).strip()
            heading_path = heading_path[: level - 1] + [title]
            add_block(
                "prose",
                start,
                line_starts[i] + len(line),
                heading_path,
            )
            i += 1
            continue

        end = line_starts[i] + len(line)
        i += 1
        while i < len(raw_lines):
            next_line = raw_lines[i]
            next_content = next_line.rstrip("\r\n")
            if not next_content.strip() or _is_prose_interrupt(next_content):
                break
            end = line_starts[i] + len(next_line)
            i += 1
        add_block("prose", start, end, path)

    return blocks


def _block_units(text: str, block: _Block) -> list[tuple[int, int]]:
    if block.kind != "prose":
        return [(block.start, block.end)]
    return _sentence_spans(text, block.start, block.end)


def _sentence_spans(text: str, start: int, end: int) -> list[tuple[int, int]]:
    spans: list[tuple[int, int]] = []
    sentence_start = start
    i = start

    while i < end:
        if text[i] not in _SENTENCE_END:
            i += 1
            continue

        punctuation_end = i + 1
        while punctuation_end < end and text[punctuation_end] in _SENTENCE_END:
            punctuation_end += 1
        boundary_end = punctuation_end
        while boundary_end < end and text[boundary_end] in _CLOSERS:
            boundary_end += 1

        if _is_false_boundary(text, start, i):
            i = punctuation_end
            continue

        sentence_end = boundary_end
        while sentence_end < end and text[sentence_end].isspace():
            sentence_end += 1
        if sentence_end < end:
            span = _trim_span(text, sentence_start, sentence_end)
            if span[0] < span[1]:
                spans.append(span)
            sentence_start = sentence_end
        i = max(punctuation_end, sentence_end)

    tail = _trim_span(text, sentence_start, end)
    if tail[0] < tail[1]:
        spans.append(tail)
    return spans


def _is_false_boundary(text: str, block_start: int, punctuation_index: int) -> bool:
    prefix = text[block_start:punctuation_index]
    match = re.search(r"([^\s]+)$", prefix)
    if not match:
        return False
    token = match.group(1)
    normalized = token.rstrip('"\')]}»”').lower()
    following = text[punctuation_index + 1 : punctuation_index + 24]

    if re.fullmatch(r"[a-zа-яё]", normalized) and re.match(
        r"\s*[a-zа-яё]+\.", following, flags=re.IGNORECASE
    ):
        return True
    if re.search(r"(?:[a-zа-яё]{1,3}\.\s*)+[a-zа-яё]$", prefix.lower()):
        return True
    if re.search(r"(?:https?://|www\.)\S*$", normalized):
        return True
    if re.fullmatch(r"\d+(?:[.,]\d+)*", normalized):
        return True
    if re.fullmatch(r"(?:[A-Za-zА-Яа-яЁё]\.){1,4}[A-Za-zА-Яа-яЁё]?", normalized):
        return True
    if normalized in _ABBREVIATIONS:
        return True

    short_prefix = prefix[-12:].lower()
    return any(short_prefix.endswith(value) for value in _ABBREVIATIONS)


def _pack_units(
    text: str,
    units: list[tuple[int, int]],
    block: _Block,
    target_tokens: int,
    max_tokens: int,
    params: ChunkingParams,
) -> list[Chunk]:
    chunks: list[Chunk] = []
    buffer: list[tuple[tuple[int, int], int]] = []
    buffered_tokens = 0

    def flush() -> None:
        nonlocal buffered_tokens
        if not buffer:
            return
        start, end = buffer[0][0][0], buffer[-1][0][1]
        actual_tokens = params.token_counter(text[start:end])
        if actual_tokens <= max_tokens:
            chunks.append(
                _make_chunk(text, start, end, block, params, forced_split=False)
            )
        else:
            for (unit_start, unit_end), unit_tokens in buffer:
                if unit_tokens <= max_tokens:
                    chunks.append(
                        _make_chunk(
                            text, unit_start, unit_end, block, params, forced_split=False
                        )
                    )
                else:
                    chunks.extend(
                        _force_split(text, unit_start, unit_end, block, params, max_tokens)
                    )
        buffer.clear()
        buffered_tokens = 0

    for start, end in units:
        token_count = params.token_counter(text[start:end])
        if token_count > max_tokens:
            flush()
            chunks.extend(_force_split(text, start, end, block, params, max_tokens))
            continue
        if buffer and buffered_tokens + token_count > target_tokens:
            flush()
        buffer.append(((start, end), token_count))
        buffered_tokens += token_count
    flush()
    return chunks


def _make_chunk(
    text: str,
    start: int,
    end: int,
    block: _Block,
    params: ChunkingParams,
    forced_split: bool,
) -> Chunk:
    start, end = _trim_span(text, start, end)
    return Chunk(
        idx=0,
        text=text[start:end],
        heading_path=block.heading_path.copy(),
        char_span=(start, end),
        token_count=params.token_counter(text[start:end]),
        kind=block.kind,
        forced_split=forced_split,
    )


def _force_split(
    text: str,
    start: int,
    end: int,
    block: _Block,
    params: ChunkingParams,
    max_tokens: int,
) -> list[Chunk]:
    pieces = _split_to_fit(text, start, end, block.kind, params, max_tokens)
    packed: list[tuple[int, int]] = []
    buffer: list[tuple[int, int]] = []
    buffered_tokens = 0

    def flush() -> None:
        nonlocal buffered_tokens
        if not buffer:
            return
        packed.append((buffer[0][0], buffer[-1][1]))
        buffer.clear()
        buffered_tokens = 0

    for piece_start, piece_end in pieces:
        token_count = params.token_counter(text[piece_start:piece_end])
        if buffer and buffered_tokens + token_count > max_tokens:
            flush()
        buffer.append((piece_start, piece_end))
        buffered_tokens += token_count
    flush()

    return [
        _make_chunk(text, piece_start, piece_end, block, params, forced_split=True)
        for piece_start, piece_end in packed
    ]


def _split_to_fit(
    text: str,
    start: int,
    end: int,
    kind: ChunkKind,
    params: ChunkingParams,
    max_tokens: int,
) -> list[tuple[int, int]]:
    if params.token_counter(text[start:end]) <= max_tokens:
        return [(start, end)]

    candidates = _line_spans(text, start, end) if kind in {"code", "table"} else []
    if len(candidates) <= 1:
        candidates = _clause_spans(text, start, end)
    if len(candidates) <= 1:
        candidates = [m.span() for m in _WORD_RE.finditer(text, start, end)]

    if len(candidates) > 1:
        pieces: list[tuple[int, int]] = []
        for candidate_start, candidate_end in candidates:
            pieces.extend(
                _split_to_fit(
                    text,
                    candidate_start,
                    candidate_end,
                    kind,
                    params,
                    max_tokens,
                )
            )
        return pieces

    return _char_spans(text, start, end, params, max_tokens)


def _clause_spans(text: str, start: int, end: int) -> list[tuple[int, int]]:
    spans: list[tuple[int, int]] = []
    clause_start = start
    i = start
    while i < end:
        if text[i] not in ";:,":
            i += 1
            continue
        clause_end = i + 1
        while clause_end < end and text[clause_end].isspace():
            clause_end += 1
        span = _trim_span(text, clause_start, clause_end)
        if span[0] < span[1]:
            spans.append(span)
        clause_start = clause_end
        i = clause_end
    tail = _trim_span(text, clause_start, end)
    if tail[0] < tail[1]:
        spans.append(tail)
    return spans


def _char_spans(
    text: str,
    start: int,
    end: int,
    params: ChunkingParams,
    max_tokens: int,
) -> list[tuple[int, int]]:
    spans: list[tuple[int, int]] = []
    cursor = start
    while cursor < end:
        low = cursor + 1
        high = end
        best = cursor + 1
        while low <= high:
            middle = (low + high) // 2
            if params.token_counter(text[cursor:middle]) <= max_tokens:
                best = middle
                low = middle + 1
            else:
                high = middle - 1
        span = _trim_span(text, cursor, best)
        if span[0] >= span[1]:
            span = (cursor, min(cursor + 1, end))
        spans.append(span)
        cursor = span[1]
        while cursor < end and text[cursor].isspace():
            next_start = cursor
            while next_start < end and text[next_start].isspace():
                next_start += 1
            if next_start > cursor:
                cursor = next_start
                break
    return spans
