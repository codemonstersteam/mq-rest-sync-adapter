package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// SpecLoader предоставляет единый интерфейс для загрузки и кеширования спецификаций
type SpecLoader struct {
	config                *Config
	consumerSpecCache     *AsyncAPISpec
	providerSpecCache     *AsyncAPISpec
	protocolRegistryCache *ProtocolRegistry
	refExtractor          *RefExtractor
	mu                    sync.RWMutex
}

// NewSpecLoader создает новый экземпляр SpecLoader
func NewSpecLoader(config *Config) *SpecLoader {
	return &SpecLoader{
		config:       config,
		refExtractor: NewRefExtractor(),
	}
}

// GetConsumerSpec возвращает спецификацию потребителя, загружая её при первом вызове
func (sl *SpecLoader) GetConsumerSpec() (*AsyncAPISpec, error) {
	sl.mu.RLock()
	if sl.consumerSpecCache != nil {
		defer sl.mu.RUnlock()
		return sl.consumerSpecCache, nil
	}
	sl.mu.RUnlock()

	sl.mu.Lock()
	defer sl.mu.Unlock()

	// Проверяем повторно после получения lock
	if sl.consumerSpecCache != nil {
		return sl.consumerSpecCache, nil
	}

	content, err := os.ReadFile(sl.config.ContractTests.Consumer.SpecPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read consumer spec file: %w", err)
	}

	var spec AsyncAPISpec
	if err := yaml.Unmarshal(content, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse consumer spec: %w", err)
	}

	sl.consumerSpecCache = &spec
	return sl.consumerSpecCache, nil
}

// GetProviderSpec возвращает спецификацию провайдера, загружая её при первом вызове
func (sl *SpecLoader) GetProviderSpec() (*AsyncAPISpec, error) {
	sl.mu.RLock()
	if sl.providerSpecCache != nil {
		defer sl.mu.RUnlock()
		return sl.providerSpecCache, nil
	}
	sl.mu.RUnlock()

	sl.mu.Lock()
	defer sl.mu.Unlock()

	// Проверяем повторно после получения lock
	if sl.providerSpecCache != nil {
		return sl.providerSpecCache, nil
	}

	var content []byte
	var err error

	// Определяем путь к спецификации (приоритет у SpecURL, затем SpecPath)
	specPath := sl.config.ContractTests.Provider.SpecURL
	if specPath == "" {
		specPath = sl.config.ContractTests.Provider.SpecPath
	}

	if strings.HasPrefix(specPath, "http://") || strings.HasPrefix(specPath, "https://") {
		// Загружаем по HTTP
		content, err = sl.loadFromHTTP(specPath)
	} else {
		// Загружаем из файла
		content, err = os.ReadFile(specPath)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read provider spec: %w", err)
	}

	var spec AsyncAPISpec
	if err := yaml.Unmarshal(content, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse provider spec: %w", err)
	}

	sl.providerSpecCache = &spec
	return sl.providerSpecCache, nil
}

// GetProtocolRegistry возвращает реестр протоколов, загружая его при первом вызове
func (sl *SpecLoader) GetProtocolRegistry() (*ProtocolRegistry, error) {
	sl.mu.RLock()
	if sl.protocolRegistryCache != nil {
		defer sl.mu.RUnlock()
		return sl.protocolRegistryCache, nil
	}
	sl.mu.RUnlock()

	sl.mu.Lock()
	defer sl.mu.Unlock()

	// Проверяем повторно после получения lock
	if sl.protocolRegistryCache != nil {
		return sl.protocolRegistryCache, nil
	}

	registry, err := LoadProtocols()
	if err != nil {
		return nil, fmt.Errorf("failed to load protocols: %w", err)
	}

	sl.protocolRegistryCache = registry
	return sl.protocolRegistryCache, nil
}

// loadFromHTTP загружает спецификацию по HTTP
func (sl *SpecLoader) loadFromHTTP(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch spec from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch spec from %s: status %d", url, resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec content from %s: %w", url, err)
	}

	return content, nil
}

// GetRefExtractor возвращает экземпляр RefExtractor для извлечения имен из ссылок
func (sl *SpecLoader) GetRefExtractor() *RefExtractor {
	return sl.refExtractor
}

// ClearCache очищает кеш (полезно для тестирования)
func (sl *SpecLoader) ClearCache() {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	sl.consumerSpecCache = nil
	sl.providerSpecCache = nil
	sl.protocolRegistryCache = nil
}

// RefExtractor содержит утилиты для извлечения имен из ссылок AsyncAPI
type RefExtractor struct{}

// NewRefExtractor создает новый экземпляр извлекателя ссылок
func NewRefExtractor() *RefExtractor {
	return &RefExtractor{}
}

// ExtractRefName извлекает имя из ссылки AsyncAPI по заданному пути
func (re *RefExtractor) ExtractRefName(ref, expectedPath string) string {
	if ref == "" || expectedPath == "" {
		return ""
	}

	// Создаем полный префикс для поиска
	prefix := "#/" + expectedPath + "/"
	
	if len(ref) > len(prefix) && strings.HasPrefix(ref, prefix) {
		return ref[len(prefix):]
	}
	
	return ""
}

// ExtractServerName извлекает имя сервера из ссылки формата "#/servers/serverName"
func (re *RefExtractor) ExtractServerName(ref string) string {
	return re.ExtractRefName(ref, "servers")
}

// ExtractComponentName извлекает имя компонента сообщения из ссылки формата "#/components/messages/componentName"
func (re *RefExtractor) ExtractComponentName(ref string) string {
	return re.ExtractRefName(ref, "components/messages")
}

// ExtractMessageName извлекает имя сообщения из ссылки формата "#/components/messages/messageName"
// Это алиас для ExtractComponentName для обратной совместимости
func (re *RefExtractor) ExtractMessageName(ref string) string {
	return re.ExtractComponentName(ref)
}

// ExtractSchemaName извлекает имя схемы из ссылки формата "#/components/schemas/schemaName"
func (re *RefExtractor) ExtractSchemaName(ref string) string {
	return re.ExtractRefName(ref, "components/schemas")
}

// ExtractChannelName извлекает имя канала из ссылки формата "#/channels/channelName"
func (re *RefExtractor) ExtractChannelName(ref string) string {
	return re.ExtractRefName(ref, "channels")
}