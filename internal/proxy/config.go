package proxy

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
