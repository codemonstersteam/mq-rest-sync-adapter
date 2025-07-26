package validator

import (
	"testing"

	"github.com/codemonsters/contract-tests-v2/parser"
	"github.com/stretchr/testify/assert"
)

// Helper функции для создания тестовых сообщений

func createMessageWithProperties(properties map[string]interface{}) *MessageInfo {
	return &MessageInfo{
		Name:        "TestMessage",
		ContentType: "application/json",
		Payload:     properties,
	}
}

func createMessageWithRequiredFields(fields []string) *MessageInfo {
	payload := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":   map[string]interface{}{"type": "string"},
			"name": map[string]interface{}{"type": "string"},
			"age":  map[string]interface{}{"type": "integer"},
		},
		"required": fields,
	}
	return &MessageInfo{
		Name:        "TestMessage",
		ContentType: "application/json",
		Payload:     payload,
	}
}

// TestAreMessagesCompatible_NilMessages тестирует обработку nil сообщений
func TestAreMessagesCompatible_NilMessages(t *testing.T) {
	v := NewChannelValidator()
	spec := &parser.AsyncAPISpec{}

	tests := []struct {
		name     string
		msg1     *MessageInfo
		msg2     *MessageInfo
		expected bool
	}{
		{
			name:     "both messages nil should return false",
			msg1:     nil,
			msg2:     nil,
			expected: false,
		},
		{
			name:     "first message nil should return false",
			msg1:     nil,
			msg2:     createMessageWithProperties(map[string]interface{}{"type": "object"}),
			expected: false,
		},
		{
			name:     "second message nil should return false",
			msg1:     createMessageWithProperties(map[string]interface{}{"type": "object"}),
			msg2:     nil,
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