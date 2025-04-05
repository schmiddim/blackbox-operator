package monitoring

import (
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/schmiddim/blackbox-operator/pkg/config"
	"github.com/schmiddim/blackbox-operator/test/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		se, err := utils.LoadServiceEntry(tt.serviceEntryFilename)
		if err != nil {
			t.Errorf("%s: loadServiceEntry failed: '%v'", tt.name, err)
		}

		serviceMonitor, err := utils.LoadServiceMonitor(tt.serviceEntryMonitor)
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
