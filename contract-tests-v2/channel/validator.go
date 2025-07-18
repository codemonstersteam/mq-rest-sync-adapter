package channel

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Consumer ConsumerConfig
	Provider ProviderConfig
}

type ConsumerConfig struct {
	SpecPath string
	Channel  string
}

type ProviderConfig struct {
	SpecPath string
}

type ValidationResult struct {
	ConsumerChannel Channel
	ProviderChannel Channel
}

type Channel struct {
	Name       string
	Protocol   string
	OutMessage *Message
	InMessage  *Message
}

type Message struct {
	Name        string
	ContentType string
	Payload     map[string]interface{}
}

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateChannels(config Config) (*ValidationResult, error) {
	consumerSpec, err := loadAsyncAPISpec(config.Consumer.SpecPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load consumer spec: %w", err)
	}

	providerSpec, err := loadAsyncAPISpec(config.Provider.SpecPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load provider spec: %w", err)
	}

	consumerChannel, err := extractChannelInfo(consumerSpec, config.Consumer.Channel)
	if err != nil {
		return nil, fmt.Errorf("failed to extract consumer channel: %w", err)
	}

	providerChannel, err := findMatchingProviderChannel(providerSpec, consumerChannel)
	if err != nil {
		return nil, fmt.Errorf("failed to find matching provider channel: %w", err)
	}

	return &ValidationResult{
		ConsumerChannel: *consumerChannel,
		ProviderChannel: *providerChannel,
	}, nil
}

func loadAsyncAPISpec(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var spec map[string]interface{}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	return spec, nil
}

func extractChannelInfo(spec map[string]interface{}, channelName string) (*Channel, error) {
	channels, ok := spec["channels"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("channels not found in spec")
	}

	channelData, ok := channels[channelName].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("channel %s not found", channelName)
	}

	protocol := extractProtocol(spec, channelData)
	
	operations := spec["operations"].(map[string]interface{})
	outMessage, inMessage := extractMessagesFromOperations(spec, operations, channelName)

	return &Channel{
		Name:       channelName,
		Protocol:   protocol,
		OutMessage: outMessage,
		InMessage:  inMessage,
	}, nil
}

func extractProtocol(spec map[string]interface{}, channelData map[string]interface{}) string {
	if servers, ok := channelData["servers"].([]interface{}); ok && len(servers) > 0 {
		if serverRef, ok := servers[0].(map[string]interface{})["$ref"].(string); ok {
			serverPath := serverRef[len("#/servers/"):]
			if serversData, ok := spec["servers"].(map[string]interface{}); ok {
				if server, ok := serversData[serverPath].(map[string]interface{}); ok {
					if protocol, ok := server["protocol"].(string); ok {
						return protocol
					}
				}
			}
		}
	}
	return ""
}

func extractMessagesFromOperations(spec map[string]interface{}, operations map[string]interface{}, channelName string) (*Message, *Message) {
	var outMessage, inMessage *Message

	for _, op := range operations {
		operation, ok := op.(map[string]interface{})
		if !ok {
			continue
		}

		if channel, ok := operation["channel"].(map[string]interface{}); ok {
			if ref, ok := channel["$ref"].(string); ok && containsChannel(ref, channelName) {
				if action, ok := operation["action"].(string); ok && action == "send" {
					if messages, ok := operation["messages"].([]interface{}); ok && len(messages) > 0 {
						outMessage = extractMessage(spec, messages[0])
					}
					if reply, ok := operation["reply"].(map[string]interface{}); ok {
						if replyMessages, ok := reply["messages"].([]interface{}); ok && len(replyMessages) > 0 {
							inMessage = extractMessage(spec, replyMessages[0])
						}
					}
				}
			}
		}
	}

	return outMessage, inMessage
}

func containsChannel(ref, channelName string) bool {
	escapedName := escapeChannelName(channelName)
	return ref == "#/channels/"+escapedName
}

func escapeChannelName(name string) string {
	result := ""
	for _, c := range name {
		if c == '/' {
			result += "~1"
		} else {
			result += string(c)
		}
	}
	return result
}

func unescapeChannelName(name string) string {
	return strings.ReplaceAll(name, "~1", "/")
}

