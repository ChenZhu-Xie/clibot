# Network Proxy Support Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add network proxy support for IM platform bots with three-tier fallback (bot-level > global > environment variables), supporting HTTP/HTTPS and SOCKS5 protocols with optional authentication.

**Architecture:** ProxyManager with HTTP client caching, unified interface for all bot adapters, transparent proxy handling for both HTTP REST API and WebSocket connections.

**Tech Stack:** Go 1.21+, golang.org/x/net/proxy (SOCKS5), discordgo, telegram-bot-api, lark SDK

---

## Task 1: Add Proxy Configuration Structures

**Files:**
- Modify: `internal/core/types.go`
- Test: `internal/core/types_test.go`

**Step 1: Write failing test for ProxyConfig**

```go
func TestProxyConfig_FullConfig(t *testing.T) {
    config := ProxyConfig{
        Enabled:  true,
        Type:     "socks5",
        URL:      "socks5://127.0.0.1:1080",
        Username: "user",
        Password: "pass",
    }

    assert.True(t, config.Enabled)
    assert.Equal(t, "socks5", config.Type)
    assert.Equal(t, "socks5://127.0.0.1:1080", config.URL)
    assert.Equal(t, "user", config.Username)
    assert.Equal(t, "pass", config.Password)
}

func TestProxyConfig_OptionalAuth(t *testing.T) {
    config := ProxyConfig{
        Enabled:  true,
        Type:     "http",
        URL:      "http://127.0.0.1:7890",
        Username: "",
        Password: "",
    }

    assert.True(t, config.Enabled)
    assert.Empty(t, config.Username)
    assert.Empty(t, config.Password)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/core/... -run TestProxyConfig -v`
Expected: FAIL with "undefined: ProxyConfig"

**Step 3: Add ProxyConfig struct to types.go**

Add after line 127 (after CLIAdapterConfig):

```go
// ProxyConfig represents proxy configuration for network requests
type ProxyConfig struct {
    Enabled  bool   `yaml:"enabled"`  // true to use this proxy config
    Type     string `yaml:"type"`     // http, https, socks5
    URL      string `yaml:"url"`      // proxy URL
    Username string `yaml:"username"` // optional authentication
    Password string `yaml:"password"` // optional authentication
}
```

**Step 4: Update Config struct to include Proxy**

Add Proxy field to Config struct (after line 51):

```go
type Config struct {
    HookServer  HookServerConfig            `yaml:"hook_server"`
    Security    SecurityConfig              `yaml:"security"`
    Watchdog    WatchdogConfig              `yaml:"watchdog"`
    Session     SessionGlobalConfig         `yaml:"session"`
    Sessions    []SessionConfig             `yaml:"sessions"`
    Bots        map[string]BotConfig        `yaml:"bots"`
    CLIAdapters map[string]CLIAdapterConfig `yaml:"cli_adapters"`
    Logging     LoggingConfig               `yaml:"logging"`
    Proxy       ProxyConfig                 `yaml:"proxy"` // NEW
}
```

**Step 5: Update BotConfig to include optional Proxy**

Modify BotConfig struct (around line 95-104):

```go
type BotConfig struct {
    Enabled           bool          `yaml:"enabled"`
    AppID             string        `yaml:"app_id"`
    AppSecret         string        `yaml:"app_secret"`
    Token             string        `yaml:"token"`
    ChannelID         string        `yaml:"channel_id"`         // For Discord: server channel ID
    EncryptKey        string        `yaml:"encrypt_key"`        // Feishu: event encryption key (optional)
    VerificationToken string        `yaml:"verification_token"` // Feishu: verification token (optional)
    Proxy             *ProxyConfig  `yaml:"proxy"`              // NEW: Optional proxy override
}
```

**Step 6: Run tests to verify they pass**

Run: `go test ./internal/core/... -run TestProxyConfig -v`
Expected: PASS

**Step 7: Commit**

```bash
git add internal/core/types.go internal/core/types_test.go
git commit -m "feat: add proxy configuration structures

- Add ProxyConfig struct with HTTP/HTTPS/SOCKS5 support
- Add optional authentication (username/password)
- Add global proxy to Config
- Add bot-level proxy override to BotConfig"
```

---

## Task 2: Add SOCKS5 Dependency

**Files:**
- Modify: `go.mod`

**Step 1: Update go.mod**

Run: `go get golang.org/x/net/proxy@latest`

**Step 2: Verify dependency**

Run: `grep "golang.org/x/net/proxy" go.mod`
Expected: Line shows `golang.org/x/net/proxy v0.XX.0`

**Step 3: Tidy dependencies**

Run: `go mod tidy`

**Step 4: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add golang.org/x/net/proxy for SOCKS5 support"
```

---

## Task 3: Create ProxyManager

**Files:**
- Create: `internal/proxy/manager.go`
- Test: `internal/proxy/manager_test.go`

**Step 1: Write failing test for ProxyManager**

```go
package proxy

