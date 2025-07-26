package validator

import (
	"testing"

	"github.com/codemonsters/contract-tests-v2/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createValidConsumerSpec создает валидную спецификацию потребителя для тестов
func createValidConsumerSpec() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Channels: map[string]parser.Channel{
			"user/signedup": {
				Servers: []parser.ServerRef{{Ref: "#/servers/production"}},
			},
		},
		Servers: map[string]parser.Server{
			"production": {
				Host:     "broker.example.com",
				Protocol: "amqp",
			},
		},
		Operations: map[string]parser.Operation{
			"sendUserSignup": {
				Action:  "send",
				Channel: parser.ChannelRef{Ref: "#/channels/user~1signedup"},
				Messages: []parser.MessageRef{
					{Ref: "#/components/messages/UserSignup"},
				},
				Reply: &parser.Reply{
					Messages: []parser.MessageRef{
						{Ref: "#/components/messages/UserSignupReply"},
					},
				},
			},
		},
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
				"UserSignupReply": {
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"success": {Type: "boolean"},
							"message": {Type: "string"},
						},
						Required: []string{"success"},
					},
				},
			},
		},
	}
}

// createSpecWithChannels создает спецификацию с указанными каналами
func createSpecWithChannels(channels map[string]parser.Channel) *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Channels: channels,
	}
}

// createSpecWithChannelNoServers создает спецификацию с каналом без серверов
func createSpecWithChannelNoServers() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Channels: map[string]parser.Channel{
			"user/signedup": {
				// Нет серверов
			},
		},
	}
}

// createSpecWithChannelNoOperations создает спецификацию с каналом без операций
func createSpecWithChannelNoOperations() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Channels: map[string]parser.Channel{
			"user/signedup": {
				Servers: []parser.ServerRef{{Ref: "#/servers/production"}},
			},
		},
		Servers: map[string]parser.Server{
			"production": {
				Host:     "broker.example.com",
				Protocol: "amqp",
			},
		},
		Operations: map[string]parser.Operation{
			// Нет операций для канала
		},
		Components: &parser.Components{
			Messages: map[string]parser.Message{},
		},
	}
}

func TestExtractConsumerChannel(t *testing.T) {
	v := NewChannelValidator()

	t.Run("Валидное имя канала", func(t *testing.T) {
		// Arrange
		spec := createValidConsumerSpec()
		contractValidate := &ContractValidate{
			ConsumerChannelName: "user/signedup",
			ConsumerSpec:        spec,
		}

		// Act
		result, err := v.extractConsumerChannel(contractValidate)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "user/signedup", result.Name)
		assert.Equal(t, "amqp", result.Protocol)
		assert.NotNil(t, result.OutMessage)
		assert.Equal(t, "UserSignup", result.OutMessage.Name)
		assert.NotNil(t, result.InMessage)
		assert.Equal(t, "UserSignupReply", result.InMessage.Name)
	})

	t.Run("Несуществующий канал", func(t *testing.T) {
		// Arrange
		spec := createValidConsumerSpec()
		contractValidate := &ContractValidate{
			ConsumerChannelName: "non/existent",
			ConsumerSpec:        spec,
		}

		// Act
		result, err := v.extractConsumerChannel(contractValidate)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "channel non/existent not found in consumer spec")
	})

	t.Run("Канал без серверов (ошибка протокола)", func(t *testing.T) {
		// Arrange
		spec := createSpecWithChannelNoServers()
		contractValidate := &ContractValidate{
			ConsumerChannelName: "user/signedup",
			ConsumerSpec:        spec,
		}

		// Act
		result, err := v.extractConsumerChannel(contractValidate)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to extract protocol")
		assert.Contains(t, err.Error(), "no servers defined for channel")
	})

	t.Run("Канал без операций (нет сообщений)", func(t *testing.T) {
		// Arrange
		spec := createSpecWithChannelNoOperations()
		contractValidate := &ContractValidate{
			ConsumerChannelName: "user/signedup",
			ConsumerSpec:        spec,
		}

		// Act
		result, err := v.extractConsumerChannel(contractValidate)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "user/signedup", result.Name)
		assert.Equal(t, "amqp", result.Protocol)
		assert.Nil(t, result.OutMessage)
		assert.Nil(t, result.InMessage)
	})

	t.Run("Nil спецификация", func(t *testing.T) {
		// Arrange
		contractValidate := &ContractValidate{
			ConsumerChannelName: "user/signedup",
			ConsumerSpec:        nil,
		}

		// Act & Assert
		assert.Panics(t, func() {
			v.extractConsumerChannel(contractValidate)
		})
	})

	t.Run("Пустая спецификация", func(t *testing.T) {
		// Arrange
		spec := createSpecWithChannels(map[string]parser.Channel{})
		contractValidate := &ContractValidate{
			ConsumerChannelName: "user/signedup",
			ConsumerSpec:        spec,
		}

		// Act
		result, err := v.extractConsumerChannel(contractValidate)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "channel user/signedup not found in consumer spec")
	})
}