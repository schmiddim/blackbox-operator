package monitoring

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/schmiddim/blackbox-operator/pkg/config"
	"istio.io/api/networking/v1alpha3"
	"regexp"
	"strings"
)

type Replace struct {
	cfg *config.Config
	log *logr.Logger
}

func NewReplace(cfg *config.Config, log *logr.Logger) *Replace {
	return &Replace{cfg: cfg, log: log}
}

func (r *Replace) GetModifiedModule(host string, port *v1alpha3.ServicePort) string {
	for _, mm := range r.cfg.ModuleMappings {
		re := regexp.MustCompile(mm.MatchPattern)
		if mm.Port == port.Number && re.MatchString(host) {
			return mm.ReplaceModule
		}
	}

	for protocol, module := range r.cfg.ProtocolModuleMappings {
		if strings.ToUpper(port.Protocol) == strings.ToUpper(protocol) {
			return module
		}
	}

	r.log.Info(fmt.Sprintf("No module for protocol %s - configuring Default (%s)", port.Protocol, r.cfg.DefaultModule))
	return r.cfg.DefaultModule
}

func (r *Replace) GetModifiedHostname(host string, port *v1alpha3.ServicePort) string {
	for _, hm := range r.cfg.HostMappings {
		re := regexp.MustCompile(hm.ReplacePattern)
		if hm.Port == port.Number && re.MatchString(host) {
			modified := strings.Replace(hm.ReplaceWith, "*", host[len(hm.ReplacePattern):], 1)
			parts := strings.SplitN(modified, "/", 2) // Teilt in maximal zwei Teile
			if len(parts) == 2 {
				formated := fmt.Sprintf("%s:%d/%s", parts[0], port.Number, parts[1])
				return formated
			}
			return fmt.Sprintf("%s:%d", modified, port.Number)
		}
	}
	return fmt.Sprintf("%s:%d", host, port.Number)
}
