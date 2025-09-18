package validator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	
	"github.com/stretchr/testify/require"
	"github.com/codemonsters/contract-tests-v2/parser"
)

func TestPayloadDebug_RealVsMock(t *testing.T) {
	validator := NewChannelValidator()
	
	t.Run("compare payload extraction from real specs vs mock specs", func(t *testing.T) {
		// 1. Load real specifications
		consumerPath := filepath.Join("..", "testdata", "contract_validator", "consumer_local.yaml")
		consumerData, err := os.ReadFile(consumerPath)
		require.NoError(t, err, "Should read consumer spec file")
		
		parserInstance := parser.New()
		consumerSpec, err := parserInstance.ParseFromBytes(consumerData)
		require.NoError(t, err, "Should parse consumer spec")
		
		providerPath := filepath.Join("..", "testdata", "contract_validator", "provider_external.yaml")
		providerData, err := os.ReadFile(providerPath)
		require.NoError(t, err, "Should read provider spec file")
		
		providerSpec, err := parserInstance.ParseFromBytes(providerData)
		require.NoError(t, err, "Should parse provider spec")
		
		// 2. Extract response messages from real specs
		contractValidate := &ContractValidate{
			ConsumerChannelName: "restGetBalanceRequest",
			ProviderSpec:        providerSpec,
			ConsumerSpec:        consumerSpec,
		}
		
		consumerChannel, err := validator.extractConsumerChannel(contractValidate)
		require.NoError(t, err, "Should extract consumer channel")
		require.NotNil(t, consumerChannel.InMessage, "Consumer should have InMessage")
		
		// Find provider response message
		_, providerOutMessage, err := validator.extractProviderMessages(providerSpec, "walletBalanceRequest")
		require.NoError(t, err, "Should extract provider messages")
		require.NotNil(t, providerOutMessage, "Provider should have OutMessage")
		
		realConsumerMsg := consumerChannel.InMessage
		realProviderMsg := providerOutMessage
		
		t.Logf("=== REAL SPECS COMPARISON ===")
		t.Logf("Consumer response message name: %s", realConsumerMsg.Name)
		t.Logf("Provider response message name: %s", realProviderMsg.Name)
		
		// Pretty print payloads for comparison
		consumerPayloadJSON, _ := json.MarshalIndent(realConsumerMsg.Payload, "", "  ")
		providerPayloadJSON, _ := json.MarshalIndent(realProviderMsg.Payload, "", "  ")
		
		t.Logf("Consumer payload:\n%s", string(consumerPayloadJSON))
		t.Logf("Provider payload:\n%s", string(providerPayloadJSON))
		
		// Test compatibility
		compatible := validator.areMessagesCompatible(realConsumerMsg, realProviderMsg, consumerSpec, providerSpec)
		t.Logf("Real specs compatible: %v", compatible)
		
		// 3. Create mock specs like in the working test
		mockConsumerSpec := &parser.AsyncAPISpec{
			Components: &parser.Components{
				Schemas: map[string]parser.Schema{
					"walletBalanceData": {
						Type: "object",
						Required: []string{"createTime", "operationId", "balance"},
						Properties: map[string]parser.Schema{
							"createTime": {
								Type:   "integer",
								Format: "int64",
							},
							"operationId": {
								Type: "string",
							},
							"balance": {
								Ref: "#/components/schemas/balanceInfo",
							},
						},
					},
					"balanceInfo": {
						Type: "object", 
						Required: []string{"value", "currency"},
						Properties: map[string]parser.Schema{
							"value": {
								Type:   "number",
								Format: "double",
							},
							"currency": {
								Type: "string",
							},
						},
					},
				},
			},
		}
		
		mockProviderSpec := &parser.AsyncAPISpec{
			Components: &parser.Components{
				Schemas: map[string]parser.Schema{
					"walletBalanceData": {
						Type: "object",
						Required: []string{"createTime", "operationId", "balance"},
						Properties: map[string]parser.Schema{
							"createTime": {
								Type:   "integer", 
								Format: "int64",
							},
							"operationId": {
								Type: "string",
							},
							"balance": {
								Ref: "#/components/schemas/balanceInfo",
							},
						},
					},
					"balanceInfo": {
						Type: "object",
						Required: []string{"value", "currency"},
						Properties: map[string]parser.Schema{
							"value": {
								Type:   "number",
								Format: "double",
							},
							"currency": {
								Type: "string",
							},
						},
					},
				},
			},
		}
		
		// Mock messages like in the working test
		mockConsumerMsg := &MessageInfo{
			Name:        "PsGetBalanceResponse",
			ContentType: "application/json", 
			Payload: map[string]interface{}{
				"type": "object",
				"required": []string{"status", "actualTimestamp", "data"},
				"properties": map[string]interface{}{
					"status": map[string]interface{}{
						"type": "string",
						"enum": []interface{}{"success", "error"},
					},
					"actualTimestamp": map[string]interface{}{
						"type":   "integer",
						"format": "int64",
					},
					"data": map[string]interface{}{
						"$ref": "#/components/schemas/walletBalanceData",
					},
				},
			},
		}
		
		mockProviderMsg := &MessageInfo{
			Name:        "GetBalanceResponse",
			ContentType: "application/json",
			Payload: map[string]interface{}{
				"type": "object", 
				"required": []string{"status", "actualTimestamp", "data"},
				"properties": map[string]interface{}{
					"status": map[string]interface{}{
						"type": "string",
						"enum": []interface{}{"success", "error"},
					},
					"actualTimestamp": map[string]interface{}{
						"type":   "integer",
						"format": "int64",
					},
					"data": map[string]interface{}{
						"$ref": "#/components/schemas/walletBalanceData",
					},
				},
			},
		}
		
		t.Logf("\n=== MOCK SPECS COMPARISON ===")
		t.Logf("Mock consumer response message name: %s", mockConsumerMsg.Name)
		t.Logf("Mock provider response message name: %s", mockProviderMsg.Name)
		
		// Pretty print mock payloads
		mockConsumerPayloadJSON, _ := json.MarshalIndent(mockConsumerMsg.Payload, "", "  ")
		mockProviderPayloadJSON, _ := json.MarshalIndent(mockProviderMsg.Payload, "", "  ")
		
		t.Logf("Mock consumer payload:\n%s", string(mockConsumerPayloadJSON))
		t.Logf("Mock provider payload:\n%s", string(mockProviderPayloadJSON))
		
		// Test mock compatibility
		mockCompatible := validator.areMessagesCompatible(mockConsumerMsg, mockProviderMsg, mockConsumerSpec, mockProviderSpec)
		t.Logf("Mock specs compatible: %v", mockCompatible)
		
		t.Logf("\n=== COMPARISON ANALYSIS ===")
		t.Logf("Real specs compatible: %v", compatible)
		t.Logf("Mock specs compatible: %v", mockCompatible)
		
		// Compare data field specifically
		realConsumerDataField := realConsumerMsg.Payload["properties"].(map[string]interface{})["data"]
		realProviderDataField := realProviderMsg.Payload["properties"].(map[string]interface{})["data"]
		mockConsumerDataField := mockConsumerMsg.Payload["properties"].(map[string]interface{})["data"]
		mockProviderDataField := mockProviderMsg.Payload["properties"].(map[string]interface{})["data"]
		
		t.Logf("Real consumer data field: %+v", realConsumerDataField)
		t.Logf("Real provider data field: %+v", realProviderDataField)
		t.Logf("Mock consumer data field: %+v", mockConsumerDataField)
		t.Logf("Mock provider data field: %+v", mockProviderDataField)
		
		// Test individual field compatibility
		realDataCompatible := validator.arePropertyTypesCompatible(realConsumerDataField, realProviderDataField, consumerSpec, providerSpec)
		mockDataCompatible := validator.arePropertyTypesCompatible(mockConsumerDataField, mockProviderDataField, mockConsumerSpec, mockProviderSpec)
		
		t.Logf("Real data field compatible: %v", realDataCompatible)
		t.Logf("Mock data field compatible: %v", mockDataCompatible)
	})
}