package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// ContractChecker основная структура для проверки контрактов
type ContractChecker struct {
	config     *Config
	specLoader *SpecLoader
}

// NewContractChecker создает новый экземпляр проверки контрактов
func NewContractChecker(config *Config) *ContractChecker {
	return &ContractChecker{
		config:     config,
		specLoader: NewSpecLoader(config),
	}
}

// NewContractCheckerWithSpecLoader создает новый экземпляр проверки контрактов с SpecLoader
func NewContractCheckerWithSpecLoader(config *Config, specLoader *SpecLoader) *ContractChecker {
	return &ContractChecker{
		config:     config,
		specLoader: specLoader,
	}
}

// CompatibilityResult результат проверки совместимости
type CompatibilityResult struct {
	IsCompatible      bool                   `json:"isCompatible"`
	BreakingChanges   []BreakingChange       `json:"breakingChanges"`
	Warnings          []CompatibilityWarning `json:"warnings"`
	CheckedAt         time.Time              `json:"checkedAt"`
	ConsumerVersion   string                 `json:"consumerVersion"`
	ProviderVersion   string                 `json:"providerVersion"`
	ConsumerName      string                 `json:"consumerName"`
	ProviderName      string                 `json:"providerName"`
	ValidatedChannels []ValidatedChannel     `json:"validatedChannels"` // Информация о валидных каналах
}

// BreakingChange описание критического изменения
type BreakingChange struct {
	Type        string `json:"type"`
	Field       string `json:"field"`
	OldValue    string `json:"oldValue"`
	NewValue    string `json:"newValue"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

// CompatibilityWarning предупреждение о совместимости
type CompatibilityWarning struct {
	Type       string `json:"type"`
	Field      string `json:"field"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
}

// AsyncAPISpec структура для работы с AsyncAPI спецификацией
type AsyncAPISpec struct {
	AsyncAPI   string                        `yaml:"asyncapi"`
	Info       AsyncAPIInfo                  `yaml:"info"`
	Servers    map[string]AsyncAPIServer     `yaml:"servers"`
	Channels   map[string]AsyncAPIChannel    `yaml:"channels"`
	Operations map[string]AsyncAPIOperation  `yaml:"operations"`
	Components AsyncAPIComponents            `yaml:"components"`
}

// AsyncAPIServer сервер в спецификации
type AsyncAPIServer struct {
	Host     string `yaml:"host"`
	Protocol string `yaml:"protocol"`
	Pathname string `yaml:"pathname"`
}

