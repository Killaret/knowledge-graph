package com.alximac.knowledgegraph.texthandler.infrastructure.chunking;

import com.alximac.knowledgegraph.texthandler.domain.model.DocumentChunk;
import com.alximac.knowledgegraph.texthandler.domain.model.ImportOptions;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.*;

class HybridChunkerTest {

    private final HybridChunker chunker = new HybridChunker();

    private ImportOptions opts(int chunkSize, int overlap, int minChunkLength) {
        return new ImportOptions(chunkSize, overlap, minChunkLength, 3, false);
    }

    @Test
    @DisplayName("Must return empty list for blank text")
    void mustReturnEmptyForBlankText() {
        List<DocumentChunk> result = chunker.chunk("", opts(100, 10, 50));
        assertThat(result).isEmpty();

        result = chunker.chunk("   \n\n  ", opts(100, 10, 50));
        assertThat(result).isEmpty();
    }

    @Test
    @DisplayName("Must create single chunk for short paragraph longer than minChunkLength")
    void mustCreateSingleChunkForParagraphLongerThanMin() {
        // Текст длиной 60 символов
        String text = "This is a paragraph that is long enough to exceed fifty characters.";
        ImportOptions options = opts(100, 10, 50);
        List<DocumentChunk> chunks = chunker.chunk(text, options);

        assertThat(chunks).hasSize(1);
        DocumentChunk chunk = chunks.get(0);
        assertThat(chunk.text()).isEqualTo(text);
        assertThat(chunk.index()).isEqualTo(0);
    }

    @Test
    @DisplayName("Must create multiple chunks for multiple long paragraphs")
    void mustCreateMultipleChunksForParagraphs() {
        String text = "First paragraph is long enough to pass the minimum length threshold easily.\n\n" +
                "Second paragraph also contains sufficient amount of text for testing.\n\n" +
                "Third paragraph is here just to have three chunks in the result list.";
        ImportOptions options = opts(100, 10, 50);
        List<DocumentChunk> chunks = chunker.chunk(text, options);

        assertThat(chunks).hasSize(3);
        assertThat(chunks.get(0).index()).isEqualTo(0);
        assertThat(chunks.get(1).index()).isEqualTo(1);
        assertThat(chunks.get(2).index()).isEqualTo(2);
    }

    @Test
    @DisplayName("Must split long paragraph with sliding window")
    void mustSplitLongParagraph() {
        // 70 слов , chunkSize=50, minChunkLength=50
        String text = "longword ".repeat(70).trim();
        ImportOptions options = opts(50, 0, 50);
        List<DocumentChunk> chunks = chunker.chunk(text, options);

        // 70 слов при chunkSize=50 даст 2 чанка (50 + 20), каждый длиной >50 символов
        assertThat(chunks).hasSize(2);
        assertThat(chunks.get(0).text().split("\\s+")).hasSize(50);
        assertThat(chunks.get(1).text().split("\\s+")).hasSize(20);
    }

    @Test
    @DisplayName("Must apply overlap correctly")
    void mustApplyOverlap() {
        // 110 слов "longword" от "longword1" до "longword110", chunkSize=50, overlap=10, minChunkLength=50
        StringBuilder sb = new StringBuilder();
        for (int i = 1; i <= 110; i++) sb.append("longword").append(i).append(" ");
        String text = sb.toString().trim();
        ImportOptions options = opts(50, 10, 50);
        List<DocumentChunk> chunks = chunker.chunk(text, options);

        // Ожидаем 3 чанка: 1..50, 41..90, 81..110
        assertThat(chunks).hasSize(3);

        // Проверяем границы и порядок
        assertThat(chunks.get(0).text()).startsWith("longword1").endsWith("longword50");
        assertThat(chunks.get(1).text()).startsWith("longword41").endsWith("longword90");
        assertThat(chunks.get(2).text()).startsWith("longword81").endsWith("longword110");
    }

    @Test
    @DisplayName("Must merge short chunk with previous when below minChunkLength")
    void mustMergeShortChunkWithPreviousWhenExists() {
        // Первый параграф длинный (>50), второй короткий (<50) => присоединяется к первому
        String text = "This paragraph is definitely long enough to exceed the minimum length of fifty characters.\n\nShort.";
        ImportOptions options = opts(100, 10, 50);
        List<DocumentChunk> chunks = chunker.chunk(text, options);

        // Должен остаться один чанк, содержащий оба параграфа
        assertThat(chunks).hasSize(1);
        assertThat(chunks.get(0).text()).contains("This paragraph").contains("Short.");
    }

    @Test
    @DisplayName("Must preserve sequential indexes")
    void mustPreserveSequentialIndexes() {
        String text = "Chunk A is a rather long text that easily passes the minimum character requirement.\n\n" +
                "Chunk B also contains enough words and letters to be considered valid.\n\n" +
                "Chunk C is the final one and it is sufficiently lengthy to create a chunk.";
        ImportOptions options = opts(100, 10, 50);
        List<DocumentChunk> chunks = chunker.chunk(text, options);

        assertThat(chunks).extracting(DocumentChunk::index)
                .containsExactly(0, 1, 2);
    }

