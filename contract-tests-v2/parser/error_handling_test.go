package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParserStandardizedErrors тестирует стандартизированные сообщения об ошибках
func TestParserStandardizedErrors(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedError  string
		expectedCode   string
		expectedLocation string
	}{
		{
			name:           "empty_specification_content",
			input:          "",
			expectedError:  "specification content is empty",
			expectedCode:   "PARSE_ERROR",
			expectedLocation: "at input",
		},
		{
			name:           "whitespace_only_content",
			input:          "   \n\t  ",
			expectedError:  "specification content is empty", 
			expectedCode:   "PARSE_ERROR",
			expectedLocation: "at input",
		},
		{
			name: "invalid_yaml_syntax",
			input: `asyncapi: 3.0.0
info:
  title: [invalid yaml
  version: 1.0.0`,
			expectedError:  "failed to parse YAML",
			expectedCode:   "YAML_PARSE_ERROR", 
			expectedLocation: "at input",
		},
		{
			name: "missing_asyncapi_version",
			input: `info:
  title: Test Service
  version: 1.0.0`,
			expectedError:  "asyncapi version is required",
			expectedCode:   "PARSE_ERROR",
			expectedLocation: "at spec.asyncapi",
		},
		{
			name: "unsupported_asyncapi_version",
			input: `asyncapi: 2.6.0
info:
  title: Test Service
  version: 1.0.0`,
			expectedError:  "unsupported AsyncAPI version '2.6.0'",
			expectedCode:   "INVALID_VERSION_ERROR",
			expectedLocation: "at spec.asyncapi",
		},
	}

	parser := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := parser.ParseFromString(tt.input)
			
			require.Error(t, err)
			assert.Nil(t, spec)
			
			// Проверяем стандартизированные компоненты сообщения об ошибке
			assert.Contains(t, err.Error(), tt.expectedCode, 
				"Expected error code '%s' in message: %s", tt.expectedCode, err.Error())
			assert.Contains(t, err.Error(), tt.expectedError,
				"Expected error text '%s' in message: %s", tt.expectedError, err.Error())
			assert.Contains(t, err.Error(), tt.expectedLocation,
				"Expected location '%s' in message: %s", tt.expectedLocation, err.Error())
		})
	}
}

func TestGetChannelByNameErrors(t *testing.T) {
	tests := []struct {
		name           string
		channels       map[string]Channel
		searchName     string
		shouldFindChannel bool
		expectedError  string
		expectedCode   string
		expectedLocation string
	}{
		{
			name: "existing_channel_found",
			channels: map[string]Channel{
				"user/events": {},
			},
			searchName: "user/events",
			shouldFindChannel: true,
		},
		{
			name: "escaped_channel_found",
			channels: map[string]Channel{
				"user~1signedup": {},
			},
			searchName: "user/signedup",
			shouldFindChannel: true,
		},
		{
			name: "unescaped_channel_found",
			channels: map[string]Channel{
				"user/events": {},
			},
			searchName: "user~1events",
			shouldFindChannel: true,
		},
		{
			name: "channel_not_found",
			channels: map[string]Channel{
				"user/events": {},
			},
			searchName: "missing/channel",
			shouldFindChannel: false,
			expectedError: "channel 'missing/channel' not found",
			expectedCode: "CHANNEL_NOT_FOUND",
			expectedLocation: "at spec.channels",
		},
	}

	parser := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &AsyncAPISpec{
				Channels: tt.channels,
			}
			
			channel, err := parser.GetChannelByName(spec, tt.searchName)
			
			if tt.shouldFindChannel {
				assert.NoError(t, err)
				assert.NotNil(t, channel)
			} else {
				require.Error(t, err)
				assert.Nil(t, channel)
				
				// Проверяем стандартизированное сообщение об ошибке
				assert.Contains(t, err.Error(), tt.expectedCode)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Contains(t, err.Error(), tt.expectedLocation)
			}
		})
	}
}

