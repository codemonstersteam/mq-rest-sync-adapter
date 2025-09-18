package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractChannelProtocol(t *testing.T) {
	t.Run("should extract protocol from channel servers", func(t *testing.T) {
		spec := &AsyncAPISpec{
			Servers: map[string]Server{
				"rabbitmq": {
					Host:     "localhost:5672",
					Protocol: "amqp",
				},
			},
			Channels: map[string]Channel{
				"test/channel": {
					Servers: []ServerRef{
						{Ref: "#/servers/rabbitmq"},
					},
				},
			},
		}

		parser := New()
		channel := spec.Channels["test/channel"]
		protocol, err := parser.ExtractChannelProtocol(spec, &channel)

		require.NoError(t, err)
		assert.Equal(t, "amqp", protocol)
	})
}

func TestCompareMessageSchemas(t *testing.T) {
	t.Run("should find compatible messages", func(t *testing.T) {
		msg1 := &Message{
			ContentType: "application/json",
			Headers: &Schema{
				Type: "object",
				Properties: map[string]Schema{
					"correlationId": {Type: "string"},
					"timestamp":     {Type: "string", Format: "date-time"},
				},
				Required: []string{"correlationId"},
			},
			Payload: &Schema{
				Type: "object",
				Properties: map[string]Schema{
					"userId": {Type: "string"},
					"email":  {Type: "string", Format: "email"},
				},
				Required: []string{"userId", "email"},
			},
		}

		msg2 := &Message{
			ContentType: "application/json",
			Headers: &Schema{
				Type: "object",
				Properties: map[string]Schema{
					"correlationId": {Type: "string"},
					"timestamp":     {Type: "string", Format: "date-time"},
					"source":        {Type: "string"}, // Дополнительное поле - OK
				},
				Required: []string{"correlationId"}, // Те же required поля
			},
			Payload: &Schema{
				Type: "object",
				Properties: map[string]Schema{
					"userId":   {Type: "string"},
					"email":    {Type: "string", Format: "email"},
					"username": {Type: "string"}, // Дополнительное поле - OK
				},
				Required: []string{"userId", "email"}, // Те же required поля
			},
		}

		parser := New()
		compatible, details := parser.CompareMessageSchemas(msg1, msg2)

		assert.True(t, compatible)
		assert.Equal(t, "messages are compatible", details)
	})

	t.Run("should find incompatible messages with different required fields", func(t *testing.T) {
		msg1 := &Message{
			ContentType: "application/json",
			Payload: &Schema{
				Type: "object",
				Properties: map[string]Schema{
					"userId": {Type: "string"},
					"email":  {Type: "string"},
				},
				Required: []string{"userId", "email"},
			},
		}

		msg2 := &Message{
			ContentType: "application/json",
			Payload: &Schema{
				Type: "object",
				Properties: map[string]Schema{
					"userId": {Type: "string"},
				},
				Required: []string{"userId"}, // Отсутствует required поле email
			},
		}

		parser := New()
		compatible, details := parser.CompareMessageSchemas(msg1, msg2)

		assert.False(t, compatible)
		assert.Equal(t, "payload schemas are incompatible", details)
	})

	t.Run("should find incompatible messages with different content types", func(t *testing.T) {
		msg1 := &Message{
			ContentType: "application/json",
			Payload:     &Schema{Type: "object"},
		}

		msg2 := &Message{
			ContentType: "application/xml",
			Payload:     &Schema{Type: "object"},
		}

		parser := New()
		compatible, details := parser.CompareMessageSchemas(msg1, msg2)

		assert.False(t, compatible)
		assert.Contains(t, details, "content types differ")
	})
}