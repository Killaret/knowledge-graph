// Package quality implements NOTE-QUALITY-1 stage 1: a weight-free quality
// pass over a note. Signals are computed deterministically from the note text
// (normalized when the NLP-4 artifact exists) plus a few counters from the
// database; the only model-dependent inputs (chunk coherence, title/text
// similarity) are injected by the caller. Nothing here mutates the note —
// stage 1 measures and reports, it never edits.
package quality

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Verdicts — the stage-1 decision format that later becomes the Java contract.
const (
	VerdictCreate = "create"
	VerdictEnrich = "enrich"
	VerdictManual = "manual"
)

// Kind values — coarse shape of the note body.
const (
	KindStub       = "stub"
	KindCollection = "collection"
	KindText       = "text"
)

// Origins.
const (
	OriginImport = "import"
	OriginManual = "manual"
)

// Thresholds are feature boundaries, not measure weights. They live under
// nlp.quality.* in config and are revisited together with the stage-2 weights.
type Thresholds struct {
	CollectionProseShare float64 // prose_share below this + enough link lines => collection
	CollectionMinLinks   int     // minimum link/list lines for a collection
	SentenceMinWords     int     // words in a "real" sentence
	FragmentMaxWords     int     // a non-empty non-heading line this short is a fragment
	MojibakeShare        float64 // share of replacement/mojibake runes that trips the gate
	LegacyTruncatedRunes int     // body after the import line >= this => old 5000-rune cut
}

// DefaultThresholds are the stage-1 numbers from the task spec.
func DefaultThresholds() Thresholds {
	return Thresholds{
		CollectionProseShare: 0.3,
		CollectionMinLinks:   3,
		SentenceMinWords:     4,
		FragmentMaxWords:     3,
		MojibakeShare:        0.01,
		LegacyTruncatedRunes: 4990,
	}
}

// Input carries everything the signal computation cannot derive from the text
// itself — model-derived numbers and database counters are injected so this
// stays pure and testable without a network.
type Input struct {
	Title     string
	Content   string // artifact normalized text when present, raw note body otherwise
	SourceURL string

	// ImportTruncated is metadata.import_truncated written by stage-A import.
	ImportTruncated bool

	// Model/db-derived (computed by the assessor, never by this package).
	CoherenceMin        *float64
	CoherenceMedian     *float64
	TitleTextSimilarity *float64
	Keywords            int
	Links               int
	HasEmbedding        bool
}

// Signals is the stage-1 signal set; one record per assessment, stored in the
// nlp_artifacts quality field and the quality_log collection.
type Signals struct {
	Words               int      `json:"words" bson:"words"`
	ProseWords          int      `json:"prose_words" bson:"prose_words"`
	EndsWithSentence    bool     `json:"ends_with_sentence" bson:"ends_with_sentence"`
	TruncatedByImport   bool     `json:"truncated_by_import" bson:"truncated_by_import"`
	UnclosedFence       bool     `json:"unclosed_fence" bson:"unclosed_fence"`
	Headings            int      `json:"headings" bson:"headings"`
	CoherenceMin        *float64 `json:"coherence_min" bson:"coherence_min,omitempty"`
	CoherenceMedian     *float64 `json:"coherence_median" bson:"coherence_median,omitempty"`
	Sentences           int      `json:"sentences" bson:"sentences"`
	ProseShare          float64  `json:"prose_share" bson:"prose_share"`
	Kind                string   `json:"kind" bson:"kind"`
	FragmentShare       float64  `json:"fragment_share" bson:"fragment_share"`
	MaxBlockWords       int      `json:"max_block_words" bson:"max_block_words"`
	HasSourceLink       bool     `json:"has_source_link" bson:"has_source_link"`
	TitleWords          int      `json:"title_words" bson:"title_words"`
	TitleGeneric        bool     `json:"title_generic" bson:"title_generic"`
	TitleTextSimilarity *float64 `json:"title_text_similarity" bson:"title_text_similarity,omitempty"`
	Origin              string   `json:"origin" bson:"origin"`
	Keywords            int      `json:"keywords" bson:"keywords"`
	Links               int      `json:"links" bson:"links"`
	HasEmbedding        bool     `json:"has_embedding" bson:"has_embedding"`
	Mojibake            bool     `json:"mojibake" bson:"mojibake"`
}

