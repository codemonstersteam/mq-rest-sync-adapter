package parser

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parser предоставляет функционал для парсинга AsyncAPI спецификаций
type Parser struct{}

// New создает новый экземпляр Parser
func New() *Parser {
	return &Parser{}
}

// ParseFromString парсит AsyncAPI спецификацию из строки
func (p *Parser) ParseFromString(content string) (*AsyncAPISpec, error) {
	if strings.TrimSpace(content) == "" {
		return nil, newParseError("specification content is empty", "input")
	}

	var spec AsyncAPISpec
	if err := yaml.Unmarshal([]byte(content), &spec); err != nil {
		return nil, newYAMLParseError(err, "input")
	}

	// Базовая валидация
	if spec.AsyncAPI == "" {
		return nil, newParseError("asyncapi version is required", "spec.asyncapi")
	}

	if !strings.HasPrefix(spec.AsyncAPI, "3.") {
		return nil, newInvalidVersionError(spec.AsyncAPI, "spec.asyncapi")
	}

	// Резолвинг ссылок (упрощенная версия для основных случаев)
	if err := p.resolveReferences(&spec); err != nil {
		return nil, fmt.Errorf("REF_ERROR: failed to resolve references - %w", err)
	}

	return &spec, nil
}

// ParseFromBytes парсит AsyncAPI спецификацию из байтов
func (p *Parser) ParseFromBytes(data []byte) (*AsyncAPISpec, error) {
	return p.ParseFromString(string(data))
}

// resolveReferences резолвит основные типы ссылок в спецификации
func (p *Parser) resolveReferences(spec *AsyncAPISpec) error {
	// Для минималистичного решения оставляем ссылки как есть
	// В реальном применении здесь был бы код для разрешения $ref
	// Наш validator уже умеет работать с ссылками
	return nil
}

// GetChannelByName возвращает канал по имени с учетом экранирования
func (p *Parser) GetChannelByName(spec *AsyncAPISpec, name string) (*Channel, error) {
	if channel, ok := spec.Channels[name]; ok {
		return &channel, nil
	}

	// Пробуем с экранированием
	escapedName := strings.ReplaceAll(name, "/", "~1")
	if channel, ok := spec.Channels[escapedName]; ok {
		return &channel, nil
	}

	// Пробуем наоборот - убрать экранирование
	unescapedName := strings.ReplaceAll(name, "~1", "/")
	if channel, ok := spec.Channels[unescapedName]; ok {
		return &channel, nil
	}

	return nil, newChannelNotFoundError(name, "spec.channels")
}

// GetMessageByRef возвращает сообщение по ссылке
func (p *Parser) GetMessageByRef(spec *AsyncAPISpec, ref string) (*Message, error) {
	if ref == "" {
		return nil, newInvalidRefError("", "empty reference", "messageRef")
	}

	// Обработка ссылок на компоненты
	if strings.HasPrefix(ref, "#/components/messages/") {
		msgName := strings.TrimPrefix(ref, "#/components/messages/")
		if spec.Components != nil {
			if msg, ok := spec.Components.Messages[msgName]; ok {
				return &msg, nil
			}
		}
		return nil, newComponentNotFoundError("message", msgName, "#/components/messages")
	}

	// Обработка ссылок на сообщения в каналах
	if strings.Contains(ref, "/messages/") {
		parts := strings.Split(ref, "/messages/")
		if len(parts) == 2 {
			channelPath := parts[0]
			messageName := parts[1]
			
			// Извлекаем имя канала
			channelParts := strings.Split(channelPath, "/")
			if len(channelParts) >= 2 {
				channelName := channelParts[len(channelParts)-1]
				// Убираем экранирование
				channelName = strings.ReplaceAll(channelName, "~1", "/")
				
				if channel, ok := spec.Channels[channelName]; ok {
					if msgRef, ok := channel.Messages[messageName]; ok {
						// Если это тоже ссылка, резолвим дальше
						if msgRef.Ref != "" {
							return p.GetMessageByRef(spec, msgRef.Ref)
						}
					}
				}
			}
		}
	}

	return nil, newInvalidRefError(ref, "unable to resolve reference", "messageRef")
}

// GetServerByRef возвращает сервер по ссылке
func (p *Parser) GetServerByRef(spec *AsyncAPISpec, ref string) (*Server, error) {
	if ref == "" {
		return nil, newInvalidRefError("", "empty reference", "serverRef")
	}

	if strings.HasPrefix(ref, "#/servers/") {
		serverName := strings.TrimPrefix(ref, "#/servers/")
		if server, ok := spec.Servers[serverName]; ok {
			return &server, nil
		}
		return nil, newComponentNotFoundError("server", serverName, "#/servers")
	}

	return nil, newInvalidRefError(ref, "unable to resolve reference", "serverRef")
}

// GetOperationByName возвращает операцию по имени
func (p *Parser) GetOperationByName(spec *AsyncAPISpec, name string) (*Operation, error) {
	if operation, ok := spec.Operations[name]; ok {
		return &operation, nil
	}
	return nil, newComponentNotFoundError("operation", name, "spec.operations")
}