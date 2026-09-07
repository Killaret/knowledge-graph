package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashToStringAndStringToHash(t *testing.T) {
	data := map[string]interface{}{"role": "admin", "active": true}

	str, err := HashToString(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, str)

	parsed, err := StringToHash(str)
	assert.NoError(t, err)
	assert.Equal(t, "admin", parsed["role"])
	assert.Equal(t, true, parsed["active"])
}

func TestStringToHash_InvalidJSON(t *testing.T) {
	_, err := StringToHash("not-json")
	assert.Error(t, err)
}

func TestHashToString_InvalidValue(t *testing.T) {
	_, err := HashToString(map[string]interface{}{"bad": make(chan int)})
	assert.Error(t, err)
}
