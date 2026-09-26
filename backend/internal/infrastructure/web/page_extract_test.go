package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"

	importer "knowledge-graph/internal/application/import"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

// goldenEntry mirrors expected.json: the human-verified expected values from
// the probe report (docs/tasks/URL-HEADING-1-findings-probe.md).
type goldenEntry struct {
	URL          string   `json:"url"`
	Title        string   `json:"title"`
	OutlineFirst []string `json:"outline_first"` // first outline texts, in order
	FetchError   bool     `json:"fetch_error"`   // page was a 4xx/5xx at probe time
}

func loadGolden(t *testing.T) map[string]goldenEntry {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "urlheading", "expected.json"))
	if err != nil {
		t.Fatalf("read expected.json: %v", err)
	}
	var m map[string]goldenEntry
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse expected.json: %v", err)
	}
	return m
}

func extractSnapshot(t *testing.T, id, rawURL string) *importer.ExtractedPage {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "urlheading", id+".html"))
	if err != nil {
		t.Fatalf("open snapshot %s: %v", id, err)
	}
	// Mirror ImportFetcher.Extract: certain detection decodes, uncertain
	// guesses default to UTF-8 (HTML5).
	enc, _, certain := charset.DetermineEncoding(data, "")
	var reader io.Reader = bytes.NewReader(data)
	if certain && enc != nil {
		reader = enc.NewDecoder().Reader(reader)
	}
	doc, err := html.Parse(reader)
	if err != nil {
		t.Fatalf("parse %s: %v", id, err)
	}
	return extractPage(doc, rawURL, defaultContentMaxRunes)
}

// TestGoldenURLHeading runs the 19-snapshot golden set (URL-HEADING-1 stage A,
// criterion 1): title matches expected.json exactly, outline entries are
// checked as an ordered subsequence — extra headings between the expected
// ones are fine. The spec's bar is ≥15/19 for both metrics; the per-page
// report is printed for docs/tasks/URL-HEADING-1-heading-extraction.md.
func TestGoldenURLHeading(t *testing.T) {
	expected := loadGolden(t)
	titleHits, outlineHits, graded := 0, 0, 0
	var report []string

	ids := make([]string, 0, len(expected))
	for id := range expected {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		exp := expected[id]
		if exp.FetchError {
			continue // 17: the 404 page never reaches extraction
		}
		graded++
		page := extractSnapshot(t, id, exp.URL)

		if page.Title == exp.Title {
			titleHits++
			report = append(report, fmt.Sprintf("%s title OK", id))
		} else {
			report = append(report, fmt.Sprintf("%s title MISS: got %q, want %q (candidates %v)",
				id, page.Title, exp.Title, page.TitleCandidates))
		}

		if outlineContainsSubsequence(page.Outline, exp.OutlineFirst) {
			outlineHits++
		} else {
			var got []string
			for _, h := range page.Outline {
				got = append(got, h.Text)
			}
			report = append(report, fmt.Sprintf("%s outline MISS: want subsequence %v, got %v",
				id, exp.OutlineFirst, got))
		}
	}

	for _, line := range report {
		t.Log(line)
	}
	t.Logf("titles %d/%d, outlines %d/%d", titleHits, graded, outlineHits, graded)
	// Spec bar: ≥15/19. The tuned extractor reaches 18/18 (all non-404
	// fixtures), so every miss is now a hard regression — an intentional
	// extractor change must update expected.json in the same commit.
	assert.GreaterOrEqual(t, titleHits, 15, "golden titles below the 15/19 bar")
	assert.GreaterOrEqual(t, outlineHits, 15, "golden outlines below the 15/19 bar")
	for _, line := range report {
		if strings.Contains(line, "MISS") {
			t.Errorf("golden regression: %s", line)
		}
	}
}

// outlineContainsSubsequence reports whether want's texts appear in got in
// order.
func outlineContainsSubsequence(got []importer.OutlineEntry, want []string) bool {
	i := 0
	for _, h := range got {
		if i < len(want) && h.Text == want[i] {
			i++
		}
	}
	return i == len(want)
}

