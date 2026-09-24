package com.alximac.knowledgegraph.texthandler.infrastructure.validation;
import java.text.BreakIterator;
import java.util.*;
import java.util.regex.Pattern;

public class ChunkValidator {

    private static final Set<String> STOP_WORDS = Set.of(
            "и", "в", "на", "с", "а", "но", "или", "что", "как", "то", "это",
            "the", "is", "a", "an", "and", "or", "but", "this", "that", "to"
    );

    private static final int MIN_CHUNK_LENGTH = 50;
    private static final double MIN_TOPIC_OVERLAP = 0.2;

    public static ValidationResult validate(String chunk) {
        List<String> errors = new ArrayList<>();

        // Правило 1: Минимальная длина
        if (chunk.length() < MIN_CHUNK_LENGTH) {
            errors.add("Chunk too short: " + chunk.length() + " chars");
        }

        // Правило 2: Целостность предложений
        if (!hasCompleteSentences(chunk)) {
            errors.add("Incomplete sentences detected");
        }

        // Правило 3: Отсутствие обрывов слов
        if (hasBrokenWords(chunk)) {
            errors.add("Broken words detected (hyphenation issues)");
        }

        // Правило 4: Topic consistency для чанков с несколькими предложениями
        if (getSentenceCount(chunk) > 2 && !hasTopicConsistency(chunk, MIN_TOPIC_OVERLAP)) {
            errors.add("Possible topic shift detected (low semantic overlap)");
        }

        // Правило 5: Отсутствие повторяющихся паттернов
        if (hasRepetitivePatterns(chunk)) {
            errors.add("Repetitive patterns detected");
        }

        return new ValidationResult(errors.isEmpty(), errors);
    }

    private static boolean hasCompleteSentences(String chunk) {
        String trimmed = chunk.trim();
        if (trimmed.length() == 0) return false;

        // Не должен начинаться с маленькой буквы (вероятно, обрыв)
        if (Character.isLowerCase(trimmed.charAt(0))) {
            return false;
        }

        // Должен заканчиваться пунктуацией если > 10 слов
        if (countWords(trimmed) > 10) {
            char lastChar = trimmed.charAt(trimmed.length() - 1);
            return lastChar == '.' || lastChar == '!' || lastChar == '?' || lastChar == '…';
        }

        return true;
    }

    private static boolean hasBrokenWords(String chunk) {
        // Проверяем дефисы в конце строк (обрывы при переносе)
        return chunk.contains("-\n") || chunk.matches(".*\\w-\\s.*");
    }

    private static boolean hasTopicConsistency(String chunk, double minOverlapRatio) {
        List<String> sentences = splitIntoSentences(chunk);
        if (sentences.size() < 2) return true;

        Set<String> significantWords1 = extractSignificantWords(sentences.get(0));
        Set<String> significantWords2 = extractSignificantWords(sentences.get(1));

        if (significantWords1.isEmpty() || significantWords2.isEmpty()) {
            return true; // не можем проверить, считаем ок
        }

        Set<String> intersection = new HashSet<>(significantWords1);
        intersection.retainAll(significantWords2);

        double overlapRatio = (double) intersection.size() /
                Math.max(significantWords1.size(), significantWords2.size());

        return overlapRatio >= minOverlapRatio;
    }

    private static Set<String> extractSignificantWords(String sentence) {
        String[] words = sentence.toLowerCase().split("\\s+");
        Set<String> significant = new HashSet<>();

        for (String word : words) {
            word = word.replaceAll("[^а-яА-Яa-zA-Z]", "");
            if (word.length() > 2 && !STOP_WORDS.contains(word)) {
                significant.add(word);
            }
        }
        return significant;
    }

    private static boolean hasRepetitivePatterns(String chunk) {
        List<String> sentences = splitIntoSentences(chunk);
        if (sentences.size() < 3) return false;

        for (int i = 0; i < sentences.size() - 1; i++) {
            if (sentences.get(i).equalsIgnoreCase(sentences.get(i + 1))) {
                return true;
            }
        }
        return false;
    }

    private static List<String> splitIntoSentences(String text) {
        List<String> sentences = new ArrayList<>();
        BreakIterator bi = BreakIterator.getSentenceInstance(Locale.getDefault());
        bi.setText(text);
        int start = bi.first();
        for (int end = bi.next(); end != BreakIterator.DONE; start = end, end = bi.next()) {
            String sentence = text.substring(start, end).trim();
            if (!sentence.isEmpty()) {
                sentences.add(sentence);
            }
        }
        return sentences;
    }

    private static int countWords(String s) {
        if (s.isBlank()) return 0;
        return s.trim().split("\\s+").length;
    }

    private static int getSentenceCount(String chunk) {
        return splitIntoSentences(chunk).size();
    }

    public record ValidationResult(boolean isValid, List<String> errors) {}
}
