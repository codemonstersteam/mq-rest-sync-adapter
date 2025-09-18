package main

import (
	"testing"
)

// TestValidateIncompatibleResponseStructure проверяет случай, когда ответ провайдера
// не соответствует по структуре ответу, который ожидает потребитель
func TestValidateIncompatibleResponseStructure(t *testing.T) {
	// Создаем конфигурацию для тестирования
	config := &Config{
		ContractTests: ContractTestsConfig{
			Consumer: ConsumerConfig{
				Name:     "test-consumer",
				SpecPath: "testdata/structure_validator/incompatible_response_consumer_asyncapi.yml",
				Channels: []string{"restGetBalanceRequest"},
			},
			Provider: ProviderConfig{
				Name:     "test-provider",
				SpecPath: "testdata/structure_validator/incompatible_response_provider_asyncapi.yml",
			},
			Settings: SettingsConfig{
				LogLevel: "info",
			},
		},
	}

	// Создаем валидатор контрактов
	validator := NewContractValidator(config)

	// Выполняем валидацию контрактов
	result, err := validator.ValidateContracts()
	if err != nil {
		t.Fatalf("Ошибка валидации контрактов: %v", err)
	}

	// Проверяем, что контракты определены как несовместимые
	if result.IsCompatible {
		t.Error("❌ Контракты должны быть несовместимы из-за различий в структуре ответа")
	} else {
		t.Log("✅ Контракты корректно определены как несовместимые")
	}

	// Проверяем наличие ошибок валидации
	if len(result.BreakingChanges) == 0 {
		t.Error("❌ Должны быть ошибки валидации структуры ответа")
	} else {
		t.Logf("🔍 Найдены ошибки совместимости (%d):", len(result.BreakingChanges))
		for i, change := range result.BreakingChanges {
			t.Logf("  %d. %s (%s): %s", i+1, change.Type, change.Severity, change.Description)
		}
	}

	// Проверяем, что есть критическая ошибка о несовместимости структуры сообщения
	hasStructureError := false
	for _, change := range result.BreakingChanges {
		if change.Type == "message_structure_incompatible" && change.Severity == "high" {
			hasStructureError = true
			t.Logf("✅ Найдена ошибка несовместимости структуры сообщения: %s", change.Description)
			break
		}
	}

	if !hasStructureError {
		t.Error("❌ Не найдена критическая ошибка о несовместимости структуры сообщения")
	}

	// Проверяем информацию о валидированных сообщениях
	if validator.messageResult != nil {
		t.Log("\n📤 Исходящие сообщения потребителя:")
		for _, msg := range validator.messageResult.ConsumerOutbound {
			t.Logf("  • %s (%s) - канал: %s", msg.Name, msg.ContentType, msg.Channel)
		}

		t.Log("\n📥 Входящие сообщения провайдера:")
		for _, msg := range validator.messageResult.ProviderInbound {
			t.Logf("  • %s (%s) - канал: %s", msg.Name, msg.ContentType, msg.Channel)
		}

		t.Log("\n📤 Исходящие сообщения провайдера:")
		for _, msg := range validator.messageResult.ProviderOutbound {
			t.Logf("  • %s (%s) - канал: %s", msg.Name, msg.ContentType, msg.Channel)
		}

		t.Log("\n📥 Входящие сообщения потребителя:")
		for _, msg := range validator.messageResult.ConsumerInbound {
			t.Logf("  • %s (%s) - канал: %s", msg.Name, msg.ContentType, msg.Channel)
		}

		// Проверяем ошибки валидации сообщений
		if len(validator.messageResult.MessageErrors) > 0 {
			t.Log("\n🔍 Ошибки валидации сообщений:")
			for _, msgError := range validator.messageResult.MessageErrors {
				severityIcon := "ℹ️"
				if msgError.Severity == "high" {
					severityIcon = "❌"
				} else if msgError.Severity == "medium" {
					severityIcon = "⚠️"
				}
				t.Logf("  %s %s (%s): %s", severityIcon, msgError.Type, msgError.Severity, msgError.Description)
			}
		}
	}

	// Проверяем, что есть информация о валидных каналах
	if len(result.ValidatedChannels) == 0 {
		t.Error("❌ Должна быть информация о валидных каналах")
	} else {
		t.Log("\n🔗 Валидные каналы:")
		for _, channel := range result.ValidatedChannels {
			t.Logf("  • %s (%s) -> %s (%s)", 
				channel.ConsumerChannelName, channel.Protocol,
				channel.ProviderChannelName, channel.Address)
		}
	}
}