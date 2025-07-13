package main

import (
	"strings"
	"testing"
)

func TestCheckChannels(t *testing.T) {
	// Загружаем тестовую конфигурацию
	config, err := LoadConfig("testdata/channel_validator/testdata-contract-tests.yaml")
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	// Создаем проверщик
	checker := NewContractChecker(config)

	// Выполняем проверку совместимости
	result, err := checker.CheckCompatibility()
	if err != nil {
		t.Fatalf("Failed to check compatibility: %v", err)
	}

	// Проверяем, что результат успешный
	if !result.IsCompatible {
		t.Errorf("Expected compatibility to be true, but got false")

		// Выводим детали для отладки
		for _, change := range result.BreakingChanges {
			t.Logf("Breaking change: %s - %s", change.Type, change.Description)
		}
	}

	// Проверяем, что нет критических изменений
	if len(result.BreakingChanges) > 0 {
		t.Errorf("Expected no breaking changes, but got %d", len(result.BreakingChanges))
	}

	// Проверяем базовую информацию
	if result.ConsumerName != config.ContractTests.Consumer.Name {
		t.Errorf("Expected consumer name '%s', got '%s'", config.ContractTests.Consumer.Name, result.ConsumerName)
	}

	if result.ProviderName != config.ContractTests.Provider.Name {
		t.Errorf("Expected provider name '%s', got '%s'", config.ContractTests.Provider.Name, result.ProviderName)
	}
}

func TestFindChannelByAddress(t *testing.T) {
	config := &Config{
		ContractTests: ContractTestsConfig{
			Settings: SettingsConfig{
				LogLevel: "info",
			},
		},
	}
	checker := NewContractChecker(config)

	// Создаем тестовую спецификацию провайдера
	provider := &AsyncAPISpec{
		Channels: map[string]AsyncAPIChannel{
			"channel1": {
				Address: "/api/client/{clientId}/wallet/balance",
			},
			"channel2": {
				Address: "IN.QUEUE.PS",
			},
		},
	}

	// Тест поиска REST endpoint
	channelName := checker.findChannelByAddress(provider, "/api/client/{clientId}/wallet/balance")
	if channelName != "channel1" {
		t.Errorf("Expected to find 'channel1', got '%s'", channelName)
	}

	// Тест поиска MQ канала
	channelName = checker.findChannelByAddress(provider, "IN.QUEUE.PS")
	if channelName != "channel2" {
		t.Errorf("Expected to find 'channel2', got '%s'", channelName)
	}

	// Тест поиска несуществующего канала
	channelName = checker.findChannelByAddress(provider, "/api/nonexistent")
	if channelName != "" {
		t.Errorf("Expected empty string for non-existent channel, got '%s'", channelName)
	}
}

