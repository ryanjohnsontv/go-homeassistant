package utils_test

import (
	"testing"

	"github.com/ryanjohnsontv/go-homeassistant/shared/utils"
	"github.com/stretchr/testify/assert"
)

func TestJSONToGo_StringField(t *testing.T) {
	rawJSON := []byte(`{"key1": "value1"}`)
	structName := "Example"

	t.Run("StringField", func(t *testing.T) {
		generated, err := utils.JSONToGo(rawJSON, structName)
		assert.NoError(t, err)
		assert.Contains(t, generated, "Key1 string `json:\"key1\"`")
	})
}

func TestJSONToGo_NumberField(t *testing.T) {
	rawJSON := []byte(`{"key2": 123}`)
	structName := "Example"

	t.Run("NumberField", func(t *testing.T) {
		generated, err := utils.JSONToGo(rawJSON, structName)
		assert.NoError(t, err)
		assert.Contains(t, generated, "Key2 float64 `json:\"key2\"`")
	})
}

func TestJSONToGo_BooleanField(t *testing.T) {
	rawJSON := []byte(`{"key3": true}`)
	structName := "Example"

	t.Run("BooleanField", func(t *testing.T) {
		generated, err := utils.JSONToGo(rawJSON, structName)
		assert.NoError(t, err)
		assert.Contains(t, generated, "Key3 bool `json:\"key3\"`")
	})
}

func TestJSONToGo_NestedObject(t *testing.T) {
	rawJSON := []byte(`{"key4": {"nestedKey": "nestedValue"}}`)
	structName := "Example"

	t.Run("NestedObject", func(t *testing.T) {
		generated, err := utils.JSONToGo(rawJSON, structName)
		assert.NoError(t, err)
		assert.Contains(t, generated, "Key4 map[string]any `json:\"key4\"`")
	})
}

func TestJSONToGo_ArrayField(t *testing.T) {
	rawJSON := []byte(`{"key5": [1, 2, 3]}`)
	structName := "Example"

	t.Run("ArrayField", func(t *testing.T) {
		generated, err := utils.JSONToGo(rawJSON, structName)
		assert.NoError(t, err)
		assert.Contains(t, generated, "Key5 []any `json:\"key5\"`")
	})
}

func TestJSONToGo_InvalidJSON(t *testing.T) {
	rawJSON := []byte(`{"key1": "value1"`) // Missing closing brace
	structName := "Example"

	t.Run("InvalidJSON", func(t *testing.T) {
		_, err := utils.JSONToGo(rawJSON, structName)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse JSON")
	})
}
