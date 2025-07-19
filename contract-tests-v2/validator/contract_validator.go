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
		return nil, fmt.Errorf("failed to load config: %w", err)
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
		return nil, fmt.Errorf("failed to parse consumer spec: %w", err)
	}

	// Шаг 3: Парсинг спецификации поставщика
	providerSpec, err := v.parseSpecWithBaseDir(config.ContractTests.Provider.SpecURL, configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provider spec: %w", err)
	}

	// Для простоты берем первый канал из списка
	// В будущем можно расширить для валидации всех каналов
	if len(config.ContractTests.Consumer.Channels) == 0 {
		return nil, fmt.Errorf("no consumer channels specified")
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
		return nil, fmt.Errorf("channel validation failed: %w", err)
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
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	// Валидация обязательных полей
	if config.ContractTests.Consumer.SpecPath == "" {
		return nil, fmt.Errorf("consumer.spec_path is required")
	}
	if len(config.ContractTests.Consumer.Channels) == 0 {
		return nil, fmt.Errorf("consumer.channels is required")
	}
	if config.ContractTests.Provider.SpecURL == "" {
		return nil, fmt.Errorf("provider.spec_url is required")
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
			return nil, fmt.Errorf("failed to load spec from URL: %w", err)
		}
	} else {
		data, err = v.loadFromFile(pathOrURL)
		if err != nil {
			return nil, fmt.Errorf("failed to load spec from file: %w", err)
		}
	}

	spec, err := v.parser.ParseFromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spec: %w", err)
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
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}

// loadFromURL загружает данные по URL
func (v *ContractValidator) loadFromURL(url string) ([]byte, error) {
	resp, err := v.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

