import random
import re
import sys
import os

import numpy as np
import pytest

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from app.core.chunking import (
    ChunkingParams,
    aggregate,
    chunk,
    embedding_inputs,
)


def word_tokens(text: str) -> int:
    return len(re.findall(r"[\w]+", text, flags=re.UNICODE))


def params(target: int = 8, max_tokens: int = 12, title: str | None = None):
    return ChunkingParams(
        token_counter=word_tokens,
        target_tokens=target,
        max_tokens=max_tokens,
        title=title,
        title_injection=True,
    )


def assert_exact_spans(text: str, chunks):
    cursor = -1
    for item in chunks:
        start, end = item.char_span
        assert start >= 0
        assert end <= len(text)
        assert start < end
        assert start >= cursor
        assert text[start:end] == item.text
        cursor = end


def assert_no_non_whitespace_loss(text: str, chunks):
    assert_exact_spans(text, chunks)
    source = "".join(text.split())
    emitted = "".join("".join(item.text for item in chunks).split())
    assert emitted == source


class TestChunking:
    def test_window_invariant_for_long_runs_code_and_table(self):
        text = (
            " ".join(f"word{i}" for i in range(40))
            + "\n\n```python\n"
            + "code " * 30
            + "\n```\n\n"
            + "| h1 | h2 |\n|---|---|\n"
            + "| " + "cell " * 20 + "|\n"
            + "| tail | row |"
        )
        chunks = chunk(text, params(target=6, max_tokens=8))

        assert chunks
        assert all(item.token_count <= 8 for item in chunks)
        assert any(item.forced_split for item in chunks)
        assert {item.kind for item in chunks} >= {"prose", "code", "table"}
        assert_no_non_whitespace_loss(text, chunks)

    @pytest.mark.parametrize(
        "fragment",
        [
            "Это т.е. проверка",
            "Это т. е. проверка",
            "Один др. вариант",
            "For example e.g. this",
            "Pi is 3.14 exactly",
            "Pushkin А.С. wrote",
            "See https://example.com/path. today",
        ],
    )
    def test_false_sentence_boundaries_do_not_split(self, fragment):
        text = f"{fragment}. Final sentence."
        chunks = chunk(text, params(target=1, max_tokens=60))

        assert chunks
        assert fragment in chunks[0].text
        assert "Final sentence." not in chunks[0].text
        assert_no_non_whitespace_loss(text, chunks)

    def test_random_texts_preserve_non_whitespace_and_spans(self):
        rng = random.Random(42)
        alphabet = "abcde fg."
        for _ in range(50):
            text = "".join(rng.choice(alphabet) for _ in range(rng.randint(1, 300)))
            chunks = chunk(text, params(target=7, max_tokens=11))
            assert_no_non_whitespace_loss(text, chunks)

    def test_structured_text_preserves_all_blocks_and_spans(self):
        text = (
            "# Root\n\nIntro sentence.\n\n## Child\n"
            "Body one. Body two.\n\n```\ncode line\n```\n\n"
            "| a |\n|---|\n| b |"
        )
        chunks = chunk(text, params(target=50, max_tokens=60))

        assert_no_non_whitespace_loss(text, chunks)
        prose = [item for item in chunks if item.kind == "prose"]
        assert prose[-1].heading_path == ["Root", "Child"]

    def test_nested_heading_path(self):
        text = "# Root\n\n## Child\n\n### Leaf\n\nDeep text."
        chunks = chunk(text, params(target=50, max_tokens=60))
        prose = [item for item in chunks if item.kind == "prose"]

        assert prose[-1].heading_path == ["Root", "Child", "Leaf"]

    def test_sibling_and_level_up_heading_paths(self):
        text = (
            "# Guide\n\n## Install\n\nInstall text.\n\n"
            "## Usage\n\nUsage text.\n\n"
            "### Advanced\n\nAdvanced text.\n\n"
            "# Appendix\n\nAppendix text."
        )
        chunks = chunk(text, params(target=50, max_tokens=60))
        prose = [item for item in chunks if item.kind == "prose"]

        by_text = {item.text.split("\n")[-1]: item.heading_path for item in prose}
        assert by_text["Install text."] == ["Guide", "Install"]
        assert by_text["Usage text."] == ["Guide", "Usage"]
        assert by_text["Advanced text."] == ["Guide", "Usage", "Advanced"]
        assert by_text["Appendix text."] == ["Appendix"]

    def test_heading_link_stub_is_no_content(self):
        text = "## [Imported page](https://example.com/page)"
        assert chunk(text, params()) == []

    def test_empty_and_whitespace_text(self):
        assert chunk("", params()) == []
        assert chunk("   \n\n  ", params()) == []

    def test_one_token_text(self):
        chunks = chunk("word", params())
        assert len(chunks) == 1
        assert chunks[0].text == "word"
        assert chunks[0].token_count == 1

    def test_exact_max_boundary_no_split(self):
        text = "one two three four five six seven eight"
        chunks = chunk(text, params(target=8, max_tokens=8))
        assert len(chunks) == 1
        assert chunks[0].forced_split is False
        assert chunks[0].token_count == 8

    def test_long_unbreakable_token_forces_char_split(self):
        char_params = ChunkingParams(
            token_counter=len, target_tokens=5, max_tokens=10
        )
        text = "x" * 200
        chunks = chunk(text, char_params)
        assert len(chunks) > 1
        assert all(item.forced_split for item in chunks)
        assert all(item.token_count <= 10 for item in chunks)
        assert_no_non_whitespace_loss(text, chunks)

    def test_title_injection_adds_title_to_model_input_not_source_span(self):
        text = "First sentence. Second sentence. Third sentence."
        chunk_params = params(target=5, max_tokens=6, title="Note title")
        chunks = chunk(text, chunk_params)
        inputs = embedding_inputs(chunks, chunk_params)

        assert len(chunks) > 1
        assert all(item.text.startswith("First") or "title" not in item.text.lower() for item in chunks)
        assert all(value.startswith("Note title\n\n") for value in inputs)
        assert all(word_tokens(value) <= 6 for value in inputs)

    def test_title_overhead_is_reserved_from_chunk_budget(self):
        text = "one two three four five six"
        chunk_params = params(target=6, max_tokens=6, title="seven eight")
        chunks = chunk(text, chunk_params)
        inputs = embedding_inputs(chunks, chunk_params)

        assert all(word_tokens(value) <= 6 for value in inputs)
        assert_no_non_whitespace_loss(text, chunks)


class TestAggregate:
    def test_mean_with_l2_normalization(self):
        result = aggregate(np.asarray([[1.0, 0.0], [0.0, 1.0]], dtype=np.float32))

        assert result == pytest.approx([2**-0.5, 2**-0.5])
        assert float(np.linalg.norm(result)) == pytest.approx(1.0)

    def test_empty_vectors(self):
        result = aggregate([])

        assert result is None
