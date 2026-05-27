package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
)

// DraftRepository implements the DraftRepository interface using MongoDB
type DraftRepository struct {
	collection *mongo.Collection
}

// NewDraftRepository creates a new DraftRepository
func NewDraftRepository(client *Client) *DraftRepository {
	return &DraftRepository{
		collection: client.GetCollection("drafts"),
	}
}

// ensureIndex ensures the TTL index exists for automatic cleanup
func (r *DraftRepository) ensureIndex(ctx context.Context) error {
	// Create TTL index on updated_at field (7 days = 604800 seconds)
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"updated_at", 1}},
		Options: options.Index().SetExpireAfterSeconds(604800),
	}

	_, err := r.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}

	// Create compound index on note_id and user_id for faster lookups
	compoundIndex := mongo.IndexModel{
		Keys:    bson.D{{"note_id", 1}, {"user_id", 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err = r.collection.Indexes().CreateOne(ctx, compoundIndex)
	return err
}

// draftModel represents the MongoDB document structure
type draftModel struct {
	ID        uuid.UUID `bson:"_id"`
	NoteID    uuid.UUID `bson:"note_id"`
	UserID    uuid.UUID `bson:"user_id"`
	Content   string    `bson:"content"`
	Title     string    `bson:"title"`
	State     string    `bson:"state"`
	UpdatedAt time.Time `bson:"updated_at"`
	CreatedAt time.Time `bson:"created_at"`
}

func (m *draftModel) toDomain() *note.Draft {
	return note.ReconstructDraft(
		m.ID,
		m.NoteID,
		m.UserID,
		m.Content,
		m.Title,
		note.DraftState(m.State),
		m.CreatedAt,
		m.UpdatedAt,
	)
}

func draftToModel(d *note.Draft) *draftModel {
	return &draftModel{
		ID:        d.ID(),
		NoteID:    d.NoteID(),
		UserID:    d.UserID(),
		Content:   d.Content(),
		Title:     d.Title(),
		State:     string(d.State()),
		UpdatedAt: d.UpdatedAt(),
		CreatedAt: d.CreatedAt(),
	}
}

// Save saves a draft to MongoDB
func (r *DraftRepository) Save(ctx context.Context, draft *note.Draft) error {
	// Ensure indexes exist
	if err := r.ensureIndex(ctx); err != nil {
		return err
	}

	model := draftToModel(draft)
	filter := bson.D{{"_id", draft.ID()}}
	update := bson.D{
		{"$set", model},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// FindByNoteAndUser finds a draft by note ID and user ID
func (r *DraftRepository) FindByNoteAndUser(ctx context.Context, noteID, userID uuid.UUID) (*note.Draft, error) {
	filter := bson.D{
		{"note_id", noteID},
		{"user_id", userID},
	}

	var model draftModel
	err := r.collection.FindOne(ctx, filter).Decode(&model)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return model.toDomain(), nil
}

// FindActiveByUser finds all active drafts for a user
func (r *DraftRepository) FindActiveByUser(ctx context.Context, userID uuid.UUID) ([]*note.Draft, error) {
	filter := bson.D{
		{"user_id", userID},
		{"state", note.DraftStateActive},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var models []*draftModel
	if err := cursor.All(ctx, &models); err != nil {
		return nil, err
	}

	drafts := make([]*note.Draft, len(models))
	for i, model := range models {
		drafts[i] = model.toDomain()
	}

	return drafts, nil
}

// FindByID finds a draft by its ID
func (r *DraftRepository) FindByID(ctx context.Context, id uuid.UUID) (*note.Draft, error) {
	filter := bson.D{{"_id", id}}

	var model draftModel
	err := r.collection.FindOne(ctx, filter).Decode(&model)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return model.toDomain(), nil
}

// DeleteByID deletes a draft by its ID
func (r *DraftRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	filter := bson.D{{"_id", id}}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// DeleteExpired deletes drafts that haven't been updated since the given time
func (r *DraftRepository) DeleteExpired(ctx context.Context, before time.Time) (int, error) {
	filter := bson.D{
		{"updated_at", bson.D{{"$lt", before}}},
		{"state", bson.D{{"$ne", note.DraftStatePublished}}},
	}

	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return int(result.DeletedCount), nil
}

// Update updates an existing draft
func (r *DraftRepository) Update(ctx context.Context, draft *note.Draft) error {
	model := draftToModel(draft)
	filter := bson.D{{"_id", draft.ID()}}
	update := bson.D{{"$set", model}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}