func extractMessage(spec map[string]interface{}, messageRef interface{}) *Message {
	if msgRef, ok := messageRef.(map[string]interface{}); ok {
		if ref, ok := msgRef["$ref"].(string); ok {
			// Handle both component references and channel message references
			if strings.HasPrefix(ref, "#/components/messages/") {
				if components, ok := spec["components"].(map[string]interface{}); ok {
					if messages, ok := components["messages"].(map[string]interface{}); ok {
						msgName := ref[len("#/components/messages/"):]
						if msgData, ok := messages[msgName].(map[string]interface{}); ok {
							return &Message{
								Name:        msgName,
								ContentType: getStringValue(msgData, "contentType"),
								Payload:     getMapValue(msgData, "payload"),
							}
						}
					}
				}
			} else if strings.Contains(ref, "/messages/") {
				// Handle channel message references
				parts := strings.Split(ref, "/messages/")
				if len(parts) == 2 {
					channelPath := parts[0]
					messageName := parts[1]
					
					// Extract channel name from path
					channelParts := strings.Split(channelPath, "/")
					if len(channelParts) >= 2 {
						channelName := unescapeChannelName(channelParts[len(channelParts)-1])
						
						if channels, ok := spec["channels"].(map[string]interface{}); ok {
							if channel, ok := channels[channelName].(map[string]interface{}); ok {
								if messages, ok := channel["messages"].(map[string]interface{}); ok {
									if msgData, ok := messages[messageName].(map[string]interface{}); ok {
										// Check if it's a reference to components
										if msgRef2, ok := msgData["$ref"].(string); ok {
											return extractMessage(spec, map[string]interface{}{"$ref": msgRef2})
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func findMatchingProviderChannel(spec map[string]interface{}, consumerChannel *Channel) (*Channel, error) {
	channels, ok := spec["channels"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("channels not found in provider spec")
	}

	operations := spec["operations"].(map[string]interface{})

	for channelName, channelData := range channels {
		channel, ok := channelData.(map[string]interface{})
		if !ok {
			continue
		}

		protocol := extractProtocol(spec, channel)
		if protocol != consumerChannel.Protocol {
			continue
		}

		inMessage, outMessage := extractProviderMessages(spec, operations, channelName)
		
		if inMessage != nil && messageMatches(inMessage, consumerChannel.OutMessage) {
			return &Channel{
				Name:       channelName,
				Protocol:   protocol,
				InMessage:  inMessage,
				OutMessage: outMessage,
			}, nil
		}
	}

	return nil, fmt.Errorf("no matching provider channel found")
}

func extractProviderMessages(spec map[string]interface{}, operations map[string]interface{}, channelName string) (*Message, *Message) {
	var inMessage, outMessage *Message

	for _, op := range operations {
		operation, ok := op.(map[string]interface{})
		if !ok {
			continue
		}

		if channel, ok := operation["channel"].(map[string]interface{}); ok {
			if ref, ok := channel["$ref"].(string); ok && containsChannel(ref, channelName) {
				if action, ok := operation["action"].(string); ok && action == "receive" {
					if messages, ok := operation["messages"].([]interface{}); ok && len(messages) > 0 {
						inMessage = extractMessage(spec, messages[0])
					}
					if reply, ok := operation["reply"].(map[string]interface{}); ok {
						if replyMessages, ok := reply["messages"].([]interface{}); ok && len(replyMessages) > 0 {
							outMessage = extractMessage(spec, replyMessages[0])
						}
					}
				}
			}
		}
	}

	return inMessage, outMessage
}

func messageMatches(msg1, msg2 *Message) bool {
	if msg1 == nil || msg2 == nil {
		return false
	}

	payload1, ok1 := msg1.Payload["properties"].(map[string]interface{})
	payload2, ok2 := msg2.Payload["properties"].(map[string]interface{})
	
	if !ok1 || !ok2 {
		return false
	}

	required1 := getRequiredFields(msg1.Payload)
	required2 := getRequiredFields(msg2.Payload)

	if len(required1) != len(required2) {
		return false
	}

	for _, field := range required1 {
		if _, exists := payload2[field]; !exists {
			return false
		}
		if _, exists := payload1[field]; !exists {
			return false
		}
	}

	return true
}

func getRequiredFields(payload map[string]interface{}) []string {
	if required, ok := payload["required"].([]interface{}); ok {
		result := make([]string, len(required))
		for i, field := range required {
			if str, ok := field.(string); ok {
				result[i] = str
			}
		}
		return result
	}
	return []string{}
}

func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getMapValue(m map[string]interface{}, key string) map[string]interface{} {
	if val, ok := m[key].(map[string]interface{}); ok {
		return val
	}
	return nil
}