package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// ContractValidator объединяет валидацию каналов и формирование итогового сообщения
type ContractValidator struct {
	config           *Config
	specLoader       *SpecLoader
	checker          *ContractChecker
	messageValidator *MessageValidator
	messageResult    *MessageValidationResult // Результат валидации сообщений для вывода
}

// NewContractValidator создает новый валидатор контрактов
func NewContractValidator(config *Config) *ContractValidator {
	specLoader := NewSpecLoader(config)
	return &ContractValidator{
		config:           config,
		specLoader:       specLoader,
		checker:          NewContractCheckerWithSpecLoader(config, specLoader),
		messageValidator: NewMessageValidatorWithSpecLoader(config, specLoader),
	}
}

// ValidateContracts выполняет валидацию контрактов и выводит результат
func (cv *ContractValidator) ValidateContracts() (*CompatibilityResult, error) {
	// Выполняем проверку совместимости каналов
	result, err := cv.checker.CheckCompatibility()
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки совместимости: %w", err)
	}

	// Если каналы совместимы, выполняем валидацию сообщений
	if result.IsCompatible {
		err = cv.validateMessages(result)
		if err != nil {
			return nil, fmt.Errorf("ошибка валидации сообщений: %w", err)
		}
	}

	// Выводим результат в консоль
	cv.printResult(result)

	// Сохраняем JSON отчет если нужно
	if cv.config.ContractTests.Settings.SaveJSONReport {
		cv.saveJSONReport(result)
	}

	return result, nil
}

// validateMessages выполняет валидацию сообщений для совместимых каналов
func (cv *ContractValidator) validateMessages(result *CompatibilityResult) error {
	// Используем информацию о валидных каналах из результата проверки
	validatedChannels := result.ValidatedChannels
	
	// Выполняем валидацию структуры сообщений
	messageResult, err := cv.messageValidator.ValidateMessages(validatedChannels)
	if err != nil {
		return fmt.Errorf("ошибка валидации структуры сообщений: %w", err)
	}

	// Сохраняем результат валидации сообщений для вывода
	cv.messageResult = messageResult

	// Если есть ошибки валидации сообщений, добавляем их в результат
	if !messageResult.IsValid {
		result.IsCompatible = false
		// Преобразуем ошибки валидации сообщений в BreakingChange
		for _, msgError := range messageResult.MessageErrors {
			if msgError.Severity == "high" {
				result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
					Type:        msgError.Type,
					Field:       msgError.Field,
					Description: msgError.Description,
					Severity:    msgError.Severity,
				})
			}
		}
	}

	return nil
}


// printResult выводит результат валидации в консоль
func (cv *ContractValidator) printResult(result *CompatibilityResult) {
	// Выводим основную информацию
	fmt.Printf("📊 Результат проверки совместимости контрактов\n")
	fmt.Printf("📋 Потребитель: %s v%s\n", result.ConsumerName, result.ConsumerVersion)
	fmt.Printf("   📄 Спецификация: %s\n", cv.config.ContractTests.Consumer.SpecPath)
	fmt.Printf("🏢 Провайдер: %s v%s\n", result.ProviderName, result.ProviderVersion)
	fmt.Printf("   🌐 Спецификация: %s\n", cv.config.ContractTests.Provider.SpecURL)

	// Выводим информацию о проверяемых каналах
	fmt.Printf("\n📡 Проверяемые каналы взаимодействия:\n")
	for _, channelName := range cv.config.ContractTests.Consumer.Channels {
		fmt.Printf("  • %s\n", channelName)
	}

	// Выводим адреса каналов потребителя
	cv.printConsumerChannels()

	// Выводим результаты валидации сообщений если она была выполнена
	if cv.messageResult != nil {
		cv.printMessageValidationResults()
	}

	// Выводим итоговый результат
	if result.IsCompatible {
		fmt.Printf("\n✅ Контракты совместимы\n")
		fmt.Printf("🎯 Все адреса каналов потребителя поддерживаются провайдером\n")

		if len(result.Warnings) > 0 {
			fmt.Printf("\n⚠️  Предупреждений: %d\n", len(result.Warnings))
			if cv.config.GetLogLevel() == "debug" {
				for _, warning := range result.Warnings {
					fmt.Printf("  • %s: %s\n", warning.Type, warning.Message)
				}
			}
		}
	} else {
		fmt.Printf("\n❌ Контракты НЕ совместимы\n")
		fmt.Printf("🚨 Критических изменений: %d\n", len(result.BreakingChanges))

		fmt.Printf("\n🔍 Детали несовместимости:\n")
		for _, change := range result.BreakingChanges {
			fmt.Printf("  • %s: %s\n", change.Type, change.Description)
		}
	}
}

