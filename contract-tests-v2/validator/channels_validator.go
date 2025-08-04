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
	
	var protocolErrors []string
	var candidateChannels []string
	channelsWithProtocol := 0

	// Перебираем все каналы поставщика
	for channelName, channel := range spec.Channels {
		// Проверяем протокол
		protocol, err := v.extractChannelProtocol(spec, &channel)
		if err != nil {
			protocolErrors = append(protocolErrors, fmt.Sprintf("%s: %v", channelName, err))
			continue // Пропускаем каналы с проблемными протоколами
		}

		if protocol != consumerChannel.Protocol {
			continue // Протокол не совпадает
		}
		
		channelsWithProtocol++
		candidateInfo := fmt.Sprintf("%s (protocol: %s)", channelName, protocol)

		// Извлекаем сообщения поставщика
		inMessage, outMessage, err := v.extractProviderMessages(spec, channelName)
		if err != nil {
			candidateChannels = append(candidateChannels, candidateInfo + fmt.Sprintf(" - error extracting messages: %v", err))
			continue // Пропускаем каналы с проблемными сообщениями
		}

		// Проверяем совместимость сообщений
		// 1. Проверяем структуру входящего сообщения поставщика с запросом потребителя
		if consumerChannel.OutMessage != nil && inMessage != nil {
			if !v.areMessagesCompatible(consumerChannel.OutMessage, inMessage, contractValidate.ConsumerSpec, contractValidate.ProviderSpec) {
				// Получаем детали о required полях для отладки
				consumerReq, _ := v.getRequiredFields(consumerChannel.OutMessage.Payload)
				providerReq, _ := v.getRequiredFields(inMessage.Payload)
				
				// Добавим информацию о payload для отладки
				consumerPayloadInfo := "nil"
				if consumerChannel.OutMessage.Payload != nil {
					consumerPayloadInfo = fmt.Sprintf("type: %T", consumerChannel.OutMessage.Payload)
				}
				providerPayloadInfo := "nil"
				if inMessage.Payload != nil {
					providerPayloadInfo = fmt.Sprintf("type: %T", inMessage.Payload)
				}
				
				candidateChannels = append(candidateChannels, candidateInfo + 
					fmt.Sprintf(" - request message incompatible (consumer: payload=%s, required=%v; provider: payload=%s, required=%v)", 
						consumerPayloadInfo, consumerReq, providerPayloadInfo, providerReq))
				continue // Структуры не совпадают, переходим к следующему каналу
			}
		} else {
			// Если нет сообщений для проверки, пропускаем канал
			if consumerChannel.OutMessage == nil {
				candidateChannels = append(candidateChannels, candidateInfo + " - consumer has no outgoing message")
			} else {
				candidateChannels = append(candidateChannels, candidateInfo + " - provider has no incoming message")
			}
			continue
		}

		// 2. Проверяем структуру исходящего сообщения поставщика с ожидаемым ответом потребителя
		if consumerChannel.InMessage != nil && outMessage != nil {
			if !v.areMessagesCompatible(consumerChannel.InMessage, outMessage, contractValidate.ConsumerSpec, contractValidate.ProviderSpec) {
				// Получаем детали о required полях для отладки
				consumerReq, _ := v.getRequiredFields(consumerChannel.InMessage.Payload)
				providerReq, _ := v.getRequiredFields(outMessage.Payload)
				candidateChannels = append(candidateChannels, candidateInfo + 
					fmt.Sprintf(" - response message incompatible (consumer required fields: %v, provider required fields: %v)", 
						consumerReq, providerReq))
				continue // Структуры ответа не совпадают, переходим к следующему каналу
			}
		} else if consumerChannel.InMessage != nil || outMessage != nil {
			// Если у одного есть reply, а у другого нет - несовместимо
			if consumerChannel.InMessage == nil {
				candidateChannels = append(candidateChannels, candidateInfo + " - consumer expects no reply but provider sends one")
			} else {
				candidateChannels = append(candidateChannels, candidateInfo + " - consumer expects reply but provider sends none")
			}
			continue
		}

		// Все проверки пройдены - возвращаем подходящий канал
		return &ChannelInfo{
			Name:       channelName,
			Protocol:   protocol,
			InMessage:  inMessage,
			OutMessage: outMessage,
		}, nil
	}

	// Формируем детальное сообщение об ошибке
	errorMsg := fmt.Sprintf("no compatible provider channel found for protocol %s", consumerChannel.Protocol)
	errorMsg += fmt.Sprintf("\nConsumer channel: %s", consumerChannel.Name)
	
	if len(spec.Channels) == 0 {
		errorMsg += "\nProvider has no channels defined"
	} else {
		errorMsg += fmt.Sprintf("\nProvider channels analyzed: %d", len(spec.Channels))
		errorMsg += fmt.Sprintf("\nChannels with matching protocol: %d", channelsWithProtocol)
	}
	
	if len(protocolErrors) > 0 {
		errorMsg += fmt.Sprintf("\n\nProtocol extraction errors:\n  - %s", strings.Join(protocolErrors, "\n  - "))
	}
	
	if len(candidateChannels) > 0 {
		errorMsg += fmt.Sprintf("\n\nChannels with matching protocol but failed validation:\n  - %s", strings.Join(candidateChannels, "\n  - "))
	}
	
	return nil, fmt.Errorf("%s", errorMsg)
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
	escapedChannelName, _ := v.escapeChannelName(channelName) // ошибки не ожидаются для валидных имен
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
	escapedChannelName, _ := v.escapeChannelName(channelName) // ошибки не ожидаются для валидных имен
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
	if spec == nil {
		return nil, fmt.Errorf("specification is nil")
	}
	
	// Проверяем ссылку на компонент
	if strings.HasPrefix(msgRef.Ref, "#/components/messages/") {
		msgName := strings.TrimPrefix(msgRef.Ref, "#/components/messages/")
		if spec.Components != nil && spec.Components.Messages != nil {
			if msg, exists := spec.Components.Messages[msgName]; exists {
				// Обрабатываем payload - может быть ссылкой
				var payload map[string]interface{}
				if msg.Payload != nil {
					if msg.Payload.Ref != "" {
						// Разрешаем ссылку на схему
						resolvedSchema := v.resolveSchemaRef(spec, msg.Payload.Ref)
						if resolvedSchema != nil {
							payload = v.convertSchema(resolvedSchema)
						}
					} else {
						payload = v.convertSchema(msg.Payload)
					}
				}
				
				return &MessageInfo{
					Name:        msgName, // Используем имя из ключа map
					ContentType: msg.ContentType,
					Payload:     payload,
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
				channelName, _ := v.unescapeChannelName(channelParts[len(channelParts)-1]) // ошибки игнорируются для обратной совместимости

				if channel, exists := spec.Channels[channelName]; exists && channel.Messages != nil {
					if msgData, exists := channel.Messages[messageName]; exists {
						// Если это ссылка на компонент, разыменовываем её
						if msgData.Ref != "" {
							return v.extractMessageInfo(spec, parser.MessageRef{Ref: msgData.Ref})
						}
						
						// Если это inline сообщение
						if msgData.Message != nil {
							// Обрабатываем payload - может быть ссылкой
							var payload map[string]interface{}
							if msgData.Message.Payload != nil {
								if msgData.Message.Payload.Ref != "" {
									// Разрешаем ссылку на схему
									resolvedSchema := v.resolveSchemaRef(spec, msgData.Message.Payload.Ref)
									if resolvedSchema != nil {
										payload = v.convertSchema(resolvedSchema)
									}
								} else {
									payload = v.convertSchema(msgData.Message.Payload)
								}
							}
							
							// Используем имя из сообщения или имя из ключа канала
							msgName := msgData.Message.Name
							if msgName == "" {
								msgName = messageName
							}
							
							return &MessageInfo{
								Name:        msgName,
								ContentType: msgData.Message.ContentType,
								Payload:     payload,
							}, nil
						}
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("message not found: %s", msgRef.Ref)
}

// resolveSchemaRef разрешает ссылку на схему согласно AsyncAPI 3.0 спецификации
func (v *ChannelValidator) resolveSchemaRef(spec *parser.AsyncAPISpec, ref string) *parser.Schema {
	if spec == nil || !strings.HasPrefix(ref, "#/components/schemas/") {
		return nil
	}
	
	schemaName := strings.TrimPrefix(ref, "#/components/schemas/")
	
	// Валидация имени схемы согласно AsyncAPI 3.0: ^[a-zA-Z0-9\.\-_]+$
	if !isValidSchemaName(schemaName) {
		return nil
	}
	
	if spec.Components != nil && spec.Components.Schemas != nil {
		if schema, exists := spec.Components.Schemas[schemaName]; exists {
			// Если схема тоже содержит ссылку, рекурсивно разрешаем её с защитой от циклов
			if schema.Ref != "" {
				return v.resolveSchemaRefWithDepth(spec, schema.Ref, 0, make(map[string]bool))
			}
			return &schema
		}
	}
	
	return nil
}

// resolveSchemaRefWithDepth разрешает ссылку с защитой от циклических ссылок
func (v *ChannelValidator) resolveSchemaRefWithDepth(spec *parser.AsyncAPISpec, ref string, depth int, visited map[string]bool) *parser.Schema {
	// Защита от слишком глубокой рекурсии
	if depth > 10 {
		return nil
	}
	
	// Защита от циклических ссылок
	if visited[ref] {
		return nil
	}
	
	if spec == nil || !strings.HasPrefix(ref, "#/components/schemas/") {
		return nil
	}
	
	schemaName := strings.TrimPrefix(ref, "#/components/schemas/")
	
	// Валидация имени схемы согласно AsyncAPI 3.0
	if !isValidSchemaName(schemaName) {
		return nil
	}
	
	if spec.Components != nil && spec.Components.Schemas != nil {
		if schema, exists := spec.Components.Schemas[schemaName]; exists {
			if schema.Ref != "" {
				// Отмечаем текущую ссылку как посещенную
				visited[ref] = true
				result := v.resolveSchemaRefWithDepth(spec, schema.Ref, depth+1, visited)
				// Убираем отметку после обработки
				delete(visited, ref)
				return result
			}
			return &schema
		}
	}
	
	return nil
}

// isValidSchemaName проверяет валидность имени схемы согласно AsyncAPI 3.0: ^[a-zA-Z0-9\.\-_]+$
func isValidSchemaName(name string) bool {
	if name == "" {
		return false
	}
	
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || 
			 (r >= 'A' && r <= 'Z') || 
			 (r >= '0' && r <= '9') || 
			 r == '.' || r == '-' || r == '_') {
			return false
		}
	}
	
	return true
}

// convertSchema преобразует схему в map[string]interface{}
func (v *ChannelValidator) convertSchema(schema *parser.Schema) map[string]interface{} {
	if schema == nil {
		return nil
	}

	result := make(map[string]interface{})
	
	// Если есть ссылка, сохраняем её как $ref
	if schema.Ref != "" {
		result["$ref"] = schema.Ref
		return result
	}
	
	result["type"] = schema.Type

	// Format поддержка (int64, double, email, uri, etc.)
	if schema.Format != "" {
		result["format"] = schema.Format
	}

	// Enum поддержка
	if len(schema.Enum) > 0 {
		result["enum"] = schema.Enum
	}

	// Properties поддержка
	if len(schema.Properties) > 0 {
		properties := make(map[string]interface{})
		for name, prop := range schema.Properties {
			properties[name] = v.convertSchema(&prop)
		}
		result["properties"] = properties
	}

	// Required поддержка
	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}

	// Items поддержка для массивов
	if schema.Items != nil {
		result["items"] = v.convertSchema(schema.Items)
	}
	
	// Description поддержка
	if schema.Description != "" {
		result["description"] = schema.Description
	}
	
	// Example поддержка
	if schema.Example != nil {
		result["example"] = schema.Example
	}
	
	// Копируем все дополнительные поля из Additional
	for key, value := range schema.Additional {
		// Не переопределяем уже установленные поля
		if _, exists := result[key]; !exists {
			result[key] = value
		}
	}

	return result
}

// areMessagesCompatible проверяет совместимость сообщений по именам полей и их типам
func (v *ChannelValidator) areMessagesCompatible(msg1, msg2 *MessageInfo, spec1, spec2 *parser.AsyncAPISpec) bool {
	if msg1 == nil || msg2 == nil {
		return false
	}

	// Проверяем совместимость ContentType
	if msg1.ContentType != msg2.ContentType {
		return false
	}

	// Проверяем наличие свойств (обязательно по нашему стандарту)
	props1, ok1 := msg1.Payload["properties"].(map[string]interface{})
	props2, ok2 := msg2.Payload["properties"].(map[string]interface{})

	if !ok1 || !ok2 {
		return false
	}

	// Проверяем обязательные поля
	required1, _ := v.getRequiredFields(msg1.Payload)
	required2, _ := v.getRequiredFields(msg2.Payload)

	if !v.areRequiredFieldsSetsIdentical(required1, required2) {
		return false
	}

	if !v.areAllRequiredFieldsCompatible(required1, props1, props2, spec1, spec2) {
		return false
	}

	return true
}

// areRequiredFieldsSetsIdentical проверяет что множества required полей идентичны
func (v *ChannelValidator) areRequiredFieldsSetsIdentical(required1, required2 []string) bool {
	if len(required1) != len(required2) {
		return false
	}
	
	// Создаем карты для быстрого поиска
	required1Map := make(map[string]bool)
	for _, field := range required1 {
		required1Map[field] = true
	}
	
	// Проверяем что все поля из required2 присутствуют в required1
	for _, field := range required2 {
		if !required1Map[field] {
			return false
		}
	}
	
	return true
}

// areAllRequiredFieldsCompatible проверяет совместимость типов всех required полей
func (v *ChannelValidator) areAllRequiredFieldsCompatible(required1 []string, props1, props2 map[string]interface{}, spec1, spec2 *parser.AsyncAPISpec) bool {
	for _, field := range required1 {
		// Проверяем наличие поля в обеих схемах
		prop1, exists1 := props1[field]
		prop2, exists2 := props2[field]
		
		if !exists1 || !exists2 {
			return false
		}

		// Проверяем совпадение типов полей
		if !v.arePropertyTypesCompatible(prop1, prop2, spec1, spec2) {
			return false
		}
	}
	return true
}

// arePropertyTypesCompatible проверяет совместимость типов свойств
func (v *ChannelValidator) arePropertyTypesCompatible(prop1, prop2 interface{}, spec1, spec2 *parser.AsyncAPISpec) bool {
	// Обрабатываем ссылки ($ref)
	ref1 := v.getPropertyRef(prop1)
	ref2 := v.getPropertyRef(prop2)
	
	// Если оба свойства содержат ссылки, сравниваем сами ссылки
	if ref1 != "" && ref2 != "" {
		// Для $ref ссылок важна точная идентичность, а не структурная совместимость
		return ref1 == ref2
	}
	
	// Если одно содержит ссылку, а другое нет, они несовместимы
	if (ref1 != "") != (ref2 != "") {
		return false
	}
	
	// Извлекаем типы из свойств (тип обязателен по стандарту)
	type1 := v.getPropertyType(prop1)
	type2 := v.getPropertyType(prop2)
	
	// Типы должны совпадать и не быть пустыми (валидация стандарта)
	if type1 == "" || type2 == "" || type1 != type2 {
		return false
	}
	
	// Для массивов проверяем совместимость элементов
	if type1 == "array" {
		return v.areArrayItemsCompatible(prop1, prop2, spec1, spec2)
	}
	
	// Для объектов проверяем совместимость свойств
	if type1 == "object" {
		return v.areObjectPropertiesCompatible(prop1, prop2, spec1, spec2)
	}
	
	// Для простых типов достаточно проверки равенства типов
	return true
}

// getPropertyRef извлекает ссылку из свойства
func (v *ChannelValidator) getPropertyRef(prop interface{}) string {
	if propMap, ok := prop.(map[string]interface{}); ok {
		if refVal, exists := propMap["$ref"]; exists {
			if refStr, ok := refVal.(string); ok {
				return refStr
			}
		}
	}
	return ""
}

// getPropertyType извлекает тип свойства из схемы
func (v *ChannelValidator) getPropertyType(prop interface{}) string {
	if propMap, ok := prop.(map[string]interface{}); ok {
		if typeVal, exists := propMap["type"]; exists {
			if typeStr, ok := typeVal.(string); ok {
				return typeStr
			}
		}
	}
	return ""
}

// getRequiredFields извлекает список обязательных полей
func (v *ChannelValidator) getRequiredFields(payload map[string]interface{}) ([]string, error) {
	if payload == nil {
		return nil, fmt.Errorf("payload is nil")
	}
	
	// Проверяем наличие поля required
	requiredField, exists := payload["required"]
	if !exists {
		return []string{}, nil // Нет required полей - это нормально
	}
	
	// Обработка []interface{} (из YAML парсера)
	if required, ok := requiredField.([]interface{}); ok {
		result := make([]string, len(required))
		for i, field := range required {
			if str, ok := field.(string); ok {
				result[i] = str
			} else {
				return nil, fmt.Errorf("invalid required field at index %d: expected string, got %T", i, field)
			}
		}
		return result, nil
	}
	
	// Обработка []string (из convertSchema)
	if required, ok := requiredField.([]string); ok {
		return required, nil
	}
	
	// Если поле required есть, но имеет неправильный тип - это ошибка
	return nil, fmt.Errorf("invalid required field type: expected []string or []interface{}, got %T", requiredField)
}

// escapeChannelName экранирует имя канала для использования в JSON Pointer по RFC 6901
func (v *ChannelValidator) escapeChannelName(name string) (string, error) {
	if name == "" {
		return "", nil // пустая строка валидна
	}
	
	// RFC 6901: сначала ~ → ~0, затем / → ~1
	result := strings.ReplaceAll(name, "~", "~0")
	result = strings.ReplaceAll(result, "/", "~1")
	return result, nil
}

// unescapeChannelName разэкранирует имя канала по RFC 6901
func (v *ChannelValidator) unescapeChannelName(name string) (string, error) {
	if name == "" {
		return "", nil // пустая строка валидна
	}
	
	// Валидация escape последовательностей
	if err := v.validateEscapeSequences(name); err != nil {
		return "", err
	}
	
	// RFC 6901: сначала ~1 → /, затем ~0 → ~
	result := strings.ReplaceAll(name, "~1", "/")
	result = strings.ReplaceAll(result, "~0", "~")
	return result, nil
}

// validateEscapeSequences проверяет корректность escape последовательностей
func (v *ChannelValidator) validateEscapeSequences(name string) error {
	for i := 0; i < len(name); i++ {
		if name[i] != '~' {
			continue
		}
		
		// Проверяем, что после ~ есть символ
		if i+1 >= len(name) {
			return fmt.Errorf("invalid escape: incomplete '~' at end")
		}
		
		// Проверяем, что после ~ идет только 0 или 1
		next := name[i+1]
		if next != '0' && next != '1' {
			return fmt.Errorf("invalid escape: '~%c' not allowed", next)
		}
	}
	return nil
}

// areArrayItemsCompatible проверяет совместимость элементов массивов
func (v *ChannelValidator) areArrayItemsCompatible(prop1, prop2 interface{}, spec1, spec2 *parser.AsyncAPISpec) bool {
	// Извлекаем items из обоих массивов
	items1 := v.getArrayItems(prop1)
	items2 := v.getArrayItems(prop2)
	
	if items1 == nil || items2 == nil {
		return false
	}
	
	// Рекурсивно проверяем совместимость элементов массива
	return v.arePropertyTypesCompatible(items1, items2, spec1, spec2)
}

// getArrayItems извлекает items из array свойства
func (v *ChannelValidator) getArrayItems(prop interface{}) interface{} {
	if propMap, ok := prop.(map[string]interface{}); ok {
		if items, exists := propMap["items"]; exists {
			return items
		}
	}
	return nil
}

// areObjectPropertiesCompatible проверяет совместимость свойств объектов
func (v *ChannelValidator) areObjectPropertiesCompatible(prop1, prop2 interface{}, spec1, spec2 *parser.AsyncAPISpec) bool {
	// Извлекаем properties и required из обоих объектов
	props1 := v.getObjectProperties(prop1)
	props2 := v.getObjectProperties(prop2)
	
	// Используем существующую функцию для извлечения required полей
	objMap1, ok1 := prop1.(map[string]interface{})
	objMap2, ok2 := prop2.(map[string]interface{})
	if !ok1 || !ok2 {
		return false
	}
	
	required1, _ := v.getRequiredFields(objMap1)
	required2, _ := v.getRequiredFields(objMap2)
	
	// Проверяем что множества required полей идентичны
	if !v.areRequiredFieldsSetsIdentical(required1, required2) {
		return false
	}
	
	// Проверяем что все required поля имеют совместимые типы
	return v.areAllObjectFieldsCompatible(required1, props1, props2, spec1, spec2)
}

// areAllObjectFieldsCompatible проверяет совместимость всех полей объекта
func (v *ChannelValidator) areAllObjectFieldsCompatible(requiredFields []string, props1, props2 map[string]interface{}, spec1, spec2 *parser.AsyncAPISpec) bool {
	for _, field := range requiredFields {
		if !v.areObjectFieldsExistInBothSchemas(field, props1, props2) {
			return false
		}
		
		// Рекурсивно проверяем совместимость свойств
		if !v.arePropertyTypesCompatible(props1[field], props2[field], spec1, spec2) {
			return false
		}
	}
	return true
}

// areObjectFieldsExistInBothSchemas проверяет существование поля в обеих схемах
func (v *ChannelValidator) areObjectFieldsExistInBothSchemas(field string, props1, props2 map[string]interface{}) bool {
	_, exists1 := props1[field]
	_, exists2 := props2[field]
	return exists1 && exists2
}

// getObjectProperties извлекает properties из object свойства
func (v *ChannelValidator) getObjectProperties(prop interface{}) map[string]interface{} {
	if propMap, ok := prop.(map[string]interface{}); ok {
		if properties, exists := propMap["properties"]; exists {
			if propsMap, ok := properties.(map[string]interface{}); ok {
				return propsMap
			}
		}
	}
	return make(map[string]interface{})
}