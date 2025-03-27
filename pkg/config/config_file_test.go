package config

import (
	"os"
	"reflect"
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Test YAML data
const testYAML = `
logLevel: "debug"
defaultModule: "http_test"
interval: "10s"
scrapeTimeout: "5s"
selector:
  matchLabels:
    app.kubernetes.io/name: "test-app"
protocolModuleMappings:
  TCP: tcp_connect
exclude:
  matchLabels:
    blackbox-operator-scrape: false
`

// Helper function to create a temporary YAML file
func createTempFile(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "config_test_*.yaml")
	if err != nil {
		t.Fatalf("Error creating temporary file: %v", err)
	}
	defer tmpFile.Close()

	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatalf("Error writing to file: %v", err)
	}

	return tmpFile.Name()
}

// Test loading the configuration from YAML
func TestLoadConfig(t *testing.T) {
	// Create a temporary file with test YAML
	filePath := createTempFile(t, testYAML)
	defer os.Remove(filePath) // Clean up after test

	config, err := LoadConfig(filePath)
	if err != nil {
		t.Fatalf("Error loading configuration: %v", err)
	}

	// Expected values
	expectedConfig := &Config{
		LogLevel:      "debug",
		DefaultModule: "http_test",
		Interval:      monitoringv1.Duration("10s"),
		ScrapeTimeout: monitoringv1.Duration("5s"),
		LabelSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{"app.kubernetes.io/name": "test-app"},
		},
		ProtocolModuleMappings: map[string]string{"TCP": "tcp_connect"},
	}

	// Compare actual values with expected ones
	if config.LogLevel != expectedConfig.LogLevel {
		t.Errorf("Expected LogLevel: %s, got: %s", expectedConfig.LogLevel, config.LogLevel)
	}
	if config.DefaultModule != expectedConfig.DefaultModule {
		t.Errorf("Expected DefaultModule: %s, got: %s", expectedConfig.DefaultModule, config.DefaultModule)
	}
	if config.Interval != expectedConfig.Interval {
		t.Errorf("Expected Interval: %s, got: %s", expectedConfig.Interval, config.Interval)
	}
	if config.ScrapeTimeout != expectedConfig.ScrapeTimeout {
		t.Errorf("Expected ScrapeTimeout: %s, got: %s", expectedConfig.ScrapeTimeout, config.ScrapeTimeout)
	}
	if !reflect.DeepEqual(config.LabelSelector, expectedConfig.LabelSelector) {
		t.Errorf("Expected LabelSelector: %v, got: %v", expectedConfig.LabelSelector, config.LabelSelector)
	}
	if !reflect.DeepEqual(config.ProtocolModuleMappings, expectedConfig.ProtocolModuleMappings) {
		t.Errorf("Expected LabelSelector: %v, got: %v", expectedConfig.ProtocolModuleMappings, config.ProtocolModuleMappings)
	}
}
func TestExcludeSelector(t *testing.T) {
	filePath := createTempFile(t, testYAML)
	defer os.Remove(filePath)
	config, err := LoadConfig(filePath)
	if err != nil {
		t.Fatalf("Error loading default configuration: %v", err)
	}

	if len(config.ExcludeSelector.MatchLabels) == 0 {
		t.Errorf("ExcludeMatchLabels is empty")
	}
}

// Test default values when `selector` is not set
func TestLoadConfig_Defaults(t *testing.T) {
	const yamlWithoutSelector = `
logLevel: "warn"
module: "http_default"
interval: "20s"
scrapeTimeout: "15s"
`
	filePath := createTempFile(t, yamlWithoutSelector)
	defer os.Remove(filePath)

	config, err := LoadConfig(filePath)
	if err != nil {
		t.Fatalf("Error loading default configuration: %v", err)
	}

	if len(config.LabelSelector.MatchLabels) != 0 {
		t.Errorf("Expected no MatchLabels, got: %v", config.LabelSelector.MatchLabels)
	}
	if len(config.LabelSelector.MatchExpressions) != 0 {
		t.Errorf("Expected no MatchExpressions, got: %v", config.LabelSelector.MatchExpressions)
	}
}
