package tag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid", "golang", false},
		{"empty", "", true},
		{"too long", string(make([]byte, 51)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag, err := New(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, tag)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, tag.ID())
			assert.Equal(t, tt.value, tag.Name())
			assert.False(t, tag.CreatedAt().IsZero())
		})
	}
}

func TestReconstruct(t *testing.T) {
	tag, err := New("initial")
	require.NoError(t, err)

	reconstructed, err := Reconstruct(tag.ID(), tag.Name(), tag.CreatedAt())
	require.NoError(t, err)
	assert.Equal(t, tag.ID(), reconstructed.ID())
	assert.Equal(t, tag.Name(), reconstructed.Name())
	assert.Equal(t, tag.CreatedAt(), reconstructed.CreatedAt())

	_, err = Reconstruct(tag.ID(), "", tag.CreatedAt())
	assert.Error(t, err)
}

func TestTag_Rename(t *testing.T) {
	tag, err := New("old")
	require.NoError(t, err)

	require.NoError(t, tag.Rename("new"))
	assert.Equal(t, "new", tag.Name())

	err = tag.Rename("")
	assert.Error(t, err)

	err = tag.Rename(string(make([]byte, 51)))
	assert.Error(t, err)
}
