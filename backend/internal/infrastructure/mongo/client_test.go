package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient_InvalidURI(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, err := NewClient(ctx, "not-a-valid-uri", "testdb")
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestClientMethods_Nil(t *testing.T) {
	var c *Client
	assert.Nil(t, c.GetDatabase())
	assert.Nil(t, c.GetCollection("drafts"))
}
