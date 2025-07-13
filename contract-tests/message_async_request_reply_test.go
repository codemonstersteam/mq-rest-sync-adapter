package main

import (
	"testing"
)

// TestValidateAsyncRequestReplyMessageCompatibility тестирует валидацию структуры сообщений
// для асинхронного взаимодействия по паттерну request-reply с использованием RabbitMQ.
// 
// Сценарий:
// - Потребитель отправляет запрос на обработку события через AMQP
// - Провайдер получает запрос и отправляет ответ с результатами обработки
// - Проверяется совместимость структур сообщений между потребителем и провайдером
// - Используется корреляция через correlationId в headers (AMQP message properties)
func TestValidateAsyncRequestReplyMessageCompatibility(t *testing.T) {
	// Создаем конфигурации для потребителя и провайдера async request-reply
	consumerConfig := &Config{
		ContractTests: ContractTestsConfig{
			Consumer: ConsumerConfig{
				SpecPath: "testdata/structure_validator/request_reply_consumer_asyncapi.yml",
				Name:     "event-processing-consumer",
			},
			Settings: SettingsConfig{
				LogLevel: "debug",
			},
		},
	}

	providerConfig := &Config{
		ContractTests: ContractTestsConfig{
			Provider: ProviderConfig{
				SpecPath: "testdata/structure_validator/request_reply_provider_asyncapi.yml",
				Name:     "event-processing-provider",
			},
			Settings: SettingsConfig{
				LogLevel: "debug",
			},
		},
	}

	// Создаем message validator
	messageValidator := NewMessageValidator(consumerConfig)

	// Загружаем спецификации
	consumerSpec, err := messageValidator.loadConsumerSpec()
	if err != nil {
		t.Fatalf("Failed to load consumer spec: %v", err)
	}

	messageValidator.config = providerConfig
	providerSpec, err := messageValidator.loadProviderSpec()
	if err != nil {
		t.Fatalf("Failed to load provider spec: %v", err)
	}

	// Проверяем совместимость запроса: потребитель отправляет -> провайдер получает
	t.Run("Request Compatibility", func(t *testing.T) {
		// Находим исходящее сообщение потребителя (запрос)
		consumerChannelName := "process_event_request"
		consumerOutboundMessage, err := messageValidator.FindConsumerOutboundMessage(consumerChannelName, consumerSpec)
		if err != nil {
			t.Fatalf("Failed to find consumer outbound message: %v", err)
		}

		// Находим входящее сообщение провайдера (тот же запрос)
		providerChannelName := "process_event_request"
		providerInboundMessage, err := messageValidator.FindProviderInboundMessage(providerChannelName, providerSpec)
		if err != nil {
			t.Fatalf("Failed to find provider inbound message: %v", err)
		}

		// Валидируем совместимость запроса
		isCompatible, errors, err := messageValidator.ValidateMessageCompatibility(
			consumerOutboundMessage, consumerSpec,
			providerInboundMessage, providerSpec,
		)
		if err != nil {
			t.Fatalf("Failed to validate request message compatibility: %v", err)
		}

		// Проверяем результат валидации запроса
		if !isCompatible {
			t.Error("Expected request messages to be compatible")
			for _, compatError := range errors {
				t.Logf("Request compatibility error: %s - %s (field: %s, severity: %s)", 
					compatError.Type, compatError.Description, compatError.Field, compatError.Severity)
			}
		} else {
			t.Logf("✅ Request messages are compatible")
		}

		// Проверяем основные аспекты совместимости запроса
		if consumerOutboundMessage.ContentType != providerInboundMessage.ContentType {
			t.Errorf("Request ContentType mismatch: consumer '%s' vs provider '%s'", 
				consumerOutboundMessage.ContentType, providerInboundMessage.ContentType)
		}

		t.Logf("Consumer outbound request: %s (%s)", consumerOutboundMessage.Name, consumerOutboundMessage.ContentType)
		t.Logf("Provider inbound request: %s (%s)", providerInboundMessage.Name, providerInboundMessage.ContentType)
	})

	// Проверяем совместимость ответа: провайдер отправляет -> потребитель получает
	t.Run("Response Compatibility", func(t *testing.T) {
		// Находим исходящее сообщение провайдера (ответ)
		providerChannelName := "process_event_response"
		providerOutboundMessage, err := messageValidator.FindProviderOutboundMessage(providerChannelName, providerSpec)
		if err != nil {
			t.Fatalf("Failed to find provider outbound message: %v", err)
		}

		// Находим входящее сообщение потребителя (тот же ответ)
		consumerChannelName := "process_event_response"
		consumerInboundMessage, err := messageValidator.FindConsumerInboundMessage(consumerChannelName, consumerSpec)
		if err != nil {
			t.Fatalf("Failed to find consumer inbound message: %v", err)
		}

		// Валидируем совместимость ответа
		isCompatible, errors, err := messageValidator.ValidateMessageCompatibility(
			providerOutboundMessage, providerSpec,
			consumerInboundMessage, consumerSpec,
		)
		if err != nil {
			t.Fatalf("Failed to validate response message compatibility: %v", err)
		}

		// Проверяем результат валидации ответа
		if !isCompatible {
			t.Error("Expected response messages to be compatible")
			for _, compatError := range errors {
				t.Logf("Response compatibility error: %s - %s (field: %s, severity: %s)", 
					compatError.Type, compatError.Description, compatError.Field, compatError.Severity)
			}
		} else {
			t.Logf("✅ Response messages are compatible")
		}

		// Проверяем основные аспекты совместимости ответа
		if providerOutboundMessage.ContentType != consumerInboundMessage.ContentType {
			t.Errorf("Response ContentType mismatch: provider '%s' vs consumer '%s'", 
				providerOutboundMessage.ContentType, consumerInboundMessage.ContentType)
		}

		t.Logf("Provider outbound response: %s (%s)", providerOutboundMessage.Name, providerOutboundMessage.ContentType)
		t.Logf("Consumer inbound response: %s (%s)", consumerInboundMessage.Name, consumerInboundMessage.ContentType)
	})

	// Проверяем структуру payload полей для запроса
	t.Run("Request Payload Fields Validation", func(t *testing.T) {
		// Получаем сообщения запроса
		consumerOutboundMessage, _ := messageValidator.FindConsumerOutboundMessage("process_event_request", consumerSpec)
		messageValidator.config = providerConfig
		providerInboundMessage, _ := messageValidator.FindProviderInboundMessage("process_event_request", providerSpec)

		// Извлекаем поля payload запроса
		consumerPayloadFields, err := messageValidator.ExtractMessagePayloadFields(consumerOutboundMessage, consumerSpec)
		if err != nil {
			t.Fatalf("Failed to extract consumer request payload fields: %v", err)
		}

		providerPayloadFields, err := messageValidator.ExtractMessagePayloadFields(providerInboundMessage, providerSpec)
		if err != nil {
			t.Fatalf("Failed to extract provider request payload fields: %v", err)
		}

		// Проверяем структуру запроса: 3 строковых поля, 2 обязательных
		t.Logf("Consumer request payload fields (%d):", len(consumerPayloadFields))
		expectedFields := map[string]bool{
			"eventType":   false, // опциональное
			"eventSource": true,  // обязательное
			"eventData":   true,  // обязательное
		}

		for _, field := range consumerPayloadFields {
			t.Logf("  - %s: %s (required: %t)", field.Name, field.Type, field.Required)
			
			expectedRequired, exists := expectedFields[field.Name]
			if !exists {
				t.Errorf("Unexpected field in consumer request: %s", field.Name)
			} else if field.Required != expectedRequired {
				t.Errorf("Field %s required mismatch: expected %t, got %t", field.Name, expectedRequired, field.Required)
			}
			if field.Type != "string" {
				t.Errorf("Field %s type mismatch: expected string, got %s", field.Name, field.Type)
			}
		}

		// Проверяем количество полей (должно быть 3)
		if len(consumerPayloadFields) != 3 {
			t.Errorf("Expected 3 payload fields in consumer request, got %d", len(consumerPayloadFields))
		}

		// Проверяем совместимость полей между потребителем и провайдером
		fieldCompatibility, fieldErrors, err := messageValidator.ValidatePayloadFieldsCompatibility(
			consumerPayloadFields, providerPayloadFields,
		)
		if err != nil {
			t.Fatalf("Failed to validate request payload fields compatibility: %v", err)
		}

		if !fieldCompatibility {
			t.Error("Expected request payload fields to be compatible")
			for _, fieldError := range fieldErrors {
				t.Logf("Request field compatibility error: %s - %s (field: %s)", 
					fieldError.Type, fieldError.Description, fieldError.Field)
			}
		} else {
			t.Logf("✅ Request payload fields are compatible")
		}
	})

	// Проверяем корреляционные headers (AMQP message properties)
	t.Run("Correlation Headers Validation", func(t *testing.T) {
		// Получаем сообщения запроса
		consumerOutboundMessage, _ := messageValidator.FindConsumerOutboundMessage("process_event_request", consumerSpec)
		messageValidator.config = providerConfig
		providerInboundMessage, _ := messageValidator.FindProviderInboundMessage("process_event_request", providerSpec)

		// Извлекаем поля headers
		consumerHeaderFields, err := messageValidator.ExtractMessageHeaderFields(consumerOutboundMessage, consumerSpec)
		if err != nil {
			t.Fatalf("Failed to extract consumer request header fields: %v", err)
		}

		providerHeaderFields, err := messageValidator.ExtractMessageHeaderFields(providerInboundMessage, providerSpec)
		if err != nil {
			t.Fatalf("Failed to extract provider request header fields: %v", err)
		}

		t.Logf("Consumer request header fields (%d):", len(consumerHeaderFields))
		for _, field := range consumerHeaderFields {
			t.Logf("  - %s: %s (required: %t)", field.Name, field.Type, field.Required)
		}

		// Проверяем наличие обязательных AMQP корреляционных полей
		requiredHeaders := []string{"correlationId", "replyTo"}
		for _, requiredHeader := range requiredHeaders {
			found := false
			for _, field := range consumerHeaderFields {
				if field.Name == requiredHeader && field.Required {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Required correlation header '%s' not found or not marked as required", requiredHeader)
			}
		}

		// Проверяем совместимость headers между потребителем и провайдером
		headerCompatibility, headerErrors, err := messageValidator.ValidateHeaderFieldsCompatibility(
			consumerHeaderFields, providerHeaderFields,
		)
		if err != nil {
			t.Fatalf("Failed to validate request header fields compatibility: %v", err)
		}

		if !headerCompatibility {
			t.Error("Expected request header fields to be compatible")
			for _, headerError := range headerErrors {
				t.Logf("Request header compatibility error: %s - %s (field: %s)", 
					headerError.Type, headerError.Description, headerError.Field)
			}
		} else {
			t.Logf("✅ Request header fields are compatible")
		}
	})
}