// TestGoldenDump prints actual extraction for every snapshot; used during
// development to fill expected.json. KG_DUMP=1 go test -run TestGoldenDump.
func TestGoldenDump(t *testing.T) {
	if os.Getenv("KG_DUMP") == "" {
		t.Skip("set KG_DUMP=1 to dump golden extraction")
	}
	entries, _ := os.ReadDir(filepath.Join("testdata", "urlheading"))
	expected := loadGolden(t)
	for _, e := range entries {
		name := e.Name()
		if filepath.Ext(name) != ".html" {
			continue
		}
		id := name[:len(name)-5]
		exp := expected[id]
		page := extractSnapshot(t, id, exp.URL)
		fmt.Printf("=== %s %s\n", id, exp.URL)
		fmt.Printf("TITLE: %q (source=%s)\n", page.Title, page.TitleSource)
		fmt.Printf("CANDS: %q\n", page.TitleCandidates)
		fmt.Printf("NOISE: %d dropped, sections_dropped=%d, links=%d\n",
			page.NoiseDropped, page.SectionsDropped, len(page.RelatedLinks))
		var ol []string
		for _, h := range page.Outline {
			ol = append(ol, fmt.Sprintf("h%d:%s", h.Level, h.Text))
		}
		fmt.Printf("OUTLINE: %v\n", ol)
		fmt.Printf("TEXT[%d runes]: %.300s\n", len([]rune(page.Text)), page.Text)
	}
}

// --- Mutation detectors and adversarial coverage (URL-HEADING-1, criteria
// 2/4/5 and the adversarial phase) ---

func parseHTML(t *testing.T, body string) *html.Node {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(body))
	require.NoError(t, err)
	return doc
}

func TestPromoHeadingsExcludedFromOutline(t *testing.T) {
	// Criterion 2: dropping "promo|subscribe" from noiseIDClass lets a
	// promo/subscribe h2 leak into the outline. Synthetic fixture: on the
	// real Skillbox snapshot (12) the subscribe blocks sit OUTSIDE the
	// selected container, so fixture alone cannot catch this mutation.
	doc := parseHTML(t, `<html><body><article><h1>Art</h1><div class="subscribe-widget"><h2>Спасибо за подписку</h2></div><h2>Real</h2></article></body></html>`)
	page := extractPage(doc, "https://example.com/a", defaultContentMaxRunes)
	texts := []string{}
	for _, h := range page.Outline {
		texts = append(texts, h.Text)
		assert.NotContains(t, h.Text, "подписку", "promo h2 leaked into outline")
	}
	assert.Equal(t, []string{"Art", "Real"}, texts)

	// Fixture check stays as a guard that real Skillbox promo never leaks.
	exp := loadGolden(t)["12"]
	for _, h := range extractSnapshot(t, "12", exp.URL).Outline {
		assert.NotContains(t, h.Text, "подписку")
	}
}

func TestArticleHeaderKeepsH1(t *testing.T) {
	// Criterion 2: mutating the <header> rule to drop every header loses the
	// h1 that article containers keep inside a header (fixture 15: babok).
	doc := parseHTML(t, `<html><body><div class="content"><header><h1>Статья в шапке</h1></header><p>text</p></div></body></html>`)
	page := extractPage(doc, "https://example.com/a", defaultContentMaxRunes)
	assert.Equal(t, "Статья в шапке", page.Title)

	// A top-level <header> under a <body> fallback IS chrome.
	doc = parseHTML(t, `<html><body><header><h1>Site Logo</h1><nav><a href="/x">x</a></nav></header><div class="content"><h1>Real Title</h1><p>text</p></div></body></html>`)
	page = extractPage(doc, "https://example.com/a", defaultContentMaxRunes)
	assert.Equal(t, "Real Title", page.Title)
}

func TestCodeFenceParityOnBudgetCut(t *testing.T) {
	// Criterion 2: a budget cut must never split a fenced code block —
	// the ``` count stays even and the dropped-sections trailer is
	// appended. Single kept fence + one dropped section: dropping the
	// closing fence (or cutting mid-block) makes the count odd.
	doc := parseHTML(t, `<html><body><article>
		<h1>T</h1><pre>`+strings.Repeat("x", 80)+`</pre>
		<h2>Second</h2><pre>`+strings.Repeat("y", 800)+`</pre><p>tail</p>
		</article></body></html>`)
	page := extractPage(doc, "https://example.com/a", 150)
	assert.Equal(t, 1, page.SectionsDropped)
	assert.Contains(t, page.Text, "обрезано")
	fences := strings.Count(page.Text, "```")
	assert.Equal(t, 2, fences, "expected exactly one balanced code block")
}

func TestSectionBudgetReplacesOld5000Cap(t *testing.T) {
	// Criterion 2 mutation detector: pinning the budget back to the old
	// 5000-rune text cap must fail — the Habr snapshot keeps far more.
	exp := loadGolden(t)["10"]
	page := extractSnapshot(t, "10", exp.URL)
	assert.Greater(t, utf8.RuneCountInString(page.Text), 5000,
		"IMPORT_CONTENT_MAX_RUNES budget regressed to the old 5000 cap")
	assert.Zero(t, page.SectionsDropped)
}

