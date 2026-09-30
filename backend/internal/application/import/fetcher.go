package importer

import "context"

// OutlineEntry is one heading of the extracted page outline. Level is
// normalized: the first heading of the page becomes 2 (`##` — the
// `## [title](url)` first line of the note is level 1 semantics), the rest
// shift relative to it, clamped to [2, 6].
type OutlineEntry struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
}

// RelatedLink is an anchor collected from noise areas (nav/aside/footer and
// page-level headers). The link text and URL are kept as metadata but never
// reach the note body (URL-HEADING-1, question 3).
type RelatedLink struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

// ExtractedPage is the stage-A URL-HEADING-1 extraction contract: a
// deterministic, rule-based snapshot of a web page — no model involved.
type ExtractedPage struct {
	Title           string         `json:"title"`
	Text            string         `json:"text"`
	TitleCandidates []string       `json:"title_candidates,omitempty"`
	TitleSource     string         `json:"title_source"` // "rule" at stage A; "model" arrives with stage B
	Outline         []OutlineEntry `json:"outline,omitempty"`
	RelatedLinks    []RelatedLink  `json:"related_links,omitempty"`
	NoiseDropped    int            `json:"noise_dropped,omitempty"`
	SectionsDropped int            `json:"sections_dropped,omitempty"`
}

// ContentExtractor fetches a web page and extracts its structured content.
// Infrastructure implementations live outside the application layer.
type ContentExtractor interface {
	Extract(ctx context.Context, rawURL string) (*ExtractedPage, error)
}
