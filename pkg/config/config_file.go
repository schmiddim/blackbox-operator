package config

import (
	"fmt"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

type Config struct {
	LogLevel      string                `yaml:"logLevel"`
	Module        string                `yaml:"module"`
	Interval      monitoringv1.Duration `yaml:"interval"`
	ScrapeTimeout monitoringv1.Duration `yaml:"scrapeTimeout"`
	Selector      metav1.LabelSelector  `yaml:"selector"`
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
	config.Selector = metav1.LabelSelector{
		MatchLabels: map[string]string{"app.kubernetes.io/instance": "blackbox-exporter"},
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing YAML: %w", err)
	}

	return &config, nil
}
