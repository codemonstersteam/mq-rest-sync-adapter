package main

import (
	"testing"
)

func TestValidateMessageStructure(t *testing.T) {
	// Создаем конфигурацию с валидными спецификациями
	config := &Config{
		ContractTests: ContractTestsConfig{
			Consumer: ConsumerConfig{
				SpecPath: "testdata/structure_validator/valid_consumer_asyncapi.yml",
				Name:     "mq-rest-sync-adapter",
				Channels: []string{"restGetBalanceRequest"},
			},
			Provider: ProviderConfig{
				SpecURL: "", // Используем локальный файл для тестирования
				SpecPath: "testdata/structure_validator/valid_provider_asyncapi.yml",
				Name:    "wallet-balance-service",
			},
			Settings: SettingsConfig{
				LogLevel: "info",
			},
		},
	}

	// Сначала валидируем каналы через channel validator
	channelChecker := NewContractChecker(config)
	channelResult, err := channelChecker.CheckCompatibility()
	if err != nil {
		t.Fatalf("Failed to check channel compatibility: %v", err)
	}

	// Проверяем, что каналы совместимы
	if !channelResult.IsCompatible {
		t.Error("Expected channels to be compatible for message validation test")
		for _, change := range channelResult.BreakingChanges {
			t.Logf("Channel breaking change: %s - %s", change.Type, change.Description)
		}
		return
	}

	// Извлекаем информацию о валидных каналах из результата channel checker
	validatedChannels := extractValidatedChannels(channelResult, config)

	// Создаем message validator
	messageValidator := NewMessageValidator(config)

	// Выполняем валидацию структуры сообщений
	result, err := messageValidator.ValidateMessages(validatedChannels)
	if err != nil {
		t.Fatalf("Failed to validate messages: %v", err)
	}

	// Проверяем базовую информацию
	if result.ConsumerName != config.ContractTests.Consumer.Name {
		t.Errorf("Expected consumer name '%s', got '%s'", config.ContractTests.Consumer.Name, result.ConsumerName)
	}

	if result.ProviderName != config.ContractTests.Provider.Name {
		t.Errorf("Expected provider name '%s', got '%s'", config.ContractTests.Provider.Name, result.ProviderName)
	}

	// Проверяем, что есть валидированные каналы
	if len(result.ValidatedChannels) == 0 {
		t.Error("Expected at least one validated channel")
	}

	// Проверяем результат валидации сообщений
	// Для HTTP каналов может быть нормально иметь предупреждения о недостающих направлениях
	hasStructuralErrors := false
	for _, msgError := range result.MessageErrors {
		if msgError.Type == "message_structure_incompatible" || msgError.Type == "payload_structure_mismatch" {
			hasStructuralErrors = true
		}
		t.Logf("Message validation info: %s - %s (channel: %s, message: %s, severity: %s)", msgError.Type, msgError.Description, msgError.Channel, msgError.Message, msgError.Severity)
	}

	if hasStructuralErrors {
		t.Error("Expected no structural incompatibilities for compatible message structures")
	}

	// Проверяем, что найден канал restGetBalanceRequest
	foundChannel := false
	for _, channel := range result.ValidatedChannels {
		if channel.ConsumerChannelName == "restGetBalanceRequest" {
			foundChannel = true
			
			// Проверяем детали канала
			if channel.Address != "/api/client/{Api-clientId}/wallet/balance" &&
			   channel.Address != "/api/client/{clientId}/wallet/balance" {
				t.Errorf("Expected address '/api/client/{clientId}/wallet/balance' or similar, got '%s'", channel.Address)
			}
			
			if channel.Protocol != "http" {
				t.Errorf("Expected protocol 'http', got '%s'", channel.Protocol)
			}
			
			break
		}
	}

	if !foundChannel {
		t.Error("Expected to find validated channel 'restGetBalanceRequest'")
	}
}

// extractValidatedChannels извлекает информацию о валидных каналах из результата channel checker
func extractValidatedChannels(channelResult *CompatibilityResult, config *Config) []ValidatedChannel {
	var validatedChannels []ValidatedChannel

	// В реальной реализации это будет делать channel checker,
	// здесь мы симулируем для тестирования
	for _, channelName := range config.ContractTests.Consumer.Channels {
		// Для канала restGetBalanceRequest мы знаем соответствие
		if channelName == "restGetBalanceRequest" {
			validatedChannels = append(validatedChannels, ValidatedChannel{
				ConsumerChannelName: "restGetBalanceRequest",
				ProviderChannelName: "walletBalanceRequest", // Соответствующий канал в провайдере
				Address:             "/api/client/{clientId}/wallet/balance",
				Protocol:            "http",
			})
		}
	}

	return validatedChannels
}


func TestFindMessageByDirection(t *testing.T) {
	// Тест для проверки определения направления сообщений через operations
	config := &Config{
		ContractTests: ContractTestsConfig{
			Consumer: ConsumerConfig{
				SpecPath: "testdata/structure_validator/valid_consumer_asyncapi.yml",
				Name:     "test-consumer",
			},
			Provider: ProviderConfig{
				SpecPath: "testdata/structure_validator/valid_provider_asyncapi.yml",
				Name:     "test-provider",
			},
			Settings: SettingsConfig{
				LogLevel: "info",
			},
		},
	}

	messageValidator := NewMessageValidator(config)

	// Загружаем спецификации
	consumerSpec, err := messageValidator.loadConsumerSpec()
	if err != nil {
		t.Fatalf("Failed to load consumer spec: %v", err)
	}

	providerSpec, err := messageValidator.loadProviderSpec()
	if err != nil {
		t.Fatalf("Failed to load provider spec: %v", err)
	}

	// Тестируем определение исходящих сообщений потребителя для канала restGetBalanceRequest
	consumerOutbound := messageValidator.findOutboundMessages("restGetBalanceRequest", consumerSpec)
	if len(consumerOutbound) == 0 {
		t.Error("Expected to find outbound messages for consumer channel restGetBalanceRequest")
	}

	// Тестируем определение входящих сообщений провайдера для канала walletBalanceRequest
	providerInbound := messageValidator.findInboundMessages("walletBalanceRequest", providerSpec)
	if len(providerInbound) == 0 {
		t.Error("Expected to find inbound messages for provider channel walletBalanceRequest")
	}

	t.Logf("Consumer outbound messages: %v", consumerOutbound)
	t.Logf("Provider inbound messages: %v", providerInbound)
}