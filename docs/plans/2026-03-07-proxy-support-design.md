# Network Proxy Support Design

**Date**: 2026-03-07
**Status**: Proposed
**Author**: Claude Code

## Overview

Add network proxy support for IM platform bots (Telegram, Discord, Feishu, DingTalk) to enable access in restricted network environments (e.g., mainland China).

## Requirements

### Functional Requirements
1. Support HTTP/HTTPS proxy protocols
2. Support SOCKS5 proxy protocol
3. Optional username/password authentication for proxies
4. Three-tier fallback hierarchy:
   - **Bot-level proxy** (highest priority)
   - **Global proxy** (middle priority)
   - **Environment variables** (fallback, lowest priority)

### Non-Functional Requirements
1. Support both HTTP REST API and WebSocket connections
2. Thread-safe HTTP client caching
3. Minimal code changes to existing bot adapters
4. Backward compatible (no proxy = use environment variables)

## Design

### Architecture

```
┌─────────────────────────────────────────────────────┐
│                   Engine                            │
│  - Creates ProxyManager                            │
│  - Passes to bot adapters                          │
└──────────────────┬──────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│              ProxyManager                           │
│  - Config: global + bot-level settings              │
│  - Cache: map[botType]*http.Client                  │
│  - Priority: bot > global > env vars                │
│  - Thread-safe (sync.RWMutex)                       │
└──────────────────┬──────────────────────────────────┘
                   │
                   ├──────────────┬──────────────┐
                   ▼              ▼              ▼
              ┌─────────┐  ┌─────────┐  ┌─────────┐
              │ Discord │  │Telegram │  │ Feishu  │
              │   Bot   │  │   Bot   │  │   Bot   │
              └─────────┘  └─────────┘  └─────────┘
```

### Configuration Structure

#### Global Proxy Config
```yaml
proxy:
  enabled: false  # Global switch, default false uses env vars
  type: "http"    # http, https, socks5
  url: "http://127.0.0.1:7890"
  username: ""   # Optional authentication
  password: ""   # Optional authentication
```

#### Bot-Level Proxy Config
```yaml
bots:
  telegram:
    enabled: true
    token: "xxx"
    proxy:
      enabled: false  # false=use global, true=use bot config
      type: "socks5"
      url: "socks5://127.0.0.1:1080"
      username: "user"  # Optional
      password: "pass"  # Optional
```

### Priority Logic

```
1. bots.<name>.proxy.enabled: true
   → Use bot-level proxy configuration

2. bots.<name>.proxy.enabled: false + proxy.enabled: true
   → Use global proxy configuration

3. Both enabled: false
   → Use environment variables (HTTP_PROXY, HTTPS_PROXY, ALL_PROXY)
```

### ProxyManager Implementation

```go
type ProxyManager struct {
    config  *Config
    mu      sync.RWMutex
    clients map[string]*http.Client  // Cache per bot type
}

// GetHTTPClient returns cached or creates new HTTP client for bot
func (pm *ProxyManager) GetHTTPClient(botType string) (*http.Client, error)

// GetProxyURL returns proxy URL for specific bot (for logging/debug)
func (pm *ProxyManager) GetProxyURL(botType string) string

// TestConnection validates proxy connection to bot API
func (pm *ProxyManager) TestConnection(botType string) error

// ClearCache clears all cached clients (for testing/reconfig)
func (pm *ProxyManager) ClearCache()
```

### HTTP Transport Creation

#### HTTP/HTTPS Proxy
```go
transport := &http.Transport{
    Proxy: http.ProxyURL(proxyURL),
    MaxIdleConns:        100,
    IdleConnTimeout:     90 * time.Second,
    TLSHandshakeTimeout: 10 * time.Second,
    ForceAttemptHTTP2:   true,
}
```

#### SOCKS5 Proxy
```go
import "golang.org/x/net/proxy"

dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, &proxy.Auth{
    User:     cfg.Username,
    Password: cfg.Password,
}, proxy.Direct)

transport := &http.Transport{
    DialContext: dialer.(proxy.ContextDialer).DialContext,
    MaxIdleConns:        100,
    IdleConnTimeout:     90 * time.Second,
    TLSHandshakeTimeout: 10 * time.Second,
}
```

### WebSocket Proxy Support

#### Mechanism
- **HTTP Proxy**: WebSocket connections tunnel through HTTP CONNECT
- **SOCKS5 Proxy**: Native WebSocket support

#### Bot Integration

