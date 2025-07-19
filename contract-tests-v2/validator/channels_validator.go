package validator

import (
	"fmt"
	"strings"

	"github.com/codemonsters/contract-tests-v2/parser"
)

// ChannelValidator структура валидатора каналов
type ChannelValidator struct{}

// NewChannelValidator создает новый экземпляр валидатора каналов
func NewChannelValidator() *ChannelValidator {
	return &ChannelValidator{}
}

// ValidateChannels принимает структуру ContractValidate и возвращает полную информацию о каналах
func (v *ChannelValidator) ValidateChannels(contractValidate *ContractValidate) (*ChannelValidationResult, error) {
	// Извлекаем информацию о канале потребителя
	consumerChannel, err := v.extractConsumerChannel(contractValidate)
	if err != nil {
		return nil, fmt.Errorf("failed to extract consumer channel: %w", err)
	}

	// Ищем соответствующий канал поставщика
	providerChannel, err := v.findMatchingProviderChannel(contractValidate, consumerChannel)
	if err != nil {
		return nil, fmt.Errorf("failed to find matching provider channel: %w", err)
	}

	return &ChannelValidationResult{
		ConsumerChannel: *consumerChannel,
		ProviderChannel: *providerChannel,
	}, nil
}

// extractConsumerChannel извлекает информацию о канале потребителя
func (v *ChannelValidator) extractConsumerChannel(contractValidate *ContractValidate) (*ChannelInfo, error) {
	channelName := contractValidate.ConsumerChannelName
	spec := contractValidate.ConsumerSpec

	// Проверяем существование канала
	channel, exists := spec.Channels[channelName]
	if !exists {
		return nil, fmt.Errorf("channel %s not found in consumer spec", channelName)
	}

	// Извлекаем протокол канала
	protocol, err := v.extractChannelProtocol(spec, &channel)
	if err != nil {
		return nil, fmt.Errorf("failed to extract protocol for channel %s: %w", channelName, err)
	}

	// Извлекаем сообщения из операций
	outMessage, inMessage, err := v.extractConsumerMessages(spec, channelName)
	if err != nil {
		return nil, fmt.Errorf("failed to extract messages for channel %s: %w", channelName, err)
	}

	return &ChannelInfo{
		Name:       channelName,
		Protocol:   protocol,
		OutMessage: outMessage,
		InMessage:  inMessage,
	}, nil
}

// findMatchingProviderChannel находит соответствующий канал поставщика
func (v *ChannelValidator) findMatchingProviderChannel(contractValidate *ContractValidate, consumerChannel *ChannelInfo) (*ChannelInfo, error) {
	spec := contractValidate.ProviderSpec

	// Перебираем все каналы поставщика
	for channelName, channel := range spec.Channels {
		// Проверяем протокол
		protocol, err := v.extractChannelProtocol(spec, &channel)
		if err != nil {
			continue // Пропускаем каналы с проблемными протоколами
		}

		if protocol != consumerChannel.Protocol {
			continue // Протокол не совпадает
		}

		// Извлекаем сообщения поставщика
		inMessage, outMessage, err := v.extractProviderMessages(spec, channelName)
		if err != nil {
			continue // Пропускаем каналы с проблемными сообщениями
		}

		// Проверяем совместимость сообщений
		// Если сообщения не найдены или не совместимы - канал не подходит
		if consumerChannel.OutMessage != nil && inMessage != nil {
			if v.areMessagesCompatible(consumerChannel.OutMessage, inMessage) {
				return &ChannelInfo{
					Name:       channelName,
					Protocol:   protocol,
					InMessage:  inMessage,
					OutMessage: outMessage,
				}, nil
			}
		}
		// Если сообщения не найдены или не совместимы - продолжаем поиск
	}

	return nil, fmt.Errorf("no compatible provider channel found for protocol %s", consumerChannel.Protocol)
}

// extractChannelProtocol извлекает протокол канала из серверов
func (v *ChannelValidator) extractChannelProtocol(spec *parser.AsyncAPISpec, channel *parser.Channel) (string, error) {
	if len(channel.Servers) == 0 {
		return "", fmt.Errorf("no servers defined for channel")
	}

	// Берем первый сервер
	serverRef := channel.Servers[0].Ref
	if !strings.HasPrefix(serverRef, "#/servers/") {
		return "", fmt.Errorf("invalid server reference: %s", serverRef)
	}

	serverName := strings.TrimPrefix(serverRef, "#/servers/")
	server, exists := spec.Servers[serverName]
	if !exists {
		return "", fmt.Errorf("server %s not found", serverName)
	}

	return server.Protocol, nil
}

// extractConsumerMessages извлекает сообщения потребителя из операций
func (v *ChannelValidator) extractConsumerMessages(spec *parser.AsyncAPISpec, channelName string) (*MessageInfo, *MessageInfo, error) {
	var outMessage, inMessage *MessageInfo

	// Экранируем имя канала для поиска в ссылках
	escapedChannelName := v.escapeChannelName(channelName)
	channelRef := "#/channels/" + escapedChannelName

	// Ищем операции, которые ссылаются на этот канал
	for _, operation := range spec.Operations {
		if operation.Channel.Ref == channelRef && operation.Action == "send" {
			// Для потребителя action "send" означает исходящее сообщение
			if len(operation.Messages) > 0 {
				msg, err := v.extractMessageInfo(spec, operation.Messages[0])
				if err == nil {
					outMessage = msg
				}
			}

			// Проверяем есть ли reply для входящего сообщения
			if operation.Reply != nil && len(operation.Reply.Messages) > 0 {
				msg, err := v.extractMessageInfo(spec, operation.Reply.Messages[0])
				if err == nil {
					inMessage = msg
				}
			}
			break
		}
	}

	return outMessage, inMessage, nil
}

