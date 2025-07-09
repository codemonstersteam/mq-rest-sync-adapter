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

	// Проверяем различные аспекты совместимости
	c.checkSchemaCompatibility(consumerSpec, providerSpec, result)
	c.checkChannelCompatibility(consumerSpec, providerSpec, result)
	c.checkOperationCompatibility(consumerSpec, providerSpec, result)
	c.checkMessageCompatibility(consumerSpec, providerSpec, result)

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
	return gjson.ParseBytes(content), nil
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

	return gjson.ParseBytes(content), nil
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
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Type:        "field_type_changed",
				Field:       fmt.Sprintf("components.schemas.%s.properties.%s.type", schemaName, fieldNameStr),
				OldValue:    consumerType,
				NewValue:    providerType,
				Description: fmt.Sprintf("Field '%s' type changed from '%s' to '%s' in schema '%s'", fieldNameStr, consumerType, providerType, schemaName),
				Severity:    "high",
			})
		}

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