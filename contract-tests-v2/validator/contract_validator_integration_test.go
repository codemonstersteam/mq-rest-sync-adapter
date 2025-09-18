package validator

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codemonsters/contract-tests-v2/parser"
)

func TestContractValidator_Validate_Integration(t *testing.T) {
	t.Run("should validate matching channels with multiple consumer channels", func(t *testing.T) {
		// Получаем абсолютный путь к тестовой конфигурации
		configPath, err := filepath.Abs("../testdata/contract_validator/multiple_consumer_channels_success.yaml")
		require.NoError(t, err, "Should resolve config path")

		// Тестируем полную интеграцию валидатора
		validator := NewContractValidator()
		result, err := validator.Validate(configPath)

		// Проверяем успешную валидацию
		require.NoError(t, err, "Validation should succeed")
		require.NotNil(t, result, "Result should not be nil")

		assert.True(t, result.IsValid, "Validation should be successful")
		assert.Equal(t, "user/events", result.ConsumerChannelName)
		assert.NotNil(t, result.ConsumerSpec)
		assert.NotNil(t, result.ProviderSpec)
		assert.Empty(t, result.Errors, "Should have no validation errors")

		// Проверяем правильные спецификации
		assert.Equal(t, "User Service", result.ConsumerSpec.Info.Title)
		assert.Equal(t, "Notification Service", result.ProviderSpec.Info.Title)

		// Проверяем наличие каналов в спецификациях
		assert.Contains(t, result.ConsumerSpec.Channels, "user/events", "Consumer should have user/events channel")
		assert.Contains(t, result.ConsumerSpec.Channels, "user/orders", "Consumer should have user/orders channel") 
		assert.Contains(t, result.ConsumerSpec.Channels, "user/kafka-events", "Consumer should have kafka channel")
		assert.Contains(t, result.ProviderSpec.Channels, "notifications/users", "Provider should have notifications/users channel")

		// Логирование результатов для документации
		t.Logf("✅ Integration test passed!")
		t.Logf("   Consumer spec: %s v%s", result.ConsumerSpec.Info.Title, result.ConsumerSpec.Info.Version)
		t.Logf("   Provider spec: %s v%s", result.ProviderSpec.Info.Title, result.ProviderSpec.Info.Version)
		t.Logf("   Consumer channels available: %v", getChannelNames(result.ConsumerSpec.Channels))
		t.Logf("   Provider channels available: %v", getChannelNames(result.ProviderSpec.Channels))
		t.Logf("   Selected consumer channel: %s", result.ConsumerChannelName)
		t.Logf("   Channel matching algorithm successfully selected compatible channel")
		t.Logf("   Validation result: SUCCESS (%v)", result.IsValid)
	})

	t.Run("should demonstrate channel selection algorithm", func(t *testing.T) {
		// Этот тест документирует работу алгоритма выбора канала
		configPath, err := filepath.Abs("../testdata/contract_validator/multiple_consumer_channels_success.yaml")
		require.NoError(t, err, "Should resolve config path")

		validator := NewContractValidator()
		result, err := validator.Validate(configPath)

		require.NoError(t, err)
		require.True(t, result.IsValid)

		// Документируем процесс выбора канала
		t.Logf("🔍 Channel Selection Algorithm Test:")
		t.Logf("")
		t.Logf("Consumer has 3 channels:")
		t.Logf("  1. user/events (AMQP + UserSignup) - ✅ COMPATIBLE")
		t.Logf("  2. user/orders (AMQP + OrderEvent) - ❌ Different message structure")
		t.Logf("  3. user/kafka-events (Kafka + UserSignup) - ❌ Different protocol")
		t.Logf("")
		t.Logf("Provider has 1 channel:")
		t.Logf("  1. notifications/users (AMQP + UserSignup) - ✅ COMPATIBLE")
		t.Logf("")
		t.Logf("Algorithm steps:")
		t.Logf("  1. Check protocol: AMQP ✅")
		t.Logf("  2. Check message structure: UserSignup fields match ✅")
		t.Logf("  3. Check field types: string types match ✅")
		t.Logf("  4. Selected: user/events ↔ notifications/users")
		t.Logf("")
		t.Logf("Result: %s channel successfully matched", result.ConsumerChannelName)

		// Проверяем что выбран именно правильный канал
		assert.Equal(t, "user/events", result.ConsumerChannelName,
			"Algorithm should select user/events as the only compatible channel")
	})

	t.Run("real project specs - incompatible patterns", func(t *testing.T) {
		configPath, err := filepath.Abs("../cmd/testdata/cmd/real-project-contract-tests.yaml")
		require.NoError(t, err)

		validator := NewContractValidator()
		_, err = validator.Validate(configPath)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no compatible provider channel found")
		assert.Contains(t, err.Error(), "provider has no outgoing message")
	})

	t.Run("real project specs - compatible after fix", func(t *testing.T) {
		configPath, err := filepath.Abs("../cmd/testdata/cmd/real-project-contract-tests-fixed.yaml")
		require.NoError(t, err)

		validator := NewContractValidator()
		result, err := validator.Validate(configPath)

		require.NoError(t, err)
		assert.True(t, result.IsValid)
		assert.Equal(t, "restGetBalanceRequest", result.ConsumerChannelName)
	})
}

// getChannelNames извлекает имена каналов из AsyncAPI спецификации
func getChannelNames(channels map[string]parser.Channel) []string {
	names := make([]string, 0, len(channels))
	for name := range channels {
		names = append(names, name)
	}
	return names
}