// extractProviderMessages извлекает сообщения поставщика из операций
func (v *ChannelValidator) extractProviderMessages(spec *parser.AsyncAPISpec, channelName string) (*MessageInfo, *MessageInfo, error) {
	var inMessage, outMessage *MessageInfo

	// Экранируем имя канала для поиска в ссылках
	escapedChannelName := v.escapeChannelName(channelName)
	channelRef := "#/channels/" + escapedChannelName

	// Ищем операции, которые ссылаются на этот канал
	for _, operation := range spec.Operations {
		if operation.Channel.Ref == channelRef && operation.Action == "receive" {
			// Для поставщика action "receive" означает входящее сообщение
			if len(operation.Messages) > 0 {
				msg, err := v.extractMessageInfo(spec, operation.Messages[0])
				if err == nil {
					inMessage = msg
				}
			}

			// Проверяем есть ли reply для исходящего сообщения
			if operation.Reply != nil && len(operation.Reply.Messages) > 0 {
				msg, err := v.extractMessageInfo(spec, operation.Reply.Messages[0])
				if err == nil {
					outMessage = msg
				}
			}
			break
		}
	}

	return inMessage, outMessage, nil
}

// extractMessageInfo извлекает информацию о сообщении
func (v *ChannelValidator) extractMessageInfo(spec *parser.AsyncAPISpec, msgRef parser.MessageRef) (*MessageInfo, error) {
	// Проверяем ссылку на компонент
	if strings.HasPrefix(msgRef.Ref, "#/components/messages/") {
		msgName := strings.TrimPrefix(msgRef.Ref, "#/components/messages/")
		if spec.Components != nil && spec.Components.Messages != nil {
			if msg, exists := spec.Components.Messages[msgName]; exists {
				return &MessageInfo{
					Name:        msg.Name,
					ContentType: msg.ContentType,
					Payload:     v.convertSchema(msg.Payload),
				}, nil
			}
		}
	}

	// Проверяем ссылку на сообщение в канале
	if strings.Contains(msgRef.Ref, "/messages/") {
		parts := strings.Split(msgRef.Ref, "/messages/")
		if len(parts) == 2 {
			channelPath := parts[0]
			messageName := parts[1]

			// Извлекаем имя канала из пути
			channelParts := strings.Split(channelPath, "/")
			if len(channelParts) >= 2 {
				channelName := v.unescapeChannelName(channelParts[len(channelParts)-1])

				if channel, exists := spec.Channels[channelName]; exists && channel.Messages != nil {
					if msgData, exists := channel.Messages[messageName]; exists {
						// Если это ссылка на компонент, разыменовываем её
						if msgData.Ref != "" {
							return v.extractMessageInfo(spec, parser.MessageRef{Ref: msgData.Ref})
						}
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("message not found: %s", msgRef.Ref)
}

// convertSchema преобразует схему в map[string]interface{}
func (v *ChannelValidator) convertSchema(schema *parser.Schema) map[string]interface{} {
	if schema == nil {
		return nil
	}

	result := make(map[string]interface{})
	result["type"] = schema.Type

	if len(schema.Properties) > 0 {
		properties := make(map[string]interface{})
		for name, prop := range schema.Properties {
			properties[name] = v.convertSchema(&prop)
		}
		result["properties"] = properties
	}

	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}

	return result
}

// areMessagesCompatible проверяет совместимость сообщений
func (v *ChannelValidator) areMessagesCompatible(msg1, msg2 *MessageInfo) bool {
	if msg1 == nil || msg2 == nil {
		return false
	}

	// Проверяем наличие свойств
	props1, ok1 := msg1.Payload["properties"].(map[string]interface{})
	props2, ok2 := msg2.Payload["properties"].(map[string]interface{})

	if !ok1 || !ok2 {
		return false
	}

	// Проверяем обязательные поля
	required1 := v.getRequiredFields(msg1.Payload)
	required2 := v.getRequiredFields(msg2.Payload)

	if len(required1) != len(required2) {
		return false
	}

	// Проверяем, что все обязательные поля присутствуют в обеих схемах
	for _, field := range required1 {
		if _, exists := props2[field]; !exists {
			return false
		}
		if _, exists := props1[field]; !exists {
			return false
		}
	}

	return true
}

// getRequiredFields извлекает список обязательных полей
func (v *ChannelValidator) getRequiredFields(payload map[string]interface{}) []string {
	if required, ok := payload["required"].([]interface{}); ok {
		result := make([]string, len(required))
		for i, field := range required {
			if str, ok := field.(string); ok {
				result[i] = str
			}
		}
		return result
	}
	return []string{}
}

// escapeChannelName экранирует имя канала для использования в JSON Pointer
func (v *ChannelValidator) escapeChannelName(name string) string {
	return strings.ReplaceAll(name, "/", "~1")
}

// unescapeChannelName разэкранирует имя канала
func (v *ChannelValidator) unescapeChannelName(name string) string {
	return strings.ReplaceAll(name, "~1", "/")
}