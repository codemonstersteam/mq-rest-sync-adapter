package validator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContractValidator_LoadConfig(t *testing.T) {
	t.Run("should load config from file", func(t *testing.T) {
		// Создаем временную директорию для тестов
		tempDir := t.TempDir()

		// Создаем тестовую конфигурацию
		configContent := `contract_tests:
  consumer:
    spec_path: "consumer.yaml"
    name: "test-consumer"
    channels:
      - "test/channel"
  provider:
    spec_url: "provider.yaml"
    name: "test-provider"
  settings:
    log_level: "debug"
    save_json_report: true
    json_report_file: "report.json"
    timeout: 60
    ignore_warnings: true`

		configPath := filepath.Join(tempDir, "test-config.yaml")
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Тестируем загрузку конфигурации
		validator := NewContractValidator()
		config, err := validator.loadConfig(configPath)

		require.NoError(t, err, "Should load config successfully")
		require.NotNil(t, config, "Config should not be nil")

		// Проверяем секцию consumer
		assert.Equal(t, "consumer.yaml", config.ContractTests.Consumer.SpecPath)
		assert.Equal(t, "test-consumer", config.ContractTests.Consumer.Name)
		assert.Equal(t, []string{"test/channel"}, config.ContractTests.Consumer.Channels)

		// Проверяем секцию provider
		assert.Equal(t, "provider.yaml", config.ContractTests.Provider.SpecURL)
		assert.Equal(t, "test-provider", config.ContractTests.Provider.Name)

		// Проверяем секцию settings
		assert.Equal(t, "debug", config.ContractTests.Settings.LogLevel)
		assert.True(t, config.ContractTests.Settings.SaveJSONReport)
		assert.Equal(t, "report.json", config.ContractTests.Settings.JSONReportFile)
		assert.Equal(t, 60, config.ContractTests.Settings.Timeout)
		assert.True(t, config.ContractTests.Settings.IgnoreWarnings)

		t.Logf("✅ Config loaded successfully:")
		t.Logf("   Consumer: %s (%s)", config.ContractTests.Consumer.Name, config.ContractTests.Consumer.SpecPath)
		t.Logf("   Provider: %s (%s)", config.ContractTests.Provider.Name, config.ContractTests.Provider.SpecURL)
		t.Logf("   Channels: %v", config.ContractTests.Consumer.Channels)
		t.Logf("   Settings: %+v", config.ContractTests.Settings)
	})

	t.Run("should return error for non-existent config file", func(t *testing.T) {
		validator := NewContractValidator()
		config, err := validator.loadConfig("non-existent-file.yaml")

		assert.Error(t, err, "Should return error for non-existent file")
		assert.Nil(t, config, "Config should be nil on error")
		// Проверяем стандартизированное сообщение об ошибке
		assert.Contains(t, err.Error(), "FILE_NOT_FOUND", "Error should contain FILE_NOT_FOUND code")
		assert.Contains(t, err.Error(), "non-existent-file.yaml", "Error should mention config file")
		assert.Contains(t, err.Error(), "at config.load", "Error should mention location")
	})

	t.Run("should return error for invalid YAML", func(t *testing.T) {
		// Создаем временную директорию
		tempDir := t.TempDir()

		// Создаем невалидный YAML
		invalidContent := `contract_tests:
  consumer:
    spec_path: "test.yaml"
    name: [invalid yaml structure
  provider:`

		configPath := filepath.Join(tempDir, "invalid-config.yaml")
		err := os.WriteFile(configPath, []byte(invalidContent), 0644)
		require.NoError(t, err)

		// Тестируем загрузку невалидной конфигурации
		validator := NewContractValidator()
		config, err := validator.loadConfig(configPath)

		assert.Error(t, err, "Should return error for invalid YAML")
		assert.Nil(t, config, "Config should be nil on error")
		
		// Проверяем стандартизированное сообщение об ошибке
		assert.Contains(t, err.Error(), "YAML_PARSE_ERROR", "Error should contain YAML_PARSE_ERROR code")
		assert.Contains(t, err.Error(), "failed to parse config", "Error should mention parsing")
		assert.Contains(t, err.Error(), "invalid-config.yaml", "Error should mention the file path")
	})

	t.Run("should check file paths resolution", func(t *testing.T) {
		// Проверяем, как loadConfig работает с относительными путями
		tempDir := t.TempDir()
		
		// Создаем поддиректорию для конфигурации
		configDir := filepath.Join(tempDir, "configs")
		err := os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		configContent := `contract_tests:
  consumer:
    spec_path: "../specs/consumer.yaml"
    name: "relative-path-consumer"
    channels:
      - "test/channel"
  provider:
    spec_url: "../specs/provider.yaml"
    name: "relative-path-provider"
  settings:
    log_level: "info"`

		configPath := filepath.Join(configDir, "relative-config.yaml")
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Загружаем конфигурацию
		validator := NewContractValidator()
		config, err := validator.loadConfig(configPath)

		require.NoError(t, err, "Should load config with relative paths")
		assert.Equal(t, "../specs/consumer.yaml", config.ContractTests.Consumer.SpecPath)
		assert.Equal(t, "../specs/provider.yaml", config.ContractTests.Provider.SpecURL)

		t.Logf("✅ Relative paths preserved:")
		t.Logf("   Consumer spec path: %s", config.ContractTests.Consumer.SpecPath)  
		t.Logf("   Provider spec URL: %s", config.ContractTests.Provider.SpecURL)
	})
}