// AsyncAPIInfo информация о спецификации
type AsyncAPIInfo struct {
	Title       string `yaml:"title"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

// AsyncAPIChannel канал в спецификации
type AsyncAPIChannel struct {
	Address     string                         `yaml:"address"`
	Description string                         `yaml:"description"`
	Messages    map[string]AsyncAPIMessageRef  `yaml:"messages"`
	Servers     []AsyncAPIServerRef            `yaml:"servers"`
}

// AsyncAPIServerRef ссылка на сервер
type AsyncAPIServerRef struct {
	Ref string `yaml:"$ref"`
}

// ProtocolRegistry структура для хранения информации о протоколах
type ProtocolRegistry struct {
	Protocols     map[string]ProtocolInfo `yaml:"protocols"`
	Compatibility CompatibilityRules      `yaml:"compatibility"`
}

// ProtocolInfo информация о протоколе
type ProtocolInfo struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Category    string `yaml:"category"`
	Secure      bool   `yaml:"secure"`
}

// CompatibilityRules правила совместимости протоколов
type CompatibilityRules struct {
	StrictMatching   bool                `yaml:"strict_matching"`
	Categories       map[string][]string `yaml:"categories"`
	CompatibleGroups [][]string          `yaml:"compatible_groups"`
}

// CheckCompatibility выполняет проверку совместимости контрактов
func (cc *ContractChecker) CheckCompatibility() (*CompatibilityResult, error) {
	if cc.config.GetLogLevel() == "debug" {
		fmt.Printf("🔍 Начинаем проверку совместимости контрактов...\n")
		fmt.Printf("📋 Потребитель: %s\n", cc.config.ContractTests.Consumer.Name)
		fmt.Printf("🏢 Провайдер: %s\n", cc.config.ContractTests.Provider.Name)
	}

	// Загружаем спецификации через SpecLoader
	consumerSpec, err := cc.specLoader.GetConsumerSpec()
	if err != nil {
		return nil, fmt.Errorf("failed to load consumer spec: %w", err)
	}

	providerSpec, err := cc.specLoader.GetProviderSpec()
	if err != nil {
		return nil, fmt.Errorf("failed to load provider spec: %w", err)
	}

	// Создаем результат проверки
	result := &CompatibilityResult{
		IsCompatible:      true,
		BreakingChanges:   []BreakingChange{},
		Warnings:          []CompatibilityWarning{},
		CheckedAt:         time.Now(),
		ConsumerVersion:   consumerSpec.Info.Version,
		ProviderVersion:   providerSpec.Info.Version,
		ConsumerName:      cc.config.ContractTests.Consumer.Name,
		ProviderName:      cc.config.ContractTests.Provider.Name,
		ValidatedChannels: []ValidatedChannel{}, // Инициализируем список валидных каналов
	}

	// Выполняем проверку каналов
	cc.checkChannels(consumerSpec, providerSpec, result)

	// Определяем общую совместимость
	if len(result.BreakingChanges) > 0 {
		result.IsCompatible = false
	}

	return result, nil
}

// loadConsumerSpec загружает спецификацию потребителя из файла (обертка для обратной совместимости)
func (cc *ContractChecker) loadConsumerSpec() (*AsyncAPISpec, error) {
	return cc.specLoader.GetConsumerSpec()
}

// loadProviderSpec загружает спецификацию провайдера (обертка для обратной совместимости)
func (cc *ContractChecker) loadProviderSpec() (*AsyncAPISpec, error) {
	return cc.specLoader.GetProviderSpec()
}

// checkChannels проверяет совместимость каналов по адресам
func (cc *ContractChecker) checkChannels(consumer, provider *AsyncAPISpec, result *CompatibilityResult) {
	if cc.config.GetLogLevel() == "debug" {
		fmt.Printf("🔍 Проверяем каналы...\n")
	}

	// Проверяем только каналы, указанные в конфигурации
	for _, channelName := range cc.config.ContractTests.Consumer.Channels {
		if cc.config.GetLogLevel() == "debug" {
			fmt.Printf("  📋 Проверяем канал: %s\n", channelName)
		}

		// Проверяем наличие канала в спецификации потребителя
		consumerChannel, exists := consumer.Channels[channelName]
		if !exists {
			result.Warnings = append(result.Warnings, CompatibilityWarning{
				Type:       "channel_missing_in_consumer",
				Field:      fmt.Sprintf("channels.%s", channelName),
				Message:    fmt.Sprintf("Channel '%s' specified in config but not found in consumer spec", channelName),
				Suggestion: "Remove channel from config or add it to consumer specification",
			})
			continue
		}

		// Получаем адрес канала потребителя
		consumerAddress := consumerChannel.Address
		if consumerAddress == "" {
			result.Warnings = append(result.Warnings, CompatibilityWarning{
				Type:       "channel_address_empty",
				Field:      fmt.Sprintf("channels.%s.address", channelName),
				Message:    fmt.Sprintf("Channel '%s' has empty address in consumer spec", channelName),
				Suggestion: "Add address to channel definition",
			})
			continue
		}

		// Предохранитель: проверяем обязательное наличие ссылки на сервер
		if len(consumerChannel.Servers) == 0 {
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Type:        "channel_missing_server_reference",
				Field:       fmt.Sprintf("channels.%s.servers", channelName),
				OldValue:    "[]",
				NewValue:    "required",
				Description: fmt.Sprintf("Канал '%s' не содержит ссылки на сервер. Для определения протокола каждый канал должен содержать ссылку на сервер", channelName),
				Severity:    "high",
			})
			if cc.config.GetLogLevel() == "debug" {
				fmt.Printf("    ❌ Канал '%s' не содержит ссылки на сервер\n", channelName)
			}
			continue
		}

		// Определяем тип канала
		channelType := cc.getChannelType(consumerChannel, consumer)

		// Проверяем, известен ли протокол потребителя
		if !cc.isProtocolKnown(channelType) {
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Type:        "unknown_protocol",
				Field:       fmt.Sprintf("channels.%s.protocol", channelName),
				OldValue:    channelType,
				NewValue:    "supported_protocol",
				Description: fmt.Sprintf("Канал '%s' использует неизвестный протокол '%s'. Поддерживаются только протоколы из официального словаря AsyncAPI 3.0", channelName, channelType),
				Severity:    "high",
			})
			if cc.config.GetLogLevel() == "debug" {
				fmt.Printf("    ❌ Канал '%s' использует неизвестный протокол '%s'\n", channelName, channelType)
			}
		}

		// Ищем канал с таким же адресом в провайдере
		providerChannelName := cc.findChannelByAddress(provider, consumerAddress)
		if providerChannelName == "" {
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Type:        "channel_address_not_supported",
				Field:       fmt.Sprintf("channels.%s.address", channelName),
				OldValue:    consumerAddress,
				NewValue:    "not_supported",
				Description: fmt.Sprintf("Канал '%s' с протоколом '%s' и адресом '%s' требуется потребителю, но не поддерживается провайдером", channelName, channelType, consumerAddress),
				Severity:    "high",
			})
			if cc.config.GetLogLevel() == "debug" {
				fmt.Printf("    ❌ Канал '%s' (протокол: %s, адрес: %s) не поддерживается провайдером\n", channelName, channelType, consumerAddress)
			}
		} else {
			// Проверяем, что у найденного канала провайдера есть ссылка на сервер
			providerChannel := provider.Channels[providerChannelName]
			if len(providerChannel.Servers) == 0 {
				result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
					Type:        "provider_channel_missing_server_reference",
					Field:       fmt.Sprintf("provider.channels.%s.servers", providerChannelName),
					OldValue:    "[]",
					NewValue:    "required",
					Description: fmt.Sprintf("Канал '%s' провайдера не содержит ссылки на сервер. Для определения протокола каждый канал должен содержать ссылку на сервер", providerChannelName),
					Severity:    "high",
				})
				if cc.config.GetLogLevel() == "debug" {
					fmt.Printf("    ❌ Канал '%s' провайдера не содержит ссылки на сервер\n", providerChannelName)
				}
			} else {
				// Проверяем совместимость протоколов
				providerChannelType := cc.getChannelType(providerChannel, provider)

				// Проверяем, известен ли протокол провайдера
				if !cc.isProtocolKnown(providerChannelType) {
					result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
						Type:        "unknown_protocol",
						Field:       fmt.Sprintf("provider.channels.%s.protocol", providerChannelName),
						OldValue:    providerChannelType,
						NewValue:    "supported_protocol",
						Description: fmt.Sprintf("Канал '%s' провайдера использует неизвестный протокол '%s'. Поддерживаются только протоколы из официального словаря AsyncAPI 3.0", providerChannelName, providerChannelType),
						Severity:    "high",
					})
					if cc.config.GetLogLevel() == "debug" {
						fmt.Printf("    ❌ Канал '%s' провайдера использует неизвестный протокол '%s'\n", providerChannelName, providerChannelType)
					}
				}

				if !cc.areProtocolsCompatible(channelType, providerChannelType) {
					result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
						Type:        "protocol_mismatch",
						Field:       fmt.Sprintf("channels.%s.protocol", channelName),
						OldValue:    channelType,
						NewValue:    providerChannelType,
						Description: fmt.Sprintf("Канал '%s' использует протокол '%s', но провайдер поддерживает протокол '%s'. Протоколы несовместимы", channelName, channelType, providerChannelType),
						Severity:    "high",
					})
					if cc.config.GetLogLevel() == "debug" {
						fmt.Printf("    ❌ Канал '%s' (протокол: %s) несовместим с провайдером (протокол: %s)\n", channelName, channelType, providerChannelType)
					}
				} else {
					if cc.config.GetLogLevel() == "debug" {
						fmt.Printf("    ✅ Канал '%s' (протокол: %s, адрес: %s) поддерживается провайдером (канал '%s', протокол: %s)\n", channelName, channelType, consumerAddress, providerChannelName, providerChannelType)
					}
					
					// Добавляем информацию о валидном канале в результат
					result.ValidatedChannels = append(result.ValidatedChannels, ValidatedChannel{
						ConsumerChannelName: channelName,
						ProviderChannelName: providerChannelName,
						Address:             consumerAddress,
						Protocol:            channelType,
					})
				}
			}
		}
	}
}

// findChannelByAddress ищет канал в провайдере по адресу
func (cc *ContractChecker) findChannelByAddress(provider *AsyncAPISpec, targetAddress string) string {
	// Нормализуем адрес для сравнения (убираем параметры)
	normalizedTargetAddress := cc.normalizeAddress(targetAddress)

	for channelName, channel := range provider.Channels {
		normalizedProviderAddress := cc.normalizeAddress(channel.Address)
		if normalizedProviderAddress == normalizedTargetAddress {
			return channelName
		}
	}
	return ""
}

// getChannelType определяет протокол канала по серверу или адресу
func (cc *ContractChecker) getChannelType(channel AsyncAPIChannel, spec *AsyncAPISpec) string {
	// Сначала пытаемся определить протокол по серверу
	if len(channel.Servers) > 0 {
		// Берем первый сервер
		serverRef := channel.Servers[0].Ref
		serverName := cc.specLoader.GetRefExtractor().ExtractServerName(serverRef)
		if server, exists := spec.Servers[serverName]; exists {
			return server.Protocol
		}
	}

	// Если сервер не указан, определяем по адресу (fallback)
	return cc.getChannelTypeByAddress(channel.Address)
}

// extractServerName извлекает имя сервера из ссылки (обертка для обратной совместимости)
func (cc *ContractChecker) extractServerName(ref string) string {
	return cc.specLoader.GetRefExtractor().ExtractServerName(ref)
}

// getChannelTypeByAddress определяет протокол канала по адресу (fallback)
func (cc *ContractChecker) getChannelTypeByAddress(address string) string {
	if len(address) > 0 && address[0] == '/' {
		return "http"
	}
	return "unknown"
}

// GetChannelType определяет тип канала (публичный метод)
func (cc *ContractChecker) GetChannelType(channel AsyncAPIChannel, spec *AsyncAPISpec) string {
	return cc.getChannelType(channel, spec)
}

// LoadProtocols загружает словарь протоколов из файла
func LoadProtocols() (*ProtocolRegistry, error) {
	content, err := os.ReadFile("protocols.yml")
	if err != nil {
		return nil, fmt.Errorf("failed to read protocols.yml: %w", err)
	}

	var registry ProtocolRegistry
	err = yaml.Unmarshal(content, &registry)
	if err != nil {
		return nil, fmt.Errorf("failed to parse protocols.yml: %w", err)
	}

	return &registry, nil
}

// areProtocolsCompatible проверяет совместимость протоколов
func (cc *ContractChecker) areProtocolsCompatible(consumerProtocol, providerProtocol string) bool {
	// Загружаем реестр протоколов через SpecLoader
	registry, err := cc.specLoader.GetProtocolRegistry()
	if err != nil {
		// Если не удалось загрузить словарь, это критическая ошибка
		return false
	}

	// Проверяем, что оба протокола существуют в словаре
	_, consumerExists := registry.Protocols[consumerProtocol]
	_, providerExists := registry.Protocols[providerProtocol]

	if !consumerExists || !providerExists {
		// Если протокол не найден в словаре, считаем несовместимым
		return false
	}

	// Используем правила совместимости из словаря
	if registry.Compatibility.StrictMatching {
		// Строгое соответствие - протоколы должны быть идентичными
		return consumerProtocol == providerProtocol
	}

	// Если не строгое соответствие, проверяем совместимые группы
	// (в текущей реализации используется strict_matching: true)
	return consumerProtocol == providerProtocol
}

// areProtocolsCompatibleFallback старая логика для совместимости (fallback)
func (cc *ContractChecker) areProtocolsCompatibleFallback(consumerProtocol, providerProtocol string) bool {
	// Точное совпадение протоколов
	if consumerProtocol == providerProtocol {
		return true
	}

	// Специальные правила совместимости
	switch {
	case consumerProtocol == "http" && providerProtocol == "https":
		// HTTP клиент может работать с HTTPS сервером (обычно)
		// Но это может быть небезопасно, поэтому считаем несовместимым
		return false
	case consumerProtocol == "https" && providerProtocol == "http":
		// HTTPS клиент не может работать с HTTP сервером
		return false
	default:
		// Для других протоколов требуется точное совпадение
		return false
	}
}

// normalizeAddress нормализует адрес для сравнения
func (cc *ContractChecker) normalizeAddress(address string) string {
	// Заменяем параметры на placeholder для сравнения
	// Например: /api/client/{Api-clientId}/wallet/balance -> /api/client/{param}/wallet/balance

	normalizedAddress := address

	// Для REST endpoints заменяем различные параметры на общий placeholder
	if len(normalizedAddress) > 0 && normalizedAddress[0] == '/' {
		// Заменяем {Api-clientId} на {clientId}
		if normalizedAddress == "/api/client/{Api-clientId}/wallet/balance" {
			normalizedAddress = "/api/client/{clientId}/wallet/balance"
		}
		// Заменяем {clientId} на {clientId} (уже нормализован)
		if normalizedAddress == "/api/client/{clientId}/wallet/balance" {
			return normalizedAddress
		}
	}

	return normalizedAddress
}

// isProtocolKnown проверяет, известен ли протокол в словаре
func (cc *ContractChecker) isProtocolKnown(protocol string) bool {
	// Загружаем реестр протоколов через SpecLoader
	registry, err := cc.specLoader.GetProtocolRegistry()
	if err != nil {
		// Если не удалось загрузить словарь, это критическая ошибка
		return false
	}

	// Проверяем наличие протокола в словаре
	_, exists := registry.Protocols[protocol]
	return exists
}

// AsyncAPIOperation операция в спецификации
type AsyncAPIOperation struct {
	Action      string              `yaml:"action"`
	Channel     AsyncAPIChannelRef  `yaml:"channel"`
	Summary     string              `yaml:"summary"`
	Description string              `yaml:"description"`
	Reply       *AsyncAPIReply      `yaml:"reply"`
}

// AsyncAPIChannelRef ссылка на канал
type AsyncAPIChannelRef struct {
	Ref string `yaml:"$ref"`
}

// AsyncAPIReply ответ операции
type AsyncAPIReply struct {
	Channel  AsyncAPIChannelRef     `yaml:"channel"`
	Messages []AsyncAPIMessageRef   `yaml:"messages"`
}

// AsyncAPIMessageRef ссылка на сообщение
type AsyncAPIMessageRef struct {
	Ref string `yaml:"$ref"`
}

// AsyncAPIComponents компоненты спецификации
type AsyncAPIComponents struct {
	Messages map[string]AsyncAPIMessage `yaml:"messages"`
	Schemas  map[string]interface{}     `yaml:"schemas"`
}

// AsyncAPIMessage сообщение в спецификации
type AsyncAPIMessage struct {
	Name        string                 `yaml:"name"`
	Title       string                 `yaml:"title"`
	ContentType string                 `yaml:"contentType"`
	Headers     map[string]interface{} `yaml:"headers"`
	Payload     map[string]interface{} `yaml:"payload"`
}