func TestGetMessageByRefErrors(t *testing.T) {
	tests := []struct {
		name           string
		spec           *AsyncAPISpec
		ref            string
		shouldFindMessage bool
		expectedError  string
		expectedCode   string
		expectedLocation string
	}{
		{
			name: "existing_component_message_found",
			spec: &AsyncAPISpec{
				Components: &Components{
					Messages: map[string]Message{
						"UserEvent": {Name: "UserEvent"},
					},
				},
			},
			ref: "#/components/messages/UserEvent",
			shouldFindMessage: true,
		},
		{
			name: "empty_reference_error",
			spec: &AsyncAPISpec{},
			ref: "",
			shouldFindMessage: false,
			expectedError: "empty reference",
			expectedCode: "INVALID_REF_ERROR",
			expectedLocation: "at messageRef",
		},
		{
			name: "missing_component_message_error",
			spec: &AsyncAPISpec{
				Components: &Components{
					Messages: map[string]Message{
						"UserEvent": {Name: "UserEvent"},
					},
				},
			},
			ref: "#/components/messages/MissingMessage",
			shouldFindMessage: false,
			expectedError: "message 'MissingMessage' not found",
			expectedCode: "COMPONENT_NOT_FOUND",
			expectedLocation: "at #/components/messages",
		},
		{
			name: "invalid_reference_format",
			spec: &AsyncAPISpec{},
			ref: "invalid/ref/format",
			shouldFindMessage: false,
			expectedError: "unable to resolve reference",
			expectedCode: "INVALID_REF_ERROR",
			expectedLocation: "at messageRef",
		},
	}

	parser := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, err := parser.GetMessageByRef(tt.spec, tt.ref)
			
			if tt.shouldFindMessage {
				assert.NoError(t, err)
				assert.NotNil(t, message)
			} else {
				require.Error(t, err)
				assert.Nil(t, message)
				
				// Проверяем стандартизированное сообщение об ошибке
				assert.Contains(t, err.Error(), tt.expectedCode)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Contains(t, err.Error(), tt.expectedLocation)
			}
		})
	}
}

func TestGetServerByRefErrors(t *testing.T) {
	tests := []struct {
		name           string
		servers        map[string]Server
		ref            string
		shouldFindServer bool
		expectedError  string
		expectedCode   string
		expectedLocation string
	}{
		{
			name: "existing_server_found",
			servers: map[string]Server{
				"rabbitmq": {Protocol: "amqp"},
			},
			ref: "#/servers/rabbitmq",
			shouldFindServer: true,
		},
		{
			name: "empty_reference_error",
			servers: map[string]Server{},
			ref: "",
			shouldFindServer: false,
			expectedError: "empty reference",
			expectedCode: "INVALID_REF_ERROR",
			expectedLocation: "at serverRef",
		},
		{
			name: "missing_server_error",
			servers: map[string]Server{
				"rabbitmq": {Protocol: "amqp"},
			},
			ref: "#/servers/missing",
			shouldFindServer: false,
			expectedError: "server 'missing' not found",
			expectedCode: "COMPONENT_NOT_FOUND",
			expectedLocation: "at #/servers",
		},
		{
			name: "invalid_reference_format",
			servers: map[string]Server{},
			ref: "invalid/server/ref",
			shouldFindServer: false,
			expectedError: "unable to resolve reference",
			expectedCode: "INVALID_REF_ERROR",
			expectedLocation: "at serverRef",
		},
	}

	parser := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &AsyncAPISpec{
				Servers: tt.servers,
			}
			
			server, err := parser.GetServerByRef(spec, tt.ref)
			
			if tt.shouldFindServer {
				assert.NoError(t, err)
				assert.NotNil(t, server)
			} else {
				require.Error(t, err)
				assert.Nil(t, server)
				
				// Проверяем стандартизированное сообщение об ошибке
				assert.Contains(t, err.Error(), tt.expectedCode)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Contains(t, err.Error(), tt.expectedLocation)
			}
		})
	}
}

func TestGetOperationByNameErrors(t *testing.T) {
	tests := []struct {
		name           string
		operations     map[string]Operation
		operationName  string
		shouldFindOperation bool
		expectedError  string
		expectedCode   string
		expectedLocation string
	}{
		{
			name: "existing_operation_found",
			operations: map[string]Operation{
				"publishEvent": {Action: "send"},
			},
			operationName: "publishEvent",
			shouldFindOperation: true,
		},
		{
			name: "missing_operation_error",
			operations: map[string]Operation{
				"publishEvent": {Action: "send"},
			},
			operationName: "missingOperation",
			shouldFindOperation: false,
			expectedError: "operation 'missingOperation' not found",
			expectedCode: "COMPONENT_NOT_FOUND",
			expectedLocation: "at spec.operations",
		},
	}

	parser := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &AsyncAPISpec{
				Operations: tt.operations,
			}
			
			operation, err := parser.GetOperationByName(spec, tt.operationName)
			
			if tt.shouldFindOperation {
				assert.NoError(t, err)
				assert.NotNil(t, operation)
			} else {
				require.Error(t, err)
				assert.Nil(t, operation)
				
				// Проверяем стандартизированное сообщение об ошибке
				assert.Contains(t, err.Error(), tt.expectedCode)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Contains(t, err.Error(), tt.expectedLocation)
			}
		})
	}
}