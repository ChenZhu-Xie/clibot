package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsValidSessionName_ValidNames tests valid session names
func TestIsValidSessionName_ValidNames(t *testing.T) {
	validNames := []string{
		"session",
		"my-session",
		"my_session",
		"session123",
		"123session",
		"a",
		"test-session-123",
		"UPPERCASE",
		"MixedCase",
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			assert.True(t, isValidSessionName(name), "%s should be valid", name)
		})
	}
}

// TestIsValidSessionName_InvalidNames tests invalid session names
func TestIsValidSessionName_InvalidNames(t *testing.T) {
	invalidNames := []string{
		"",
		"session with spaces",
		"session/with/slash",
		"session\\with\\backslash",
		"session.with.dots",
		"session@with@special",
		"session:with:colon",
		"session;with;semicolon",
	}

	for _, name := range invalidNames {
		t.Run(name, func(t *testing.T) {
			assert.False(t, isValidSessionName(name), "%s should be invalid", name)
		})
	}
}

// TestIsValidSessionName_LengthLimits tests length restrictions
func TestIsValidSessionName_LengthLimits(t *testing.T) {
	t.Run("exactly 100 characters is valid", func(t *testing.T) {
		name := ""
		for i := 0; i < 100; i++ {
			name += "a"
		}
		assert.True(t, isValidSessionName(name))
	})

	t.Run("101 characters is invalid", func(t *testing.T) {
		name := ""
		for i := 0; i < 101; i++ {
			name += "a"
		}
		assert.False(t, isValidSessionName(name))
	})
}

// TestTruncateString tests the truncateString function
func TestTruncateString(t *testing.T) {
	t.Run("string shorter than max", func(t *testing.T) {
		result := truncateString("hello", 10)
		assert.Equal(t, "hello", result)
	})

	t.Run("string equal to max", func(t *testing.T) {
		result := truncateString("hello", 5)
		assert.Equal(t, "hello", result)
	})

	t.Run("string longer than max", func(t *testing.T) {
		result := truncateString("hello world", 5)
		assert.Equal(t, "hello", result)
	})

	t.Run("empty string", func(t *testing.T) {
		result := truncateString("", 10)
		assert.Equal(t, "", result)
	})

	t.Run("max length zero", func(t *testing.T) {
		result := truncateString("hello", 0)
		assert.Equal(t, "", result)
	})

	t.Run("unicode characters", func(t *testing.T) {
		result := truncateString("你好世界", 6)
		assert.Equal(t, "你好", result)
	})

	t.Run("unicode characters truncated mid-character", func(t *testing.T) {
		result := truncateString("你好", 2)
		assert.Equal(t, 2, len(result))
	})
}

// TestEngine_RegisterCLIAdapter tests RegisterCLIAdapter method
func TestEngine_RegisterCLIAdapter(t *testing.T) {
	config := &Config{
		Sessions: []SessionConfig{
			{Name: "test", CLIType: "claude", WorkDir: "/tmp"},
		},
	}
	engine := NewEngine(config)

	engine.RegisterCLIAdapter("test-cli", nil)

	_, exists := engine.cliAdapters["test-cli"]
	assert.True(t, exists)
}

// TestEngine_RegisterCLIAdapter_MultipleAdapters tests registering multiple CLI adapters
func TestEngine_RegisterCLIAdapter_MultipleAdapters(t *testing.T) {
	config := &Config{
		Sessions: []SessionConfig{
			{Name: "test", CLIType: "claude", WorkDir: "/tmp"},
		},
	}
	engine := NewEngine(config)

	adapters := []string{"cli1", "cli2", "cli3"}
	for _, adapterType := range adapters {
		engine.RegisterCLIAdapter(adapterType, nil)
	}

	for _, adapterType := range adapters {
		_, exists := engine.cliAdapters[adapterType]
		assert.True(t, exists)
	}
}

// TestEngine_RegisterCLIAdapter_Overwrite tests overwriting an existing adapter
func TestEngine_RegisterCLIAdapter_Overwrite(t *testing.T) {
	config := &Config{
		Sessions: []SessionConfig{
			{Name: "test", CLIType: "claude", WorkDir: "/tmp"},
		},
	}
	engine := NewEngine(config)

	engine.RegisterCLIAdapter("test-cli", nil)
	engine.RegisterCLIAdapter("test-cli", nil)

	_, exists := engine.cliAdapters["test-cli"]
	assert.True(t, exists)
}
