package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEngine_UpdateSessionState_ExistingSession tests updating state of existing session
func TestEngine_UpdateSessionState_ExistingSession(t *testing.T) {
	config := &Config{
		Sessions: []SessionConfig{
			{Name: "test", CLIType: "claude", WorkDir: "/tmp"},
		},
	}
	engine := NewEngine(config)

	// Create a session
	engine.sessions["test"] = &Session{
		Name:  "test",
		State: StateIdle,
	}

	// Update state
	engine.updateSessionState("test", StateProcessing)

	// Verify state was updated
	assert.Equal(t, StateProcessing, engine.sessions["test"].State)
}

// TestEngine_UpdateSessionState_NonExistingSession tests updating state of non-existing session
func TestEngine_UpdateSessionState_NonExistingSession(t *testing.T) {
	config := &Config{
		Sessions: []SessionConfig{
			{Name: "test", CLIType: "claude", WorkDir: "/tmp"},
		},
	}
	engine := NewEngine(config)

	// Try to update non-existing session (should not panic)
	engine.updateSessionState("nonexistent", StateProcessing)

	// No session should be created
	_, exists := engine.sessions["nonexistent"]
	assert.False(t, exists)
}

// TestEngine_GetActiveSession_WithIdleSessions tests getting active session with only idle sessions
func TestEngine_GetActiveSession_WithIdleSessions(t *testing.T) {
	config := &Config{
		Sessions: []SessionConfig{
			{Name: "session1", CLIType: "claude", WorkDir: "/tmp"},
			{Name: "session2", CLIType: "gemini", WorkDir: "/tmp"},
		},
	}
	engine := NewEngine(config)

	// Both sessions are idle
	engine.sessions["session1"] = &Session{Name: "session1", State: StateIdle}
	engine.sessions["session2"] = &Session{Name: "session2", State: StateIdle}

	// Should return one of the idle sessions (random due to map iteration)
	session := engine.GetActiveSession("test-channel")
	assert.NotNil(t, session)
	assert.Contains(t, []string{"session1", "session2"}, session.Name)
	assert.Equal(t, StateIdle, session.State)
}

// TestEngine_GetActiveSession_WithProcessingSession tests getting active session with processing session
func TestEngine_GetActiveSession_WithProcessingSession(t *testing.T) {
	config := &Config{
		Sessions: []SessionConfig{
			{Name: "idle", CLIType: "claude", WorkDir: "/tmp"},
			{Name: "processing", CLIType: "gemini", WorkDir: "/tmp"},
		},
	}
	engine := NewEngine(config)

	engine.sessions["idle"] = &Session{Name: "idle", State: StateIdle}
	engine.sessions["processing"] = &Session{Name: "processing", State: StateProcessing}

	// Should return one of the sessions (map iteration order is random)
	session := engine.GetActiveSession("test-channel")
	assert.NotNil(t, session)
	assert.Contains(t, []string{"idle", "processing"}, session.Name)
}
