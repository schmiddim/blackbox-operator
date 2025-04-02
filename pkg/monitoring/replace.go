package monitoring

import (
	"fmt"
	"github.com/schmiddim/blackbox-operator/pkg/config"
	"istio.io/api/networking/v1alpha3"
	"regexp"
	"strings"
)

type Replace struct {
	cfg *config.Config
}

func NewReplace(cfg *config.Config) *Replace {
	return &Replace{cfg: cfg}
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
