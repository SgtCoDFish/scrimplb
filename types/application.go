package types

// Application is a service running on a backend. A backend will respond
// with a list of Applications when queried by a load balancer.
type Application struct {
	Name     string `json:"name"`
	Domain   string `json:"domain"`
	Port     string `json:"port"`
	Protocol string `json:"protocol"`
}
