package channel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateChannels(t *testing.T) {
	t.Run("should build channel structure from consumer and provider specs", func(t *testing.T) {
		config := Config{
			Consumer: ConsumerConfig{
				SpecPath: "../testdata/channel_validator/consumer.yaml",
				Channel:  "user/signedup",
			},
			Provider: ProviderConfig{
				SpecPath: "../testdata/channel_validator/provider.yaml",
			},
		}

		validator := NewValidator()
		result, err := validator.ValidateChannels(config)

		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "user/signedup", result.ConsumerChannel.Name)
		assert.Equal(t, "amqp", result.ConsumerChannel.Protocol)
		assert.NotNil(t, result.ConsumerChannel.OutMessage)
		assert.NotNil(t, result.ConsumerChannel.InMessage)

		assert.NotEmpty(t, result.ProviderChannel.Name)
		assert.Equal(t, "amqp", result.ProviderChannel.Protocol)
		assert.NotNil(t, result.ProviderChannel.InMessage)
		assert.NotNil(t, result.ProviderChannel.OutMessage)
	})
}