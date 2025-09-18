package validator

import (
	"testing"

	"github.com/codemonsters/contract-tests-v2/parser"
	"github.com/stretchr/testify/assert"
)

// Helper функции для создания array свойств

func createArrayProperty(itemType string) map[string]interface{} {
	return map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": itemType,
		},
	}
}

func createArrayPropertyWithObjectItems(objectProps map[string]string) map[string]interface{} {
	itemProperties := make(map[string]interface{})
	itemRequired := make([]string, 0, len(objectProps))
	
	for propName, propType := range objectProps {
		itemProperties[propName] = map[string]interface{}{"type": propType}
		itemRequired = append(itemRequired, propName)
	}
	
	return map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type":       "object",
			"properties": itemProperties,
			"required":   itemRequired,
		},
	}
}

func createArrayPropertyWithRefItems(ref string) map[string]interface{} {
	return map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"$ref": ref,
		},
	}
}

// TestAreArrayItemsCompatible тестирует совместимость элементов массивов
func TestAreArrayItemsCompatible(t *testing.T) {
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
		prop1    interface{}
		prop2    interface{}
		expected bool
	}{
		{
			name:     "same primitive array items should be compatible",
			prop1:    createArrayProperty("string"),
			prop2:    createArrayProperty("string"),
			expected: true,
		},
		{
			name:     "different primitive array items should be incompatible",
			prop1:    createArrayProperty("string"),
			prop2:    createArrayProperty("number"),
			expected: false,
		},
		{
			name: "same object array items should be compatible",
			prop1: createArrayPropertyWithObjectItems(map[string]string{
				"id":   "string",
				"name": "string",
			}),
			prop2: createArrayPropertyWithObjectItems(map[string]string{
				"id":   "string",
				"name": "string",
			}),
			expected: true,
		},
		{
			name: "different object array items should be incompatible",
			prop1: createArrayPropertyWithObjectItems(map[string]string{
				"id":   "string",
				"name": "string",
			}),
			prop2: createArrayPropertyWithObjectItems(map[string]string{
				"id":    "string",
				"email": "string", // different property
			}),
			expected: false,
		},
		{
			name:     "same ref array items should be compatible",
			prop1:    createArrayPropertyWithRefItems("#/components/schemas/User"),
			prop2:    createArrayPropertyWithRefItems("#/components/schemas/User"),
			expected: true,
		},
		{
			name:     "different ref array items should be incompatible",
			prop1:    createArrayPropertyWithRefItems("#/components/schemas/User"),
			prop2:    createArrayPropertyWithRefItems("#/components/schemas/Product"),
			expected: false,
		},
		{
			name:     "ref vs inline object array items should be incompatible",
			prop1:    createArrayPropertyWithRefItems("#/components/schemas/User"),
			prop2:    createArrayPropertyWithObjectItems(map[string]string{"id": "string"}),
			expected: false,
		},
		{
			name:     "primitive vs object array items should be incompatible",
			prop1:    createArrayProperty("string"),
			prop2:    createArrayPropertyWithObjectItems(map[string]string{"value": "string"}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.areArrayItemsCompatible(tt.prop1, tt.prop2, spec, spec)
			assert.Equal(t, tt.expected, result)
		})
	}
}