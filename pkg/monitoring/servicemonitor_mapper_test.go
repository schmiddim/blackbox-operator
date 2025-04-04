package monitoring

import (
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/schmiddim/blackbox-operator/pkg/config"
	istioNetworking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	yaml "sigs.k8s.io/yaml/goyaml.v3"
	"testing"
)

func getCfg() config.Config {
	return config.Config{
		LogLevel:                    "debug",
		DefaultModule:               "http_test",
		ServiceMonitorNamingPattern: "buah-%s",
		Interval:                    "10s",
		ScrapeTimeout:               "5s",
		LabelSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{"app.kubernetes.io/name": "test-app"},
		},
		ProtocolModuleMappings: map[string]string{"TCP": "tcp_connect"},
	}
}

func loadServiceEntryJson(filename string) (*istioNetworking.ServiceEntry, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading YAML file:", err)
		return nil, err
	}
	var jsonData interface{}
	if err := yaml.Unmarshal(data, &jsonData); err != nil {
		fmt.Println("Error unmarshalling YAML:", err)
		return nil, err
	}
	result, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return nil, err
	}

	seJson := istioNetworking.ServiceEntry{}
	err = json.Unmarshal(result, &seJson)

	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}

	return &seJson, err
}

func loadServiceMonitorJson(filename string) (*v1.ServiceMonitor, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading YAML file:", err)
		return nil, err
	}
	var jsonData interface{}
	if err := yaml.Unmarshal(data, &jsonData); err != nil {
		fmt.Println("Error unmarshalling YAML:", err)
		return nil, err
	}
	result, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return nil, err
	}

	smJson := v1.ServiceMonitor{}
	err = json.Unmarshal(result, &smJson)

	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}
	return &smJson, err

}
func TestServiceEntries(t *testing.T) {
	tests := []*struct {
		name                 string
		configFileName       string
		serviceEntryFilename string
		serviceEntryMonitor  string
	}{
		{
			name:                 "1 Smoke Test",
			configFileName:       "./testdata/1-config.yaml",
			serviceEntryFilename: "./testdata/1-service-entry.yaml",
			serviceEntryMonitor:  "./testdata/1-service-monitor.yaml",
		},
		{
			name:                 "2 No Probe for Port",
			configFileName:       "./testdata/2-config.yaml",
			serviceEntryFilename: "./testdata/2-service-entry.yaml",
			serviceEntryMonitor:  "./testdata/2-service-monitor.yaml",
		},
		{
			name:                 "3 Test Rewrite Urls by Regex",
			configFileName:       "./testdata/3-config.yaml",
			serviceEntryFilename: "./testdata/3-service-entry.yaml",
			serviceEntryMonitor:  "./testdata/3-service-monitor.yaml",
		},
		{
			name:                 "4 Test Module Overwrite",
			configFileName:       "./testdata/4-config.yaml",
			serviceEntryFilename: "./testdata/4-service-entry.yaml",
			serviceEntryMonitor:  "./testdata/4-service-monitor.yaml",
		},
		{
			name:                 "5 Test Module Naming",
			configFileName:       "./testdata/5-config.yaml",
			serviceEntryFilename: "./testdata/5-service-entry.yaml",
			serviceEntryMonitor:  "./testdata/5-service-monitor.yaml",
		},
	}
	for _, tt := range tests {
		se, err := loadServiceEntryJson(tt.serviceEntryFilename)
		if err != nil {
			t.Errorf("%s: loadServiceEntry failed: '%v'", tt.name, err)
		}

		serviceMonitor, err := loadServiceMonitorJson(tt.serviceEntryMonitor)
		if err != nil {
			t.Errorf("%s: loadServiceMonitor failed: '%v'", tt.name, err)
		}
		cfg, err := config.LoadConfig(tt.configFileName)
		if err != nil {
			t.Errorf("%s: loadServiceEntry failed: '%v'", tt.name, err)
		}

		smm := ServiceMonitorMapper{
			config: cfg,
			log:    &(logr.Logger{}),
		}
		generatedSm := smm.MapperForService(se)
		generatedSm.TypeMeta.Kind = serviceMonitor.Kind
		generatedSm.TypeMeta.Kind = serviceMonitor.Kind
		generatedSm.TypeMeta.APIVersion = serviceMonitor.APIVersion

		if diff := cmp.Diff(serviceMonitor, generatedSm); diff != "" {
			t.Errorf("%s: ServiceMonitor mismatch (-want +got):\n%s", tt.name, diff)
		}

	}

}

// @todo move to file  tt
func TestNamingPattern(t *testing.T) {
	cfg := getCfg()
	logger := logr.Logger{}
	mapper := NewServiceMonitorMapper(&cfg, &logger)

	got, _ := mapper.GetNameForServiceMonitor("hansi")
	want := "buah-hansi"
	if got != want {
		t.Errorf("expected %s, got %s", want, got)
	}

	cfg.ServiceMonitorNamingPattern = "invalid"
	_, err := mapper.GetNameForServiceMonitor("hansi")
	if err == nil {
		t.Errorf("expected %s, got %s", err, "nil")
	}
}
