package main

import (
	"testing"
)

// TestFindConsumerOutboundMessage тестирует поиск исходящего сообщения от потребителя
func TestFindConsumerOutboundMessage(t *testing.T) {
	// Создаем конфигурацию для тестирования
	config := &Config{
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

	// Создаем message validator
	messageValidator := NewMessageValidator(config)

	// Загружаем спецификацию потребителя
	consumerSpec, err := messageValidator.loadConsumerSpec()
	if err != nil {
		t.Fatalf("Failed to load consumer spec: %v", err)
	}

	// Тестируем поиск исходящего сообщения для канала restGetBalanceRequest
	channelName := "restGetBalanceRequest"
	outboundMessage, err := messageValidator.FindConsumerOutboundMessage(channelName, consumerSpec)
	if err != nil {
		t.Fatalf("Failed to find consumer outbound message: %v", err)
	}

	// Проверяем, что сообщение найдено
	if outboundMessage == nil {
		t.Error("Expected to find outbound message for consumer channel restGetBalanceRequest")
		return
	}

	// Проверяем конкретные значения из спецификации
	if outboundMessage.Name != "PsGetBalanceRequest" {
		t.Errorf("Expected message name 'PsGetBalanceRequest', got '%s'", outboundMessage.Name)
	}

	if outboundMessage.Title != "Http Request PsGetBalanceRequest" {
		t.Errorf("Expected title 'Http Request PsGetBalanceRequest', got '%s'", outboundMessage.Title)
	}

	if outboundMessage.ContentType != "application/json" {
		t.Errorf("Expected ContentType 'application/json', got '%s'", outboundMessage.ContentType)
	}

	// Проверяем, что сообщение имеет headers и payload
	if outboundMessage.Headers == nil {
		t.Error("Expected outbound message to have headers")
	}

	if outboundMessage.Payload == nil {
		t.Error("Expected outbound message to have payload")
	}

	t.Logf("✅ Found correct outbound message: %s", outboundMessage.Name)
	t.Logf("   Title: %s", outboundMessage.Title)
	t.Logf("   ContentType: %s", outboundMessage.ContentType)
}

// TestExtractMessageHeaderFields тестирует извлечение полей заголовков из сообщения
func TestExtractMessageHeaderFields(t *testing.T) {
	// Создаем конфигурацию для тестирования
	config := &Config{
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

	// Создаем message validator
	messageValidator := NewMessageValidator(config)

	// Загружаем спецификацию потребителя
	consumerSpec, err := messageValidator.loadConsumerSpec()
	if err != nil {
		t.Fatalf("Failed to load consumer spec: %v", err)
	}

	// Находим исходящее сообщение
	channelName := "restGetBalanceRequest"
	outboundMessage, err := messageValidator.FindConsumerOutboundMessage(channelName, consumerSpec)
	if err != nil || outboundMessage == nil {
		t.Fatalf("Failed to find consumer outbound message for header field extraction test")
	}

	// Извлекаем поля заголовков
	headerFields, err := messageValidator.ExtractMessageHeaderFields(outboundMessage, consumerSpec)
	if err != nil {
		t.Fatalf("Failed to extract message header fields: %v", err)
	}

	// Проверяем, что поля заголовков найдены
	if len(headerFields) == 0 {
		t.Error("Expected to find header fields")
		return
	}

	// Ожидаемые поля заголовков из спецификации psGetBalanceRequestHeaders
	expectedHeaderFields := []string{"Api-clientId", "Content-Type"} // Из спецификации
	foundHeaderFields := make(map[string]bool)
	
	for _, field := range headerFields {
		foundHeaderFields[field.Name] = true
		t.Logf("Found header field: %s (type: %s)", field.Name, field.Type)
	}

	// Проверяем наличие всех ожидаемых полей заголовков
	for _, expectedField := range expectedHeaderFields {
		if !foundHeaderFields[expectedField] {
			t.Errorf("Expected to find header field '%s'", expectedField)
		}
	}

	t.Logf("✅ Found all expected header fields")
}

// TestExtractMessagePayloadFields тестирует извлечение полей payload из сообщения
func TestExtractMessagePayloadFields(t *testing.T) {
	// Создаем конфигурацию для тестирования
	config := &Config{
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

	// Создаем message validator
	messageValidator := NewMessageValidator(config)

	// Загружаем спецификацию потребителя
	consumerSpec, err := messageValidator.loadConsumerSpec()
	if err != nil {
		t.Fatalf("Failed to load consumer spec: %v", err)
	}

	// Находим исходящее сообщение
	channelName := "restGetBalanceRequest"
	outboundMessage, err := messageValidator.FindConsumerOutboundMessage(channelName, consumerSpec)
	if err != nil || outboundMessage == nil {
		t.Fatalf("Failed to find consumer outbound message for payload field extraction test")
	}

	// Извлекаем поля payload
	payloadFields, err := messageValidator.ExtractMessagePayloadFields(outboundMessage, consumerSpec)
	if err != nil {
		t.Fatalf("Failed to extract message payload fields: %v", err)
	}

	// Проверяем, что поля payload найдены
	if len(payloadFields) == 0 {
		t.Error("Expected to find payload fields")
		return
	}

	// Ожидаемые поля payload из спецификации psGetBalanceRequestPayload
	expectedPayloadFields := []string{"requestType", "dateFrom", "dateTo", "operTime"}
	requiredFields := []string{"requestType", "dateFrom", "dateTo", "operTime"} // Все поля обязательные
	
	foundPayloadFields := make(map[string]MessageField)
	
	for _, field := range payloadFields {
		foundPayloadFields[field.Name] = field
		t.Logf("Found payload field: %s (type: %s, required: %t)", field.Name, field.Type, field.Required)
	}

	// Проверяем наличие всех ожидаемых полей payload
	for _, expectedField := range expectedPayloadFields {
		if _, found := foundPayloadFields[expectedField]; !found {
			t.Errorf("Expected to find payload field '%s'", expectedField)
		}
	}

	// Проверяем типы полей
	expectedTypes := map[string]string{
		"requestType": "string",
		"dateFrom":    "number",
		"dateTo":      "number", 
		"operTime":    "number",
	}

	for fieldName, expectedType := range expectedTypes {
		if field, found := foundPayloadFields[fieldName]; found {
			if field.Type != expectedType {
				t.Errorf("Expected field '%s' to have type '%s', got '%s'", fieldName, expectedType, field.Type)
			}
		}
	}

	// Проверяем, что все поля обязательные
	for _, requiredField := range requiredFields {
		if field, found := foundPayloadFields[requiredField]; found {
			if !field.Required {
				t.Errorf("Expected field '%s' to be required", requiredField)
			}
		}
	}

	t.Logf("✅ Found all expected payload fields with correct types and requirements")
}


// TestFindConsumerOutboundMessageNotFound тестирует случай, когда сообщение не найдено
func TestFindConsumerOutboundMessageNotFound(t *testing.T) {
	config := &Config{
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

	messageValidator := NewMessageValidator(config)
	consumerSpec, err := messageValidator.loadConsumerSpec()
	if err != nil {
		t.Fatalf("Failed to load consumer spec: %v", err)
	}

	// Тестируем поиск несуществующего канала
	nonExistentChannel := "nonExistentChannel"
	outboundMessage, err := messageValidator.FindConsumerOutboundMessage(nonExistentChannel, consumerSpec)
	
	// Ожидаем ошибку или nil сообщение
	if err == nil && outboundMessage != nil {
		t.Error("Expected no outbound message for non-existent channel")
	}

	if err != nil {
		t.Logf("Expected error for non-existent channel: %v", err)
	}
}

// TestFindProviderInboundMessage тестирует поиск входящего сообщения у провайдера
func TestFindProviderInboundMessage(t *testing.T) {
	// Создаем конфигурацию для тестирования
	config := &Config{
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
	messageValidator := NewMessageValidator(config)

	// Загружаем спецификацию провайдера
	providerSpec, err := messageValidator.loadProviderSpec()
	if err != nil {
		t.Fatalf("Failed to load provider spec: %v", err)
	}

	// Тестируем поиск входящего сообщения для канала walletBalanceRequest
	channelName := "walletBalanceRequest"
	inboundMessage, err := messageValidator.FindProviderInboundMessage(channelName, providerSpec)
	if err != nil {
		t.Fatalf("Failed to find provider inbound message: %v", err)
	}

	// Проверяем, что сообщение найдено
	if inboundMessage == nil {
		t.Error("Expected to find inbound message for provider channel walletBalanceRequest")
		return
	}

	// Проверяем конкретные значения из спецификации провайдера
	if inboundMessage.Name != "GetBalanceRequest" {
		t.Errorf("Expected message name 'GetBalanceRequest', got '%s'", inboundMessage.Name)
	}

	if inboundMessage.Title != "Запрос на получение баланса кошелька" {
		t.Errorf("Expected title 'Запрос на получение баланса кошелька', got '%s'", inboundMessage.Title)
	}

	if inboundMessage.ContentType != "application/json" {
		t.Errorf("Expected ContentType 'application/json', got '%s'", inboundMessage.ContentType)
	}

	// Проверяем, что сообщение имеет headers и payload
	if inboundMessage.Headers == nil {
		t.Error("Expected inbound message to have headers")
	}

	if inboundMessage.Payload == nil {
		t.Error("Expected inbound message to have payload")
	}

	t.Logf("✅ Found correct inbound message: %s", inboundMessage.Name)
	t.Logf("   Title: %s", inboundMessage.Title)
	t.Logf("   ContentType: %s", inboundMessage.ContentType)
}

// TestExtractProviderInboundMessagePayloadFields тестирует извлечение полей payload входящего сообщения провайдера
func TestExtractProviderInboundMessagePayloadFields(t *testing.T) {
	// Создаем конфигурацию для тестирования
	config := &Config{
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
	messageValidator := NewMessageValidator(config)

	// Загружаем спецификацию провайдера
	providerSpec, err := messageValidator.loadProviderSpec()
	if err != nil {
		t.Fatalf("Failed to load provider spec: %v", err)
	}

	// Находим входящее сообщение
	channelName := "walletBalanceRequest"
	inboundMessage, err := messageValidator.FindProviderInboundMessage(channelName, providerSpec)
	if err != nil || inboundMessage == nil {
		t.Fatalf("Failed to find provider inbound message for payload field extraction test")
	}

	// Извлекаем поля payload
	payloadFields, err := messageValidator.ExtractMessagePayloadFields(inboundMessage, providerSpec)
	if err != nil {
		t.Fatalf("Failed to extract message payload fields: %v", err)
	}

	// Проверяем, что поля payload найдены
	if len(payloadFields) == 0 {
		t.Error("Expected to find payload fields")
		return
	}

	// Ожидаемые поля payload из спецификации balanceRequestPayload
	expectedPayloadFields := []string{"requestType", "dateFrom", "dateTo", "operTime"}
	foundPayloadFields := make(map[string]MessageField)
	
	for _, field := range payloadFields {
		foundPayloadFields[field.Name] = field
		t.Logf("Found provider payload field: %s (type: %s, required: %t)", field.Name, field.Type, field.Required)
	}

	// Проверяем наличие всех ожидаемых полей payload
	for _, expectedField := range expectedPayloadFields {
		if _, found := foundPayloadFields[expectedField]; !found {
			t.Errorf("Expected to find payload field '%s' in provider message", expectedField)
		}
	}

	// Проверяем типы полей (должны совпадать с потребителем)
	expectedTypes := map[string]string{
		"requestType": "string",
		"dateFrom":    "number",
		"dateTo":      "number", 
		"operTime":    "number",
	}

	for fieldName, expectedType := range expectedTypes {
		if field, found := foundPayloadFields[fieldName]; found {
			if field.Type != expectedType {
				t.Errorf("Expected provider field '%s' to have type '%s', got '%s'", fieldName, expectedType, field.Type)
			}
		}
	}

	t.Logf("✅ Found all expected payload fields in provider message with correct types")
}

// TestFindProviderInboundMessageNotFound тестирует случай, когда входящее сообщение провайдера не найдено
func TestFindProviderInboundMessageNotFound(t *testing.T) {
	config := &Config{
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

	messageValidator := NewMessageValidator(config)
	providerSpec, err := messageValidator.loadProviderSpec()
	if err != nil {
		t.Fatalf("Failed to load provider spec: %v", err)
	}

	// Тестируем поиск несуществующего канала
	nonExistentChannel := "nonExistentChannel"
	inboundMessage, err := messageValidator.FindProviderInboundMessage(nonExistentChannel, providerSpec)
	
	// Ожидаем ошибку или nil сообщение
	if err == nil && inboundMessage != nil {
		t.Error("Expected no inbound message for non-existent channel")
	}

	if err != nil {
		t.Logf("Expected error for non-existent channel: %v", err)
	}
}