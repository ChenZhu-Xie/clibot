package cli

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/keepmind9/clibot/internal/bot"
	"github.com/keepmind9/clibot/internal/logger"
)

// listGeminiSessionsByWorkDir is a shared package-level helper that scans
// ~/.gemini/tmp/{hash}/chats for session-*.json files and returns a list of session info.
func listGeminiSessionsByWorkDir(workDir string) ([]GeminiSession, error) {
	chatsDir, err := findGeminiChatsDir(workDir)
	if err != nil {
		return nil, err
	}

	matches, err := filepath.Glob(filepath.Join(chatsDir, "session-*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to find session files: %w", err)
	}
	if len(matches) == 0 {
		return []GeminiSession{}, nil
	}

	sort.Slice(matches, func(i, j int) bool {
		infoI, _ := os.Stat(matches[i])
		infoJ, _ := os.Stat(matches[j])
		return infoI.ModTime().After(infoJ.ModTime())
	})

	var sessions []GeminiSession
	for _, file := range matches {
		info, err := os.Stat(file)
		if err != nil {
			logger.WithField("file", file).Warn("failed-to-stat-session-file")
			continue
		}

		id := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(file), "session-"), ".json")
		summary := bot.TruncateRuneSafe(geminiSessionSummary(file), 40)

		sessions = append(sessions, GeminiSession{
			ID:      id,
			Summary: summary,
			ModTime: info.ModTime().Unix(),
		})
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].ModTime > sessions[j].ModTime // Newest first
	})

	return sessions, nil
}

// geminiSessionSummary extracts the first user prompt (≤50 chars) from a Gemini CLI
// session JSON file. Gemini stores user messages with content as [{text: "..."}]
// and assistant messages with content as a plain string. We handle both.
func geminiSessionSummary(sessionFile string) string {
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return "(unreadable)"
	}

	var sd struct {
		Title    string `json:"title"`
		Name     string `json:"name"`
		Messages []struct {
			Type    string          `json:"type"`
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(data, &sd); err != nil {
		return "(parse error)"
	}

	// Prefer explicit title or name
	if sd.Title != "" {
		return bot.TruncateRuneSafe(sd.Title, 50)
	}
	if sd.Name != "" {
		return bot.TruncateRuneSafe(sd.Name, 50)
	}

	for _, msg := range sd.Messages {
		if msg.Type != "user" {
			continue
		}
		// Try array form: [{"text": "..."}, ...]
		var parts []struct {
			Text string `json:"text"`
		}
		if json.Unmarshal(msg.Content, &parts) == nil && len(parts) > 0 {
			return bot.TruncateRuneSafe(parts[0].Text, 50)
		}
		// Try plain string form: "..."
		var plain string
		if json.Unmarshal(msg.Content, &plain) == nil {
			return bot.TruncateRuneSafe(plain, 50)
		}
	}
	return "(No messages)"
}

// findGeminiChatsDir finds the Gemini chats directory by searching for the .project_root file
// if the hash-based approach fails.
func findGeminiChatsDir(workDir string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		absWorkDir = workDir
	}

	// 1. Try name-based directory (e.g. .gemini/tmp/clibot)
	projectName := filepath.Base(absWorkDir)
	nameDir := filepath.Join(homeDir, ".gemini", "tmp", projectName)
	if _, err := os.Stat(filepath.Join(nameDir, ".project_root")); err == nil {
		// Verify project root matches
		rootContent, err := os.ReadFile(filepath.Join(nameDir, ".project_root"))
		if err == nil && strings.TrimSpace(strings.ToLower(string(rootContent))) == strings.TrimSpace(strings.ToLower(absWorkDir)) {
			return filepath.Join(nameDir, "chats"), nil
		}
	}

	// 2. Try hash-based directory
	projectHash := computeProjectHash(absWorkDir)
	hashDir := filepath.Join(homeDir, ".gemini", "tmp", projectHash)
	if _, err := os.Stat(filepath.Join(hashDir, "chats")); err == nil {
		return filepath.Join(hashDir, "chats"), nil
	}

	// 3. Scan all directories in .gemini/tmp for .project_root matching absWorkDir
	tmpDir := filepath.Join(homeDir, ".gemini", "tmp")
	entries, err := os.ReadDir(tmpDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			rootFile := filepath.Join(tmpDir, entry.Name(), ".project_root")
			if _, err := os.Stat(rootFile); err == nil {
				rootContent, err := os.ReadFile(rootFile)
				if err == nil && strings.TrimSpace(strings.ToLower(string(rootContent))) == strings.TrimSpace(strings.ToLower(absWorkDir)) {
					return filepath.Join(tmpDir, entry.Name(), "chats"), nil
				}
			}
		}
	}

	// Fallback to the hash-based path
	return filepath.Join(hashDir, "chats"), nil
}

// computeProjectHash computes SHA256 hash of project path
// This is used by Gemini to organize conversation history by project
func computeProjectHash(projectPath string) string {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		logger.WithField("error", err).Warn("failed-to-get-absolute-path-using-raw-path")
		absPath = projectPath
	}

	hash := sha256.Sum256([]byte(absPath))
	return fmt.Sprintf("%x", hash)
}
