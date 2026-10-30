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

func TestConnect_InvalidDSN(t *testing.T) {
	db, err := Connect("not a valid dsn")
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to connect")
}

func TestGetPoolStats_InvalidDB(t *testing.T) {
	stats := GetPoolStats(nil)
	assert.Contains(t, stats["error"], "nil")
}

func TestNormalizePoolConfig(t *testing.T) {
	defaults := DefaultPoolConfig()

	tests := []struct {
		name string
		pool PoolConfig
		want PoolConfig
	}{
		{"zero value falls back entirely", PoolConfig{}, defaults},
		{"negative values fall back", PoolConfig{-1, -5, -300, -60}, defaults},
		{"custom values kept", PoolConfig{50, 10, 600, 120}, PoolConfig{50, 10, 600, 120}},
		{"partial override keeps defaults for unset fields", PoolConfig{MaxOpenConns: 50}, PoolConfig{50, defaults.MaxIdleConns, defaults.ConnMaxLifetimeSeconds, defaults.ConnMaxIdleTimeSeconds}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizePoolConfig(tt.pool))
		})
	}
}

func TestConnectWithPool_EmptyDSN(t *testing.T) {
	db, err := ConnectWithPool("", DefaultPoolConfig())
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "DSN is empty")
}
