package monitoring

import "github.com/schmiddim/blackbox-operator/pkg/config"

type Excluded struct {
	cfg *config.Config
}

func NewExcluded(cfg *config.Config) *Excluded {
	return &Excluded{cfg: cfg}
}

func (e *Excluded) IsExcluded(labels map[string]string) bool {
	for k, v := range labels {
		for ek, ev := range e.cfg.ExcludeSelector.MatchLabels {
			if k == ek && v == ev {
				return true
			}
		}
	}
	return false
}
