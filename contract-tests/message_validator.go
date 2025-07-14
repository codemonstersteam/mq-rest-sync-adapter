package main

import (
	"fmt"
)

// MessageValidator валидирует структуру сообщений между потребителем и провайдером
type MessageValidator struct {
	config     *Config
	specLoader *SpecLoader
}

// NewMessageValidator создает новый валидатор сообщений
func NewMessageValidator(config *Config) *MessageValidator {
	return &MessageValidator{
		config:     config,
		specLoader: NewSpecLoader(config),
	}
}

// NewMessageValidatorWithSpecLoader создает новый валидатор сообщений с SpecLoader
func NewMessageValidatorWithSpecLoader(config *Config, specLoader *SpecLoader) *MessageValidator {
	return &MessageValidator{
		config:     config,
		specLoader: specLoader,
	}
}

// ValidatedChannel содержит информацию о валидном канале
type ValidatedChannel struct {
	ConsumerChannelName string
	ProviderChannelName string
	Address             string
	Protocol            string
}

// ValidatedMessage информация о валидированном сообщении
type ValidatedMessage struct {
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Direction   string `json:"direction"` // "send" или "receive"
	Compatible  bool   `json:"compatible"` // Совместимо ли с соответствующим сообщением
	Channel     string `json:"channel"`    // Имя канала для этого сообщения
}

// MessageValidationResult результат валидации сообщений
type MessageValidationResult struct {
	IsValid          bool                     `json:"is_valid"`
	ConsumerName     string                   `json:"consumer_name"`
	ProviderName     string                   `json:"provider_name"`
	ValidatedChannels []ValidatedChannel      `json:"validated_channels"`
	MessageErrors    []MessageValidationError `json:"message_errors"`
	// Детальная информация о сообщениях (плоская структура)
	ConsumerOutbound []ValidatedMessage `json:"consumer_outbound"` // Исходящие от потребителя
	ProviderInbound  []ValidatedMessage `json:"provider_inbound"`  // Входящие к провайдеру
	ProviderOutbound []ValidatedMessage `json:"provider_outbound"` // Исходящие от провайдера  
	ConsumerInbound  []ValidatedMessage `json:"consumer_inbound"`  // Входящие к потребителю
}

