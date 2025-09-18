package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFromString(t *testing.T) {
	t.Run("should parse valid AsyncAPI 3.0 specification", func(t *testing.T) {
		specContent := `asyncapi: 3.0.0
info:
  title: Test Service
  version: 1.0.0
  description: Test service for parser
servers:
  rabbitmq:
    host: localhost:5672
    protocol: amqp
    description: RabbitMQ broker
channels:
  user/signedup:
    address: user.signedup
    servers:
      - $ref: '#/servers/rabbitmq'
    messages:
      userSignedUp:
        $ref: '#/components/messages/UserSignedUp'
operations:
  publishUserSignedUp:
    action: send
    channel:
      $ref: '#/channels/user~1signedup'
    messages:
      - $ref: '#/channels/user~1signedup/messages/userSignedUp'
components:
  messages:
    UserSignedUp:
      name: UserSignedUp
      title: User signed up event
      contentType: application/json
      headers:
        type: object
        properties:
          correlationId:
            type: string
            description: Correlation ID for tracking
          timestamp:
            type: string
            format: date-time
        required:
          - correlationId
      payload:
        type: object
        properties:
          userId:
            type: string
            description: User identifier
          email:
            type: string
            format: email
        required:
          - userId
          - email`

		parser := New()
		spec, err := parser.ParseFromString(specContent)

		require.NoError(t, err)
		require.NotNil(t, spec)

		// Проверка основных полей
		assert.Equal(t, "3.0.0", spec.AsyncAPI)
		assert.Equal(t, "Test Service", spec.Info.Title)
		assert.Equal(t, "1.0.0", spec.Info.Version)

		// Проверка серверов
		assert.Len(t, spec.Servers, 1)
		assert.Contains(t, spec.Servers, "rabbitmq")
		assert.Equal(t, "amqp", spec.Servers["rabbitmq"].Protocol)

		// Проверка каналов
		assert.Len(t, spec.Channels, 1)
		assert.Contains(t, spec.Channels, "user/signedup")

		// Проверка операций
		assert.Len(t, spec.Operations, 1)
		assert.Contains(t, spec.Operations, "publishUserSignedUp")
		assert.Equal(t, "send", spec.Operations["publishUserSignedUp"].Action)

		// Проверка компонентов
		require.NotNil(t, spec.Components)
		assert.Len(t, spec.Components.Messages, 1)
		
		// Проверка сообщения с headers
		msg := spec.Components.Messages["UserSignedUp"]
		assert.Equal(t, "UserSignedUp", msg.Name)
		assert.Equal(t, "application/json", msg.ContentType)
		
		// Проверка headers
		require.NotNil(t, msg.Headers)
		assert.Equal(t, "object", msg.Headers.Type)
		assert.Contains(t, msg.Headers.Properties, "correlationId")
		assert.Contains(t, msg.Headers.Required, "correlationId")
		
		// Проверка payload
		require.NotNil(t, msg.Payload)
		assert.Equal(t, "object", msg.Payload.Type)
		assert.Contains(t, msg.Payload.Properties, "userId")
		assert.Contains(t, msg.Payload.Properties, "email")
		assert.Contains(t, msg.Payload.Required, "userId")
		assert.Contains(t, msg.Payload.Required, "email")
	})

	t.Run("should parse specification with message bindings", func(t *testing.T) {
		specContent := `asyncapi: 3.0.0
info:
  title: Test Service
  version: 1.0.0
channels:
  test/channel:
    address: test.channel
    bindings:
      amqp:
        is: queue
        queue:
          durable: true
components:
  messages:
    TestMessage:
      name: TestMessage
      contentType: application/json
      correlationId:
        location: $message.header#/correlationId
        description: Correlation ID for message tracking
      bindings:
        amqp:
          contentEncoding: gzip
          messageType: TestMessage
          bindingVersion: 0.2.0
      payload:
        type: object`

		parser := New()
		spec, err := parser.ParseFromString(specContent)

		require.NoError(t, err)
		require.NotNil(t, spec)

		// Проверка channel bindings
		channel := spec.Channels["test/channel"]
		assert.NotNil(t, channel.Bindings)
		assert.Contains(t, channel.Bindings, "amqp")

		// Проверка message bindings и correlationId
		msg := spec.Components.Messages["TestMessage"]
		assert.NotNil(t, msg.Bindings)
		assert.Contains(t, msg.Bindings, "amqp")
		
		require.NotNil(t, msg.CorrelationId)
		assert.Equal(t, "$message.header#/correlationId", msg.CorrelationId.Location)
		assert.Equal(t, "Correlation ID for message tracking", msg.CorrelationId.Description)
	})

	t.Run("should handle invalid YAML", func(t *testing.T) {
		specContent := `asyncapi: 3.0.0
info:
  title: [invalid yaml
  version: 1.0.0`

		parser := New()
		spec, err := parser.ParseFromString(specContent)

		assert.Error(t, err)
		assert.Nil(t, spec)
		// Проверяем стандартизированное сообщение об ошибке
		assert.Contains(t, err.Error(), "YAML_PARSE_ERROR")
		assert.Contains(t, err.Error(), "at input")
	})

	t.Run("should handle empty string", func(t *testing.T) {
		parser := New()
		spec, err := parser.ParseFromString("")

		assert.Error(t, err)
		assert.Nil(t, spec)
		// Проверяем стандартизированное сообщение об ошибке
		assert.Contains(t, err.Error(), "PARSE_ERROR")
		assert.Contains(t, err.Error(), "specification content is empty")
		assert.Contains(t, err.Error(), "at input")
	})

	t.Run("should handle missing asyncapi version", func(t *testing.T) {
		specContent := `info:
  title: Test Service
  version: 1.0.0`

		parser := New()
		spec, err := parser.ParseFromString(specContent)

		assert.Error(t, err)
		assert.Nil(t, spec)
		// Проверяем стандартизированное сообщение об ошибке
		assert.Contains(t, err.Error(), "PARSE_ERROR")
		assert.Contains(t, err.Error(), "asyncapi version is required")
		assert.Contains(t, err.Error(), "at spec.asyncapi")
	})

	t.Run("should handle invalid asyncapi version", func(t *testing.T) {
		specContent := `asyncapi: 2.6.0
info:
  title: Test Service
  version: 1.0.0`

		parser := New()
		spec, err := parser.ParseFromString(specContent)

		assert.Error(t, err)
		assert.Nil(t, spec)
		// Проверяем стандартизированное сообщение об ошибке
		assert.Contains(t, err.Error(), "INVALID_VERSION_ERROR")
		assert.Contains(t, err.Error(), "unsupported AsyncAPI version '2.6.0'")
		assert.Contains(t, err.Error(), "at spec.asyncapi")
	})
}