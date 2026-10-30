package note

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewType(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid star", "star", false},
		{"valid galaxy", "galaxy", false},
		{"valid moon", "moon", false},
		{"valid debris", "debris", false},
		{"empty", "", true},
		{"invalid", "banana", true},
		{"case sensitive uppercase", "STAR", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nt, err := NewType(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, nt.String())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.value, nt.String())
		})
	}
}

func TestNewTypeOrUnknown(t *testing.T) {
	assert.Equal(t, "star", NewTypeOrUnknown("star").String())
	assert.Equal(t, "unknown", NewTypeOrUnknown("").String())
	assert.Equal(t, "unknown", NewTypeOrUnknown("banana").String())
}

func TestMustType(t *testing.T) {
	assert.Equal(t, "star", MustType("star").String())
	assert.Panics(t, func() { MustType("banana") })
}

func TestDefaultAndUnknown(t *testing.T) {
	assert.Equal(t, "star", Default().String())
	assert.Equal(t, "unknown", Unknown().String())
}

func TestNoteTypeProperties(t *testing.T) {
	star, err := NewType("star")
	require.NoError(t, err)
	assert.True(t, star.IsValid())
	assert.True(t, star.IsUserSelectable())
	assert.Equal(t, 80, star.ScaleRank())

	unknown := Unknown()
	assert.True(t, unknown.IsValid())
	assert.False(t, unknown.IsUserSelectable())
	assert.Equal(t, 0, unknown.ScaleRank())

	invalid := NoteType{}
	assert.False(t, invalid.IsValid())
	assert.False(t, invalid.IsUserSelectable())
	assert.Equal(t, -1, invalid.ScaleRank())

	technical, err := NewType("technical")
	require.NoError(t, err)
	assert.True(t, technical.IsValid())
	assert.False(t, technical.IsUserSelectable())
}

func TestAllTypesOrdering(t *testing.T) {
	all := AllTypeStrings()
	want := []string{
		"galaxy", "nebula", "blackhole", "star", "planet", "moon",
		"comet", "satellite", "asteroid", "dust", "debris",
		"chromatic_maw", "cosmic_abomination", "reality_rift",
		"technical", "unknown", "void_whisper",
	}
	assert.Equal(t, want, all)
}

func TestUserSelectableTypes(t *testing.T) {
	ui := UserSelectableTypeStrings()
	assert.Equal(t, []string{
		"galaxy", "nebula", "blackhole", "star", "planet", "moon",
		"comet", "satellite", "asteroid", "dust", "debris",
	}, ui)
	assert.NotContains(t, ui, "technical")
	assert.NotContains(t, ui, "unknown")
	assert.NotContains(t, ui, "reality_rift")
}

func TestNoteTypeJSON(t *testing.T) {
	nt := MustType("planet")
	data, err := json.Marshal(nt)
	require.NoError(t, err)
	assert.Equal(t, `"planet"`, string(data))

	var parsed NoteType
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Equal(t, "planet", parsed.String())

	var invalid NoteType
	assert.Error(t, json.Unmarshal([]byte(`"banana"`), &invalid))
}
