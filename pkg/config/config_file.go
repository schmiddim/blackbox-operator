package config

import (
	"fmt"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	yaml "sigs.k8s.io/yaml/goyaml.v3"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

type LabelSelectorYAML struct {
	MatchLabels      map[string]string                 `yaml:"matchLabels,omitempty"`
	MatchExpressions []metav1.LabelSelectorRequirement `yaml:"matchExpressions,omitempty"`
}

type Config struct {
	LogLevel               string                `yaml:"logLevel"`
	DefaultModule          string                `yaml:"defaultModule"`
	Interval               monitoringv1.Duration `yaml:"interval"`
	ScrapeTimeout          monitoringv1.Duration `yaml:"scrapeTimeout"`
	TmpSelector            LabelSelectorYAML     `yaml:"selector"`
	LabelSelector          metav1.LabelSelector  // `yaml:"selector"`
	TmpExclude             LabelSelectorYAML     `yaml:"exclude"`
	ExcludeSelector        metav1.LabelSelector
	ProtocolModuleMappings map[string]string `yaml:"protocolModuleMappings,omitempty"`
}

func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	// Default Values
	config.DefaultModule = "http_2xx"
	config.LogLevel = "info"
	config.ScrapeTimeout = "30s"
	config.Interval = "30s"

	err = yaml.Unmarshal(data, &config)

	config.LabelSelector = metav1.LabelSelector{
		MatchLabels:      config.TmpSelector.MatchLabels,
		MatchExpressions: config.TmpSelector.MatchExpressions,
	}

	config.ExcludeSelector = metav1.LabelSelector{
		MatchLabels:      config.TmpExclude.MatchLabels,
		MatchExpressions: config.TmpExclude.MatchExpressions,
	}

	if len(config.LabelSelector.MatchLabels) > 0 {
		config.TmpSelector = LabelSelectorYAML{
			MatchLabels:      map[string]string{"app.kubernetes.io/name": "blackbox-exporter"},
			MatchExpressions: nil,
		}
	}

	if err != nil {
		return nil, fmt.Errorf("error parsing YAML: %w", err)
	}

	return &config, nil
}
