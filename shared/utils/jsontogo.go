package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Dynamically generates a Go struct from JSON data.
func JSONToGo(rawJSON []byte, structName string) (string, error) {
	var parsed map[string]any

	err := json.Unmarshal(rawJSON, &parsed)
	if err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("type %s struct {\n", structName))

	for key, value := range parsed {
		fieldName := toCamelCase(key)
		fieldType := inferType(value)
		buffer.WriteString(fmt.Sprintf("    %s %s `json:\"%s\"`\n", fieldName, fieldType, key))
	}

	buffer.WriteString("}\n")

	return buffer.String(), nil
}

func toCamelCase(input string) string {
	parts := strings.Split(input, "_")
	for i := range parts {
		caser := cases.Title(language.AmericanEnglish)
		parts[i] = caser.String(parts[i])
	}

	return strings.Join(parts, "")
}

func inferType(value any) string {
	switch v := value.(type) {
	case float64:
		return "float64"
	case string:
		return "string"
	case bool:
		return "bool"
	case map[string]any:
		return "map[string]any"
	case []any:
		return "[]any"
	default:
		return fmt.Sprintf("unknown (type: %T)", v)
	}
}