func TestNormalizeAddress(t *testing.T) {
	config := &Config{
		ContractTests: ContractTestsConfig{
			Settings: SettingsConfig{
				LogLevel: "info",
			},
		},
	}
	checker := NewContractChecker(config)

	tests := []struct {
		name     string
		address  string
		expected string
	}{
		{
			name:     "REST endpoint",
			address:  "/api/client/{clientId}/wallet/balance",
			expected: "/api/client/{clientId}/wallet/balance",
		},
		{
			name:     "MQ channel",
			address:  "IN.QUEUE.PS",
			expected: "IN.QUEUE.PS",
		},
		{
			name:     "Empty address",
			address:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.normalizeAddress(tt.address)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestLoadSpecs(t *testing.T) {
	config, err := LoadConfig("testdata/channel_validator/testdata-contract-tests.yaml")
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	checker := NewContractChecker(config)

	// Тест загрузки спецификации потребителя
	consumerSpec, err := checker.loadConsumerSpec()
	if err != nil {
		t.Fatalf("Failed to load consumer spec: %v", err)
	}

	if consumerSpec.Info.Title == "" {
		t.Error("Consumer spec title should not be empty")
	}

	// Тест загрузки спецификации провайдера
	providerSpec, err := checker.loadProviderSpec()
	if err != nil {
		t.Fatalf("Failed to load provider spec: %v", err)
	}

	if providerSpec.Info.Title == "" {
		t.Error("Provider spec title should not be empty")
	}

	// Проверяем, что каналы загружены
	if len(consumerSpec.Channels) == 0 {
		t.Error("Consumer spec should have channels")
	}

	if len(providerSpec.Channels) == 0 {
		t.Error("Provider spec should have channels")
	}
}

func TestCheckChannelsIncompatible(t *testing.T) {
	// Создаем конфигурацию с провайдером, который не поддерживает нужный канал
	config := &Config{
		ContractTests: ContractTestsConfig{
			Consumer: ConsumerConfig{
				SpecPath: "testdata/channel_validator/testdata_consumer_asyncapi.yml",
				Name:     "test-consumer",
				Channels: []string{"restGetBalanceRequest"},
			},
			Provider: ProviderConfig{
				SpecURL: "testdata/channel_validator/testdata_provider_fails_channel.yml",
				Name:    "test-provider",
			},
			Settings: SettingsConfig{
				LogLevel: "info", // Отключаем debug логи для теста
			},
		},
	}

	checker := NewContractChecker(config)

	// Выполняем проверку совместимости
	result, err := checker.CheckCompatibility()
	if err != nil {
		t.Fatalf("Failed to check compatibility: %v", err)
	}

	// Проверяем, что результат НЕ совместим
	if result.IsCompatible {
		t.Error("Expected compatibility to be false, but got true")
	}

	// Проверяем, что есть критические изменения
	if len(result.BreakingChanges) == 0 {
		t.Error("Expected breaking changes, but got none")
	}

	// Проверяем, что есть изменение типа channel_address_not_supported
	foundAddressNotSupported := false
	for _, change := range result.BreakingChanges {
		if change.Type == "channel_address_not_supported" {
			foundAddressNotSupported = true

			// Проверяем детали изменения
			if change.Field != "channels.restGetBalanceRequest.address" {
				t.Errorf("Expected field 'channels.restGetBalanceRequest.address', got '%s'", change.Field)
			}
			if change.OldValue != "/api/client/{Api-clientId}/wallet/balance" {
				t.Errorf("Expected old value '/api/client/{Api-clientId}/wallet/balance', got '%s'", change.OldValue)
			}
			if change.NewValue != "not_supported" {
				t.Errorf("Expected new value 'not_supported', got '%s'", change.NewValue)
			}
			if change.Severity != "high" {
				t.Errorf("Expected severity 'high', got '%s'", change.Severity)
			}
			break
		}
	}

	if !foundAddressNotSupported {
		t.Error("Expected to find breaking change of type 'channel_address_not_supported'")
	}
}

func TestCheckChannelsIncompatibleWithConfigFile(t *testing.T) {
	// Загружаем конфигурацию с каналом, который не поддерживает провайдер
	config, err := LoadConfig("testdata/channel_validator/testdata-contract-tests-fail-channel-not-supported.yaml")
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	checker := NewContractChecker(config)

	// Выполняем проверку совместимости
	result, err := checker.CheckCompatibility()
	if err != nil {
		t.Fatalf("Failed to check compatibility: %v", err)
	}

	// Проверяем, что результат НЕ совместим
	if result.IsCompatible {
		t.Error("Expected compatibility to be false, but got true")

		// Выводим детали для отладки
		for _, change := range result.BreakingChanges {
			t.Logf("Breaking change: %s - %s", change.Type, change.Description)
		}
	}

	// Проверяем, что есть критические изменения
	if len(result.BreakingChanges) == 0 {
		t.Error("Expected breaking changes, but got none")
	}

	// Проверяем, что есть изменение типа channel_address_not_supported для канала sendRestResponse
	foundMQChannelNotSupported := false
	for _, change := range result.BreakingChanges {
		if change.Type == "channel_address_not_supported" && change.Field == "channels.sendRestResponse.address" {
			foundMQChannelNotSupported = true

			// Проверяем детали изменения
			if change.OldValue != "OUT.QUEUE.PS" {
				t.Errorf("Expected old value 'OUT.QUEUE.PS', got '%s'", change.OldValue)
			}
			// Проверяем, что описание содержит протокол
			if !strings.Contains(change.Description, "протокол") {
				t.Errorf("Expected description to contain protocol information, got: %s", change.Description)
			}
			if change.NewValue != "not_supported" {
				t.Errorf("Expected new value 'not_supported', got '%s'", change.NewValue)
			}
			if change.Severity != "high" {
				t.Errorf("Expected severity 'high', got '%s'", change.Severity)
			}
			break
		}
	}

	if !foundMQChannelNotSupported {
		t.Error("Expected to find breaking change of type 'channel_address_not_supported' for sendRestResponse channel")

		// Выводим все breaking changes для отладки
		for _, change := range result.BreakingChanges {
			t.Logf("Breaking change found: Type=%s, Field=%s, OldValue=%s", change.Type, change.Field, change.OldValue)
		}
	}
}

func TestExtractServerName(t *testing.T) {
	config := &Config{
		ContractTests: ContractTestsConfig{
			Settings: SettingsConfig{
				LogLevel: "info",
			},
		},
	}
	checker := NewContractChecker(config)

	tests := []struct {
		name     string
		ref      string
		expected string
	}{
		{
			name:     "Valid server reference",
			ref:      "#/servers/broker",
			expected: "broker",
		},
		{
			name:     "Valid REST server reference",
			ref:      "#/servers/restServer",
			expected: "restServer",
		},
		{
			name:     "Invalid reference",
			ref:      "invalid/ref",
			expected: "",
		},
		{
			name:     "Empty reference",
			ref:      "",
			expected: "",
		},
		{
			name:     "Missing server name",
			ref:      "#/servers/",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.extractServerName(tt.ref)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestChannelMissingServerReference(t *testing.T) {
	// Загружаем конфигурацию с невалидной спецификацией потребителя (без ссылки на сервер)
	config, err := LoadConfig("testdata/channel_validator/testdata_config_missing_server.yaml")
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	checker := NewContractChecker(config)

	// Выполняем проверку совместимости
	result, err := checker.CheckCompatibility()
	if err != nil {
		t.Fatalf("Failed to check compatibility: %v", err)
	}

	// Проверяем, что результат НЕ совместим
	if result.IsCompatible {
		t.Error("Expected compatibility to be false, but got true")

		// Выводим детали для отладки
		for _, change := range result.BreakingChanges {
			t.Logf("Breaking change: %s - %s", change.Type, change.Description)
		}
	}

	// Проверяем, что есть критические изменения
	if len(result.BreakingChanges) == 0 {
		t.Error("Expected breaking changes, but got none")
	}

	// Проверяем, что есть изменение типа channel_missing_server_reference
	foundMissingServerRef := false
	for _, change := range result.BreakingChanges {
		if change.Type == "channel_missing_server_reference" {
			foundMissingServerRef = true

			// Проверяем детали изменения
			if change.Field != "channels.restGetBalanceRequest.servers" {
				t.Errorf("Expected field 'channels.restGetBalanceRequest.servers', got '%s'", change.Field)
			}
			if change.OldValue != "[]" {
				t.Errorf("Expected old value '[]', got '%s'", change.OldValue)
			}
			if change.NewValue != "required" {
				t.Errorf("Expected new value 'required', got '%s'", change.NewValue)
			}
			if change.Severity != "high" {
				t.Errorf("Expected severity 'high', got '%s'", change.Severity)
			}
			// Проверяем, что описание содержит информацию о необходимости ссылки на сервер
			if !strings.Contains(change.Description, "ссылки на сервер") {
				t.Errorf("Expected description to contain 'ссылки на сервер', got: %s", change.Description)
			}
			break
		}
	}

	if !foundMissingServerRef {
		t.Error("Expected to find breaking change of type 'channel_missing_server_reference'")

		// Выводим все breaking changes для отладки
		for _, change := range result.BreakingChanges {
			t.Logf("Breaking change found: Type=%s, Field=%s, Description=%s", change.Type, change.Field, change.Description)
		}
	}
}

func TestProviderChannelMissingServerReference(t *testing.T) {
	// Загружаем конфигурацию где у потребителя есть ссылка на сервер, а у провайдера нет
	config, err := LoadConfig("testdata/channel_validator/testdata_config_provider_missing_server.yaml")
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	checker := NewContractChecker(config)

	// Выполняем проверку совместимости
	result, err := checker.CheckCompatibility()
	if err != nil {
		t.Fatalf("Failed to check compatibility: %v", err)
	}

	// Проверяем, что результат НЕ совместим
	if result.IsCompatible {
		t.Error("Expected compatibility to be false, but got true")

		// Выводим детали для отладки
		for _, change := range result.BreakingChanges {
			t.Logf("Breaking change: %s - %s", change.Type, change.Description)
		}
	}

	// Проверяем, что есть критические изменения
	if len(result.BreakingChanges) == 0 {
		t.Error("Expected breaking changes, but got none")
	}

	// Проверяем, что есть изменение типа provider_channel_missing_server_reference
	foundProviderMissingServerRef := false
	for _, change := range result.BreakingChanges {
		if change.Type == "provider_channel_missing_server_reference" {
			foundProviderMissingServerRef = true

			// Проверяем детали изменения
			if change.Field != "provider.channels.walletBalanceRequest.servers" {
				t.Errorf("Expected field 'provider.channels.walletBalanceRequest.servers', got '%s'", change.Field)
			}
			if change.OldValue != "[]" {
				t.Errorf("Expected old value '[]', got '%s'", change.OldValue)
			}
			if change.NewValue != "required" {
				t.Errorf("Expected new value 'required', got '%s'", change.NewValue)
			}
			if change.Severity != "high" {
				t.Errorf("Expected severity 'high', got '%s'", change.Severity)
			}
			// Проверяем, что описание содержит информацию о провайдере и необходимости ссылки на сервер
			if !strings.Contains(change.Description, "провайдера") || !strings.Contains(change.Description, "ссылки на сервер") {
				t.Errorf("Expected description to contain 'провайдера' and 'ссылки на сервер', got: %s", change.Description)
			}
			break
		}
	}

	if !foundProviderMissingServerRef {
		t.Error("Expected to find breaking change of type 'provider_channel_missing_server_reference'")

		// Выводим все breaking changes для отладки
		for _, change := range result.BreakingChanges {
			t.Logf("Breaking change found: Type=%s, Field=%s, Description=%s", change.Type, change.Field, change.Description)
		}
	}
}

func TestProtocolMismatch(t *testing.T) {
	// Загружаем конфигурацию где у потребителя HTTP протокол, а у провайдера HTTPS
	config, err := LoadConfig("testdata/channel_validator/testdata_config_protocol_mismatch.yaml")
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	checker := NewContractChecker(config)

	// Выполняем проверку совместимости
	result, err := checker.CheckCompatibility()
	if err != nil {
		t.Fatalf("Failed to check compatibility: %v", err)
	}

	// Проверяем, что результат НЕ совместим
	if result.IsCompatible {
		t.Error("Expected compatibility to be false, but got true")

		// Выводим детали для отладки
		for _, change := range result.BreakingChanges {
			t.Logf("Breaking change: %s - %s", change.Type, change.Description)
		}
	}

	// Проверяем, что есть критические изменения
	if len(result.BreakingChanges) == 0 {
		t.Error("Expected breaking changes, but got none")
	}

	// Проверяем, что есть изменение типа protocol_mismatch
	foundProtocolMismatch := false
	for _, change := range result.BreakingChanges {
		if change.Type == "protocol_mismatch" {
			foundProtocolMismatch = true

			// Проверяем детали изменения
			if change.Field != "channels.restGetBalanceRequest.protocol" {
				t.Errorf("Expected field 'channels.restGetBalanceRequest.protocol', got '%s'", change.Field)
			}
			if change.OldValue != "http" {
				t.Errorf("Expected old value 'http', got '%s'", change.OldValue)
			}
			if change.NewValue != "https" {
				t.Errorf("Expected new value 'https', got '%s'", change.NewValue)
			}
			if change.Severity != "high" {
				t.Errorf("Expected severity 'high', got '%s'", change.Severity)
			}
			// Проверяем, что описание содержит информацию о несовместимости протоколов
			if !strings.Contains(change.Description, "несовместимы") {
				t.Errorf("Expected description to contain 'несовместимы', got: %s", change.Description)
			}
			break
		}
	}

	if !foundProtocolMismatch {
		t.Error("Expected to find breaking change of type 'protocol_mismatch'")

		// Выводим все breaking changes для отладки
		for _, change := range result.BreakingChanges {
			t.Logf("Breaking change found: Type=%s, Field=%s, Description=%s", change.Type, change.Field, change.Description)
		}
	}
}

func TestAreProtocolsCompatible(t *testing.T) {
	config := &Config{
		ContractTests: ContractTestsConfig{
			Settings: SettingsConfig{
				LogLevel: "info",
			},
		},
	}
	checker := NewContractChecker(config)

	tests := []struct {
		name               string
		consumerProtocol   string
		providerProtocol   string
		expectedCompatible bool
	}{
		{
			name:               "Same protocols - http",
			consumerProtocol:   "http",
			providerProtocol:   "http",
			expectedCompatible: true,
		},
		{
			name:               "HTTP consumer, HTTPS provider",
			consumerProtocol:   "http",
			providerProtocol:   "https",
			expectedCompatible: false,
		},
		{
			name:               "HTTPS consumer, HTTP provider",
			consumerProtocol:   "https",
			providerProtocol:   "http",
			expectedCompatible: false,
		},
		{
			name:               "Different protocols - ibmmq vs https",
			consumerProtocol:   "ibmmq",
			providerProtocol:   "https",
			expectedCompatible: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.areProtocolsCompatible(tt.consumerProtocol, tt.providerProtocol)
			if result != tt.expectedCompatible {
				t.Errorf("Expected compatibility %v, got %v", tt.expectedCompatible, result)
			}
		})
	}
}

// TestIsProtocolKnown тестирует проверку протоколов в словаре
func TestIsProtocolKnown(t *testing.T) {
	config := &Config{
		ContractTests: ContractTestsConfig{
			Consumer: ConsumerConfig{
				Name:     "test-consumer",
				SpecPath: "testdata/channel_validator/testdata_consumer_asyncapi.yml",
				Channels: []string{"restGetBalanceRequest"},
			},
			Provider: ProviderConfig{
				Name:    "test-provider",
				SpecURL: "testdata/channel_validator/testdata_provider_success.yml",
			},
			Settings: SettingsConfig{
				LogLevel: "info",
			},
		},
	}

	checker := NewContractChecker(config)

	tests := []struct {
		name     string
		protocol string
		expected bool
	}{
		{
			name:     "Known protocol - http",
			protocol: "http",
			expected: true,
		},
		{
			name:     "Known protocol - ibmmq",
			protocol: "ibmmq",
			expected: true,
		},
		{
			name:     "Known protocol - kafka",
			protocol: "kafka",
			expected: true,
		},
		{
			name:     "Unknown protocol - https",
			protocol: "https",
			expected: false,
		},
		{
			name:     "Unknown protocol - custom",
			protocol: "custom-protocol",
			expected: false,
		},
		{
			name:     "Empty protocol",
			protocol: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.isProtocolKnown(tt.protocol)
			if result != tt.expected {
				t.Errorf("Expected protocol '%s' known status %v, got %v", tt.protocol, tt.expected, result)
			}
		})
	}
}

// TestUnknownProtocolBreakingChanges тестирует генерацию breaking changes для неизвестных протоколов
func TestUnknownProtocolBreakingChanges(t *testing.T) {
	config := &Config{
		ContractTests: ContractTestsConfig{
			Consumer: ConsumerConfig{
				Name:     "test-consumer",
				SpecPath: "testdata/channel_validator/testdata_consumer_unknown_protocol.yml",
				Channels: []string{"testChannel"},
			},
			Provider: ProviderConfig{
				Name:    "test-provider",
				SpecURL: "testdata/channel_validator/testdata_provider_unknown_protocol.yml",
			},
			Settings: SettingsConfig{
				LogLevel: "info",
			},
		},
	}

	checker := NewContractChecker(config)
	result, err := checker.CheckCompatibility()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Проверяем, что контракт несовместим из-за неизвестного протокола
	if result.IsCompatible {
		t.Error("Expected compatibility to be false for unknown protocol")
	}

	// Проверяем наличие breaking changes о неизвестных протоколах
	foundConsumerBreakingChange := false
	foundProviderBreakingChange := false

	for _, change := range result.BreakingChanges {
		if change.Type == "unknown_protocol" {
			if strings.Contains(change.Field, "channels.testChannel.protocol") {
				foundConsumerBreakingChange = true
			}
			if strings.Contains(change.Field, "provider.channels.testChannel.protocol") {
				foundProviderBreakingChange = true
			}
		}
	}

	if !foundConsumerBreakingChange {
		t.Error("Expected breaking change about unknown consumer protocol")
	}
	if !foundProviderBreakingChange {
		t.Error("Expected breaking change about unknown provider protocol")
	}

	// Проверяем содержимое breaking change
	if len(result.BreakingChanges) > 0 {
		change := result.BreakingChanges[0]
		if !strings.Contains(change.Description, "custom-protocol") {
			t.Errorf("Expected breaking change description to contain 'custom-protocol', got: %s", change.Description)
		}
		if change.Severity != "high" {
			t.Errorf("Expected breaking change severity to be 'high', got: %s", change.Severity)
		}
	}
}

// TestConsumerUnknownProtocol тестирует случай, когда у потребителя указан неизвестный протокол
func TestConsumerUnknownProtocol(t *testing.T) {
	config := &Config{
		ContractTests: ContractTestsConfig{
			Consumer: ConsumerConfig{
				Name:     "test-consumer-https",
				SpecPath: "testdata/channel_validator/testdata_consumer_https.yml",
				Channels: []string{"restGetBalanceRequest"},
			},
			Provider: ProviderConfig{
				Name:    "test-provider-http",
				SpecURL: "testdata/channel_validator/testdata_provider_success.yml",
			},
			Settings: SettingsConfig{
				LogLevel: "debug",
			},
		},
	}

	checker := NewContractChecker(config)
	result, err := checker.CheckCompatibility()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Проверяем, что контракт несовместим
	if result.IsCompatible {
		t.Error("Expected compatibility to be false when consumer uses unknown protocol")
	}

	// Проверяем наличие breaking change о неизвестном протоколе потребителя
	foundConsumerBreakingChange := false
	for _, change := range result.BreakingChanges {
		if change.Type == "unknown_protocol" &&
			strings.Contains(change.Field, "channels.restGetBalanceRequest.protocol") &&
			strings.Contains(change.Description, "https") {
			foundConsumerBreakingChange = true
			break
		}
	}

	if !foundConsumerBreakingChange {
		t.Error("Expected breaking change about unknown consumer protocol 'https'")
		t.Logf("Breaking changes found: %+v", result.BreakingChanges)
	}

	// Проверяем наличие breaking change о несовместимости протоколов
	foundProtocolMismatch := false
	for _, change := range result.BreakingChanges {
		if change.Type == "protocol_mismatch" &&
			strings.Contains(change.Description, "https") &&
			strings.Contains(change.Description, "http") {
			foundProtocolMismatch = true
			break
		}
	}

	if !foundProtocolMismatch {
		t.Error("Expected breaking change about protocol mismatch (https vs http)")
		t.Logf("Breaking changes found: %+v", result.BreakingChanges)
	}
}

// TestProviderUnknownProtocol тестирует случай, когда у провайдера указан неизвестный протокол
func TestProviderUnknownProtocol(t *testing.T) {
	config := &Config{
		ContractTests: ContractTestsConfig{
			Consumer: ConsumerConfig{
				Name:     "test-consumer-http",
				SpecPath: "testdata/channel_validator/testdata_consumer_asyncapi.yml",
				Channels: []string{"restGetBalanceRequest"},
			},
			Provider: ProviderConfig{
				Name:    "test-provider-grpc",
				SpecURL: "testdata/channel_validator/testdata_provider_unknown_protocol_only.yml",
			},
			Settings: SettingsConfig{
				LogLevel: "debug",
			},
		},
	}

	checker := NewContractChecker(config)
	result, err := checker.CheckCompatibility()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Проверяем, что контракт несовместим
	if result.IsCompatible {
		t.Error("Expected compatibility to be false when provider uses unknown protocol")
	}

	// Проверяем наличие breaking change о неизвестном протоколе провайдера
	foundProviderBreakingChange := false
	for _, change := range result.BreakingChanges {
		if change.Type == "unknown_protocol" &&
			strings.Contains(change.Field, "provider.channels.") &&
			strings.Contains(change.Description, "grpc") {
			foundProviderBreakingChange = true
			break
		}
	}

	if !foundProviderBreakingChange {
		t.Error("Expected breaking change about unknown provider protocol 'grpc'")
		t.Logf("Breaking changes found: %+v", result.BreakingChanges)
	}

	// Проверяем наличие breaking change о несовместимости протоколов
	foundProtocolMismatch := false
	for _, change := range result.BreakingChanges {
		if change.Type == "protocol_mismatch" &&
			strings.Contains(change.Description, "grpc") &&
			strings.Contains(change.Description, "http") {
			foundProtocolMismatch = true
			break
		}
	}

	if !foundProtocolMismatch {
		t.Error("Expected breaking change about protocol mismatch (http vs grpc)")
		t.Logf("Breaking changes found: %+v", result.BreakingChanges)
	}
}

// TestProtocolValidationIntegration интеграционный тест для демонстрации всех случаев валидации протоколов
func TestProtocolValidationIntegration(t *testing.T) {
	tests := []struct {
		name                   string
		configFile             string
		expectedCompatible     bool
		expectedBreakingType   string
		expectedProtocol       string
		expectedBreakingChange string
	}{
		{
			name:                   "Consumer with unknown protocol (https)",
			configFile:             "testdata/channel_validator/testdata_config_consumer_https.yaml",
			expectedCompatible:     false,
			expectedBreakingType:   "unknown_protocol",
			expectedProtocol:       "https",
			expectedBreakingChange: "protocol_mismatch",
		},
		{
			name:                   "Provider with unknown protocol (grpc)",
			configFile:             "testdata/channel_validator/testdata_config_provider_unknown.yaml",
			expectedCompatible:     false,
			expectedBreakingType:   "unknown_protocol",
			expectedProtocol:       "grpc",
			expectedBreakingChange: "protocol_mismatch",
		},
		{
			name:                   "Both with unknown protocols (custom)",
			configFile:             "testdata/channel_validator/testdata_config_unknown_protocol.yaml",
			expectedCompatible:     false,
			expectedBreakingType:   "unknown_protocol",
			expectedProtocol:       "custom-protocol",
			expectedBreakingChange: "protocol_mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Загружаем конфигурацию
			config, err := LoadConfig(tt.configFile)
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			// Создаем checker и проверяем совместимость
			checker := NewContractChecker(config)
			result, err := checker.CheckCompatibility()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Проверяем совместимость
			if result.IsCompatible != tt.expectedCompatible {
				t.Errorf("Expected compatibility %v, got %v", tt.expectedCompatible, result.IsCompatible)
			}

			// Проверяем breaking changes для неизвестных протоколов
			foundBreakingChange := false
			for _, change := range result.BreakingChanges {
				if change.Type == tt.expectedBreakingType &&
					strings.Contains(change.Description, tt.expectedProtocol) {
					foundBreakingChange = true
					break
				}
			}

			if !foundBreakingChange {
				t.Errorf("Expected breaking change about %s protocol '%s'", tt.expectedBreakingType, tt.expectedProtocol)
				t.Logf("Breaking changes found: %+v", result.BreakingChanges)
			}

			// Проверяем дополнительные breaking changes (например, protocol_mismatch)
			foundAdditionalBreakingChange := false
			for _, change := range result.BreakingChanges {
				if change.Type == tt.expectedBreakingChange {
					foundAdditionalBreakingChange = true
					break
				}
			}

			if !foundAdditionalBreakingChange {
				t.Errorf("Expected breaking change of type '%s'", tt.expectedBreakingChange)
				t.Logf("Breaking changes found: %+v", result.BreakingChanges)
			}
		})
	}
}
