package main

import (
	"testing"
	"os"
	"strings"
)

func TestChannelRestrictions_RestOnly(t *testing.T) {
	// Skip if test files don't exist
	if _, err := os.Stat("testdata_consumer_asyncapi.yml"); os.IsNotExist(err) {
		t.Skip("Consumer test data not available")
	}
	if _, err := os.Stat("testdata_provider_main.yml"); os.IsNotExist(err) {
		t.Skip("Provider test data not available")
	}

	// Test with rest_only configuration
	checker, err := NewTestContractCompatibilityCheckerFromConfig("test_config_rest_only.yaml")
	if err != nil {
		t.Fatalf("Failed to create checker from config: %v", err)
	}

	// Capture stdout to verify messages
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	result, err := checker.CheckCompatibility()

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout
	capturedOutput := make([]byte, 1024*10) // 10KB buffer
	n, _ := r.Read(capturedOutput)
	output := string(capturedOutput[:n])

	if err != nil {
		t.Fatalf("CheckCompatibility() error = %v", err)
	}

	// Verify that MQ channel restrictions are working
	if !strings.Contains(output, "режим: rest_only") {
		t.Errorf("Expected 'режим: rest_only' message in output")
	}

	if !strings.Contains(output, "Проверка MQ каналов пропущена согласно конфигурации") {
		t.Errorf("Expected MQ channels skip message in output")
	}

	if !strings.Contains(output, "wallet-balance-service") {
		t.Errorf("Expected provider name in skip message")
	}

	// Verify plan section mentions the restriction
	if !strings.Contains(output, "ПРОПУЩЕНО - провайдер") {
		t.Errorf("Expected plan section to mention skipped MQ checks")
	}

	// Should still have REST endpoint checks
	if !strings.Contains(output, "🌐 ПРОВЕРКА REST ENDPOINTS") {
		t.Errorf("Expected REST endpoints check to still run")
	}

	// Should have fewer warnings since MQ channels are skipped
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// Count MQ-related warnings - should be 0 with rest_only mode
	mqWarnings := 0
	for _, warning := range result.Warnings {
		if strings.Contains(warning.Type, "mq_channel") {
			mqWarnings++
		}
	}

	if mqWarnings > 0 {
		t.Errorf("Expected 0 MQ warnings with rest_only mode, got %d", mqWarnings)
	}

	t.Logf("Successfully tested rest_only mode: %d breaking changes, %d warnings", 
		len(result.BreakingChanges), len(result.Warnings))
}

func TestChannelRestrictions_All(t *testing.T) {
	// Test with default (all channels) configuration
	checker := NewTestContractCompatibilityChecker(
		"testdata_consumer_asyncapi.yml",
		"testdata_provider_main.yml",
	)

	// Note: TestContractCompatibilityChecker doesn't have config, so it should check all channels
	result, err := checker.CheckCompatibility()
	if err != nil {
		t.Fatalf("CheckCompatibility() error = %v", err)
	}

	// Should have MQ-related warnings since all channels are checked
	mqWarnings := 0
	for _, warning := range result.Warnings {
		if strings.Contains(warning.Type, "mq_channel") {
			mqWarnings++
		}
	}

	if mqWarnings == 0 {
		t.Errorf("Expected some MQ warnings with default mode, got 0")
	}

	t.Logf("Successfully tested default mode: %d breaking changes, %d warnings (including %d MQ warnings)", 
		len(result.BreakingChanges), len(result.Warnings), mqWarnings)
}

func TestChannelRestrictions_ConfigParsing(t *testing.T) {
	// Test that configuration is parsed correctly
	checker, err := NewTestContractCompatibilityCheckerFromConfig("test_config_rest_only.yaml")
	if err != nil {
		t.Fatalf("Failed to create checker from config: %v", err)
	}

	if checker.Config == nil {
		t.Fatal("Config should not be nil")
	}

	// Verify channel restrictions are parsed
	checkMode := checker.Config.ContractTests.Provider.ChannelRestrictions.CheckMode
	if checkMode != "rest_only" {
		t.Errorf("Expected check_mode 'rest_only', got '%s'", checkMode)
	}

	// Verify provider name
	providerName := checker.Config.ContractTests.Provider.Name
	if providerName != "wallet-balance-service" {
		t.Errorf("Expected provider name 'wallet-balance-service', got '%s'", providerName)
	}

	t.Logf("Configuration parsed correctly: provider=%s, check_mode=%s", 
		providerName, checkMode)
}