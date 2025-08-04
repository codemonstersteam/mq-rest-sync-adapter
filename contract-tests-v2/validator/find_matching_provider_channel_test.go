package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codemonsters/contract-tests-v2/parser"
)

// TestCase структура для table-driven тестов функции findMatchingProviderChannel
type FindMatchingProviderChannelTestCase struct {
	name                 string
	consumerSpec         *parser.AsyncAPISpec
	providerSpec         *parser.AsyncAPISpec
	consumerChannelName  string
	expectedMatch        bool
	expectedChannelName  string
	expectedErrorContains string
	communicationPattern string // "request-reply", "fire-and-forget", "pub-sub"
	protocol             string // "amqp", "mqtt", "kafka", "ws"
	description          string
}

// Базовая функция для создания тестовых спецификаций
func createBasicConsumerSpec() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		AsyncAPI: "3.0.0",
		Info: parser.Info{
			Title:   "Consumer Service",
			Version: "1.0.0",
		},
		Servers: map[string]parser.Server{
			"production": {
				Host:     "localhost:5672",
				Protocol: "amqp",
			},
		},
		Channels: map[string]parser.Channel{
			"user/events": {
				Address: "user.events",
				Messages: map[string]parser.MessageRef{
					"userSignup": {
						Ref: "#/components/messages/UserSignup",
					},
				},
				Servers: []parser.ServerRef{
					{Ref: "#/servers/production"},
				},
			},
		},
		Operations: map[string]parser.Operation{
			"sendUserSignup": {
				Action: "send",
				Channel: parser.ChannelRef{
					Ref: "#/channels/user~1events",
				},
				Messages: []parser.MessageRef{
					{Ref: "#/components/messages/UserSignup"},
				},
				Reply: &parser.Reply{
					Channel: parser.ChannelRef{
						Ref: "#/channels/user~1events",
					},
					Messages: []parser.MessageRef{
						{Ref: "#/components/messages/UserSignupResponse"},
					},
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
				"UserSignupResponse": {
					Name:        "UserSignupResponse",
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"success": {Type: "boolean"},
							"userId":  {Type: "string"},
						},
						Required: []string{"success", "userId"},
					},
				},
			},
		},
	}
}

func createBasicProviderSpec() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		AsyncAPI: "3.0.0",
		Info: parser.Info{
			Title:   "Provider Service",
			Version: "1.0.0",
		},
		Servers: map[string]parser.Server{
			"production": {
				Host:     "localhost:5672",
				Protocol: "amqp",
			},
		},
		Channels: map[string]parser.Channel{
			"user/events": {
				Address: "user.events",
				Messages: map[string]parser.MessageRef{
					"userSignup": {
						Ref: "#/components/messages/UserSignup",
					},
				},
				Servers: []parser.ServerRef{
					{Ref: "#/servers/production"},
				},
			},
		},
		Operations: map[string]parser.Operation{
			"receiveUserSignup": {
				Action: "receive",
				Channel: parser.ChannelRef{
					Ref: "#/channels/user~1events",
				},
				Messages: []parser.MessageRef{
					{Ref: "#/components/messages/UserSignup"},
				},
				Reply: &parser.Reply{
					Channel: parser.ChannelRef{
						Ref: "#/channels/user~1events",
					},
					Messages: []parser.MessageRef{
						{Ref: "#/components/messages/UserSignupResponse"},
					},
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
				"UserSignupResponse": {
					Name:        "UserSignupResponse",
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"success": {Type: "boolean"},
							"userId":  {Type: "string"},
						},
						Required: []string{"success", "userId"},
					},
				},
			},
		},
	}
}

func createProviderSpecWithExtraFields() *parser.AsyncAPISpec {
	spec := createBasicProviderSpec()
	
	// Добавляем дополнительные поля к UserSignup схеме
	spec.Components.Messages["UserSignup"].Payload.Properties["timestamp"] = parser.Schema{Type: "string"}
	spec.Components.Messages["UserSignup"].Payload.Properties["source"] = parser.Schema{Type: "string"}
	
	// Добавляем дополнительные поля к UserSignupResponse схеме  
	spec.Components.Messages["UserSignupResponse"].Payload.Properties["timestamp"] = parser.Schema{Type: "string"}
	spec.Components.Messages["UserSignupResponse"].Payload.Properties["code"] = parser.Schema{Type: "integer"}
	
	return spec
}

