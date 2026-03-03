package constants

import "time"

// Message length limits for different platforms
const (
	// MaxDiscordMessageLength is Discord's message character limit
	MaxDiscordMessageLength = 2000
	// MaxTelegramMessageLength is Telegram's message character limit
	MaxTelegramMessageLength = 4096
	// MaxFeishuMessageLength is Feishu's message character limit
	MaxFeishuMessageLength = 20000
	// MaxDingTalkMessageLength is DingTalk's message character limit
	MaxDingTalkMessageLength = 20000
)

// Timeouts and delays
const (
	// DefaultConnectionTimeout is the timeout for establishing connections
	DefaultConnectionTimeout = 2 * time.Second
	// DefaultPollTimeout is the timeout for long polling operations
	DefaultPollTimeout = 60 * time.Second
	// HookNotificationDelay is the delay for hook notification to send
	HookNotificationDelay = 300 * time.Millisecond
	// HookHTTPTimeout is the timeout for hook HTTP requests
	HookHTTPTimeout = 5 * time.Second
)

// Message buffer sizes
const (
	// MessageChannelBufferSize is the buffer size for the message channel
	MessageChannelBufferSize = 100
)

// Secret masking
const (
	// MinSecretLengthForMasking is the minimum secret length to apply masking
	MinSecretLengthForMasking = 10
	// SecretMaskPrefixLength is the length of prefix to show before masking
	SecretMaskPrefixLength = 4
	// SecretMaskSuffixLength is the length of suffix to show after masking
	SecretMaskSuffixLength = 4
)

// Logging defaults
const (
	// DefaultLogMaxSize is the default maximum log file size in MB
	DefaultLogMaxSize = 100
	// DefaultLogMaxAge is the default maximum number of days to retain old logs
	DefaultLogMaxAge = 30
	// HTTPSuccessStatusCode is the standard HTTP success status code
	HTTPSuccessStatusCode = 200
)
