package validator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codemonsters/contract-tests-v2/parser"
)

func TestContractValidator(t *testing.T) {
	t.Run("should validate channels from ContractValidate structure", func(t *testing.T) {
		// Создаем mock спецификации
		consumerSpec := &parser.AsyncAPISpec{
			AsyncAPI: "3.0.0",
			Servers: map[string]parser.Server{
				"rabbitmq": {
					Host:     "localhost:5672",
					Protocol: "amqp",
				},
			},
			Channels: map[string]parser.Channel{
				"user/signedup": {
					Servers: []parser.ServerRef{
						{Ref: "#/servers/rabbitmq"},
					},
				},
			},
			Operations: map[string]parser.Operation{
				"sendUserSignup": {
					Action: "send",
					Channel: parser.ChannelRef{
						Ref: "#/channels/user~1signedup",
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
								"userId": {Type: "string"},
								"email":  {Type: "string"},
							},
							Required: []string{"userId", "email"},
						},
					},
				},
			},
		}

		providerSpec := &parser.AsyncAPISpec{
			AsyncAPI: "3.0.0",
			Servers: map[string]parser.Server{
				"rabbitmq": {
					Host:     "localhost:5672",
					Protocol: "amqp",
				},
			},
			Channels: map[string]parser.Channel{
				"notifications/user-events": {
					Servers: []parser.ServerRef{
						{Ref: "#/servers/rabbitmq"},
					},
				},
			},
			Operations: map[string]parser.Operation{
				"receiveUserSignup": {
					Action: "receive",
					Channel: parser.ChannelRef{
						Ref: "#/channels/notifications~1user-events",
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
								"userId": {Type: "string"},
								"email":  {Type: "string"},
							},
							Required: []string{"userId", "email"},
						},
					},
				},
			},
		}

		contractValidate := &ContractValidate{
			ConsumerChannelName: "user/signedup",
			ConsumerSpec:        consumerSpec,
			ProviderSpec:        providerSpec,
		}

		validator := NewChannelValidator()
		result, err := validator.ValidateChannels(contractValidate)

		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "user/signedup", result.ConsumerChannel.Name)
		assert.Equal(t, "amqp", result.ConsumerChannel.Protocol)
		assert.Equal(t, "notifications/user-events", result.ProviderChannel.Name)
		assert.Equal(t, "amqp", result.ProviderChannel.Protocol)
	})
}


func TestParser(t *testing.T) {
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

		parser := NewParser()
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

	t.Run("should handle invalid YAML", func(t *testing.T) {
		specContent := `asyncapi: 3.0.0
info:
  title: [invalid yaml
  version: 1.0.0`

		parser := NewParser()
		spec, err := parser.ParseFromString(specContent)

		assert.Error(t, err)
		assert.Nil(t, spec)
		assert.Contains(t, err.Error(), "failed to parse YAML")
	})
}

func createTempFile(t *testing.T, name, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, name)
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)
	return filePath
}
