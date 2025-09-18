package validator

import (
	"testing"

	"github.com/codemonsters/contract-tests-v2/parser"
	"github.com/stretchr/testify/assert"
)

// Helper функции для создания сообщений с массивами

func createMessageWithArrayField(fieldName, itemType string) *MessageInfo {
	payload := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			fieldName: map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": itemType,
				},
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

func createMessageWithArrayOfObjects(fieldName string, objectProps map[string]string) *MessageInfo {
	objectProperties := make(map[string]interface{})
	objectRequired := make([]string, 0, len(objectProps))
	
	for propName, propType := range objectProps {
		objectProperties[propName] = map[string]interface{}{"type": propType}
		objectRequired = append(objectRequired, propName)
	}
	
	payload := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			fieldName: map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type":       "object",
					"properties": objectProperties,
					"required":   objectRequired,
				},
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

func createMessageWithArrayRef(fieldName, ref string) *MessageInfo {
	payload := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			fieldName: map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"$ref": ref,
				},
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

// TestAreMessagesCompatible_Arrays тестирует совместимость массивов
func TestAreMessagesCompatible_Arrays(t *testing.T) {
	v := NewChannelValidator()
	spec := &parser.AsyncAPISpec{
		Components: &parser.Components{
			Schemas: map[string]parser.Schema{
				"User": {
					Type: "object",
					Properties: map[string]parser.Schema{
						"id":   {Type: "string"},
						"name": {Type: "string"},
					},
				},
				"Product": {
					Type: "object",
					Properties: map[string]parser.Schema{
						"id":    {Type: "string"},
						"title": {Type: "string"},
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		msg1     *MessageInfo
		msg2     *MessageInfo
		expected bool
	}{
		{
			name:     "same array of primitives should be compatible",
			msg1:     createMessageWithArrayField("tags", "string"),
			msg2:     createMessageWithArrayField("tags", "string"),
			expected: true,
		},
		{
			name:     "different array item types should be incompatible",
			msg1:     createMessageWithArrayField("values", "string"),
			msg2:     createMessageWithArrayField("values", "number"),
			expected: false, // Recursive validation now correctly detects different item types
		},
		{
			name:     "array vs non-array should be incompatible",
			msg1:     createMessageWithArrayField("data", "string"),
			msg2:     createMessageWithFieldTypes(map[string]string{"data": "string"}),
			expected: false,
		},
		{
			name: "same array of objects should be compatible",
			msg1: createMessageWithArrayOfObjects("users", map[string]string{
				"id":   "string",
				"name": "string",
			}),
			msg2: createMessageWithArrayOfObjects("users", map[string]string{
				"id":   "string",
				"name": "string",
			}),
			expected: true,
		},
		{
			name: "different array object structures should be incompatible",
			msg1: createMessageWithArrayOfObjects("users", map[string]string{
				"id":   "string",
				"name": "string",
			}),
			msg2: createMessageWithArrayOfObjects("users", map[string]string{
				"id":    "string",
				"email": "string", // different property
			}),
			expected: false, // Recursive validation now correctly detects different object structures in arrays
		},
		{
			name:     "same array refs should be compatible",
			msg1:     createMessageWithArrayRef("users", "#/components/schemas/User"),
			msg2:     createMessageWithArrayRef("users", "#/components/schemas/User"),
			expected: true,
		},
		{
			name:     "different array refs should be incompatible",
			msg1:     createMessageWithArrayRef("items", "#/components/schemas/User"),
			msg2:     createMessageWithArrayRef("items", "#/components/schemas/Product"),
			expected: false,
		},
		{
			name:     "array ref vs array inline object should be incompatible",
			msg1:     createMessageWithArrayRef("users", "#/components/schemas/User"),
			msg2:     createMessageWithArrayOfObjects("users", map[string]string{"id": "string"}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.areMessagesCompatible(tt.msg1, tt.msg2, spec, spec)
			assert.Equal(t, tt.expected, result)
		})
	}
}