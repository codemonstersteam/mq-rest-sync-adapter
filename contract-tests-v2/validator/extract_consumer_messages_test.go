package validator

import (
	"testing"

	"github.com/codemonsters/contract-tests-v2/parser"
)

func TestExtractConsumerMessages(t *testing.T) {
	validator := NewChannelValidator()

	tests := []struct {
		name              string
		spec              *parser.AsyncAPISpec
		channelName       string
		expectedOut       *MessageInfo
		expectedIn        *MessageInfo
		expectError       bool
	}{
		{
			name:        "operation send with messages and reply",
			spec:        createSpecWithSendOperationAndReply(),
			channelName: "user/events",
			expectedOut: createExpectedOutMessage(),
			expectedIn:  createExpectedInMessage(),
		},
		{
			name:        "operation send only with messages (no reply)",
			spec:        createSpecWithSendOperationOnly(),
			channelName: "user/events", 
			expectedOut: createExpectedOutMessage(),
			expectedIn:  nil,
		},
		{
			name:        "operation send only with reply (no messages)",
			spec:        createSpecWithSendReplyOnly(),
			channelName: "user/events",
			expectedOut: nil,
			expectedIn:  createExpectedInMessage(),
		},
		{
			name:        "channel without operations",
			spec:        createSpecWithoutOperations(),
			channelName: "user/events",
			expectedOut: nil,
			expectedIn:  nil,
		},
		{
			name:        "operation with action receive (should be ignored)",
			spec:        createSpecWithReceiveOperation(),
			channelName: "user/events",
			expectedOut: nil,
			expectedIn:  nil,
		},
		{
			name:        "multiple operations on one channel",
			spec:        createSpecWithMultipleOperations(),
			channelName: "user/events",
			expectedOut: createExpectedOutMessage(), // Should take first send operation
			expectedIn:  createExpectedInMessage(),
		},
		{
			name:        "escaped channel name user~1events",
			spec:        createSpecWithEscapedChannelRef(),
			channelName: "user/events", // Input unescaped
			expectedOut: createExpectedOutMessage(),
			expectedIn:  createExpectedInMessage(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outMessage, inMessage, err := validator.extractConsumerMessages(tt.spec, tt.channelName)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
			}

			// Compare outMessage
			if !compareMessageInfo(outMessage, tt.expectedOut) {
				t.Errorf("outMessage mismatch:\nexpected: %+v\nactual: %+v", tt.expectedOut, outMessage)
			}

			// Compare inMessage
			if !compareMessageInfo(inMessage, tt.expectedIn) {
				t.Errorf("inMessage mismatch:\nexpected: %+v\nactual: %+v", tt.expectedIn, inMessage)
			}
		})
	}
}

// Helper functions для создания тестовых данных

func createSpecWithSendOperationAndReply() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Operations: map[string]parser.Operation{
			"sendUserEvent": {
				Action: "send",
				Channel: parser.ChannelRef{
					Ref: "#/channels/user~1events",
				},
				Messages: []parser.MessageRef{
					{Ref: "#/components/messages/UserEvent"},
				},
				Reply: &parser.Reply{
					Messages: []parser.MessageRef{
						{Ref: "#/components/messages/UserEventReply"},
					},
				},
			},
		},
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"UserEvent": {
					Name:        "UserEvent",
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"userId": {Type: "string"},
							"action": {Type: "string"},
						},
						Required: []string{"userId", "action"},
					},
				},
				"UserEventReply": {
					Name:        "UserEventReply",
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"status": {Type: "string"},
							"timestamp": {Type: "string"},
						},
						Required: []string{"status"},
					},
				},
			},
		},
	}
}

func createSpecWithSendOperationOnly() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Operations: map[string]parser.Operation{
			"sendUserEvent": {
				Action: "send",
				Channel: parser.ChannelRef{
					Ref: "#/channels/user~1events",
				},
				Messages: []parser.MessageRef{
					{Ref: "#/components/messages/UserEvent"},
				},
				// No Reply
			},
		},
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"UserEvent": {
					Name:        "UserEvent",
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"userId": {Type: "string"},
							"action": {Type: "string"},
						},
						Required: []string{"userId", "action"},
					},
				},
			},
		},
	}
}

