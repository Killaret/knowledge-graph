package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client represents MongoDB client wrapper
type Client struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewClient creates a new MongoDB client
func NewClient(ctx context.Context, uri, databaseName string) (*Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &Client{
		client:   client,
		database: client.Database(databaseName),
	}, nil
}

// GetDatabase returns the MongoDB database
func (c *Client) GetDatabase() *mongo.Database {
	if c == nil {
		return nil
	}
	return c.database
}

// Close closes the MongoDB connection
func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// GetCollection returns a collection by name
func (c *Client) GetCollection(name string) *mongo.Collection {
	if c == nil || c.database == nil {
		return nil
	}
	return c.database.Collection(name)
}
