package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

type ContractCompatibilityChecker struct {
	ConsumerSpecPath string
	ProviderGitURL   string
	Config           *Config
}

type Config struct {
	ContractTests struct {
		Consumer struct {
			SpecPath string `yaml:"spec_path"`
			Name     string `yaml:"name"`
		} `yaml:"consumer"`
		Provider struct {
			SpecURL string `yaml:"spec_url"`
			Name    string `yaml:"name"`
			ChannelRestrictions struct {
				CheckMode string `yaml:"check_mode"`
			} `yaml:"channel_restrictions"`
		} `yaml:"provider"`
		Settings struct {
			LogLevel         string `yaml:"log_level"`
			SaveJSONReport   bool   `yaml:"save_json_report"`
			JSONReportFile   string `yaml:"json_report_file"`
			Timeout          int    `yaml:"timeout"`
			IgnoreWarnings   bool   `yaml:"ignore_warnings"`
		} `yaml:"settings"`
	} `yaml:"contract_tests"`
}

type CompatibilityResult struct {
	IsCompatible     bool                `json:"isCompatible"`
	BreakingChanges  []BreakingChange    `json:"breakingChanges"`
	Warnings         []CompatibilityWarning `json:"warnings"`
	CheckedAt        time.Time           `json:"checkedAt"`
	ConsumerVersion  string              `json:"consumerVersion"`
	ProviderVersion  string              `json:"providerVersion"`
}

