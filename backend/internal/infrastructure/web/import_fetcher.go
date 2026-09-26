package web

import (
	"bytes"
	"context"
	"fmt"
	"io"
	importer "knowledge-graph/internal/application/import"
	"knowledge-graph/internal/shared/textutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

const (
	maxFetchBodySize = 1 << 20 // 1 MiB
	maxTitleRunes    = 200
)

// ImportFetcher fetches a web page and extracts its structured content
// (URL-HEADING-1 stage A). It is the production implementation of
// importer.ContentExtractor.
type ImportFetcher struct {
	client          *http.Client
	contentMaxRunes int
}

// NewImportFetcher creates an ImportFetcher with safe defaults. The section
// budget comes from IMPORT_CONTENT_MAX_RUNES (default 20000, stage A).
func NewImportFetcher() *ImportFetcher {
	return &ImportFetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		contentMaxRunes: contentMaxRunesFromEnv(),
	}
}

// NewImportFetcherWithClient creates an ImportFetcher with a custom HTTP client.
func NewImportFetcherWithClient(client *http.Client) *ImportFetcher {
	f := NewImportFetcher()
	f.client = client
	return f
}

func contentMaxRunesFromEnv() int {
	if v := os.Getenv("IMPORT_CONTENT_MAX_RUNES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultContentMaxRunes
}

// Extract fetches the URL, parses HTML and returns the structured page
// (title candidates, outline, Markdown body, related links, truncation
// metadata). The caller (application layer) must already validate and
// normalize the URL.
func (f *ImportFetcher) Extract(ctx context.Context, rawURL string) (*importer.ExtractedPage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "KnowledgeGraphBot/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Read the (1 MiB-capped) body fully: charset detection needs to inspect
	// the bytes anyway, and buffering lets us override an uncertain guess —
	// x/net falls back to windows-1252 on pages without a declaration
	// (Confluence: <meta charset="">), which garbles legitimate UTF-8.
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxFetchBodySize))
	if err != nil {
		return nil, err
	}

	// Convert to UTF-8 based on the declared Content-Type or a <meta charset>
	// declaration. When detection is not certain, HTML5 mandates UTF-8.
	enc, _, certain := charset.DetermineEncoding(body, resp.Header.Get("Content-Type"))
	var reader io.Reader = bytes.NewReader(body)
	if certain && enc != nil {
		reader = enc.NewDecoder().Reader(reader)
	}

	doc, err := html.Parse(reader)
	if err != nil {
		return nil, err
	}

	page := extractPage(doc, rawURL, f.contentMaxRunes)

	// Sanitize so we never pass an invalid UTF-8 string further up.
	page.Title = textutil.SanitizeUTF8(cleanText(page.Title))
	page.Text = textutil.SanitizeUTF8(page.Text)
	for i := range page.TitleCandidates {
		page.TitleCandidates[i] = textutil.SanitizeUTF8(cleanText(page.TitleCandidates[i]))
	}
	for i := range page.Outline {
		page.Outline[i].Text = textutil.SanitizeUTF8(page.Outline[i].Text)
	}

	// Truncate the title without splitting multi-byte runes.
	page.Title = textutil.TruncateToMaxRunes(page.Title, maxTitleRunes)

	return page, nil
}

func extractTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {
		return cleanText(extractText(n))
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if t := extractTitle(c); t != "" {
			return t
		}
	}
	return ""
}

func cleanText(s string) string {
	s = strings.TrimSpace(s)
	lines := strings.Split(s, "\n")
	var out []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		out = append(out, strings.Join(fields, " "))
	}
	return strings.Join(out, "\n")
}

func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		b.WriteString(extractText(c))
	}
	return b.String()
}
