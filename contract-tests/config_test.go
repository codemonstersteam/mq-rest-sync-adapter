package main

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Тест загрузки валидной конфигурации
	config, err := LoadConfig("testdata/testdata-contract-tests.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.ContractTests.Consumer.Name != "test-consumer" {
		t.Errorf("Expected consumer name 'test-consumer', got '%s'", config.ContractTests.Consumer.Name)
	}

	if config.ContractTests.Provider.Name != "test-provider" {
		t.Errorf("Expected provider name 'test-provider', got '%s'", config.ContractTests.Provider.Name)
	}

	// Тест несуществующего файла
	_, err = LoadConfig("non-existent-file.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestConfigValidation(t *testing.T) {
	// Тест валидной конфигурации
	validConfig := Config{
		ContractTests: ContractTestsConfig{
			Consumer: ConsumerConfig{
				SpecPath: "test.yml",
				Name:     "test-consumer",
			},
			Provider: ProviderConfig{
				SpecURL: "http://example.com/spec.yml",
				Name:    "test-provider",
			},
		},
	}

	if err := validConfig.Validate(); err != nil {
		t.Errorf("Valid config should not fail validation: %v", err)
	}

	// Тест отсутствующих обязательных полей
	invalidConfig := Config{
		ContractTests: ContractTestsConfig{
			Consumer: ConsumerConfig{
				Name: "test-consumer",
				// SpecPath отсутствует
			},
			Provider: ProviderConfig{
				SpecURL: "http://example.com/spec.yml",
				Name:    "test-provider",
			},
		},
	}

	if err := invalidConfig.Validate(); err == nil {
		t.Error("Expected validation error for missing spec_path")
	}
}
