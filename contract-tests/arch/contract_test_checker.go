package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

// TestContractCompatibilityChecker is a version of the checker for testing with local files
type TestContractCompatibilityChecker struct {
	ConsumerSpecPath string
	ProviderSpecPath string
	Config           *Config
}

// NewTestContractCompatibilityChecker creates a new test checker that works with local files
func NewTestContractCompatibilityChecker(consumerSpecPath, providerSpecPath string) *TestContractCompatibilityChecker {
	return &TestContractCompatibilityChecker{
		ConsumerSpecPath: consumerSpecPath,
		ProviderSpecPath: providerSpecPath,
	}
}

// CheckCompatibility performs compatibility check using local files
func (c *TestContractCompatibilityChecker) CheckCompatibility() (*CompatibilityResult, error) {
	// Load consumer specification
	consumerSpec, err := c.loadLocalSpec(c.ConsumerSpecPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load consumer spec: %w", err)
	}

	// Load provider specification
	providerSpec, err := c.loadLocalSpec(c.ProviderSpecPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load provider spec: %w", err)
	}

	// Check compatibility
	result := &CompatibilityResult{
		IsCompatible:     true,
		BreakingChanges:  []BreakingChange{},
		Warnings:         []CompatibilityWarning{},
		CheckedAt:        time.Now(),
		ConsumerVersion:  consumerSpec.Get("info.version").String(),
		ProviderVersion:  providerSpec.Get("info.version").String(),
	}

	// Use the same compatibility checking logic
	checker := &ContractCompatibilityChecker{
		Config: c.Config,
	}
	
	// Add analysis if config is available (for testing channel restrictions)
	if c.Config != nil {
		checker.analyzeAndExplainCompatibilityPlan(consumerSpec, providerSpec)
	}
	
	checker.checkRESTEndpointCompatibility(consumerSpec, providerSpec, result)
	checker.checkMQChannelCompatibility(consumerSpec, providerSpec, result)
	checker.checkSchemaCompatibility(consumerSpec, providerSpec, result)

	// If there are breaking changes, mark as incompatible
	if len(result.BreakingChanges) > 0 {
		result.IsCompatible = false
	}

	return result, nil
}

// loadLocalSpec loads a specification from a local file
func (c *TestContractCompatibilityChecker) loadLocalSpec(filePath string) (gjson.Result, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return gjson.Result{}, err
	}
	return c.parseYAMLToJSON(content)
}

// parseYAMLToJSON converts YAML content to JSON for gjson parsing
func (c *TestContractCompatibilityChecker) parseYAMLToJSON(yamlContent []byte) (gjson.Result, error) {
	var yamlData interface{}
	err := yaml.Unmarshal(yamlContent, &yamlData)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("failed to parse YAML: %w", err)
	}

	jsonData, err := json.Marshal(yamlData)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("failed to convert YAML to JSON: %w", err)
	}

	return gjson.ParseBytes(jsonData), nil
}

// GenerateReport generates a compatibility report
func (c *TestContractCompatibilityChecker) GenerateReport(result *CompatibilityResult) string {
	checker := &ContractCompatibilityChecker{}
	return checker.GenerateReport(result)
}