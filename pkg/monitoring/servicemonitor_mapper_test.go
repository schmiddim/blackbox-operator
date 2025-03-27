package monitoring

import (
	"github.com/go-logr/logr"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/schmiddim/blackbox-operator/pkg/config"
	"istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestModuleForProtocol(t *testing.T) {
	cfg := &config.Config{
		LogLevel:      "debug",
		DefaultModule: "http_test",
		Interval:      monitoringv1.Duration("10s"),
		ScrapeTimeout: monitoringv1.Duration("5s"),
		LabelSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{"app.kubernetes.io/name": "test-app"},
		},
		ProtocolModuleMappings: map[string]string{"TCP": "tcp_connect"},
	}

	logger := logr.Logger{}
	mapper := NewServiceMonitorMapper(cfg, &logger)

	servicePorts := v1alpha3.ServicePort{

		Number:   9093,
		Protocol: "TCP",
		Name:     "tcp",
	}
	module := mapper.getModuleForProtocol(&servicePorts)
	if module != "tcp_connect" {
		t.Errorf("expect tcp_connect, got %s", module)
	}

	servicePorts = v1alpha3.ServicePort{

		Number:   9093,
		Protocol: "fooobarto",
		Name:     "tcp",
	}
	module = mapper.getModuleForProtocol(&servicePorts)
	if module != cfg.DefaultModule {
		t.Errorf("expect %s, got %s", cfg.DefaultModule, module)
	}

}