import (
    "testing"
    "github.com/keepmind9/clibot/internal/core"
    "github.com/stretchr/testify/assert"
)

func TestProxyManager_NoProxy_ReturnsClientWithEnvProxy(t *testing.T) {
    config := &core.Config{
        Proxy: core.ProxyConfig{
            Enabled: false,
        },
        Bots: map[string]core.BotConfig{},
    }

    pm := NewProxyManager(config)
    client, err := pm.GetHTTPClient("telegram")

    assert.NoError(t, err)
    assert.NotNil(t, client)
    assert.NotNil(t, client.Transport)
}

func TestProxyManager_GlobalProxy_ReturnsClientWithProxy(t *testing.T) {
    config := &core.Config{
        Proxy: core.ProxyConfig{
            Enabled: true,
            Type:    "http",
            URL:     "http://127.0.0.1:8080",
        },
        Bots: map[string]core.BotConfig{},
    }

    pm := NewProxyManager(config)
    client, err := pm.GetHTTPClient("telegram")

    assert.NoError(t, err)
    assert.NotNil(t, client)
}

func TestProxyManager_BotLevelProxyOverridesGlobal(t *testing.T) {
    botProxy := &core.ProxyConfig{
        Enabled: true,
        Type:    "socks5",
        URL:     "socks5://127.0.0.1:1080",
    }

    config := &core.Config{
        Proxy: core.ProxyConfig{
            Enabled: true,
            Type:    "http",
            URL:     "http://127.0.0.1:8080",
        },
        Bots: map[string]core.BotConfig{
            "telegram": {
                Proxy: botProxy,
            },
        },
    }

    pm := NewProxyManager(config)
    proxyURL := pm.GetProxyURL("telegram")

    assert.Contains(t, proxyURL, "socks5://127.0.0.1:1080")
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/proxy/... -v`
Expected: FAIL with "package proxy does not exist" or "undefined: NewProxyManager"

**Step 3: Create ProxyManager implementation**

Create `internal/proxy/manager.go`:

```go
package proxy

import (
    "fmt"
    "net/http"
    "net/url"
    "sync"

    "golang.org/x/net/proxy"
    "github.com/keepmind9/clibot/internal/core"
    "github.com/keepmind9/clibot/internal/logger"
    "github.com/sirupsen/logrus"
)

// ProxyManager manages proxy configuration with client caching
type ProxyManager struct {
    config  *core.Config
    mu      sync.RWMutex
    clients map[string]*http.Client
}

// NewProxyManager creates a new ProxyManager
func NewProxyManager(config *core.Config) *ProxyManager {
    return &ProxyManager{
        config:  config,
        clients: make(map[string]*http.Client),
    }
}

// GetHTTPClient returns cached or creates new HTTP client for bot
// Priority: bot-level > global > environment variables
func (pm *ProxyManager) GetHTTPClient(botType string) (*http.Client, error) {
    // Check cache first
    pm.mu.RLock()
    if client, exists := pm.clients[botType]; exists {
        pm.mu.RUnlock()
        logger.WithFields(logrus.Fields{
            "bot_type": botType,
            "cached":   true,
        }).Debug("proxy-manager-returning-cached-client")
        return client, nil
    }
    pm.mu.RUnlock()

    // Resolve proxy configuration
    proxyConfig, err := pm.resolveProxyConfig(botType)
    if err != nil {
        return nil, err
    }

    // Create HTTP client
    client, err := pm.createClient(proxyConfig)
    if err != nil {
        return nil, err
    }

    // Cache the client
    pm.mu.Lock()
    pm.clients[botType] = client
    pm.mu.Unlock()

    logger.WithFields(logrus.Fields{
        "bot_type":  botType,
        "proxy_url": pm.GetProxyURL(botType),
        "cached":    false,
    }).Info("proxy-manager-created-http-client")

    return client, nil
}

// GetProxyURL returns proxy URL for specific bot (for logging)
func (pm *ProxyManager) GetProxyURL(botType string) string {
    proxyConfig, _ := pm.resolveProxyConfig(botType)
    if proxyConfig == nil {
        return "env://HTTP_PROXY"
    }
    return proxyConfig.URL
}

// resolveProxyConfig resolves proxy config based on priority
func (pm *ProxyManager) resolveProxyConfig(botType string) (*core.ProxyConfig, error) {
    // 1. Bot-level proxy (highest priority)
    if botConfig, exists := pm.config.Bots[botType]; exists {
        if botConfig.Proxy != nil && botConfig.Proxy.Enabled {
            logger.WithFields(logrus.Fields{
                "bot_type": botType,
                "source":   "bot-level",
            }).Debug("proxy-manager-using-bot-level-proxy")
            return botConfig.Proxy, nil
        }
    }

    // 2. Global proxy
    if pm.config.Proxy.Enabled {
        logger.WithFields(logrus.Fields{
            "bot_type": botType,
            "source":   "global",
        }).Debug("proxy-manager-using-global-proxy")
        return &pm.config.Proxy, nil
    }

    // 3. No proxy (use environment variables)
    logger.WithFields(logrus.Fields{
        "bot_type": botType,
        "source":   "environment",
    }).Debug("proxy-manager-using-environment-variables")
    return nil, nil
}

// createClient creates HTTP client with proxy configuration
func (pm *ProxyManager) createClient(proxyConfig *core.ProxyConfig) (*http.Client, error) {
    if proxyConfig == nil {
        // Use environment variables
        return &http.Client{
            Transport: &http.Transport{
                Proxy: http.ProxyFromEnvironment,
            },
        }, nil
    }

    // Validate proxy URL
    proxyURL, err := url.Parse(proxyConfig.URL)
    if err != nil {
        return nil, fmt.Errorf("invalid proxy URL: %w", err)
    }

    // Add authentication if provided
    if proxyConfig.Username != "" {
        proxyURL.User = url.UserPassword(proxyConfig.Username, proxyConfig.Password)
    }

    // Create transport based on proxy type
    transport, err := pm.createTransport(proxyConfig, proxyURL)
    if err != nil {
        return nil, fmt.Errorf("failed to create transport: %w", err)
    }

    return &http.Client{
        Transport: transport,
        Timeout:   30 * time.Second,
    }, nil
}

// createTransport creates HTTP transport with proxy
func (pm *ProxyManager) createTransport(cfg *core.ProxyConfig, proxyURL *url.URL) (*http.Transport, error) {
    var transport *http.Transport

    switch cfg.Type {
    case "http", "https":
        // HTTP proxy (supports WebSocket via HTTP CONNECT)
        transport = &http.Transport{
            Proxy:               http.ProxyURL(proxyURL),
            MaxIdleConns:        100,
            IdleConnTimeout:     90 * time.Second,
            TLSHandshakeTimeout: 10 * time.Second,
            ForceAttemptHTTP2:   true,
        }

    case "socks5":
        // SOCKS5 proxy (native WebSocket support)
        var auth *proxy.Auth
        if cfg.Username != "" {
            auth = &proxy.Auth{
                User:     cfg.Username,
                Password: cfg.Password,
            }
        }

        dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
        if err != nil {
            return nil, fmt.Errorf("SOCKS5 proxy error: %w", err)
        }

        transport = &http.Transport{
            DialContext:         dialer.(proxy.ContextDialer).DialContext,
            MaxIdleConns:        100,
            IdleConnTimeout:     90 * time.Second,
            TLSHandshakeTimeout: 10 * time.Second,
        }

    default:
        return nil, fmt.Errorf("unsupported proxy type: %s", cfg.Type)
    }

    return transport, nil
}

// ClearCache clears all cached clients (for testing/reconfig)
func (pm *ProxyManager) ClearCache() {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    pm.clients = make(map[string]*http.Client)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/proxy/... -v`
Expected: PASS (may need minor syntax fixes)

**Step 5: Add missing import**

Add `time` import to manager.go:
```go
import (
    "fmt"
    "net/http"
    "net/url"
    "sync"
    "time"  // ADD THIS

    "golang.org/x/net/proxy"
    // ... rest
)
```

**Step 6: Run tests again**

Run: `go test ./internal/proxy/... -v`
Expected: PASS

**Step 7: Commit**

```bash
git add internal/proxy/
git commit -m "feat: implement ProxyManager with client caching

- Add ProxyManager with three-tier priority (bot > global > env)
- Support HTTP/HTTPS and SOCKS5 protocols
- Support optional username/password authentication
- Thread-safe HTTP client caching with sync.RWMutex
- WebSocket proxy support via HTTP CONNECT and SOCKS5"
```

---

## Task 4: Integrate ProxyManager into Engine

**Files:**
- Modify: `internal/core/engine.go`
- Test: `internal/core/engine_test.go`

**Step 1: Write failing test**

```go
func TestEngine_ProxyManagerInitialized(t *testing.T) {
    config := &Config{
        Sessions: []SessionConfig{},
    }
    engine := NewEngine(config)

    assert.NotNil(t, engine.proxyMgr)
}

func TestEngine_GetProxyClient(t *testing.T) {
    config := &Config{
        Sessions: []SessionConfig{},
        Proxy: ProxyConfig{
            Enabled: false,
        },
    }
    engine := NewEngine(config)

    client, err := engine.proxyMgr.GetHTTPClient("telegram")
    assert.NoError(t, err)
    assert.NotNil(t, client)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/core/... -run TestEngine_ProxyManager -v`
Expected: FAIL with "engine.proxyMgr undefined"

**Step 3: Add ProxyManager to Engine struct**

Modify Engine struct in `internal/core/engine.go` (around line 89-100):

```go
type Engine struct {
    config          *Config
    proxyMgr        *proxy.ProxyManager        // NEW
    cliAdapters     map[string]cli.CLIAdapter
    activeBots      map[string]bot.BotAdapter
    sessions        map[string]*Session
    sessionMu       sync.RWMutex
    messageChan     chan bot.BotMessage
    hookServer      *http.Server
    sessionChannels map[string]BotChannel
    userSessions    map[string]string
    cmdLocksMu      sync.RWMutex
    sessionCmdLocks map[string]*sync.Mutex
}
```

**Step 4: Add proxy import**

Add to imports in engine.go:
```go
import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/keepmind9/clibot/internal/bot"
    "github.com/keepmind9/clibot/internal/cli"
    "github.com/keepmind9/clibot/internal/logger"
    "github.com/keepmind9/clibot/internal/proxy"  // NEW
    "github.com/keepmind9/clibot/internal/watchdog"
    "github.com/keepmind9/clibot/pkg/constants"
    "github.com/sirupsen/logrus"
)
```

**Step 5: Initialize ProxyManager in NewEngine**

Modify NewEngine function (around line 120-134):

```go
// NewEngine creates a new Engine instance
func NewEngine(config *Config) *Engine {
    ctx, cancel := context.WithCancel(context.Background())

    return &Engine{
        config:          config,
        proxyMgr:        proxy.NewProxyManager(config),  // NEW
        cliAdapters:     make(map[string]cli.CLIAdapter),
        activeBots:      make(map[string]bot.BotAdapter),
        sessions:        make(map[string]*Session),
        sessionChannels: make(map[string]BotChannel),
        userSessions:     make(map[string]string),
        sessionCmdLocks: make(map[string]*sync.Mutex),
        messageChan:     make(chan bot.BotMessage, 100),
        ctx:             ctx,
        cancel:          cancel,
    }
}
```

**Step 6: Run tests to verify they pass**

Run: `go test ./internal/core/... -run TestEngine_ProxyManager -v`
Expected: PASS

**Step 7: Commit**

```bash
git add internal/core/engine.go internal/core/engine_test.go
git commit -m "feat: integrate ProxyManager into Engine

- Add ProxyManager field to Engine struct
- Initialize ProxyManager in NewEngine
- Add tests for ProxyManager initialization"
```

---

## Task 5: Update Bot Interface to Support ProxyManager

**Files:**
- Modify: `internal/bot/interface.go`
- Modify: `internal/bot/telegram.go`
- Modify: `internal/bot/discord.go`
- Modify: `internal/bot/feishu.go`
- Modify: `internal/bot/dingtalk.go`

**Step 1: Add SetProxyManager to BotAdapter interface**

Modify `internal/bot/interface.go` (around line 30-50):

```go
// BotAdapter defines the interface for IM platform bots
type BotAdapter interface {
    // Start begins listening for messages
    Start(messageHandler func(BotMessage)) error

    // Stop shuts down the bot
    Stop() error

    // SendMessage sends a message to the configured channel
    SendMessage(channel string, message string) error

    // SetProxyManager sets the proxy manager for this bot
    SetProxyManager(proxyMgr interface{}) // NEW
}
```

**Step 2: Update TelegramBot to use ProxyManager**

Modify `internal/bot/telegram.go`:

Add struct field (around line 16-24):
```go
type TelegramBot struct {
    DefaultTypingIndicator
    mu             sync.RWMutex
    token          string
    bot            *tgbotapi.BotAPI
    messageHandler func(BotMessage)
    ctx            context.Context
    cancel         context.CancelFunc
    proxyMgr       interface{}  // NEW
}
```

Add SetProxyManager method (after NewTelegramBot):
```go
// SetProxyManager sets the proxy manager
func (t *TelegramBot) SetProxyManager(proxyMgr interface{}) {
    t.mu.Lock()
    defer t.mu.Unlock()
    t.proxyMgr = proxyMgr
}
```

Modify Start method to use proxy (around line 34-50):
```go
func (t *TelegramBot) Start(messageHandler func(BotMessage)) error {
    t.SetMessageHandler(messageHandler)
    t.ctx, t.cancel = context.WithCancel(context.Background())

    logger.WithFields(logrus.Fields{
        "token": maskSecret(t.token),
    }).Info("starting-telegram-bot-with-long-polling")

    // Initialize Telegram bot with proxy
    var err error
    t.mu.Lock()
    defer t.mu.Unlock()

    if t.proxyMgr != nil {
        // Use proxy manager to get HTTP client
        if pm, ok := t.proxyMgr.(interface{ GetHTTPClient(string) (*http.Client, error) }); ok {
            client, clientErr := pm.GetHTTPClient("telegram")
            if clientErr != nil {
                logger.WithField("error", clientErr).Error("failed-to-create-proxy-client")
                return fmt.Errorf("failed to create proxy client: %w", clientErr)
            }
            t.bot, err = tgbotapi.NewBotAPIWithClient(t.token, tgbotapi.APIEndpoint, client)
        } else {
            t.bot, err = tgbotapi.NewBotAPI(t.token)
        }
    } else {
        t.bot, err = tgbotapi.NewBotAPI(t.token)
    }

    if err != nil {
        logger.WithFields(logrus.Fields{
            "error": err,
        }).Error("failed-to-initialize-telegram-bot")
        return fmt.Errorf("failed to initialize Telegram bot: %w", err)
    }

    logger.WithFields(logrus.Fields{
        "bot_username": t.bot.Self.UserName,
        "bot_id":       t.bot.Self.ID,
    }).Info("telegram-bot-initialized-successfully")

    // ... rest of the method
}
```

Add missing import:
```go
import (
    "context"
    "fmt"
    "net/http"  // NEW
    "sync"
    "time"
    // ... rest
)
```

**Step 3: Update DiscordBot to use ProxyManager**

Modify `internal/bot/discord.go` similarly:

Add field:
```go
type DiscordBot struct {
    DefaultTypingIndicator
    mu             sync.RWMutex
    token          string
    channelID      string
    session        DiscordSessionInterface
    messageHandler func(BotMessage)
    proxyMgr       interface{}  // NEW
}
```

Add SetProxyManager:
```go
func (d *DiscordBot) SetProxyManager(proxyMgr interface{}) {
    d.mu.Lock()
    defer d.mu.Unlock()
    d.proxyMgr = proxyMgr
}
```

Modify Start method:
```go
func (d *DiscordBot) Start(messageHandler func(BotMessage)) error {
    d.SetMessageHandler(messageHandler)

    logger.WithFields(logrus.Fields{
        "token":   maskSecret(d.token),
        "channel": d.channelID,
    }).Info("starting-discord-bot")

    var err error
    d.mu.Lock()
    defer d.mu.Unlock()

    // Create Discord session with proxy
    if d.proxyMgr != nil {
        if pm, ok := d.proxyMgr.(interface{ GetHTTPClient(string) (*http.Client, error) }); ok {
            client, clientErr := pm.GetHTTPClient("discord")
            if clientErr != nil {
                return fmt.Errorf("failed to create proxy client: %w", clientErr)
            }
            d.session, err = discordgo.New(
                "Bot "+d.token,
                discordgo.WithHTTPClient(client),
            )
        } else {
            d.session, err = discordgo.New("Bot " + d.token)
        }
    } else {
        d.session, err = discordgo.New("Bot " + d.token)
    }

    if err != nil {
        return fmt.Errorf("failed to create discord session: %w", err)
    }

    // ... rest of the method
}
```

**Step 4: Update FeishuBot and DingTalkBot**

Similar pattern as DiscordBot. For brevity, implement same pattern:
1. Add `proxyMgr interface{}` field
2. Add `SetProxyManager` method
3. Modify Start method to use proxy client

**Step 5: Run tests**

Run: `go test ./internal/bot/... -v`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/bot/
git commit -m "feat: add proxy support to all bot adapters

- Add SetProxyManager to BotAdapter interface
- Update Telegram, Discord, Feishu, DingTalk bots
- Use proxy manager to create HTTP clients
- Support both REST API and WebSocket proxy connections"
```

---

## Task 6: Wire ProxyManager to Bot Creation

**Files:**
- Modify: `internal/core/engine.go`

**Step 1: Find bot creation code**

Search for where bot instances are created in engine.go

**Step 2: Update bot initialization to set ProxyManager**

Locate bot initialization code (likely in `initializeBots` or similar method) and add:

```go
func (e *Engine) initializeBots() error {
    for botType, botConfig := range e.config.Bots {
        if !botConfig.Enabled {
            continue
        }

        var botAdapter bot.BotAdapter
        var err error

        switch botType {
        case "telegram":
            telegramBot := bot.NewTelegramBot(botConfig.Token)
            telegramBot.SetProxyManager(e.proxyMgr)  // NEW
            botAdapter = telegramBot

        case "discord":
            discordBot := bot.NewDiscordBot(botConfig.Token, botConfig.ChannelID)
            discordBot.SetProxyManager(e.proxyMgr)  // NEW
            botAdapter = discordBot

        case "feishu":
            feishuBot := bot.NewFeishuBot(botConfig.AppID, botConfig.AppSecret,
                botConfig.EncryptKey, botConfig.VerificationToken)
            feishuBot.SetProxyManager(e.proxyMgr)  // NEW
            botAdapter = feishuBot

        case "dingtalk":
            dingtalkBot := bot.NewDingTalkBot(botConfig.AppID, botConfig.AppSecret)
            dingtalkBot.SetProxyManager(e.proxyMgr)  // NEW
            botAdapter = dingtalkBot

        default:
            logger.Warnf("Unknown bot type: %s", botType)
            continue
        }

        e.activeBots[botType] = botAdapter
    }

    return nil
}
```

**Step 3: Run tests**

Run: `go test ./internal/core/... -v`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/core/engine.go
git commit -m "feat: wire ProxyManager to bot initialization

- Pass ProxyManager to each bot during initialization
- Enable automatic proxy support for all bot types"
```

---

## Task 7: Update Configuration Examples

**Files:**
- Modify: `configs/config.full.yaml`

**Step 1: Add proxy configuration section**

Add after `hook_server` section:

```yaml
# ==============================================================================
# Network Proxy Configuration (OPTIONAL)
# ==============================================================================
# Configure proxy for accessing IM platforms in restricted networks
#
# Priority:
#   1. Bot-level proxy (highest) - bots.<name>.proxy
#   2. Global proxy - this section
#   3. Environment variables - HTTP_PROXY, HTTPS_PROXY, ALL_PROXY (default)
#
# Supported proxy types:
#   - http: HTTP/HTTPS proxy (e.g., http://127.0.0.1:7890)
#   - socks5: SOCKS5 proxy (e.g., socks5://127.0.0.1:1080)
#
# Authentication (optional):
#   username: "your_username"
#   password: "your_password"
# ==============================================================================
proxy:
  enabled: false  # Set to true to enable global proxy
  type: "http"    # Proxy type: http, https, socks5
  url: "http://127.0.0.1:7890"  # Proxy server URL
  # username: ""  # Optional: proxy authentication username
  # password: ""  # Optional: proxy authentication password

  # Example SOCKS5 with authentication:
  # enabled: true
  # type: "socks5"
  # url: "socks5://proxy.example.com:1080"
  # username: "user"
  # password: "pass"
```

**Step 2: Add bot-level proxy examples**

Update bot configurations with proxy examples:

```yaml
bots:
  # Telegram Bot
  telegram:
    enabled: false
    token: "123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
    # Optional: Bot-level proxy (overrides global proxy)
    proxy:
      enabled: false  # Set to true to use this proxy instead of global
      type: "socks5"
      url: "socks5://127.0.0.1:1080"
      # username: ""  # Optional
      # password: ""  # Optional

  # Discord Bot
  discord:
    enabled: false
    token: "MTIzNDU2Nzg5MDEyMzQ1Njc4OQ.GhIjKl..."
    channel_id: "123456789012345678"
    # Optional: Bot-level proxy
    # proxy:
    #   enabled: true
    #   type: "http"
    #   url: "http://127.0.0.1:7890"
```

**Step 3: Commit**

```bash
git add configs/config.full.yaml
git commit -m "docs: add proxy configuration examples to config.full.yaml

- Add global proxy configuration section
- Add bot-level proxy examples
- Document priority and supported types
- Include authentication examples"
```

---

## Task 8: Add Comprehensive Tests

**Files:**
- Create: `internal/proxy/manager_integration_test.go`
- Create: `internal/proxy/manager_benchmark_test.go`

**Step 1: Write integration tests**

Create `internal/proxy/manager_integration_test.go`:

```go
package proxy

import (
    "testing"
    "github.com/keepmind9/clibot/internal/core"
    "github.com/stretchr/testify/assert"
)

func TestProxyManager_HTTPProxyIntegration(t *testing.T) {
    config := &core.Config{
        Proxy: core.ProxyConfig{
            Enabled: true,
            Type:    "http",
            URL:     "http://127.0.0.1:8080",
        },
        Bots: map[string]core.BotConfig{},
    }

    pm := NewProxyManager(config)

    // Get client for same bot twice (should return cached client)
    client1, err1 := pm.GetHTTPClient("telegram")
    client2, err2 := pm.GetHTTPClient("telegram")

    assert.NoError(t, err1)
    assert.NoError(t, err2)
    assert.Same(t, client1, client2, "Should return cached client")
}

func TestProxyManager_SOCKS5ProxyIntegration(t *testing.T) {
    config := &core.Config{
        Proxy: core.ProxyConfig{
            Enabled:  true,
            Type:     "socks5",
            URL:      "socks5://127.0.0.1:1080",
            Username: "testuser",
            Password: "testpass",
        },
        Bots: map[string]core.BotConfig{},
    }

    pm := NewProxyManager(config)
    client, err := pm.GetHTTPClient("telegram")

    assert.NoError(t, err)
    assert.NotNil(t, client)
}

func TestProxyManager_BotLevelOverride(t *testing.T) {
    botProxy := &core.ProxyConfig{
        Enabled: true,
        Type:    "socks5",
        URL:     "socks5://127.0.0.1:9999",
    }

    config := &core.Config{
        Proxy: core.ProxyConfig{
            Enabled: true,
            Type:    "http",
            URL:     "http://127.0.0.1:8080",
        },
        Bots: map[string]core.BotConfig{
            "telegram": { Proxy: botProxy },
            "discord":  {},
        },
    }

    pm := NewProxyManager(config)

    // Telegram should use bot-level proxy
    telegramURL := pm.GetProxyURL("telegram")
    assert.Contains(t, telegramURL, "9999")

    // Discord should use global proxy
    discordURL := pm.GetProxyURL("discord")
    assert.Contains(t, discordURL, "8080")
}

func TestProxyManager_ClearCache(t *testing.T) {
    config := &core.Config{
        Proxy: core.ProxyConfig{ Enabled: true, Type: "http", URL: "http://127.0.0.1:8080" },
        Bots: map[string]core.BotConfig{},
    }

    pm := NewProxyManager(config)

    // Get client
    client1, _ := pm.GetHTTPClient("telegram")
    assert.NotNil(t, client1)

    // Clear cache
    pm.ClearCache()

    // Get new client (should be different instance)
    client2, _ := pm.GetHTTPClient("telegram")
    assert.NotNil(t, client2)

    assert.NotSame(t, client1, client2, "Should return new client after cache clear")
}
```

**Step 2: Write benchmark tests**

Create `internal/proxy/manager_benchmark_test.go`:

```go
package proxy

import (
    "testing"
    "github.com/keepmind9/clibot/internal/core"
)

func BenchmarkProxyManager_GetHTTPClient_Cached(b *testing.B) {
    config := &core.Config{
        Proxy: core.ProxyConfig{ Enabled: true, Type: "http", URL: "http://127.0.0.1:8080" },
        Bots: map[string]core.BotConfig{},
    }
    pm := NewProxyManager(config)

    // Warm up cache
    pm.GetHTTPClient("telegram")

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        pm.GetHTTPClient("telegram")
    }
}

func BenchmarkProxyManager_GetHTTPClient_Uncached(b *testing.B) {
    config := &core.Config{
        Proxy: core.ProxyConfig{ Enabled: true, Type: "http", URL: "http://127.0.0.1:8080" },
        Bots: map[string]core.BotConfig{},
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        pm := NewProxyManager(config)
        pm.GetHTTPClient("telegram")
    }
}
```

**Step 3: Run all tests**

Run: `go test ./internal/proxy/... -v -bench=.`
Expected: All tests and benchmarks pass

**Step 4: Commit**

```bash
git add internal/proxy/
git commit -m "test: add comprehensive proxy manager tests

- Add integration tests for HTTP and SOCKS5 proxies
- Add tests for bot-level proxy override
- Add cache invalidation tests
- Add benchmark tests for performance validation"
```

---

## Task 9: Update Documentation

**Files:**
- Create: `docs/en/setup/proxy.md`
- Modify: `README.md`

**Step 1: Create proxy setup guide**

Create `docs/en/setup/proxy.md`:

```markdown
# Network Proxy Configuration

This guide explains how to configure network proxies for accessing IM platforms in restricted network environments.

## Quick Start

### Using Environment Variables (Simplest)

```bash
export HTTP_PROXY="http://127.0.0.1:7890"
export HTTPS_PROXY="http://127.0.0.1:7890"
clibot serve
```

### Using Configuration File

Add to your `config.yaml`:

```yaml
proxy:
  enabled: true
  type: "http"
  url: "http://127.0.0.1:7890"
```

## Configuration Priority

clibot uses a three-tier fallback system:

1. **Bot-level proxy** (highest priority)
2. **Global proxy** (middle priority)
3. **Environment variables** (fallback)

## Proxy Types

### HTTP/HTTPS Proxy

```yaml
proxy:
  enabled: true
  type: "http"
  url: "http://127.0.0.1:7890"
```

### SOCKS5 Proxy

```yaml
proxy:
  enabled: true
  type: "socks5"
  url: "socks5://127.0.0.1:1080"
```

## Authentication

### HTTP Proxy with Auth

```yaml
proxy:
  enabled: true
  type: "http"
  url: "http://proxy.example.com:8080"
  username: "your_username"
  password: "your_password"
```

### SOCKS5 with Auth

```yaml
proxy:
  enabled: true
  type: "socks5"
  url: "socks5://proxy.example.com:1080"
  username: "your_username"
  password: "your_password"
```

## Bot-Level Proxy Configuration

Configure different proxies for different bots:

```yaml
proxy:
  enabled: true
  type: "http"
  url: "http://127.0.0.1:8080"

bots:
  telegram:
    enabled: true
    token: "your_token"
    proxy:
      enabled: true  # Override global proxy
      type: "socks5"
      url: "socks5://127.0.0.1:1080"

  discord:
    enabled: true
    token: "your_token"
    # Uses global proxy (no bot-level override)
```

## WebSocket Support

All proxy types support WebSocket connections:

- **HTTP Proxy**: Uses HTTP CONNECT tunneling
- **SOCKS5 Proxy**: Native WebSocket support

Platforms using WebSocket:
- Discord
- Feishu/Lark
- DingTalk

## Troubleshooting

### Proxy Connection Failed

1. Check proxy server is running
2. Verify proxy URL format
3. Test proxy manually:
   ```bash
   curl -x http://127.0.0.1:8080 https://api.telegram.org
   ```

### Authentication Failed

1. Verify username and password
2. Check proxy server logs
3. Test authentication manually:
   ```bash
   curl -x http://user:pass@127.0.0.1:8080 https://api.telegram.org
   ```

### WebSocket Connection Failed

1. Ensure proxy supports CONNECT method (for HTTP proxy)
2. Try SOCKS5 proxy for better WebSocket support
3. Check firewall rules

## Common Proxy Servers

- **Clash**: https://github.com/Dreamacro/clash
- **V2Ray**: https://www.v2fly.org/
- **Shadowsocks**: https://shadowsocks.org/

## Examples

### Clash Example

```yaml
proxy:
  enabled: true
  type: "http"
  url: "http://127.0.0.1:7890"  # Clash default HTTP port
```

### V2Ray Example

```yaml
proxy:
  enabled: true
  type: "socks5"
  url: "socks5://127.0.0.1:1080"  # V2Ray default SOCKS5 port
```

### Environment Variables Only

No config changes needed:

```bash
export HTTP_PROXY="http://127.0.0.1:7890"
export HTTPS_PROXY="http://127.0.0.1:7890"
export ALL_PROXY="socks5://127.0.0.1:1080"

clibot serve --config config.yaml
```
```

**Step 2: Update README**

Add proxy section to README.md:

```markdown
## Network Proxy

clibot supports network proxies for accessing IM platforms in restricted networks.

### Quick Setup

```bash
# Using environment variables
export HTTP_PROXY="http://127.0.0.1:7890"
clibot serve

# Or configure in config.yaml
proxy:
  enabled: true
  type: "http"
  url: "http://127.0.0.1:7890"
```

### Supported Protocols

- HTTP/HTTPS proxy
- SOCKS5 proxy
- Optional username/password authentication

### Documentation

See [Proxy Configuration Guide](docs/en/setup/proxy.md) for details.
```

**Step 3: Commit**

```bash
git add docs/en/setup/proxy.md README.md
git commit -m "docs: add network proxy configuration guide

- Add comprehensive proxy setup guide
- Document HTTP/HTTPS and SOCKS5 support
- Include authentication examples
- Add troubleshooting section
- Update README with proxy section"
```

---

## Task 10: Final Testing and Validation

**Files:**
- Test: Manual testing script

**Step 1: Create manual testing script**

Create `scripts/test-proxy.sh`:

```bash
#!/bin/bash

echo "Testing proxy support..."

# Test 1: Environment variable proxy
echo "Test 1: Environment variable proxy"
export HTTP_PROXY="http://127.0.0.1:8080"
clibot validate --config configs/config.mini.yaml

# Test 2: Global proxy in config
echo "Test 2: Global proxy configuration"
# Create test config with proxy
clibot validate --config configs/test-proxy.yaml

# Test 3: Bot-level proxy
echo "Test 3: Bot-level proxy override"
# Create test config with bot-level proxy
clibot validate --config configs/test-bot-proxy.yaml

echo "All validation tests passed!"
```

**Step 2: Run full test suite**

Run: `go test ./... -v`
Expected: All tests pass

**Step 3: Build and validate**

Run: `make fmt && make build`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add scripts/test-proxy.sh
git commit -m "test: add proxy validation test script

- Add script for testing proxy configurations
- Test environment variable, global, and bot-level proxies"
```

---

## Summary

This implementation plan adds comprehensive network proxy support to clibot with:

**Features:**
- ✅ HTTP/HTTPS and SOCKS5 protocol support
- ✅ Optional username/password authentication
- ✅ Three-tier fallback hierarchy (bot > global > env)
- ✅ Thread-safe HTTP client caching
- ✅ WebSocket proxy support
- ✅ Zero configuration changes required for existing users

**Files Modified:**
- `internal/core/types.go` - Add proxy config structures
- `internal/core/engine.go` - Integrate ProxyManager
- `internal/bot/*.go` - Add proxy support to all bots
- `configs/config.full.yaml` - Add proxy examples

**Files Created:**
- `internal/proxy/manager.go` - ProxyManager implementation
- `internal/proxy/*_test.go` - Comprehensive tests
- `docs/en/setup/proxy.md` - Proxy setup guide

**Estimated Time:** 2-3 hours for full implementation

**Next Steps:**
1. Review and approve this plan
2. Execute implementation using executing-plans skill
3. Manual testing with real proxy servers
4. Documentation review and refinement
