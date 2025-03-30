package monitoring

import (
	"github.com/go-logr/logr"
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/schmiddim/blackbox-operator/pkg/config"
	"istio.io/api/networking/v1alpha3"
	istioNetworking "istio.io/client-go/pkg/apis/networking/v1alpha3"
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

func TestServiceMonitorMapper(t *testing.T) {

	cfg := getCfg()
	logger := logr.Logger{}
	mapper := NewServiceMonitorMapper(&cfg, &logger)

	tests := []*struct {
		name           string
		serviceMonitor v1.ServiceMonitor
		serviceEntry   istioNetworking.ServiceEntry
		config         config.Config
	}{

		{
			name: "SmokeTest",
			serviceMonitor: v1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "sm-example-entry",
					Labels: nil,
				},
			},
			config: getCfg(),
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
						Protocol: "https",
						Name:     "https",
					}},
				},
			},
		},
	}

	for _, test := range tests {
		sm := mapper.MapperForService(&test.serviceEntry)
		if test.name == "smokeTest" {
			want := "sm-example-entry"
			got := sm.Name
			if got != want {
				t.Errorf("expected %s, got %s", want, got)
			}

			if len(sm.Spec.Endpoints) != 1 {
				t.Errorf("expected 1 endpoints, got %d", len(sm.Spec.Endpoints))
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
		name         string
		host         string
		servicePort  v1alpha3.ServicePort
		expectedHost string
	}{
		{
			name: "HTTPS with port 9093",
			host: "google.de",
			servicePort: v1alpha3.ServicePort{
				Number:   9093,
				Protocol: "HTTPS",
				Name:     "tcp",
			},
			expectedHost: "https://google.de:9093",
		},
		{
			name: "no Protocol",
			host: "test.com",
			servicePort: v1alpha3.ServicePort{
				Number:   8080,
				Protocol: "UNKNOWN",
				Name:     "unknown",
			},
			expectedHost: "test.com:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapper.getHost(tt.host, &tt.servicePort)
			if got != tt.expectedHost {
				t.Errorf("expected %s, got %s", tt.expectedHost, got)
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
