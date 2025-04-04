package monitoring

import (
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/schmiddim/blackbox-operator/pkg/config"
	"istio.io/api/networking/v1alpha3"
	istioNetworking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
)

type ServiceMonitorMapper struct {
	config *config.Config
	log    *logr.Logger
}

func NewServiceMonitorMapper(cfg *config.Config, log *logr.Logger) *ServiceMonitorMapper {
	return &ServiceMonitorMapper{
		config: cfg,
		log:    log,
	}
}

func (smm *ServiceMonitorMapper) GetNameForServiceMonitor(ServiceEntryName string) (string, error) {

	count := strings.Count(smm.config.ServiceMonitorNamingPattern, "%s")
	if count != 1 {
		return "", errors.New("ServiceMonitorNamingPattern must contain 1 %s")
	}
	name := fmt.Sprintf(smm.config.ServiceMonitorNamingPattern, ServiceEntryName)
	return name, nil
}

func (smm *ServiceMonitorMapper) isPortIgnored(port *v1alpha3.ServicePort, labels map[string]string) bool {
	for key, value := range labels {
		if key == "skip-probe-for-port" && value == strconv.FormatUint(uint64(port.Number), 10) {
			return true
		}

	}
	return false
}
func (smm *ServiceMonitorMapper) generateEndpoints(hosts []string, ports []*v1alpha3.ServicePort, labels map[string]string) (endpoints []monitoringv1.Endpoint, labelsForModifications map[string]string) {
	labelsForModifications = make(map[string]string)

	replace := NewReplace(smm.config, smm.log)
	for _, port := range ports {
		if smm.isPortIgnored(port, labels) {
			continue
		}
		for _, host := range hosts {

			hostWithPort := replace.GetModifiedHostname(host, port)
			modifiedModule, labelsFromModule := replace.GetModifiedModule(host, port)
			for k, v := range labelsFromModule {
				labelsForModifications[k] = v
			}

			if strings.ToUpper(port.GetProtocol()) == "HTTPS" {
				hostWithPort = fmt.Sprintf("https://%s", hostWithPort)
			}
			e := monitoringv1.Endpoint{
				Interval:      smm.config.Interval,
				Port:          "http",
				Scheme:        "http",
				Path:          "/probe",
				ScrapeTimeout: smm.config.ScrapeTimeout,
				Params: map[string][]string{
					"module": {modifiedModule},
					"target": {hostWithPort},
				},
				RelabelConfigs: []monitoringv1.RelabelConfig{
					{
						SourceLabels: []monitoringv1.LabelName{"__param_target"},
						TargetLabel:  "instance",
						Action:       "replace",
					},
					{
						SourceLabels: []monitoringv1.LabelName{"__param_module"},
						TargetLabel:  "module",
						Action:       "replace",
					},
					{
						Action: "labeldrop",
						Regex:  "pod|service|container",
					},
					{
						SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_namespace"},
						TargetLabel:  "namespace",
						Action:       "replace",
					},
				},
			}
			endpoints = append(endpoints, e)

		}
	}
	return endpoints, labelsForModifications
}

func (smm *ServiceMonitorMapper) MapperForService(se *istioNetworking.ServiceEntry) *monitoringv1.ServiceMonitor {

	endpoints, additionalLabels := smm.generateEndpoints(se.Spec.Hosts, se.Spec.Ports, se.ObjectMeta.Labels)
	labels := map[string]string{
		"managed-by": "blackbox-operator",
		"for":        se.Name,
	}
	for k, v := range additionalLabels {
		labels[k] = v
	}

	sm := &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sm-" + se.Name,
			Namespace: se.Namespace,
			Labels:    labels,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			NamespaceSelector: monitoringv1.NamespaceSelector{
				Any: true,
			},

			Selector:  smm.config.LabelSelector,
			Endpoints: endpoints,
		},
	}

	return sm
}
