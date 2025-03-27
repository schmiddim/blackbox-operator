package config

import (
	"fmt"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

type LabelSelectorYAML struct {
	MatchLabels      map[string]string                 `yaml:"matchLabels,omitempty"`
	MatchExpressions []metav1.LabelSelectorRequirement `yaml:"matchExpressions,omitempty"`
}

type Config struct {
	LogLevel      string                `yaml:"logLevel"`
	Module        string                `yaml:"module"`
	Interval      monitoringv1.Duration `yaml:"interval"`
	ScrapeTimeout monitoringv1.Duration `yaml:"scrapeTimeout"`
	TmpSelector   LabelSelectorYAML     `yaml:"selector"`
	LabelSelector metav1.LabelSelector  // `yaml:"selector"`
}

func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	// Default Values
	config.Module = "http_2xx"
	config.LogLevel = "info"
	config.ScrapeTimeout = "30s"
	config.Interval = "30s"

	err = yaml.Unmarshal(data, &config)

	config.LabelSelector = metav1.LabelSelector{
		MatchLabels:      config.TmpSelector.MatchLabels,
		MatchExpressions: config.TmpSelector.MatchExpressions,
	}

	if len(config.LabelSelector.MatchLabels) > 0 {
		config.TmpSelector = LabelSelectorYAML{
			MatchLabels:      map[string]string{"app.kubernetes.io/name": "blackbox-exporter"},
			MatchExpressions: nil,
		}
	}
	fmt.Println(config.LabelSelector, len(config.LabelSelector.MatchLabels), len(config.LabelSelector.MatchExpressions))

	if err != nil {
		return nil, fmt.Errorf("error parsing YAML: %w", err)
	}

	return &config, nil
}
