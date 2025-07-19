package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codemonsters/contract-tests-v2/parser"
)

func TestChannelValidator_NoMatchingMessageStructure(t *testing.T) {
	t.Run("should return error when provider has channels with same protocol but incompatible message structures", func(t *testing.T) {
		// Arrange
		consumerSpec := createConsumerSpecWithUserSignupMessage()
		providerSpec := createProviderSpecWithIncompatibleMessages()
		contractValidate := &ContractValidate{
			ConsumerChannelName: "user/events",
			ProviderSpec:        providerSpec,
			ConsumerSpec:        consumerSpec,
		}

		// Act
		validator := NewChannelValidator()
		result, err := validator.ValidateChannels(contractValidate)

		// Assert
		require.Error(t, err, "Should return error when no matching message structures found")
		assert.Nil(t, result, "Result should be nil on error")
		assert.Contains(t, err.Error(), "no compatible provider channel found", "Error should mention no compatible channel")

		// Логирование для документации
		t.Logf("❌ Test case: Provider has 2 AMQP channels but none match consumer message structure")
		t.Logf("   Consumer channel: user/events (AMQP) with UserSignup message")
		t.Logf("   Provider channel 1: orders/created (AMQP) with OrderEvent - INCOMPATIBLE")
		t.Logf("   Provider channel 2: products/updated (AMQP) with ProductEvent - INCOMPATIBLE")
		t.Logf("   Expected result: Error 'no compatible provider channel found'")
		t.Logf("   Actual error: %v", err)
	})
}

// createConsumerSpecWithUserSignupMessage создает спецификацию потребителя
// с каналом user/events и сообщением UserSignup (userId, email, name)
func createConsumerSpecWithUserSignupMessage() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		AsyncAPI: "3.0.0",
		Info: parser.Info{
			Title:   "Consumer Service",
			Version: "1.0.0",
		},
		Servers: map[string]parser.Server{
			"rabbitmq": {
				Host:     "localhost:5672",
				Protocol: "amqp",
			},
		},
		Channels: map[string]parser.Channel{
			"user/events": {
				Address: "user.events",
				Servers: []parser.ServerRef{
					{Ref: "#/servers/rabbitmq"},
				},
				Messages: map[string]parser.MessageRef{
					"userSignedUp": {Ref: "#/components/messages/UserSignup"},
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
					{Ref: "#/channels/user~1events/messages/userSignedUp"},
				},
			},
		},
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"UserSignup": {
					Name:        "UserSignup",
					Title:       "User signed up event",
					ContentType: "application/json",
					Payload:     createUserSignupSchema(),
				},
			},
		},
	}
}

// createProviderSpecWithIncompatibleMessages создает спецификацию поставщика
// с 2 AMQP каналами, которые имеют несовместимые структуры сообщений:
// - orders/created с OrderEvent (orderId, customerId, amount)
// - products/updated с ProductEvent (productId, productName, price)
func createProviderSpecWithIncompatibleMessages() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		AsyncAPI: "3.0.0",
		Info: parser.Info{
			Title:   "Provider Service",
			Version: "1.0.0",
		},
		Servers: map[string]parser.Server{
			"rabbitmq": {
				Host:     "localhost:5672",
				Protocol: "amqp",
			},
		},
		Channels: map[string]parser.Channel{
			"orders/created": {
				Address: "orders.created",
				Servers: []parser.ServerRef{
					{Ref: "#/servers/rabbitmq"},
				},
				Messages: map[string]parser.MessageRef{
					"orderCreated": {Ref: "#/components/messages/OrderEvent"},
				},
			},
			"products/updated": {
				Address: "products.updated",
				Servers: []parser.ServerRef{
					{Ref: "#/servers/rabbitmq"},
				},
				Messages: map[string]parser.MessageRef{
					"productUpdated": {Ref: "#/components/messages/ProductEvent"},
				},
			},
		},
		Operations: map[string]parser.Operation{
			"receiveOrderEvent": {
				Action: "receive",
				Channel: parser.ChannelRef{
					Ref: "#/channels/orders~1created",
				},
				Messages: []parser.MessageRef{
					{Ref: "#/channels/orders~1created/messages/orderCreated"},
				},
			},
			"receiveProductEvent": {
				Action: "receive",
				Channel: parser.ChannelRef{
					Ref: "#/channels/products~1updated",
				},
				Messages: []parser.MessageRef{
					{Ref: "#/channels/products~1updated/messages/productUpdated"},
				},
			},
		},
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"OrderEvent": {
					Name:        "OrderEvent",
					Title:       "Order created event",
					ContentType: "application/json",
					Payload:     createOrderEventSchema(),
				},
				"ProductEvent": {
					Name:        "ProductEvent",
					Title:       "Product updated event",
					ContentType: "application/json",
					Payload:     createProductEventSchema(),
				},
			},
		},
	}
}

// createUserSignupSchema создает схему сообщения UserSignup
func createUserSignupSchema() *parser.Schema {
	return &parser.Schema{
		Type: "object",
		Properties: map[string]parser.Schema{
			"userId": {
				Type:    "string",
				Example: "user123",
			},
			"email": {
				Type:    "string",
				Example: "user@example.com",
			},
			"name": {
				Type:    "string",
				Example: "John Doe",
			},
		},
		Required: []string{"userId", "email", "name"},
	}
}

// createOrderEventSchema создает схему сообщения OrderEvent
func createOrderEventSchema() *parser.Schema {
	return &parser.Schema{
		Type: "object",
		Properties: map[string]parser.Schema{
			"orderId": {
				Type:    "string",
				Example: "order123",
			},
			"customerId": {
				Type:    "string",
				Example: "customer456",
			},
			"amount": {
				Type:    "number",
				Example: 99.99,
			},
		},
		Required: []string{"orderId", "customerId", "amount"},
	}
}

// createProductEventSchema создает схему сообщения ProductEvent
func createProductEventSchema() *parser.Schema {
	return &parser.Schema{
		Type: "object",
		Properties: map[string]parser.Schema{
			"productId": {
				Type:    "string",
				Example: "prod789",
			},
			"productName": {
				Type:    "string",
				Example: "Widget",
			},
			"price": {
				Type:    "number",
				Example: 19.99,
			},
		},
		Required: []string{"productId", "productName", "price"},
	}
}