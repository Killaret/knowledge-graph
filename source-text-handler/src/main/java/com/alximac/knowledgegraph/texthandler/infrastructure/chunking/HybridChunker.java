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
                // short paragraph → immediate chunk, if not shorter than minChunkLength
                if (paragraph.length() >= options.minChunkLength()) {
                    result.add(new DocumentChunk(paragraph, chunkIndex.getAndIncrement(), Map.of(), null));
                } else if (!result.isEmpty()) {
                    // too short → merge with previous chunk
                    DocumentChunk prev = result.remove(result.size() - 1);
                    result.add(new DocumentChunk(prev.text() + " " + paragraph, prev.index(), Map.of(), null));
                }
            } else {
                // long paragraph → sliding window by sentences
                List<String> sentences = splitIntoSentences(paragraph);
                result.addAll(chunkParagraph(sentences, options, chunkIndex));
            }
        }
        return result;
    }

    // splits text by 2+ newlines
    private List<String> splitIntoParagraphs(String text) {
        return List.of(text.split("\\n\\s*\\n"));
    }

    // word count (non-empty tokens)
    private int countWords(String s) {
        if (s.isBlank()) return 0;
        return s.trim().split("\\s+").length;
    }

    // text splitting into sentences
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

    // sliding window for a list of sentences
    private List<DocumentChunk> chunkParagraph(List<String> sentences, ImportOptions opts, AtomicInteger indexCounter) {
        List<DocumentChunk> chunks = new ArrayList<>();
        int chunkSizeWords = opts.chunkSize();
        int overlapWords = opts.overlap();
        int minLength = opts.minChunkLength();

        LinkedList<String> wordBuffer = new LinkedList<>(); // current word buffer
        int currentWordCount = 0;

        for (String sentence : sentences) {
            String[] words = sentence.split("\\s+");
            for (String w : words) {
                if (!w.isBlank()) {
                    wordBuffer.add(w);
                    currentWordCount++;
                }
            }

            // when buffer is full or sentence ends, form chunk(s)
            while (currentWordCount >= chunkSizeWords) {
                // take chunkSizeWords from buffer
                List<String> chunkWords = new ArrayList<>();
                for (int i = 0; i < chunkSizeWords && !wordBuffer.isEmpty(); i++) {
                    chunkWords.add(wordBuffer.pollFirst());
                    currentWordCount--;
                }
                addChunkOrMerge(String.join(" ", chunkWords), minLength, chunks, indexCounter);

                // Return overlap words back to buffer for overlap
                if (overlapWords > 0 && !chunkWords.isEmpty()) {
                    // Take last overlap words from chunkWords
                    int overlapStart = Math.max(0, chunkWords.size() - overlapWords);
                    List<String> overlapList = chunkWords.subList(overlapStart, chunkWords.size());
                    wordBuffer.addAll(0, overlapList);
                    currentWordCount += overlapList.size();
                }
            }
        }

        // remaining buffer (less than chunkSize) → last chunk
        if (!wordBuffer.isEmpty()) {
            String lastText = String.join(" ", wordBuffer);
            addChunkOrMerge(lastText, minLength, chunks, indexCounter);
        }

        return chunks;
    }

    private void addChunkOrMerge(String chunkText, int minLength, List<DocumentChunk> chunks, AtomicInteger indexCounter) {
        if (chunkText.length() >= minLength) {
            chunks.add(new DocumentChunk(chunkText, indexCounter.getAndIncrement(), Map.of(), null));
        } else if (!chunks.isEmpty()) {
            DocumentChunk prev = chunks.remove(chunks.size() - 1);
            chunks.add(new DocumentChunk(prev.text() + " " + chunkText, prev.index(), Map.of(), null));
        }
        // If list is empty and text is short, chunk is lost.
    }
}