package validator

import (
	"testing"

	"github.com/codemonsters/contract-tests-v2/parser"
	"github.com/stretchr/testify/assert"
)

// Helper функции для создания object свойств

func createObjectProperty(objProps map[string]string) map[string]interface{} {
	properties := make(map[string]interface{})
	required := make([]string, 0, len(objProps))
	
	for propName, propType := range objProps {
		properties[propName] = map[string]interface{}{"type": propType}
		required = append(required, propName)
	}
	
	return map[string]interface{}{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}

func createObjectPropertyWithRefs(objProps map[string]string) map[string]interface{} {
	properties := make(map[string]interface{})
	required := make([]string, 0, len(objProps))
	
	for propName, ref := range objProps {
		properties[propName] = map[string]interface{}{"$ref": ref}
		required = append(required, propName)
	}
	
	return map[string]interface{}{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}

func createObjectPropertyWithMixedTypes(props map[string]interface{}) map[string]interface{} {
	required := make([]string, 0, len(props))
	for propName := range props {
		required = append(required, propName)
	}
	
	return map[string]interface{}{
		"type":       "object",
		"properties": props,
		"required":   required,
	}
}

// TestAreObjectPropertiesCompatible тестирует совместимость объектов
func TestAreObjectPropertiesCompatible(t *testing.T) {
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
				"Profile": {
					Type: "object",
					Properties: map[string]parser.Schema{
						"age":   {Type: "integer"},
						"email": {Type: "string"},
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
			name: "same object properties should be compatible",
			prop1: createObjectProperty(map[string]string{
				"id":   "string",
				"name": "string",
			}),
			prop2: createObjectProperty(map[string]string{
				"id":   "string",
				"name": "string",
			}),
			expected: true,
		},
		{
			name: "different object properties should be incompatible",
			prop1: createObjectProperty(map[string]string{
				"id":   "string",
				"name": "string",
			}),
			prop2: createObjectProperty(map[string]string{
				"id":    "string",
				"email": "string", // different property
			}),
			expected: false,
		},
		{
			name: "different object property types should be incompatible",
			prop1: createObjectProperty(map[string]string{
				"id":  "string",
				"age": "integer",
			}),
			prop2: createObjectProperty(map[string]string{
				"id":  "string",
				"age": "string", // different type
			}),
			expected: false,
		},
		{
			name: "same refs in object properties should be compatible",
			prop1: createObjectPropertyWithRefs(map[string]string{
				"user":    "#/components/schemas/User",
				"profile": "#/components/schemas/Profile",
			}),
			prop2: createObjectPropertyWithRefs(map[string]string{
				"user":    "#/components/schemas/User",
				"profile": "#/components/schemas/Profile",
			}),
			expected: true,
		},
		{
			name: "different refs in object properties should be incompatible",
			prop1: createObjectPropertyWithRefs(map[string]string{
				"user": "#/components/schemas/User",
			}),
			prop2: createObjectPropertyWithRefs(map[string]string{
				"user": "#/components/schemas/Profile", // different ref
			}),
			expected: false,
		},
		{
			name: "ref vs inline object property should be incompatible",
			prop1: createObjectPropertyWithRefs(map[string]string{
				"user": "#/components/schemas/User",
			}),
			prop2: createObjectProperty(map[string]string{
				"user": "object",
			}),
			expected: false,
		},
		{
			name: "objects with different required fields count should be incompatible",
			prop1: createObjectProperty(map[string]string{
				"id":   "string",
				"name": "string",
				"age":  "integer",
			}),
			prop2: createObjectProperty(map[string]string{
				"id":   "string",
				"name": "string",
			}),
			expected: false,
		},
		{
			name: "empty objects should be compatible",
			prop1: createObjectProperty(map[string]string{}),
			prop2: createObjectProperty(map[string]string{}),
			expected: true,
		},
		{
			name: "nested objects should be validated recursively",
			prop1: createObjectPropertyWithMixedTypes(map[string]interface{}{
				"user": createObjectProperty(map[string]string{
					"id":   "string",
					"name": "string",
				}),
			}),
			prop2: createObjectPropertyWithMixedTypes(map[string]interface{}{
				"user": createObjectProperty(map[string]string{
					"id":   "string",
					"name": "string",
				}),
			}),
			expected: true,
		},
		{
			name: "different nested objects should be incompatible",
			prop1: createObjectPropertyWithMixedTypes(map[string]interface{}{
				"user": createObjectProperty(map[string]string{
					"id":   "string",
					"name": "string",
				}),
			}),
			prop2: createObjectPropertyWithMixedTypes(map[string]interface{}{
				"user": createObjectProperty(map[string]string{
					"id":    "string",
					"email": "string", // different nested property
				}),
			}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.areObjectPropertiesCompatible(tt.prop1, tt.prop2, spec, spec)
			assert.Equal(t, tt.expected, result)
		})
	}
}