package importer

import "context"

// ContentExtractor fetches a web page and extracts a readable title and text.
// Infrastructure implementations live outside the application layer.
type ContentExtractor interface {
	Extract(ctx context.Context, rawURL string) (title string, text string, err error)
}
