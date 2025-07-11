package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func main() {
	// Определяем путь к конфигурационному файлу
	configPath := "contract-tests.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	// Загружаем конфигурацию
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("❌ Ошибка загрузки конфигурации: %v", err)
	}

	// Создаем проверщик контрактов
	checker := NewContractChecker(config)

	// Выполняем проверку совместимости
	result, err := checker.CheckCompatibility()
	if err != nil {
		log.Fatalf("❌ Ошибка проверки совместимости: %v", err)
	}

	// Выводим результат
	fmt.Printf("📊 Результат проверки совместимости контрактов\n")
	fmt.Printf("📋 Потребитель: %s v%s\n", result.ConsumerName, result.ConsumerVersion)
	fmt.Printf("🏢 Провайдер: %s v%s\n", result.ProviderName, result.ProviderVersion)

	// Выводим информацию о проверяемых каналах
	fmt.Printf("\n📡 Проверяемые каналы взаимодействия:\n")
	for _, channelName := range config.ContractTests.Consumer.Channels {
		fmt.Printf("  • %s\n", channelName)
	}

	// Загружаем спецификации для получения адресов каналов
	consumerSpec, err := checker.loadConsumerSpec()
	if err == nil {
		fmt.Printf("\n🔗 Адреса каналов потребителя:\n")
		for _, channelName := range config.ContractTests.Consumer.Channels {
			if channel, exists := consumerSpec.Channels[channelName]; exists {
				channelType := checker.GetChannelType(channel, consumerSpec)
				fmt.Printf("  • %s (протокол: %s): %s\n", channelName, channelType, channel.Address)
			}
		}
	}

	if result.IsCompatible {
		fmt.Printf("\n✅ Контракты совместимы\n")
		fmt.Printf("🎯 Все адреса каналов потребителя поддерживаются провайдером\n")

		if len(result.Warnings) > 0 {
			fmt.Printf("\n⚠️  Предупреждений: %d\n", len(result.Warnings))
			if config.GetLogLevel() == "debug" {
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

	// Сохраняем JSON отчет если нужно
	if config.ContractTests.Settings.SaveJSONReport {
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Printf("⚠️ Ошибка создания JSON отчета: %v", err)
		} else {
			filename := config.ContractTests.Settings.JSONReportFile
			if filename == "" {
				filename = "compatibility_report.json"
			}

			err = os.WriteFile(filename, jsonData, 0644)
			if err != nil {
				log.Printf("⚠️ Ошибка сохранения JSON отчета: %v", err)
			} else {
				fmt.Printf("📄 JSON отчет сохранен: %s\n", filename)
			}
		}
	}

	// Выходим с кодом 1 если есть критические изменения
	if !result.IsCompatible {
		os.Exit(1)
	}
}
