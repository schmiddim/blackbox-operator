package config

import (
	"encoding/json"
	"fmt"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	yaml "sigs.k8s.io/yaml/goyaml.v3"
)

type Config struct {
	LogLevel                    string                `json:"logLevel"`
	DefaultModule               string                `json:"defaultModule"`
	ServiceMonitorNamingPattern string                `json:"serviceMonitorNamingPattern"`
	Interval                    monitoringv1.Duration `json:"interval"`
	ScrapeTimeout               monitoringv1.Duration `json:"scrapeTimeout"`
	HostMappings                []struct {
		Port           uint32 `yaml:"port,omitempty"`
		ReplacePattern string `yaml:"replacePattern"`
		ReplaceWith    string `yaml:"replaceWith"`
	} `json:"hostMappings,omitempty"`
	ModuleMappings []struct {
		Port          uint32 `yaml:"port,omitempty"`
		MatchPattern  string `yaml:"matchPattern"`
		ReplaceModule string `yaml:"replaceModule"`
	} `json:"moduleMappings,omitempty"`
	LabelSelector          metav1.LabelSelector `json:"selector"`
	ExcludeSelector        metav1.LabelSelector `json:"exclude,omitempty"`
	ProtocolModuleMappings map[string]string    `json:"protocolModuleMappings,omitempty"`
}

func LoadConfig(filePath string) (*Config, error) {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}
	var jsonData interface{}
	if err := yaml.Unmarshal(data, &jsonData); err != nil {
		return nil, err
	}
	result, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return nil, err
	}

	config := Config{}
	// Default Values
	config.DefaultModule = "http_2xx"
	config.LogLevel = "info"
	config.ScrapeTimeout = "30s"
	config.Interval = "30s"

	err = json.Unmarshal(result, &config)

	if err != nil {
		return nil, err
	}
	return &config, nil
}
