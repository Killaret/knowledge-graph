package web

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/charmap"
)

func TestImportFetcher_Extract(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		server    http.Handler
		wantTitle string
		wantText  string
		wantErr   bool
	}{
		{
			name:      "extracts title and visible text",
			url:       "/page",
			server:    simpleHTMLServer(`<html><head><title>Test Page</title></head><body><h1>Hello</h1><p>World</p></body></html>`),
			wantTitle: "Hello", // h1 is candidate #1 and beats <title> (A5)
			wantText:  "## Hello\n\nWorld",
		},
		{
			name:      "skips script and style content",
			url:       "/page",
			server:    simpleHTMLServer(`<html><head><title>Clean</title><script>alert(1)</script></head><body><p>visible</p><style>.x{}</style></body></html>`),
			wantTitle: "Clean",
			wantText:  "visible",
		},
		{
			name:    "returns error on 404",
			url:     "/missing",
			server:  statusServer(http.StatusNotFound),
			wantErr: true,
		},
		{
			name:    "returns error on 500",
			url:     "/error",
			server:  statusServer(http.StatusInternalServerError),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.server)
			defer srv.Close()

			url := tt.url
			if url == "/page" || url == "/missing" || url == "/error" {
				url = srv.URL + url
			}

			f := NewImportFetcherWithClient(srv.Client())
			page, err := f.Extract(context.Background(), url)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, page)
			assert.Equal(t, tt.wantTitle, page.Title)
			assert.Equal(t, tt.wantText, page.Text)
		})
	}
}

func TestImportFetcher_Extract_TitleFallback(t *testing.T) {
	htmlBody := `<html><head><title>  </title></head><body><p>Only body text</p></body></html>`
	srv := httptest.NewServer(simpleHTMLServer(htmlBody))
	defer srv.Close()

	f := NewImportFetcherWithClient(srv.Client())
	page, err := f.Extract(context.Background(), srv.URL)
	require.NoError(t, err)
	// Empty <title> and no h1 → last URL path segment, then the URL itself
	// (srv.URL has an empty path, so the URL fallback wins).
	assert.Equal(t, srv.URL, page.Title, "empty title should fall back to URL")
	assert.Equal(t, "Only body text", page.Text)
}

func TestImportFetcher_Extract_Cancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// never respond
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	f := NewImportFetcherWithClient(srv.Client())
	_, err := f.Extract(ctx, srv.URL)
	require.Error(t, err)
}

func simpleHTMLServer(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, body)
	})
}

func statusServer(status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	})
}

func TestImportFetcher_Extract_CharsetConversion(t *testing.T) {
	// A page encoded in windows-1251 that contains Cyrillic text.
	htmlBody := `<html><head><title>Полное руководство по IDOR</title></head><body><p>Уязвимости</p></body></html>`
	encoded, err := charmap.Windows1251.NewEncoder().String(htmlBody)
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=windows-1251")
		_, _ = w.Write([]byte(encoded))
	}))
	defer srv.Close()

	f := NewImportFetcherWithClient(srv.Client())
	page, err := f.Extract(context.Background(), srv.URL)
	require.NoError(t, err)

	assert.Equal(t, "Полное руководство по IDOR", page.Title)
	assert.Contains(t, page.Text, "Уязвимости")
	assert.True(t, utf8.ValidString(page.Title))
	assert.True(t, utf8.ValidString(page.Text))
}

func TestImportFetcher_Extract_RuneSafeTextTruncation(t *testing.T) {
	// 6000 multi-byte Cyrillic runes — one paragraph fits under the
	// IMPORT_CONTENT_MAX_RUNES=20000 budget, so nothing is dropped and no
	// rune is split.
	longText := strings.Repeat("ы", 6000)
	body := fmt.Sprintf(`<html><head><title>Title</title></head><body><p>%s</p></body></html>`, longText)

	srv := httptest.NewServer(simpleHTMLServer(body))
	defer srv.Close()

	f := NewImportFetcherWithClient(srv.Client())
	page, err := f.Extract(context.Background(), srv.URL)
	require.NoError(t, err)

	assert.Equal(t, "Title", page.Title)
	assert.True(t, utf8.ValidString(page.Text))
	assert.Equal(t, 6000, utf8.RuneCountInString(page.Text))
}

func TestImportFetcher_Extract_TruncatesLongTitle(t *testing.T) {
	// 250 Cyrillic runes exceed the 3–200 rune candidate rule, so the rule
	// falls through to the URL; the extraction-level cap is a second guard.
	longTitle := strings.Repeat("ы", 250)
	body := fmt.Sprintf(`<html><head><title>%s</title></head><body><p>text</p></body></html>`, longTitle)

	srv := httptest.NewServer(simpleHTMLServer(body))
	defer srv.Close()

	f := NewImportFetcherWithClient(srv.Client())
	page, err := f.Extract(context.Background(), srv.URL)
	require.NoError(t, err)

	assert.True(t, utf8.ValidString(page.Title))
	assert.LessOrEqual(t, utf8.RuneCountInString(page.Title), 200)
}

func TestImportFetcher_Extract_UncertainCharsetKeepsUTF8(t *testing.T) {
	// Atlassian/Skillbox regression: <meta charset=""> with no HTTP charset
	// makes charset.DetermineEncoding guess windows-1252 with certain=false —
	// an uncertain guess must not decode, UTF-8 survives as-is.
	htmlBody := `<html><head><meta charset=""><title>Настройка Keycloak</title></head><body><h1>Настройка Keycloak и интеграция LDAP</h1><p>Шаги</p></body></html>`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, htmlBody)
	}))
	defer srv.Close()

	f := NewImportFetcherWithClient(srv.Client())
	page, err := f.Extract(context.Background(), srv.URL)
	require.NoError(t, err)
	assert.Equal(t, "Настройка Keycloak и интеграция LDAP", page.Title)
	assert.Contains(t, page.Text, "Шаги")
	assert.True(t, utf8.ValidString(page.Text))
}
