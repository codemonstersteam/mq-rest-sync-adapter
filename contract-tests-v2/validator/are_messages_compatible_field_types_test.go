package validator

import (
	"testing"

	"github.com/codemonsters/contract-tests-v2/parser"
	"github.com/stretchr/testify/assert"
)

// Helper функции для создания сообщений с разными типами полей

func createMessageWithFieldTypes(fieldTypes map[string]string) *MessageInfo {
	properties := make(map[string]interface{})
	requiredFields := make([]string, 0, len(fieldTypes))
	
	for fieldName, fieldType := range fieldTypes {
		properties[fieldName] = map[string]interface{}{"type": fieldType}
		requiredFields = append(requiredFields, fieldName)
	}
	
	payload := map[string]interface{}{
		"type":       "object",
		"properties": properties,
		"required":   requiredFields,
	}
	
	return &MessageInfo{
		Name:        "TestMessage",
		ContentType: "application/json",
		Payload:     payload,
	}
}

func createMessageWithFieldRef(fieldName, ref string) *MessageInfo {
	payload := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			fieldName: map[string]interface{}{"$ref": ref},
		},
		"required": []string{fieldName},
	}
	
	return &MessageInfo{
		Name:        "TestMessage",
		ContentType: "application/json",
		Payload:     payload,
	}
}

// TestAreMessagesCompatible_DifferentFieldTypes тестирует совместимость разных типов полей
func TestAreMessagesCompatible_DifferentFieldTypes(t *testing.T) {
	v := NewChannelValidator()
	spec := &parser.AsyncAPISpec{}

	tests := []struct {
		name     string
		msg1     *MessageInfo
		msg2     *MessageInfo
		expected bool
	}{
		{
			name: "same field types should be compatible",
			msg1: createMessageWithFieldTypes(map[string]string{
				"id":   "string",
				"age":  "integer",
				"name": "string",
			}),
			msg2: createMessageWithFieldTypes(map[string]string{
				"id":   "string",
				"age":  "integer", 
				"name": "string",
			}),
			expected: true,
		},
		{
			name: "different field types should be incompatible",
			msg1: createMessageWithFieldTypes(map[string]string{
				"id":  "string",
				"age": "integer",
			}),
			msg2: createMessageWithFieldTypes(map[string]string{
				"id":  "string",
				"age": "string", // integer -> string
			}),
			expected: false,
		},
		{
			name: "string vs number should be incompatible",
			msg1: createMessageWithFieldTypes(map[string]string{
				"value": "string",
			}),
			msg2: createMessageWithFieldTypes(map[string]string{
				"value": "number",
			}),
			expected: false,
		},
		{
			name: "integer vs number should be incompatible",
			msg1: createMessageWithFieldTypes(map[string]string{
				"count": "integer",
			}),
			msg2: createMessageWithFieldTypes(map[string]string{
				"count": "number",
			}),
			expected: false,
		},
		{
			name: "boolean vs string should be incompatible",
			msg1: createMessageWithFieldTypes(map[string]string{
				"isActive": "boolean",
			}),
			msg2: createMessageWithFieldTypes(map[string]string{
				"isActive": "string",
			}),
			expected: false,
		},
		{
			name: "array vs object should be incompatible",
			msg1: createMessageWithFieldTypes(map[string]string{
				"data": "array",
			}),
			msg2: createMessageWithFieldTypes(map[string]string{
				"data": "object",
			}),
			expected: false,
		},
		{
			name: "mixed compatible types",
			msg1: createMessageWithFieldTypes(map[string]string{
				"id":       "string",
				"count":    "integer",
				"price":    "number",
				"active":   "boolean",
				"tags":     "array",
				"metadata": "object",
			}),
			msg2: createMessageWithFieldTypes(map[string]string{
				"id":       "string",
				"count":    "integer",
				"price":    "number",
				"active":   "boolean",
				"tags":     "array",
				"metadata": "object",
			}),
			expected: true,
		},
		{
			name: "one field type mismatch in many fields",
			msg1: createMessageWithFieldTypes(map[string]string{
				"id":       "string",
				"count":    "integer",
				"price":    "number",
				"active":   "boolean",
			}),
			msg2: createMessageWithFieldTypes(map[string]string{
				"id":       "string",
				"count":    "string", // Mismatch: integer -> string
				"price":    "number",
				"active":   "boolean",
			}),
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

// TestAreMessagesCompatible_RefFields тестирует поля с $ref ссылками
func TestAreMessagesCompatible_RefFields(t *testing.T) {
	v := NewChannelValidator()
	
	// Создаем spec с компонентами для $ref ссылок
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
				"Admin": {
					Type: "object", 
					Properties: map[string]parser.Schema{
						"id":    {Type: "string"},
						"role":  {Type: "string"},
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
			name:     "same ref should be compatible",
			msg1:     createMessageWithFieldRef("user", "#/components/schemas/User"),
			msg2:     createMessageWithFieldRef("user", "#/components/schemas/User"),
			expected: true,
		},
		{
			name:     "different refs should be incompatible",
			msg1:     createMessageWithFieldRef("user", "#/components/schemas/User"),
			msg2:     createMessageWithFieldRef("user", "#/components/schemas/Admin"),
			expected: false,
		},
		{
			name: "ref vs inline type should be incompatible",
			msg1: createMessageWithFieldRef("user", "#/components/schemas/User"),
			msg2: createMessageWithFieldTypes(map[string]string{
				"user": "object",
			}),
			expected: false,
		},
		{
			name: "inline type vs ref should be incompatible",
			msg1: createMessageWithFieldTypes(map[string]string{
				"user": "string",
			}),
			msg2: createMessageWithFieldRef("user", "#/components/schemas/User"),
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