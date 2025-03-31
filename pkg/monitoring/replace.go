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
		if hm.Port == port.Number && hm.Host == host {

			//host

			//expected www.example.com:443/health, got https://www.example.com/health:443
			re := regexp.MustCompile(hm.ReplacePattern)
			result := re.ReplaceAllString(host, hm.ReplaceWith)

			parts := strings.SplitN(result, "/", 2) // Teilt in maximal zwei Teile

			if len(parts) == 2 {
				formated := fmt.Sprintf("%s:%d/%s", parts[0], port.Number, parts[1])
				// Port nur an den Host anhängen, dann den Pfad wieder ergänzen
				return formated
			}

			// Falls kein Pfad vorhanden ist, einfach den Port anhängen
			return fmt.Sprintf("%s:%d", result, port.Number)

		}
	}
	return fmt.Sprintf("%s:%d", host, port.Number)
}
