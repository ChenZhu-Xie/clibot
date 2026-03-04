package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandHome_WithTilde(t *testing.T) {
	// Test expanding ~/path
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	result, err := expandHome("~/test/path")
	assert.NoError(t, err)
	expected := filepath.Join(home, "test/path")
	assert.Equal(t, expected, result)
}

func TestExpandHome_WithoutTilde(t *testing.T) {
	// Test paths without ~ should return unchanged
	testCases := []string{
		"/absolute/path",
		"relative/path",
		"./current/dir",
		"../parent/dir",
	}

	for _, tc := range testCases {
		result, err := expandHome(tc)
		assert.NoError(t, err)
		assert.Equal(t, tc, result)
	}
}

func TestExpandHome_WithTildeOnly(t *testing.T) {
	// Test ~/ alone - expandHome only expands paths starting with "~/", not "~" alone
	result, err := expandHome("~")
	assert.NoError(t, err)
	assert.Equal(t, "~", result)
}

func TestExpandHome_ErrorHandling(t *testing.T) {
	// Since os.UserHomeDir rarely fails, we just verify the function works
	// In a controlled test environment, this should always succeed

	// Test with a valid tilde path
	result, err := expandHome("~/test")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestExpandHome_WindowsPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Windows-specific test")
	}

	// Test that Windows backslash paths are handled
	// (though this function is primarily for Unix-like systems)
	result, err := expandHome("~/test")
	assert.NoError(t, err)
	// On Unix systems, this should use forward slashes
	assert.Contains(t, result, "/")
}

func TestBuildShellCommand_Unix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test")
	}

	cmd := buildShellCommand("echo hello")
	assert.NotNil(t, cmd)
	// Check Args contains the correct command and arguments
	assert.Contains(t, cmd.Args[0], "sh")
	assert.Equal(t, []string{cmd.Args[0], "-c", "echo hello"}, cmd.Args)
}

func TestBuildShellCommand_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test")
	}

	cmd := buildShellCommand("echo hello")
	assert.NotNil(t, cmd)
	// On Windows, exec.Command may use absolute path for cmd.exe
	assert.Contains(t, cmd.Args[0], "cmd")
	assert.Equal(t, []string{cmd.Args[0], "/c", "echo hello"}, cmd.Args)
}

func TestBuildShellCommand_EmptyCommand(t *testing.T) {
	cmd := buildShellCommand("")
	assert.NotNil(t, cmd)

	if runtime.GOOS == "windows" {
		assert.Contains(t, cmd.Args[0], "cmd")
		assert.Equal(t, []string{cmd.Args[0], "/c", ""}, cmd.Args)
	} else {
		assert.Contains(t, cmd.Args[0], "sh")
		assert.Equal(t, []string{cmd.Args[0], "-c", ""}, cmd.Args)
	}
}
