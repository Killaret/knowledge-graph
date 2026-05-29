package com.alximac.knowledgegraph.texthandler.infrastructure.chunking;

import com.alximac.knowledgegraph.texthandler.domain.model.DocumentChunk;
import com.alximac.knowledgegraph.texthandler.domain.model.ImportOptions;
import com.alximac.knowledgegraph.texthandler.domain.service.ChunkingStrategy;

import java.text.BreakIterator;
import java.util.*;
import java.util.concurrent.atomic.AtomicInteger;

public class HybridChunker implements ChunkingStrategy {

    @Override
    public List<DocumentChunk> chunk(String text, ImportOptions options) {
        AtomicInteger chunkIndex = new AtomicInteger(0);
        List<DocumentChunk> result = new ArrayList<>();
        List<String> paragraphs = splitIntoParagraphs(text);

        for (String paragraph : paragraphs) {
            if (countWords(paragraph) <= options.chunkSize()) {
                // короткий параграф — сразу чанк, если не короче minChunkLength
                if (paragraph.length() >= options.minChunkLength()) {
                    result.add(new DocumentChunk(paragraph, chunkIndex.getAndIncrement(), Map.of(), null));
                } else if (!result.isEmpty()) {
                    // слишком короткий — присоединяем к предыдущему чанку
                    DocumentChunk prev = result.remove(result.size() - 1);
                    result.add(new DocumentChunk(prev.text() + " " + paragraph, prev.index(), Map.of(), null));
                }
            } else {
                // длинный параграф — sliding window по предложениям
                List<String> sentences = splitIntoSentences(paragraph);
                result.addAll(chunkParagraph(sentences, options, chunkIndex));
            }
        }
        return result;
    }

    // разбивает текст по 2+ переводам строки
    private List<String> splitIntoParagraphs(String text) {
        return List.of(text.split("\\n\\s*\\n"));
    }

    // количество слов (непустых токенов)
    private int countWords(String s) {
        if (s.isBlank()) return 0;
        return s.trim().split("\\s+").length;
    }

    // разбитие текста на предложения
    private List<String> splitIntoSentences(String text) {
        List<String> sentences = new ArrayList<>();
        BreakIterator bi = BreakIterator.getSentenceInstance(Locale.getDefault());
        bi.setText(text);
        int start = bi.first();
        for (int end = bi.next(); end != BreakIterator.DONE; start = end, end = bi.next()) {
            sentences.add(text.substring(start, end).trim());
        }
        return sentences;
    }

    // sliding window для списка предложений
    private List<DocumentChunk> chunkParagraph(List<String> sentences, ImportOptions opts, AtomicInteger indexCounter) {
        List<DocumentChunk> chunks = new ArrayList<>();
        int chunkSizeWords = opts.chunkSize();
        int overlapWords = opts.overlap();
        int minLength = opts.minChunkLength();

        LinkedList<String> wordBuffer = new LinkedList<>(); // текущий буфер слов
        int currentWordCount = 0;

        for (String sentence : sentences) {
            String[] words = sentence.split("\\s+");
            for (String w : words) {
                if (!w.isBlank()) {
                    wordBuffer.add(w);
                    currentWordCount++;
                }
            }

            // когда буфер наполнился или предложение закончилось, формируем чанк(и)
            while (currentWordCount >= chunkSizeWords) {
                // берём chunkSizeWords из буфера
                List<String> chunkWords = new ArrayList<>();
                for (int i = 0; i < chunkSizeWords && !wordBuffer.isEmpty(); i++) {
                    chunkWords.add(wordBuffer.pollFirst());
                    currentWordCount--;
                }
                String chunkText = String.join(" ", chunkWords);
                if (chunkText.length() >= minLength) {
                    chunks.add(new DocumentChunk(chunkText, indexCounter.getAndIncrement(), Map.of(), null));
                } else if (!chunks.isEmpty()) {
                    // короткий чанк присоединяем к предыдущему
                    DocumentChunk prev = chunks.remove(chunks.size() - 1);
                    chunks.add(new DocumentChunk(prev.text() + " " + chunkText, prev.index(), Map.of(), null));
                }

                // Возвращаем overlap слов обратно в буфер для перекрытия
                if (overlapWords > 0 && !chunkWords.isEmpty()) {
                    // Берём последние overlap слов из chunkWords
                    int overlapStart = Math.max(0, chunkWords.size() - overlapWords);
                    List<String> overlapList = chunkWords.subList(overlapStart, chunkWords.size());
                    wordBuffer.addAll(0, overlapList);
                    currentWordCount += overlapList.size();
                }
            }
        }

        // остаток буфера (меньше chunkSize) — последний чанк
        if (!wordBuffer.isEmpty()) {
            String lastText = String.join(" ", wordBuffer);
            if (lastText.length() >= minLength) {
                chunks.add(new DocumentChunk(lastText, indexCounter.getAndIncrement(), Map.of(), null));
            } else if (!chunks.isEmpty()) {
                DocumentChunk prev = chunks.remove(chunks.size() - 1);
                chunks.add(new DocumentChunk(prev.text() + " " + lastText, prev.index(), Map.of(), null));
            }
        }

        return chunks;
    }
}