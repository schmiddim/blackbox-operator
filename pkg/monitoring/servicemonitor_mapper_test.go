package monitoring

import (
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/schmiddim/blackbox-operator/pkg/config"
	"gopkg.in/yaml.v3"
	"istio.io/api/networking/v1alpha3"
	istioNetworking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
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

func loadServiceEntry(filename string) (*istioNetworking.ServiceEntry, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	type ServiceEntryWrapper struct {
		APIVersion string                       `yaml:"apiVersion"`
		Kind       string                       `yaml:"kind"`
		Metadata   metav1.ObjectMeta            `yaml:"metadata"`
		Spec       istioNetworking.ServiceEntry `yaml:"spec"`
	}

	var wrapper ServiceEntryWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	wrapper.Spec.ObjectMeta = wrapper.Metadata
	return &wrapper.Spec, nil
}
func loadServiceMonitor(filename string) (*v1.ServiceMonitor, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	type ServiceMonitorWrapper struct {
		APIVersion string            `yaml:"apiVersion"`
		Kind       string            `yaml:"kind"`
		Metadata   metav1.ObjectMeta `yaml:"metadata"`
		Spec       v1.ServiceMonitor `yaml:"spec"`
	}

	type LabelSelectorYAML struct {
		MatchLabels      map[string]string                 `yaml:"matchLabels,omitempty"`
		MatchExpressions []metav1.LabelSelectorRequirement `yaml:"selector"`
	}

	type YamlServiceMonitor struct {
		Spec struct {
			Endpoints []v1.Endpoint     `yaml:"endpoints"`
			Selector  LabelSelectorYAML `yaml:"selector"`
		} `yaml:"spec"`
	}

	var nwrapper YamlServiceMonitor
	var wrapper ServiceMonitorWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &nwrapper)
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	wrapper.Spec.ObjectMeta = wrapper.Metadata
	wrapper.Spec.Spec.Selector.MatchLabels = nwrapper.Spec.Selector.MatchLabels
	wrapper.Spec.Spec.NamespaceSelector.Any = true

	return &wrapper.Spec, nil
}

func TestLoadFromFS(t *testing.T) {
	tests := []*struct {
		name                 string
		configFileName       string
		serviceEntryFilename string
		serviceEntryMonitor  string
	}{
		{
			name:                 "Smoke Test",
			configFileName:       "./testdata/1-config.yaml",
			serviceEntryFilename: "./testdata/1-ServiceEntry.yaml",
			serviceEntryMonitor:  "./testdata/1-ServiceMonitor.yaml",
		},
	}
	for _, tt := range tests {
		se, err := loadServiceEntry(tt.serviceEntryFilename)
		serviceMonitor, err := loadServiceMonitor(tt.serviceEntryMonitor)
		if err != nil {
			t.Errorf("%s: loadServiceEntry failed: '%v'", tt.name, err)
		}
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
		if diff := cmp.Diff(serviceMonitor, generatedSm); diff != "" {
			t.Errorf("%s: ServiceMonitor mismatch (-want +got):\n%s", tt.name, diff)
		}
	}

}
func TestServiceMonitorMapper(t *testing.T) {

	tests := []*struct {
		name           string
		serviceMonitor v1.ServiceMonitor
		serviceEntry   istioNetworking.ServiceEntry
		config         config.Config
	}{
		{
			name: "HostReplaceTest",
			serviceMonitor: v1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "sm-example-entry",
					Labels: nil,
				},
			},
			config: config.Config{
				LogLevel:                    "debug",
				DefaultModule:               "http_test",
				ServiceMonitorNamingPattern: "buah-%s",
				Interval:                    "10s",
				ScrapeTimeout:               "5s",
				HostMappings: []config.HostMapping{{
					//ServiceEntryName: "example-entry",
					Host:           "www.example.com",
					Port:           443,
					ReplacePattern: `www.example.com`,
					ReplaceWith:    "www.example.com/health",
				},
				},
				LabelSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{"app.kubernetes.io/name": "test-app"},
				},
				ProtocolModuleMappings: map[string]string{"TCP": "tcp_connect"},
			},
			serviceEntry: istioNetworking.ServiceEntry{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{Name: "example-entry"},
				Spec: v1alpha3.ServiceEntry{
					Hosts: []string{
						"www.example.com",
					},
					//Addresses:        nil,
					Ports: []*v1alpha3.ServicePort{{
						Number:   443,
						Protocol: "htps",
						Name:     "ttps",
					}},
				},
			},
		},
	}

	for _, test := range tests {
		logger := logr.Logger{}
		mapper := NewServiceMonitorMapper(&test.config, &logger)
		sm := mapper.MapperForService(&test.serviceEntry)
		if test.name == "HostReplaceTest" {
			if len(sm.Spec.Endpoints) != 1 {
				t.Errorf("expected 1 endpoints, got %d", len(sm.Spec.Endpoints))
			}

			got := sm.Spec.Endpoints[0].Params["target"][0]
			want := "www.example.com:443/health"
			if got != want {
				t.Errorf("expected %s, got %s", want, got)
			}
		}

	}
}

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

