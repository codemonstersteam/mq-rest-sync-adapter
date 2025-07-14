package main

import (
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

	// Создаем валидатор контрактов
	validator := NewContractValidator(config)

	// Выполняем валидацию контрактов
	result, err := validator.ValidateContracts()
	if err != nil {
		log.Fatalf("❌ %v", err)
	}

	// Выходим с кодом 1 если есть критические изменения
	if !result.IsCompatible {
		os.Exit(1)
	}
}
