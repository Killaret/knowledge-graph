package importer

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

// ContentExtractor fetches a web page and extracts a readable title and text.
type ContentExtractor interface {
	Extract(ctx context.Context, rawURL string) (title string, text string, err error)
}

// DefaultExtractor is a production implementation of ContentExtractor.
// It uses an HTTP client with a short timeout and SSRF-safe URL checks.
type DefaultExtractor struct {
	client *http.Client
}

// NewDefaultExtractor creates a ContentExtractor with safe defaults.
func NewDefaultExtractor() *DefaultExtractor {
	return &DefaultExtractor{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Extract fetches the URL, parses HTML and returns title and plain text.
func (e *DefaultExtractor) Extract(ctx context.Context, rawURL string) (string, string, error) {
	if !IsAllowedURL(rawURL) {
		return "", "", fmt.Errorf("URL is not allowed: %s", rawURL)
	}

	u, err := NormalizeURL(rawURL)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "KnowledgeGraphBot/1.0")

	resp, err := e.client.Do(req)
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
		title = u
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
	if n.Type != html.ElementNode {
		return ""
	}

	switch n.Data {
	case "script", "style", "noscript", "nav", "footer", "header", "aside", "button", "input", "select", "textarea":
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
