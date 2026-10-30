package web

import (
	"context"
	"fmt"
	"io"
	"knowledge-graph/internal/shared/textutil"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

const (
	maxFetchBodySize = 1 << 20 // 1 MiB
	maxTitleRunes    = 200
	maxTextRunes     = 5000
)

// ImportFetcher fetches a web page and extracts a readable title and text.
// It is a production implementation of importer.ContentExtractor.
type ImportFetcher struct {
	client *http.Client
}

// NewImportFetcher creates an ImportFetcher with safe defaults.
func NewImportFetcher() *ImportFetcher {
	return &ImportFetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewImportFetcherWithClient creates an ImportFetcher with a custom HTTP client.
func NewImportFetcherWithClient(client *http.Client) *ImportFetcher {
	return &ImportFetcher{client: client}
}

// Extract fetches the URL, parses HTML and returns title and plain text.
// The caller (application layer) must already validate and normalize the URL.
func (f *ImportFetcher) Extract(ctx context.Context, rawURL string) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "KnowledgeGraphBot/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	lr := io.LimitReader(resp.Body, maxFetchBodySize)

	// Convert to UTF-8 based on the declared Content-Type or a <meta charset>
	// declaration. If detection fails, fall back to reading the body as-is.
	reader, err := charset.NewReader(lr, resp.Header.Get("Content-Type"))
	if err != nil {
		reader = lr
	}

	doc, err := html.Parse(reader)
	if err != nil {
		return "", "", err
	}

	title := extractTitle(doc)
	text := extractVisibleText(doc)

	if title == "" {
		title = rawURL
	}

	// Sanitize so we never pass an invalid UTF-8 string further up.
	title = textutil.SanitizeUTF8(cleanText(title))
	text = textutil.SanitizeUTF8(cleanText(text))

	// Truncate without splitting multi-byte runes.
	title = textutil.TruncateToMaxRunes(title, maxTitleRunes)
	text = textutil.TruncateToMaxRunes(text, maxTextRunes)

	return title, text, nil
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

func extractVisibleText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	if n.Type == html.DocumentNode {
		var parts []string
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if t := extractVisibleText(c); t != "" {
				parts = append(parts, t)
			}
		}
		return strings.Join(parts, " ")
	}
	if n.Type != html.ElementNode {
		return ""
	}

	switch n.Data {
	case "title", "script", "style", "noscript", "nav", "footer", "header", "aside", "button", "input", "select", "textarea":
		return ""
	}

	var parts []string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if t := extractVisibleText(c); t != "" {
			parts = append(parts, t)
		}
	}

	if n.Data == "p" || n.Data == "div" || n.Data == "br" || n.Data == "li" || n.Data == "h1" || n.Data == "h2" || n.Data == "h3" || n.Data == "h4" || n.Data == "h5" || n.Data == "h6" {
		return strings.Join(parts, " ") + "\n"
	}
	return strings.Join(parts, " ")
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
