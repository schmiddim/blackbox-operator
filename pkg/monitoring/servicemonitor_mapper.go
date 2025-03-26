package monitoring

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/schmiddim/blackbox-operator/pkg/config"
	istioNetworking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceMonitorMapper struct {
	config *config.Config
}

func NewServiceMonitorMapper(cfg *config.Config) *ServiceMonitorMapper {
	return &ServiceMonitorMapper{
		config: cfg,
	}
}

func (smm *ServiceMonitorMapper) generateEndpoints(hosts []string) []monitoringv1.Endpoint {
	var endpoints []monitoringv1.Endpoint
	for _, host := range hosts {
		e := monitoringv1.Endpoint{
			Interval:      smm.config.Interval,
			Port:          "http", //@todo port
			Scheme:        "http", //@todo scheme
			Path:          "/probe",
			ScrapeTimeout: smm.config.ScrapeTimeout,
			Params: map[string][]string{
				"module": {smm.config.Module},
				"target": {host},
			},
			RelabelConfigs: []monitoringv1.RelabelConfig{
				{
					SourceLabels: []monitoringv1.LabelName{"__param_target"},
					TargetLabel:  "instance",
				},
				{
					SourceLabels: []monitoringv1.LabelName{"__param_module"},
					TargetLabel:  "module",
				},
				{
					Action: "labeldrop",
					Regex:  "pod|service|container",
				},
				{
					SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_namespace"},
					TargetLabel:  "namespace",
				},
			},
		}
		endpoints = append(endpoints, e)

	}
	return endpoints
}

func (smm *ServiceMonitorMapper) MapperForService(se *istioNetworking.ServiceEntry) *monitoringv1.ServiceMonitor {
	sm := &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sm-" + se.Name,
			Namespace: se.Namespace,
			Labels: map[string]string{
				"managed-by": "blackbox-operator",
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			NamespaceSelector: monitoringv1.NamespaceSelector{
				Any: true,
			},
			Selector:  smm.config.LabelSelector,
			Endpoints: smm.generateEndpoints(se.Spec.Hosts),
		},
	}
	return sm
}
