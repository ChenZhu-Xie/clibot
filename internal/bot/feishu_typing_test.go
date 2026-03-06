package bot

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFeishuBot_SupportsTypingIndicator tests that Feishu supports typing indicators
func TestFeishuBot_SupportsTypingIndicator(t *testing.T) {
	bot := NewFeishuBot("test-app-id", "test-app-secret")

	assert.True(t, bot.SupportsTypingIndicator(), "Feishu should support typing indicators")
}

// TestFeishuBot_AddTypingIndicator_NoClient tests AddTypingIndicator when client is not initialized
func TestFeishuBot_AddTypingIndicator_NoClient(t *testing.T) {
	bot := &FeishuBot{}

	// Don't call Start, so larkClient is nil
	success := bot.AddTypingIndicator("test-message-id")
	assert.False(t, success, "Should return false when client is nil")
}

// TestFeishuBot_AddTypingIndicator_AlreadyExists tests that duplicate calls return true
func TestFeishuBot_AddTypingIndicator_AlreadyExists(t *testing.T) {
	bot := NewFeishuBot("test-app-id", "test-app-secret")

	messageID := "test-message-id"

	// Manually add to map to simulate existing reaction
	bot.mu.Lock()
	bot.typingReactions[messageID] = "test-reaction-id"
	bot.mu.Unlock()

	// Should return true even without client since reaction already exists
	success := bot.AddTypingIndicator(messageID)
	assert.True(t, success, "Should return true when reaction already exists")
}

// TestFeishuBot_RemoveTypingIndicator_NoReaction tests RemoveTypingIndicator when no reaction exists
func TestFeishuBot_RemoveTypingIndicator_NoReaction(t *testing.T) {
	bot := NewFeishuBot("test-app-id", "test-app-secret")

	// No reaction added, should return nil (no error)
	err := bot.RemoveTypingIndicator("test-message-id")
	assert.NoError(t, err, "Should return nil when no reaction exists")
}

// TestFeishuBot_RemoveTypingIndicator_NoClient tests RemoveTypingIndicator when client is not initialized
func TestFeishuBot_RemoveTypingIndicator_NoClient(t *testing.T) {
	bot := &FeishuBot{}

	// Manually add to map
	bot.typingReactions = map[string]string{
		"test-message-id": "test-reaction-id",
	}

	// Should return error when trying to remove without client
	err := bot.RemoveTypingIndicator("test-message-id")
	assert.Error(t, err, "Should return error when client is nil")
	assert.Contains(t, err.Error(), "not initialized", "Error should mention not initialized")
}

// TestFeishuBot_RemoveTypingIndicator_DeleteFromMap tests that reaction is deleted from map
func TestFeishuBot_RemoveTypingIndicator_DeleteFromMap(t *testing.T) {
	bot := NewFeishuBot("test-app-id", "test-app-secret")

	// Initialize context to avoid panic in RemoveTypingIndicator
	ctx, cancel := bot.ctx, bot.cancel
	if ctx == nil {
		var contextCancel context.CancelFunc
		ctx, contextCancel = context.WithCancel(context.Background())
		bot.ctx, bot.cancel = ctx, contextCancel
		defer contextCancel()
	}
	if cancel != nil {
		defer cancel()
	}

	messageID := "test-message-id"

	// Manually add to map
	bot.mu.Lock()
	bot.typingReactions[messageID] = "test-reaction-id"
	bot.mu.Unlock()

	// Remove without client - should delete from map and return error
	_ = bot.RemoveTypingIndicator(messageID)

	// Verify it was deleted from map
	bot.mu.RLock()
	_, exists := bot.typingReactions[messageID]
	bot.mu.RUnlock()

	assert.False(t, exists, "Reaction should be deleted from map")
}