    @Test
    @DisplayName("Sliding window: must produce exact text chunks, not just word counts")
    void mustProduceExactChunksForLongParagraph() {
        // 150 длинных слов (по 10 символов), chunkSize=50 даёт 3 чанка (50+50+50)
        StringBuilder sb = new StringBuilder();
        for (int i = 1; i <= 150; i++) {
            sb.append(String.format("Word%04d", i)).append(" ");
        }
        String text = sb.toString().trim();
        ImportOptions options = opts(50, 0, 50);
        List<DocumentChunk> chunks = chunker.chunk(text, options);

        assertThat(chunks).hasSize(3);
        // Проверяем первые слова каждого чанка
        assertThat(chunks.get(0).text()).startsWith("Word0001").endsWith("Word0050");
        assertThat(chunks.get(1).text()).startsWith("Word0051").endsWith("Word0100");
        assertThat(chunks.get(2).text()).startsWith("Word0101").endsWith("Word0150");
    }

    @Test
    @DisplayName("Must correctly handle real sentences with BreakIterator")
    void mustHandleRealSentences() {
        //  предложения из длинных "слов", чтобы 50 слов было достаточно длинным
        String longWord = "ABCDEFGHIJ"; // 10 символов
        String sentence1 = (longWord + " ").repeat(50).trim() + ". ";
        String sentence2 = (longWord + " ").repeat(30).trim() + ". ";
        String sentence3 = (longWord + " ").repeat(20).trim() + ".";
        String text = sentence1 + sentence2 + sentence3;
        ImportOptions options = opts(50, 0, 50);
        List<DocumentChunk> chunks = chunker.chunk(text, options);

        // Ожидаем 2 чанка: первый из 50 слов, второй из 50 слов (30+20)
        assertThat(chunks).hasSize(2);
        assertThat(chunks.get(0).text().split("\\s+")).hasSize(50);
        assertThat(chunks.get(1).text().split("\\s+")).hasSize(50);
        // Проверка, что второе предложение начинается с того же слова
        assertThat(chunks.get(1).text()).contains(longWord);
    }

    @Test
    @DisplayName("Overlap must work correctly with multiple sliding windows")
    void mustApplyOverlapWithMultipleChunks() {
        // 120 длинных слов, chunkSize=50, overlap=10 -> должны получить 3 чанка
        StringBuilder sb = new StringBuilder();
        for (int i = 1; i <= 120; i++) {
            sb.append(String.format("W%03d", i)).append(" ");
        }
        String text = sb.toString().trim();
        ImportOptions options = opts(50, 10, 50);
        List<DocumentChunk> chunks = chunker.chunk(text, options);

        // 120 слов, chunkSize=50, overlap=10:
        // Первый: 1-50, второй: 41-90, третий: 81-120
        assertThat(chunks).hasSize(3);
        assertThat(chunks.get(0).text()).startsWith("W001").endsWith("W050");
        assertThat(chunks.get(1).text()).startsWith("W041").endsWith("W090");
        assertThat(chunks.get(2).text()).startsWith("W081").endsWith("W120");
    }

    @Test
    @DisplayName("Must merge short sliding-window chunk with previous inside long paragraph")
    void mustMergeShortSlidingWindowChunk() {
        // Первая часть: 50 длинных слов (длина >50), остаток: одно короткое слово "xy"
        String longWord = "LongEnoughWord"; // 14 символов
        String text = (longWord + " ").repeat(50).trim() + " xy";
        ImportOptions options = opts(50, 0, 50);
        List<DocumentChunk> chunks = chunker.chunk(text, options);

        // Ожидаем 1 чанк, т.к. " xy" короче 50 символов и присоединяется
        assertThat(chunks).hasSize(1);
        assertThat(chunks.get(0).text()).endsWith("xy");
        assertThat(chunks.get(0).index()).isEqualTo(0);
    }

    @Test
    @DisplayName("Must create chunk when length equals minChunkLength exactly")
    void mustCreateChunkWhenExactlyMinLength() {
        String text = "0".repeat(50); // 50 символов
        ImportOptions options = opts(100, 10, 50);
        List<DocumentChunk> chunks = chunker.chunk(text, options);

        assertThat(chunks).hasSize(1);
        assertThat(chunks.get(0).text()).isEqualTo(text);
    }

    @Test
    @DisplayName("Multiple short paragraphs in a row are all lost (normal behaviour)")
    void multipleShortParagraphsLost() {
        String text = "Tiny.\n\nSmall.\n\nShort."; // все <50 символов
        ImportOptions options = opts(100, 10, 50);
        List<DocumentChunk> chunks = chunker.chunk(text, options);

        assertThat(chunks).isEmpty();
    }
}