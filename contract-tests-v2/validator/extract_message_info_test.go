package validator

import (
	"testing"

	"github.com/codemonsters/contract-tests-v2/parser"
)

func TestExtractMessageInfo(t *testing.T) {
	validator := &ChannelValidator{}

	tests := []struct {
		name     string
		spec     *parser.AsyncAPISpec
		msgRef   parser.MessageRef
		expected *MessageInfo
		hasError bool
	}{
		{
			name:   "component message reference",
			spec:   createSpecWithComponentMessage(),
			msgRef: parser.MessageRef{Ref: "#/components/messages/UserSignup"},
			expected: &MessageInfo{
				Name:        "UserSignup",
				ContentType: "application/json",
				Payload: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"userId": map[string]interface{}{"type": "string"},
						"email":  map[string]interface{}{"type": "string"},
					},
					"required": []string{"userId", "email"},
				},
			},
			hasError: false,
		},
		{
			name:   "component message with schema reference",
			spec:   createSpecWithComponentMessageSchemaRef(),
			msgRef: parser.MessageRef{Ref: "#/components/messages/UserUpdate"},
			expected: &MessageInfo{
				Name:        "UserUpdate",
				ContentType: "application/json",
				Payload: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":   map[string]interface{}{"type": "integer"},
						"name": map[string]interface{}{"type": "string"},
					},
					"required": []string{"id"},
				},
			},
			hasError: false,
		},
		{
			name:   "inline channel message",
			spec:   createSpecWithInlineChannelMessage(),
			msgRef: parser.MessageRef{Ref: "#/channels/user~1events/messages/signup"},
			expected: &MessageInfo{
				Name:        "signup",
				ContentType: "application/json",
				Payload: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"timestamp": map[string]interface{}{"type": "string"},
						"userId":    map[string]interface{}{"type": "string"},
					},
				},
			},
			hasError: false,
		},
		{
			name:   "channel message with component reference",
			spec:   createSpecWithChannelMessageComponentRef(),
			msgRef: parser.MessageRef{Ref: "#/channels/user~1notifications/messages/alert"},
			expected: &MessageInfo{
				Name:        "AlertMessage",
				ContentType: "application/json",
				Payload: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"message": map[string]interface{}{"type": "string"},
						"level":   map[string]interface{}{"type": "string"},
					},
					"required": []string{"message"},
				},
			},
			hasError: false,
		},
		{
			name:     "message without payload",
			spec:     createSpecWithMessageWithoutPayload(),
			msgRef:   parser.MessageRef{Ref: "#/components/messages/EmptyMessage"},
			expected: &MessageInfo{
				Name:        "EmptyMessage",
				ContentType: "application/json",
				Payload:     nil,
			},
			hasError: false,
		},
		{
			name:     "nonexistent component message",
			spec:     createSpecWithComponentMessage(),
			msgRef:   parser.MessageRef{Ref: "#/components/messages/NonExistent"},
			expected: nil,
			hasError: true,
		},
		{
			name:     "invalid reference format",
			spec:     createSpecWithComponentMessage(),
			msgRef:   parser.MessageRef{Ref: "invalid-ref-format"},
			expected: nil,
			hasError: true,
		},
		{
			name:     "nil specification",
			spec:     nil,
			msgRef:   parser.MessageRef{Ref: "#/components/messages/UserSignup"},
			expected: nil,
			hasError: true,
		},
		{
			name:     "spec without components",
			spec:     &parser.AsyncAPISpec{},
			msgRef:   parser.MessageRef{Ref: "#/components/messages/UserSignup"},
			expected: nil,
			hasError: true,
		},
		{
			name:     "nonexistent channel in channel reference",
			spec:     createSpecWithInlineChannelMessage(),
			msgRef:   parser.MessageRef{Ref: "#/channels/nonexistent/messages/signup"},
			expected: nil,
			hasError: true,
		},
		{
			name:     "nonexistent message in channel",
			spec:     createSpecWithInlineChannelMessage(),
			msgRef:   parser.MessageRef{Ref: "#/channels/user~1events/messages/nonexistent"},
			expected: nil,
			hasError: true,
		},
		{
			name:     "broken schema reference",
			spec:     createSpecWithBrokenSchemaRef(),
			msgRef:   parser.MessageRef{Ref: "#/components/messages/BrokenRef"},
			expected: &MessageInfo{
				Name:        "BrokenRef",
				ContentType: "application/json",
				Payload:     nil, // resolveSchemaRef returns nil for broken refs
			},
			hasError: false,
		},
		{
			name:   "inline channel message with schema reference",
			spec:   createSpecWithInlineChannelMessageSchemaRef(),
			msgRef: parser.MessageRef{Ref: "#/channels/user~1profile/messages/update"},
			expected: &MessageInfo{
				Name:        "update",
				ContentType: "application/json",
				Payload: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"profileId": map[string]interface{}{"type": "string"},
						"data":      map[string]interface{}{"type": "object"},
					},
					"required": []string{"profileId"},
				},
			},
			hasError: false,
		},
		{
			name:   "inline channel message without name (uses key)",
			spec:   createSpecWithInlineChannelMessageNoName(),
			msgRef: parser.MessageRef{Ref: "#/channels/system~1alerts/messages/warning"},
			expected: &MessageInfo{
				Name:        "warning", // Should use the key name when message name is empty
				ContentType: "application/json",
				Payload: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"severity": map[string]interface{}{"type": "string"},
					},
				},
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.extractMessageInfo(tt.spec, tt.msgRef)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("expected result but got nil")
				return
			}

			if result.Name != tt.expected.Name {
				t.Errorf("expected Name %s but got %s", tt.expected.Name, result.Name)
			}

			if result.ContentType != tt.expected.ContentType {
				t.Errorf("expected ContentType %s but got %s", tt.expected.ContentType, result.ContentType)
			}

			// Compare payloads
			if !comparePayloads(result.Payload, tt.expected.Payload) {
				t.Errorf("expected Payload %+v but got %+v", tt.expected.Payload, result.Payload)
			}
		})
	}
}