func createProviderSpecWithIncompatibleResponse() *parser.AsyncAPISpec {
	spec := createBasicProviderSpec()
	
	// Делаем response несовместимым - меняем тип success на string вместо boolean
	spec.Components.Messages["UserSignupResponse"].Payload.Properties["success"] = parser.Schema{Type: "string"}
	// Удаляем обязательное поле userId
	spec.Components.Messages["UserSignupResponse"].Payload.Required = []string{"success"}
	
	return spec
}

func createProviderSpecWithIncompatibleRequest() *parser.AsyncAPISpec {
	spec := createBasicProviderSpec()
	
	// Делаем request несовместимым - меняем тип email на integer вместо string
	spec.Components.Messages["UserSignup"].Payload.Properties["email"] = parser.Schema{Type: "integer"}
	// Добавляем дополнительное обязательное поле
	spec.Components.Messages["UserSignup"].Payload.Properties["phone"] = parser.Schema{Type: "string"}
	spec.Components.Messages["UserSignup"].Payload.Required = []string{"userId", "email", "phone"}
	
	return spec
}

func createProviderSpecWithBothIncompatibleMessages() *parser.AsyncAPISpec {
	spec := createBasicProviderSpec()
	
	// Делаем request несовместимым
	spec.Components.Messages["UserSignup"].Payload.Properties["email"] = parser.Schema{Type: "integer"}
	spec.Components.Messages["UserSignup"].Payload.Properties["phone"] = parser.Schema{Type: "string"}
	spec.Components.Messages["UserSignup"].Payload.Required = []string{"userId", "email", "phone"}
	
	// Делаем response несовместимым
	spec.Components.Messages["UserSignupResponse"].Payload.Properties["success"] = parser.Schema{Type: "string"}
	spec.Components.Messages["UserSignupResponse"].Payload.Required = []string{"success"}
	
	return spec
}

func createFireAndForgetConsumerSpec() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		AsyncAPI: "3.0.0",
		Info: parser.Info{
			Title:   "Consumer Service Fire-and-Forget",
			Version: "1.0.0",
		},
		Servers: map[string]parser.Server{
			"production": {
				Host:     "localhost:1883",
				Protocol: "mqtt",
			},
		},
		Channels: map[string]parser.Channel{
			"notifications/send": {
				Address: "notifications.send",
				Messages: map[string]parser.MessageRef{
					"notification": {
						Ref: "#/components/messages/NotificationMessage",
					},
				},
				Servers: []parser.ServerRef{
					{Ref: "#/servers/production"},
				},
			},
		},
		Operations: map[string]parser.Operation{
			"sendNotification": {
				Action: "send",
				Channel: parser.ChannelRef{
					Ref: "#/channels/notifications~1send",
				},
				Messages: []parser.MessageRef{
					{Ref: "#/components/messages/NotificationMessage"},
				},
				// Нет Reply - это Fire-and-Forget
			},
		},
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"NotificationMessage": {
					Name:        "NotificationMessage",
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"message": {Type: "string"},
							"userId":  {Type: "string"},
						},
						Required: []string{"message", "userId"},
					},
				},
			},
		},
	}
}

