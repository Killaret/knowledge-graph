package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	noteDomain "knowledge-graph/internal/domain/note"
)

// DraftRepository implements the DraftRepository interface using MongoDB
type DraftRepository struct {
	client     *Client
	collection *mongo.Collection
	ttlHours   int
}

// DraftModel represents the MongoDB document structure for drafts
type DraftModel struct {
	ID        primitive.ObjectID `bson:"_id"`
	NoteID    string             `bson:"note_id"`
	UserID    string             `bson:"user_id"`
	Content   string             `bson:"content"`
	Title     string             `bson:"title"`
	State     string             `bson:"state"`
	UpdatedAt time.Time          `bson:"updated_at"`
	CreatedAt time.Time          `bson:"created_at"`
}

// NewDraftRepository creates a new DraftRepository
func NewDraftRepository(client *Client, ttlHours int) (*DraftRepository, error) {
	if client == nil {
		return nil, fmt.Errorf("MongoDB client is required")
	}

	collection := client.Collection("drafts")

	// Create TTL index on updated_at for automatic cleanup
	indexModel := mongo.IndexModel{
		Keys: bson.M{
			"updated_at": 1,
		},
		Options: options.Index().
			SetExpireAfterSeconds(int32(ttlHours * 3600)).
			SetName("draft_ttl_idx"),
	}

	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create TTL index: %w", err)
	}

	return &DraftRepository{
		client:     client,
		collection: collection,
		ttlHours:   ttlHours,
	}, nil
}

// Save saves a draft to MongoDB
func (r *DraftRepository) Save(ctx context.Context, draft *noteDomain.Draft) error {
	model := r.toModel(draft)

	_, err := r.collection.InsertOne(ctx, model)
	if err != nil {
		return fmt.Errorf("failed to save draft: %w", err)
	}

	return nil
}

// FindByNoteAndUser finds a draft by note ID and user ID
func (r *DraftRepository) FindByNoteAndUser(ctx context.Context, noteID, userID uuid.UUID) (*noteDomain.Draft, error) {
	filter := bson.M{
		"note_id": noteID.String(),
		"user_id": userID.String(),
	}

	var model DraftModel
	err := r.collection.FindOne(ctx, filter).Decode(&model)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find draft: %w", err)
	}

	return r.toDomain(&model), nil
}

// FindActiveByUser finds all active drafts for a user
func (r *DraftRepository) FindActiveByUser(ctx context.Context, userID uuid.UUID) ([]*noteDomain.Draft, error) {
	filter := bson.M{
		"user_id": userID.String(),
		"state":   string(noteDomain.DraftStateActive),
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find active drafts: %w", err)
	}
	defer cursor.Close(ctx)

	var models []DraftModel
	if err := cursor.All(ctx, &models); err != nil {
		return nil, fmt.Errorf("failed to decode drafts: %w", err)
	}

	drafts := make([]*noteDomain.Draft, len(models))
	for i, model := range models {
		drafts[i] = r.toDomain(&model)
	}

	return drafts, nil
}

// FindByID finds a draft by its ID
func (r *DraftRepository) FindByID(ctx context.Context, id uuid.UUID) (*noteDomain.Draft, error) {
	filter := bson.M{"_id": id}

	var model DraftModel
	err := r.collection.FindOne(ctx, filter).Decode(&model)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find draft by ID: %w", err)
	}

	return r.toDomain(&model), nil
}

// DeleteByID deletes a draft by its ID
func (r *DraftRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"_id": id}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete draft: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("draft not found")
	}

	return nil
}

// DeleteExpired deletes drafts that haven't been updated since the given time
func (r *DraftRepository) DeleteExpired(ctx context.Context, before time.Time) (int, error) {
	filter := bson.M{
		"updated_at": bson.M{"$lt": before},
	}

	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired drafts: %w", err)
	}

	return int(result.DeletedCount), nil
}

// Update updates an existing draft
func (r *DraftRepository) Update(ctx context.Context, draft *noteDomain.Draft) error {
	model := r.toModel(draft)

	filter := bson.M{"_id": model.ID}
	update := bson.M{"$set": model}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update draft: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("draft not found")
	}

	return nil
}

// toModel converts a domain Draft to a DraftModel
func (r *DraftRepository) toModel(draft *noteDomain.Draft) *DraftModel {
	return &DraftModel{
		ID:        primitive.NewObjectID(),
		NoteID:    draft.NoteID().String(),
		UserID:    draft.UserID().String(),
		Content:   draft.Content(),
		Title:     draft.Title(),
		State:     string(draft.State()),
		UpdatedAt: draft.UpdatedAt(),
		CreatedAt: draft.CreatedAt(),
	}
}

// toDomain converts a DraftModel to a domain Draft
func (r *DraftRepository) toDomain(model *DraftModel) *noteDomain.Draft {
	noteID, _ := uuid.Parse(model.NoteID)
	userID, _ := uuid.Parse(model.UserID)
	draftID, _ := uuid.Parse(model.ID.Hex())

	return noteDomain.ReconstructDraft(
		draftID,
		noteID,
		userID,
		model.Content,
		model.Title,
		noteDomain.DraftState(model.State),
		model.CreatedAt,
		model.UpdatedAt,
	)
}
