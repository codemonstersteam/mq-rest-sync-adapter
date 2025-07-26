package validator

import (
	"testing"

	"github.com/codemonsters/contract-tests-v2/parser"
	"github.com/stretchr/testify/assert"
)

// Helper functions for creating test schemas

func createNilSchema() *parser.Schema {
	return nil
}

func createSchemaWithRef(ref string) *parser.Schema {
	return &parser.Schema{
		Ref: ref,
	}
}

func createSimpleStringSchema() *parser.Schema {
	return &parser.Schema{
		Type: "string",
	}
}

func createSchemaWithFormat(schemaType, format string) *parser.Schema {
	return &parser.Schema{
		Type:   schemaType,
		Format: format,
	}
}

func createSchemaWithEnum(schemaType string, enum []interface{}) *parser.Schema {
	return &parser.Schema{
		Type: schemaType,
		Enum: enum,
	}
}

func createArraySchemaWithItems() *parser.Schema {
	return &parser.Schema{
		Type: "array",
		Items: &parser.Schema{
			Type: "string",
		},
	}
}

func createObjectSchemaWithProperties() *parser.Schema {
	return &parser.Schema{
		Type: "object",
		Properties: map[string]parser.Schema{
			"userId": {Type: "string"},
			"email":  {Type: "string"},
			"age":    {Type: "integer"},
		},
	}
}

func createSchemaWithRequired(required []string) *parser.Schema {
	return &parser.Schema{
		Type: "object",
		Properties: map[string]parser.Schema{
			"userId": {Type: "string"},
			"email":  {Type: "string"},
		},
		Required: required,
	}
}

func createComplexSchemaWithRefInProperties() *parser.Schema {
	return &parser.Schema{
		Type: "object",
		Properties: map[string]parser.Schema{
			"user": {Ref: "#/components/schemas/User"},
			"data": {Type: "string"},
		},
		Required: []string{"user", "data"},
	}
}

// Helper functions for creating expected results

func createExpectedNilResult() map[string]interface{} {
	return nil
}

func createExpectedRefResult(ref string) map[string]interface{} {
	return map[string]interface{}{
		"$ref": ref,
	}
}

func createExpectedStringResult() map[string]interface{} {
	return map[string]interface{}{
		"type": "string",
	}
}

func createExpectedResultWithFormat(schemaType, format string) map[string]interface{} {
	return map[string]interface{}{
		"type":   schemaType,
		"format": format,
	}
}

func createExpectedResultWithEnum(schemaType string, enum []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type": schemaType,
		"enum": enum,
	}
}

func createExpectedArrayResult() map[string]interface{} {
	return map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": "string",
		},
	}
}

func createExpectedObjectResult() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"userId": map[string]interface{}{"type": "string"},
			"email":  map[string]interface{}{"type": "string"},
			"age":    map[string]interface{}{"type": "integer"},
		},
	}
}

func createExpectedResultWithRequired(required []string) map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"userId": map[string]interface{}{"type": "string"},
			"email":  map[string]interface{}{"type": "string"},
		},
		"required": required,
	}
}

func createExpectedComplexResult() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"user": map[string]interface{}{"$ref": "#/components/schemas/User"},
			"data": map[string]interface{}{"type": "string"},
		},
		"required": []string{"user", "data"},
	}
}