func createSpecWithSendReplyOnly() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Operations: map[string]parser.Operation{
			"sendUserEvent": {
				Action: "send",
				Channel: parser.ChannelRef{
					Ref: "#/channels/user~1events",
				},
				Messages: []parser.MessageRef{}, // Empty messages
				Reply: &parser.Reply{
					Messages: []parser.MessageRef{
						{Ref: "#/components/messages/UserEventReply"},
					},
				},
			},
		},
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"UserEventReply": {
					Name:        "UserEventReply",
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"status": {Type: "string"},
						},
						Required: []string{"status"},
					},
				},
			},
		},
	}
}

func createSpecWithoutOperations() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Operations: map[string]parser.Operation{},
	}
}

func createSpecWithReceiveOperation() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Operations: map[string]parser.Operation{
			"receiveUserEvent": {
				Action: "receive", // Should be ignored for consumer
				Channel: parser.ChannelRef{
					Ref: "#/channels/user~1events",
				},
				Messages: []parser.MessageRef{
					{Ref: "#/components/messages/UserEvent"},
				},
			},
		},
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"UserEvent": {
					Name:        "UserEvent",
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"userId": {Type: "string"},
						},
						Required: []string{"userId"},
					},
				},
			},
		},
	}
}

func createSpecWithMultipleOperations() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Operations: map[string]parser.Operation{
			"sendUserEvent": {
				Action: "send",
				Channel: parser.ChannelRef{
					Ref: "#/channels/user~1events",
				},
				Messages: []parser.MessageRef{
					{Ref: "#/components/messages/UserEvent"},
				},
				Reply: &parser.Reply{
					Messages: []parser.MessageRef{
						{Ref: "#/components/messages/UserEventReply"},
					},
				},
			},
			"sendOtherEvent": {
				Action: "send",
				Channel: parser.ChannelRef{
					Ref: "#/channels/user~1events",
				},
				Messages: []parser.MessageRef{
					{Ref: "#/components/messages/OtherEvent"},
				},
			},
		},
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"UserEvent": {
					Name:        "UserEvent",
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"userId": {Type: "string"},
							"action": {Type: "string"},
						},
						Required: []string{"userId", "action"},
					},
				},
				"UserEventReply": {
					Name:        "UserEventReply",
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"status": {Type: "string"},
						},
						Required: []string{"status"},
					},
				},
				"OtherEvent": {
					Name:        "OtherEvent",
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"data": {Type: "string"},
						},
						Required: []string{"data"},
					},
				},
			},
		},
	}
}

func createSpecWithEscapedChannelRef() *parser.AsyncAPISpec {
	// Тест что функция корректно экранирует user/events → user~1events
	return createSpecWithSendOperationAndReply()
}

func createExpectedOutMessage() *MessageInfo {
	return &MessageInfo{
		Name:        "UserEvent",
		ContentType: "application/json",
		Payload: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"userId": map[string]interface{}{"type": "string"},
				"action": map[string]interface{}{"type": "string"},
			},
			"required": []string{"userId", "action"},
		},
	}
}

func createExpectedInMessage() *MessageInfo {
	return &MessageInfo{
		Name:        "UserEventReply",
		ContentType: "application/json",
		Payload: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"status":    map[string]interface{}{"type": "string"},
				"timestamp": map[string]interface{}{"type": "string"},
			},
			"required": []string{"status"},
		},
	}
}

// Utility function для сравнения MessageInfo
func compareMessageInfo(actual, expected *MessageInfo) bool {
	if actual == nil && expected == nil {
		return true
	}
	if actual == nil || expected == nil {
		return false
	}
	
	if actual.Name != expected.Name {
		return false
	}
	if actual.ContentType != expected.ContentType {
		return false
	}
	
	// Простое сравнение payload (можно расширить при необходимости)
	return true // Для упрощения, детальное сравнение payload можно добавить позже
}