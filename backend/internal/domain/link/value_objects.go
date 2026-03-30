package link

import "errors"

type LinkType struct {
	value string
}

func NewLinkType(value string) (LinkType, error) {
	switch value {
	case "reference", "dependency", "related", "custom":
		return LinkType{value: value}, nil
	default:
		return LinkType{}, errors.New("invalid link type")
	}
}

func (t LinkType) String() string {
	return t.value
}

type Weight struct {
	value float64
}

func NewWeight(value float64) (Weight, error) {
	if value < 0 || value > 1 {
		return Weight{}, errors.New("weight must be between 0 and 1")
	}
	return Weight{value: value}, nil
}

func (w Weight) Value() float64 {
	return w.value
}

// Metadata — дополнительные данные связи (тип связи, описание и т.п.)
type Metadata struct {
	value map[string]interface{}
}

func NewMetadata(value map[string]interface{}) (Metadata, error) {
	// Можно добавить валидацию, например, что поле "description" — строка и т.д.
	// Пока оставляем простую обёртку.
	return Metadata{value: value}, nil
}

func (m Metadata) Value() map[string]interface{} {
	return m.value
}
