package validator

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/codemonsters/contract-tests-v2/parser"
)

// ContractValidator основная структура валидатора контрактов
type ContractValidator struct {
	parser           *Parser
	channelValidator *ChannelValidator
	httpClient       *http.Client
}

// NewContractValidator создает новый экземпляр валидатора
func NewContractValidator() *ContractValidator {
	return &ContractValidator{
		parser:           NewParser(),
		channelValidator: NewChannelValidator(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Validate основная функция валидации с ROP паттерном
func (v *ContractValidator) Validate(configPath string) (*ValidationResult, error) {
	// Шаг 1: Загрузка конфигурации
	config, err := v.loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("CONFIG_ERROR: failed to load config - %w", err)
	}

	// Обновляем timeout если указан в конфигурации
	if config.ContractTests.Settings.Timeout > 0 {
		v.httpClient.Timeout = time.Duration(config.ContractTests.Settings.Timeout) * time.Second
	}

	// Получаем базовую директорию конфигурационного файла
	configDir := filepath.Dir(configPath)

	// Шаг 2: Парсинг спецификации потребителя
	consumerSpec, err := v.parseSpecWithBaseDir(config.ContractTests.Consumer.SpecPath, configDir)
	if err != nil {
		return nil, fmt.Errorf("PARSE_ERROR: failed to parse consumer spec at %s - %w", config.ContractTests.Consumer.SpecPath, err)
	}

	// Шаг 3: Парсинг спецификации поставщика
	providerPath := config.ContractTests.Provider.SpecURL
	if providerPath == "" {
		providerPath = config.ContractTests.Provider.SpecPath
	}
	providerSpec, err := v.parseSpecWithBaseDir(providerPath, configDir)
	if err != nil {
		return nil, fmt.Errorf("PARSE_ERROR: failed to parse provider spec at %s - %w", providerPath, err)
	}

	// Для простоты берем первый канал из списка
	// В будущем можно расширить для валидации всех каналов
	if len(config.ContractTests.Consumer.Channels) == 0 {
		return nil, newMissingFieldError("consumer.channels", "config.contract_tests")
	}
	consumerChannel := config.ContractTests.Consumer.Channels[0]

	// Шаг 4: Формирование структуры ContractValidate
	contractValidate := &ContractValidate{
		ConsumerChannelName: consumerChannel,
		ProviderSpec:        providerSpec,
		ConsumerSpec:        consumerSpec,
	}

	// Шаг 5: Валидация каналов
	_, err = v.channelValidator.ValidateChannels(contractValidate)
	if err != nil {
		return nil, fmt.Errorf("VALIDATION_ERROR: channel validation failed - %w", err)
	}

	// Формируем финальный результат валидации
	result := &ValidationResult{
		IsValid:             true,
		ConsumerChannelName: contractValidate.ConsumerChannelName,
		ConsumerSpec:        contractValidate.ConsumerSpec,
		ProviderSpec:        contractValidate.ProviderSpec,
		Errors:              []string{},
	}

	return result, nil
}

// loadConfig загружает конфигурацию из YAML файла
func (v *ContractValidator) loadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, newFileNotFoundError(configPath, "config.load")
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("YAML_PARSE_ERROR: failed to parse config YAML at %s - %w", configPath, err)
	}

	// Валидация обязательных полей
	if config.ContractTests.Consumer.SpecPath == "" {
		return nil, newMissingFieldError("consumer.spec_path", "config.contract_tests.consumer")
	}
	if len(config.ContractTests.Consumer.Channels) == 0 {
		return nil, newMissingFieldError("consumer.channels", "config.contract_tests.consumer")
	}
	if config.ContractTests.Provider.SpecURL == "" && config.ContractTests.Provider.SpecPath == "" {
		return nil, newMissingFieldError("provider.spec_url or provider.spec_path", "config.contract_tests.provider")
	}

	return &config, nil
}

// parseSpec парсит спецификацию из файла или URL
func (v *ContractValidator) parseSpec(pathOrURL string) (*parser.AsyncAPISpec, error) {
	var data []byte
	var err error

	// Определяем, это URL или локальный путь
	if strings.HasPrefix(pathOrURL, "http://") || strings.HasPrefix(pathOrURL, "https://") {
		data, err = v.loadFromURL(pathOrURL)
		if err != nil {
			return nil, fmt.Errorf("HTTP_ERROR: failed to load spec from URL %s - %w", pathOrURL, err)
		}
	} else {
		data, err = v.loadFromFile(pathOrURL)
		if err != nil {
			return nil, fmt.Errorf("FILE_NOT_FOUND: failed to load spec from file %s - %w", pathOrURL, err)
		}
	}

	spec, err := v.parser.ParseFromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("PARSE_ERROR: failed to parse spec from %s - %w", pathOrURL, err)
	}

	return spec, nil
}

// parseSpecWithBaseDir парсит спецификацию с учетом базовой директории для относительных путей
func (v *ContractValidator) parseSpecWithBaseDir(pathOrURL, baseDir string) (*parser.AsyncAPISpec, error) {
	var actualPath string
	
	// Определяем, это URL или локальный путь
	if strings.HasPrefix(pathOrURL, "http://") || strings.HasPrefix(pathOrURL, "https://") {
		// Для URL используем исходный parseSpec
		return v.parseSpec(pathOrURL)
	} else {
		// Для локальных файлов, если путь относительный, делаем его относительно baseDir
		if !filepath.IsAbs(pathOrURL) {
			actualPath = filepath.Join(baseDir, pathOrURL)
		} else {
			actualPath = pathOrURL
		}
		return v.parseSpec(actualPath)
	}
}

// loadFromFile загружает данные из локального файла
func (v *ContractValidator) loadFromFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, newFileNotFoundError(filePath, "file.read")
	}
	return data, nil
}

// loadFromURL загружает данные по URL
func (v *ContractValidator) loadFromURL(url string) ([]byte, error) {
	resp, err := v.httpClient.Get(url)
	if err != nil {
		// Проверяем на timeout
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded") {
			return nil, newTimeoutError("HTTP request", "url.fetch")
		}
		return nil, fmt.Errorf("HTTP_ERROR: failed to fetch URL %s - %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, newHTTPError(resp.StatusCode, url, "url.fetch")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HTTP_ERROR: failed to read response body from %s - %w", url, err)
	}

	return data, nil
}

