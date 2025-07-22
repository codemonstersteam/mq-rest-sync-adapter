package validator

import (
	"github.com/codemonsters/contract-tests-v2/parser"
)

// ===== ОБЩИЕ ТИПЫ =====

// Config представляет конфигурацию из contract-tests.yaml
type Config struct {
	ContractTests ContractTests `yaml:"contract_tests"`
}

// ContractTests основная секция конфигурации
type ContractTests struct {
	Consumer ConsumerConfig `yaml:"consumer"`
	Provider ProviderConfig `yaml:"provider"`
	Settings Settings       `yaml:"settings"`
}

// ConsumerConfig конфигурация потребителя
type ConsumerConfig struct {
	SpecPath string   `yaml:"spec_path"`
	Name     string   `yaml:"name"`
	Channels []string `yaml:"channels"`
}

// ProviderConfig конфигурация поставщика
type ProviderConfig struct {
	SpecURL  string `yaml:"spec_url"`
	SpecPath string `yaml:"spec_path"`
	Name     string `yaml:"name"`
}

// Settings настройки валидации
type Settings struct {
	LogLevel       string `yaml:"log_level"`
	SaveJSONReport bool   `yaml:"save_json_report"`
	JSONReportFile string `yaml:"json_report_file"`
	Timeout        int    `yaml:"timeout"`
	IgnoreWarnings bool   `yaml:"ignore_warnings"`
}

// ===== КОНТРАКТНЫЕ ТИПЫ =====

// ContractValidate структура для валидации контракта
type ContractValidate struct {
	ConsumerChannelName string
	ProviderSpec        *parser.AsyncAPISpec
	ConsumerSpec        *parser.AsyncAPISpec
}

// ValidationResult результат валидации
type ValidationResult struct {
	IsValid             bool
	ConsumerChannelName string
	ConsumerSpec        *parser.AsyncAPISpec
	ProviderSpec        *parser.AsyncAPISpec
	Errors              []string
}

// ===== LEGACY CHANNEL VALIDATOR ТИПЫ =====

// LegacyChannelConfig старая структура для обратной совместимости
type LegacyChannelConfig struct {
	Consumer LegacyConsumerConfig
	Provider LegacyProviderConfig
}

type LegacyConsumerConfig struct {
	SpecPath string
	Channel  string
}

type LegacyProviderConfig struct {
	SpecPath string
}

// ChannelValidationResult результат валидации каналов
type ChannelValidationResult struct {
	ConsumerChannel ChannelInfo
	ProviderChannel ChannelInfo
}

// ChannelInfo информация о канале
type ChannelInfo struct {
	Name       string
	Protocol   string
	OutMessage *MessageInfo
	InMessage  *MessageInfo
}

// MessageInfo информация о сообщении
type MessageInfo struct {
	Name        string
	ContentType string
	Payload     map[string]interface{}
}
