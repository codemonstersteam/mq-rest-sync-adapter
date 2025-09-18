package validator

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContractValidator_OfflineValidation(t *testing.T) {
	// Тест проверяет валидацию контрактов с локальными спецификациями
	// без необходимости доступа к интернету
	t.Run("successful validation with local specifications", func(t *testing.T) {
		// Arrange
		validator := NewContractValidator()
		configPath := filepath.Join("..", "testdata", "contract_validator", "contract-tests-local.yaml")
		
		// Проверяем, что все необходимые файлы существуют
		require.FileExists(t, configPath, "Config file should exist")
		require.FileExists(t, filepath.Join("..", "testdata", "contract_validator", "consumer_local.yaml"), "Consumer spec should exist")
		require.FileExists(t, filepath.Join("..", "testdata", "contract_validator", "provider_external.yaml"), "Provider spec should exist")

		// Act
		result, err := validator.Validate(configPath)

		// Assert
		if err != nil {
			// Добавим отладку для понимания проблемы
			t.Logf("Validation error: %v", err)
		}
		require.NoError(t, err, "Validation should not return error")
		require.NotNil(t, result, "Result should not be nil")
		
		// Проверяем основные поля результата
		assert.True(t, result.IsValid, "Validation should be successful")
		assert.Equal(t, "restGetBalanceRequest", result.ConsumerChannelName)
		assert.NotNil(t, result.ConsumerSpec, "Consumer spec should be loaded")
		assert.NotNil(t, result.ProviderSpec, "Provider spec should be loaded")
		assert.Empty(t, result.Errors, "Should have no errors")
		
		// Проверяем, что спецификации загружены корректно
		assert.Equal(t, "3.0.0", result.ConsumerSpec.AsyncAPI)
		assert.Equal(t, "MQ Adapter for REST wallet balance service", result.ConsumerSpec.Info.Title)
		assert.Equal(t, "3.0.0", result.ProviderSpec.AsyncAPI)
		assert.Equal(t, "Wallet Balance Service", result.ProviderSpec.Info.Title)
		
		// Проверяем наличие каналов
		assert.NotEmpty(t, result.ConsumerSpec.Channels, "Consumer should have channels")
		assert.NotEmpty(t, result.ProviderSpec.Channels, "Provider should have channels")
		
		// Проверяем конкретный канал потребителя
		consumerChannel, exists := result.ConsumerSpec.Channels["restGetBalanceRequest"]
		assert.True(t, exists, "Consumer should have restGetBalanceRequest channel")
		assert.Equal(t, "/api/client/{Api-clientId}/wallet/balance", consumerChannel.Address)
		
		// Проверяем, что провайдер имеет совместимый канал
		providerChannel, exists := result.ProviderSpec.Channels["walletBalanceRequest"]
		assert.True(t, exists, "Provider should have walletBalanceRequest channel")
		assert.Equal(t, "/api/client/{clientId}/wallet/balance", providerChannel.Address)
		
		// Проверяем наличие операций
		assert.NotEmpty(t, result.ConsumerSpec.Operations, "Consumer should have operations")
		assert.NotEmpty(t, result.ProviderSpec.Operations, "Provider should have operations")
		
		// Проверяем операцию потребителя
		consumerOp, exists := result.ConsumerSpec.Operations["getBalanceRequest"]
		assert.True(t, exists, "Consumer should have getBalanceRequest operation")
		assert.Equal(t, "send", consumerOp.Action)
		assert.NotNil(t, consumerOp.Reply, "Operation should have reply")
		
		// Проверяем операцию провайдера
		providerOp, exists := result.ProviderSpec.Operations["receiveBalanceRequest"]
		assert.True(t, exists, "Provider should have receiveBalanceRequest operation")
		assert.Equal(t, "receive", providerOp.Action)
	})
}