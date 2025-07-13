package main

import (
	"testing"
)

// TestValidateConsumerToProviderMessageCompatibility тестирует совместимость исходящего сообщения потребителя с входящим сообщением провайдера
func TestValidateConsumerToProviderMessageCompatibility(t *testing.T) {
	// Создаем конфигурации для потребителя и провайдера
	consumerConfig := &Config{
		ContractTests: ContractTestsConfig{
			Consumer: ConsumerConfig{
				SpecPath: "testdata/structure_validator/valid_consumer_asyncapi.yml",
				Name:     "mq-rest-sync-adapter",
			},
			Settings: SettingsConfig{
				LogLevel: "info",
			},
		},
	}

	providerConfig := &Config{
		ContractTests: ContractTestsConfig{
			Provider: ProviderConfig{
				SpecPath: "testdata/structure_validator/valid_provider_asyncapi.yml",
				Name:     "wallet-balance-service",
			},
			Settings: SettingsConfig{
				LogLevel: "info",
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

	// Обновляем конфигурацию для загрузки провайдера
	messageValidator.config = providerConfig
	providerSpec, err := messageValidator.loadProviderSpec()
	if err != nil {
		t.Fatalf("Failed to load provider spec: %v", err)
	}

	// Находим исходящее сообщение потребителя
	consumerChannelName := "restGetBalanceRequest"
	consumerOutboundMessage, err := messageValidator.FindConsumerOutboundMessage(consumerChannelName, consumerSpec)
	if err != nil {
		t.Fatalf("Failed to find consumer outbound message: %v", err)
	}

	// Находим входящее сообщение провайдера
	providerChannelName := "walletBalanceRequest"
	providerInboundMessage, err := messageValidator.FindProviderInboundMessage(providerChannelName, providerSpec)
	if err != nil {
		t.Fatalf("Failed to find provider inbound message: %v", err)
	}

	// Валидируем совместимость сообщений
	isCompatible, errors, err := messageValidator.ValidateMessageCompatibility(
		consumerOutboundMessage, consumerSpec,
		providerInboundMessage, providerSpec,
	)
	if err != nil {
		t.Fatalf("Failed to validate message compatibility: %v", err)
	}

	// Проверяем результат валидации
	if !isCompatible {
		t.Error("Expected messages to be compatible")
		for _, compatError := range errors {
			t.Logf("Compatibility error: %s - %s (field: %s, severity: %s)", 
				compatError.Type, compatError.Description, compatError.Field, compatError.Severity)
		}
	} else {
		t.Logf("✅ Messages are compatible")
		if len(errors) > 0 {
			t.Logf("Non-critical compatibility issues:")
			for _, compatError := range errors {
				t.Logf("  - %s: %s", compatError.Type, compatError.Description)
			}
		}
	}

	// Проверяем основные аспекты совместимости
	if consumerOutboundMessage.ContentType != providerInboundMessage.ContentType {
		t.Errorf("ContentType mismatch: consumer '%s' vs provider '%s'", 
			consumerOutboundMessage.ContentType, providerInboundMessage.ContentType)
	}

	t.Logf("Consumer outbound: %s (%s)", consumerOutboundMessage.Name, consumerOutboundMessage.ContentType)
	t.Logf("Provider inbound: %s (%s)", providerInboundMessage.Name, providerInboundMessage.ContentType)
}

// TestValidatePayloadFieldsCompatibility тестирует совместимость полей payload между сообщениями
func TestValidatePayloadFieldsCompatibility(t *testing.T) {
	// Подготовка данных аналогично предыдущему тесту
	consumerConfig := &Config{
		ContractTests: ContractTestsConfig{
			Consumer: ConsumerConfig{
				SpecPath: "testdata/structure_validator/valid_consumer_asyncapi.yml",
				Name:     "mq-rest-sync-adapter",
			},
			Settings: SettingsConfig{
				LogLevel: "info",
			},
		},
	}

	providerConfig := &Config{
		ContractTests: ContractTestsConfig{
			Provider: ProviderConfig{
				SpecPath: "testdata/structure_validator/valid_provider_asyncapi.yml",
				Name:     "wallet-balance-service",
			},
			Settings: SettingsConfig{
				LogLevel: "info",
			},
		},
	}

	messageValidator := NewMessageValidator(consumerConfig)

	// Загружаем спецификации и находим сообщения
	consumerSpec, _ := messageValidator.loadConsumerSpec()
	messageValidator.config = providerConfig
	providerSpec, _ := messageValidator.loadProviderSpec()

	consumerOutboundMessage, _ := messageValidator.FindConsumerOutboundMessage("restGetBalanceRequest", consumerSpec)
	providerInboundMessage, _ := messageValidator.FindProviderInboundMessage("walletBalanceRequest", providerSpec)

	// Извлекаем поля payload
	consumerPayloadFields, err := messageValidator.ExtractMessagePayloadFields(consumerOutboundMessage, consumerSpec)
	if err != nil {
		t.Fatalf("Failed to extract consumer payload fields: %v", err)
	}

	providerPayloadFields, err := messageValidator.ExtractMessagePayloadFields(providerInboundMessage, providerSpec)
	if err != nil {
		t.Fatalf("Failed to extract provider payload fields: %v", err)
	}

	// Валидируем совместимость полей payload
	fieldCompatibility, fieldErrors, err := messageValidator.ValidatePayloadFieldsCompatibility(
		consumerPayloadFields, providerPayloadFields,
	)
	if err != nil {
		t.Fatalf("Failed to validate payload fields compatibility: %v", err)
	}

	// Проверяем результат
	if !fieldCompatibility {
		t.Error("Expected payload fields to be compatible")
		for _, fieldError := range fieldErrors {
			t.Logf("Field compatibility error: %s - %s (field: %s)", 
				fieldError.Type, fieldError.Description, fieldError.Field)
		}
	} else {
		t.Logf("✅ Payload fields are compatible")
	}

	// Логируем детали полей для проверки
	t.Logf("Consumer payload fields:")
	for _, field := range consumerPayloadFields {
		t.Logf("  - %s: %s (required: %t)", field.Name, field.Type, field.Required)
	}

	t.Logf("Provider payload fields:")
	for _, field := range providerPayloadFields {
		t.Logf("  - %s: %s (required: %t)", field.Name, field.Type, field.Required)
	}
}

// TestValidateIncompatibleMessages тестирует случай несовместимых сообщений
func TestValidateIncompatibleMessages(t *testing.T) {
	// Этот тест перенесен в отдельный модуль message_missing_required_field_consumer_test.go
	// для лучшей организации тестовой базы - один тест кейс на модуль
	t.Skip("Test moved to message_missing_required_field_consumer_test.go")
}