func createFireAndForgetProviderSpec() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		AsyncAPI: "3.0.0",
		Info: parser.Info{
			Title:   "Provider Service Fire-and-Forget",
			Version: "1.0.0",
		},
		Servers: map[string]parser.Server{
			"production": {
				Host:     "localhost:1883",
				Protocol: "mqtt",
			},
		},
		Channels: map[string]parser.Channel{
			"notifications/send": {
				Address: "notifications.send",
				Messages: map[string]parser.MessageRef{
					"notification": {
						Ref: "#/components/messages/NotificationMessage",
					},
				},
				Servers: []parser.ServerRef{
					{Ref: "#/servers/production"},
				},
			},
		},
		Operations: map[string]parser.Operation{
			"receiveNotification": {
				Action: "receive",
				Channel: parser.ChannelRef{
					Ref: "#/channels/notifications~1send",
				},
				Messages: []parser.MessageRef{
					{Ref: "#/components/messages/NotificationMessage"},
				},
				// Нет Reply - это Fire-and-Forget
			},
		},
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"NotificationMessage": {
					Name:        "NotificationMessage",
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"message": {Type: "string"},
							"userId":  {Type: "string"},
						},
						Required: []string{"message", "userId"},
					},
				},
			},
		},
	}
}

func createProviderSpecWithDifferentProtocol() *parser.AsyncAPISpec {
	spec := createBasicProviderSpec()
	// Создаем новый сервер с другим протоколом
	kafkaServer := spec.Servers["production"]
	kafkaServer.Protocol = "kafka"
	spec.Servers["production"] = kafkaServer
	return spec
}

func createProviderSpecWithoutServers() *parser.AsyncAPISpec {
	spec := createBasicProviderSpec()
	spec.Servers = map[string]parser.Server{} // Пустые серверы
	return spec
}

func createEmptyProviderSpec() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		AsyncAPI: "3.0.0",
		Info: parser.Info{
			Title:   "Empty Provider Service",
			Version: "1.0.0",
		},
		Servers: map[string]parser.Server{
			"production": {
				Host:     "localhost:5672",
				Protocol: "amqp",
			},
		},
		Channels: map[string]parser.Channel{}, // Пустые каналы
		Operations: map[string]parser.Operation{},
		Components: &parser.Components{
			Messages: map[string]parser.Message{},
		},
	}
}

func createProviderSpecWithoutOperations() *parser.AsyncAPISpec {
	spec := createBasicProviderSpec()
	spec.Operations = map[string]parser.Operation{} // Пустые операции
	return spec
}

func createProviderSpecWithXMLContentType() *parser.AsyncAPISpec {
	spec := createBasicProviderSpec()
	
	// Меняем ContentType на XML
	userSignup := spec.Components.Messages["UserSignup"]
	userSignup.ContentType = "application/xml"
	spec.Components.Messages["UserSignup"] = userSignup
	
	userSignupResponse := spec.Components.Messages["UserSignupResponse"]
	userSignupResponse.ContentType = "application/xml"
	spec.Components.Messages["UserSignupResponse"] = userSignupResponse
	
	return spec
}

