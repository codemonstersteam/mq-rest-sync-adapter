package main

import (
	"os"
	"strings"
	"testing"
)

// Integration tests using real provider specifications
func TestContractCompatibilityChecker_RealProviderSpecs(t *testing.T) {
	// Skip if provider spec files don't exist
	if _, err := os.Stat("testdata_provider_main.yml"); os.IsNotExist(err) {
		t.Skip("Real provider specification files not available")
	}

	tests := []struct {
		name                    string
		consumerSpecPath        string
		providerSpecPath        string
		expectedCompatible      bool
		minBreakingChanges      int
		minWarnings             int
		shouldContainBreaking   []string
		shouldContainWarnings   []string
	}{
		{
			name:                    "consumer vs main provider",
			consumerSpecPath:        "testdata_consumer_asyncapi.yml",
			providerSpecPath:        "testdata_provider_main.yml",
			expectedCompatible:      false,
			minBreakingChanges:      1,
			minWarnings:             1,
			shouldContainBreaking:   []string{"rest_message_type_missing"},
			shouldContainWarnings:   []string{"mq_channel_missing"},
		},
		{
			name:                    "consumer vs failing provider",
			consumerSpecPath:        "testdata_consumer_asyncapi.yml",
			providerSpecPath:        "testdata_provider_fails.yml",
			expectedCompatible:      false,
			minBreakingChanges:      1,
			minWarnings:             1,
			shouldContainBreaking:   []string{"rest_message_type_missing"},
			shouldContainWarnings:   []string{"mq_channel_missing"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout to verify explanation output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			checker := NewTestContractCompatibilityChecker(tt.consumerSpecPath, tt.providerSpecPath)
			result, err := checker.CheckCompatibility()

			// Restore stdout and read captured output
			w.Close()
			os.Stdout = oldStdout
			capturedOutput := make([]byte, 1024*10) // 10KB buffer
			n, _ := r.Read(capturedOutput)
			output := string(capturedOutput[:n])

			if err != nil {
				t.Errorf("CheckCompatibility() unexpected error = %v", err)
				return
			}

			// Verify explanation output (TestContractCompatibilityChecker doesn't show full analysis)
			if !strings.Contains(output, "🌐 ПРОВЕРКА REST ENDPOINTS...") {
				t.Errorf("Expected REST endpoints check in output, got: %v", output)
			}


			if !strings.Contains(output, "📨 ПРОВЕРКА MQ КАНАЛОВ...") {
				t.Errorf("Expected MQ channels check in output")
			}

			// Verify compatibility result
			if result.IsCompatible != tt.expectedCompatible {
				t.Errorf("CheckCompatibility() IsCompatible = %v, want %v", result.IsCompatible, tt.expectedCompatible)
			}

			// Verify minimum breaking changes
			if len(result.BreakingChanges) < tt.minBreakingChanges {
				t.Errorf("CheckCompatibility() BreakingChanges count = %v, want at least %v", len(result.BreakingChanges), tt.minBreakingChanges)
			}

			// Verify minimum warnings
			if len(result.Warnings) < tt.minWarnings {
				t.Errorf("CheckCompatibility() Warnings count = %v, want at least %v", len(result.Warnings), tt.minWarnings)
			}

			// Verify specific breaking change types
			breakingChangeTypes := make(map[string]bool)
			for _, change := range result.BreakingChanges {
				breakingChangeTypes[change.Type] = true
			}

			for _, expectedType := range tt.shouldContainBreaking {
				if !breakingChangeTypes[expectedType] {
					t.Errorf("Expected breaking change type '%s' not found in results", expectedType)
				}
			}

			// Verify specific warning types
			warningTypes := make(map[string]bool)
			for _, warning := range result.Warnings {
				warningTypes[warning.Type] = true
			}

			for _, expectedType := range tt.shouldContainWarnings {
				if !warningTypes[expectedType] {
					t.Errorf("Expected warning type '%s' not found in results", expectedType)
				}
			}

			// Verify version information is extracted
			if result.ConsumerVersion == "" {
				t.Errorf("Expected consumer version to be extracted")
			}

			if result.ProviderVersion == "" {
				t.Errorf("Expected provider version to be extracted")
			}

			// Verify timestamp is set
			if result.CheckedAt.IsZero() {
				t.Errorf("Expected CheckedAt timestamp to be set")
			}
		})
	}
}

