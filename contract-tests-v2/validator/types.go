package validator

import (
	"fmt"
	
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
	Name          string
	ContentType   string
	Payload       map[string]interface{}
	Headers       map[string]interface{} // Заголовки сообщения (headers schema)
	CorrelationId *CorrelationIdInfo     // Информация о correlation ID
}

// CorrelationIdInfo информация о correlation ID
type CorrelationIdInfo struct {
	Location    string // Расположение correlation ID в сообщении
	Description string // Описание использования correlation ID
}

// ===== ТИПИЗИРОВАННЫЕ ОШИБКИ =====

// ErrorCode константы для типов ошибок
type ErrorCode string

const (
	// Ошибки парсинга
	ParseErrorCode           ErrorCode = "PARSE_ERROR"
	YAMLParseErrorCode      ErrorCode = "YAML_PARSE_ERROR"
	InvalidVersionErrorCode ErrorCode = "INVALID_VERSION_ERROR"
	
	// Ошибки конфигурации
	ConfigErrorCode         ErrorCode = "CONFIG_ERROR"
	MissingFieldErrorCode   ErrorCode = "MISSING_FIELD_ERROR"
	InvalidPathErrorCode    ErrorCode = "INVALID_PATH_ERROR"
	
	// Ошибки валидации
	ValidationErrorCode     ErrorCode = "VALIDATION_ERROR"
	SchemaMismatchErrorCode ErrorCode = "SCHEMA_MISMATCH"
	TypeMismatchErrorCode   ErrorCode = "TYPE_MISMATCH"
	FieldMissingErrorCode   ErrorCode = "FIELD_MISSING_ERROR"
	
	// Ошибки ссылок
	RefErrorCode            ErrorCode = "REF_ERROR"
	ComponentNotFoundErrorCode ErrorCode = "COMPONENT_NOT_FOUND"
	InvalidRefErrorCode     ErrorCode = "INVALID_REF_ERROR"
	
	// Ошибки каналов
	ChannelNotFoundErrorCode ErrorCode = "CHANNEL_NOT_FOUND"
	ProtocolMismatchErrorCode ErrorCode = "PROTOCOL_MISMATCH"
	MessageNotFoundErrorCode ErrorCode = "MESSAGE_NOT_FOUND"
	
	// Ошибки HTTP
	HTTPErrorCode           ErrorCode = "HTTP_ERROR"
	TimeoutErrorCode        ErrorCode = "TIMEOUT_ERROR"
	FileNotFoundErrorCode   ErrorCode = "FILE_NOT_FOUND"
)

// ValidationError типизированная ошибка с контекстом
type ValidationError struct {
	Code     ErrorCode              `json:"code"`
	Message  string                 `json:"message"`
	Context  map[string]interface{} `json:"context,omitempty"`
	Location string                 `json:"location,omitempty"`
	Details  string                 `json:"details,omitempty"`
}

// Error реализует интерфейс error
func (e *ValidationError) Error() string {
	if e.Location != "" {
		return fmt.Sprintf("%s: %s at %s", e.Code, e.Message, e.Location)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewValidationError создает новую типизированную ошибку
func NewValidationError(code ErrorCode, message string, location string) *ValidationError {
	return &ValidationError{
		Code:     code,
		Message:  message,
		Location: location,
		Context:  make(map[string]interface{}),
	}
}

// WithContext добавляет контекст к ошибке
func (e *ValidationError) WithContext(key string, value interface{}) *ValidationError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithDetails добавляет детальное описание ошибки
func (e *ValidationError) WithDetails(details string) *ValidationError {
	e.Details = details
	return e
}

// IsValidationError проверяет, является ли ошибка ValidationError
func IsValidationError(err error) (*ValidationError, bool) {
	if verr, ok := err.(*ValidationError); ok {
		return verr, true
	}
	return nil, false
}