// TestFindMatchingProviderChannel_RequestReply тестирует Request-Reply паттерн
func TestFindMatchingProviderChannel_RequestReply(t *testing.T) {
	validator := NewChannelValidator()

	testCases := []FindMatchingProviderChannelTestCase{
		{
			name:                "request_reply_perfect_match",
			consumerSpec:        createBasicConsumerSpec(),
			providerSpec:        createBasicProviderSpec(),
			consumerChannelName: "user/events",
			expectedMatch:       true,
			expectedChannelName: "user/events",
			expectedErrorContains: "",
			communicationPattern: "request-reply",
			protocol:            "amqp",
			description:         "Полная совместимость: consumer(out+in) ↔ provider(in+out)",
		},
		{
			name:                "request_reply_compatible_with_extra_fields",
			consumerSpec:        createBasicConsumerSpec(),
			providerSpec:        createProviderSpecWithExtraFields(),
			consumerChannelName: "user/events",
			expectedMatch:       true,
			expectedChannelName: "user/events",
			expectedErrorContains: "",
			communicationPattern: "request-reply",
			protocol:            "amqp",
			description:         "Совместимые типы с расширенными полями у поставщика",
		},
		{
			name:                "request_reply_identical_schemas",
			consumerSpec:        createBasicConsumerSpec(),
			providerSpec:        createBasicProviderSpec(),
			consumerChannelName: "user/events",
			expectedMatch:       true,
			expectedChannelName: "user/events",
			expectedErrorContains: "",
			communicationPattern: "request-reply",
			protocol:            "amqp",
			description:         "Идентичные схемы сообщений для request и response",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			contractValidate := &ContractValidate{
				ConsumerChannelName: tc.consumerChannelName,
				ProviderSpec:        tc.providerSpec,
				ConsumerSpec:        tc.consumerSpec,
			}

			// Первоначально извлекаем канал потребителя
			consumerChannel, err := validator.extractConsumerChannel(contractValidate)
			require.NoError(t, err, "Failed to extract consumer channel for test: %s", tc.name)

			// Act
			result, err := validator.findMatchingProviderChannel(contractValidate, consumerChannel)

			// Assert
			if tc.expectedMatch {
				require.NoError(t, err, "Expected successful match for test case: %s", tc.name)
				require.NotNil(t, result, "Expected non-nil result for test case: %s", tc.name)
				assert.Equal(t, tc.expectedChannelName, result.Name, "Channel name mismatch for test case: %s", tc.name)
				assert.Equal(t, tc.protocol, result.Protocol, "Protocol mismatch for test case: %s", tc.name)
				
				// Проверяем, что у нас есть как входящие так и исходящие сообщения для Request-Reply
				if tc.communicationPattern == "request-reply" {
					assert.NotNil(t, result.InMessage, "Expected InMessage for Request-Reply pattern: %s", tc.name)
					assert.NotNil(t, result.OutMessage, "Expected OutMessage for Request-Reply pattern: %s", tc.name)
				}
			} else {
				require.Error(t, err, "Expected error for test case: %s", tc.name)
				if tc.expectedErrorContains != "" {
					assert.Contains(t, err.Error(), tc.expectedErrorContains, "Error message mismatch for test case: %s", tc.name)
				}
				assert.Nil(t, result, "Expected nil result for failed test case: %s", tc.name)
			}
		})
	}
}

// TestFindMatchingProviderChannel_RequestReply_Incompatible тестирует несовместимые кейсы Request-Reply паттерна
func TestFindMatchingProviderChannel_RequestReply_Incompatible(t *testing.T) {
	validator := NewChannelValidator()

	testCases := []FindMatchingProviderChannelTestCase{
		{
			name:                "request_reply_request_compatible_response_incompatible",
			consumerSpec:        createBasicConsumerSpec(),
			providerSpec:        createProviderSpecWithIncompatibleResponse(),
			consumerChannelName: "user/events",
			expectedMatch:       false,
			expectedChannelName: "",
			expectedErrorContains: "no compatible provider channel found",
			communicationPattern: "request-reply",
			protocol:            "amqp",
			description:         "Request совместим, response несовместим",
		},
		{
			name:                "request_reply_request_incompatible_response_compatible",
			consumerSpec:        createBasicConsumerSpec(),
			providerSpec:        createProviderSpecWithIncompatibleRequest(),
			consumerChannelName: "user/events",
			expectedMatch:       false,
			expectedChannelName: "",
			expectedErrorContains: "no compatible provider channel found",
			communicationPattern: "request-reply",
			protocol:            "amqp",
			description:         "Request несовместим, response совместим",
		},
		{
			name:                "request_reply_both_messages_incompatible",
			consumerSpec:        createBasicConsumerSpec(),
			providerSpec:        createProviderSpecWithBothIncompatibleMessages(),
			consumerChannelName: "user/events",
			expectedMatch:       false,
			expectedChannelName: "",
			expectedErrorContains: "no compatible provider channel found",
			communicationPattern: "request-reply",
			protocol:            "amqp",
			description:         "Оба сообщения несовместимы",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			contractValidate := &ContractValidate{
				ConsumerChannelName: tc.consumerChannelName,
				ProviderSpec:        tc.providerSpec,
				ConsumerSpec:        tc.consumerSpec,
			}

			// Первоначально извлекаем канал потребителя
			consumerChannel, err := validator.extractConsumerChannel(contractValidate)
			require.NoError(t, err, "Failed to extract consumer channel for test: %s", tc.name)

			// Act
			result, err := validator.findMatchingProviderChannel(contractValidate, consumerChannel)

			// Assert
			if tc.expectedMatch {
				require.NoError(t, err, "Expected successful match for test case: %s", tc.name)
				require.NotNil(t, result, "Expected non-nil result for test case: %s", tc.name)
				assert.Equal(t, tc.expectedChannelName, result.Name, "Channel name mismatch for test case: %s", tc.name)
				assert.Equal(t, tc.protocol, result.Protocol, "Protocol mismatch for test case: %s", tc.name)
			} else {
				require.Error(t, err, "Expected error for test case: %s", tc.name)
				if tc.expectedErrorContains != "" {
					assert.Contains(t, err.Error(), tc.expectedErrorContains, "Error message mismatch for test case: %s", tc.name)
				}
				assert.Nil(t, result, "Expected nil result for failed test case: %s", tc.name)
			}
		})
	}
}

