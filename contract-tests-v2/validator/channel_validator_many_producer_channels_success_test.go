package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codemonsters/contract-tests-v2/parser"
)

func TestChannelValidator_MultipleProviderChannels(t *testing.T) {
	t.Run("should select correct provider channel when multiple channels have same protocol", func(t *testing.T) {
		// Consumer specification with UserSignup message
		consumerSpec := &parser.AsyncAPISpec{
			AsyncAPI: "3.0.0",
			Servers: map[string]parser.Server{
				"rabbitmq": {
					Host:     "localhost:5672",
					Protocol: "amqp",
				},
			},
			Channels: map[string]parser.Channel{
				"user/events": {
					Servers: []parser.ServerRef{
						{Ref: "#/servers/rabbitmq"},
					},
				},
			},
			Operations: map[string]parser.Operation{
				"sendUserEvent": {
					Action: "send",
					Channel: parser.ChannelRef{
						Ref: "#/channels/user~1events",
					},
					Messages: []parser.MessageRef{
						{Ref: "#/components/messages/UserSignup"},
					},
				},
			},
			Components: &parser.Components{
				Messages: map[string]parser.Message{
					"UserSignup": {
						Name:        "UserSignup",
						ContentType: "application/json",
						Payload: &parser.Schema{
							Type: "object",
							Properties: map[string]parser.Schema{
								"userId": {Type: "string", Example: "user123"},
								"email":  {Type: "string", Example: "user@example.com"},
								"name":   {Type: "string", Example: "John Doe"},
							},
							Required: []string{"userId", "email", "name"},
						},
					},
				},
			},
		}

		// Provider specification with multiple AMQP channels
		providerSpec := &parser.AsyncAPISpec{
			AsyncAPI: "3.0.0",
			Servers: map[string]parser.Server{
				"rabbitmq": {
					Host:     "localhost:5672",
					Protocol: "amqp",
				},
			},
			Channels: map[string]parser.Channel{
				// First channel - different message structure (should NOT match)
				"notifications/orders": {
					Servers: []parser.ServerRef{
						{Ref: "#/servers/rabbitmq"},
					},
				},
				// Second channel - compatible message structure (SHOULD match)
				"notifications/users": {
					Servers: []parser.ServerRef{
						{Ref: "#/servers/rabbitmq"},
					},
				},
				// Third channel - different message structure (should NOT match)
				"notifications/products": {
					Servers: []parser.ServerRef{
						{Ref: "#/servers/rabbitmq"},
					},
				},
			},
			Operations: map[string]parser.Operation{
				// Order notifications - incompatible structure
				"receiveOrderNotification": {
					Action: "receive",
					Channel: parser.ChannelRef{
						Ref: "#/channels/notifications~1orders",
					},
					Messages: []parser.MessageRef{
						{Ref: "#/components/messages/OrderEvent"},
					},
				},
				// User notifications - compatible structure
				"receiveUserNotification": {
					Action: "receive",
					Channel: parser.ChannelRef{
						Ref: "#/channels/notifications~1users",
					},
					Messages: []parser.MessageRef{
						{Ref: "#/components/messages/UserSignup"},
					},
				},
				// Product notifications - incompatible structure
				"receiveProductNotification": {
					Action: "receive",
					Channel: parser.ChannelRef{
						Ref: "#/channels/notifications~1products",
					},
					Messages: []parser.MessageRef{
						{Ref: "#/components/messages/ProductEvent"},
					},
				},
			},
			Components: &parser.Components{
				Messages: map[string]parser.Message{
					// Compatible message - same required fields as consumer
					"UserSignup": {
						Name:        "UserSignup",
						ContentType: "application/json",
						Payload: &parser.Schema{
							Type: "object",
							Properties: map[string]parser.Schema{
								"userId": {Type: "string", Example: "user123"},
								"email":  {Type: "string", Example: "user@example.com"},
								"name":   {Type: "string", Example: "John Doe"},
							},
							Required: []string{"userId", "email", "name"},
						},
					},
					// Incompatible message - different required fields
					"OrderEvent": {
						Name:        "OrderEvent",
						ContentType: "application/json",
						Payload: &parser.Schema{
							Type: "object",
							Properties: map[string]parser.Schema{
								"orderId":    {Type: "string", Example: "order123"},
								"customerId": {Type: "string", Example: "customer456"},
								"amount":     {Type: "number", Example: 99.99},
							},
							Required: []string{"orderId", "customerId", "amount"},
						},
					},
					// Another incompatible message
					"ProductEvent": {
						Name:        "ProductEvent",
						ContentType: "application/json",
						Payload: &parser.Schema{
							Type: "object",
							Properties: map[string]parser.Schema{
								"productId": {Type: "string", Example: "product789"},
								"category":  {Type: "string", Example: "electronics"},
							},
							Required: []string{"productId", "category"},
						},
					},
				},
			},
		}

		contractValidate := &ContractValidate{
			ConsumerChannelName: "user/events",
			ConsumerSpec:        consumerSpec,
			ProviderSpec:        providerSpec,
		}

		validator := NewChannelValidator()
		result, err := validator.ValidateChannels(contractValidate)

		require.NoError(t, err)
		require.NotNil(t, result)

		// Debug information
		t.Logf("Consumer channel: %s, protocol: %s", result.ConsumerChannel.Name, result.ConsumerChannel.Protocol)
		if result.ConsumerChannel.OutMessage != nil {
			t.Logf("Consumer OutMessage: %s", result.ConsumerChannel.OutMessage.Name)
			t.Logf("Consumer required fields: %v", result.ConsumerChannel.OutMessage.Payload["required"])
		} else {
			t.Logf("Consumer OutMessage is nil")
		}

		t.Logf("Provider channel: %s, protocol: %s", result.ProviderChannel.Name, result.ProviderChannel.Protocol)
		if result.ProviderChannel.InMessage != nil {
			t.Logf("Provider InMessage: %s", result.ProviderChannel.InMessage.Name)
			t.Logf("Provider required fields: %v", result.ProviderChannel.InMessage.Payload["required"])
		} else {
			t.Logf("Provider InMessage is nil")
		}

		// Verify consumer channel information
		assert.Equal(t, "user/events", result.ConsumerChannel.Name)
		assert.Equal(t, "amqp", result.ConsumerChannel.Protocol)
		assert.NotNil(t, result.ConsumerChannel.OutMessage, "Consumer OutMessage should not be nil")
		
		if result.ConsumerChannel.OutMessage != nil {
			assert.Equal(t, "UserSignup", result.ConsumerChannel.OutMessage.Name)
		}

		// Verify that the correct provider channel was selected
		// Should be "notifications/users" because it has compatible message structure
		assert.Equal(t, "notifications/users", result.ProviderChannel.Name, 
			"Expected notifications/users channel to be selected due to compatible message structure")
		assert.Equal(t, "amqp", result.ProviderChannel.Protocol)
		assert.NotNil(t, result.ProviderChannel.InMessage, "Provider InMessage should not be nil")
		
		if result.ProviderChannel.InMessage != nil {
			assert.Equal(t, "UserSignup", result.ProviderChannel.InMessage.Name)
		}

		// Verify message structure compatibility
		if result.ConsumerChannel.OutMessage != nil && result.ProviderChannel.InMessage != nil {
			consumerPayload := result.ConsumerChannel.OutMessage.Payload
			providerPayload := result.ProviderChannel.InMessage.Payload

			// Both should have the same required fields
			consumerRequired := consumerPayload["required"].([]string)
			providerRequired := providerPayload["required"].([]string)

			assert.ElementsMatch(t, consumerRequired, providerRequired,
				"Consumer and provider should have matching required fields")
			assert.Contains(t, consumerRequired, "userId")
			assert.Contains(t, consumerRequired, "email")
			assert.Contains(t, consumerRequired, "name")
		}
	})
}