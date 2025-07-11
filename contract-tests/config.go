package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config представляет конфигурацию для проверки совместимости контрактов
type Config struct {
	ContractTests ContractTestsConfig `yaml:"contract_tests"`
}

// ContractTestsConfig содержит настройки для контрактных тестов
type ContractTestsConfig struct {
	Consumer ConsumerConfig `yaml:"consumer"`
	Provider ProviderConfig `yaml:"provider"`
	Settings SettingsConfig `yaml:"settings"`
}

// ConsumerConfig содержит настройки потребителя
type ConsumerConfig struct {
	SpecPath string   `yaml:"spec_path"`
	Name     string   `yaml:"name"`
	Channels []string `yaml:"channels"`
}

// ProviderConfig содержит настройки поставщика
type ProviderConfig struct {
	SpecURL string `yaml:"spec_url"`
	Name    string `yaml:"name"`
}

// SettingsConfig содержит общие настройки
type SettingsConfig struct {
	LogLevel       string `yaml:"log_level"`
	SaveJSONReport bool   `yaml:"save_json_report"`
	JSONReportFile string `yaml:"json_report_file"`
	Timeout        int    `yaml:"timeout"`
	IgnoreWarnings bool   `yaml:"ignore_warnings"`
}

// LoadConfig загружает конфигурацию из YAML файла
func LoadConfig(configPath string) (*Config, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Валидация конфигурации
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	if c.ContractTests.Consumer.SpecPath == "" {
		return fmt.Errorf("consumer spec_path is required")
	}
	if c.ContractTests.Consumer.Name == "" {
		return fmt.Errorf("consumer name is required")
	}
	if c.ContractTests.Provider.SpecURL == "" {
		return fmt.Errorf("provider spec_url is required")
	}
	if c.ContractTests.Provider.Name == "" {
		return fmt.Errorf("provider name is required")
	}

	// Валидация уровня логирования
	validLogLevels := []string{"debug", "info", "warn", "error"}
	logLevel := c.ContractTests.Settings.LogLevel
	if logLevel == "" {
		logLevel = "info" // значение по умолчанию
	}

	validLevel := false
	for _, level := range validLogLevels {
		if level == logLevel {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("invalid log_level: %s, must be one of: %v", logLevel, validLogLevels)
	}

	// Валидация timeout
	if c.ContractTests.Settings.Timeout <= 0 {
		c.ContractTests.Settings.Timeout = 30 // значение по умолчанию
	}

	// Установка значений по умолчанию
	if c.ContractTests.Settings.JSONReportFile == "" {
		c.ContractTests.Settings.JSONReportFile = "compatibility_report.json"
	}

	return nil
}

// GetTimeout возвращает timeout как time.Duration
func (c *Config) GetTimeout() time.Duration {
	return time.Duration(c.ContractTests.Settings.Timeout) * time.Second
}

// GetLogLevel возвращает уровень логирования
func (c *Config) GetLogLevel() string {
	if c.ContractTests.Settings.LogLevel == "" {
		return "info"
	}
	return c.ContractTests.Settings.LogLevel
}
