package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// ContractValidator объединяет валидацию каналов и формирование итогового сообщения
type ContractValidator struct {
	config  *Config
	checker *ContractChecker
}

// NewContractValidator создает новый валидатор контрактов
func NewContractValidator(config *Config) *ContractValidator {
	return &ContractValidator{
		config:  config,
		checker: NewContractChecker(config),
	}
}

// ValidateContracts выполняет валидацию контрактов и выводит результат
func (cv *ContractValidator) ValidateContracts() (*CompatibilityResult, error) {
	// Выполняем проверку совместимости
	result, err := cv.checker.CheckCompatibility()
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки совместимости: %w", err)
	}

	// Выводим результат в консоль
	cv.printResult(result)

	// Сохраняем JSON отчет если нужно
	if cv.config.ContractTests.Settings.SaveJSONReport {
		cv.saveJSONReport(result)
	}

	return result, nil
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
	consumerSpec, err := cv.checker.loadConsumerSpec()
	if err == nil {
		fmt.Printf("\n🔗 Адреса каналов потребителя:\n")
		for _, channelName := range cv.config.ContractTests.Consumer.Channels {
			if channel, exists := consumerSpec.Channels[channelName]; exists {
				channelType := cv.checker.GetChannelType(channel, consumerSpec)
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