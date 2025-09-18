package validator

import (
	"fmt"
	"testing"
)

func TestGetPropertyRef(t *testing.T) {
	validator := &ChannelValidator{}

	// Базовые тест-кейсы
	t.Run("Базовые тест-кейсы", func(t *testing.T) {
		tests := []struct {
			name     string
			prop     interface{}
			expected string
		}{
			{
				name: "валидная $ref ссылка на компонент",
				prop: map[string]interface{}{
					"$ref": "#/components/schemas/UserData",
				},
				expected: "#/components/schemas/UserData",
			},
			{
				name: "свойство без $ref (возвращает пустую строку)",
				prop: map[string]interface{}{
					"type": "string",
					"description": "User name",
				},
				expected: "",
			},
			{
				name:     "nil свойство (безопасная обработка)",
				prop:     nil,
				expected: "",
			},
			{
				name: "неправильный тип $ref (не string)",
				prop: map[string]interface{}{
					"$ref": 123,
				},
				expected: "",
			},
			{
				name: "пустая ссылка (refVal = \"\")",
				prop: map[string]interface{}{
					"$ref": "",
				},
				expected: "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := validator.getPropertyRef(tt.prop)
				if result != tt.expected {
					t.Errorf("getPropertyRef() = %v, expected %v", result, tt.expected)
				}
			})
		}
	})

	// AsyncAPI 3.0 специфичные форматы ссылок
	t.Run("AsyncAPI 3.0 специфичные форматы ссылок", func(t *testing.T) {
		tests := []struct {
			name     string
			prop     interface{}
			expected string
		}{
			{
				name: "ссылка на schema",
				prop: map[string]interface{}{
					"$ref": "#/components/schemas/UserData",
				},
				expected: "#/components/schemas/UserData",
			},
			{
				name: "ссылка на message",
				prop: map[string]interface{}{
					"$ref": "#/components/messages/UserSignup",
				},
				expected: "#/components/messages/UserSignup",
			},
			{
				name: "локальный файл",
				prop: map[string]interface{}{
					"$ref": "./schemas/user.json",
				},
				expected: "./schemas/user.json",
			},
			{
				name: "внешний URL",
				prop: map[string]interface{}{
					"$ref": "https://api.example.com/user-schema.json",
				},
				expected: "https://api.example.com/user-schema.json",
			},
			{
				name: "относительная ссылка",
				prop: map[string]interface{}{
					"$ref": "../common/types.yaml#/User",
				},
				expected: "../common/types.yaml#/User",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := validator.getPropertyRef(tt.prop)
				if result != tt.expected {
					t.Errorf("getPropertyRef() = %v, expected %v", result, tt.expected)
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
				name: "case sensitivity: $Ref вместо $ref (должно игнорироваться)",
				prop: map[string]interface{}{
					"$Ref": "#/components/schemas/User",
				},
				expected: "",
			},
			{
				name: "дополнительные поля: приоритет $ref",
				prop: map[string]interface{}{
					"$ref": "#/components/schemas/User",
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{"type": "string"},
					},
				},
				expected: "#/components/schemas/User",
			},
			{
				name:     "невалидное свойство: не map[string]interface{}",
				prop:     "invalid property type",
				expected: "",
			},
			{
				name: "множественные $ref в одном объекте (невалидный JSON Schema)",
				prop: map[string]interface{}{
					"$ref": "#/components/schemas/User",
					// В реальности второй $ref перезапишет первый, но тестируем текущее поведение
				},
				expected: "#/components/schemas/User",
			},
			{
				name: "специальные символы в ссылках",
				prop: map[string]interface{}{
					"$ref": "#/components/schemas/User-Data_With_Special-Chars",
				},
				expected: "#/components/schemas/User-Data_With_Special-Chars",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := validator.getPropertyRef(tt.prop)
				if result != tt.expected {
					t.Errorf("getPropertyRef() = %v, expected %v", result, tt.expected)
				}
			})
		}
	})

	// Дополнительные edge cases для полного покрытия
	t.Run("Дополнительные edge cases", func(t *testing.T) {
		tests := []struct {
			name     string
			prop     interface{}
			expected string
		}{
			{
				name: "пустой map",
				prop: map[string]interface{}{},
				expected: "",
			},
			{
				name: "map с другими полями но без $ref",
				prop: map[string]interface{}{
					"type": "object",
					"format": "uuid",
					"enum": []string{"value1", "value2"},
				},
				expected: "",
			},
			{
				name: "ссылка с пробелами и юникодом",
				prop: map[string]interface{}{
					"$ref": "#/components/schemas/Пользователь Data",
				},
				expected: "#/components/schemas/Пользователь Data",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := validator.getPropertyRef(tt.prop)
				if result != tt.expected {
					t.Errorf("getPropertyRef() = %v, expected %v", result, tt.expected)
				}
			})
		}
	})
}

// Тест для проверки безопасности type assertion
func TestGetPropertyRef_TypeSafety(t *testing.T) {
	validator := &ChannelValidator{}

	// Тестируем различные невалидные типы, которые могут вызвать панику
	invalidTypes := []interface{}{
		123,
		true,
		[]string{"test"},
		map[int]string{1: "test"},
		struct{ Field string }{Field: "test"},
	}

	for i, invalidType := range invalidTypes {
		t.Run(fmt.Sprintf("invalid_type_%d", i), func(t *testing.T) {
			// Этот тест должен пройти без паники
			result := validator.getPropertyRef(invalidType)
			if result != "" {
				t.Errorf("Expected empty string for invalid type, got %v", result)
			}
		})
	}
}