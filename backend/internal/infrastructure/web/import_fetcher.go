package web

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const maxFetchBodySize = 1 << 20 // 1 MiB

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

	body := io.LimitReader(resp.Body, maxFetchBodySize)
	doc, err := html.Parse(body)
	if err != nil {
		return "", "", err
	}

	title := extractTitle(doc)
	text := extractVisibleText(doc)

	if title == "" {
		title = rawURL
	}
	text = cleanText(text)
	if len(text) > 5000 {
		text = text[:5000]
	}

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
