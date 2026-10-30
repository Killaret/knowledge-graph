package note

import (
	"encoding/json"
	"fmt"
	"sort"
)

// NoteType is an immutable value object for the celestial body type of a note.
// Valid values and their ordering are defined by the canonical taxonomy.
type NoteType struct {
	value string
}

// typeDef holds the canonical metadata for a single note type.
type typeDef struct {
	scaleRank        int
	isUserSelectable bool
}

var typeRegistry = map[string]typeDef{
	"galaxy":             {scaleRank: 100, isUserSelectable: true},
	"nebula":             {scaleRank: 90, isUserSelectable: true},
	"blackhole":          {scaleRank: 85, isUserSelectable: true},
	"star":               {scaleRank: 80, isUserSelectable: true},
	"planet":             {scaleRank: 60, isUserSelectable: true},
	"moon":               {scaleRank: 50, isUserSelectable: true},
	"comet":              {scaleRank: 40, isUserSelectable: true},
	"satellite":          {scaleRank: 35, isUserSelectable: true},
	"asteroid":           {scaleRank: 30, isUserSelectable: true},
	"dust":               {scaleRank: 10, isUserSelectable: true},
	"debris":             {scaleRank: 5, isUserSelectable: true},
	"technical":          {scaleRank: 0, isUserSelectable: false},
	"unknown":            {scaleRank: 0, isUserSelectable: false},
	"reality_rift":       {scaleRank: 0, isUserSelectable: false},
	"chromatic_maw":      {scaleRank: 0, isUserSelectable: false},
	"void_whisper":       {scaleRank: 0, isUserSelectable: false},
	"cosmic_abomination": {scaleRank: 0, isUserSelectable: false},
}

// NewType validates and creates a NoteType. Empty or invalid values return an error.
func NewType(value string) (NoteType, error) {
	if value == "" {
		return NoteType{}, fmt.Errorf("note type cannot be empty")
	}
	if _, ok := typeRegistry[value]; !ok {
		return NoteType{}, fmt.Errorf("invalid note type: %q", value)
	}
	return NoteType{value: value}, nil
}

// NewTypeOrUnknown creates a NoteType from a raw string. Empty or invalid
// values fall back to the "unknown" type. This is intended for reconstruction
// from external storage where legacy data might be incomplete.
func NewTypeOrUnknown(value string) NoteType {
	if value == "" {
		return Unknown()
	}
	if _, ok := typeRegistry[value]; !ok {
		return Unknown()
	}
	return NoteType{value: value}
}

// MustType creates a NoteType and panics if the value is invalid.
// It is intended for tests and seed data where the value is a known constant.
func MustType(value string) NoteType {
	t, err := NewType(value)
	if err != nil {
		panic(err)
	}
	return t
}

// Default returns the default type for newly created notes.
func Default() NoteType {
	return NoteType{value: "star"}
}

// Unknown returns the fallback unknown type.
func Unknown() NoteType {
	return NoteType{value: "unknown"}
}

// String returns the raw type value.
func (t NoteType) String() string {
	return t.value
}

// IsValid reports whether the type is one of the canonical values.
func (t NoteType) IsValid() bool {
	_, ok := typeRegistry[t.value]
	return ok
}

// IsUserSelectable reports whether the type should be shown in user-facing selectors.
func (t NoteType) IsUserSelectable() bool {
	def, ok := typeRegistry[t.value]
	return ok && def.isUserSelectable
}

// ScaleRank returns the cosmic scale rank of the type (larger = broader scope).
// Invalid types return -1.
func (t NoteType) ScaleRank() int {
	def, ok := typeRegistry[t.value]
	if !ok {
		return -1
	}
	return def.scaleRank
}

// AllTypes returns all canonical note types sorted from largest to smallest.
func AllTypes() []NoteType {
	return sortedTypes(func(def typeDef) bool { return true })
}

// UserSelectableTypes returns all user-selectable types sorted from largest to smallest.
func UserSelectableTypes() []NoteType {
	return sortedTypes(func(def typeDef) bool { return def.isUserSelectable })
}

// AllTypeStrings returns the string values of all canonical types sorted from largest to smallest.
func AllTypeStrings() []string {
	return toStrings(AllTypes())
}

// UserSelectableTypeStrings returns the string values of all user-selectable types sorted from largest to smallest.
func UserSelectableTypeStrings() []string {
	return toStrings(UserSelectableTypes())
}

// MarshalJSON serializes the type as a JSON string.
func (t NoteType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.value)
}

// UnmarshalJSON reads the type from a JSON string.
func (t *NoteType) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	parsed, err := NewType(raw)
	if err != nil {
		return err
	}
	*t = parsed
	return nil
}

func sortedTypes(filter func(typeDef) bool) []NoteType {
	var types []NoteType
	for name, def := range typeRegistry {
		if filter(def) {
			types = append(types, NoteType{value: name})
		}
	}
	sort.Slice(types, func(i, j int) bool {
		rankI := types[i].ScaleRank()
		rankJ := types[j].ScaleRank()
		if rankI == rankJ {
			return types[i].value < types[j].value
		}
		return rankI > rankJ
	})
	return types
}

func toStrings(types []NoteType) []string {
	out := make([]string, len(types))
	for i, t := range types {
		out[i] = t.value
	}
	return out
}
