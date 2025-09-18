package validator

import (
	"fmt"
	"strings"
)

// ===== HELPER ФУНКЦИИ ДЛЯ СТАНДАРТИЗИРОВАННЫХ ОШИБОК VALIDATOR =====

// newValidationError создает стандартизированную ошибку валидации
func newValidationError(message, location string) error {
	return fmt.Errorf("VALIDATION_ERROR: %s at %s", message, location)
}

// newSchemaMismatchError создает ошибку несовпадения схемы
func newSchemaMismatchError(field, expected, actual, location string) error {
	return fmt.Errorf("SCHEMA_MISMATCH: field '%s' type mismatch - expected %s, got %s at %s", 
		field, expected, actual, location)
}

// newFieldMissingError создает ошибку отсутствующего поля
func newFieldMissingError(field, location string) error {
	return fmt.Errorf("FIELD_MISSING_ERROR: required field '%s' missing at %s", field, location)
}

// newTypeMismatchError создает ошибку несовместимости типов
func newTypeMismatchError(expected, actual, location string) error {
	return fmt.Errorf("TYPE_MISMATCH: expected %s, got %s at %s", expected, actual, location)
}

// newProtocolMismatchError создает ошибку несовпадения протоколов
func newProtocolMismatchError(consumerProtocol, providerProtocol, location string) error {
	return fmt.Errorf("PROTOCOL_MISMATCH: consumer protocol '%s' incompatible with provider protocol '%s' at %s", 
		consumerProtocol, providerProtocol, location)
}

// newChannelNotFoundError создает ошибку отсутствующего канала
func newChannelNotFoundErrorValidator(channelName, location string) error {
	return fmt.Errorf("CHANNEL_NOT_FOUND: channel '%s' not found at %s", channelName, location)
}

// newMessageNotFoundError создает ошибку отсутствующего сообщения
func newMessageNotFoundError(messageName, location string) error {
	return fmt.Errorf("MESSAGE_NOT_FOUND: message '%s' not found at %s", messageName, location)
}

// newConfigError создает ошибку конфигурации
func newConfigError(message, location string) error {
	return fmt.Errorf("CONFIG_ERROR: %s at %s", message, location)
}

// newMissingFieldError создает ошибку отсутствующего обязательного поля
func newMissingFieldError(fieldName, location string) error {
	return fmt.Errorf("MISSING_FIELD_ERROR: required field '%s' is missing at %s", fieldName, location)
}

// newInvalidPathError создает ошибку неверного пути
func newInvalidPathError(path, reason, location string) error {
	return fmt.Errorf("INVALID_PATH_ERROR: path '%s' invalid - %s at %s", path, reason, location)
}

// newHTTPError создает ошибку HTTP запроса
func newHTTPError(statusCode int, url, location string) error {
	return fmt.Errorf("HTTP_ERROR: HTTP %d when fetching %s at %s", statusCode, url, location)
}

// newTimeoutError создает ошибку таймаута
func newTimeoutError(operation, location string) error {
	return fmt.Errorf("TIMEOUT_ERROR: timeout during %s at %s", operation, location)
}

// newFileNotFoundError создает ошибку отсутствующего файла  
func newFileNotFoundError(filePath, location string) error {
	return fmt.Errorf("FILE_NOT_FOUND: file '%s' not found at %s", filePath, location)
}

// newDetailedValidationError создает подробную ошибку валидации с контекстом
func newDetailedValidationError(code ErrorCode, message, location string, context map[string]interface{}) *ValidationError {
	err := NewValidationError(code, message, location)
	for k, v := range context {
		err.WithContext(k, v)
	}
	return err
}

// formatChannelCompatibilityError форматирует ошибку совместимости каналов с деталями
func formatChannelCompatibilityError(consumerChannelName string, providerChannelsAnalyzed int, incompatibleDetails []string) error {
	var detailsStr string
	if len(incompatibleDetails) > 0 {
		detailsStr = fmt.Sprintf("\nAnalyzed channels: %d\nIncompatibility details:\n- %s", 
			providerChannelsAnalyzed, strings.Join(incompatibleDetails, "\n- "))
	}
	
	return fmt.Errorf("VALIDATION_ERROR: no compatible provider channel found for consumer channel '%s'%s at channel.matching", 
		consumerChannelName, detailsStr)
}

// formatRequiredFieldsError форматирует ошибку с информацией о required полях
func formatRequiredFieldsError(field string, consumerRequired, providerRequired []string, location string) error {
	return fmt.Errorf("FIELD_MISSING_ERROR: required field '%s' missing at %s\nConsumer required: [%s]\nProvider has: [%s]", 
		field, location, strings.Join(consumerRequired, ", "), strings.Join(providerRequired, ", "))
}

// formatContentTypeError форматирует ошибку несовпадения ContentType
func formatContentTypeError(consumerType, providerType, location string) error {
	return fmt.Errorf("TYPE_MISMATCH: content type mismatch - consumer expects '%s', provider provides '%s' at %s", 
		consumerType, providerType, location)
}