func TestModuleForHost(t *testing.T) {
	cfg := getCfg()
	logger := logr.Logger{}
	mapper := NewServiceMonitorMapper(&cfg, &logger)
	tests := []*struct {
		name           string
		host           string
		servicePort    v1alpha3.ServicePort
		expectedHost   string
		expectedModule string
	}{
		{
			name: "HTTPS with port 9093",
			host: "google.de",
			servicePort: v1alpha3.ServicePort{
				Number:   9093,
				Protocol: "HTTPS",
				Name:     "tcp",
			},
			expectedHost:   "https://google.de:9093",
			expectedModule: cfg.DefaultModule,
		},
		{
			name: "no Protocol",
			host: "test.com",
			servicePort: v1alpha3.ServicePort{
				Number:   8080,
				Protocol: "UNKNOWN",
				Name:     "unknown",
			},
			expectedHost:   "test.com:8080",
			expectedModule: cfg.DefaultModule,
		},
	}

	for _, tt := range tests {
		serviceEntry := istioNetworking.ServiceEntry{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{Name: "example-entry"},
			Spec: v1alpha3.ServiceEntry{
				Hosts: []string{
					tt.host,
				},

				Ports: []*v1alpha3.ServicePort{&tt.servicePort},
			},
		}
		t.Run(tt.name, func(t *testing.T) {
			sm := mapper.MapperForService(&serviceEntry)
			got := sm.Spec.Endpoints[0].Params["target"][0]
			if got != tt.expectedHost {
				t.Errorf("expected %s, got %s", tt.expectedHost, got)
			}

			//module test
			gotModule := sm.Spec.Endpoints[0].Params["module"][0]
			if gotModule != tt.expectedModule {
				t.Errorf("expected %s, got %s", tt.expectedModule, gotModule)
			}

		})
	}
}
func TestModuleForProtocol(t *testing.T) {
	cfg := getCfg()
	logger := logr.Logger{}
	mapper := NewServiceMonitorMapper(&cfg, &logger)

	tests := []*struct {
		name           string
		servicePort    v1alpha3.ServicePort
		expectedModule string
	}{
		{
			name:           "TCP Protocol",
			servicePort:    v1alpha3.ServicePort{Number: 9093, Name: "tcp", Protocol: "TCP"},
			expectedModule: "tcp_connect",
		},
		{
			name:           "TCP Protocol not upper case",
			servicePort:    v1alpha3.ServicePort{Number: 9093, Name: "tcp", Protocol: "TcP"},
			expectedModule: "tcp_connect",
		},
		{
			name:           "Back to default",
			servicePort:    v1alpha3.ServicePort{Number: 9093, Name: "heinzi", Protocol: "HTTPS"},
			expectedModule: cfg.DefaultModule,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapper.getModuleForProtocol(&tt.servicePort)
			if got != tt.expectedModule {
				t.Errorf("expected %s, got %s", tt.expectedModule, got)
			}
		})
	}

}
