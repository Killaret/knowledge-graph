package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnect_EmptyDSN(t *testing.T) {
	db, err := Connect("")
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "DSN is empty")
}

func TestGetPoolStats_InvalidDB(t *testing.T) {
	stats := GetPoolStats(nil)
	assert.Contains(t, stats["error"], "nil")
}