// TestFindMatchingProviderChannel_FireAndForget тестирует Fire-and-Forget паттерн
func TestFindMatchingProviderChannel_FireAndForget(t *testing.T) {
	validator := NewChannelValidator()

	testCases := []FindMatchingProviderChannelTestCase{
		{
			name:                "fire_and_forget_consumer_send_provider_receive",
			consumerSpec:        createFireAndForgetConsumerSpec(),
			providerSpec:        createFireAndForgetProviderSpec(),
			consumerChannelName: "notifications/send",
			expectedMatch:       true,
			expectedChannelName: "notifications/send",
			expectedErrorContains: "",
			communicationPattern: "fire-and-forget",
			protocol:            "mqtt",
			description:         "Consumer только send (OutMessage), Provider только receive (InMessage)",
		},
		{
			name:                "fire_and_forget_no_reply_operations",
			consumerSpec:        createFireAndForgetConsumerSpec(),
			providerSpec:        createFireAndForgetProviderSpec(),
			consumerChannelName: "notifications/send",
			expectedMatch:       true,
			expectedChannelName: "notifications/send",
			expectedErrorContains: "",
			communicationPattern: "fire-and-forget",
			protocol:            "mqtt",
			description:         "Отсутствие reply операций у обеих сторон",
		},
		{
			name:                "fire_and_forget_consumer_expects_reply_provider_no_reply",
			consumerSpec:        createBasicConsumerSpec(), // У него есть reply
			providerSpec:        createFireAndForgetProviderSpec(), // У него нет reply
			consumerChannelName: "user/events",
			expectedMatch:       false,
			expectedChannelName: "",
			expectedErrorContains: "no compatible provider channel found",
			communicationPattern: "fire-and-forget",
			protocol:            "amqp",
			description:         "Consumer ожидает ответ, но Provider не отвечает",
		},
		{
			name:                "fire_and_forget_provider_sends_reply_consumer_no_expect",
			consumerSpec:        createFireAndForgetConsumerSpec(), // У него нет reply
			providerSpec:        createBasicProviderSpec(), // У него есть reply
			consumerChannelName: "notifications/send",
			expectedMatch:       false,
			expectedChannelName: "",
			expectedErrorContains: "no compatible provider channel found",
			communicationPattern: "fire-and-forget",
			protocol:            "mqtt",
			description:         "Provider отправляет ответ, но Consumer не ждет",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			contractValidate := &ContractValidate{
				ConsumerChannelName: tc.consumerChannelName,
				ProviderSpec:        tc.providerSpec,
				ConsumerSpec:        tc.consumerSpec,
			}

			// Первоначально извлекаем канал потребителя
			consumerChannel, err := validator.extractConsumerChannel(contractValidate)
			require.NoError(t, err, "Failed to extract consumer channel for test: %s", tc.name)

			// Act
			result, err := validator.findMatchingProviderChannel(contractValidate, consumerChannel)

			// Assert
			if tc.expectedMatch {
				require.NoError(t, err, "Expected successful match for test case: %s", tc.name)
				require.NotNil(t, result, "Expected non-nil result for test case: %s", tc.name)
				assert.Equal(t, tc.expectedChannelName, result.Name, "Channel name mismatch for test case: %s", tc.name)
				assert.Equal(t, tc.protocol, result.Protocol, "Protocol mismatch for test case: %s", tc.name)
				
				// Проверяем, что у нас есть только входящие сообщения для Fire-and-Forget
				if tc.communicationPattern == "fire-and-forget" {
					assert.NotNil(t, result.InMessage, "Expected InMessage for Fire-and-Forget pattern: %s", tc.name)
					assert.Nil(t, result.OutMessage, "Expected no OutMessage for Fire-and-Forget pattern: %s", tc.name)
				}
			} else {
				require.Error(t, err, "Expected error for test case: %s", tc.name)
				if tc.expectedErrorContains != "" {
					assert.Contains(t, err.Error(), tc.expectedErrorContains, "Error message mismatch for test case: %s", tc.name)
				}
				assert.Nil(t, result, "Expected nil result for failed test case: %s", tc.name)
			}
		})
	}
}

