package quality

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// corpusNote mirrors one row of the owner's local notes dataset. The dataset
// lives outside git (work-nlp4/) — this test skips silently when absent.
type corpusNote struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func corpusPath() string {
	// backend/internal/application/quality -> repo root -> work-nlp4
	return filepath.Join("..", "..", "..", "..", "work-nlp4", "notes_dataset.json")
}

// TestCorpusSignals runs the stage-1 signal set over the owner's corpus and
// asserts the expected counts from the NOTE-QUALITY-1 spec:
// truncated_by_import = 39, stub = 8, collection = 0.
// No text is logged — counts only.
func TestCorpusSignals(t *testing.T) {
	path := corpusPath()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("corpus dataset not available at %s", path)
	}
	var notes []corpusNote
	require.NoError(t, json.Unmarshal(raw, &notes))
	require.NotEmpty(t, notes)

	th := DefaultThresholds()
	var truncated, stubs, collections, texts, imports int
	for _, n := range notes {
		// Old imports carry no metadata — mimic production by deriving
		// source_url from the first-line link pattern only.
		in := Input{Title: n.Title, Content: n.Content}
		s := ComputeSignals(in, th)
		if s.TruncatedByImport {
			truncated++
		}
		if s.Origin == OriginImport {
			imports++
		}
		switch s.Kind {
		case KindStub:
			stubs++
		case KindCollection:
			collections++
		case KindText:
			texts++
		}
	}
	t.Logf("corpus: %d notes, %d imports, truncated=%d stub=%d collection=%d text=%d",
		len(notes), imports, truncated, stubs, collections, texts)

	assert.Equal(t, 108, len(notes))
	assert.Equal(t, 39, truncated, "truncated_by_import mismatch — legacy 5000-rune heuristic drifted")
	assert.Equal(t, 8, stubs, "stub count mismatch — link-line detection drifted")
	assert.Equal(t, 0, collections, "collection should be absent from this corpus")
}