func TestContractCompatibilityChecker_ErrorHandling(t *testing.T) {
	tests := []struct {
		name               string
		consumerSpecPath   string
		providerSpecPath   string
		expectError        bool
		errorShouldContain string
	}{
		{
			name:               "non-existent consumer spec",
			consumerSpecPath:   "non-existent-consumer.yml",
			providerSpecPath:   "testdata_provider_main.yml",
			expectError:        true,
			errorShouldContain: "failed to load consumer spec",
		},
		{
			name:               "non-existent provider spec",
			consumerSpecPath:   "testdata_consumer_asyncapi.yml",
			providerSpecPath:   "non-existent-provider.yml",
			expectError:        true,
			errorShouldContain: "failed to load provider spec",
		},
		{
			name:             "both specs exist",
			consumerSpecPath: "testdata_consumer_asyncapi.yml",
			providerSpecPath: "testdata_provider_main.yml",
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if required files don't exist
			if tt.providerSpecPath == "testdata_provider_main.yml" {
				if _, err := os.Stat(tt.providerSpecPath); os.IsNotExist(err) {
					t.Skip("Real provider specification file not available")
				}
			}

			if tt.consumerSpecPath == "testdata_consumer_asyncapi.yml" {
				if _, err := os.Stat(tt.consumerSpecPath); os.IsNotExist(err) {
					t.Skip("Consumer specification file not available")
				}
			}

			checker := NewTestContractCompatibilityChecker(tt.consumerSpecPath, tt.providerSpecPath)
			_, err := checker.CheckCompatibility()

			if tt.expectError {
				if err == nil {
					t.Errorf("CheckCompatibility() expected error, got nil")
					return
				}
				if tt.errorShouldContain != "" && !strings.Contains(err.Error(), tt.errorShouldContain) {
					t.Errorf("CheckCompatibility() error = %v, should contain %v", err, tt.errorShouldContain)
				}
			} else {
				if err != nil {
					t.Errorf("CheckCompatibility() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestContractCompatibilityChecker_DetailedFieldComparison(t *testing.T) {
	// Skip if provider spec files don't exist
	if _, err := os.Stat("testdata_provider_main.yml"); os.IsNotExist(err) {
		t.Skip("Real provider specification files not available")
	}

	checker := NewTestContractCompatibilityChecker(
		"testdata_consumer_asyncapi.yml",
		"testdata_provider_main.yml",
	)

	result, err := checker.CheckCompatibility()
	if err != nil {
		t.Fatalf("CheckCompatibility() error = %v", err)
	}

	// Test specific field-level validations
	t.Run("enum_validation", func(t *testing.T) {
		// Look for enum-related changes
		var enumChanges []BreakingChange
		for _, change := range result.BreakingChanges {
			if strings.Contains(change.Type, "enum") {
				enumChanges = append(enumChanges, change)
			}
		}

		// Should detect enum differences if they exist
		t.Logf("Found %d enum-related changes", len(enumChanges))
	})

	t.Run("type_validation", func(t *testing.T) {
		// Look for type-related changes
		var typeChanges []BreakingChange
		for _, change := range result.BreakingChanges {
			if strings.Contains(change.Type, "type") {
				typeChanges = append(typeChanges, change)
			}
		}

		t.Logf("Found %d type-related changes", len(typeChanges))
	})

	t.Run("required_fields_validation", func(t *testing.T) {
		// Look for required field changes
		var requiredFieldChanges []BreakingChange
		for _, change := range result.BreakingChanges {
			if strings.Contains(change.Type, "required") {
				requiredFieldChanges = append(requiredFieldChanges, change)
			}
		}

		t.Logf("Found %d required field changes", len(requiredFieldChanges))
	})
}

func TestContractCompatibilityChecker_PerformanceWithRealSpecs(t *testing.T) {
	// Skip if provider spec files don't exist
	if _, err := os.Stat("testdata_provider_main.yml"); os.IsNotExist(err) {
		t.Skip("Real provider specification files not available")
	}

	checker := NewTestContractCompatibilityChecker(
		"testdata_consumer_asyncapi.yml",
		"testdata_provider_main.yml",
	)

	// Performance test - should complete within reasonable time
	const maxIterations = 10
	for i := 0; i < maxIterations; i++ {
		result, err := checker.CheckCompatibility()
		if err != nil {
			t.Fatalf("CheckCompatibility() iteration %d error = %v", i, err)
		}

		if result == nil {
			t.Fatalf("CheckCompatibility() iteration %d returned nil result", i)
		}

		// Verify result structure is consistent
		if len(result.BreakingChanges) == 0 && len(result.Warnings) == 0 && result.IsCompatible {
			t.Errorf("CheckCompatibility() iteration %d: unexpected fully compatible result", i)
		}
	}
}

func TestContractCompatibilityChecker_OutputFormat(t *testing.T) {
	// Skip if provider spec files don't exist
	if _, err := os.Stat("testdata_provider_main.yml"); os.IsNotExist(err) {
		t.Skip("Real provider specification files not available")
	}

	checker := NewTestContractCompatibilityChecker(
		"testdata_consumer_asyncapi.yml",
		"testdata_provider_main.yml",
	)

	result, err := checker.CheckCompatibility()
	if err != nil {
		t.Fatalf("CheckCompatibility() error = %v", err)
	}

	// Test report generation
	report := checker.GenerateReport(result)

	// Verify report structure
	expectedSections := []string{
		"# Contract Compatibility Report",
		"**Checked at:**",
		"**Consumer Version:**",
		"**Provider Version:**",
		"**Compatible:**",
	}

	for _, section := range expectedSections {
		if !strings.Contains(report, section) {
			t.Errorf("Report missing expected section: %s", section)
		}
	}

	// Verify breaking changes section if any exist
	if len(result.BreakingChanges) > 0 {
		if !strings.Contains(report, "## 🚨 Breaking Changes") {
			t.Errorf("Report missing breaking changes section")
		}
	}

	// Verify warnings section if any exist
	if len(result.Warnings) > 0 {
		if !strings.Contains(report, "## ⚠️ Warnings") {
			t.Errorf("Report missing warnings section")
		}
	}

	// Verify result section
	if result.IsCompatible {
		if !strings.Contains(report, "## ✅ Result") {
			t.Errorf("Report missing success result section")
		}
	} else {
		if !strings.Contains(report, "## ❌ Result") {
			t.Errorf("Report missing failure result section")
		}
	}
}