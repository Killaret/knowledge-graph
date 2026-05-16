package mongo

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client wraps the MongoDB client
type Client struct {
	client   *mongo.Client
	database string
}

// Config holds MongoDB configuration
type Config struct {
	URI      string
	Database string
}

// NewClient creates a new MongoDB client
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.URI == "" {
		return nil, fmt.Errorf("MongoDB URI is required")
	}
	if cfg.Database == "" {
		return nil, fmt.Errorf("MongoDB database name is required")
	}

	// Create client options
	clientOpts := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(10 * time.Minute)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Printf("[Mongo] Connected to MongoDB: %s", cfg.Database)

	return &Client{
		client:   client,
		database: cfg.Database,
	}, nil
}

// Close closes the MongoDB connection
func (c *Client) Close(ctx context.Context) error {
	if err := c.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}
	log.Printf("[Mongo] Disconnected from MongoDB")
	return nil
}

// Database returns the MongoDB database
func (c *Client) Database() *mongo.Database {
	return c.client.Database(c.database)
}

// Collection returns a MongoDB collection
func (c *Client) Collection(name string) *mongo.Collection {
	return c.client.Database(c.database).Collection(name)
}

// Client returns the underlying MongoDB client
func (c *Client) Client() *mongo.Client {
	return c.client
}
