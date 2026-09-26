package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/uuid"

	appquality "knowledge-graph/internal/application/quality"
)

// QualityLogCollection holds the NOTE-QUALITY-1 assessment log — one entry
// per assessment, trimmed to the last LogKeepPerNote per note.
const QualityLogCollection = "quality_log"

// qualityLogDoc is the stored shape: the log entry plus an id and a write
// timestamp used for trimming.
type qualityLogDoc struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	NoteID     uuid.UUID          `bson:"note_id"`
	SourceHash string             `bson:"source_hash"`
	Trigger    string             `bson:"trigger"`
	Record     appquality.Record  `bson:"record"`
	Changed    bool               `bson:"changed"`
	CreatedAt  time.Time          `bson:"created_at"`
}

// QualityLogRepository stores quality assessments per note.
type QualityLogRepository struct {
	collection *mongo.Collection
}

// NewQualityLogRepository creates the repository for the quality_log collection.
func NewQualityLogRepository(client *Client) *QualityLogRepository {
	return &QualityLogRepository{collection: client.GetCollection(QualityLogCollection)}
}

// EnsureIndexes creates the (note_id, created_at) index used for listing.
func (r *QualityLogRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{primitive.E{Key: "note_id", Value: 1}, primitive.E{Key: "created_at", Value: -1}},
		Options: options.Index().SetName("note_created"),
	})
	return err
}

// Append writes a log entry and trims the note's history to the last
// LogKeepPerNote documents.
func (r *QualityLogRepository) Append(ctx context.Context, entry appquality.LogEntry) error {
	doc := qualityLogDoc{
		ID:         primitive.NewObjectID(),
		NoteID:     entry.NoteID,
		SourceHash: entry.SourceHash,
		Trigger:    entry.Trigger,
		Record:     entry.Record,
		Changed:    entry.Changed,
		CreatedAt:  time.Now().UTC(),
	}
	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return err
	}
	return r.trim(ctx, entry.NoteID)
}

func (r *QualityLogRepository) trim(ctx context.Context, noteID uuid.UUID) error {
	// Find ids beyond the keep-window (newest first) and delete them.
	opts := options.Find().
		SetSort(bson.D{primitive.E{Key: "created_at", Value: -1}, primitive.E{Key: "_id", Value: -1}}).
		SetSkip(appquality.LogKeepPerNote).
		SetProjection(bson.D{primitive.E{Key: "_id", Value: 1}})
	cur, err := r.collection.Find(ctx, bson.D{primitive.E{Key: "note_id", Value: noteID}}, opts)
	if err != nil {
		return err
	}
	var stale []struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	if err := cur.All(ctx, &stale); err != nil {
		return err
	}
	if len(stale) == 0 {
		return nil
	}
	ids := make([]primitive.ObjectID, len(stale))
	for i, d := range stale {
		ids[i] = d.ID
	}
	_, err = r.collection.DeleteMany(ctx, bson.D{primitive.E{Key: "_id", Value: bson.M{"$in": ids}}})
	return err
}

// List returns the note's entries, newest first, capped at limit.
func (r *QualityLogRepository) List(ctx context.Context, noteID uuid.UUID, limit int) ([]appquality.LogEntry, error) {
	opts := options.Find().
		SetSort(bson.D{primitive.E{Key: "created_at", Value: -1}, primitive.E{Key: "_id", Value: -1}}).
		SetLimit(int64(limit))
	cur, err := r.collection.Find(ctx, bson.D{primitive.E{Key: "note_id", Value: noteID}}, opts)
	if err != nil {
		return nil, err
	}
	var docs []qualityLogDoc
	if err := cur.All(ctx, &docs); err != nil {
		return nil, err
	}
	out := make([]appquality.LogEntry, len(docs))
	for i, d := range docs {
		out[i] = appquality.LogEntry{
			NoteID:     d.NoteID,
			SourceHash: d.SourceHash,
			Trigger:    d.Trigger,
			Record:     d.Record,
			Changed:    d.Changed,
		}
	}
	return out, nil
}

// ListQuality is the QualityReader-facing alias of List.
func (r *QualityLogRepository) ListQuality(ctx context.Context, noteID uuid.UUID, limit int) ([]appquality.LogEntry, error) {
	return r.List(ctx, noteID, limit)
}

// Latest returns the newest entry for the note, or nil when none exists.
func (r *QualityLogRepository) LatestQuality(ctx context.Context, noteID uuid.UUID) (*appquality.LogEntry, error) {
	entries, err := r.List(ctx, noteID, 1)
	if err != nil || len(entries) == 0 {
		return nil, err
	}
	return &entries[0], nil
}

// MarkManualReview sets needs_manual_review on the newest entry of the given
// text version — the stop rule's flag for a gate that survived all attempts.
func (r *QualityLogRepository) MarkManualReview(ctx context.Context, noteID uuid.UUID, sourceHash string) (bool, error) {
	opts := options.FindOne().
		SetSort(bson.D{primitive.E{Key: "created_at", Value: -1}, primitive.E{Key: "_id", Value: -1}})
	var doc qualityLogDoc
	err := r.collection.FindOne(ctx, bson.D{
		primitive.E{Key: "note_id", Value: noteID},
		primitive.E{Key: "source_hash", Value: sourceHash},
	}, opts).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	if doc.Record.NeedsManualReview {
		return true, nil
	}
	_, err = r.collection.UpdateOne(ctx,
		bson.D{primitive.E{Key: "_id", Value: doc.ID}},
		bson.D{primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "record.needs_manual_review", Value: true},
		}}})
	return err == nil, err
}

// DeleteByNoteID — cascade: the log goes away with the note.
func (r *QualityLogRepository) DeleteByNoteID(ctx context.Context, noteID uuid.UUID) (int64, error) {
	res, err := r.collection.DeleteMany(ctx, bson.D{primitive.E{Key: "note_id", Value: noteID}})
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}
