package validator

import (
	"testing"
)

// TestEscapeChannelName тестирует функцию экранирования имен каналов по RFC 6901
func TestEscapeChannelName(t *testing.T) {
	validator := &ChannelValidator{}
	
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		// Базовые тест-кейсы экранирования (7 кейсов)
		{
			name:     "slash_should_be_escaped_to_tilde1",
			input:    "user/events",
			expected: "user~1events",
			hasError: false,
		},
		{
			name:     "tilde_should_be_escaped_to_tilde0",
			input:    "user~events",
			expected: "user~0events",
			hasError: false,
		},
		{
			name:     "both_symbols_should_be_escaped_correctly",
			input:    "config~data/test",
			expected: "config~0data~1test",
			hasError: false,
		},
		{
			name:     "multiple_slashes_should_be_escaped",
			input:    "a/b/c",
			expected: "a~1b~1c",
			hasError: false,
		},
		{
			name:     "multiple_tildes_should_be_escaped",
			input:    "a~b~c",
			expected: "a~0b~0c",
			hasError: false,
		},
		{
			name:     "name_without_special_symbols_should_remain_unchanged",
			input:    "simple",
			expected: "simple",
			hasError: false,
		},
		{
			name:     "empty_string_should_remain_empty",
			input:    "",
			expected: "",
			hasError: false,
		},
		
		// Критичные edge cases RFC 6901 (3 кейса)
		{
			name:     "sequence_tilde01_should_not_become_slash",
			input:    "test~01",
			expected: "test~001",
			hasError: false,
		},
		{
			name:     "sequence_tilde10_should_be_escaped_properly",
			input:    "data~10",
			expected: "data~010",
			hasError: false,
		},
		{
			name:     "combination_of_all_symbols_should_be_escaped_correctly",
			input:    "~path/to~data/",
			expected: "~0path~1to~0data~1",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.escapeChannelName(tt.input)
			
			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if result != tt.expected {
				t.Errorf("escapeChannelName(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestUnescapeChannelName тестирует функцию разэкранирования имен каналов по RFC 6901
func TestUnescapeChannelName(t *testing.T) {
	validator := &ChannelValidator{}
	
	tests := []struct {
		name          string
		input         string
		expected      string
		hasError      bool
		expectedError string
	}{
		// Тест-кейсы разэкранирования - правильный порядок (4 кейса)
		{
			name:     "tilde1_should_be_unescaped_to_slash_first",
			input:    "user~1events",
			expected: "user/events",
			hasError: false,
		},
		{
			name:     "tilde0_should_be_unescaped_to_tilde_second",
			input:    "user~0events",
			expected: "user~events",
			hasError: false,
		},
		{
			name:     "correct_order_both_symbols_unescaped",
			input:    "config~0data~1test",
			expected: "config~data/test",
			hasError: false,
		},
		{
			name:     "sequence_tilde001_should_not_become_slash1",
			input:    "test~001",
			expected: "test~01",
			hasError: false,
		},
		
		// Базовые валидные кейсы
		{
			name:     "name_without_escapes_should_remain_unchanged",
			input:    "simple",
			expected: "simple",
			hasError: false,
		},
		{
			name:     "empty_string_should_remain_empty",
			input:    "",
			expected: "",
			hasError: false,
		},
		
		// Тест-кейсы с ошибками валидации
		{
			name:          "incomplete_escape_at_end_should_return_error",
			input:         "test~",
			expected:      "",
			hasError:      true,
			expectedError: "invalid escape: incomplete '~' at end",
		},
		{
			name:          "invalid_escape_sequence_tilde2_should_return_error",
			input:         "test~2data",
			expected:      "",
			hasError:      true,
			expectedError: "invalid escape: '~2' not allowed",
		},
		{
			name:          "invalid_escape_sequence_tilde_a_should_return_error",
			input:         "test~avalue",
			expected:      "",
			hasError:      true,
			expectedError: "invalid escape: '~a' not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.unescapeChannelName(tt.input)
			
			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error, but got none")
					return
				}
				if tt.expectedError != "" && err.Error() != tt.expectedError {
					t.Errorf("Expected error %q, but got %q", tt.expectedError, err.Error())
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if result != tt.expected {
				t.Errorf("unescapeChannelName(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestEscapeUnescapeRoundTrip тестирует regression кейсы round-trip преобразований
func TestEscapeUnescapeRoundTrip(t *testing.T) {
	validator := &ChannelValidator{}
	
	tests := []struct {
		name     string
		original string
		escaped  string
	}{
		// Regression тесты round-trip (2 кейса)
		{
			name:     "already_escaped_string_should_be_double_escaped_then_unescaped_correctly",
			original: "user~1events",       // уже escaped строка
			escaped:  "user~01events",      // при повторном escape тильда экранируется
		},
		{
			name:     "unescape_then_escape_should_return_original_for_all_source_strings", 
			original: "user/events~data",  // исходная строка
			escaped:  "user~1events~0data", // должна правильно экранироваться
		},
		
		// AsyncAPI 3.0 реальные примеры (2 кейса)
		{
			name:     "asyncapi_example_user_signedup_from_claude_md",
			original: "user/signedup",
			escaped:  "user~1signedup",
		},
		{
			name:     "complex_asyncapi_example_with_both_symbols",
			original: "notifications/user~events",
			escaped:  "notifications~1user~0events",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Тест: escape(original) должен дать escaped
			actualEscaped, err := validator.escapeChannelName(tt.original)
			if err != nil {
				t.Errorf("escapeChannelName(%q) returned error: %v", tt.original, err)
				return
			}
			if actualEscaped != tt.escaped {
				t.Errorf("escapeChannelName(%q) = %q, expected %q", tt.original, actualEscaped, tt.escaped)
			}

			// Тест: unescape(escaped) должен дать original
			actualUnescaped, err := validator.unescapeChannelName(tt.escaped)
			if err != nil {
				t.Errorf("unescapeChannelName(%q) returned error: %v", tt.escaped, err)
				return
			}
			if actualUnescaped != tt.original {
				t.Errorf("unescapeChannelName(%q) = %q, expected %q", tt.escaped, actualUnescaped, tt.original)
			}

			// Round-trip тест: unescape(escape(original)) == original
			roundTrip, err := validator.unescapeChannelName(actualEscaped)
			if err != nil {
				t.Errorf("Round-trip unescapeChannelName(%q) returned error: %v", actualEscaped, err)
				return
			}
			if roundTrip != tt.original {
				t.Errorf("Round-trip failed: original=%q, escaped=%q, unescaped=%q", tt.original, actualEscaped, roundTrip)
			}
		})
	}
}