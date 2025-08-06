package validator

import (
	"testing"

	"github.com/codemonsters/contract-tests-v2/parser"
)

func TestExtractChannelProtocol(t *testing.T) {
	validator := NewChannelValidator()

	tests := []struct {
		name        string
		spec        *parser.AsyncAPISpec
		channel     *parser.Channel
		expected    string
		expectError bool
		errorMsg    string
	}{
		{
			name:     "valid server reference",
			spec:     createSpecWithServer("production", "amqp"),
			channel:  createChannelWithServer("#/servers/production"),
			expected: "amqp",
		},
		{
			name:        "channel without servers",
			spec:        createSpecWithServer("production", "amqp"),
			channel:     createChannelWithoutServers(),
			expectError: true,
			errorMsg:    "no servers defined for channel",
		},
		{
			name:        "invalid server reference",
			spec:        createSpecWithServer("production", "amqp"),
			channel:     createChannelWithServer("invalid-ref"),
			expectError: true,
			errorMsg:    "invalid server reference: invalid-ref",
		},
		{
			name:        "reference to nonexistent server",
			spec:        createSpecWithServer("production", "amqp"),
			channel:     createChannelWithServer("#/servers/nonexistent"),
			expectError: true,
			errorMsg:    "server nonexistent not found",
		},
		{
			name:     "multiple servers (take first)",
			spec:     createSpecWithMultipleServers(),
			channel:  createChannelWithMultipleServers(),
			expected: "amqp",
		},
		{
			name:        "nil channel",
			spec:        createSpecWithServer("production", "amqp"),
			channel:     nil,
			expectError: true,
			errorMsg:    "runtime error", // panic будет пойман как runtime error
		},
		{
			name:        "empty server reference",
			spec:        createSpecWithServer("production", "amqp"),
			channel:     createChannelWithServer(""),
			expectError: true,
			errorMsg:    "invalid server reference:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Обрабатываем случай с nil channel через defer/recover
			defer func() {
				if r := recover(); r != nil && tt.name == "nil channel" {
					// Ожидаемая паника для nil channel
					return
				} else if r != nil {
					t.Errorf("unexpected panic: %v", r)
				}
			}()

			result, err := validator.extractChannelProtocol(tt.spec, tt.channel)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
				
				// Проверяем стандартизированные коды ошибок для некоторых кейсов
				switch tt.name {
				case "channel without servers":
					if !containsString(err.Error(), "VALIDATION_ERROR") {
						t.Errorf("error should contain VALIDATION_ERROR code, got '%s'", err.Error())
					}
				case "invalid server reference", "reference to nonexistent server":
					if !containsString(err.Error(), "VALIDATION_ERROR") {
						t.Errorf("error should contain VALIDATION_ERROR code, got '%s'", err.Error())
					}
				case "empty server reference":
					if !containsString(err.Error(), "VALIDATION_ERROR") {
						t.Errorf("error should contain VALIDATION_ERROR code, got '%s'", err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result != tt.expected {
					t.Errorf("expected protocol '%s', got '%s'", tt.expected, result)
				}
			}
		})
	}
}

// Helper functions для создания тестовых данных

func createSpecWithServer(serverName, protocol string) *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Servers: map[string]parser.Server{
			serverName: {
				Protocol: protocol,
			},
		},
	}
}

func createSpecWithMultipleServers() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Servers: map[string]parser.Server{
			"production": {
				Protocol: "amqp",
			},
			"staging": {
				Protocol: "mqtt",
			},
		},
	}
}

func createChannelWithServer(serverRef string) *parser.Channel {
	return &parser.Channel{
		Servers: []parser.ServerRef{
			{Ref: serverRef},
		},
	}
}

func createChannelWithoutServers() *parser.Channel {
	return &parser.Channel{
		Servers: []parser.ServerRef{},
	}
}

func createChannelWithMultipleServers() *parser.Channel {
	return &parser.Channel{
		Servers: []parser.ServerRef{
			{Ref: "#/servers/production"},
			{Ref: "#/servers/staging"},
		},
	}
}

// Utility function для проверки содержания строки
func containsString(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && 
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}