func TestRelatedLinksCapAndSchemeFilter(t *testing.T) {
	// Criterion 4: ≤20 links, http(s) only; the 21st link and non-http(s)
	// schemes (javascript:, mailto:, ftp:, relative) never reach metadata.
	var b strings.Builder
	b.WriteString(`<html><body><nav>`)
	for i := 0; i < 25; i++ {
		fmt.Fprintf(&b, `<a href="https://x.example/%d">l%d</a>`, i, i)
	}
	b.WriteString(`<a href="javascript:void(0)">js</a><a href="mailto:a@b.c">mail</a>` +
		`<a href="ftp://x/f">ftp</a><a href="/relative">rel</a>`)
	b.WriteString(`</nav><article><h1>T</h1><p>text</p></article></body></html>`)
	page := extractPage(parseHTML(t, b.String()), "https://example.com/", defaultContentMaxRunes)
	assert.Len(t, page.RelatedLinks, maxRelatedLinks)
	for _, l := range page.RelatedLinks {
		assert.True(t, strings.HasPrefix(l.URL, "https://"), "non-http link leaked: %s", l.URL)
	}
}

func TestHasSidebarModifierIsNotNoise(t *testing.T) {
	// Habr regression: tm-page__main_has-sidebar wraps the whole article —
	// a BEM modifier is not a sidebar (noiseIDClassAllowed).
	doc := parseHTML(t, `<html><body><div class="tm-page__main_has-sidebar"><article><h1>Habr Post</h1><p>body</p></article></div></body></html>`)
	page := extractPage(doc, "https://habr.com/x", defaultContentMaxRunes)
	assert.Equal(t, "Habr Post", page.Title)
	assert.Contains(t, page.Text, "body")
}

func TestModalContentIsNotAContainer(t *testing.T) {
	// jsonformatter regression: .modal-content must lose to the real
	// content block (chromeIDClass skips chrome-looking candidates).
	doc := parseHTML(t, `<html><body><div class="modal-content"><h2>Sign in</h2></div><div class="page-content"><h1>JSON Formatter</h1><p>real</p></div></body></html>`)
	page := extractPage(doc, "https://jsonformatter.org/", defaultContentMaxRunes)
	assert.Equal(t, "JSON Formatter", page.Title)
}

func TestOutlineNormalizesWithoutH1(t *testing.T) {
	// Adversarial: a page whose first heading is h3 must still start the
	// outline at level 2.
	doc := parseHTML(t, `<html><body><article><h3>Deep</h3><h4>Deeper</h4><p>x</p></article></body></html>`)
	page := extractPage(doc, "https://example.com/a", defaultContentMaxRunes)
	require.Len(t, page.Outline, 2)
	assert.Equal(t, 2, page.Outline[0].Level)
	assert.Equal(t, 3, page.Outline[1].Level)
}

func TestWideTableFallsBackToParagraphs(t *testing.T) {
	// Adversarial: a >6-column table renders as paragraphs, not Markdown.
	doc := parseHTML(t, `<html><body><article><h1>T</h1><table><tr><td>a</td><td>b</td><td>c</td><td>d</td><td>e</td><td>f</td><td>g</td></tr></table></article></body></html>`)
	page := extractPage(doc, "https://example.com/a", defaultContentMaxRunes)
	assert.NotContains(t, page.Text, "| --- |")
	assert.Contains(t, page.Text, "a b c d e f g")
}

func TestEmptyDocumentDoesNotPanic(t *testing.T) {
	page := extractPage(parseHTML(t, `<html><body></body></html>`), "https://example.com/", defaultContentMaxRunes)
	assert.Equal(t, "https://example.com/", page.Title)
	assert.Empty(t, page.Outline)
}

func TestNestedListRendering(t *testing.T) {
	doc := parseHTML(t, `<html><body><article><h1>T</h1><ul><li>one<ul><li>nested</li></ul></li><li>two</li></ul></article></body></html>`)
	page := extractPage(doc, "https://example.com/a", defaultContentMaxRunes)
	assert.Contains(t, page.Text, "- one")
	assert.Contains(t, page.Text, "  - nested")
	assert.Contains(t, page.Text, "- two")
}

func TestTitleSuffixStripping(t *testing.T) {
	// Criterion 2 mutation detector: without suffix stripping the <title>
	// candidate keeps " - jwt.io" and wins over the URL segment.
	doc := parseHTML(t, `<html><head><title>JSON Web Tokens - jwt.io</title></head><body><p>x</p></body></html>`)
	page := extractPage(doc, "https://jwt.io/", defaultContentMaxRunes)
	assert.Equal(t, "JSON Web Tokens", page.Title)
	assert.Contains(t, page.TitleCandidates, "JSON Web Tokens")
}
