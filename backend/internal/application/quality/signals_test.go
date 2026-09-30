package quality

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func baseThresholds() Thresholds { return DefaultThresholds() }

func TestComputeSignals_ProseAndWords(t *testing.T) {
	in := Input{
		Title: "Заметка о тестах",
		Content: "# Заметка о тестах\n\n" +
			"Это первое предложение с достаточным числом слов. " +
			"А это второе предложение, тоже длинное.\n\n" +
			"- [Ссылка раз](https://a.example/1)\n" +
			"- [Ссылка два](https://a.example/2)\n",
	}
	s := ComputeSignals(in, baseThresholds())
	assert.Greater(t, s.Words, 10)
	assert.Equal(t, s.Headings, 1)
	assert.Equal(t, 2, s.Sentences)
	assert.Equal(t, KindText, s.Kind)
	assert.False(t, s.TruncatedByImport)
	assert.False(t, s.UnclosedFence)
	// Last non-empty line is a bare link bullet — not a sentence ending.
	assert.False(t, s.EndsWithSentence)
	assert.Equal(t, OriginManual, s.Origin)
	assert.Greater(t, s.ProseShare, 0.5)
	assert.Equal(t, 3, s.TitleWords)
	assert.False(t, s.TitleGeneric)
}

func TestComputeSignals_Stub(t *testing.T) {
	// Bare-link import: heading line is the source link, nothing else.
	in := Input{
		Title:     "https://example.com/page",
		Content:   "## [Example](https://example.com/page)\n",
		SourceURL: "https://example.com/page",
	}
	s := ComputeSignals(in, baseThresholds())
	assert.Equal(t, KindStub, s.Kind)
	// The anchor text still counts as one word — the line is a link line,
	// which is what makes it a stub regardless.
	assert.Equal(t, 1, s.Words)
	assert.True(t, s.HasSourceLink)
	assert.Equal(t, OriginImport, s.Origin)
}

func TestComputeSignals_Collection(t *testing.T) {
	in := Input{
		Title: "Подборка ссылок",
		Content: "# Подборка\n\n" +
			"- [Один](https://a.example/1) короткая заметка\n" +
			"- [Два](https://a.example/2) короткая заметка\n" +
			"- [Три](https://a.example/3) короткая заметка\n" +
			"- [Четыре](https://a.example/4) короткая заметка\n",
	}
	s := ComputeSignals(in, baseThresholds())
	assert.Equal(t, KindCollection, s.Kind)
	assert.Less(t, s.ProseShare, 0.3)
}

func TestComputeSignals_TruncatedImport_Flag(t *testing.T) {
	in := Input{Title: "x", Content: "текст", ImportTruncated: true}
	s := ComputeSignals(in, baseThresholds())
	assert.True(t, s.TruncatedByImport)
}

func TestComputeSignals_TruncatedImport_Legacy(t *testing.T) {
	// Old imports have no metadata flag — the heuristic catches a body that
	// sits at the old 5000-rune cap after the first link line.
	body := strings.Repeat("а", 4995)
	in := Input{
		Title:   "L",
		Content: "## [t](https://x.example)\n\n" + body,
	}
	s := ComputeSignals(in, baseThresholds())
	assert.True(t, s.TruncatedByImport)

	in.Content = "## [t](https://x.example)\n\n" + strings.Repeat("а", 100)
	s = ComputeSignals(in, baseThresholds())
	assert.False(t, s.TruncatedByImport)
}

func TestComputeSignals_UnclosedFence(t *testing.T) {
	in := Input{Title: "x", Content: "текст\n\n```go\nкод\n"}
	s := ComputeSignals(in, baseThresholds())
	assert.True(t, s.UnclosedFence)

	in.Content = "текст\n\n```go\nкод\n```\n"
	s = ComputeSignals(in, baseThresholds())
	assert.False(t, s.UnclosedFence)
}

func TestComputeSignals_EndsWithSentence(t *testing.T) {
	in := Input{Title: "x", Content: "Первое предложение завершено."}
	assert.True(t, ComputeSignals(in, baseThresholds()).EndsWithSentence)

	in.Content = "Обрезано на середине слов"
	assert.False(t, ComputeSignals(in, baseThresholds()).EndsWithSentence)

	// Closing markup after the period still counts.
	in.Content = "Фраза [ссылкой.](https://x.example)"
	assert.True(t, ComputeSignals(in, baseThresholds()).EndsWithSentence)
}

func TestComputeSignals_Mojibake(t *testing.T) {
	in := Input{Title: "x", Content: strings.Repeat("Ð", 10) + strings.Repeat("а", 20)}
	assert.True(t, ComputeSignals(in, baseThresholds()).Mojibake)

	in.Content = "Чистый русский текст без поломок."
	assert.False(t, ComputeSignals(in, baseThresholds()).Mojibake)
}

func TestComputeSignals_TitleGeneric(t *testing.T) {
	in := Input{Title: "Untitled", Content: "текст"}
	assert.True(t, ComputeSignals(in, baseThresholds()).TitleGeneric)

	in.Title = "https://example.com/x"
	assert.True(t, ComputeSignals(in, baseThresholds()).TitleGeneric)

	in.Title = "Осмысленный заголовок заметки"
	assert.False(t, ComputeSignals(in, baseThresholds()).TitleGeneric)
}

func TestComputeSignals_Passthrough(t *testing.T) {
	minV, medV, sim := 0.1, 0.5, 0.7
	in := Input{
		Title: "x", Content: "текст",
		CoherenceMin: &minV, CoherenceMedian: &medV, TitleTextSimilarity: &sim,
		Keywords: 4, Links: 2, HasEmbedding: true,
	}
	s := ComputeSignals(in, baseThresholds())
	assert.Equal(t, &minV, s.CoherenceMin)
	assert.Equal(t, &medV, s.CoherenceMedian)
	assert.Equal(t, &sim, s.TitleTextSimilarity)
	assert.Equal(t, 4, s.Keywords)
	assert.Equal(t, 2, s.Links)
	assert.True(t, s.HasEmbedding)
}
