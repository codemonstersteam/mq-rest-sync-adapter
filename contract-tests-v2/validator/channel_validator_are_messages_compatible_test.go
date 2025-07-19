package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChannelValidator_areMessagesCompatible(t *testing.T) {
	validator := NewChannelValidator()

	t.Run("should return true for identical messages", func(t *testing.T) {
		msg1 := &MessageInfo{
			Name:        "UserSignup",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userId": map[string]interface{}{"type": "string"},
					"email":  map[string]interface{}{"type": "string"},
				},
				"required": []interface{}{"userId", "email"},
			},
		}

		msg2 := &MessageInfo{
			Name:        "UserSignup",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userId": map[string]interface{}{"type": "string"},
					"email":  map[string]interface{}{"type": "string"},
				},
				"required": []interface{}{"userId", "email"},
			},
		}

		result := validator.areMessagesCompatible(msg1, msg2)
		assert.True(t, result, "Identical messages should be compatible")
	})

	t.Run("should return false for different field names", func(t *testing.T) {
		msg1 := &MessageInfo{
			Name:        "UserSignup",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userId": map[string]interface{}{"type": "string"},
					"email":  map[string]interface{}{"type": "string"},
				},
				"required": []interface{}{"userId", "email"},
			},
		}

		msg2 := &MessageInfo{
			Name:        "OrderEvent",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"orderId":    map[string]interface{}{"type": "string"},
					"customerId": map[string]interface{}{"type": "string"},
				},
				"required": []interface{}{"orderId", "customerId"},
			},
		}

		result := validator.areMessagesCompatible(msg1, msg2)
		assert.False(t, result, "Messages with different field names should not be compatible")
	})

	t.Run("should return false for different number of required fields", func(t *testing.T) {
		msg1 := &MessageInfo{
			Name:        "UserSignup",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userId": map[string]interface{}{"type": "string"},
					"email":  map[string]interface{}{"type": "string"},
					"name":   map[string]interface{}{"type": "string"},
				},
				"required": []interface{}{"userId", "email", "name"},
			},
		}

		msg2 := &MessageInfo{
			Name:        "ProductEvent",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"productId": map[string]interface{}{"type": "string"},
					"category":  map[string]interface{}{"type": "string"},
				},
				"required": []interface{}{"productId", "category"},
			},
		}

		result := validator.areMessagesCompatible(msg1, msg2)
		assert.False(t, result, "Messages with different number of required fields should not be compatible")
	})

	t.Run("should return false for different field types", func(t *testing.T) {
		msg1 := &MessageInfo{
			Name:        "UserSignup",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userId": map[string]interface{}{"type": "string"},
					"age":    map[string]interface{}{"type": "number"},
				},
				"required": []interface{}{"userId", "age"},
			},
		}

		msg2 := &MessageInfo{
			Name:        "UserSignup",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userId": map[string]interface{}{"type": "string"},
					"age":    map[string]interface{}{"type": "string"}, // different type!
				},
				"required": []interface{}{"userId", "age"},
			},
		}

		result := validator.areMessagesCompatible(msg1, msg2)
		assert.False(t, result, "Messages with different field types should not be compatible")
	})

	t.Run("should return false for nil messages", func(t *testing.T) {
		msg1 := &MessageInfo{
			Name:        "UserSignup",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userId": map[string]interface{}{"type": "string"},
				},
				"required": []interface{}{"userId"},
			},
		}

		result1 := validator.areMessagesCompatible(nil, msg1)
		assert.False(t, result1, "Nil message should not be compatible")

		result2 := validator.areMessagesCompatible(msg1, nil)
		assert.False(t, result2, "Message compared with nil should not be compatible")

		result3 := validator.areMessagesCompatible(nil, nil)
		assert.False(t, result3, "Both nil messages should not be compatible")
	})

	t.Run("should return false for messages without properties", func(t *testing.T) {
		msg1 := &MessageInfo{
			Name:        "UserSignup",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type":     "object",
				"required": []interface{}{"userId"},
			},
		}

		msg2 := &MessageInfo{
			Name:        "UserSignup",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userId": map[string]interface{}{"type": "string"},
				},
				"required": []interface{}{"userId"},
			},
		}

		result := validator.areMessagesCompatible(msg1, msg2)
		assert.False(t, result, "Message without properties should not be compatible")
	})

	t.Run("should return false for properties without type", func(t *testing.T) {
		msg1 := &MessageInfo{
			Name:        "UserSignup",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userId": map[string]interface{}{"type": "string"},
					"email":  map[string]interface{}{}, // no type!
				},
				"required": []interface{}{"userId", "email"},
			},
		}

		msg2 := &MessageInfo{
			Name:        "UserSignup",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userId": map[string]interface{}{"type": "string"},
					"email":  map[string]interface{}{"type": "string"},
				},
				"required": []interface{}{"userId", "email"},
			},
		}

		result := validator.areMessagesCompatible(msg1, msg2)
		assert.False(t, result, "Property without type should make messages incompatible")
	})

	t.Run("should return false for convertSchema bug with []string required", func(t *testing.T) {
		// Воспроизводим проблему convertSchema где required передается как []string
		userSignup := &MessageInfo{
			Name:        "UserSignup",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userId": map[string]interface{}{"type": "string"},
					"email":  map[string]interface{}{"type": "string"},
					"name":   map[string]interface{}{"type": "string"},
				},
				"required": []string{"userId", "email", "name"}, // []string!
			},
		}

		orderEvent := &MessageInfo{
			Name:        "OrderEvent", 
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"orderId":    map[string]interface{}{"type": "string"},
					"customerId": map[string]interface{}{"type": "string"},
					"amount":     map[string]interface{}{"type": "number"},
				},
				"required": []string{"orderId", "customerId", "amount"}, // []string!
			},
		}

		result := validator.areMessagesCompatible(userSignup, orderEvent)
		assert.False(t, result, "Should handle []string required and return false for different messages")
	})

	t.Run("should return false for same structure but different field type", func(t *testing.T) {
		// Consumer message: UserSignup с полем age как number
		consumerMessage := &MessageInfo{
			Name:        "UserSignup",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userId": map[string]interface{}{"type": "string"},
					"email":  map[string]interface{}{"type": "string"}, 
					"age":    map[string]interface{}{"type": "number"}, // number
				},
				"required": []string{"userId", "email", "age"},
			},
		}

		// Provider message: те же поля, но age как string
		providerMessage := &MessageInfo{
			Name:        "UserSignup",
			ContentType: "application/json", 
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userId": map[string]interface{}{"type": "string"},
					"email":  map[string]interface{}{"type": "string"},
					"age":    map[string]interface{}{"type": "string"}, // string вместо number!
				},
				"required": []string{"userId", "email", "age"},
			},
		}

		result := validator.areMessagesCompatible(consumerMessage, providerMessage)
		assert.False(t, result, "Messages with same structure but different field type should not be compatible")
	})

	t.Run("should return true when provider has additional optional field", func(t *testing.T) {
		// Consumer message: знает только об обязательных полях userId и email
		consumerMessage := &MessageInfo{
			Name:        "UserSignup",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userId": map[string]interface{}{"type": "string"},
					"email":  map[string]interface{}{"type": "string"},
				},
				"required": []string{"userId", "email"},
			},
		}

		// Provider message: имеет те же обязательные поля + дополнительное необязательное поле
		providerMessage := &MessageInfo{
			Name:        "UserSignup",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userId":    map[string]interface{}{"type": "string"},
					"email":     map[string]interface{}{"type": "string"},
					"timestamp": map[string]interface{}{"type": "string"}, // дополнительное поле
				},
				"required": []string{"userId", "email"}, // timestamp не обязательное
			},
		}

		result := validator.areMessagesCompatible(consumerMessage, providerMessage)
		assert.True(t, result, "Provider with additional optional field should be compatible - consumer doesn't need to know about extra fields")
	})

	t.Run("should return false when consumer missing required field from provider", func(t *testing.T) {
		// Consumer message: имеет только 1 обязательное поле userId + 1 необязательное email
		consumerMessage := &MessageInfo{
			Name:        "UserSignup",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userId": map[string]interface{}{"type": "string"},
					"email":  map[string]interface{}{"type": "string"}, // есть в properties, но не обязательное
				},
				"required": []string{"userId"}, // только userId обязательное
			},
		}

		// Provider message: имеет 2 обязательных поля + 1 необязательное
		providerMessage := &MessageInfo{
			Name:        "UserSignup",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userId":    map[string]interface{}{"type": "string"},
					"email":     map[string]interface{}{"type": "string"}, // обязательное у провайдера!
					"timestamp": map[string]interface{}{"type": "string"}, // необязательное
				},
				"required": []string{"userId", "email"}, // email обязательное у провайдера
			},
		}

		result := validator.areMessagesCompatible(consumerMessage, providerMessage)
		assert.False(t, result, "Consumer missing required field 'email' - incompatible because provider requires email but consumer treats it as optional")
	})
}