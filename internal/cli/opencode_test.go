package cli

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewOpenCodeAdapter tests the NewOpenCodeAdapter function
func TestNewOpenCodeAdapter(t *testing.T) {
	t.Run("creates adapter with default config", func(t *testing.T) {
		config := OpenCodeAdapterConfig{}

		adapter, err := NewOpenCodeAdapter(config)

		assert.NoError(t, err)
		assert.NotNil(t, adapter)
	})
}

// TestOpenCodeAdapter_HandleHookData tests the HandleHookData function
func TestOpenCodeAdapter_HandleHookData(t *testing.T) {
	t.Run("valid hook data with cwd", func(t *testing.T) {
		config := OpenCodeAdapterConfig{}
		adapter, err := NewOpenCodeAdapter(config)
		require.NoError(t, err)

		hookData := map[string]interface{}{
			"cwd":        "/home/user/project",
			"session_id": "test-session",
		}
		data, _ := json.Marshal(hookData)

		cwd, prompt, response, err := adapter.HandleHookData(data)

		assert.NoError(t, err)
		assert.Equal(t, "/home/user/project", cwd)
		assert.Equal(t, "", prompt)
		assert.Equal(t, "", response)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		config := OpenCodeAdapterConfig{}
		adapter, err := NewOpenCodeAdapter(config)
		require.NoError(t, err)

		data := []byte("invalid json")

		cwd, prompt, response, err := adapter.HandleHookData(data)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse JSON data")
		assert.Equal(t, "", cwd)
		assert.Equal(t, "", prompt)
		assert.Equal(t, "", response)
	})

	t.Run("missing cwd in hook data", func(t *testing.T) {
		config := OpenCodeAdapterConfig{}
		adapter, err := NewOpenCodeAdapter(config)
		require.NoError(t, err)

		hookData := map[string]interface{}{
			"session_id": "test-session",
		}
		data, _ := json.Marshal(hookData)

		cwd, prompt, response, err := adapter.HandleHookData(data)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing cwd in hook data")
		assert.Equal(t, "", cwd)
		assert.Equal(t, "", prompt)
		assert.Equal(t, "", response)
	})

	t.Run("notification event clears response", func(t *testing.T) {
		config := OpenCodeAdapterConfig{}
		adapter, err := NewOpenCodeAdapter(config)
		require.NoError(t, err)

		hookData := map[string]interface{}{
			"cwd":             "/home/user/project",
			"session_id":      "test-session",
			"hook_event_name": "Notification",
		}
		data, _ := json.Marshal(hookData)

		cwd, prompt, response, err := adapter.HandleHookData(data)

		assert.NoError(t, err)
		assert.Equal(t, "/home/user/project", cwd)
		assert.Equal(t, "", prompt)
		assert.Equal(t, "", response)
	})
}