var (
	wordRe     = regexp.MustCompile(`[\p{L}\p{N}]+`)
	mdLinkRe   = regexp.MustCompile(`!?\[([^\]]*)\]\(([^)\s]+)[^)]*\)`)
	bareURLRe  = regexp.MustCompile(`(?i)\bhttps?://\S+|www\.\S+`)
	headingRe  = regexp.MustCompile(`^#{1,6}\s`)
	listItemRe = regexp.MustCompile(`^\s*(?:[-*+]|\d+[.)])\s+`)
	fenceRe    = regexp.MustCompile("^\\s*```")
	tableRe    = regexp.MustCompile(`^\s*\|`)
	// First-line import marker. Greedy `.*` tolerates inner brackets inside
	// the link title (e.g. "[423 Be Supposed To Exercises [21 Online Tests]]").
	importFirstRe = regexp.MustCompile(`^##\s+\[.*\]\(https?://`)
	blockquoteRe  = regexp.MustCompile(`^\s*>`)
	// A "link-only" line is just a link, optionally under a heading/bullet
	// marker — the anchor text does not count (stub detection).
	soleMdLinkRe  = regexp.MustCompile(`^\s*(?:#{1,6}\s+)?(?:[-*+]\s+|\d+[.)]\s+)?!?\[[^\]]*\]\([^)]*\)\s*$`)
	soleBareURLRe = regexp.MustCompile(`^\s*(?:#{1,6}\s+)?(?:[-*+]\s+|\d+[.)]\s+)?(?:https?://\S+|www\.\S+)\s*$`)
)

var genericTitles = map[string]bool{
	"home": true, "главная": true, "untitled": true, "index": true,
	"welcome": true, "без названия": true, "новая заметка": true,
	"new note": true, "page": true, "страница": true,
}

// stripInlineMarkup removes markdown link syntax and bare URLs so word counts
// see words, not addresses.
func stripInlineMarkup(s string) string {
	s = mdLinkRe.ReplaceAllString(s, "$1")
	s = bareURLRe.ReplaceAllString(s, " ")
	return s
}

func countWords(s string) int {
	return len(wordRe.FindAllString(stripInlineMarkup(s), -1))
}

// lineFacts classifies one line for the line-level signals.
type lineFacts struct {
	empty    bool
	heading  bool
	listItem bool
	fence    bool
	table    bool
	quote    bool
	linkOnly bool // nothing left after stripping links/markup
	words    int
}

func analyzeLine(line string) lineFacts {
	trimmed := strings.TrimSpace(line)
	f := lineFacts{empty: trimmed == ""}
	if f.empty {
		return f
	}
	f.fence = fenceRe.MatchString(line)
	f.heading = headingRe.MatchString(line)
	f.listItem = listItemRe.MatchString(line)
	f.table = tableRe.MatchString(line)
	f.quote = blockquoteRe.MatchString(line)
	// linkOnly: the whole line is a link (marker + link + nothing else).
	f.linkOnly = soleMdLinkRe.MatchString(line) || soleBareURLRe.MatchString(line)
	f.words = countWords(line)
	return f
}

func countWordsFromMarkupStripped(s string) int {
	return len(wordRe.FindAllString(stripInlineMarkup(s), -1))
}

var sentenceEndRe = regexp.MustCompile(`[.!?…]+["'»)\]]*\s*`)

func splitSentences(s string) []string {
	parts := sentenceEndRe.Split(s, -1)
	out := parts[:0]
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			out = append(out, p)
		}
	}
	return out
}

var mojibakeRe = regexp.MustCompile(`[ÐÃÑð¢]`)

