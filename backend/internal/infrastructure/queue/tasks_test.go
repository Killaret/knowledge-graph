package queue

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestExtractKeywordsTaskPayload(t *testing.T) {
	t.Run("marshal and unmarshal", func(t *testing.T) {
		noteID := uuid.New()
		payload := ExtractKeywordsTaskPayload{
			NoteID: noteID.String(),
			TopN:   10,
		}

		data, err := json.Marshal(payload)
		assert.NoError(t, err)

		var decoded ExtractKeywordsTaskPayload
		err = json.Unmarshal(data, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, payload.NoteID, decoded.NoteID)
		assert.Equal(t, payload.TopN, decoded.TopN)
	})
}

func TestComputeEmbeddingTaskPayload(t *testing.T) {
	t.Run("marshal and unmarshal", func(t *testing.T) {
		noteID := uuid.New()
		payload := ComputeEmbeddingTaskPayload{
			NoteID: noteID.String(),
		}

		data, err := json.Marshal(payload)
		assert.NoError(t, err)

		var decoded ComputeEmbeddingTaskPayload
		err = json.Unmarshal(data, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, payload.NoteID, decoded.NoteID)
	})
}

func TestTaskTypeConstants(t *testing.T) {
	assert.Equal(t, "extract:keywords", TypeExtractKeywords)
	assert.Equal(t, "compute:embedding", TypeComputeEmbedding)
}
