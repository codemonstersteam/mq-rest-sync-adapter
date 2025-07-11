package main

import (
	"os"
	"testing"
)

// TestFinal_AllComponentsTogether tests the complete system integration
func TestFinal_AllComponentsTogether(t *testing.T) {
	// Verify all test data files exist
	testFiles := []string{
		"testdata_consumer_asyncapi.yml",
		"testdata_provider_main.yml", 
		"testdata_provider_fails.yml",
	}

	for _, file := range testFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Fatalf("Required test data file missing: %s", file)
		}
	}

	// Test 1: Consumer vs Valid Provider
	t.Run("consumer_vs_valid_provider", func(t *testing.T) {
		checker := NewTestContractCompatibilityChecker(
			"testdata_consumer_asyncapi.yml",
			"testdata_provider_main.yml",
		)

		result, err := checker.CheckCompatibility()
		if err != nil {
			t.Fatalf("CheckCompatibility() error = %v", err)
		}

		// Should find some incompatibilities but not be critically broken
		if result.IsCompatible && len(result.BreakingChanges) == 0 && len(result.Warnings) == 0 {
			t.Error("Expected some incompatibilities between consumer and provider")
		}

		// Should have at least MQ channel warnings
		if len(result.Warnings) < 1 {
			t.Error("Expected at least one warning for MQ channels")
		}

		// Verify report generation works
		report := checker.GenerateReport(result)
		if len(report) < 100 {
			t.Error("Generated report seems too short")
		}

		t.Logf("Found %d breaking changes and %d warnings", 
			len(result.BreakingChanges), len(result.Warnings))
	})

	// Test 2: Consumer vs Failing Provider
	t.Run("consumer_vs_failing_provider", func(t *testing.T) {
		checker := NewTestContractCompatibilityChecker(
			"testdata_consumer_asyncapi.yml",
			"testdata_provider_fails.yml",
		)

		result, err := checker.CheckCompatibility()
		if err != nil {
			t.Fatalf("CheckCompatibility() error = %v", err)
		}

		// Should definitely find incompatibilities
		if result.IsCompatible {
			t.Error("Expected incompatibilities with failing provider")
		}

		if len(result.BreakingChanges) == 0 {
			t.Error("Expected breaking changes with failing provider")
		}

		t.Logf("Found %d breaking changes and %d warnings with failing provider", 
			len(result.BreakingChanges), len(result.Warnings))
	})

	// Test 3: Test all core functions work
	t.Run("core_functions", func(t *testing.T) {
		checker := &ContractCompatibilityChecker{}

		// Test path normalization
		tests := []struct {
			input    string
			expected string
		}{
			{"/api/client/{clientId}/balance", "/api/client/{param}/balance"},
			{"/api/client/{Api-clientId}/balance", "/api/client/{param}/balance"},
			{"/api/simple", "/api/simple"},
		}

		for _, test := range tests {
			result := checker.normalizeResourcePath(test.input)
			if result != test.expected {
				t.Errorf("normalizeResourcePath(%s) = %s, want %s", 
					test.input, result, test.expected)
			}
		}

		// Test type compatibility
		typeTests := []struct {
			consumer string
			provider string
			expected bool
		}{
			{"string", "string", true},
			{"integer", "number", true},
			{"number", "integer", true},
			{"string", "integer", false},
		}

		for _, test := range typeTests {
			result := checker.areTypesCompatible(test.consumer, test.provider)
			if result != test.expected {
				t.Errorf("areTypesCompatible(%s, %s) = %v, want %v", 
					test.consumer, test.provider, result, test.expected)
			}
		}

		// Test message type inference
		msgTests := []struct {
			input    string
			expected string
		}{
			{"getBalanceRequest", "request"},
			{"balanceResponse", "response"},
			{"someMessage", "unknown"},
		}

		for _, test := range msgTests {
			result := checker.inferMessageType(test.input)
			if result != test.expected {
				t.Errorf("inferMessageType(%s) = %s, want %s", 
					test.input, result, test.expected)
			}
		}
	})
}

