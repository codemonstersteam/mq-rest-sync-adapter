package validator

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/stretchr/testify/require"
	"github.com/codemonsters/contract-tests-v2/parser"
)

func TestMessageExtraction_RealSpecs(t *testing.T) {
	validator := NewChannelValidator()
	
	t.Run("extract messages from real consumer and provider specs", func(t *testing.T) {
		// Загружаем реальные спецификации
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
		
		t.Logf("Consumer spec loaded: %s v%s", consumerSpec.Info.Title, consumerSpec.AsyncAPI)
		t.Logf("Provider spec loaded: %s v%s", providerSpec.Info.Title, providerSpec.AsyncAPI)
		
		// Создаем ContractValidate как в реальном интеграционном тесте
		contractValidate := &ContractValidate{
			ConsumerChannelName: "restGetBalanceRequest",
			ProviderSpec:        providerSpec,
			ConsumerSpec:        consumerSpec,
		}
		
		// 1. Тестируем извлечение канала потребителя
		consumerChannel, err := validator.extractConsumerChannel(contractValidate)
		require.NoError(t, err, "Should extract consumer channel")
		require.NotNil(t, consumerChannel, "Consumer channel should not be nil")
		
		t.Logf("Consumer channel extracted:")
		t.Logf("  Name: %s", consumerChannel.Name)
		t.Logf("  Protocol: %s", consumerChannel.Protocol)
		t.Logf("  Has OutMessage: %v", consumerChannel.OutMessage != nil)
		t.Logf("  Has InMessage: %v", consumerChannel.InMessage != nil)
		
		if consumerChannel.OutMessage != nil {
			t.Logf("  OutMessage name: %s", consumerChannel.OutMessage.Name)
			t.Logf("  OutMessage contentType: %s", consumerChannel.OutMessage.ContentType)
			// t.Logf("  OutMessage payload keys: %v", getMapKeys(consumerChannel.OutMessage.Payload))
			if reqFields, _ := validator.getRequiredFields(consumerChannel.OutMessage.Payload); len(reqFields) > 0 {
				t.Logf("  OutMessage required fields: %v", reqFields)
			}
		}
		
		if consumerChannel.InMessage != nil {
			t.Logf("  InMessage name: %s", consumerChannel.InMessage.Name)
			t.Logf("  InMessage contentType: %s", consumerChannel.InMessage.ContentType)
			// t.Logf("  InMessage payload keys: %v", getMapKeys(consumerChannel.InMessage.Payload))
			if reqFields, _ := validator.getRequiredFields(consumerChannel.InMessage.Payload); len(reqFields) > 0 {
				t.Logf("  InMessage required fields: %v", reqFields)
			}
		}
		
		// 2. Тестируем поиск подходящего канала провайдера
		providerChannel, err := validator.findMatchingProviderChannel(contractValidate, consumerChannel)
		
		if err != nil {
			t.Logf("Provider channel matching failed: %v", err)
			
			// Детальная диагностика всех каналов провайдера
			t.Logf("Provider channels analysis:")
			for channelName, channel := range providerSpec.Channels {
				protocol, protocolErr := validator.extractChannelProtocol(providerSpec, &channel)
				t.Logf("  Channel: %s", channelName)
				if protocolErr != nil {
					t.Logf("    Protocol error: %v", protocolErr)
				} else {
					t.Logf("    Protocol: %s", protocol)
					if protocol == consumerChannel.Protocol {
						t.Logf("    *** PROTOCOL MATCH ***")
						
						// Извлекаем сообщения провайдера для этого канала
						inMessage, outMessage, msgErr := validator.extractProviderMessages(providerSpec, channelName)
						if msgErr != nil {
							t.Logf("    Messages extraction error: %v", msgErr)
						} else {
							t.Logf("    Has InMessage: %v", inMessage != nil)
							t.Logf("    Has OutMessage: %v", outMessage != nil)
							
							if inMessage != nil {
								t.Logf("    InMessage name: %s", inMessage.Name)
								if reqFields, _ := validator.getRequiredFields(inMessage.Payload); len(reqFields) > 0 {
									t.Logf("    InMessage required: %v", reqFields)
								}
							}
							
							if outMessage != nil {
								t.Logf("    OutMessage name: %s", outMessage.Name)
								if reqFields, _ := validator.getRequiredFields(outMessage.Payload); len(reqFields) > 0 {
									t.Logf("    OutMessage required: %v", reqFields)
								}
								
								// Тестируем совместимость response сообщений
								if consumerChannel.InMessage != nil && outMessage != nil {
									compatible := validator.areMessagesCompatible(
										consumerChannel.InMessage, outMessage, 
										consumerSpec, providerSpec)
									t.Logf("    Response messages compatible: %v", compatible)
									
									if !compatible {
										t.Logf("    INCOMPATIBILITY DETAILS:")
										consumerReq, _ := validator.getRequiredFields(consumerChannel.InMessage.Payload)
										providerReq, _ := validator.getRequiredFields(outMessage.Payload)
										t.Logf("      Consumer response required: %v", consumerReq)
										t.Logf("      Provider response required: %v", providerReq)
										
										// Проверяем properties
										_, ok1 := consumerChannel.InMessage.Payload["properties"].(map[string]interface{})
										_, ok2 := outMessage.Payload["properties"].(map[string]interface{})
										t.Logf("      Consumer has properties: %v, Provider has properties: %v", ok1, ok2)
									}
								}
							}
						}
					}
				}
			}
		} else {
			require.NotNil(t, providerChannel, "Provider channel should not be nil")
			t.Logf("Provider channel matched: %s", providerChannel.Name)
		}
		
		// Дополнительно: проверим операции потребителя
		t.Logf("\nConsumer operations analysis:")
		for opName, operation := range consumerSpec.Operations {
			t.Logf("  Operation: %s (action: %s)", opName, operation.Action)
			if operation.Channel.Ref != "" {
				t.Logf("    Channel ref: %s", operation.Channel.Ref)
			}
			if len(operation.Messages) > 0 {
				t.Logf("    Messages count: %d", len(operation.Messages))
				for i, msg := range operation.Messages {
					t.Logf("      Message[%d] ref: %s", i, msg.Ref)
				}
			}
			if operation.Reply != nil {
				t.Logf("    Has reply with %d messages", len(operation.Reply.Messages))
				for i, msg := range operation.Reply.Messages {
					t.Logf("      Reply[%d] ref: %s", i, msg.Ref)
				}
			}
		}
		
		// Проверим операции провайдера
		t.Logf("\nProvider operations analysis:")
		for opName, operation := range providerSpec.Operations {
			t.Logf("  Operation: %s (action: %s)", opName, operation.Action)
			if operation.Channel.Ref != "" {
				t.Logf("    Channel ref: %s", operation.Channel.Ref)
			}
			if len(operation.Messages) > 0 {
				t.Logf("    Messages count: %d", len(operation.Messages))
				for i, msg := range operation.Messages {
					t.Logf("      Message[%d] ref: %s", i, msg.Ref)
				}
			}
			if operation.Reply != nil {
				t.Logf("    Has reply with %d messages", len(operation.Reply.Messages))
				for i, msg := range operation.Reply.Messages {
					t.Logf("      Reply[%d] ref: %s", i, msg.Ref)
				}
			}
		}
	})
}

