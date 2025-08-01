package validator

import (
	"testing"
)

func TestGetPropertyType(t *testing.T) {
	validator := &ChannelValidator{}

	// Базовые тест-кейсы
	t.Run("Базовые тест-кейсы", func(t *testing.T) {
		tests := []struct {
			name     string
			prop     interface{}
			expected string
		}{
			{
				name: "валидный type (string)",
				prop: map[string]interface{}{
					"type": "string",
				},
				expected: "string",
			},
			{
				name: "свойство без type (возвращает пустую строку)",
				prop: map[string]interface{}{
					"format": "email",
					"description": "User email",
				},
				expected: "",
			},
			{
				name:     "nil свойство (безопасная обработка)",
				prop:     nil,
				expected: "",
			},
			{
				name: "неправильный тип type (не string)",
				prop: map[string]interface{}{
					"type": 123,
				},
				expected: "",
			},
			{
				name: "пустой type (type = \"\")",
				prop: map[string]interface{}{
					"type": "",
				},
				expected: "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := validator.getPropertyType(tt.prop)
				if result != tt.expected {
					t.Errorf("getPropertyType() = %v, expected %v", result, tt.expected)
				}
			})
		}
	})

	// AsyncAPI 3.0 валидные JSON Schema типы
	t.Run("AsyncAPI 3.0 валидные JSON Schema типы", func(t *testing.T) {
		tests := []struct {
			name     string
			prop     interface{}
			expected string
		}{
			{
				name: "примитивный тип: string",
				prop: map[string]interface{}{
					"type": "string",
				},
				expected: "string",
			},
			{
				name: "примитивный тип: number",
				prop: map[string]interface{}{
					"type": "number",
				},
				expected: "number",
			},
			{
				name: "примитивный тип: integer",
				prop: map[string]interface{}{
					"type": "integer",
				},
				expected: "integer",
			},
			{
				name: "примитивный тип: boolean",
				prop: map[string]interface{}{
					"type": "boolean",
				},
				expected: "boolean",
			},
			{
				name: "примитивный тип: null",
				prop: map[string]interface{}{
					"type": "null",
				},
				expected: "null",
			},
			{
				name: "составной тип: object",
				prop: map[string]interface{}{
					"type": "object",
				},
				expected: "object",
			},
			{
				name: "составной тип: array",
				prop: map[string]interface{}{
					"type": "array",
				},
				expected: "array",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := validator.getPropertyType(tt.prop)
				if result != tt.expected {
					t.Errorf("getPropertyType() = %v, expected %v", result, tt.expected)
				}
			})
		}
	})

	// AsyncAPI 3.0 специфичные кейсы
	t.Run("AsyncAPI 3.0 специфичные кейсы", func(t *testing.T) {
		tests := []struct {
			name     string
			prop     interface{}
			expected string
		}{
			{
				name: "case sensitivity: String vs string (должно распознавать точно)",
				prop: map[string]interface{}{
					"type": "String", // Неправильный регистр
				},
				expected: "String", // Функция возвращает как есть
			},
			{
				name: "тип с format модификатором",
				prop: map[string]interface{}{
					"type":   "string",
					"format": "email",
				},
				expected: "string", // format не влияет на type
			},
			{
				name: "приоритет полей: type и $ref присутствуют",
				prop: map[string]interface{}{
					"type": "string",
					"$ref": "#/components/schemas/User",
				},
				expected: "string", // Должен вернуть type, не $ref
			},
			{
				name: "свойство с дополнительными полями",
				prop: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{"type": "string"},
					},
					"required": []string{"name"},
				},
				expected: "object",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := validator.getPropertyType(tt.prop)
				if result != tt.expected {
					t.Errorf("getPropertyType() = %v, expected %v", result, tt.expected)
				}
			})
		}
	})

	// Edge cases и ошибки
	t.Run("Edge cases и ошибки", func(t *testing.T) {
		tests := []struct {
			name     string
			prop     interface{}
			expected string
		}{
			{
				name: "невалидный тип: datetime (не JSON Schema тип)",
				prop: map[string]interface{}{
					"type": "datetime",
				},
				expected: "datetime", // Функция возвращает как есть, валидация на уровне выше
			},
			{
				name: "невалидный тип: uuid (не JSON Schema тип)",
				prop: map[string]interface{}{
					"type": "uuid",
				},
				expected: "uuid", // Функция возвращает как есть
			},
			{
				name: "невалидный тип: unknown",
				prop: map[string]interface{}{
					"type": "unknown",
				},
				expected: "unknown", // Функция возвращает как есть
			},
			{
				name: "type как число",
				prop: map[string]interface{}{
					"type": 123,
				},
				expected: "", // type должен быть string
			},
			{
				name: "type как boolean",
				prop: map[string]interface{}{
					"type": true,
				},
				expected: "", // type должен быть string
			},
			{
				name: "type как массив (недопустимо в AsyncAPI)",
				prop: map[string]interface{}{
					"type": []string{"string", "null"},
				},
				expected: "", // type должен быть string, не массив
			},
			{
				name:     "невалидное свойство: не map[string]interface{}",
				prop:     "invalid property type",
				expected: "",
			},
			{
				name:     "пустой map",
				prop:     map[string]interface{}{},
				expected: "",
			},
			{
				name: "map с другими полями но без type",
				prop: map[string]interface{}{
					"format": "email",
					"enum":   []string{"user@example.com"},
					"minLength": 5,
				},
				expected: "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := validator.getPropertyType(tt.prop)
				if result != tt.expected {
					t.Errorf("getPropertyType() = %v, expected %v", result, tt.expected)
				}
			})
		}
	})

	// Дополнительные тесты для полного покрытия
	t.Run("Дополнительные edge cases", func(t *testing.T) {
		tests := []struct {
			name     string
			prop     interface{}
			expected string
		}{
			{
				name: "type с null значением",
				prop: map[string]interface{}{
					"type": nil,
				},
				expected: "", // type должен быть string
			},
			{
				name: "вложенный объект с type",
				prop: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"nested": map[string]interface{}{
							"type": "string",
						},
					},
				},
				expected: "object", // Извлекаем только верхний уровень
			},
			{
				name: "type с дополнительными пробелами",
				prop: map[string]interface{}{
					"type": " string ",
				},
				expected: " string ", // Функция возвращает как есть, trim на уровне выше
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := validator.getPropertyType(tt.prop)
				if result != tt.expected {
					t.Errorf("getPropertyType() = %v, expected %v", result, tt.expected)
				}
			})
		}
	})
}