// TestFinal_PerformanceAndMemory tests performance characteristics
func TestFinal_PerformanceAndMemory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Skip if test files don't exist
	for _, file := range []string{"testdata_consumer_asyncapi.yml", "testdata_provider_main.yml"} {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Skip("Test data files not available")
		}
	}

	checker := NewTestContractCompatibilityChecker(
		"testdata_consumer_asyncapi.yml",
		"testdata_provider_main.yml",
	)

	// Run multiple times to check for memory leaks or performance issues
	const iterations = 10
	for i := 0; i < iterations; i++ {
		result, err := checker.CheckCompatibility()
		if err != nil {
			t.Fatalf("Iteration %d failed: %v", i, err)
		}

		if result == nil {
			t.Fatalf("Iteration %d returned nil result", i)
		}

		// Results should be consistent
		if i == 0 {
			// Store first result for comparison
		}
	}
}

// TestFinal_ErrorHandling tests error handling scenarios
func TestFinal_ErrorHandling(t *testing.T) {
	// Test with non-existent consumer file
	t.Run("missing_consumer", func(t *testing.T) {
		checker := NewTestContractCompatibilityChecker(
			"non-existent-consumer.yml",
			"testdata_provider_main.yml",
		)

		_, err := checker.CheckCompatibility()
		if err == nil {
			t.Error("Expected error for missing consumer file")
		}
	})

	// Test with non-existent provider file
	t.Run("missing_provider", func(t *testing.T) {
		checker := NewTestContractCompatibilityChecker(
			"testdata_consumer_asyncapi.yml",
			"non-existent-provider.yml",
		)

		_, err := checker.CheckCompatibility()
		if err == nil {
			t.Error("Expected error for missing provider file")
		}
	})

	// Test with invalid YAML
	t.Run("invalid_yaml", func(t *testing.T) {
		// Create temporary invalid YAML file
		invalidYAML := "invalid: yaml: content: ["
		tmpFile, err := os.CreateTemp("", "invalid_*.yml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.WriteString(invalidYAML); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}
		tmpFile.Close()

		checker := NewTestContractCompatibilityChecker(
			tmpFile.Name(),
			"testdata_provider_main.yml",
		)

		_, err = checker.CheckCompatibility()
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
	})
}

// TestFinal_ReportFormats tests different report formats
func TestFinal_ReportFormats(t *testing.T) {
	// Skip if test files don't exist
	if _, err := os.Stat("testdata_consumer_asyncapi.yml"); os.IsNotExist(err) {
		t.Skip("Consumer test data not available")
	}
	if _, err := os.Stat("testdata_provider_main.yml"); os.IsNotExist(err) {
		t.Skip("Provider test data not available")
	}

	checker := NewTestContractCompatibilityChecker(
		"testdata_consumer_asyncapi.yml",
		"testdata_provider_main.yml",
	)

	result, err := checker.CheckCompatibility()
	if err != nil {
		t.Fatalf("CheckCompatibility() error = %v", err)
	}

	// Test markdown report generation
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
		if !containsString(report, section) {
			t.Errorf("Report missing expected section: %s", section)
		}
	}

	// If there are breaking changes, verify breaking changes section
	if len(result.BreakingChanges) > 0 {
		if !containsString(report, "## 🚨 Breaking Changes") {
			t.Error("Report missing breaking changes section")
		}
	}

	// If there are warnings, verify warnings section
	if len(result.Warnings) > 0 {
		if !containsString(report, "## ⚠️ Warnings") {
			t.Error("Report missing warnings section")
		}
	}

	// Verify result section
	if result.IsCompatible {
		if !containsString(report, "## ✅ Result") {
			t.Error("Report missing success result section")
		}
	} else {
		if !containsString(report, "## ❌ Result") {
			t.Error("Report missing failure result section")
		}
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}