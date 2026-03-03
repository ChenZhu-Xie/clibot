package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewBaseAdapter tests NewBaseAdapter creation
func TestNewBaseAdapter(t *testing.T) {
	adapter := NewBaseAdapter("test-cli", "test-command", 300)

	assert.Equal(t, "test-cli", adapter.cliName)
	assert.Equal(t, "test-command", adapter.startCmd)
	assert.Equal(t, 300, adapter.inputDelayMs)
}

// TestBaseAdapter_Name tests that adapter name is set correctly
func TestBaseAdapter_Name(t *testing.T) {
	adapter := NewBaseAdapter("mycli", "mycommand", 100)

	// Check that the adapter has the right name
	// The name is stored in cliName field
	assert.Equal(t, "mycli", adapter.cliName)
}

// TestBaseAdapter_StartCmd tests that start command is set correctly
func TestBaseAdapter_StartCmd(t *testing.T) {
	adapter := NewBaseAdapter("test", "claude", 0)

	assert.Equal(t, "claude", adapter.startCmd)
}

// TestBaseAdapter_InputDelayMs tests that input delay is set correctly
func TestBaseAdapter_InputDelayMs(t *testing.T) {
	t.Run("custom delay", func(t *testing.T) {
		adapter := NewBaseAdapter("test", "cmd", 500)
		assert.Equal(t, 500, adapter.inputDelayMs)
	})

	t.Run("zero delay", func(t *testing.T) {
		adapter := NewBaseAdapter("test", "cmd", 0)
		assert.Equal(t, 0, adapter.inputDelayMs)
	})
}
