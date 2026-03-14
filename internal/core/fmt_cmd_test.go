package core

import (
	"fmt"
	"strings"
	"testing"

	"github.com/keepmind9/clibot/internal/bot"
	"github.com/stretchr/testify/assert"
)

func TestEngine_FmtCmd(t *testing.T) {
	engine := NewEngine(&Config{})
	
	// Register dummy bots that implement LinkFormatter
	engine.RegisterBotAdapter("telegram", &mockTelegramBot{})

	tests := []struct {
		platform string
		cmd      string
		expected string
	}{
		{
			platform: "telegram",
			cmd:      "slist",
			expected: "[slist](tg://msg?text=slist)",
		},
		{
			platform: "telegram",
			cmd:      "suse [name]",
			expected: "[suse [name]](tg://msg?text=suse%20)",
		},
		{
			platform: "discord",
			cmd:      "slist",
			expected: "`slist`",
		},
	}

	for _, tt := range tests {
		msg := bot.BotMessage{Platform: tt.platform}
		result := engine.fmtCmd(msg, tt.cmd)
		assert.Equal(t, tt.expected, result)
	}
}

type mockTelegramBot struct {
	bot.BotAdapter
}

func (m *mockTelegramBot) FormatCommandLink(command string, args ...string) string {
	if command == "suse" && len(args) == 0 {
		return "[suse [name]](tg://msg?text=suse%20)"
	}
	text := command
	if len(args) > 0 {
		text += " " + strings.Join(args, " ")
	}
	return fmt.Sprintf("[%s](tg://msg?text=%s)", command, text)
}

func (m *mockTelegramBot) GetBotUsername() string { return "" }
func (m *mockTelegramBot) Start(callback func(bot.BotMessage)) error { return nil }
func (m *mockTelegramBot) Stop() error { return nil }
func (m *mockTelegramBot) SendMessage(channel, message string) error { return nil }
func (m *mockTelegramBot) FormatSessionLink(sessionName, projectPath string) string { return sessionName }
func (m *mockTelegramBot) FormatLogLink(sessionName, logFile string) string { return logFile }
func (m *mockTelegramBot) SupportsTypingIndicator() bool { return false }
func (m *mockTelegramBot) AddTypingIndicator(messageID string) bool { return false }
func (m *mockTelegramBot) RemoveTypingIndicator(messageID string) error { return nil }