// TestFindMatchingProviderChannel_EdgeCases тестирует граничные случаи и ошибки
func TestFindMatchingProviderChannel_EdgeCases(t *testing.T) {
	validator := NewChannelValidator()

	testCases := []FindMatchingProviderChannelTestCase{
		{
			name:                "no_channels_with_matching_protocol",
			consumerSpec:        createBasicConsumerSpec(), // amqp
			providerSpec:        createProviderSpecWithDifferentProtocol(), // kafka
			consumerChannelName: "user/events",
			expectedMatch:       false,
			expectedChannelName: "",
			expectedErrorContains: "no compatible provider channel found",
			communicationPattern: "request-reply",
			protocol:            "amqp",
			description:         "Нет каналов с нужным протоколом",
		},
		{
			name:                "provider_spec_missing_servers",
			consumerSpec:        createBasicConsumerSpec(),
			providerSpec:        createProviderSpecWithoutServers(),
			consumerChannelName: "user/events",
			expectedMatch:       false,
			expectedChannelName: "",
			expectedErrorContains: "no compatible provider channel found",
			communicationPattern: "request-reply",
			protocol:            "amqp",
			description:         "Отсутствующие servers в спецификации поставщика",
		},
		{
			name:                "provider_spec_missing_channels",
			consumerSpec:        createBasicConsumerSpec(),
			providerSpec:        createEmptyProviderSpec(),
			consumerChannelName: "user/events",
			expectedMatch:       false,
			expectedChannelName: "",
			expectedErrorContains: "no compatible provider channel found",
			communicationPattern: "request-reply",
			protocol:            "amqp",
			description:         "Отсутствующие channels в спецификации поставщика",
		},
		{
			name:                "provider_spec_missing_operations",
			consumerSpec:        createBasicConsumerSpec(),
			providerSpec:        createProviderSpecWithoutOperations(),
			consumerChannelName: "user/events",
			expectedMatch:       false,
			expectedChannelName: "",
			expectedErrorContains: "no compatible provider channel found",
			communicationPattern: "request-reply",
			protocol:            "amqp",
			description:         "Отсутствующие operations в спецификации поставщика",
		},
		{
			name:                "different_content_types",
			consumerSpec:        createBasicConsumerSpec(), // application/json
			providerSpec:        createProviderSpecWithXMLContentType(), // application/xml
			consumerChannelName: "user/events",
			expectedMatch:       false,
			expectedChannelName: "",
			expectedErrorContains: "no compatible provider channel found",
			communicationPattern: "request-reply",
			protocol:            "amqp",
			description:         "Различные ContentType (application/json vs application/xml)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			contractValidate := &ContractValidate{
				ConsumerChannelName: tc.consumerChannelName,
				ProviderSpec:        tc.providerSpec,
				ConsumerSpec:        tc.consumerSpec,
			}

			// Первоначально извлекаем канал потребителя
			consumerChannel, err := validator.extractConsumerChannel(contractValidate)
			require.NoError(t, err, "Failed to extract consumer channel for test: %s", tc.name)

			// Act
			result, err := validator.findMatchingProviderChannel(contractValidate, consumerChannel)

			// Assert
			require.Error(t, err, "Expected error for test case: %s", tc.name)
			assert.Contains(t, err.Error(), tc.expectedErrorContains, "Error message mismatch for test case: %s", tc.name)
			assert.Nil(t, result, "Expected nil result for failed test case: %s", tc.name)
		})
	}
}

