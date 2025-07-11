package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/tidwall/gjson"
)

// TestHelper provides utility functions for contract testing
type TestHelper struct {
	t *testing.T
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{t: t}
}

// CreateTempSpec creates a temporary specification file for testing
func (h *TestHelper) CreateTempSpec(name, content string) string {
	file, err := os.CreateTemp("", name+"_*.yml")
	if err != nil {
		h.t.Fatalf("Failed to create temp file: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		h.t.Fatalf("Failed to write temp file: %v", err)
	}

	return file.Name()
}

// AssertContainsBreakingChange checks if result contains specific breaking change type
func (h *TestHelper) AssertContainsBreakingChange(result *CompatibilityResult, changeType string) {
	for _, change := range result.BreakingChanges {
		if change.Type == changeType {
			return
		}
	}
	h.t.Errorf("Expected breaking change type '%s' not found in results", changeType)
}

// AssertContainsWarning checks if result contains specific warning type
func (h *TestHelper) AssertContainsWarning(result *CompatibilityResult, warningType string) {
	for _, warning := range result.Warnings {
		if warning.Type == warningType {
			return
		}
	}
	h.t.Errorf("Expected warning type '%s' not found in results", warningType)
}

// AssertFieldChange checks if result contains specific field change
func (h *TestHelper) AssertFieldChange(result *CompatibilityResult, field, oldValue, newValue string) {
	for _, change := range result.BreakingChanges {
		if change.Field == field && change.OldValue == oldValue && change.NewValue == newValue {
			return
		}
	}
	h.t.Errorf("Expected field change for '%s' from '%s' to '%s' not found", field, oldValue, newValue)
}

// AssertJSONStructure validates the JSON structure of the result
func (h *TestHelper) AssertJSONStructure(result *CompatibilityResult) {
	jsonData, err := json.Marshal(result)
	if err != nil {
		h.t.Fatalf("Failed to marshal result to JSON: %v", err)
	}

	parsed := gjson.ParseBytes(jsonData)

	// Check required fields
	requiredFields := []string{
		"isCompatible",
		"breakingChanges",
		"warnings",
		"checkedAt",
		"consumerVersion",
		"providerVersion",
	}

	for _, field := range requiredFields {
		if !parsed.Get(field).Exists() {
			h.t.Errorf("JSON result missing required field: %s", field)
		}
	}

	// Validate breaking changes structure
	parsed.Get("breakingChanges").ForEach(func(_, change gjson.Result) bool {
		changeFields := []string{"type", "field", "oldValue", "newValue", "description", "severity"}
		for _, field := range changeFields {
			if !change.Get(field).Exists() {
				h.t.Errorf("Breaking change missing required field: %s", field)
			}
		}
		return true
	})

	// Validate warnings structure
	parsed.Get("warnings").ForEach(func(_, warning gjson.Result) bool {
		warningFields := []string{"type", "field", "message", "suggestion"}
		for _, field := range warningFields {
			if !warning.Get(field).Exists() {
				h.t.Errorf("Warning missing required field: %s", field)
			}
		}
		return true
	})
}

// MockChecker provides a mock implementation for testing
type MockChecker struct {
	ConsumerSpec gjson.Result
	ProviderSpec gjson.Result
	ShouldError  bool
	Result       *CompatibilityResult
}

// CheckCompatibility mock implementation
func (m *MockChecker) CheckCompatibility() (*CompatibilityResult, error) {
	if m.ShouldError {
		return nil, &TestError{Message: "mock error"}
	}
	return m.Result, nil
}

// TestError represents a test error
type TestError struct {
	Message string
}

func (e *TestError) Error() string {
	return e.Message
}

// SpecBuilder helps build test specifications
type SpecBuilder struct {
	spec map[string]interface{}
}

// NewSpecBuilder creates a new spec builder
func NewSpecBuilder() *SpecBuilder {
	return &SpecBuilder{
		spec: map[string]interface{}{
			"asyncapi": "3.0.0",
			"info": map[string]interface{}{
				"title":   "Test API",
				"version": "1.0.0",
			},
			"channels":   map[string]interface{}{},
			"operations": map[string]interface{}{},
			"components": map[string]interface{}{
				"messages": map[string]interface{}{},
				"schemas":  map[string]interface{}{},
			},
		},
	}
}

// WithTitle sets the API title
func (b *SpecBuilder) WithTitle(title string) *SpecBuilder {
	info := b.spec["info"].(map[string]interface{})
	info["title"] = title
	return b
}

// WithVersion sets the API version
func (b *SpecBuilder) WithVersion(version string) *SpecBuilder {
	info := b.spec["info"].(map[string]interface{})
	info["version"] = version
	return b
}

// WithRESTChannel adds a REST channel
func (b *SpecBuilder) WithRESTChannel(name, address string) *SpecBuilder {
	channels := b.spec["channels"].(map[string]interface{})
	channels[name] = map[string]interface{}{
		"address":     address,
		"description": "Test REST channel",
		"messages": map[string]interface{}{
			"testRequest": map[string]interface{}{
				"$ref": "#/components/messages/testRequest",
			},
			"testResponse": map[string]interface{}{
				"$ref": "#/components/messages/testResponse",
			},
		},
	}
	return b
}

// WithMQChannel adds an MQ channel
func (b *SpecBuilder) WithMQChannel(name, address string) *SpecBuilder {
	channels := b.spec["channels"].(map[string]interface{})
	channels[name] = map[string]interface{}{
		"address":     address,
		"description": "Test MQ channel",
		"messages": map[string]interface{}{
			"testMessage": map[string]interface{}{
				"$ref": "#/components/messages/testMessage",
			},
		},
	}
	return b
}

// WithMessage adds a message definition
func (b *SpecBuilder) WithMessage(name, payloadSchemaRef string) *SpecBuilder {
	components := b.spec["components"].(map[string]interface{})
	messages := components["messages"].(map[string]interface{})
	messages[name] = map[string]interface{}{
		"name":        name,
		"title":       "Test " + name,
		"contentType": "application/json",
		"payload": map[string]interface{}{
			"$ref": payloadSchemaRef,
		},
	}
	return b
}

// WithSchema adds a schema definition
func (b *SpecBuilder) WithSchema(name string, schema map[string]interface{}) *SpecBuilder {
	components := b.spec["components"].(map[string]interface{})
	schemas := components["schemas"].(map[string]interface{})
	schemas[name] = schema
	return b
}

// Build returns the built specification as JSON bytes
func (b *SpecBuilder) Build() []byte {
	data, _ := json.Marshal(b.spec)
	return data
}

// BuildYAML returns the built specification as YAML string
func (b *SpecBuilder) BuildYAML() string {
	// Simple JSON to YAML conversion for testing
	data, _ := json.MarshalIndent(b.spec, "", "  ")
	return string(data)
}

// Common test schemas
func GetTestRequestSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"required": []string{
			"requestType",
			"dateFrom",
			"dateTo",
			"operTime",
		},
		"properties": map[string]interface{}{
			"requestType": map[string]interface{}{
				"type": "string",
				"enum": []string{"operationHistory"},
			},
			"dateFrom": map[string]interface{}{
				"type":   "number",
				"format": "double",
			},
			"dateTo": map[string]interface{}{
				"type":   "number",
				"format": "double",
			},
			"operTime": map[string]interface{}{
				"type":   "number",
				"format": "double",
			},
		},
	}
}

func GetTestResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"required": []string{
			"status",
			"actualTimestamp",
			"data",
		},
		"properties": map[string]interface{}{
			"status": map[string]interface{}{
				"type": "string",
				"enum": []string{"success", "error"},
			},
			"actualTimestamp": map[string]interface{}{
				"type":   "integer",
				"format": "int64",
			},
			"data": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"balance": map[string]interface{}{
						"type":   "number",
						"format": "double",
					},
					"currency": map[string]interface{}{
						"type": "string",
						"enum": []string{"rub", "usd", "eur"},
					},
				},
			},
		},
	}
}

// GetIncompatibleResponseSchema returns a schema with breaking changes
func GetIncompatibleResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"required": []string{
			"status",
			"actualTimestamp",
			"data",
			"newRequiredField", // Breaking change: new required field
		},
		"properties": map[string]interface{}{
			"status": map[string]interface{}{
				"type": "string",
				"enum": []string{"success", "error", "pending"}, // Breaking change: new enum value
			},
			"actualTimestamp": map[string]interface{}{
				"type":   "string", // Breaking change: type changed from integer to string
				"format": "date-time",
			},
			"data": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"balance": map[string]interface{}{
						"type":   "integer", // Breaking change: type changed from number to integer
						"format": "int64",
					},
					"currency": map[string]interface{}{
						"type": "string",
						"enum": []string{"rub", "usd", "eur", "btc"}, // Breaking change: new enum value
					},
				},
			},
			"newRequiredField": map[string]interface{}{
				"type": "string",
			},
		},
	}
}

// NewTestContractCompatibilityCheckerFromConfig creates a test checker from config with local files
func NewTestContractCompatibilityCheckerFromConfig(configPath string) (*TestContractCompatibilityChecker, error) {
	// Load the config first to get settings
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, err
	}

	// Create test checker using local files
	checker := &TestContractCompatibilityChecker{
		ConsumerSpecPath: config.ContractTests.Consumer.SpecPath,
		ProviderSpecPath: config.ContractTests.Provider.SpecURL, // Treat as local file path for tests
		Config:          config,
	}

	return checker, nil
}