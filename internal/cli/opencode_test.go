package cli

import (
	"encoding/json"
	"os"
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

// TestGetOpencodeMessageText tests the getOpencodeMessageText function
func TestGetOpencodeMessageText(t *testing.T) {
	t.Run("extracts text from message with single text part", func(t *testing.T) {
		msg := OpenCodeMessageInfo{
			Parts: []OpenCodePart{
				{Type: "text", Text: "Hello, world!"},
			},
		}
		result := getOpencodeMessageText(msg)
		assert.Equal(t, "Hello, world!", result)
	})

	t.Run("extracts and joins multiple text parts", func(t *testing.T) {
		msg := OpenCodeMessageInfo{
			Parts: []OpenCodePart{
				{Type: "text", Text: "First paragraph"},
				{Type: "reasoning", Text: "Should be ignored"},
				{Type: "text", Text: "Second paragraph"},
			},
		}
		result := getOpencodeMessageText(msg)
		assert.Equal(t, "First paragraph\n\nSecond paragraph", result)
	})

	t.Run("returns empty string for message with no text parts", func(t *testing.T) {
		msg := OpenCodeMessageInfo{
			Parts: []OpenCodePart{
				{Type: "reasoning", Text: "Some reasoning"},
				{Type: "tool-invocation", Text: "Tool call"},
			},
		}
		result := getOpencodeMessageText(msg)
		assert.Equal(t, "", result)
	})

	t.Run("returns empty string for message with empty parts", func(t *testing.T) {
		msg := OpenCodeMessageInfo{
			Parts: []OpenCodePart{},
		}
		result := getOpencodeMessageText(msg)
		assert.Equal(t, "", result)
	})
}

// TestExtractLatestInteractionFromFile tests the extractLatestInteractionFromFile function
func TestExtractLatestInteractionFromFile(t *testing.T) {
	t.Run("extracts text from valid OpenCode message file", func(t *testing.T) {
		msgData := `{
			"id": "msg-123",
			"role": "assistant",
			"parts": [
				{"type": "text", "text": "This is the response"}
			],
			"metadata": {
				"sessionID": "session-456",
				"time": {"created": 1234567890}
			}
		}`

		// Create a temporary file
		tmpfile, err := os.CreateTemp("", "opencode-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.Write([]byte(msgData))
		require.NoError(t, err)
		tmpfile.Close()

		prompt, response, err := extractLatestInteractionFromFile(tmpfile.Name())

		assert.NoError(t, err)
		assert.Equal(t, "", prompt) // No user prompt in single message
		assert.Equal(t, "This is the response", response)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, _, err := extractLatestInteractionFromFile("/nonexistent/path/file.json")
		assert.Error(t, err)
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		tmpfile, err := os.CreateTemp("", "opencode-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.Write([]byte("not valid json"))
		require.NoError(t, err)
		tmpfile.Close()

		_, _, err = extractLatestInteractionFromFile(tmpfile.Name())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported opencode file format")
	})
}

// TestGetProjectID tests the getProjectID function
func TestGetProjectID(t *testing.T) {
	t.Run("returns global for non-git directory", func(t *testing.T) {
		// Use /tmp which is typically not a git repo
		projectID, err := getProjectID("/tmp")
		assert.NoError(t, err)
		assert.Equal(t, "global", projectID)
	})
}
