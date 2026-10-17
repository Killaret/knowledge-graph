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
                addChunkOrMerge(paragraph, options.minChunkLength(), result, chunkIndex);
                // короткий параграф — сразу чанк, если не короче minChunkLength
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
                addChunkOrMerge(String.join(" ", chunkWords), minLength, chunks, indexCounter);

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
            addChunkOrMerge(lastText, minLength, chunks, indexCounter);
        }

        return chunks;
    }

    private List<String> sliceOversizedChunk(String text, int maxLength) {
        List<String> parts = new ArrayList<>();
        String[] words = text.split("\\s+");
        StringBuilder current = new StringBuilder();

        for (String word : words) {

            if (word.length() > maxLength) {
                // Слово слишком длинное — режем посимвольно
                if (current.length() > 0) {
                    parts.add(current.toString().strip());
                    current.setLength(0);
                }
                while (word.length() > maxLength) {
                    parts.add(word.substring(0, maxLength));
                    word = word.substring(maxLength);
                }
                current.append(word); // остаток
                continue;
            }
            if (current.length() + word.length() + 1 > maxLength) {
                if (current.length() > 0) {
                    parts.add(current.toString().strip());
                    current.setLength(0);
                }
            }
            if (current.length() > 0) current.append(' ');
            current.append(word);
        }
        if (current.length() > 0) {
            parts.add(current.toString().strip());
        }
        return parts;
    }

    private void addChunkOrMerge(String chunkText, int minLength, List<DocumentChunk> chunks, AtomicInteger indexCounter) {
        // 1. Разрезаем слишком длинный текст на части ≤ MAX_TEXT_LENGTH
        List<String> parts = chunkText.length() <= DocumentChunk.MAX_TEXT_LENGTH
                ? List.of(chunkText)
                : sliceOversizedChunk(chunkText, DocumentChunk.MAX_TEXT_LENGTH);

        for (String part : parts) {
            if (part.length() >= minLength) {
                chunks.add(new DocumentChunk(part, indexCounter.getAndIncrement(), Map.of(), null, null));
            } else if (!chunks.isEmpty()) {
                // Пробуем присоединить к предыдущему
                DocumentChunk prev = chunks.remove(chunks.size() - 1);
                String merged = prev.text() + " " + part;

                // Проверяем, не превысит ли слияние лимит
                if (merged.length() <= DocumentChunk.MAX_TEXT_LENGTH) {
                    chunks.add(new DocumentChunk(merged, prev.index(), Map.of(), null, null));
                } else {
                    // Не влезает — кладём prev обратно и добавляем part как отдельный чанк
                    // Слияние превысит лимит — сохраняю part отдельно, даже если он короче minLength.
                    // Иначе - потеря данных
                    chunks.add(prev);
                    chunks.add(new DocumentChunk(part, indexCounter.getAndIncrement(), Map.of(), null, null));
                }
            }else {
                // chunks пустой, part короче minLength — не выбрасываем его.
                // мб  контент документа маленький.
                chunks.add(new DocumentChunk(part, indexCounter.getAndIncrement(), Map.of(), null, null));
            }
        }
    }


}