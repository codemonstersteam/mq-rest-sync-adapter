package validator

import (
	"strings"
	"testing"
)

func TestGetRequiredFields(t *testing.T) {
	validator := &ChannelValidator{}

	t.Run("Базовые валидные кейсы", func(t *testing.T) {
		t.Run("[]interface{} тип из YAML парсера", func(t *testing.T) {
			payload := createPayloadWithInterfaceRequired([]interface{}{"name", "email", "userId"})
			result, err := validator.getRequiredFields(payload)
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			expected := []string{"name", "email", "userId"}
			assertStringSlicesEqual(t, expected, result)
		})

		t.Run("[]string тип из convertSchema", func(t *testing.T) {
			payload := createPayloadWithStringRequired([]string{"name", "email", "userId"})
			result, err := validator.getRequiredFields(payload)
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			expected := []string{"name", "email", "userId"}
			assertStringSlicesEqual(t, expected, result)
		})

		t.Run("Пустой массив required", func(t *testing.T) {
			payload := createPayloadWithStringRequired([]string{})
			result, err := validator.getRequiredFields(payload)
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result == nil {
				t.Error("Expected empty slice, got nil")
			}
			if len(result) != 0 {
				t.Errorf("Expected empty slice, got %v", result)
			}
		})

		t.Run("Nil payload - должна вернуть ошибку", func(t *testing.T) {
			var payload map[string]interface{}
			result, err := validator.getRequiredFields(payload)
			
			if err == nil {
				t.Error("Expected error for nil payload, got nil")
			}
			if result != nil {
				t.Errorf("Expected nil result for error case, got %v", result)
			}
			if !strings.Contains(err.Error(), "payload is nil") {
				t.Errorf("Expected 'payload is nil' error, got: %v", err)
			}
		})

		t.Run("Отсутствующее поле required в payload", func(t *testing.T) {
			payload := createPayloadWithoutRequired()
			result, err := validator.getRequiredFields(payload)
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result == nil {
				t.Error("Expected empty slice, got nil")
			}
			if len(result) != 0 {
				t.Errorf("Expected empty slice, got %v", result)
			}
		})
	})

	t.Run("AsyncAPI 3.0 compliance кейсы", func(t *testing.T) {
		t.Run("Валидные имена свойств - стандартные", func(t *testing.T) {
			payload := createPayloadWithStringRequired([]string{"name", "email", "userId"})
			result, err := validator.getRequiredFields(payload)
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			expected := []string{"name", "email", "userId"}
			assertStringSlicesEqual(t, expected, result)
		})

		t.Run("Специальные символы - подчеркивания", func(t *testing.T) {
			payload := createPayloadWithStringRequired([]string{"user_id", "created_at", "updated_at"})
			result, err := validator.getRequiredFields(payload)
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			expected := []string{"user_id", "created_at", "updated_at"}
			assertStringSlicesEqual(t, expected, result)
		})

		t.Run("Числовые имена валидные в JSON Schema", func(t *testing.T) {
			payload := createPayloadWithStringRequired([]string{"123", "field1", "2ndAddress"})
			result, err := validator.getRequiredFields(payload)
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			expected := []string{"123", "field1", "2ndAddress"}
			assertStringSlicesEqual(t, expected, result)
		})

		t.Run("Unicode имена - международные", func(t *testing.T) {
			payload := createPayloadWithStringRequired([]string{"имя", "descripción", "адрес"})
			result, err := validator.getRequiredFields(payload)
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			expected := []string{"имя", "descripción", "адрес"}
			assertStringSlicesEqual(t, expected, result)
		})
	})

	t.Run("Невалидные типы данных", func(t *testing.T) {
		t.Run("required как строка - должна вернуть ошибку", func(t *testing.T) {
			payload := createPayloadWithInvalidRequired("name")
			result, err := validator.getRequiredFields(payload)
			
			if err == nil {
				t.Error("Expected error for string required, got nil")
			}
			if result != nil {
				t.Errorf("Expected nil result for error case, got %v", result)
			}
			if !strings.Contains(err.Error(), "invalid required field type") {
				t.Errorf("Expected 'invalid required field type' error, got: %v", err)
			}
		})

		t.Run("required как число", func(t *testing.T) {
			payload := createPayloadWithInvalidRequired(123)
			result, err := validator.getRequiredFields(payload)
			
			if err == nil {
				t.Error("Expected error for numeric required, got nil")
			}
			if result != nil {
				t.Errorf("Expected nil result for error case, got %v", result)
			}
			if !strings.Contains(err.Error(), "invalid required field type") {
				t.Errorf("Expected 'invalid required field type' error, got: %v", err)
			}
		})

		t.Run("required как boolean", func(t *testing.T) {
			payload := createPayloadWithInvalidRequired(true)
			result, err := validator.getRequiredFields(payload)
			
			if err == nil {
				t.Error("Expected error for boolean required, got nil")
			}
			if result != nil {
				t.Errorf("Expected nil result for error case, got %v", result)
			}
			if !strings.Contains(err.Error(), "invalid required field type") {
				t.Errorf("Expected 'invalid required field type' error, got: %v", err)
			}
		})

		t.Run("required как объект", func(t *testing.T) {
			payload := createPayloadWithInvalidRequired(map[string]interface{}{"name": true})
			result, err := validator.getRequiredFields(payload)
			
			if err == nil {
				t.Error("Expected error for object required, got nil")
			}
			if result != nil {
				t.Errorf("Expected nil result for error case, got %v", result)
			}
			if !strings.Contains(err.Error(), "invalid required field type") {
				t.Errorf("Expected 'invalid required field type' error, got: %v", err)
			}
		})

		t.Run("required как null", func(t *testing.T) {
			payload := createPayloadWithInvalidRequired(nil)
			result, err := validator.getRequiredFields(payload)
			
			if err == nil {
				t.Error("Expected error for null required, got nil")
			}
			if result != nil {
				t.Errorf("Expected nil result for error case, got %v", result)
			}
			if !strings.Contains(err.Error(), "invalid required field type") {
				t.Errorf("Expected 'invalid required field type' error, got: %v", err)
			}
		})
	})

	t.Run("Частично невалидные данные", func(t *testing.T) {
		t.Run("Массив с null элементами", func(t *testing.T) {
			payload := createPayloadWithInterfaceRequired([]interface{}{"name", nil, "email"})
			result, err := validator.getRequiredFields(payload)
			
			if err == nil {
				t.Error("Expected error for array with null elements, got nil")
			}
			if result != nil {
				t.Errorf("Expected nil result for error case, got %v", result)
			}
			if !strings.Contains(err.Error(), "invalid required field at index 1") {
				t.Errorf("Expected 'invalid required field at index 1' error, got: %v", err)
			}
		})

		t.Run("Массив с числовыми элементами", func(t *testing.T) {
			payload := createPayloadWithInterfaceRequired([]interface{}{"name", 123, "email"})
			result, err := validator.getRequiredFields(payload)
			
			if err == nil {
				t.Error("Expected error for array with numeric elements, got nil")
			}
			if result != nil {
				t.Errorf("Expected nil result for error case, got %v", result)
			}
			if !strings.Contains(err.Error(), "invalid required field at index 1") {
				t.Errorf("Expected 'invalid required field at index 1' error, got: %v", err)
			}
		})

		t.Run("Массив с объектами", func(t *testing.T) {
			payload := createPayloadWithInterfaceRequired([]interface{}{"name", map[string]interface{}{"invalid": true}, "email"})
			result, err := validator.getRequiredFields(payload)
			
			if err == nil {
				t.Error("Expected error for array with object elements, got nil")
			}
			if result != nil {
				t.Errorf("Expected nil result for error case, got %v", result)
			}
			if !strings.Contains(err.Error(), "invalid required field at index 1") {
				t.Errorf("Expected 'invalid required field at index 1' error, got: %v", err)
			}
		})
	})

	t.Run("Production edge cases", func(t *testing.T) {
		t.Run("Пустые строки в required", func(t *testing.T) {
			payload := createPayloadWithStringRequired([]string{"name", "", "email"})
			result, err := validator.getRequiredFields(payload)
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			expected := []string{"name", "", "email"}
			assertStringSlicesEqual(t, expected, result)
		})

		t.Run("Дублированные поля", func(t *testing.T) {
			payload := createPayloadWithStringRequired([]string{"name", "email", "name"})
			result, err := validator.getRequiredFields(payload)
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			expected := []string{"name", "email", "name"}
			assertStringSlicesEqual(t, expected, result)
		})

		t.Run("Очень длинные имена полей", func(t *testing.T) {
			longFieldName := "very_long_field_name_that_exceeds_normal_limits_and_tests_edge_cases_for_field_name_length"
			payload := createPayloadWithStringRequired([]string{longFieldName})
			result, err := validator.getRequiredFields(payload)
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			expected := []string{longFieldName}
			assertStringSlicesEqual(t, expected, result)
		})
	})
}

// Helper functions для создания тестовых данных

func createPayloadWithInterfaceRequired(required []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type":     "object",
		"required": required,
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
		},
	}
}

func createPayloadWithStringRequired(required []string) map[string]interface{} {
	return map[string]interface{}{
		"type":     "object",
		"required": required,
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
		},
	}
}

func createPayloadWithoutRequired() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
		},
	}
}

func createPayloadWithInvalidRequired(required interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type":     "object",
		"required": required,
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
		},
	}
}

func assertStringSlicesEqual(t *testing.T, expected, actual []string) {
	if len(expected) != len(actual) {
		t.Errorf("Expected length %d, got %d", len(expected), len(actual))
		return
	}
	
	for i, expectedVal := range expected {
		if actual[i] != expectedVal {
			t.Errorf("At index %d: expected '%s', got '%s'", i, expectedVal, actual[i])
		}
	}
}