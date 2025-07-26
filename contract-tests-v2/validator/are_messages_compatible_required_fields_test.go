package validator

import (
	"testing"

	"github.com/codemonsters/contract-tests-v2/parser"
	"github.com/stretchr/testify/assert"
)

// Helper функция для создания сообщения с required полями
func createMessageWithRequiredFieldsForTest(fields []string) *MessageInfo {
	payload := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":     map[string]interface{}{"type": "string"},
			"name":   map[string]interface{}{"type": "string"},
			"age":    map[string]interface{}{"type": "integer"},
			"email":  map[string]interface{}{"type": "string"},
			"userId": map[string]interface{}{"type": "string"},
		},
		"required": fields,
	}
	return &MessageInfo{
		Name:        "TestMessage",
		ContentType: "application/json",
		Payload:     payload,
	}
}

// TestAreMessagesCompatible_DifferentRequiredFields тестирует разное количество required полей
func TestAreMessagesCompatible_DifferentRequiredFields(t *testing.T) {
	v := NewChannelValidator()
	spec := &parser.AsyncAPISpec{}

	tests := []struct {
		name     string
		msg1     *MessageInfo
		msg2     *MessageInfo
		expected bool
	}{
		{
			name:     "same required fields should be compatible",
			msg1:     createMessageWithRequiredFieldsForTest([]string{"id", "name"}),
			msg2:     createMessageWithRequiredFieldsForTest([]string{"id", "name"}),
			expected: true,
		},
		{
			name:     "different order of required fields should be compatible",
			msg1:     createMessageWithRequiredFieldsForTest([]string{"id", "name"}),
			msg2:     createMessageWithRequiredFieldsForTest([]string{"name", "id"}),
			expected: true,
		},
		{
			name:     "different number of required fields should be incompatible",
			msg1:     createMessageWithRequiredFieldsForTest([]string{"id", "name", "age"}),
			msg2:     createMessageWithRequiredFieldsForTest([]string{"id", "name"}),
			expected: false,
		},
		{
			name:     "completely different required fields should be incompatible",
			msg1:     createMessageWithRequiredFieldsForTest([]string{"id", "name"}),
			msg2:     createMessageWithRequiredFieldsForTest([]string{"userId", "email"}),
			expected: false,
		},
		{
			name:     "empty required arrays should be compatible",
			msg1:     createMessageWithRequiredFieldsForTest([]string{}),
			msg2:     createMessageWithRequiredFieldsForTest([]string{}),
			expected: true,
		},
		{
			name:     "one empty required array should be incompatible",
			msg1:     createMessageWithRequiredFieldsForTest([]string{"id"}),
			msg2:     createMessageWithRequiredFieldsForTest([]string{}),
			expected: false,
		},
		{
			name:     "subset of required fields should be incompatible",
			msg1:     createMessageWithRequiredFieldsForTest([]string{"id"}),
			msg2:     createMessageWithRequiredFieldsForTest([]string{"id", "name"}),
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