// ComputeSignals evaluates every stage-1 signal. Pure: no IO, no model calls.
func ComputeSignals(in Input, th Thresholds) Signals {
	var s Signals

	content := in.Content
	lines := strings.Split(content, "\n")

	isImport := in.SourceURL != "" ||
		(len(lines) > 0 && importFirstRe.MatchString(strings.TrimSpace(lines[0])))
	if isImport {
		s.Origin = OriginImport
	} else {
		s.Origin = OriginManual
	}
	s.HasSourceLink = in.SourceURL != "" ||
		(len(lines) > 0 && mdLinkRe.MatchString(lines[0]))

	inFence := false
	linkishLines := 0
	nonEmptyLines := 0
	linkOnlyLines := 0
	fragments := 0
	blockWords, maxBlock := 0, 0
	var lastNonEmpty string

	for _, line := range lines {
		f := analyzeLine(line)
		if f.fence {
			inFence = !inFence
			continue
		}
		if inFence {
			continue // code is neither prose nor a fragment
		}
		if f.empty {
			if blockWords > maxBlock {
				maxBlock = blockWords
			}
			blockWords = 0
			continue
		}
		nonEmptyLines++
		blockWords += f.words
		if f.table || f.listItem || f.linkOnly {
			linkishLines++
		}
		if f.linkOnly {
			linkOnlyLines++
		}
		if f.heading {
			s.Headings++
		} else {
			if f.words <= th.FragmentMaxWords {
				fragments++
			}
			if !f.listItem && !f.linkOnly && !f.table {
				s.ProseWords += f.words
				for _, sent := range splitSentences(stripSentenceLine(line)) {
					if countWords(sent) >= th.SentenceMinWords {
						s.Sentences++
					}
				}
			}
		}
		lastNonEmpty = line
	}
	if blockWords > maxBlock {
		maxBlock = blockWords
	}
	s.MaxBlockWords = maxBlock
	s.UnclosedFence = inFence

	if nonEmptyLines > 0 {
		s.FragmentShare = float64(fragments) / float64(nonEmptyLines)
	}
	s.Words = countWords(content)
	if s.Words > 0 {
		s.ProseShare = float64(s.ProseWords) / float64(s.Words)
	}

	// Kind: stub (only a link line), collection (mostly link/bullet lines),
	// else text.
	body := content
	if isImport && len(lines) > 0 {
		body = strings.Join(lines[1:], "\n")
	}
	bodyFacts := countNonLinkLines(body)
	if (s.Words == 0 && bodyFacts == 0) || (nonEmptyLines == 1 && linkOnlyLines == 1) {
		s.Kind = KindStub
	} else if s.ProseShare < th.CollectionProseShare && linkishLines >= th.CollectionMinLinks {
		s.Kind = KindCollection
	} else {
		s.Kind = KindText
	}

	// ends_with_sentence: last non-empty line ends with sentence punctuation
	// or closing markup (e.g. a list item may end with a link closing bracket
	// right after the period). Link markup is stripped first so a trailing
	// [text.](url) still counts.
	tail := strings.TrimRight(strings.TrimSpace(stripInlineMarkup(lastNonEmpty)), `"'»)\]*_`+"`")
	s.EndsWithSentence = strings.HasSuffix(tail, ".") ||
		strings.HasSuffix(tail, "!") || strings.HasSuffix(tail, "?") ||
		strings.HasSuffix(tail, "…")

	// truncated_by_import: explicit flag from stage-A import, or the legacy
	// heuristic — whole content sits at/over the old 5000-rune cap.
	s.TruncatedByImport = in.ImportTruncated ||
		(isImport && utf8.RuneCountInString(content) >= th.LegacyTruncatedRunes)

	// Title signals.
	title := strings.TrimSpace(in.Title)
	s.TitleWords = countWords(title)
	s.TitleGeneric = s.TitleWords <= 1 ||
		bareURLRe.MatchString(title) || genericTitles[strings.ToLower(title)]

	// Mojibake: replacement chars and the classic Cyrillic-in-1252 leftovers.
	totalRunes := utf8.RuneCountInString(content)
	if totalRunes > 0 {
		bad := len(mojibakeRe.FindAllString(content, -1))
		s.Mojibake = float64(bad)/float64(totalRunes) > th.MojibakeShare
	}

	// Injected fields pass through.
	s.CoherenceMin = in.CoherenceMin
	s.CoherenceMedian = in.CoherenceMedian
	s.TitleTextSimilarity = in.TitleTextSimilarity
	s.Keywords = in.Keywords
	s.Links = in.Links
	s.HasEmbedding = in.HasEmbedding

	return s
}

// countNonLinkLines counts lines that still carry words after link markup is
// stripped — a stub note leaves nothing behind.
func countNonLinkLines(body string) int {
	n := 0
	inFence := false
	for _, line := range strings.Split(body, "\n") {
		if fenceRe.MatchString(line) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if countWordsFromMarkupStripped(line) > 0 {
			n++
		}
	}
	return n
}

// stripSentenceLine kept for sentence extraction on prose lines.
func stripSentenceLine(line string) string {
	return strings.TrimSpace(stripInlineMarkup(line))
}
