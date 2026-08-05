package web

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			wantTitle: "Test Page",
			wantText:  "Hello\nWorld",
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
			title, text, err := f.Extract(context.Background(), url)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantTitle, title)
			assert.Equal(t, tt.wantText, text)
		})
	}
}

func TestImportFetcher_Extract_TitleFallback(t *testing.T) {
	htmlBody := `<html><head><title>  </title></head><body><p>Only body text</p></body></html>`
	srv := httptest.NewServer(simpleHTMLServer(htmlBody))
	defer srv.Close()

	f := NewImportFetcherWithClient(srv.Client())
	title, text, err := f.Extract(context.Background(), srv.URL)
	require.NoError(t, err)
	assert.Equal(t, srv.URL, title, "empty title should fall back to URL")
	assert.Equal(t, "Only body text", text)
}

func TestImportFetcher_Extract_Cancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// never respond
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	f := NewImportFetcherWithClient(srv.Client())
	_, _, err := f.Extract(ctx, srv.URL)
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
