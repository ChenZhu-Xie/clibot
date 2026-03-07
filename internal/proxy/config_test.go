package proxy

// mockConfigProvider is a test implementation of ConfigProvider
type mockConfigProvider struct {
	globalEnabled bool
	globalURL     string
	globalUser    string
	globalPass    string
	botEnabled    bool
	botURL        string
	botUser       string
	botPass       string
}

func (m *mockConfigProvider) GetGlobalProxyEnabled() bool {
	return m.globalEnabled
}

func (m *mockConfigProvider) GetGlobalProxyURL() string {
	return m.globalURL
}

func (m *mockConfigProvider) GetGlobalProxyUsername() string {
	return m.globalUser
}

func (m *mockConfigProvider) GetGlobalProxyPassword() string {
	return m.globalPass
}

func (m *mockConfigProvider) GetBotProxyEnabled(botType string) bool {
	return m.botEnabled
}

func (m *mockConfigProvider) GetBotProxyURL(botType string) string {
	return m.botURL
}

func (m *mockConfigProvider) GetBotProxyUsername(botType string) string {
	return m.botUser
}

func (m *mockConfigProvider) GetBotProxyPassword(botType string) string {
	return m.botPass
}
