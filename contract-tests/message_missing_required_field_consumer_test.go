package main

import (
	"testing"
)

// TestValidateConsumerMissingRequiredField тестирует случай, когда у потребителя отсутствует обязательное поле,
// которое требует провайдер. Это критический случай несовместимости сообщений.
// 
// Сценарий:
// - Потребитель имеет 4 строковых поля (field1-field4), все обязательные
// - Провайдер требует 5 полей (field1-field5), все обязательные  
// - Отсутствует field5 у потребителя - это должно генерировать критическую ошибку
func TestValidateConsumerMissingRequiredField(t *testing.T) {
	// Создаем конфигурации для несовместимого потребителя и провайдера
	// Потребитель имеет 4 поля из 5, но провайдер требует все 5 полей (включая field5)
	consumerConfig := &Config{
		ContractTests: ContractTestsConfig{
			Consumer: ConsumerConfig{
				SpecPath: "testdata/structure_validator/consumer_missing_required_field_consumer_asyncapi.yml",
				Name:     "incompatible-consumer",
			},
			Settings: SettingsConfig{
				LogLevel: "debug",
			},
		},
	}

	providerConfig := &Config{
		ContractTests: ContractTestsConfig{
			Provider: ProviderConfig{
				SpecPath: "testdata/structure_validator/consumer_missing_required_field_provider_asyncapi.yml",
				Name:     "incompatible-provider",
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

	// Проверяем, что сообщения НЕ совместимы
	if isCompatible {
		t.Error("Expected messages to be incompatible - consumer missing required field5")
	} else {
		t.Logf("✅ Messages correctly identified as incompatible")
	}

	// Проверяем, что есть ошибки совместимости
	if len(errors) == 0 {
		t.Error("Expected compatibility errors for missing required field")
	} else {
		t.Logf("🔍 Found compatibility errors (%d):", len(errors))
		
		// Проверяем, что есть ошибка о недостающем обязательном поле
		foundMissingFieldError := false
		for _, compatError := range errors {
			t.Logf("  - %s: %s (field: %s, severity: %s)", 
				compatError.Type, compatError.Description, compatError.Field, compatError.Severity)
			
			if compatError.Type == "required_field_missing_in_consumer" && compatError.Field == "payload.field5" {
				foundMissingFieldError = true
			}
		}
		
		if !foundMissingFieldError {
			t.Error("Expected to find required_field_missing_in_consumer error for payload.field5")
		} else {
			t.Logf("✅ Correctly detected missing required field: payload.field5")
		}
	}

	// Дополнительная проверка payload полей
	consumerPayloadFields, _ := messageValidator.ExtractMessagePayloadFields(consumerOutboundMessage, consumerSpec)
	providerPayloadFields, _ := messageValidator.ExtractMessagePayloadFields(providerInboundMessage, providerSpec)

	t.Logf("Consumer payload fields (%d):", len(consumerPayloadFields))
	for _, field := range consumerPayloadFields {
		t.Logf("  - %s: %s (required: %t)", field.Name, field.Type, field.Required)
	}

	t.Logf("Provider payload fields (%d):", len(providerPayloadFields))
	for _, field := range providerPayloadFields {
		t.Logf("  - %s: %s (required: %t)", field.Name, field.Type, field.Required)
	}

	// Проверяем, что у потребителя 4 поля, а у провайдера 5
	if len(consumerPayloadFields) != 4 {
		t.Errorf("Expected consumer to have 4 payload fields, got %d", len(consumerPayloadFields))
	}
	
	if len(providerPayloadFields) != 5 {
		t.Errorf("Expected provider to have 5 payload fields, got %d", len(providerPayloadFields))
	}
}