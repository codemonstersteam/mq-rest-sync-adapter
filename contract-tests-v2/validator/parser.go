package validator

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/codemonsters/contract-tests-v2/parser"
)

// Parser предоставляет функционал для парсинга AsyncAPI спецификаций
type Parser struct{}

// NewParser создает новый экземпляр Parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseFromString парсит AsyncAPI спецификацию из строки
func (p *Parser) ParseFromString(content string) (*parser.AsyncAPISpec, error) {
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("specification content is empty")
	}

	var spec parser.AsyncAPISpec
	if err := yaml.Unmarshal([]byte(content), &spec); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Базовая валидация
	if spec.AsyncAPI == "" {
		return nil, fmt.Errorf("asyncapi version is required")
	}

	if !strings.HasPrefix(spec.AsyncAPI, "3.") {
		return nil, fmt.Errorf("only AsyncAPI 3.x is supported, got %s", spec.AsyncAPI)
	}

	// Резолвинг ссылок (упрощенная версия для основных случаев)
	if err := p.resolveReferences(&spec); err != nil {
		return nil, fmt.Errorf("failed to resolve references: %w", err)
	}

	return &spec, nil
}

// ParseFromBytes парсит AsyncAPI спецификацию из байтов
func (p *Parser) ParseFromBytes(data []byte) (*parser.AsyncAPISpec, error) {
	return p.ParseFromString(string(data))
}

// resolveReferences резолвит основные типы ссылок в спецификации
func (p *Parser) resolveReferences(spec *parser.AsyncAPISpec) error {
	// Для минималистичного решения оставляем ссылки как есть
	// В реальном применении здесь был бы код для разрешения $ref
	// Наш validator уже умеет работать с ссылками
	return nil
}

// GetChannelByName возвращает канал по имени с учетом экранирования
func (p *Parser) GetChannelByName(spec *parser.AsyncAPISpec, name string) (*parser.Channel, error) {
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

	return nil, fmt.Errorf("channel %s not found", name)
}

// GetMessageByRef возвращает сообщение по ссылке
func (p *Parser) GetMessageByRef(spec *parser.AsyncAPISpec, ref string) (*parser.Message, error) {
	if ref == "" {
		return nil, fmt.Errorf("empty reference")
	}

	// Обработка ссылок на компоненты
	if strings.HasPrefix(ref, "#/components/messages/") {
		msgName := strings.TrimPrefix(ref, "#/components/messages/")
		if spec.Components != nil {
			if msg, ok := spec.Components.Messages[msgName]; ok {
				return &msg, nil
			}
		}
		return nil, fmt.Errorf("message %s not found in components", msgName)
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

	return nil, fmt.Errorf("unable to resolve message reference: %s", ref)
}

// GetServerByRef возвращает сервер по ссылке
func (p *Parser) GetServerByRef(spec *parser.AsyncAPISpec, ref string) (*parser.Server, error) {
	if ref == "" {
		return nil, fmt.Errorf("empty reference")
	}

	if strings.HasPrefix(ref, "#/servers/") {
		serverName := strings.TrimPrefix(ref, "#/servers/")
		if server, ok := spec.Servers[serverName]; ok {
			return &server, nil
		}
		return nil, fmt.Errorf("server %s not found", serverName)
	}

	return nil, fmt.Errorf("unable to resolve server reference: %s", ref)
}

// GetOperationByName возвращает операцию по имени
func (p *Parser) GetOperationByName(spec *parser.AsyncAPISpec, name string) (*parser.Operation, error) {
	if operation, ok := spec.Operations[name]; ok {
		return &operation, nil
	}
	return nil, fmt.Errorf("operation %s not found", name)
}

// ExtractChannelProtocol извлекает протокол канала через связанные серверы
func (p *Parser) ExtractChannelProtocol(spec *parser.AsyncAPISpec, channel *parser.Channel) (string, error) {
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