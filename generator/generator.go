package generator

import "github.com/sgtcodfish/scrimplb/types"

// Generator provides an interface for generating configuration values based on backend configuration
type Generator interface {
	GenerateConfig(types.UpstreamApplicationMap) (string, error)
}

// MakeApplicationMap converts a map of upstreams to applications into a map
// of applications to addresses
func MakeApplicationMap(val types.UpstreamApplicationMap) map[types.Application][]string {
	appMap := make(map[types.Application][]string)

	for k, v := range val {
		for _, app := range v {
			appMap[app] = append(appMap[app], k.Address)
		}
	}

	return appMap
}