type BreakingChange struct {
	Type        string `json:"type"`
	Field       string `json:"field"`
	OldValue    string `json:"oldValue"`
	NewValue    string `json:"newValue"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

type CompatibilityWarning struct {
	Type        string `json:"type"`
	Field       string `json:"field"`
	Message     string `json:"message"`
	Suggestion  string `json:"suggestion"`
}

func NewContractCompatibilityChecker(consumerSpecPath, providerGitURL string) *ContractCompatibilityChecker {
	return &ContractCompatibilityChecker{
		ConsumerSpecPath: consumerSpecPath,
		ProviderGitURL:   providerGitURL,
	}
}

func NewContractCompatibilityCheckerFromConfig(configPath string) (*ContractCompatibilityChecker, error) {
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, err
	}

	return &ContractCompatibilityChecker{
		ConsumerSpecPath: config.ContractTests.Consumer.SpecPath,
		ProviderGitURL:   config.ContractTests.Provider.SpecURL,
		Config:          config,
	}, nil
}

func loadConfig(configPath string) (*Config, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func (c *ContractCompatibilityChecker) CheckCompatibility() (*CompatibilityResult, error) {
	// Загружаем спецификацию потребителя
	consumerSpec, err := c.loadConsumerSpec()
	if err != nil {
		return nil, fmt.Errorf("failed to load consumer spec: %w", err)
	}

	// Загружаем спецификацию поставщика из Git
	providerSpec, err := c.loadProviderSpecFromGit()
	if err != nil {
		return nil, fmt.Errorf("failed to load provider spec from git: %w", err)
	}

	// Проверяем совместимость
	result := &CompatibilityResult{
		IsCompatible:     true,
		BreakingChanges:  []BreakingChange{},
		Warnings:         []CompatibilityWarning{},
		CheckedAt:        time.Now(),
		ConsumerVersion:  consumerSpec.Get("info.version").String(),
		ProviderVersion:  providerSpec.Get("info.version").String(),
	}

	// Анализируем спецификации и выводим план проверки
	c.analyzeAndExplainCompatibilityPlan(consumerSpec, providerSpec)
	
	// Проверяем различные аспекты совместимости
	c.checkRESTEndpointCompatibility(consumerSpec, providerSpec, result)
	c.checkMQChannelCompatibility(consumerSpec, providerSpec, result)
	c.checkSchemaCompatibility(consumerSpec, providerSpec, result)

	// Если есть breaking changes, контракт несовместим
	if len(result.BreakingChanges) > 0 {
		result.IsCompatible = false
	}

	return result, nil
}

func (c *ContractCompatibilityChecker) loadConsumerSpec() (gjson.Result, error) {
	content, err := os.ReadFile(c.ConsumerSpecPath)
	if err != nil {
		return gjson.Result{}, err
	}
	return c.parseYAMLToJSON(content)
}

func (c *ContractCompatibilityChecker) loadProviderSpecFromGit() (gjson.Result, error) {
	resp, err := http.Get(c.ProviderGitURL)
	if err != nil {
		return gjson.Result{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return gjson.Result{}, fmt.Errorf("failed to fetch provider spec: status %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return gjson.Result{}, err
	}

	return c.parseYAMLToJSON(content)
}

func (c *ContractCompatibilityChecker) parseYAMLToJSON(yamlContent []byte) (gjson.Result, error) {
	var yamlData interface{}
	err := yaml.Unmarshal(yamlContent, &yamlData)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("failed to parse YAML: %w", err)
	}

	jsonData, err := json.Marshal(yamlData)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("failed to convert YAML to JSON: %w", err)
	}

	return gjson.ParseBytes(jsonData), nil
}

func (c *ContractCompatibilityChecker) analyzeAndExplainCompatibilityPlan(consumer, provider gjson.Result) {
	fmt.Println("🔍 АНАЛИЗ СПЕЦИФИКАЦИЙ И ПЛАН ПРОВЕРКИ СОВМЕСТИМОСТИ")
	fmt.Println(strings.Repeat("=", 60))
	
	// Анализируем Consumer
	fmt.Printf("📋 ПОТРЕБИТЕЛЬ: %s v%s\n", consumer.Get("info.title").String(), consumer.Get("info.version").String())
	consumerChannels := consumer.Get("channels")
	
	var restEndpoints []string
	var mqChannels []string
	
	consumerChannels.ForEach(func(channelName, channel gjson.Result) bool {
		addr := channel.Get("address").String()
		if strings.HasPrefix(addr, "/api/") || strings.HasPrefix(addr, "http") {
			restEndpoints = append(restEndpoints, fmt.Sprintf("  • %s: %s", channelName.String(), addr))
		} else {
			mqChannels = append(mqChannels, fmt.Sprintf("  • %s: %s", channelName.String(), addr))
		}
		return true
	})
	
	fmt.Println("  REST endpoints для проверки:")
	for _, endpoint := range restEndpoints {
		fmt.Println(endpoint)
	}
	fmt.Println("  MQ каналы:")
	for _, channel := range mqChannels {
		fmt.Println(channel)
	}
	
	// Анализируем Provider
	fmt.Printf("\n🏢 ПРОВАЙДЕР: %s v%s\n", provider.Get("info.title").String(), provider.Get("info.version").String())
	providerChannels := provider.Get("channels")
	
	var providerRestEndpoints []string
	var providerMqChannels []string
	
	providerChannels.ForEach(func(channelName, channel gjson.Result) bool {
		addr := channel.Get("address").String()
		if strings.HasPrefix(addr, "/api/") || strings.HasPrefix(addr, "http") {
			providerRestEndpoints = append(providerRestEndpoints, fmt.Sprintf("  • %s: %s", channelName.String(), addr))
		} else {
			providerMqChannels = append(providerMqChannels, fmt.Sprintf("  • %s: %s", channelName.String(), addr))
		}
		return true
	})
	
	fmt.Println("  REST endpoints:")
	for _, endpoint := range providerRestEndpoints {
		fmt.Println(endpoint)
	}
	fmt.Println("  MQ каналы:")
	for _, channel := range providerMqChannels {
		fmt.Println(channel)
	}
	
	// План проверки
	fmt.Println("\n📝 ПЛАН ПРОВЕРКИ СОВМЕСТИМОСТИ:")
	fmt.Println("  1. 🌐 REST Endpoints:")
	fmt.Println("     - Сопоставление путей (нормализация параметров)")
	fmt.Println("     - Сравнение структур запросов и ответов")
	fmt.Println("     - Валидация типов полей и форматов")
	
	// Проверяем ограничения каналов из конфигурации
	if c.Config != nil && c.Config.ContractTests.Provider.ChannelRestrictions.CheckMode == "rest_only" {
		fmt.Println("  2. 📨 MQ Каналы:")
		fmt.Printf("     ⚠️  ПРОПУЩЕНО - провайдер '%s' настроен только для HTTP (режим: rest_only)\n", c.Config.ContractTests.Provider.Name)
	} else {
		fmt.Println("  2. 📨 MQ Каналы:")
		fmt.Println("     - Проверка адресов очередей")
		fmt.Println("     - Сравнение структур сообщений")
	}
	
	fmt.Println("  3. 📋 Схемы данных:")
	fmt.Println("     - Совместимость типов полей")
	fmt.Println("     - Обязательные поля")
	fmt.Println("     - Enum значения")
	fmt.Println("\n" + strings.Repeat("=", 60))
}

func (c *ContractCompatibilityChecker) checkRESTEndpointCompatibility(consumer, provider gjson.Result, result *CompatibilityResult) {
	fmt.Println("\n🌐 ПРОВЕРКА REST ENDPOINTS...")
	
	consumerChannels := consumer.Get("channels")
	providerChannels := provider.Get("channels")
	
	if !consumerChannels.Exists() || !providerChannels.Exists() {
		result.Warnings = append(result.Warnings, CompatibilityWarning{
			Type:    "rest_endpoints",
			Field:   "channels",
			Message: "Channel definitions not found in one or both specifications",
			Suggestion: "Ensure both consumer and provider have channel definitions",
		})
		return
	}
	
	// Собираем REST endpoints провайдера для быстрого поиска
	providerEndpoints := make(map[string]gjson.Result)
	providerChannels.ForEach(func(channelName, channel gjson.Result) bool {
		addr := channel.Get("address").String()
		if strings.HasPrefix(addr, "/api/") || strings.HasPrefix(addr, "http") {
			normalizedPath := c.normalizeResourcePath(addr)
			providerEndpoints[normalizedPath] = channel
			fmt.Printf("  📍 Провайдер endpoint: %s -> %s\n", addr, normalizedPath)
		}
		return true
	})
	
	// Проверяем REST endpoints потребителя
	consumerChannels.ForEach(func(channelName, channel gjson.Result) bool {
		channelAddr := channel.Get("address").String()
		
		if strings.HasPrefix(channelAddr, "/api/") || strings.HasPrefix(channelAddr, "http") {
			normalizedConsumerPath := c.normalizeResourcePath(channelAddr)
			fmt.Printf("  🔍 Поиск для потребителя: %s -> %s\n", channelAddr, normalizedConsumerPath)
			
			if providerChannel, found := providerEndpoints[normalizedConsumerPath]; found {
				fmt.Printf("  ✅ Найдено соответствие для %s\n", channelAddr)
				// Сравниваем структуры сообщений
				c.compareRESTMessages(channelName.String(), channel, providerChannel, consumer, provider, result)
			} else {
				fmt.Printf("  ❌ Не найден endpoint для %s\n", channelAddr)
				result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
					Type:        "rest_endpoint_not_found",
					Field:       fmt.Sprintf("channels.%s.address", channelName.String()),
					OldValue:    channelAddr,
					NewValue:    "not_found_in_provider",
					Description: fmt.Sprintf("REST endpoint '%s' not found in provider specification", channelAddr),
					Severity:    "high",
				})
			}
		}
		return true
	})
}

func (c *ContractCompatibilityChecker) checkMQChannelCompatibility(consumer, provider gjson.Result, result *CompatibilityResult) {
	// Проверяем ограничения каналов из конфигурации
	if c.Config != nil && c.Config.ContractTests.Provider.ChannelRestrictions.CheckMode == "rest_only" {
		fmt.Println("\n📨 ПРОВЕРКА MQ КАНАЛОВ...")
		fmt.Println("  ℹ️  Проверка MQ каналов пропущена согласно конфигурации (режим: rest_only)")
		fmt.Printf("  📋 Провайдер '%s' настроен только для HTTP взаимодействия\n", c.Config.ContractTests.Provider.Name)
		return
	}
	
	fmt.Println("\n📨 ПРОВЕРКА MQ КАНАЛОВ...")
	
	consumerChannels := consumer.Get("channels")
	providerChannels := provider.Get("channels")
	
	// Логика для MQ каналов - они должны точно совпадать по именам
	consumerChannels.ForEach(func(channelName, channel gjson.Result) bool {
		channelAddr := channel.Get("address").String()
		
		// Проверяем только MQ каналы (не REST)
		if !strings.HasPrefix(channelAddr, "/api/") && !strings.HasPrefix(channelAddr, "http") {
			providerChannel := providerChannels.Get(channelName.String())
			
			if !providerChannel.Exists() {
				fmt.Printf("  ❌ MQ канал '%s' не найден в провайдере\n", channelName.String())
				result.Warnings = append(result.Warnings, CompatibilityWarning{
					Type:    "mq_channel_missing",
					Field:   fmt.Sprintf("channels.%s", channelName.String()),
					Message: fmt.Sprintf("MQ channel '%s' not found in provider (this may be expected for consumer-specific channels)", channelName.String()),
					Suggestion: "Verify if this channel should exist in provider specification",
				})
			} else {
				fmt.Printf("  ✅ MQ канал '%s' найден\n", channelName.String())
			}
		}
		return true
	})
}

func (c *ContractCompatibilityChecker) normalizeResourcePath(path string) string {
	// Нормализуем пути для сравнения, заменяя параметры
	normalized := strings.ReplaceAll(path, "{clientId}", "{param}")
	normalized = strings.ReplaceAll(normalized, "{Api-clientId}", "{param}")
	return normalized
}

func (c *ContractCompatibilityChecker) compareRESTMessages(channelName string, consumerChannel, providerChannel gjson.Result, consumer, provider gjson.Result, result *CompatibilityResult) {
	fmt.Printf("    🔍 Сравниваем сообщения для канала %s\n", channelName)
	
	consumerMessages := consumerChannel.Get("messages")
	providerMessages := providerChannel.Get("messages")
	
	if !consumerMessages.Exists() || !providerMessages.Exists() {
		result.Warnings = append(result.Warnings, CompatibilityWarning{
			Type:    "rest_messages_missing",
			Field:   fmt.Sprintf("channels.%s.messages", channelName),
			Message: "Messages not defined in one or both channel specifications",
			Suggestion: "Ensure both channels have message definitions",
		})
		return
	}
	
	// Анализируем request и response сообщения
	consumerMessages.ForEach(func(msgName, msgRef gjson.Result) bool {
		msgRefStr := msgRef.Get("$ref").String()
		if msgRefStr != "" {
			msgType := c.inferMessageType(msgName.String())
			fmt.Printf("      📨 Сообщение потребителя: %s (%s) -> %s\n", msgName.String(), msgType, msgRefStr)
			
			// Ищем соответствующий тип сообщения в провайдере
			c.findAndCompareProviderMessage(msgType, msgRefStr, providerMessages, consumer, provider, channelName, result)
		}
		return true
	})
}

func (c *ContractCompatibilityChecker) inferMessageType(messageName string) string {
	name := strings.ToLower(messageName)
	if strings.Contains(name, "request") {
		return "request"
	} else if strings.Contains(name, "response") {
		return "response"
	}
	return "unknown"
}

func (c *ContractCompatibilityChecker) findAndCompareProviderMessage(msgType, consumerMsgRef string, providerMessages gjson.Result, consumer, provider gjson.Result, channelName string, result *CompatibilityResult) {
	consumerMsgName := c.extractMessageName(consumerMsgRef)
	
	// Ищем соответствующий тип сообщения в провайдере
	var providerMsgRef string
	providerMessages.ForEach(func(msgName, msgRef gjson.Result) bool {
		providerType := c.inferMessageType(msgName.String())
		if providerType == msgType {
			providerMsgRef = msgRef.Get("$ref").String()
			fmt.Printf("        ✅ Найдено соответствие: %s -> %s\n", msgName.String(), providerMsgRef)
			return false
		}
		return true
	})
	
	if providerMsgRef == "" {
		result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
			Type:        "rest_message_type_missing",
			Field:       fmt.Sprintf("channels.%s.messages", channelName),
			OldValue:    consumerMsgRef,
			NewValue:    "not_found",
			Description: fmt.Sprintf("Message type '%s' not found in provider for channel '%s'", msgType, channelName),
			Severity:    "high",
		})
		return
	}
	
	// Сравниваем структуры сообщений
	c.compareMessageSchemas(consumerMsgName, c.extractMessageName(providerMsgRef), consumer, provider, channelName, msgType, result)
}

func (c *ContractCompatibilityChecker) compareMessageSchemas(consumerMsgName, providerMsgName string, consumer, provider gjson.Result, channelName, msgType string, result *CompatibilityResult) {
	fmt.Printf("        🔍 Сравниваем схемы: %s vs %s\n", consumerMsgName, providerMsgName)
	
	consumerMsg := consumer.Get(fmt.Sprintf("components.messages.%s", consumerMsgName))
	providerMsg := provider.Get(fmt.Sprintf("components.messages.%s", providerMsgName))
	
	if !consumerMsg.Exists() || !providerMsg.Exists() {
		result.Warnings = append(result.Warnings, CompatibilityWarning{
			Type:    "message_schema_missing",
			Field:   fmt.Sprintf("channels.%s.messages.%s", channelName, msgType),
			Message: fmt.Sprintf("Message schema not found: consumer '%s' or provider '%s'", consumerMsgName, providerMsgName),
			Suggestion: "Verify message definitions exist in components.messages",
		})
		return
	}
	
	// Сравниваем payload схемы
	consumerPayloadRef := consumerMsg.Get("payload.$ref").String()
	providerPayloadRef := providerMsg.Get("payload.$ref").String()
	
	if consumerPayloadRef != "" && providerPayloadRef != "" {
		c.comparePayloadSchemas(consumerPayloadRef, providerPayloadRef, consumer, provider, channelName, msgType, result)
	}
}

func (c *ContractCompatibilityChecker) comparePayloadSchemas(consumerPayloadRef, providerPayloadRef string, consumer, provider gjson.Result, channelName, msgType string, result *CompatibilityResult) {
	consumerSchemaName := c.extractSchemaName(consumerPayloadRef)
	providerSchemaName := c.extractSchemaName(providerPayloadRef)
	
	fmt.Printf("          📋 Сравниваем payload схемы: %s vs %s\n", consumerSchemaName, providerSchemaName)
	
	consumerSchema := consumer.Get(fmt.Sprintf("components.schemas.%s", consumerSchemaName))
	providerSchema := provider.Get(fmt.Sprintf("components.schemas.%s", providerSchemaName))
	
	if !consumerSchema.Exists() || !providerSchema.Exists() {
		result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
			Type:        "payload_schema_missing",
			Field:       fmt.Sprintf("channels.%s.messages.%s.payload", channelName, msgType),
			OldValue:    consumerPayloadRef,
			NewValue:    providerPayloadRef,
			Description: fmt.Sprintf("Payload schema missing: consumer '%s' or provider '%s'", consumerSchemaName, providerSchemaName),
			Severity:    "high",
		})
		return
	}
	
	// Детальное сравнение полей
	c.compareSchemaFields(consumerSchema, providerSchema, consumerSchemaName, providerSchemaName, channelName, msgType, result)
}

func (c *ContractCompatibilityChecker) extractSchemaName(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) >= 4 && parts[0] == "#" && parts[1] == "components" && parts[2] == "schemas" {
		return parts[3]
	}
	return ""
}

func (c *ContractCompatibilityChecker) compareSchemaFields(consumerSchema, providerSchema gjson.Result, consumerSchemaName, providerSchemaName, channelName, msgType string, result *CompatibilityResult) {
	fmt.Printf("            🔎 Детальное сравнение полей схем\n")
	
	// Сравниваем обязательные поля
	c.compareRequiredFieldsDetailed(consumerSchema, providerSchema, consumerSchemaName, providerSchemaName, channelName, msgType, result)
	
	// Сравниваем типы полей
	c.compareFieldTypesDetailed(consumerSchema, providerSchema, consumerSchemaName, providerSchemaName, channelName, msgType, result)
}

func (c *ContractCompatibilityChecker) compareRequiredFieldsDetailed(consumerSchema, providerSchema gjson.Result, consumerSchemaName, providerSchemaName, channelName, msgType string, result *CompatibilityResult) {
	consumerRequired := consumerSchema.Get("required")
	providerRequired := providerSchema.Get("required")
	
	if !consumerRequired.Exists() || !providerRequired.Exists() {
		return
	}
	
	consumerRequiredFields := make(map[string]bool)
	consumerRequired.ForEach(func(_, value gjson.Result) bool {
		consumerRequiredFields[value.String()] = true
		return true
	})
	
	providerRequiredFields := make(map[string]bool)
	providerRequired.ForEach(func(_, value gjson.Result) bool {
		providerRequiredFields[value.String()] = true
		return true
	})
	
	// Проверяем совместимость обязательных полей
	for field := range consumerRequiredFields {
		if !providerRequiredFields[field] {
			if msgType == "response" {
				// Для ответов - провайдер должен предоставлять все поля, которые ожидает потребитель
				result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
					Type:        "required_response_field_missing",
					Field:       fmt.Sprintf("channels.%s.messages.%s.payload.required.%s", channelName, msgType, field),
					OldValue:    "required",
					NewValue:    "optional_or_missing",
					Description: fmt.Sprintf("Required response field '%s' is not guaranteed by provider in schema '%s'", field, providerSchemaName),
					Severity:    "high",
				})
			}
		}
	}
	
	for field := range providerRequiredFields {
		if !consumerRequiredFields[field] && msgType == "request" {
			// Для запросов - провайдер требует поле, которое потребитель может не отправлять
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Type:        "required_request_field_added",
				Field:       fmt.Sprintf("channels.%s.messages.%s.payload.required.%s", channelName, msgType, field),
				OldValue:    "optional",
				NewValue:    "required",
				Description: fmt.Sprintf("Provider now requires field '%s' in request schema '%s' that consumer may not provide", field, providerSchemaName),
				Severity:    "high",
			})
		}
	}
}

func (c *ContractCompatibilityChecker) compareFieldTypesDetailed(consumerSchema, providerSchema gjson.Result, consumerSchemaName, providerSchemaName, channelName, msgType string, result *CompatibilityResult) {
	consumerProperties := consumerSchema.Get("properties")
	providerProperties := providerSchema.Get("properties")
	
	if !consumerProperties.Exists() || !providerProperties.Exists() {
		return
	}
	
	consumerProperties.ForEach(func(fieldName, consumerField gjson.Result) bool {
		fieldNameStr := fieldName.String()
		providerField := providerProperties.Get(fieldNameStr)
		
		if providerField.Exists() {
			consumerType := consumerField.Get("type").String()
			providerType := providerField.Get("type").String()
			
			if consumerType != "" && providerType != "" && !c.areTypesCompatible(consumerType, providerType) {
				result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
					Type:        "field_type_incompatible",
					Field:       fmt.Sprintf("channels.%s.messages.%s.payload.properties.%s.type", channelName, msgType, fieldNameStr),
					OldValue:    consumerType,
					NewValue:    providerType,
					Description: fmt.Sprintf("Field '%s' type incompatible: consumer expects '%s', provider provides '%s' in %s", fieldNameStr, consumerType, providerType, msgType),
					Severity:    "high",
				})
			}
			
			// Проверяем enum значения
			c.compareEnumValuesDetailed(consumerField, providerField, fieldNameStr, channelName, msgType, result)
		}
		return true
	})
}

func (c *ContractCompatibilityChecker) compareEnumValuesDetailed(consumerField, providerField gjson.Result, fieldName, channelName, msgType string, result *CompatibilityResult) {
	consumerEnum := consumerField.Get("enum")
	providerEnum := providerField.Get("enum")
	
	if !consumerEnum.Exists() || !providerEnum.Exists() {
		return
	}
	
	consumerValues := make(map[string]bool)
	consumerEnum.ForEach(func(_, value gjson.Result) bool {
		consumerValues[value.String()] = true
		return true
	})
	
	providerValues := make(map[string]bool)
	providerEnum.ForEach(func(_, value gjson.Result) bool {
		providerValues[value.String()] = true
		return true
	})
	
	if msgType == "response" {
		// Для ответов - провайдер может вернуть значения, которые потребитель не ожидает
		for value := range providerValues {
			if !consumerValues[value] {
				result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
					Type:        "response_enum_value_added",
					Field:       fmt.Sprintf("channels.%s.messages.%s.payload.properties.%s.enum", channelName, msgType, fieldName),
					OldValue:    "not_expected",
					NewValue:    value,
					Description: fmt.Sprintf("Provider may return unexpected enum value '%s' for field '%s' in response", value, fieldName),
					Severity:    "medium",
				})
			}
		}
	}
	
	if msgType == "request" {
		// Для запросов - потребитель может отправить значения, которые провайдер не поддерживает
		for value := range consumerValues {
			if !providerValues[value] {
				result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
					Type:        "request_enum_value_unsupported",
					Field:       fmt.Sprintf("channels.%s.messages.%s.payload.properties.%s.enum", channelName, msgType, fieldName),
					OldValue:    value,
					NewValue:    "unsupported",
					Description: fmt.Sprintf("Consumer may send unsupported enum value '%s' for field '%s' in request", value, fieldName),
					Severity:    "high",
				})
			}
		}
	}
}

func (c *ContractCompatibilityChecker) compareChannelMessages(channelName string, consumerChannel, providerChannel gjson.Result, result *CompatibilityResult) {
	// Получаем сообщения каналов
	consumerMessages := consumerChannel.Get("messages")
	providerMessages := providerChannel.Get("messages")
	
	if !consumerMessages.Exists() || !providerMessages.Exists() {
		return
	}
	
	// Сравниваем схемы сообщений
	consumerMessages.ForEach(func(msgName, msgRef gjson.Result) bool {
		// Получаем reference на схему
		consumerMsgRef := msgRef.Get("$ref").String()
		if consumerMsgRef == "" {
			return true
		}
		
		// Ищем соответствующее сообщение в провайдере
		providerMessages.ForEach(func(providerMsgName, providerMsgRef gjson.Result) bool {
			providerRefStr := providerMsgRef.Get("$ref").String()
			if providerRefStr != "" {
				// Сравниваем структуры сообщений
				c.compareMessageStructures(channelName, consumerMsgRef, providerRefStr, result)
			}
			return true
		})
		return true
	})
}

func (c *ContractCompatibilityChecker) compareMessageStructures(channelName, consumerMsgRef, providerMsgRef string, result *CompatibilityResult) {
	// Извлекаем имена сообщений из references
	consumerMsgName := c.extractMessageName(consumerMsgRef)
	providerMsgName := c.extractMessageName(providerMsgRef)
	
	if consumerMsgName == "" || providerMsgName == "" {
		result.Warnings = append(result.Warnings, CompatibilityWarning{
			Type:    "message_reference",
			Field:   fmt.Sprintf("channels.%s.messages", channelName),
			Message: fmt.Sprintf("Could not extract message names from references '%s' and '%s'", consumerMsgRef, providerMsgRef),
			Suggestion: "Verify message reference format",
		})
		return
	}
	
	// Получаем структуры сообщений из спецификаций
	// Для более детального сравнения нужно получить access к исходным спецификациям
	result.Warnings = append(result.Warnings, CompatibilityWarning{
		Type:    "message_structure",
		Field:   fmt.Sprintf("channels.%s.messages", channelName),
		Message: fmt.Sprintf("Comparing message structures: consumer '%s' vs provider '%s'", consumerMsgName, providerMsgName),
		Suggestion: "Verify payload and header structures are compatible",
	})
}

func (c *ContractCompatibilityChecker) extractMessageName(ref string) string {
	// Извлекаем имя сообщения из #/components/messages/messageName
	parts := strings.Split(ref, "/")
	if len(parts) >= 4 && parts[0] == "#" && parts[1] == "components" && parts[2] == "messages" {
		return parts[3]
	}
	return ""
}

func (c *ContractCompatibilityChecker) areTypesCompatible(consumerType, providerType string) bool {
	// Проверяем совместимость типов
	if consumerType == providerType {
		return true
	}
	
	// Некоторые типы могут быть совместимы (например, integer и number)
	compatibleTypes := map[string][]string{
		"integer": {"number"},
		"number":  {"integer"},
	}
	
	if compatibleList, exists := compatibleTypes[consumerType]; exists {
		for _, compatibleType := range compatibleList {
			if compatibleType == providerType {
				return true
			}
		}
	}
	
	return false
}

func (c *ContractCompatibilityChecker) checkFieldFormat(consumerField, providerField gjson.Result, schemaName, fieldName string, result *CompatibilityResult) {
	consumerFormat := consumerField.Get("format").String()
	providerFormat := providerField.Get("format").String()
	
	// Если форматы указаны и отличаются
	if consumerFormat != "" && providerFormat != "" && consumerFormat != providerFormat {
		// Проверяем критические различия в формате
		if c.isFormatBreakingChange(consumerFormat, providerFormat) {
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Type:        "field_format_changed",
				Field:       fmt.Sprintf("components.schemas.%s.properties.%s.format", schemaName, fieldName),
				OldValue:    consumerFormat,
				NewValue:    providerFormat,
				Description: fmt.Sprintf("Field '%s' format changed from '%s' to '%s' in schema '%s'", fieldName, consumerFormat, providerFormat, schemaName),
				Severity:    "medium",
			})
		}
	}
}

func (c *ContractCompatibilityChecker) isFormatBreakingChange(consumerFormat, providerFormat string) bool {
	// Определяем критические различия в формате
	criticalFormats := map[string][]string{
		"double":    {"int64", "int32"},
		"int64":     {"int32", "double"},
		"int32":     {"int64", "double"},
		"date-time": {"date"},
	}
	
	if incompatibleList, exists := criticalFormats[consumerFormat]; exists {
		for _, incompatible := range incompatibleList {
			if incompatible == providerFormat {
				return true
			}
		}
	}
	
	return false
}

func (c *ContractCompatibilityChecker) checkSchemaCompatibility(consumer, provider gjson.Result, result *CompatibilityResult) {
	// Проверяем совместимость схем данных
	consumerSchemas := consumer.Get("components.schemas")
	providerSchemas := provider.Get("components.schemas")

	if !consumerSchemas.Exists() || !providerSchemas.Exists() {
		result.Warnings = append(result.Warnings, CompatibilityWarning{
			Type:    "schema",
			Field:   "components.schemas",
			Message: "Schema definitions not found in one or both specifications",
			Suggestion: "Ensure both consumer and provider have schema definitions",
		})
		return
	}

	consumerSchemas.ForEach(func(key, value gjson.Result) bool {
		schemaName := key.String()
		providerSchema := providerSchemas.Get(schemaName)
		
		if !providerSchema.Exists() {
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Type:        "schema_removed",
				Field:       fmt.Sprintf("components.schemas.%s", schemaName),
				OldValue:    "exists",
				NewValue:    "not_exists",
				Description: fmt.Sprintf("Schema '%s' was removed from provider specification", schemaName),
				Severity:    "high",
			})
			return true
		}

		// Проверяем обязательные поля
		c.checkRequiredFields(value, providerSchema, schemaName, result)

		// Проверяем типы полей
		c.checkFieldTypes(value, providerSchema, schemaName, result)

		return true
	})
}

func (c *ContractCompatibilityChecker) checkRequiredFields(consumerSchema, providerSchema gjson.Result, schemaName string, result *CompatibilityResult) {
	consumerRequired := consumerSchema.Get("required")
	providerRequired := providerSchema.Get("required")

	if !consumerRequired.Exists() || !providerRequired.Exists() {
		return
	}

	consumerRequiredFields := make(map[string]bool)
	consumerRequired.ForEach(func(_, value gjson.Result) bool {
		consumerRequiredFields[value.String()] = true
		return true
	})

	providerRequiredFields := make(map[string]bool)
	providerRequired.ForEach(func(_, value gjson.Result) bool {
		providerRequiredFields[value.String()] = true
		return true
	})

	// Проверяем, добавились ли новые обязательные поля
	for field := range providerRequiredFields {
		if !consumerRequiredFields[field] {
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Type:        "required_field_added",
				Field:       fmt.Sprintf("components.schemas.%s.required.%s", schemaName, field),
				OldValue:    "optional",
				NewValue:    "required",
				Description: fmt.Sprintf("Field '%s' became required in schema '%s'", field, schemaName),
				Severity:    "high",
			})
		}
	}

	// Проверяем, удалились ли обязательные поля
	for field := range consumerRequiredFields {
		if !providerRequiredFields[field] {
			result.Warnings = append(result.Warnings, CompatibilityWarning{
				Type:    "required_field_removed",
				Field:   fmt.Sprintf("components.schemas.%s.required.%s", schemaName, field),
				Message: fmt.Sprintf("Required field '%s' is no longer required in schema '%s'", field, schemaName),
				Suggestion: "This is generally safe but verify consumer behavior",
			})
		}
	}
}

func (c *ContractCompatibilityChecker) checkFieldTypes(consumerSchema, providerSchema gjson.Result, schemaName string, result *CompatibilityResult) {
	consumerProperties := consumerSchema.Get("properties")
	providerProperties := providerSchema.Get("properties")

	if !consumerProperties.Exists() || !providerProperties.Exists() {
		return
	}

	consumerProperties.ForEach(func(fieldName, consumerField gjson.Result) bool {
		fieldNameStr := fieldName.String()
		providerField := providerProperties.Get(fieldNameStr)

		if !providerField.Exists() {
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Type:        "field_removed",
				Field:       fmt.Sprintf("components.schemas.%s.properties.%s", schemaName, fieldNameStr),
				OldValue:    "exists",
				NewValue:    "not_exists",
				Description: fmt.Sprintf("Field '%s' was removed from schema '%s'", fieldNameStr, schemaName),
				Severity:    "high",
			})
			return true
		}

		// Проверяем тип поля
		consumerType := consumerField.Get("type").String()
		providerType := providerField.Get("type").String()

		if consumerType != "" && providerType != "" && consumerType != providerType {
			// Проверяем совместимость типов
			if !c.areTypesCompatible(consumerType, providerType) {
				result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
					Type:        "field_type_changed",
					Field:       fmt.Sprintf("components.schemas.%s.properties.%s.type", schemaName, fieldNameStr),
					OldValue:    consumerType,
					NewValue:    providerType,
					Description: fmt.Sprintf("Field '%s' type changed from '%s' to '%s' in schema '%s'", fieldNameStr, consumerType, providerType, schemaName),
					Severity:    "high",
				})
			}
		}
		
		// Проверяем формат поля
		c.checkFieldFormat(consumerField, providerField, schemaName, fieldNameStr, result)

		// Проверяем enum значения
		c.checkEnumValues(consumerField, providerField, schemaName, fieldNameStr, result)

		return true
	})
}

func (c *ContractCompatibilityChecker) checkEnumValues(consumerField, providerField gjson.Result, schemaName, fieldName string, result *CompatibilityResult) {
	consumerEnum := consumerField.Get("enum")
	providerEnum := providerField.Get("enum")

	if !consumerEnum.Exists() || !providerEnum.Exists() {
		return
	}

	consumerValues := make(map[string]bool)
	consumerEnum.ForEach(func(_, value gjson.Result) bool {
		consumerValues[value.String()] = true
		return true
	})

	providerValues := make(map[string]bool)
	providerEnum.ForEach(func(_, value gjson.Result) bool {
		providerValues[value.String()] = true
		return true
	})

	// Проверяем, удалились ли enum значения
	for value := range consumerValues {
		if !providerValues[value] {
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Type:        "enum_value_removed",
				Field:       fmt.Sprintf("components.schemas.%s.properties.%s.enum", schemaName, fieldName),
				OldValue:    value,
				NewValue:    "removed",
				Description: fmt.Sprintf("Enum value '%s' was removed from field '%s' in schema '%s'", value, fieldName, schemaName),
				Severity:    "medium",
			})
		}
	}

	// Проверяем, добавились ли новые enum значения (это не breaking change)
	for value := range providerValues {
		if !consumerValues[value] {
			result.Warnings = append(result.Warnings, CompatibilityWarning{
				Type:    "enum_value_added",
				Field:   fmt.Sprintf("components.schemas.%s.properties.%s.enum", schemaName, fieldName),
				Message: fmt.Sprintf("New enum value '%s' was added to field '%s' in schema '%s'", value, fieldName, schemaName),
				Suggestion: "Ensure consumer can handle new enum values gracefully",
			})
		}
	}
}

func (c *ContractCompatibilityChecker) checkChannelCompatibility(consumer, provider gjson.Result, result *CompatibilityResult) {
	consumerChannels := consumer.Get("channels")
	providerChannels := provider.Get("channels")

	if !consumerChannels.Exists() || !providerChannels.Exists() {
		result.Warnings = append(result.Warnings, CompatibilityWarning{
			Type:    "channels",
			Field:   "channels",
			Message: "Channel definitions not found in one or both specifications",
			Suggestion: "Ensure both consumer and provider have channel definitions",
		})
		return
	}

	consumerChannels.ForEach(func(key, value gjson.Result) bool {
		channelName := key.String()
		providerChannel := providerChannels.Get(channelName)

		if !providerChannel.Exists() {
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Type:        "channel_removed",
				Field:       fmt.Sprintf("channels.%s", channelName),
				OldValue:    "exists",
				NewValue:    "not_exists",
				Description: fmt.Sprintf("Channel '%s' was removed from provider specification", channelName),
				Severity:    "high",
			})
			return true
		}

		// Проверяем address канала
		consumerAddress := value.Get("address").String()
		providerAddress := providerChannel.Get("address").String()

		if consumerAddress != providerAddress {
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Type:        "channel_address_changed",
				Field:       fmt.Sprintf("channels.%s.address", channelName),
				OldValue:    consumerAddress,
				NewValue:    providerAddress,
				Description: fmt.Sprintf("Channel '%s' address changed from '%s' to '%s'", channelName, consumerAddress, providerAddress),
				Severity:    "high",
			})
		}

		return true
	})
}

func (c *ContractCompatibilityChecker) checkOperationCompatibility(consumer, provider gjson.Result, result *CompatibilityResult) {
	consumerOperations := consumer.Get("operations")
	providerOperations := provider.Get("operations")

	if !consumerOperations.Exists() || !providerOperations.Exists() {
		result.Warnings = append(result.Warnings, CompatibilityWarning{
			Type:    "operations",
			Field:   "operations",
			Message: "Operation definitions not found in one or both specifications",
			Suggestion: "Ensure both consumer and provider have operation definitions",
		})
		return
	}

	consumerOperations.ForEach(func(key, value gjson.Result) bool {
		operationName := key.String()
		providerOperation := providerOperations.Get(operationName)

		if !providerOperation.Exists() {
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Type:        "operation_removed",
				Field:       fmt.Sprintf("operations.%s", operationName),
				OldValue:    "exists",
				NewValue:    "not_exists",
				Description: fmt.Sprintf("Operation '%s' was removed from provider specification", operationName),
				Severity:    "high",
			})
			return true
		}

		// Проверяем action операции
		consumerAction := value.Get("action").String()
		providerAction := providerOperation.Get("action").String()

		if consumerAction != providerAction {
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Type:        "operation_action_changed",
				Field:       fmt.Sprintf("operations.%s.action", operationName),
				OldValue:    consumerAction,
				NewValue:    providerAction,
				Description: fmt.Sprintf("Operation '%s' action changed from '%s' to '%s'", operationName, consumerAction, providerAction),
				Severity:    "high",
			})
		}

		return true
	})
}

func (c *ContractCompatibilityChecker) checkMessageCompatibility(consumer, provider gjson.Result, result *CompatibilityResult) {
	consumerMessages := consumer.Get("components.messages")
	providerMessages := provider.Get("components.messages")

	if !consumerMessages.Exists() || !providerMessages.Exists() {
		result.Warnings = append(result.Warnings, CompatibilityWarning{
			Type:    "messages",
			Field:   "components.messages",
			Message: "Message definitions not found in one or both specifications",
			Suggestion: "Ensure both consumer and provider have message definitions",
		})
		return
	}

	consumerMessages.ForEach(func(key, value gjson.Result) bool {
		messageName := key.String()
		providerMessage := providerMessages.Get(messageName)

		if !providerMessage.Exists() {
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Type:        "message_removed",
				Field:       fmt.Sprintf("components.messages.%s", messageName),
				OldValue:    "exists",
				NewValue:    "not_exists",
				Description: fmt.Sprintf("Message '%s' was removed from provider specification", messageName),
				Severity:    "high",
			})
			return true
		}

		// Проверяем contentType
		consumerContentType := value.Get("contentType").String()
		providerContentType := providerMessage.Get("contentType").String()

		if consumerContentType != providerContentType {
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Type:        "message_content_type_changed",
				Field:       fmt.Sprintf("components.messages.%s.contentType", messageName),
				OldValue:    consumerContentType,
				NewValue:    providerContentType,
				Description: fmt.Sprintf("Message '%s' contentType changed from '%s' to '%s'", messageName, consumerContentType, providerContentType),
				Severity:    "medium",
			})
		}

		return true
	})
}

func (c *ContractCompatibilityChecker) GenerateReport(result *CompatibilityResult) string {
	var report strings.Builder

	report.WriteString("# Contract Compatibility Report\n\n")
	report.WriteString(fmt.Sprintf("**Checked at:** %s\n", result.CheckedAt.Format("2006-01-02 15:04:05")))
	report.WriteString(fmt.Sprintf("**Consumer Version:** %s\n", result.ConsumerVersion))
	report.WriteString(fmt.Sprintf("**Provider Version:** %s\n", result.ProviderVersion))
	report.WriteString(fmt.Sprintf("**Compatible:** %t\n\n", result.IsCompatible))

	if len(result.BreakingChanges) > 0 {
		report.WriteString("## 🚨 Breaking Changes\n\n")
		for _, change := range result.BreakingChanges {
			report.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", change.Type, change.Severity, change.Description))
			report.WriteString(fmt.Sprintf("  - Field: `%s`\n", change.Field))
			report.WriteString(fmt.Sprintf("  - Old: `%s` → New: `%s`\n\n", change.OldValue, change.NewValue))
		}
	}

	if len(result.Warnings) > 0 {
		report.WriteString("## ⚠️ Warnings\n\n")
		for _, warning := range result.Warnings {
			report.WriteString(fmt.Sprintf("- **%s**: %s\n", warning.Type, warning.Message))
			report.WriteString(fmt.Sprintf("  - Field: `%s`\n", warning.Field))
			if warning.Suggestion != "" {
				report.WriteString(fmt.Sprintf("  - Suggestion: %s\n", warning.Suggestion))
			}
			report.WriteString("\n")
		}
	}

	if result.IsCompatible {
		report.WriteString("## ✅ Result\n\n")
		report.WriteString("The consumer contract is compatible with the provider specification. No breaking changes detected.\n")
	} else {
		report.WriteString("## ❌ Result\n\n")
		report.WriteString("The consumer contract is **NOT COMPATIBLE** with the provider specification. Breaking changes detected that may cause consumer failures.\n")
	}

	return report.String()
}

func main() {
	// Проверяем аргументы командной строки
	// os.Args[0] = program name, os.Args[1+] = user args
	var checker *ContractCompatibilityChecker
	var err error

	if len(os.Args) == 1 {
		// Нет аргументов - используем config файл по умолчанию
		checker, err = NewContractCompatibilityCheckerFromConfig("contract-tests.yaml")
		if err != nil {
			log.Fatalf("Error loading config: %v", err)
		}
	} else if len(os.Args) == 2 {
		// Один аргумент - либо config файл, либо показать help
		arg := os.Args[1]
		if arg == "-h" || arg == "--help" || arg == "help" {
			fmt.Println("Usage:")
			fmt.Println("  go run contract_compatibility_checker.go models.go [config_file]")
			fmt.Println("  go run contract_compatibility_checker.go models.go <consumer_spec_path> <provider_git_url>")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  go run contract_compatibility_checker.go models.go                    # Uses contract-tests.yaml")
			fmt.Println("  go run contract_compatibility_checker.go models.go custom-config.yaml")
			fmt.Println("  go run contract_compatibility_checker.go models.go ../api-specification/asyncapi.yml https://git.codemonsters.team/guides/wallet-balance/-/raw/main/api-specification/asyncapi.yml")
			os.Exit(0)
		}
		// Используем переданный config файл
		configPath := arg
		checker, err = NewContractCompatibilityCheckerFromConfig(configPath)
		if err != nil {
			log.Fatalf("Error loading config from %s: %v", configPath, err)
		}
	} else if len(os.Args) == 3 {
		// Два аргумента - прямые параметры (consumer_spec, provider_url)
		consumerSpecPath := os.Args[1]
		providerGitURL := os.Args[2]
		checker = NewContractCompatibilityChecker(consumerSpecPath, providerGitURL)
	} else {
		fmt.Println("Usage:")
		fmt.Println("  go run contract_compatibility_checker.go models.go [config_file]")
		fmt.Println("  go run contract_compatibility_checker.go models.go <consumer_spec_path> <provider_git_url>")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  go run contract_compatibility_checker.go models.go                    # Uses contract-tests.yaml")
		fmt.Println("  go run contract_compatibility_checker.go models.go custom-config.yaml")
		fmt.Println("  go run contract_compatibility_checker.go models.go ../api-specification/asyncapi.yml https://git.codemonsters.team/guides/wallet-balance/-/raw/main/api-specification/asyncapi.yml")
		os.Exit(1)
	}

	// Проверяем совместимость
	result, err := checker.CheckCompatibility()
	if err != nil {
		log.Fatalf("Error checking compatibility: %v", err)
	}

	// Генерируем отчет
	report := checker.GenerateReport(result)
	fmt.Println(report)

	// Определяем имя файла для JSON отчета
	jsonFileName := "compatibility_report.json"
	if checker.Config != nil && checker.Config.ContractTests.Settings.JSONReportFile != "" {
		jsonFileName = checker.Config.ContractTests.Settings.JSONReportFile
	}

	// Сохраняем результат в JSON (если не отключено в конфиге)
	saveJSON := true
	if checker.Config != nil {
		saveJSON = checker.Config.ContractTests.Settings.SaveJSONReport
	}

	if saveJSON {
		jsonResult, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Fatalf("Error marshaling result: %v", err)
		}

		err = os.WriteFile(jsonFileName, jsonResult, 0644)
		if err != nil {
			log.Fatalf("Error writing result file: %v", err)
		}

		fmt.Printf("Detailed JSON report saved to: %s\n", jsonFileName)
	}

	// Проверяем нужно ли игнорировать warnings
	ignoreWarnings := false
	if checker.Config != nil {
		ignoreWarnings = checker.Config.ContractTests.Settings.IgnoreWarnings
	}

	// Выходим с кодом 1 если есть breaking changes или warnings (если не игнорируются)
	if !result.IsCompatible || (!ignoreWarnings && len(result.Warnings) > 0) {
		if !result.IsCompatible {
			os.Exit(1)
		}
	}
}