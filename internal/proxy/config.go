package proxy

import "net/http"

// ConfigProvider provides proxy configuration to ProxyManager
// This avoids circular dependency between core and proxy packages
type ConfigProvider interface {
	GetGlobalProxyEnabled() bool
	GetGlobalProxyURL() string
	GetGlobalProxyUsername() string
	GetGlobalProxyPassword() string
	GetBotProxyEnabled(botType string) bool
	GetBotProxyURL(botType string) string
	GetBotProxyUsername(botType string) string
	GetBotProxyPassword(botType string) string
}

// Manager defines the interface for proxy management
// This allows bot adapters to get HTTP clients with proxy configuration
type Manager interface {
	GetHTTPClient(botType string) (*http.Client, error)
	GetProxyURL(botType string) string
}