// TestFindMatchingProviderChannel_CorrelationId тестирует поддержку Correlation ID для Request-Reply паттерна
func TestFindMatchingProviderChannel_CorrelationId(t *testing.T) {
	validator := NewChannelValidator()

	testCases := []FindMatchingProviderChannelTestCase{
		{
			name:                "correlation_id_supported_by_both",
			consumerSpec:        createConsumerSpecWithCorrelationId(),
			providerSpec:        createProviderSpecWithCorrelationId(),
			consumerChannelName: "user/events",
			expectedMatch:       true,
			expectedChannelName: "user/events",
			expectedErrorContains: "",
			communicationPattern: "request-reply",
			protocol:            "amqp",
			description:         "Consumer и Provider поддерживают Correlation ID - полная совместимость",
		},
		{
			name:                "consumer_requires_correlation_id_provider_missing",
			consumerSpec:        createConsumerSpecWithCorrelationId(),
			providerSpec:        createBasicProviderSpec(), // без correlation ID
			consumerChannelName: "user/events",
			expectedMatch:       false,
			expectedChannelName: "",
			expectedErrorContains: "no compatible provider channel found",
			communicationPattern: "request-reply",
			protocol:            "amqp",
			description:         "Consumer требует Correlation ID для связывания ответов, Provider не поддерживает - несовместимо",
		},
		{
			name:                "provider_requires_correlation_id_consumer_missing",
			consumerSpec:        createBasicConsumerSpec(), // без correlation ID
			providerSpec:        createProviderSpecWithCorrelationId(),
			consumerChannelName: "user/events",
			expectedMatch:       false,
			expectedChannelName: "",
			expectedErrorContains: "no compatible provider channel found",
			communicationPattern: "request-reply",
			protocol:            "amqp",
			description:         "Provider требует Correlation ID для обработки запросов, Consumer не предоставляет - несовместимо",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			contractValidate := &ContractValidate{
				ConsumerChannelName: tc.consumerChannelName,
				ProviderSpec:        tc.providerSpec,
				ConsumerSpec:        tc.consumerSpec,
			}

			consumerChannel, err := validator.extractConsumerChannel(contractValidate)
			require.NoError(t, err, "Failed to extract consumer channel for test: %s", tc.name)

			// Act
			result, err := validator.findMatchingProviderChannel(contractValidate, consumerChannel)

			// Assert
			if tc.expectedMatch {
				require.NoError(t, err, "Expected successful match for test case: %s", tc.name)
				require.NotNil(t, result, "Expected non-nil result for test case: %s", tc.name)
				assert.Equal(t, tc.expectedChannelName, result.Name, "Channel name mismatch for test case: %s", tc.name)
				assert.Equal(t, tc.protocol, result.Protocol, "Protocol mismatch for test case: %s", tc.name)

				t.Logf("✅ %s: Found matching channel '%s' with protocol '%s'", tc.description, result.Name, result.Protocol)
			} else {
				require.Error(t, err, "Expected error for test case: %s", tc.name)
				assert.Contains(t, err.Error(), tc.expectedErrorContains, "Error message mismatch for test case: %s", tc.name)
				assert.Nil(t, result, "Expected nil result for failed test case: %s", tc.name)

				t.Logf("❌ %s: %v", tc.description, err)
			}
		})
	}
}

// Helper функции для создания спецификаций с Correlation ID