// MessageValidationError ошибка валидации сообщения
type MessageValidationError struct {
	Type        string `json:"type"`
	Channel     string `json:"channel"`
	Message     string `json:"message"`
	Field       string `json:"field"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

// MessageDirection направление сообщения
type MessageDirection struct {
	MessageRef AsyncAPIMessageRef
	Action     string // "send" или "receive"
}

// MessageField представляет поле сообщения
type MessageField struct {
	Name        string
	Type        string
	Description string
	Required    bool
}

// ValidateMessages валидирует структуру сообщений для валидных каналов
func (mv *MessageValidator) ValidateMessages(validatedChannels []ValidatedChannel) (*MessageValidationResult, error) {
	result := &MessageValidationResult{
		IsValid:           true,
		ConsumerName:      mv.config.ContractTests.Consumer.Name,
		ProviderName:      mv.config.ContractTests.Provider.Name,
		ValidatedChannels: validatedChannels,
		MessageErrors:     []MessageValidationError{},
		ConsumerOutbound:  []ValidatedMessage{},
		ProviderInbound:   []ValidatedMessage{},
		ProviderOutbound:  []ValidatedMessage{},
		ConsumerInbound:   []ValidatedMessage{},
	}

	// Загружаем спецификации через SpecLoader
	consumerSpec, err := mv.specLoader.GetConsumerSpec()
	if err != nil {
		return nil, fmt.Errorf("failed to load consumer spec: %v", err)
	}

	providerSpec, err := mv.specLoader.GetProviderSpec()
	if err != nil {
		return nil, fmt.Errorf("failed to load provider spec: %v", err)
	}

	// Валидируем каждый канал
	for _, channel := range validatedChannels {
		err := mv.validateChannelMessages(channel, consumerSpec, providerSpec, result)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// validateChannelMessages валидирует сообщения конкретного канала
func (mv *MessageValidator) validateChannelMessages(
	channel ValidatedChannel,
	consumerSpec, providerSpec *AsyncAPISpec,
	result *MessageValidationResult,
) error {
	// Находим исходящие сообщения потребителя (то что потребитель отправляет)
	consumerOutbound := mv.findOutboundMessages(channel.ConsumerChannelName, consumerSpec)
	result.ConsumerOutbound = append(result.ConsumerOutbound, mv.convertToValidatedMessages(consumerOutbound, "send", channel.ConsumerChannelName, consumerSpec)...)
	
	// Находим входящие сообщения провайдера (то что провайдер принимает)
	providerInbound := mv.findInboundMessages(channel.ProviderChannelName, providerSpec)
	result.ProviderInbound = append(result.ProviderInbound, mv.convertToValidatedMessages(providerInbound, "receive", channel.ProviderChannelName, providerSpec)...)
	
	// Валидируем совместимость: исходящие потребителя должны совпадать с входящими провайдера
	mv.validateMessageCompatibility(channel, consumerOutbound, providerInbound, "outbound->inbound", consumerSpec, providerSpec, result)
	
	// Находим входящие сообщения потребителя (то что потребитель ожидает получить)
	consumerInbound := mv.findInboundMessages(channel.ConsumerChannelName, consumerSpec)
	result.ConsumerInbound = append(result.ConsumerInbound, mv.convertToValidatedMessages(consumerInbound, "receive", channel.ConsumerChannelName, consumerSpec)...)
	
	// Находим исходящие сообщения провайдера (то что провайдер отправляет)
	providerOutbound := mv.findOutboundMessages(channel.ProviderChannelName, providerSpec)
	result.ProviderOutbound = append(result.ProviderOutbound, mv.convertToValidatedMessages(providerOutbound, "send", channel.ProviderChannelName, providerSpec)...)
	
	// Валидируем совместимость: исходящие провайдера должны совпадать с входящими потребителя
	mv.validateMessageCompatibility(channel, providerOutbound, consumerInbound, "inbound<-outbound", providerSpec, consumerSpec, result)
	
	// Валидация уже выполнена выше для всех направлений сообщений

	return nil
}

// findOutboundMessages находит исходящие сообщения для канала (action: send)
func (mv *MessageValidator) findOutboundMessages(channelName string, spec *AsyncAPISpec) []MessageDirection {
	var outbound []MessageDirection
	
	if spec.Operations == nil {
		return outbound
	}
	
	for _, operation := range spec.Operations {
		if operation.Action == "send" && mv.isOperationForChannel(operation, channelName) {
			// Для операций send исходящие - это основные сообщения операции (не reply)
			if channel, exists := spec.Channels[channelName]; exists {
				// Получаем все сообщения канала
				allMessages := channel.Messages
				
				// Исключаем сообщения из reply (если есть)
				replyMessages := make(map[string]bool)
				if operation.Reply != nil && len(operation.Reply.Messages) > 0 {
					for _, replyMsg := range operation.Reply.Messages {
						replyMessages[replyMsg.Ref] = true
					}
				}
				
				// Добавляем только сообщения, которые НЕ являются reply
				for _, msgRef := range allMessages {
					if !replyMessages[msgRef.Ref] {
						outbound = append(outbound, MessageDirection{
							MessageRef: msgRef,
							Action:     "send",
						})
					}
				}
			}
		}
	}
	
	return outbound
}

// findInboundMessages находит входящие сообщения для канала (action: receive или reply от send)
func (mv *MessageValidator) findInboundMessages(channelName string, spec *AsyncAPISpec) []MessageDirection {
	var inbound []MessageDirection
	
	if spec.Operations == nil {
		return inbound
	}
	
	for _, operation := range spec.Operations {
		if operation.Action == "receive" && mv.isOperationForChannel(operation, channelName) {
			// Для операций receive входящие - это все сообщения канала
			if channel, exists := spec.Channels[channelName]; exists {
				for _, msgRef := range channel.Messages {
					inbound = append(inbound, MessageDirection{
						MessageRef: msgRef,
						Action:     "receive",
					})
				}
			}
		} else if operation.Action == "send" && mv.isOperationForChannel(operation, channelName) {
			// Для операций send входящие - это сообщения из reply
			if operation.Reply != nil && len(operation.Reply.Messages) > 0 {
				for _, replyMsg := range operation.Reply.Messages {
					inbound = append(inbound, MessageDirection{
						MessageRef: replyMsg,
						Action:     "receive", // reply messages являются входящими
					})
				}
			}
		}
	}
	
	return inbound
}

// isOperationForChannel проверяет, относится ли операция к указанному каналу
func (mv *MessageValidator) isOperationForChannel(operation AsyncAPIOperation, channelName string) bool {
	if operation.Channel.Ref == "" {
		return false
	}
	
	// Извлекаем имя канала из ссылки (например, "#/channels/restGetBalanceRequest")
	const prefix = "#/channels/"
	if len(operation.Channel.Ref) > len(prefix) && operation.Channel.Ref[:len(prefix)] == prefix {
		operationChannelName := operation.Channel.Ref[len(prefix):]
		return operationChannelName == channelName
	}
	
	return false
}

// validateMessageCompatibility валидирует совместимость сообщений между двумя направлениями
func (mv *MessageValidator) validateMessageCompatibility(
	channel ValidatedChannel,
	sourceMessages, targetMessages []MessageDirection,
	direction string,
	sourceSpec, targetSpec *AsyncAPISpec,
	result *MessageValidationResult,
) {
	if len(sourceMessages) == 0 {
		mv.addMessageError(result, MessageValidationError{
			Type:        "no_source_messages",
			Channel:     channel.ConsumerChannelName,
			Message:     "",
			Field:       fmt.Sprintf("channels.%s.messages", channel.ConsumerChannelName),
			Description: fmt.Sprintf("Не найдены сообщения направления %s в канале", direction),
			Severity:    "medium",
		})
		return
	}
	
	if len(targetMessages) == 0 {
		mv.addMessageError(result, MessageValidationError{
			Type:        "no_target_messages",
			Channel:     channel.ConsumerChannelName,
			Message:     "",
			Field:       fmt.Sprintf("channels.%s.messages", channel.ProviderChannelName),
			Description: fmt.Sprintf("Не найдены целевые сообщения направления %s в канале", direction),
			Severity:    "high",
		})
		return
	}
	
	// Используем универсальный метод сравнения сообщений
	for _, sourceMsg := range sourceMessages {
		sourceMessage := mv.resolveMessageRef(sourceMsg.MessageRef, sourceSpec)
		if sourceMessage == nil {
			continue
		}
		
		compatible := false
		for _, targetMsg := range targetMessages {
			targetMessage := mv.resolveMessageRef(targetMsg.MessageRef, targetSpec)
			if targetMessage == nil {
				continue
			}
			
			if mv.compareMessageStructuresWithSpecs(sourceMessage, targetMessage, sourceSpec, targetSpec) {
				compatible = true
				break
			}
		}
		
		if !compatible {
			mv.addMessageError(result, MessageValidationError{
				Type:        "message_structure_incompatible",
				Channel:     channel.ConsumerChannelName,
				Message:     sourceMessage.Name,
				Field:       fmt.Sprintf("messages.%s.payload", sourceMessage.Name),
				Description: fmt.Sprintf("Структура сообщения %s не совместима с целевыми сообщениями направления %s", sourceMessage.Name, direction),
				Severity:    "high",
			})
		}
	}
}

// compareMessageStructuresWithSpecs сравнивает две структуры сообщений с доступом к спецификациям
func (mv *MessageValidator) compareMessageStructuresWithSpecs(source, target *AsyncAPIMessage, sourceSpec, targetSpec *AsyncAPISpec) bool {
	// Проверяем ContentType
	if source.ContentType != target.ContentType {
		return false
	}
	
	// Проверяем совместимость headers
	if !mv.compareMessageHeaders(source, target) {
		return false
	}
	
	// Проверяем совместимость payload с доступом к спецификациям
	if !mv.compareMessagePayloadWithSpecs(source, target, sourceSpec, targetSpec) {
		return false
	}
	
	return true
}

// compareMessageStructures сравнивает две структуры сообщений (deprecated)
func (mv *MessageValidator) compareMessageStructures(source, target *AsyncAPIMessage) bool {
	return mv.compareMessageStructuresWithSpecs(source, target, nil, nil)
}

// compareMessageHeaders сравнивает заголовки сообщений
func (mv *MessageValidator) compareMessageHeaders(source, target *AsyncAPIMessage) bool {
	// Базовая проверка наличия headers
	if source.Headers == nil && target.Headers == nil {
		return true
	}
	
	if source.Headers == nil || target.Headers == nil {
		return false
	}
	
	// Здесь должна быть более детальная проверка схем headers
	return true
}

// compareMessagePayload сравнивает payload сообщений
func (mv *MessageValidator) compareMessagePayload(source, target *AsyncAPIMessage) bool {
	// Базовая проверка наличия payload
	if source.Payload == nil && target.Payload == nil {
		return true
	}
	
	if source.Payload == nil || target.Payload == nil {
		return false
	}
	
	// Используем существующие методы для извлечения полей
	sourceFields, err := mv.ExtractMessagePayloadFields(source, nil)
	if err != nil {
		return false
	}
	
	targetFields, err := mv.ExtractMessagePayloadFields(target, nil)
	if err != nil {
		return false
	}
	
	// Создаем map для быстрого поиска
	sourceFieldMap := make(map[string]MessageField)
	for _, field := range sourceFields {
		sourceFieldMap[field.Name] = field
	}
	
	targetFieldMap := make(map[string]MessageField)
	for _, field := range targetFields {
		targetFieldMap[field.Name] = field
	}
	
	// Проверяем, что все обязательные поля source есть в target
	for fieldName, sourceField := range sourceFieldMap {
		if sourceField.Required {
			if targetField, exists := targetFieldMap[fieldName]; !exists {
				return false // Обязательное поле отсутствует
			} else if targetField.Type != sourceField.Type {
				return false // Типы полей не совпадают
			}
		}
	}
	
	return true
}

// compareMessagePayloadWithSpecs сравнивает payload сообщений с доступом к спецификациям
func (mv *MessageValidator) compareMessagePayloadWithSpecs(source, target *AsyncAPIMessage, sourceSpec, targetSpec *AsyncAPISpec) bool {
	// Базовая проверка наличия payload
	if source.Payload == nil && target.Payload == nil {
		return true
	}
	
	if source.Payload == nil || target.Payload == nil {
		return false
	}
	
	// Используем существующие методы для извлечения полей
	sourceFields, err := mv.ExtractMessagePayloadFields(source, sourceSpec)
	if err != nil {
		return false
	}
	
	targetFields, err := mv.ExtractMessagePayloadFields(target, targetSpec)
	if err != nil {
		return false
	}
	
	// Используем существующий метод детального сравнения
	compatible, _, err := mv.ValidatePayloadFieldsCompatibility(sourceFields, targetFields)
	if err != nil {
		return false
	}
	
	return compatible
}

// areMessagesStructurallyCompatible проверяет структурную совместимость сообщений (deprecated)
func (mv *MessageValidator) areMessagesStructurallyCompatible(source, target *AsyncAPIMessage) bool {
	return mv.compareMessageStructures(source, target)
}

// resolveMessageRef разрешает ссылку на сообщение
func (mv *MessageValidator) resolveMessageRef(msgRef AsyncAPIMessageRef, spec *AsyncAPISpec) *AsyncAPIMessage {
	if msgRef.Ref != "" {
		// Извлекаем имя компонента из ссылки (например, "#/components/messages/psGetBalanceRequest")
		componentName := mv.specLoader.GetRefExtractor().ExtractComponentName(msgRef.Ref)
		if componentName != "" && spec.Components.Messages != nil {
			if msg, exists := spec.Components.Messages[componentName]; exists {
				return &msg
			}
		}
	}
	return nil
}

// extractComponentName извлекает имя компонента из ссылки (обертка для обратной совместимости)
func (mv *MessageValidator) extractComponentName(ref string) string {
	return mv.specLoader.GetRefExtractor().ExtractComponentName(ref)
}

// addMessageError добавляет ошибку валидации сообщения
func (mv *MessageValidator) addMessageError(result *MessageValidationResult, error MessageValidationError) {
	result.MessageErrors = append(result.MessageErrors, error)
	// Только ошибки высокого уровня делают контракт несовместимым
	if error.Severity == "high" {
		result.IsValid = false
	}
}

// loadConsumerSpec загружает спецификацию потребителя (обертка для обратной совместимости)
func (mv *MessageValidator) loadConsumerSpec() (*AsyncAPISpec, error) {
	return mv.specLoader.GetConsumerSpec()
}

// loadProviderSpec загружает спецификацию провайдера (обертка для обратной совместимости)
func (mv *MessageValidator) loadProviderSpec() (*AsyncAPISpec, error) {
	return mv.specLoader.GetProviderSpec()
}

// convertToValidatedMessages преобразует MessageDirection в ValidatedMessage
func (mv *MessageValidator) convertToValidatedMessages(messages []MessageDirection, direction, channelName string, spec *AsyncAPISpec) []ValidatedMessage {
	var validatedMessages []ValidatedMessage
	
	for _, msg := range messages {
		messageName := mv.specLoader.GetRefExtractor().ExtractMessageName(msg.MessageRef.Ref)
		contentType := mv.extractMessageContentType(msg.MessageRef, spec)
		
		validatedMessages = append(validatedMessages, ValidatedMessage{
			Name:        messageName,
			ContentType: contentType,
			Direction:   direction,
			Compatible:  true, // Будет обновлено при валидации совместимости
			Channel:     channelName,
		})
	}
	
	return validatedMessages
}

// extractMessageContentType извлекает Content-Type сообщения
func (mv *MessageValidator) extractMessageContentType(messageRef AsyncAPIMessageRef, spec *AsyncAPISpec) string {
	if messageRef.Ref != "" {
		// Извлекаем имя сообщения из ссылки
		messageName := mv.specLoader.GetRefExtractor().ExtractMessageName(messageRef.Ref)
		if message, exists := spec.Components.Messages[messageName]; exists {
			if message.ContentType != "" {
				return message.ContentType
			}
		}
	}
	return "application/json" // По умолчанию
}

// FindConsumerOutboundMessage находит исходящее сообщение (send) от потребителя для указанного канала
func (mv *MessageValidator) FindConsumerOutboundMessage(channelName string, consumerSpec *AsyncAPISpec) (*AsyncAPIMessage, error) {
	// Ищем операцию с action: send для данного канала
	var sendOperation *AsyncAPIOperation
	channelRef := fmt.Sprintf("#/channels/%s", channelName)
	
	for _, operation := range consumerSpec.Operations {
		if operation.Action == "send" && operation.Channel.Ref == channelRef {
			sendOperation = &operation
			break
		}
	}
	
	if sendOperation == nil {
		return nil, fmt.Errorf("no send operation found for channel %s", channelName)
	}
	
	// Получаем канал
	channel, exists := consumerSpec.Channels[channelName]
	if !exists {
		return nil, fmt.Errorf("channel %s not found in spec", channelName)
	}
	
	// Получаем список reply сообщений из операции (это исходящие сообщения провайдера)
	replyMessageRefs := make(map[string]bool)
	if sendOperation.Reply != nil && sendOperation.Reply.Messages != nil {
		for _, replyMsgRef := range sendOperation.Reply.Messages {
			// Извлекаем имя сообщения из ссылки
			if msgName := mv.specLoader.GetRefExtractor().ExtractMessageName(replyMsgRef.Ref); msgName != "" {
				replyMessageRefs[msgName] = true
			}
		}
	}
	
	// Ищем исходящее сообщение потребителя - это сообщение в канале, которое НЕ является reply
	for _, messageRef := range channel.Messages {
		// Извлекаем имя сообщения из ссылки messageRef
		msgRefName := mv.specLoader.GetRefExtractor().ExtractMessageName(messageRef.Ref)
		
		// Если это не reply сообщение - значит это исходящее сообщение потребителя
		if !replyMessageRefs[msgRefName] {
			message := mv.resolveMessageRef(messageRef, consumerSpec)
			if message != nil {
				return message, nil
			}
		}
	}
	
	// Если не нашли через исключение reply, берем первое доступное
	for _, messageRef := range channel.Messages {
		message := mv.resolveMessageRef(messageRef, consumerSpec)
		if message != nil {
			return message, nil
		}
	}
	
	return nil, fmt.Errorf("no outbound message found for channel %s", channelName)
}

// ExtractMessageHeaderFields извлекает поля заголовков из сообщения
func (mv *MessageValidator) ExtractMessageHeaderFields(message *AsyncAPIMessage, spec *AsyncAPISpec) ([]MessageField, error) {
	var fields []MessageField
	
	if message.Headers == nil {
		return fields, nil
	}
	
	// Разрешаем ссылку на схему headers
	if refValue, ok := message.Headers["$ref"].(string); ok {
		// Извлекаем имя схемы из ссылки
		schemaName := mv.specLoader.GetRefExtractor().ExtractSchemaName(refValue)
		if schemaName != "" && spec.Components.Schemas != nil {
			if schema, exists := spec.Components.Schemas[schemaName]; exists {
				return mv.extractFieldsFromSchema(schema, schemaName)
			}
		}
	}
	
	// Если это прямое определение схемы
	return mv.extractFieldsFromSchema(message.Headers, "headers")
}

// ExtractMessagePayloadFields извлекает поля payload из сообщения
func (mv *MessageValidator) ExtractMessagePayloadFields(message *AsyncAPIMessage, spec *AsyncAPISpec) ([]MessageField, error) {
	var fields []MessageField
	
	if message.Payload == nil {
		return fields, nil
	}
	
	// Разрешаем ссылку на схему payload
	if refValue, ok := message.Payload["$ref"].(string); ok {
		// Извлекаем имя схемы из ссылки
		schemaName := mv.specLoader.GetRefExtractor().ExtractSchemaName(refValue)
		if schemaName != "" && spec.Components.Schemas != nil {
			if schema, exists := spec.Components.Schemas[schemaName]; exists {
				return mv.extractFieldsFromSchema(schema, schemaName)
			}
		}
	}
	
	// Если это прямое определение схемы
	return mv.extractFieldsFromSchema(message.Payload, "payload")
}

// extractSchemaName извлекает имя схемы из ссылки (обертка для обратной совместимости)
func (mv *MessageValidator) extractSchemaName(ref string) string {
	return mv.specLoader.GetRefExtractor().ExtractSchemaName(ref)
}

// extractMessageName извлекает имя сообщения из ссылки (обертка для обратной совместимости)
func (mv *MessageValidator) extractMessageName(ref string) string {
	return mv.specLoader.GetRefExtractor().ExtractMessageName(ref)
}

// extractFieldsFromSchema извлекает поля из схемы JSON Schema
func (mv *MessageValidator) extractFieldsFromSchema(schema interface{}, schemaName string) ([]MessageField, error) {
	var fields []MessageField
	
	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		return fields, fmt.Errorf("schema %s is not a valid object", schemaName)
	}
	
	// Получаем properties (поля объекта)
	properties, ok := schemaMap["properties"].(map[string]interface{})
	if !ok {
		return fields, nil // Нет свойств
	}
	
	// Получаем список обязательных полей
	requiredFields := make(map[string]bool)
	if required, ok := schemaMap["required"].([]interface{}); ok {
		for _, field := range required {
			if fieldName, ok := field.(string); ok {
				requiredFields[fieldName] = true
			}
		}
	}
	
	// Извлекаем информацию о каждом поле
	for fieldName, fieldDef := range properties {
		field := MessageField{
			Name:     fieldName,
			Required: requiredFields[fieldName],
		}
		
		if fieldDefMap, ok := fieldDef.(map[string]interface{}); ok {
			// Извлекаем тип поля
			if fieldType, ok := fieldDefMap["type"].(string); ok {
				field.Type = fieldType
			}
			
			// Извлекаем описание поля
			if description, ok := fieldDefMap["description"].(string); ok {
				field.Description = description
			}
		}
		
		fields = append(fields, field)
	}
	
	return fields, nil
}

// FindProviderInboundMessage находит входящее сообщение (receive) у провайдера для указанного канала
func (mv *MessageValidator) FindProviderInboundMessage(channelName string, providerSpec *AsyncAPISpec) (*AsyncAPIMessage, error) {
	// Ищем операцию с action: receive для данного канала
	var receiveOperation *AsyncAPIOperation
	channelRef := fmt.Sprintf("#/channels/%s", channelName)
	
	for _, operation := range providerSpec.Operations {
		if operation.Action == "receive" && operation.Channel.Ref == channelRef {
			receiveOperation = &operation
			break
		}
	}
	
	if receiveOperation == nil {
		return nil, fmt.Errorf("no receive operation found for channel %s", channelName)
	}
	
	// Получаем канал
	channel, exists := providerSpec.Channels[channelName]
	if !exists {
		return nil, fmt.Errorf("channel %s not found in spec", channelName)
	}
	
	// Если есть reply в операции receive - исключаем reply сообщения
	if receiveOperation.Reply != nil && receiveOperation.Reply.Messages != nil {
		// Получаем список reply сообщений (это исходящие сообщения провайдера)
		replyMessageRefs := make(map[string]bool)
		for _, replyMsgRef := range receiveOperation.Reply.Messages {
			if msgName := mv.specLoader.GetRefExtractor().ExtractMessageName(replyMsgRef.Ref); msgName != "" {
				replyMessageRefs[msgName] = true
			}
		}
		
		// Ищем входящее сообщение провайдера - это сообщение в канале, которое НЕ является reply
		for _, messageRef := range channel.Messages {
			msgRefName := mv.specLoader.GetRefExtractor().ExtractMessageName(messageRef.Ref)
			
			// Если это не reply сообщение - значит это входящее сообщение провайдера
			if !replyMessageRefs[msgRefName] {
				message := mv.resolveMessageRef(messageRef, providerSpec)
				if message != nil {
					return message, nil
				}
			}
		}
	} else {
		// Если нет reply - операция receive указывает, что провайдер принимает сообщения
		// Берем любое сообщение из канала (обычно первое доступное)
		for _, messageRef := range channel.Messages {
			message := mv.resolveMessageRef(messageRef, providerSpec)
			if message != nil {
				return message, nil
			}
		}
	}
	
	return nil, fmt.Errorf("no inbound message found for channel %s", channelName)
}

// MessageCompatibilityError представляет ошибку совместимости сообщений
type MessageCompatibilityError struct {
	Type        string
	Field       string
	Description string
	Severity    string // "high", "medium", "low", "info"
}

// ValidateMessageCompatibility валидирует совместимость между исходящим сообщением потребителя и входящим сообщением провайдера
func (mv *MessageValidator) ValidateMessageCompatibility(
	consumerMessage *AsyncAPIMessage, consumerSpec *AsyncAPISpec,
	providerMessage *AsyncAPIMessage, providerSpec *AsyncAPISpec,
) (bool, []MessageCompatibilityError, error) {
	var errors []MessageCompatibilityError
	isCompatible := true

	// 1. Проверяем ContentType
	if compatError := mv.validateContentTypeCompatibility(consumerMessage, providerMessage); compatError != nil {
		errors = append(errors, *compatError)
		isCompatible = false
	}

	// 2. Валидируем совместимость полей payload
	payloadCompatible, payloadErrors, err := mv.validatePayloadCompatibility(consumerMessage, consumerSpec, providerMessage, providerSpec)
	if err != nil {
		return false, errors, err
	}
	if !payloadCompatible {
		isCompatible = false
	}
	errors = append(errors, payloadErrors...)

	// 3. Валидируем совместимость заголовков
	headerCompatible, headerErrors, err := mv.validateHeaderCompatibility(consumerMessage, consumerSpec, providerMessage, providerSpec)
	if err != nil {
		return false, errors, err
	}
	if !headerCompatible {
		isCompatible = false
	}
	errors = append(errors, headerErrors...)

	return isCompatible, errors, nil
}

// validateContentTypeCompatibility проверяет совместимость ContentType
func (mv *MessageValidator) validateContentTypeCompatibility(consumerMessage, providerMessage *AsyncAPIMessage) *MessageCompatibilityError {
	if consumerMessage.ContentType != providerMessage.ContentType {
		return &MessageCompatibilityError{
			Type:        "content_type_mismatch",
			Field:       "contentType",
			Description: fmt.Sprintf("ContentType не совпадает: потребитель '%s', провайдер '%s'", consumerMessage.ContentType, providerMessage.ContentType),
			Severity:    "high",
		}
	}
	return nil
}

// validatePayloadCompatibility валидирует совместимость payload
func (mv *MessageValidator) validatePayloadCompatibility(
	consumerMessage *AsyncAPIMessage, consumerSpec *AsyncAPISpec,
	providerMessage *AsyncAPIMessage, providerSpec *AsyncAPISpec,
) (bool, []MessageCompatibilityError, error) {
	consumerPayloadFields, err := mv.ExtractMessagePayloadFields(consumerMessage, consumerSpec)
	if err != nil {
		return false, nil, fmt.Errorf("failed to extract consumer payload fields: %v", err)
	}

	providerPayloadFields, err := mv.ExtractMessagePayloadFields(providerMessage, providerSpec)
	if err != nil {
		return false, nil, fmt.Errorf("failed to extract provider payload fields: %v", err)
	}

	return mv.ValidatePayloadFieldsCompatibility(consumerPayloadFields, providerPayloadFields)
}

// validateHeaderCompatibility валидирует совместимость заголовков
func (mv *MessageValidator) validateHeaderCompatibility(
	consumerMessage *AsyncAPIMessage, consumerSpec *AsyncAPISpec,
	providerMessage *AsyncAPIMessage, providerSpec *AsyncAPISpec,
) (bool, []MessageCompatibilityError, error) {
	consumerHeaderFields, err := mv.ExtractMessageHeaderFields(consumerMessage, consumerSpec)
	if err != nil {
		return false, nil, fmt.Errorf("failed to extract consumer header fields: %v", err)
	}

	providerHeaderFields, err := mv.ExtractMessageHeaderFields(providerMessage, providerSpec)
	if err != nil {
		return false, nil, fmt.Errorf("failed to extract provider header fields: %v", err)
	}

	return mv.ValidateHeaderFieldsCompatibility(consumerHeaderFields, providerHeaderFields)
}

// ValidatePayloadFieldsCompatibility валидирует совместимость полей payload
func (mv *MessageValidator) ValidatePayloadFieldsCompatibility(
	consumerFields, providerFields []MessageField,
) (bool, []MessageCompatibilityError, error) {
	var errors []MessageCompatibilityError
	isCompatible := true

	// Создаем карты для быстрого поиска
	consumerFieldsMap, providerFieldsMap := mv.createFieldMaps(consumerFields, providerFields)

	// 1. Проверяем обязательные поля провайдера
	requiredFieldErrors := mv.validateRequiredProviderFields(providerFields, consumerFieldsMap)
	for _, err := range requiredFieldErrors {
		if err.Severity == "high" {
			isCompatible = false
		}
	}
	errors = append(errors, requiredFieldErrors...)

	// 2. Проверяем необязательные поля провайдера
	optionalFieldErrors := mv.validateOptionalProviderFields(providerFields, consumerFieldsMap)
	errors = append(errors, optionalFieldErrors...)

	// 3. Проверяем дополнительные поля потребителя
	extraConsumerFieldErrors := mv.validateExtraConsumerFields(consumerFields, providerFieldsMap)
	errors = append(errors, extraConsumerFieldErrors...)

	// 4. Логируем дополнительные поля провайдера
	extraProviderFieldErrors := mv.validateExtraProviderFields(providerFields, consumerFieldsMap)
	errors = append(errors, extraProviderFieldErrors...)

	return isCompatible, errors, nil
}

// ValidateHeaderFieldsCompatibility валидирует совместимость полей заголовков
func (mv *MessageValidator) ValidateHeaderFieldsCompatibility(
	consumerFields, providerFields []MessageField,
) (bool, []MessageCompatibilityError, error) {
	var errors []MessageCompatibilityError
	isCompatible := true

	// Создаем карты для быстрого поиска
	consumerFieldsMap, providerFieldsMap := mv.createFieldMaps(consumerFields, providerFields)

	// Проверяем обязательные заголовки провайдера
	requiredHeaderErrors := mv.validateRequiredProviderHeaders(providerFields, consumerFieldsMap)
	for _, err := range requiredHeaderErrors {
		if err.Severity == "high" {
			isCompatible = false
		}
	}
	errors = append(errors, requiredHeaderErrors...)

	// Проверяем дополнительные заголовки потребителя
	extraHeaderErrors := mv.validateExtraConsumerHeaders(consumerFields, providerFieldsMap)
	errors = append(errors, extraHeaderErrors...)

	return isCompatible, errors, nil
}

// Вспомогательные функции

// createFieldMaps создает карты полей для быстрого поиска
func (mv *MessageValidator) createFieldMaps(consumerFields, providerFields []MessageField) (map[string]MessageField, map[string]MessageField) {
	consumerFieldsMap := make(map[string]MessageField)
	providerFieldsMap := make(map[string]MessageField)

	for _, field := range consumerFields {
		consumerFieldsMap[field.Name] = field
	}
	for _, field := range providerFields {
		providerFieldsMap[field.Name] = field
	}

	return consumerFieldsMap, providerFieldsMap
}

// validateRequiredProviderFields проверяет обязательные поля провайдера
func (mv *MessageValidator) validateRequiredProviderFields(providerFields []MessageField, consumerFieldsMap map[string]MessageField) []MessageCompatibilityError {
	var errors []MessageCompatibilityError

	for _, providerField := range providerFields {
		if providerField.Required {
			consumerField, exists := consumerFieldsMap[providerField.Name]
			if !exists {
				errors = append(errors, MessageCompatibilityError{
					Type:        "required_field_missing_in_consumer",
					Field:       fmt.Sprintf("payload.%s", providerField.Name),
					Description: fmt.Sprintf("Обязательное поле '%s' требуется провайдером, но отсутствует у потребителя", providerField.Name),
					Severity:    "high",
				})
				continue
			}

			// Проверяем совместимость типов обязательных полей
			if consumerField.Type != providerField.Type {
				errors = append(errors, MessageCompatibilityError{
					Type:        "required_field_type_mismatch",
					Field:       fmt.Sprintf("payload.%s", providerField.Name),
					Description: fmt.Sprintf("Тип обязательного поля '%s' не совпадает: потребитель '%s', провайдер '%s'", providerField.Name, consumerField.Type, providerField.Type),
					Severity:    "high",
				})
			}
		}
	}

	return errors
}

// validateOptionalProviderFields проверяет необязательные поля провайдера
func (mv *MessageValidator) validateOptionalProviderFields(providerFields []MessageField, consumerFieldsMap map[string]MessageField) []MessageCompatibilityError {
	var errors []MessageCompatibilityError

	for _, providerField := range providerFields {
		if !providerField.Required {
			if consumerField, exists := consumerFieldsMap[providerField.Name]; exists {
				// Если потребитель отправляет необязательное поле, типы должны совпадать
				if consumerField.Type != providerField.Type {
					errors = append(errors, MessageCompatibilityError{
						Type:        "optional_field_type_mismatch",
						Field:       fmt.Sprintf("payload.%s", providerField.Name),
						Description: fmt.Sprintf("Тип необязательного поля '%s' не совпадает: потребитель '%s', провайдер '%s'", providerField.Name, consumerField.Type, providerField.Type),
						Severity:    "medium",
					})
				}
			}
		}
	}

	return errors
}

// validateExtraConsumerFields проверяет дополнительные поля потребителя
func (mv *MessageValidator) validateExtraConsumerFields(consumerFields []MessageField, providerFieldsMap map[string]MessageField) []MessageCompatibilityError {
	var errors []MessageCompatibilityError

	for _, consumerField := range consumerFields {
		if _, exists := providerFieldsMap[consumerField.Name]; !exists {
			errors = append(errors, MessageCompatibilityError{
				Type:        "field_extra_in_consumer",
				Field:       fmt.Sprintf("payload.%s", consumerField.Name),
				Description: fmt.Sprintf("Потребитель отправляет дополнительное поле '%s' (может быть проигнорировано провайдером)", consumerField.Name),
				Severity:    "low",
			})
		}
	}

	return errors
}

// validateExtraProviderFields логирует дополнительные поля провайдера
func (mv *MessageValidator) validateExtraProviderFields(providerFields []MessageField, consumerFieldsMap map[string]MessageField) []MessageCompatibilityError {
	var errors []MessageCompatibilityError

	for _, providerField := range providerFields {
		if _, exists := consumerFieldsMap[providerField.Name]; !exists {
			errors = append(errors, MessageCompatibilityError{
				Type:        "field_extra_in_provider",
				Field:       fmt.Sprintf("payload.%s", providerField.Name),
				Description: fmt.Sprintf("Провайдер имеет дополнительное поле '%s' (нормально - расширяемость API)", providerField.Name),
				Severity:    "info",
			})
		}
	}

	return errors
}

// validateRequiredProviderHeaders проверяет обязательные заголовки провайдера
func (mv *MessageValidator) validateRequiredProviderHeaders(providerFields []MessageField, consumerFieldsMap map[string]MessageField) []MessageCompatibilityError {
	var errors []MessageCompatibilityError

	for _, providerField := range providerFields {
		if providerField.Required {
			consumerField, exists := consumerFieldsMap[providerField.Name]
			if !exists {
				errors = append(errors, MessageCompatibilityError{
					Type:        "required_header_missing_in_consumer",
					Field:       fmt.Sprintf("headers.%s", providerField.Name),
					Description: fmt.Sprintf("Обязательный заголовок '%s' требуется провайдером, но отсутствует у потребителя", providerField.Name),
					Severity:    "high",
				})
				continue
			}

			// Проверяем тип обязательного заголовка
			if consumerField.Type != providerField.Type {
				errors = append(errors, MessageCompatibilityError{
					Type:        "required_header_type_mismatch",
					Field:       fmt.Sprintf("headers.%s", providerField.Name),
					Description: fmt.Sprintf("Тип обязательного заголовка '%s' не совпадает: потребитель '%s', провайдер '%s'", providerField.Name, consumerField.Type, providerField.Type),
					Severity:    "high",
				})
			}
		}
	}

	return errors
}

// validateExtraConsumerHeaders проверяет дополнительные заголовки потребителя
func (mv *MessageValidator) validateExtraConsumerHeaders(consumerFields []MessageField, providerFieldsMap map[string]MessageField) []MessageCompatibilityError {
	var errors []MessageCompatibilityError

	for _, consumerField := range consumerFields {
		if _, exists := providerFieldsMap[consumerField.Name]; !exists {
			errors = append(errors, MessageCompatibilityError{
				Type:        "header_extra_in_consumer",
				Field:       fmt.Sprintf("headers.%s", consumerField.Name),
				Description: fmt.Sprintf("Потребитель отправляет дополнительный заголовок '%s' (может быть проигнорирован провайдером)", consumerField.Name),
				Severity:    "low",
			})
		}
	}

	return errors
}

// FindProviderOutboundMessage находит исходящее сообщение (send) у провайдера для указанного канала
func (mv *MessageValidator) FindProviderOutboundMessage(channelName string, providerSpec *AsyncAPISpec) (*AsyncAPIMessage, error) {
	// Ищем операцию с action: send для данного канала
	var sendOperation *AsyncAPIOperation
	channelRef := fmt.Sprintf("#/channels/%s", channelName)
	
	for _, operation := range providerSpec.Operations {
		if operation.Action == "send" && operation.Channel.Ref == channelRef {
			sendOperation = &operation
			break
		}
	}
	
	if sendOperation == nil {
		return nil, fmt.Errorf("no send operation found for channel %s", channelName)
	}
	
	// Получаем канал
	channel, exists := providerSpec.Channels[channelName]
	if !exists {
		return nil, fmt.Errorf("channel %s not found in spec", channelName)
	}
	
	// Ищем исходящее сообщение провайдера
	for _, messageRef := range channel.Messages {
		message := mv.resolveMessageRef(messageRef, providerSpec)
		if message != nil {
			return message, nil
		}
	}
	
	return nil, fmt.Errorf("no outbound message found for channel %s", channelName)
}

// FindConsumerInboundMessage находит входящее сообщение (receive) у потребителя для указанного канала
func (mv *MessageValidator) FindConsumerInboundMessage(channelName string, consumerSpec *AsyncAPISpec) (*AsyncAPIMessage, error) {
	// Ищем операцию с action: receive для данного канала
	var receiveOperation *AsyncAPIOperation
	channelRef := fmt.Sprintf("#/channels/%s", channelName)
	
	for _, operation := range consumerSpec.Operations {
		if operation.Action == "receive" && operation.Channel.Ref == channelRef {
			receiveOperation = &operation
			break
		}
	}
	
	if receiveOperation == nil {
		return nil, fmt.Errorf("no receive operation found for channel %s", channelName)
	}
	
	// Получаем канал
	channel, exists := consumerSpec.Channels[channelName]
	if !exists {
		return nil, fmt.Errorf("channel %s not found in spec", channelName)
	}
	
	// Ищем входящее сообщение потребителя
	for _, messageRef := range channel.Messages {
		message := mv.resolveMessageRef(messageRef, consumerSpec)
		if message != nil {
			return message, nil
		}
	}
	
	return nil, fmt.Errorf("no inbound message found for channel %s", channelName)
}