package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/uuid"
)

// NlpArtifactsCollection is the Mongo collection holding NLP-4 derived
// artifacts (normalized text + structural chunks). Rebuildable data —
// notes.content in Postgres stays the source of truth.
const NlpArtifactsCollection = "nlp_artifacts"

// NlpArtifactStatus* — lifecycle states of an artifact document.
const (
	NlpArtifactCurrent    = "current"
	NlpArtifactSuperseded = "superseded"
)

// NlpArtifactChunk mirrors the CHUNK-1 chunk contract returned by /normalize.
type NlpArtifactChunk struct {
	Idx         int      `bson:"idx"`
	Text        string   `bson:"text"`
	HeadingPath []string `bson:"heading_path"`
	CharSpan    [2]int   `bson:"char_span"`
	TokenCount  int      `bson:"token_count"`
	Kind        string   `bson:"kind"`
	ForcedSplit bool     `bson:"forced_split"`
}

// NlpArtifactMetrics mirrors the /normalize metrics block.
type NlpArtifactMetrics struct {
	RawTokens      int      `bson:"raw_tokens"`
	NormTokens     int      `bson:"norm_tokens"`
	Compression    float64  `bson:"compression"`
	Iterations     int      `bson:"iterations"`
	StopReason     string   `bson:"stop_reason"`
	EmbCosine      *float64 `bson:"emb_cosine,omitempty"`
	RolledBack     bool     `bson:"rolled_back"`
	RollbackReason string   `bson:"rollback_reason,omitempty"`
	Skipped        bool     `bson:"skipped"`
}

// NlpArtifact is the Mongo document for a note's normalization artifact.
type NlpArtifact struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	NoteID          uuid.UUID          `bson:"note_id"`
	SourceHash      string             `bson:"source_hash"`
	PipelineVersion string             `bson:"pipeline_version"`
	ModelVersion    string             `bson:"model_version"`
	NormalizedText  string             `bson:"normalized_text"`
	Chunks          []NlpArtifactChunk `bson:"chunks"`
	Metrics         NlpArtifactMetrics `bson:"metrics"`
	Status          string             `bson:"status"` // current|superseded
	CreatedAt       time.Time          `bson:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at"`
}

// NlpArtifactsRepository stores NLP-4 artifacts in MongoDB.
type NlpArtifactsRepository struct {
	collection *mongo.Collection
}

// NewNlpArtifactsRepository creates the repository for the nlp_artifacts collection.
func NewNlpArtifactsRepository(client *Client) *NlpArtifactsRepository {
	return &NlpArtifactsRepository{
		collection: client.GetCollection(NlpArtifactsCollection),
	}
}

// EnsureIndexes creates the collection indexes:
//   - unique (note_id, pipeline_version) restricted to status=current — one
//     live artifact per note per pipeline version;
//   - (note_id) for history lookups and cascade deletes.
func (r *NlpArtifactsRepository) EnsureIndexes(ctx context.Context) error {
	uniqueCurrent := mongo.IndexModel{
		Keys: bson.D{primitive.E{Key: "note_id", Value: 1}, primitive.E{Key: "pipeline_version", Value: 1}},
		Options: options.Index().
			SetUnique(true).
			SetPartialFilterExpression(bson.D{primitive.E{Key: "status", Value: NlpArtifactCurrent}}).
			SetName("note_pipeline_current"),
	}
	byNote := mongo.IndexModel{
		Keys:    bson.D{primitive.E{Key: "note_id", Value: 1}},
		Options: options.Index().SetName("note_id"),
	}
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{uniqueCurrent, byNote})
	return err
}

// FindCurrent returns the current artifact for (note_id, pipeline_version),
// or nil when none exists.
func (r *NlpArtifactsRepository) FindCurrent(ctx context.Context, noteID uuid.UUID, pipelineVersion string) (*NlpArtifact, error) {
	filter := bson.D{
		primitive.E{Key: "note_id", Value: noteID},
		primitive.E{Key: "pipeline_version", Value: pipelineVersion},
		primitive.E{Key: "status", Value: NlpArtifactCurrent},
	}
	var doc NlpArtifact
	err := r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &doc, nil
}

// FindCurrentSourceHash returns the source_hash of the current artifact, or
// ("", false) when none exists. Cheap probe used by the recompute command.
func (r *NlpArtifactsRepository) FindCurrentSourceHash(ctx context.Context, noteID uuid.UUID, pipelineVersion string) (string, bool, error) {
	filter := bson.D{
		primitive.E{Key: "note_id", Value: noteID},
		primitive.E{Key: "pipeline_version", Value: pipelineVersion},
		primitive.E{Key: "status", Value: NlpArtifactCurrent},
	}
	opts := options.FindOne().SetProjection(bson.D{primitive.E{Key: "source_hash", Value: 1}})
	var doc struct {
		SourceHash string `bson:"source_hash"`
	}
	err := r.collection.FindOne(ctx, filter, opts).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", false, nil
		}
		return "", false, err
	}
	return doc.SourceHash, true, nil
}

// SaveCurrent stores the artifact as the single current document for
// (note_id, pipeline_version). With historyEnabled, previous current docs are
// marked superseded; without it they are deleted so history never accrues.
// Supersede happens before insert — the unique partial index on current docs
// would otherwise reject the write.
func (r *NlpArtifactsRepository) SaveCurrent(ctx context.Context, doc *NlpArtifact, historyEnabled bool) error {
	now := time.Now().UTC()
	doc.ID = primitive.NewObjectID()
	doc.Status = NlpArtifactCurrent
	doc.CreatedAt = now
	doc.UpdatedAt = now

	filter := bson.D{
		primitive.E{Key: "note_id", Value: doc.NoteID},
		primitive.E{Key: "pipeline_version", Value: doc.PipelineVersion},
		primitive.E{Key: "status", Value: NlpArtifactCurrent},
	}
	if historyEnabled {
		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "status", Value: NlpArtifactSuperseded},
				primitive.E{Key: "updated_at", Value: now},
			}},
		}
		if _, err := r.collection.UpdateMany(ctx, filter, update); err != nil {
			return err
		}
	} else {
		if _, err := r.collection.DeleteMany(ctx, filter); err != nil {
			return err
		}
	}
	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

// DeleteByNoteID removes every artifact of the note (current and history) —
// the cascade half of the note lifecycle.
func (r *NlpArtifactsRepository) DeleteByNoteID(ctx context.Context, noteID uuid.UUID) (int64, error) {
	res, err := r.collection.DeleteMany(ctx, bson.D{primitive.E{Key: "note_id", Value: noteID}})
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}

// CountByStatus reports collection stats for diagnostics.
func (r *NlpArtifactsRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.D{primitive.E{Key: "status", Value: status}})
}