func createConsumerSpecWithCorrelationId() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		AsyncAPI: "3.0.0",
		Info: parser.Info{
			Title:   "Consumer Service with Correlation ID",
			Version: "1.0.0",
		},
		Servers: map[string]parser.Server{
			"production": {
				Host:     "localhost:5672",
				Protocol: "amqp",
			},
		},
		Channels: map[string]parser.Channel{
			"user/events": {
				Address: "user.events",
				Messages: map[string]parser.MessageRef{
					"userSignup": {
						Ref: "#/components/messages/UserSignupWithCorrelation",
					},
				},
				Servers: []parser.ServerRef{
					{Ref: "#/servers/production"},
				},
			},
		},
		Operations: map[string]parser.Operation{
			"sendUserSignup": {
				Action:  "send",
				Channel: parser.ChannelRef{Ref: "#/channels/user~1events"},
				Messages: []parser.MessageRef{
					{Ref: "#/components/messages/UserSignupWithCorrelation"},
				},
				Reply: &parser.Reply{
					Channel: parser.ChannelRef{Ref: "#/channels/user~1events"},
					Messages: []parser.MessageRef{
						{Ref: "#/components/messages/UserSignupResponseWithCorrelation"},
					},
				},
			},
		},
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"UserSignupWithCorrelation": {
					Name:        "UserSignupWithCorrelation",
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"userId": {Type: "string"},
							"email":  {Type: "string"},
						},
						Required: []string{"userId", "email"},
					},
					CorrelationId: &parser.CorrelationId{
						Location:    "$message.header#/correlationId",
						Description: "Correlation ID for message tracking",
					},
				},
				"UserSignupResponseWithCorrelation": {
					Name:        "UserSignupResponseWithCorrelation",
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"success": {Type: "boolean"},
							"userId":  {Type: "string"},
						},
						Required: []string{"success", "userId"},
					},
					CorrelationId: &parser.CorrelationId{
						Location:    "$message.header#/correlationId",
						Description: "Correlation ID for message tracking",
					},
				},
			},
		},
	}
}

func createProviderSpecWithCorrelationId() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		AsyncAPI: "3.0.0",
		Info: parser.Info{
			Title:   "Provider Service with Correlation ID",
			Version: "1.0.0",
		},
		Servers: map[string]parser.Server{
			"production": {
				Host:     "localhost:5672",
				Protocol: "amqp",
			},
		},
		Channels: map[string]parser.Channel{
			"user/events": {
				Address: "user.events",
				Messages: map[string]parser.MessageRef{
					"userSignup": {
						Ref: "#/components/messages/UserSignupWithCorrelation",
					},
				},
				Servers: []parser.ServerRef{
					{Ref: "#/servers/production"},
				},
			},
		},
		Operations: map[string]parser.Operation{
			"receiveUserSignup": {
				Action:  "receive",
				Channel: parser.ChannelRef{Ref: "#/channels/user~1events"},
				Messages: []parser.MessageRef{
					{Ref: "#/components/messages/UserSignupWithCorrelation"},
				},
				Reply: &parser.Reply{
					Channel: parser.ChannelRef{Ref: "#/channels/user~1events"},
					Messages: []parser.MessageRef{
						{Ref: "#/components/messages/UserSignupResponseWithCorrelation"},
					},
				},
			},
		},
		Components: &parser.Components{
			Messages: map[string]parser.Message{
				"UserSignupWithCorrelation": {
					Name:        "UserSignupWithCorrelation",
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"userId": {Type: "string"},
							"email":  {Type: "string"},
						},
						Required: []string{"userId", "email"},
					},
					CorrelationId: &parser.CorrelationId{
						Location:    "$message.header#/correlationId",
						Description: "Correlation ID for message tracking",
					},
				},
				"UserSignupResponseWithCorrelation": {
					Name:        "UserSignupResponseWithCorrelation",
					ContentType: "application/json",
					Payload: &parser.Schema{
						Type: "object",
						Properties: map[string]parser.Schema{
							"success": {Type: "boolean"},
							"userId":  {Type: "string"},
						},
						Required: []string{"success", "userId"},
					},
					CorrelationId: &parser.CorrelationId{
						Location:    "$message.header#/correlationId",
						Description: "Correlation ID for message tracking",
					},
				},
			},
		},
	}
}

func TestFindMatchingProviderChannel_PubSub(t *testing.T) {
	t.Skip("Будет реализовано в следующей итерации")
}