func TestConvertSchema(t *testing.T) {
	v := NewChannelValidator()

	tests := []struct {
		name     string
		schema   *parser.Schema
		expected map[string]interface{}
	}{
		{
			name:     "nil schema should return nil",
			schema:   createNilSchema(),
			expected: createExpectedNilResult(),
		},
		{
			name:     "schema with ref should return only ref (CRITICAL - THIS WAS THE BUG)",
			schema:   createSchemaWithRef("#/components/schemas/User"),
			expected: createExpectedRefResult("#/components/schemas/User"),
		},
		{
			name:     "simple string schema",
			schema:   createSimpleStringSchema(),
			expected: createExpectedStringResult(),
		},
		{
			name:     "integer schema",
			schema:   &parser.Schema{Type: "integer"},
			expected: map[string]interface{}{"type": "integer"},
		},
		{
			name:     "number schema",
			schema:   &parser.Schema{Type: "number"},
			expected: map[string]interface{}{"type": "number"},
		},
		{
			name:     "boolean schema",
			schema:   &parser.Schema{Type: "boolean"},
			expected: map[string]interface{}{"type": "boolean"},
		},
		{
			name:     "array schema",
			schema:   &parser.Schema{Type: "array"},
			expected: map[string]interface{}{"type": "array"},
		},
		{
			name:     "schema with format int64",
			schema:   createSchemaWithFormat("integer", "int64"),
			expected: createExpectedResultWithFormat("integer", "int64"),
		},
		{
			name:     "schema with format double",
			schema:   createSchemaWithFormat("number", "double"),
			expected: createExpectedResultWithFormat("number", "double"),
		},
		{
			name:     "schema with format email",
			schema:   createSchemaWithFormat("string", "email"),
			expected: createExpectedResultWithFormat("string", "email"),
		},
		{
			name:     "schema with format uri",
			schema:   createSchemaWithFormat("string", "uri"),
			expected: createExpectedResultWithFormat("string", "uri"),
		},
		{
			name:     "schema with string enum",
			schema:   createSchemaWithEnum("string", []interface{}{"success", "error"}),
			expected: createExpectedResultWithEnum("string", []interface{}{"success", "error"}),
		},
		{
			name:     "schema with integer enum",
			schema:   createSchemaWithEnum("integer", []interface{}{1, 2, 3}),
			expected: createExpectedResultWithEnum("integer", []interface{}{1, 2, 3}),
		},
		{
			name:     "array schema with items",
			schema:   createArraySchemaWithItems(),
			expected: createExpectedArrayResult(),
		},
		{
			name:     "object schema with properties",
			schema:   createObjectSchemaWithProperties(),
			expected: createExpectedObjectResult(),
		},
		{
			name:     "schema with required fields",
			schema:   createSchemaWithRequired([]string{"userId", "email"}),
			expected: createExpectedResultWithRequired([]string{"userId", "email"}),
		},
		{
			name:   "schema with empty required array",
			schema: createSchemaWithRequired([]string{}),
			expected: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userId": map[string]interface{}{"type": "string"},
					"email":  map[string]interface{}{"type": "string"},
				},
				// Пустой required массив не должен появляться в результате
			},
		},
		{
			name:     "complex schema with ref in properties and required fields",
			schema:   createComplexSchemaWithRefInProperties(),
			expected: createExpectedComplexResult(),
		},
		{
			name:   "schema with nested ref should prioritize top-level ref",
			schema: &parser.Schema{
				Ref:  "#/components/schemas/TopLevel",
				Type: "object", // Должно игнорироваться при наличии Ref
				Properties: map[string]parser.Schema{
					"nested": {Type: "string"}, // Должно игнорироваться при наличии Ref
				},
			},
			expected: map[string]interface{}{
				"$ref": "#/components/schemas/TopLevel",
			},
		},
		{
			name:     "empty object schema",
			schema:   &parser.Schema{Type: "object"},
			expected: map[string]interface{}{"type": "object"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := v.convertSchema(tt.schema)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestConvertSchema_EdgeCases проверяет граничные случаи
func TestConvertSchema_EdgeCases(t *testing.T) {
	v := NewChannelValidator()

	t.Run("schema with empty string type", func(t *testing.T) {
		schema := &parser.Schema{Type: ""}
		result := v.convertSchema(schema)
		expected := map[string]interface{}{"type": ""}
		assert.Equal(t, expected, result)
	})

	t.Run("schema with only format field", func(t *testing.T) {
		schema := &parser.Schema{Format: "int64"}
		result := v.convertSchema(schema)
		expected := map[string]interface{}{
			"type":   "",
			"format": "int64",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("schema with empty properties map", func(t *testing.T) {
		schema := &parser.Schema{
			Type:       "object",
			Properties: map[string]parser.Schema{},
		}
		result := v.convertSchema(schema)
		expected := map[string]interface{}{"type": "object"}
		assert.Equal(t, expected, result)
	})

	t.Run("schema with deeply nested properties", func(t *testing.T) {
		schema := &parser.Schema{
			Type: "object",
			Properties: map[string]parser.Schema{
				"level1": {
					Type: "object",
					Properties: map[string]parser.Schema{
						"level2": {
							Type: "object",
							Properties: map[string]parser.Schema{
								"level3": {Type: "string"},
							},
						},
					},
				},
			},
		}
		result := v.convertSchema(schema)
		
		expected := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"level1": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"level2": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"level3": map[string]interface{}{"type": "string"},
							},
						},
					},
				},
			},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("schema with empty enum array", func(t *testing.T) {
		schema := &parser.Schema{
			Type: "string",
			Enum: []interface{}{},
		}
		result := v.convertSchema(schema)
		expected := map[string]interface{}{"type": "string"}
		assert.Equal(t, expected, result)
	})

	t.Run("array schema with complex items", func(t *testing.T) {
		schema := &parser.Schema{
			Type: "array",
			Items: &parser.Schema{
				Type: "object",
				Properties: map[string]parser.Schema{
					"id":   {Type: "integer"},
					"name": {Type: "string"},
				},
				Required: []string{"id"},
			},
		}
		result := v.convertSchema(schema)
		
		expected := map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id":   map[string]interface{}{"type": "integer"},
					"name": map[string]interface{}{"type": "string"},
				},
				"required": []string{"id"},
			},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("schema with minimum and maximum", func(t *testing.T) {
		schema := &parser.Schema{
			Type: "integer",
			Additional: map[string]interface{}{
				"minimum": 0,
				"maximum": 100,
			},
		}
		result := v.convertSchema(schema)
		
		expected := map[string]interface{}{
			"type":    "integer",
			"minimum": 0,
			"maximum": 100,
		}
		assert.Equal(t, expected, result)
	})

	t.Run("schema with exclusiveMinimum and exclusiveMaximum", func(t *testing.T) {
		schema := &parser.Schema{
			Type: "number",
			Additional: map[string]interface{}{
				"minimum":          0,
				"maximum":          100,
				"exclusiveMinimum": true,
				"exclusiveMaximum": true,
			},
		}
		result := v.convertSchema(schema)
		
		expected := map[string]interface{}{
			"type":             "number",
			"minimum":          0,
			"maximum":          100,
			"exclusiveMinimum": true,
			"exclusiveMaximum": true,
		}
		assert.Equal(t, expected, result)
	})

	t.Run("schema with minLength and maxLength", func(t *testing.T) {
		schema := &parser.Schema{
			Type: "string",
			Additional: map[string]interface{}{
				"minLength": 5,
				"maxLength": 20,
			},
		}
		result := v.convertSchema(schema)
		
		expected := map[string]interface{}{
			"type":      "string",
			"minLength": 5,
			"maxLength": 20,
		}
		assert.Equal(t, expected, result)
	})

	t.Run("schema with pattern for email validation", func(t *testing.T) {
		schema := &parser.Schema{
			Type: "string",
			Additional: map[string]interface{}{
				"pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
			},
		}
		result := v.convertSchema(schema)
		
		expected := map[string]interface{}{
			"type":    "string",
			"pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("schema with combined format and enum", func(t *testing.T) {
		schema := &parser.Schema{
			Type:   "integer",
			Format: "int64",
			Enum:   []interface{}{1, 2, 3},
		}
		result := v.convertSchema(schema)
		
		expected := map[string]interface{}{
			"type":   "integer",
			"format": "int64",
			"enum":   []interface{}{1, 2, 3},
		}
		assert.Equal(t, expected, result)
	})
}

// TestConvertSchema_BugRegression тестирует конкретный баг, который был исправлен
func TestConvertSchema_BugRegression(t *testing.T) {
	v := NewChannelValidator()

	t.Run("REGRESSION: ref field should be preserved as $ref", func(t *testing.T) {
		// Этот тест проверяет критический баг, который был исправлен
		// До исправления: поле Ref терялось при конвертации
		// После исправления: поле Ref сохраняется как $ref
		
		schema := &parser.Schema{
			Ref: "#/components/schemas/walletBalanceData",
		}
		
		result := v.convertSchema(schema)
		
		// Проверяем что результат содержит именно $ref
		assert.Contains(t, result, "$ref")
		assert.Equal(t, "#/components/schemas/walletBalanceData", result["$ref"])
		
		// Проверяем что других полей нет (ref имеет приоритет)
		assert.Len(t, result, 1)
		assert.NotContains(t, result, "type")
		assert.NotContains(t, result, "properties")
	})

	t.Run("REGRESSION: complex schema with refs in data field", func(t *testing.T) {
		// Реальный случай из багрепорта - схема с ref в поле data
		schema := &parser.Schema{
			Type: "object",
			Properties: map[string]parser.Schema{
				"status": {Type: "string"},
				"actualTimestamp": {Type: "integer"},
				"data": {Ref: "#/components/schemas/walletBalanceData"}, // Критичная ссылка
			},
			Required: []string{"status", "actualTimestamp", "data"},
		}
		
		result := v.convertSchema(schema)
		
		// Проверяем что поле data содержит $ref
		properties := result["properties"].(map[string]interface{})
		dataField := properties["data"].(map[string]interface{})
		
		assert.Equal(t, "#/components/schemas/walletBalanceData", dataField["$ref"])
		assert.NotContains(t, dataField, "type")
	})
}