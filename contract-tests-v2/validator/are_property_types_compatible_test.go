package validator

import (
	"testing"

	"github.com/codemonsters/contract-tests-v2/parser"
)

func TestChannelValidator_ArePropertyTypesCompatible(t *testing.T) {
	validator := &ChannelValidator{}

	tests := []struct {
		name     string
		prop1    interface{}
		prop2    interface{}
		spec1    *parser.AsyncAPISpec
		spec2    *parser.AsyncAPISpec
		expected bool
	}{
		// Простые типы (string vs string)
		{
			name:     "identical string types",
			prop1:    createSimpleProperty("string"),
			prop2:    createSimpleProperty("string"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true,
		},
		{
			name:     "identical number types",
			prop1:    createSimpleProperty("number"),
			prop2:    createSimpleProperty("number"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true,
		},
		{
			name:     "identical integer types",
			prop1:    createSimpleProperty("integer"),
			prop2:    createSimpleProperty("integer"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true,
		},
		{
			name:     "identical boolean types",
			prop1:    createSimpleProperty("boolean"),
			prop2:    createSimpleProperty("boolean"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true,
		},

		// Разные типы (string vs number)
		{
			name:     "different types string vs number",
			prop1:    createSimpleProperty("string"),
			prop2:    createSimpleProperty("number"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: false,
		},
		{
			name:     "different types integer vs boolean",
			prop1:    createSimpleProperty("integer"),
			prop2:    createSimpleProperty("boolean"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: false,
		},

		// Оба свойства с $ref на одинаковые схемы
		{
			name:     "identical refs",
			prop1:    createRefProperty("#/components/schemas/User"),
			prop2:    createRefProperty("#/components/schemas/User"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true,
		},

		// Оба свойства с $ref на разные схемы
		{
			name:     "different refs",
			prop1:    createRefProperty("#/components/schemas/User"),
			prop2:    createRefProperty("#/components/schemas/Product"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: false,
		},

		// Одно с $ref, другое inline
		{
			name:     "ref vs inline type",
			prop1:    createRefProperty("#/components/schemas/User"),
			prop2:    createSimpleProperty("string"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: false,
		},
		{
			name:     "inline type vs ref",
			prop1:    createSimpleProperty("number"),
			prop2:    createRefProperty("#/components/schemas/Price"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: false,
		},

		// Массивы - функция вызывает areArrayItemsCompatible (рекурсивная валидация)
		{
			name:     "compatible array types",
			prop1:    createArrayPropertyForTypes("string"),
			prop2:    createArrayPropertyForTypes("string"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true,
		},
		{
			name:     "incompatible array element types",
			prop1:    createArrayPropertyForTypes("string"),
			prop2:    createArrayPropertyForTypes("number"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: false,
		},

		// Объекты - функция вызывает areObjectPropertiesCompatible (рекурсивная валидация)
		{
			name:     "compatible object types",
			prop1:    createObjectPropertyForTypes(map[string]interface{}{"name": createSimpleProperty("string")}),
			prop2:    createObjectPropertyForTypes(map[string]interface{}{"name": createSimpleProperty("string")}),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true,
		},
		{
			name:     "incompatible object property types",
			prop1:    createObjectPropertyWithRequired(map[string]interface{}{"id": createSimpleProperty("string")}, []string{"id"}),
			prop2:    createObjectPropertyWithRequired(map[string]interface{}{"id": createSimpleProperty("number")}, []string{"id"}),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: false,
		},

		// Edge cases с nil значениями (spec1 или spec2 равен nil)
		{
			name:     "nil spec1",
			prop1:    createSimpleProperty("string"),
			prop2:    createSimpleProperty("string"),
			spec1:    nil,
			spec2:    createEmptySpec(),
			expected: true, // Простые типы не требуют specs для сравнения
		},
		{
			name:     "nil spec2",
			prop1:    createSimpleProperty("number"),
			prop2:    createSimpleProperty("number"),
			spec1:    createEmptySpec(),
			spec2:    nil,
			expected: true,
		},
		{
			name:     "both specs nil",
			prop1:    createSimpleProperty("boolean"),
			prop2:    createSimpleProperty("boolean"),
			spec1:    nil,
			spec2:    nil,
			expected: true,
		},

		// Пустые свойства (map[string]interface{}{})
		{
			name:     "empty properties",
			prop1:    map[string]interface{}{},
			prop2:    map[string]interface{}{},
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: false, // Отсутствие типа должно возвращать false
		},
		{
			name:     "empty prop1 vs valid prop2",
			prop1:    map[string]interface{}{},
			prop2:    createSimpleProperty("string"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: false,
		},
		{
			name:     "valid prop1 vs empty prop2",
			prop1:    createSimpleProperty("integer"),
			prop2:    map[string]interface{}{},
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: false,
		},

		// Nil свойства
		{
			name:     "nil prop1",
			prop1:    nil,
			prop2:    createSimpleProperty("string"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: false,
		},
		{
			name:     "nil prop2",
			prop1:    createSimpleProperty("string"),
			prop2:    nil,
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: false,
		},
		{
			name:     "both props nil",
			prop1:    nil,
			prop2:    nil,
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: false,
		},

		// Format совместимость для AsyncAPI 3.0 (int32, int64, float, double)
		{
			name:     "same format int32",
			prop1:    createPropertyWithFormat("integer", "int32"),
			prop2:    createPropertyWithFormat("integer", "int32"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true,
		},
		{
			name:     "same format int64",
			prop1:    createPropertyWithFormat("integer", "int64"),
			prop2:    createPropertyWithFormat("integer", "int64"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true,
		},
		{
			name:     "different integer formats int32 vs int64",
			prop1:    createPropertyWithFormat("integer", "int32"),
			prop2:    createPropertyWithFormat("integer", "int64"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true, // Same base type, formats should be compatible
		},
		{
			name:     "same format float",
			prop1:    createPropertyWithFormat("number", "float"),
			prop2:    createPropertyWithFormat("number", "float"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true,
		},
		{
			name:     "same format double",
			prop1:    createPropertyWithFormat("number", "double"),
			prop2:    createPropertyWithFormat("number", "double"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true,
		},
		{
			name:     "different number formats float vs double",
			prop1:    createPropertyWithFormat("number", "float"),
			prop2:    createPropertyWithFormat("number", "double"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true, // Same base type, formats should be compatible
		},
		{
			name:     "format vs no format on same base type",
			prop1:    createPropertyWithFormat("integer", "int64"),
			prop2:    createSimpleProperty("integer"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true, // Format is optional, base type matches
		},

		// AsyncAPI 3.0 специфичные ссылки
		{
			name:     "message component reference vs schema component reference",
			prop1:    createRefProperty("#/components/messages/UserEvent"),
			prop2:    createRefProperty("#/components/schemas/UserEvent"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: false, // Разные типы компонентов AsyncAPI
		},
		{
			name:     "valid AsyncAPI schema references",
			prop1:    createRefProperty("#/components/schemas/UserProfile"),
			prop2:    createRefProperty("#/components/schemas/UserProfile"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true,
		},
		{
			name:     "invalid reference format",
			prop1:    createRefProperty("invalid-ref-format"),
			prop2:    createRefProperty("invalid-ref-format"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true, // Функция не валидирует формат ссылок, только сравнивает
		},

		// Контрактное тестирование scenarios (реальные кейсы)
		{
			name:     "user event timestamp field type mismatch",
			prop1:    createSimpleProperty("string"), // consumer expects ISO string
			prop2:    createSimpleProperty("integer"), // provider sends unix timestamp
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: false, // Контракты несовместимы - критическая ошибка!
		},
		{
			name:     "user ID field compatible types",
			prop1:    createSimpleProperty("string"), // consumer expects string ID
			prop2:    createSimpleProperty("string"), // provider sends string ID
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true, // Контракты совместимы
		},
		{
			name:     "amount field with ref to Money schema",
			prop1:    createRefProperty("#/components/schemas/Money"),
			prop2:    createRefProperty("#/components/schemas/Money"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true, // Оба ссылаются на одну схему денег
		},
		{
			name:     "price field different money schemas",
			prop1:    createRefProperty("#/components/schemas/USD"),
			prop2:    createRefProperty("#/components/schemas/EUR"),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: false, // Разные валюты - несовместимо
		},
		{
			name:     "complex nested object validation delegation",
			prop1:    createComplexObjectProperty(),
			prop2:    createComplexObjectProperty(),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true, // Делегирует в areObjectPropertiesCompatible
		},
		{
			name:     "complex array validation delegation",
			prop1:    createComplexArrayProperty(),
			prop2:    createComplexArrayProperty(),
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: true, // Делегирует в areArrayItemsCompatible
		},

		// Edge cases для production scenarios
		{
			name:     "boolean flag vs string enum",
			prop1:    createSimpleProperty("boolean"), // consumer: active: true/false
			prop2:    createSimpleProperty("string"),  // provider: status: "active"/"inactive"
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: false, // Несовместимые представления состояния
		},
		{
			name:     "numeric ID vs string UUID",
			prop1:    createSimpleProperty("integer"), // consumer: id: 12345
			prop2:    createSimpleProperty("string"),  // provider: id: "uuid-string"
			spec1:    createEmptySpec(),
			spec2:    createEmptySpec(),
			expected: false, // Разные форматы идентификаторов
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.arePropertyTypesCompatible(tt.prop1, tt.prop2, tt.spec1, tt.spec2)
			if result != tt.expected {
				t.Errorf("arePropertyTypesCompatible() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Helper functions for this specific test

func createSimpleProperty(propType string) map[string]interface{} {
	return map[string]interface{}{
		"type": propType,
	}
}

func createRefProperty(ref string) map[string]interface{} {
	return map[string]interface{}{
		"$ref": ref,
	}
}

func createArrayPropertyForTypes(itemType string) map[string]interface{} {
	return map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": itemType,
		},
	}
}

func createObjectPropertyForTypes(properties map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
}

func createObjectPropertyWithRequired(properties map[string]interface{}, required []string) map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}

func createPropertyWithFormat(propType, format string) map[string]interface{} {
	return map[string]interface{}{
		"type":   propType,
		"format": format,
	}
}

func createComplexObjectProperty() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"user": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id":   map[string]interface{}{"type": "string"},
					"name": map[string]interface{}{"type": "string"},
				},
				"required": []string{"id", "name"},
			},
			"timestamp": map[string]interface{}{"type": "string"},
		},
		"required": []string{"user", "timestamp"},
	}
}

func createComplexArrayProperty() map[string]interface{} {
	return map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":   map[string]interface{}{"type": "string"},
				"data": map[string]interface{}{"$ref": "#/components/schemas/EventData"},
			},
			"required": []string{"id", "data"},
		},
	}
}

func createEmptySpec() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		AsyncAPI: "3.0.0",
		Info: parser.Info{
			Title:   "Test Spec",
			Version: "1.0.0",
		},
	}
}