// printConsumerChannels выводит адреса каналов потребителя
func (cv *ContractValidator) printConsumerChannels() {
	consumerSpec, err := cv.specLoader.GetConsumerSpec()
	if err == nil {
		fmt.Printf("\n🔗 Адреса каналов потребителя:\n")
		for _, channelName := range cv.config.ContractTests.Consumer.Channels {
			if channel, exists := consumerSpec.Channels[channelName]; exists {
				channelType := cv.checker.getChannelType(channel, consumerSpec)
				fmt.Printf("  • %s (протокол: %s): %s\n", channelName, channelType, channel.Address)
			}
		}
	}
}

// saveJSONReport сохраняет JSON отчет
func (cv *ContractValidator) saveJSONReport(result *CompatibilityResult) {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("⚠️ Ошибка создания JSON отчета: %v\n", err)
		return
	}

	filename := cv.config.ContractTests.Settings.JSONReportFile
	if filename == "" {
		filename = "compatibility_report.json"
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		fmt.Printf("⚠️ Ошибка сохранения JSON отчета: %v\n", err)
	} else {
		fmt.Printf("📄 JSON отчет сохранен: %s\n", filename)
	}
}

// printMessageValidationResults выводит детальную информацию о валидации сообщений
func (cv *ContractValidator) printMessageValidationResults() {
	if cv.messageResult == nil {
		return
	}

	fmt.Printf("\n💬 Результаты валидации структуры сообщений:\n")

	// Показываем исходящие сообщения потребителя
	if len(cv.messageResult.ConsumerOutbound) > 0 {
		fmt.Printf("\n📤 Исходящие сообщения потребителя:\n")
		for _, msg := range cv.messageResult.ConsumerOutbound {
			fmt.Printf("  • %s (%s) - канал: %s\n", msg.Name, msg.ContentType, msg.Channel)
		}
	}

	// Показываем входящие сообщения провайдера
	if len(cv.messageResult.ProviderInbound) > 0 {
		fmt.Printf("\n📥 Входящие сообщения провайдера:\n")
		for _, msg := range cv.messageResult.ProviderInbound {
			fmt.Printf("  • %s (%s) - канал: %s\n", msg.Name, msg.ContentType, msg.Channel)
		}
	}

	// Показываем совместимость outbound->inbound
	if len(cv.messageResult.ConsumerOutbound) > 0 && len(cv.messageResult.ProviderInbound) > 0 {
		fmt.Printf("\n✅ Совместимость исходящих сообщений потребителя с входящими провайдера:\n")
		for _, consumerMsg := range cv.messageResult.ConsumerOutbound {
			for _, providerMsg := range cv.messageResult.ProviderInbound {
				fmt.Printf("  • Исходящее сообщение потребителя '%s' канала %s валидно по контракту с входящим сообщением '%s' канала провайдера %s\n",
					consumerMsg.Name,
					consumerMsg.Channel,
					providerMsg.Name,
					providerMsg.Channel)
			}
		}
	}

	// Показываем исходящие сообщения провайдера
	if len(cv.messageResult.ProviderOutbound) > 0 {
		fmt.Printf("\n📤 Исходящие сообщения провайдера:\n")
		for _, msg := range cv.messageResult.ProviderOutbound {
			fmt.Printf("  • %s (%s) - канал: %s\n", msg.Name, msg.ContentType, msg.Channel)
		}
	}

	// Показываем входящие сообщения потребителя
	if len(cv.messageResult.ConsumerInbound) > 0 {
		fmt.Printf("\n📥 Входящие сообщения потребителя:\n")
		for _, msg := range cv.messageResult.ConsumerInbound {
			fmt.Printf("  • %s (%s) - канал: %s\n", msg.Name, msg.ContentType, msg.Channel)
		}
	}

	// Показываем совместимость inbound<-outbound
	if len(cv.messageResult.ProviderOutbound) > 0 && len(cv.messageResult.ConsumerInbound) > 0 {
		fmt.Printf("\n✅ Совместимость исходящих сообщений провайдера с входящими потребителя:\n")
		for _, providerMsg := range cv.messageResult.ProviderOutbound {
			for _, consumerMsg := range cv.messageResult.ConsumerInbound {
				fmt.Printf("  • Исходящее сообщение провайдера '%s' канала %s валидно по контракту с входящим сообщением '%s' канала потребителя %s\n",
					providerMsg.Name,
					providerMsg.Channel,
					consumerMsg.Name,
					consumerMsg.Channel)
			}
		}
	}

	// Показываем ошибки валидации если есть
	if len(cv.messageResult.MessageErrors) > 0 {
		fmt.Printf("\n🔍 Выявленные проблемы валидации сообщений:\n")
		for _, msgError := range cv.messageResult.MessageErrors {
			severityIcon := "ℹ️"
			if msgError.Severity == "high" {
				severityIcon = "❌"
			} else if msgError.Severity == "medium" {
				severityIcon = "⚠️"
			}
			fmt.Printf("  %s %s (%s): %s\n", severityIcon, msgError.Type, msgError.Severity, msgError.Description)
		}
	}
}