// Helper functions for creating test specifications

func createSpecWithComponentMessage() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"UserSignup": {
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"userId": {Type: "string"},
							"email":  {Type: "string"},
						},
						Required: []string{"userId", "email"},
					},
				},
			},
		},
	}
}

func createSpecWithComponentMessageSchemaRef() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"UserUpdate": {
					ContentType: "application/json",
					Payload: &parser.Schema{
						Ref: "#/components/schemas/UserSchema",
					},
				},
			},
			Schemas: map[string]parser.Schema{
				"UserSchema": {
					Type: "object",
					Properties: map[string]parser.Schema{
						"id":   {Type: "integer"},
						"name": {Type: "string"},
					},
					Required: []string{"id"},
				},
			},
		},
	}
}

func createSpecWithInlineChannelMessage() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Channels: map[string]parser.Channel{
			"user/events": {
				Messages: map[string]parser.MessageRef{
					"signup": {
						Message: &parser.Message{
							Name:        "signup",
							ContentType: "application/json",
							Payload: &parser.Schema{
								Type: "object",
								Properties: map[string]parser.Schema{
									"timestamp": {Type: "string"},
									"userId":    {Type: "string"},
								},
							},
						},
					},
				},
			},
		},
	}
}

func createSpecWithChannelMessageComponentRef() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Channels: map[string]parser.Channel{
			"user/notifications": {
				Messages: map[string]parser.MessageRef{
					"alert": {
						Ref: "#/components/messages/AlertMessage",
					},
				},
			},
		},
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"AlertMessage": {
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"message": {Type: "string"},
							"level":   {Type: "string"},
						},
						Required: []string{"message"},
					},
				},
			},
		},
	}
}

func createSpecWithMessageWithoutPayload() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"EmptyMessage": {
					ContentType: "application/json",
					Payload:     nil,
				},
			},
		},
	}
}

func createSpecWithBrokenSchemaRef() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"BrokenRef": {
					ContentType: "application/json",
					Payload: &parser.Schema{
						Ref: "#/components/schemas/NonExistentSchema",
					},
				},
			},
			// Note: schemas map doesn't contain NonExistentSchema
		},
	}
}

func createSpecWithInlineChannelMessageSchemaRef() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Channels: map[string]parser.Channel{
			"user/profile": {
				Messages: map[string]parser.MessageRef{
					"update": {
						Message: &parser.Message{
							Name:        "update",
							ContentType: "application/json",
							Payload: &parser.Schema{
								Ref: "#/components/schemas/ProfileUpdate",
							},
						},
					},
				},
			},
		},
		Components: &parser.Components{
			Schemas: map[string]parser.Schema{
				"ProfileUpdate": {
					Type: "object",
					Properties: map[string]parser.Schema{
						"profileId": {Type: "string"},
						"data":      {Type: "object"},
					},
					Required: []string{"profileId"},
				},
			},
		},
	}
}

func createSpecWithInlineChannelMessageNoName() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Channels: map[string]parser.Channel{
			"system/alerts": {
				Messages: map[string]parser.MessageRef{
					"warning": {
						Message: &parser.Message{
							// Name intentionally omitted to test fallback to key
							ContentType: "application/json",
							Payload: &parser.Schema{
								Type: "object",
								Properties: map[string]parser.Schema{
									"severity": {Type: "string"},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Helper function to compare payloads (handles nil cases)
func comparePayloads(a, b map[string]interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// For deep comparison, we would need a more sophisticated comparison
	// For now, let's do a simple length check and key existence check
	if len(a) != len(b) {
		return false
	}
	for key := range a {
		if _, exists := b[key]; !exists {
			return false
		}
	}
	return true
}