// Тест для проверки всех стандартных JSON Schema типов в одном месте
func TestGetPropertyType_AllJSONSchemaTypes(t *testing.T) {
	validator := &ChannelValidator{}

	// Все валидные JSON Schema типы согласно AsyncAPI 3.0
	validTypes := []string{
		"string", "number", "integer", "boolean", "null", "object", "array",
	}

	for _, validType := range validTypes {
		t.Run("json_schema_type_"+validType, func(t *testing.T) {
			prop := map[string]interface{}{
				"type": validType,
			}
			result := validator.getPropertyType(prop)
			if result != validType {
				t.Errorf("Expected %s, got %s", validType, result)
			}
		})
	}
}

// Тест для проверки безопасности type assertion
func TestGetPropertyType_TypeSafety(t *testing.T) {
	validator := &ChannelValidator{}

	// Тестируем различные невалидные типы, которые могут вызвать панику
	invalidTypes := []interface{}{
		123,
		true,
		[]string{"test"},
		map[int]string{1: "test"},
		struct{ Field string }{Field: "test"},
		func() {},
	}

	for i, invalidType := range invalidTypes {
		t.Run("invalid_type_"+string(rune(i+'0')), func(t *testing.T) {
			// Этот тест должен пройти без паники
			result := validator.getPropertyType(invalidType)
			if result != "" {
				t.Errorf("Expected empty string for invalid type, got %v", result)
			}
		})
	}
}