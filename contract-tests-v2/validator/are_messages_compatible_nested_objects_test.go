package validator

import (
	"testing"

	"github.com/codemonsters/contract-tests-v2/parser"
	"github.com/stretchr/testify/assert"
)

// Helper функции для создания сообщений с вложенными объектами

func createMessageWithNestedObject(fieldName string, nestedProps map[string]string) *MessageInfo {
	nestedProperties := make(map[string]interface{})
	nestedRequired := make([]string, 0, len(nestedProps))
	
	for propName, propType := range nestedProps {
		nestedProperties[propName] = map[string]interface{}{"type": propType}
		nestedRequired = append(nestedRequired, propName)
	}
	
	payload := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			fieldName: map[string]interface{}{
				"type":       "object",
				"properties": nestedProperties,
				"required":   nestedRequired,
			},
		},
		"required": []string{fieldName},
	}
	
	return &MessageInfo{
		Name:        "TestMessage",
		ContentType: "application/json",
		Payload:     payload,
	}
}

func createMessageWithDeeplyNestedObject() *MessageInfo {
	payload := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"user": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"profile": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name": map[string]interface{}{"type": "string"},
							"age":  map[string]interface{}{"type": "integer"},
						},
						"required": []string{"name", "age"},
					},
					"settings": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"theme": map[string]interface{}{"type": "string"},
						},
						"required": []string{"theme"},
					},
				},
				"required": []string{"profile", "settings"},
			},
		},
		"required": []string{"user"},
	}
	
	return &MessageInfo{
		Name:        "TestMessage",
		ContentType: "application/json",
		Payload:     payload,
	}
}

// TestAreMessagesCompatible_NestedObjects тестирует совместимость вложенных объектов
func TestAreMessagesCompatible_NestedObjects(t *testing.T) {
	v := NewChannelValidator()
	spec := &parser.AsyncAPISpec{}

	tests := []struct {
		name     string
		msg1     *MessageInfo
		msg2     *MessageInfo
		expected bool
	}{
		{
			name: "same nested object structure should be compatible",
			msg1: createMessageWithNestedObject("address", map[string]string{
				"street": "string",
				"city":   "string",
			}),
			msg2: createMessageWithNestedObject("address", map[string]string{
				"street": "string",
				"city":   "string",
			}),
			expected: true,
		},
		{
			name: "different nested object properties should be incompatible",
			msg1: createMessageWithNestedObject("address", map[string]string{
				"street": "string",
				"city":   "string",
			}),
			msg2: createMessageWithNestedObject("address", map[string]string{
				"street": "string",
				"country": "string", // different property
			}),
			expected: false, // Recursive validation now correctly detects incompatible nested structures
		},
		{
			name: "different nested object property types should be incompatible",
			msg1: createMessageWithNestedObject("coordinates", map[string]string{
				"x": "number",
				"y": "number",
			}),
			msg2: createMessageWithNestedObject("coordinates", map[string]string{
				"x": "string", // different type
				"y": "number",
			}),
			expected: false, // Recursive validation now correctly detects type mismatches
		},
		{
			name: "deeply nested objects with same structure should be compatible",
			msg1: createMessageWithDeeplyNestedObject(),
			msg2: createMessageWithDeeplyNestedObject(),
			expected: true,
		},
		{
			name: "empty nested object should be compatible with another empty nested object",
			msg1: createMessageWithNestedObject("metadata", map[string]string{}),
			msg2: createMessageWithNestedObject("metadata", map[string]string{}),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.areMessagesCompatible(tt.msg1, tt.msg2, spec, spec)
			assert.Equal(t, tt.expected, result)
		})
	}
}