**Discord (discordgo)**:
```go
client, err := d.proxyMgr.GetHTTPClient("discord")
session, err := discordgo.New(
    "Bot " + token,
    discordgo.WithHTTPClient(client),  // Handles HTTP + WebSocket
)
```

**Feishu (lark SDK)**:
```go
client, err := f.proxyMgr.GetHTTPClient("feishu")
f.larkClient = lark.NewClient(appID, appSecret,
    lark.WithHTTPClient(client),
)
f.wsClient = ws.NewClient(appID, appSecret,
    ws.WithHTTPClient(client),
)
```

**DingTalk**:
```go
client, err := d.proxyMgr.GetHTTPClient("dingtalk")
d.streamClient = client.NewClient(
    client.WithHTTPClient(client),
)
```

**Telegram (long polling, no WebSocket)**:
```go
client, err := t.proxyMgr.GetHTTPClient("telegram")
bot, err := tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, client)
```

## Data Types

```go
// Config structure additions
type Config struct {
    // ... existing fields
    Proxy ProxyConfig            `yaml:"proxy"`
    Bots  map[string]BotConfig   `yaml:"bots"`
}

type ProxyConfig struct {
    Enabled  bool   `yaml:"enabled"`
    Type     string `yaml:"type"`     // http, https, socks5
    URL      string `yaml:"url"`
    Username string `yaml:"username"` // Optional
    Password string `yaml:"password"` // Optional
}

type BotConfig struct {
    // ... existing fields
    Proxy *ProxyConfig `yaml:"proxy"` // Optional, nil = use global
}
```

## Error Handling

1. **Invalid proxy URL**: Return error at startup, prevent bot creation
2. **Proxy connection failure**: Log error, retry with backoff
3. **Authentication failure**: Log masked error, prevent bot creation
4. **WebSocket tunnel failure**: Log error, attempt reconnection

## Testing

### Unit Tests
1. `TestProxyManager_PriorityOrder` - Verify bot > global > env
2. `TestProxyManager_ClientCaching` - Verify cache mechanism
3. `TestProxyManager_HTTPProxy` - Test HTTP proxy creation
4. `TestProxyManager_SOCKS5Proxy` - Test SOCKS5 proxy creation
5. `TestProxyManager_Authentication` - Test proxy authentication
6. `TestProxyManager_TestConnection` - Test connection validation

### Integration Tests
1. Test Telegram bot with proxy
2. Test Discord bot with proxy (WebSocket)
3. Test Feishu bot with proxy (WebSocket)
4. Test environment variable fallback
5. Test proxy switching (bot-level override global)

### Manual Testing
```bash
# Setup local proxy
export HTTP_PROXY="http://127.0.0.1:7890"

# Test with global proxy
# config.yaml: proxy.enabled: true

# Test with bot-level proxy
# config.yaml: bots.telegram.proxy.enabled: true

# Test environment fallback
# config.yaml: proxy.enabled: false, bots.*.proxy.enabled: false
```

## Dependencies

```go
// go.mod additions
require (
    golang.org/x/net/proxy v0.16.0  // SOCKS5 support
)
```

## Migration Path

1. **Phase 1**: Implement ProxyManager and global proxy
2. **Phase 2**: Add bot-level proxy support
3. **Phase 3**: Add connection testing and diagnostics
4. **Phase 4**: Update documentation and examples

## Security Considerations

1. Mask proxy passwords in logs
2. Support environment variable expansion for proxy credentials
3. Validate proxy URL syntax at config load time
4. Prevent SSRF via proxy URL validation

## Documentation Updates

1. Update `config.full.yaml` with proxy examples
2. Add proxy setup guide to `docs/en/setup/proxy.md`
3. Update README with proxy section
4. Add troubleshooting section for common proxy issues

## Performance Impact

- **Client caching**: Minimal overhead, cache hit > 95%
- **Connection reuse**: Benefits from HTTP keep-alive
- **Memory**: ~1KB per cached client (negligible)

## Alternatives Considered

1. **Environment variables only**: Too inflexible, rejected
2. **Per-bot proxy only**: No global setting, rejected
3. **Manual proxy in each bot**: Code duplication, rejected
4. **Auto-detect proxy**: Unreliable, rejected

## Open Questions

1. Should we support PAC (Proxy Auto-Config) files? **Decision: No, out of scope**
2. Should we support proxy rotation/load balancing? **Decision: No, future enhancement**
3. Should we add proxy health monitoring? **Decision: Phase 3**

## Sign-off

- [x] Requirements clarified
- [x] Design reviewed
- [x] Implementation plan ready
- [ ] Implementation approved
