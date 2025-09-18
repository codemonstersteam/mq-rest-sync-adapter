package parser

import (
	"fmt"
)

// ===== HELPER ФУНКЦИИ ДЛЯ СТАНДАРТИЗИРОВАННЫХ ОШИБОК =====

// newParseError создает стандартизированную ошибку парсинга
func newParseError(message, location string) error {
	if location != "" {
		return fmt.Errorf("PARSE_ERROR: %s at %s", message, location)
	}
	return fmt.Errorf("PARSE_ERROR: %s", message)
}

// newYAMLParseError создает ошибку парсинга YAML
func newYAMLParseError(err error, location string) error {
	return fmt.Errorf("YAML_PARSE_ERROR: failed to parse YAML at %s - %w", location, err)
}

// newInvalidVersionError создает ошибку неподдерживаемой версии
func newInvalidVersionError(version, location string) error {
	return fmt.Errorf("INVALID_VERSION_ERROR: unsupported AsyncAPI version '%s' at %s - only 3.x is supported", version, location)
}

// newComponentNotFoundError создает ошибку отсутствующего компонента
func newComponentNotFoundError(componentType, name, location string) error {
	return fmt.Errorf("COMPONENT_NOT_FOUND: %s '%s' not found at %s", componentType, name, location)
}

// newInvalidRefError создает ошибку неверной ссылки
func newInvalidRefError(ref, reason, location string) error {
	return fmt.Errorf("INVALID_REF_ERROR: invalid reference '%s' at %s - %s", ref, location, reason)
}

// newChannelNotFoundError создает ошибку отсутствующего канала
func newChannelNotFoundError(channelName, location string) error {
	return fmt.Errorf("CHANNEL_NOT_FOUND: channel '%s' not found at %s", channelName, location)
}

// ExtractFullMessage извлекает полную информацию о сообщении включая headers и payload
func (p *Parser) ExtractFullMessage(spec *AsyncAPISpec, messageRef MessageRef) (*Message, error) {
	if messageRef.Ref == "" {
		return nil, fmt.Errorf("empty message reference")
	}

	return p.GetMessageByRef(spec, messageRef.Ref)
}

// ExtractChannelProtocol извлекает протокол канала через связанные серверы
func (p *Parser) ExtractChannelProtocol(spec *AsyncAPISpec, channel *Channel) (string, error) {
	if len(channel.Servers) == 0 {
		return "", fmt.Errorf("channel has no associated servers")
	}

	// Берем первый сервер
	serverRef := channel.Servers[0]
	server, err := p.GetServerByRef(spec, serverRef.Ref)
	if err != nil {
		return "", fmt.Errorf("failed to resolve server: %w", err)
	}

	if server.Protocol == "" {
		return "", fmt.Errorf("server has no protocol specified")
	}

	return server.Protocol, nil
}

// ExtractOperationMessages извлекает все сообщения операции (основные и reply)
func (p *Parser) ExtractOperationMessages(spec *AsyncAPISpec, operation *Operation) (messages []*Message, replyMessages []*Message, err error) {
	// Извлекаем основные сообщения
	for _, msgRef := range operation.Messages {
		msg, err := p.ExtractFullMessage(spec, msgRef)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to extract message: %w", err)
		}
		messages = append(messages, msg)
	}

	// Извлекаем reply сообщения если есть
	if operation.Reply != nil && len(operation.Reply.Messages) > 0 {
		for _, msgRef := range operation.Reply.Messages {
			msg, err := p.ExtractFullMessage(spec, msgRef)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to extract reply message: %w", err)
			}
			replyMessages = append(replyMessages, msg)
		}
	}

	return messages, replyMessages, nil
}

// CompareMessageSchemas сравнивает схемы двух сообщений (payload и headers)
func (p *Parser) CompareMessageSchemas(msg1, msg2 *Message) (compatible bool, details string) {
	if msg1 == nil || msg2 == nil {
		return false, "one or both messages are nil"
	}

	// Сравниваем payload
	if !p.compareSchemas(msg1.Payload, msg2.Payload) {
		return false, "payload schemas are incompatible"
	}

	// Сравниваем headers если есть
	if msg1.Headers != nil || msg2.Headers != nil {
		if !p.compareSchemas(msg1.Headers, msg2.Headers) {
			return false, "header schemas are incompatible"
		}
	}

	// Сравниваем contentType
	if msg1.ContentType != msg2.ContentType {
		return false, fmt.Sprintf("content types differ: %s vs %s", msg1.ContentType, msg2.ContentType)
	}

	return true, "messages are compatible"
}

// compareSchemas сравнивает две схемы на совместимость
func (p *Parser) compareSchemas(schema1, schema2 *Schema) bool {
	if schema1 == nil && schema2 == nil {
		return true
	}
	
	if schema1 == nil || schema2 == nil {
		return false
	}

	// Проверяем тип
	if schema1.Type != schema2.Type {
		return false
	}

	// Для объектов проверяем required поля
	if schema1.Type == "object" {
		// Все required поля schema1 должны быть в schema2
		for _, field := range schema1.Required {
			found := false
			for _, field2 := range schema2.Required {
				if field == field2 {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}

		// Проверяем что все required поля присутствуют в properties
		for _, field := range schema1.Required {
			if _, exists := schema2.Properties[field]; !exists {
				return false
			}
		}
	}

	return true
}

// ValidateAsyncAPIVersion проверяет версию спецификации
func (p *Parser) ValidateAsyncAPIVersion(spec *AsyncAPISpec) error {
	if spec.AsyncAPI == "" {
		return fmt.Errorf("asyncapi version is not specified")
	}

	// Поддерживаем только версию 3.x
	if spec.AsyncAPI[0] != '3' {
		return fmt.Errorf("unsupported AsyncAPI version: %s (only 3.x is supported)", spec.AsyncAPI)
	}

	return nil
}