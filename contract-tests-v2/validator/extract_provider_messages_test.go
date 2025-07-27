package validator

import (
	"testing"

	"github.com/codemonsters/contract-tests-v2/parser"
)

func TestExtractProviderMessages(t *testing.T) {
	validator := NewChannelValidator()

	tests := []struct {
		name              string
		spec              *parser.AsyncAPISpec
		channelName       string
		expectedIn        *MessageInfo
		expectedOut       *MessageInfo
		expectError       bool
	}{
		{
			name:        "operation receive with messages and reply",
			spec:        createProviderSpecWithReceiveOperationAndReply(),
			channelName: "user/events",
			expectedIn:  createExpectedProviderInMessage(),
			expectedOut: createExpectedProviderOutMessage(),
		},
		{
			name:        "operation receive only with messages",
			spec:        createProviderSpecWithReceiveOperationOnly(),
			channelName: "user/events",
			expectedIn:  createExpectedProviderInMessage(),
			expectedOut: nil,
		},
		{
			name:        "operation receive only with reply",
			spec:        createProviderSpecWithReceiveReplyOnly(),
			channelName: "user/events",
			expectedIn:  nil,
			expectedOut: createExpectedProviderOutMessage(),
		},
		{
			name:        "channel without receive operations",
			spec:        createProviderSpecWithoutReceiveOperations(),
			channelName: "user/events",
			expectedIn:  nil,
			expectedOut: nil,
		},
		{
			name:        "operation with action send (should be ignored)",
			spec:        createProviderSpecWithSendOperation(),
			channelName: "user/events",
			expectedIn:  nil,
			expectedOut: nil,
		},
		{
			name:        "multiple receive operations",
			spec:        createProviderSpecWithMultipleReceiveOperations(),
			channelName: "user/events",
			expectedIn:  createExpectedProviderInMessage(), // Should take first receive operation
			expectedOut: createExpectedProviderOutMessage(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inMessage, outMessage, err := validator.extractProviderMessages(tt.spec, tt.channelName)

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

			// Compare inMessage
			if !compareMessageInfo(inMessage, tt.expectedIn) {
				t.Errorf("inMessage mismatch:\nexpected: %+v\nactual: %+v", tt.expectedIn, inMessage)
			}

			// Compare outMessage
			if !compareMessageInfo(outMessage, tt.expectedOut) {
				t.Errorf("outMessage mismatch:\nexpected: %+v\nactual: %+v", tt.expectedOut, outMessage)
			}
		})
	}
}

// Helper functions для создания тестовых данных поставщика

func createProviderSpecWithReceiveOperationAndReply() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Operations: map[string]parser.Operation{
			"receiveUserEvent": {
				Action: "receive",
				Channel: parser.ChannelRef{
					Ref: "#/channels/user~1events",
				},
				Messages: []parser.MessageRef{
					{Ref: "#/components/messages/UserEventRequest"},
				},
				Reply: &parser.Reply{
					Messages: []parser.MessageRef{
						{Ref: "#/components/messages/UserEventResponse"},
					},
				},
			},
		},
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"UserEventRequest": {
					Name:        "UserEventRequest",
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
				"UserEventResponse": {
					Name:        "UserEventResponse",
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"status":    {Type: "string"},
							"timestamp": {Type: "string"},
							"result":    {Type: "string"},
						},
						Required: []string{"status", "timestamp"},
					},
				},
			},
		},
	}
}

func createProviderSpecWithReceiveOperationOnly() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Operations: map[string]parser.Operation{
			"receiveUserEvent": {
				Action: "receive",
				Channel: parser.ChannelRef{
					Ref: "#/channels/user~1events",
				},
				Messages: []parser.MessageRef{
					{Ref: "#/components/messages/UserEventRequest"},
				},
				// No Reply
			},
		},
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"UserEventRequest": {
					Name:        "UserEventRequest",
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

func createProviderSpecWithReceiveReplyOnly() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Operations: map[string]parser.Operation{
			"receiveUserEvent": {
				Action: "receive",
				Channel: parser.ChannelRef{
					Ref: "#/channels/user~1events",
				},
				Messages: []parser.MessageRef{}, // Empty messages
				Reply: &parser.Reply{
					Messages: []parser.MessageRef{
						{Ref: "#/components/messages/UserEventResponse"},
					},
				},
			},
		},
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"UserEventResponse": {
					Name:        "UserEventResponse",
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"status": {Type: "string"},
							"result": {Type: "string"},
						},
						Required: []string{"status"},
					},
				},
			},
		},
	}
}

func createProviderSpecWithoutReceiveOperations() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Operations: map[string]parser.Operation{},
	}
}

func createProviderSpecWithSendOperation() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Operations: map[string]parser.Operation{
			"sendUserEvent": {
				Action: "send", // Should be ignored for provider
				Channel: parser.ChannelRef{
					Ref: "#/channels/user~1events",
				},
				Messages: []parser.MessageRef{
					{Ref: "#/components/messages/UserEventRequest"},
				},
			},
		},
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"UserEventRequest": {
					Name:        "UserEventRequest",
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

func createProviderSpecWithMultipleReceiveOperations() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Operations: map[string]parser.Operation{
			"receiveUserEvent": {
				Action: "receive",
				Channel: parser.ChannelRef{
					Ref: "#/channels/user~1events",
				},
				Messages: []parser.MessageRef{
					{Ref: "#/components/messages/UserEventRequest"},
				},
				Reply: &parser.Reply{
					Messages: []parser.MessageRef{
						{Ref: "#/components/messages/UserEventResponse"},
					},
				},
			},
			"receiveOtherEvent": {
				Action: "receive",
				Channel: parser.ChannelRef{
					Ref: "#/channels/user~1events",
				},
				Messages: []parser.MessageRef{
					{Ref: "#/components/messages/OtherEventRequest"},
				},
			},
		},
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"UserEventRequest": {
					Name:        "UserEventRequest",
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
				"UserEventResponse": {
					Name:        "UserEventResponse",
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"status":    {Type: "string"},
							"timestamp": {Type: "string"},
							"result":    {Type: "string"},
						},
						Required: []string{"status", "timestamp"},
					},
				},
				"OtherEventRequest": {
					Name:        "OtherEventRequest",
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

func createExpectedProviderInMessage() *MessageInfo {
	return &MessageInfo{
		Name:        "UserEventRequest",
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

func createExpectedProviderOutMessage() *MessageInfo {
	return &MessageInfo{
		Name:        "UserEventResponse",
		ContentType: "application/json",
		Payload: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"status":    map[string]interface{}{"type": "string"},
				"timestamp": map[string]interface{}{"type": "string"},
				"result":    map[string]interface{}{"type": "string"},
			},
			"required": []string{"status", "timestamp"},
		},
	}
}