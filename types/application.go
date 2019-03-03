package types

import (
	"sort"
	"strings"
)

// JSONApplication is a helper for loading applications with string slices for Domains
type JSONApplication struct {
	Name            string   `json:"name"`
	ListenPort      string   `json:"listen-port"`
	ApplicationPort string   `json:"application-port"`
	Protocol        string   `json:"protocol"`
	Domains         []string `json:"domains"`
}

func (a *JSONApplication) ToApplication() Application {
	sort.Slice(a.Domains, func(i, j int) bool {
		return a.Domains[i] < a.Domains[j]
	})

	domainString := strings.Join(a.Domains, " ")
	return Application{
		Name: a.Name,
		ListenPort: a.ListenPort,
		ApplicationPort: a.ApplicationPort,
		Protocol: a.Protocol,
		domains: domainString,
	}
}

// Application is a service running on a backend. A backend will respond
// with a list of Applications when queried by a load balancer.
type Application struct {
	Name            string
	ListenPort      string
	ApplicationPort string
	Protocol        string
	domains         string
}

// ApplicationsEqual implements an equality check for two Applications
func (a *Application) Equal(other Application) bool {
	return a.Name == other.Name && a.ListenPort == other.ListenPort && a.ApplicationPort == other.ApplicationPort && a.Protocol == other.Protocol && a.domains == other.domains
}

// DomainSlice returns domains as a []string
func (a *Application) DomainSlice() []string {
	return strings.Split(a.domains, " ")
}

// DomainString returns the domain list as a string separated by `sep`
func (a *Application) DomainString(sep string) string {
	return strings.Join(a.DomainSlice(), sep)
}
