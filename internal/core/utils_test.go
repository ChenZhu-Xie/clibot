package core

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExpandPath tests the expandPath function
func TestExpandPath(t *testing.T) {
	t.Run("expand tilde with path", func(t *testing.T) {
		home, err := os.UserHomeDir()
		require.NoError(t, err)

		result, err := expandPath("~/test/path")
		assert.NoError(t, err)
		expected := filepath.Join(home, "test/path")
		assert.Equal(t, expected, result)
	})

	t.Run("expand tilde only", func(t *testing.T) {
		home, err := os.UserHomeDir()
		require.NoError(t, err)

		result, err := expandPath("~")
		assert.NoError(t, err)
		assert.Equal(t, home, result)
	})

	t.Run("absolute path unchanged", func(t *testing.T) {
		result, err := expandPath("/absolute/path")
		assert.NoError(t, err)
		assert.Equal(t, "/absolute/path", result)
	})

	t.Run("relative path unchanged", func(t *testing.T) {
		result, err := expandPath("relative/path")
		assert.NoError(t, err)
		assert.Equal(t, "relative/path", result)
	})

	t.Run("empty path", func(t *testing.T) {
		result, err := expandPath("")
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

// TestNormalizePath tests the normalizePath function
func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "path with trailing slash",
			input:    "/path/to/dir/",
			expected: "/path/to/dir",
		},
		{
			name:     "path without trailing slash",
			input:    "/path/to/dir",
			expected: "/path/to/dir",
		},
		{
			name:     "path with multiple trailing slashes",
			input:    "/path/to/dir///",
			expected: "/path/to/dir//",
		},
		{
			name:     "root path",
			input:    "/",
			expected: "",
		},
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "relative path with trailing slash",
			input:    "relative/path/",
			expected: "relative/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetUserKey tests the getUserKey function
func TestGetUserKey(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		userID   string
		expected string
	}{
		{
			name:     "discord user",
			platform: "discord",
			userID:   "123456789",
			expected: "discord:123456789",
		},
		{
			name:     "telegram user",
			platform: "telegram",
			userID:   "987654321",
			expected: "telegram:987654321",
		},
		{
			name:     "feishu user",
			platform: "feishu",
			userID:   "abc123",
			expected: "feishu:abc123",
		},
		{
			name:     "empty platform",
			platform: "",
			userID:   "user123",
			expected: ":user123",
		},
		{
			name:     "empty userID",
			platform: "discord",
			userID:   "",
			expected: "discord:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getUserKey(tt.platform, tt.userID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSetServerDefaults tests the setServerDefaults function
func TestSetServerDefaults(t *testing.T) {
	t.Run("sets default port when zero", func(t *testing.T) {
		config := &Config{
			HookServer: HookServerConfig{
				Port: 0,
			},
		}

		setServerDefaults(config)

		assert.Equal(t, DefaultHookPort, config.HookServer.Port)
	})

	t.Run("preserves existing port", func(t *testing.T) {
		config := &Config{
			HookServer: HookServerConfig{
				Port: 9090,
			},
		}

		setServerDefaults(config)

		assert.Equal(t, 9090, config.HookServer.Port)
	})
}

// TestValidateSecuritySettings tests the validateSecuritySettings function
func TestValidateSecuritySettings(t *testing.T) {
	t.Run("whitelist enabled with allowed users", func(t *testing.T) {
		config := &Config{
			Security: SecurityConfig{
				WhitelistEnabled: true,
				AllowedUsers: map[string][]string{
					"discord": {"123456789"},
				},
			},
		}

		err := validateSecuritySettings(config)
		assert.NoError(t, err)
	})

	t.Run("whitelist enabled without allowed users", func(t *testing.T) {
		config := &Config{
			Security: SecurityConfig{
				WhitelistEnabled: true,
				AllowedUsers:     map[string][]string{},
			},
		}

		err := validateSecuritySettings(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty when whitelist is enabled")
	})

	t.Run("whitelist disabled", func(t *testing.T) {
		config := &Config{
			Security: SecurityConfig{
				WhitelistEnabled: false,
				AllowedUsers:     map[string][]string{},
			},
		}

		err := validateSecuritySettings(config)
		assert.NoError(t, err)
	})

	t.Run("whitelist disabled with nil allowed users", func(t *testing.T) {
		config := &Config{
			Security: SecurityConfig{
				WhitelistEnabled: false,
			},
		}

		err := validateSecuritySettings(config)
		assert.NoError(t, err)
	})
}

// TestExpandHome tests the expandHome function
func TestExpandHome(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	t.Run("expand tilde with path", func(t *testing.T) {
		home, err := os.UserHomeDir()
		require.NoError(t, err)

		result, err := expandHome("~/test/path")
		assert.NoError(t, err)
		expected := home + "/test/path"
		assert.Equal(t, expected, result)
	})

	t.Run("path without tilde unchanged", func(t *testing.T) {
		result, err := expandHome("/absolute/path")
		assert.NoError(t, err)
		assert.Equal(t, "/absolute/path", result)
	})

	t.Run("relative path unchanged", func(t *testing.T) {
		result, err := expandHome("relative/path")
		assert.NoError(t, err)
		assert.Equal(t, "relative/path", result)
	})

	t.Run("empty path", func(t *testing.T) {
		result, err := expandHome("")
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("path starting with tilde only returns unchanged", func(t *testing.T) {
		// expandHome only expands paths starting with "~/", not "~" alone
		result, err := expandHome("~")
		assert.NoError(t, err)
		assert.Equal(t, "~", result)
	})
}
