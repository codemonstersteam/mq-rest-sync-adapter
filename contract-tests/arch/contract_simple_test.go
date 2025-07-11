package main

import (
	"os"
	"testing"
)

func TestMain_BasicFunctionality(t *testing.T) {
	// Test basic functions work
	checker := &ContractCompatibilityChecker{}

	// Test path normalization
	normalized := checker.normalizeResourcePath("/api/client/{clientId}/wallet/balance")
	expected := "/api/client/{param}/wallet/balance"
	if normalized != expected {
		t.Errorf("normalizeResourcePath() = %v, want %v", normalized, expected)
	}

	// Test type compatibility
	if !checker.areTypesCompatible("integer", "number") {
		t.Errorf("areTypesCompatible(integer, number) should be true")
	}

	if checker.areTypesCompatible("string", "integer") {
		t.Errorf("areTypesCompatible(string, integer) should be false")
	}

	// Test message type inference
	if checker.inferMessageType("testRequest") != "request" {
		t.Errorf("inferMessageType(testRequest) should be 'request'")
	}

	if checker.inferMessageType("testResponse") != "response" {
		t.Errorf("inferMessageType(testResponse) should be 'response'")
	}

	// Test reference extraction
	if checker.extractMessageName("#/components/messages/testMessage") != "testMessage" {
		t.Errorf("extractMessageName() failed")
	}

	if checker.extractSchemaName("#/components/schemas/testSchema") != "testSchema" {
		t.Errorf("extractSchemaName() failed")
	}
}

func TestMain_WithRealSpecs(t *testing.T) {
	// Skip if spec files don't exist
	if _, err := os.Stat("testdata_consumer_asyncapi.yml"); os.IsNotExist(err) {
		t.Skip("Consumer specification file not available")
	}
	if _, err := os.Stat("testdata_provider_main.yml"); os.IsNotExist(err) {
		t.Skip("Provider specification file not available")
	}

	// Test with real specifications
	checker := NewTestContractCompatibilityChecker(
		"testdata_consumer_asyncapi.yml",
		"testdata_provider_main.yml",
	)

	result, err := checker.CheckCompatibility()
	if err != nil {
		t.Fatalf("CheckCompatibility() error = %v", err)
	}

	// Verify result structure
	if result == nil {
		t.Fatal("CheckCompatibility() returned nil result")
	}

	// Should find some incompatibilities
	if result.IsCompatible && len(result.BreakingChanges) == 0 && len(result.Warnings) == 0 {
		t.Error("Expected some incompatibilities between real specs")
	}

	// Verify timestamps
	if result.CheckedAt.IsZero() {
		t.Error("CheckedAt timestamp should be set")
	}

	// Verify versions are extracted
	if result.ConsumerVersion == "" {
		t.Error("Consumer version should be extracted")
	}

	if result.ProviderVersion == "" {
		t.Error("Provider version should be extracted")
	}
}

func TestMain_ReportGeneration(t *testing.T) {
	// Create a simple result
	result := &CompatibilityResult{
		IsCompatible:    false,
		BreakingChanges: []BreakingChange{
			{
				Type:        "test_breaking_change",
				Field:       "test.field",
				OldValue:    "old",
				NewValue:    "new",
				Description: "Test breaking change",
				Severity:    "high",
			},
		},
		Warnings: []CompatibilityWarning{
			{
				Type:       "test_warning",
				Field:      "test.field",
				Message:    "Test warning message",
				Suggestion: "Test suggestion",
			},
		},
		ConsumerVersion: "1.0.0",
		ProviderVersion: "1.0.0",
	}

	checker := &ContractCompatibilityChecker{}
	report := checker.GenerateReport(result)

	// Verify report contains expected sections
	expectedSections := []string{
		"# Contract Compatibility Report",
		"**Compatible:** false",
		"## 🚨 Breaking Changes",
		"## ⚠️ Warnings",
		"## ❌ Result",
	}

	for _, section := range expectedSections {
		if !contains(report, section) {
			t.Errorf("Report missing expected section: %s", section)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}