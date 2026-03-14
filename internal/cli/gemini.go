package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/keepmind9/clibot/internal/bot"
	"github.com/keepmind9/clibot/internal/logger"
	"github.com/keepmind9/clibot/internal/watchdog"
	"github.com/keepmind9/clibot/pkg/utils"
	"github.com/sirupsen/logrus"
)

// GeminiAdapterConfig configuration for Gemini CLI adapter
type GeminiAdapterConfig struct {
	Env map[string]string // Environment variables to set for the CLI process
}

// GeminiAdapter implements CLIAdapter for Gemini CLI
type GeminiAdapter struct {
	BaseAdapter
	mu                sync.Mutex
	cwdToUsage       map[string]float64
	contextUsageLimit float64
	engine            Engine
}

// NewGeminiAdapter creates a new Gemini CLI adapter
func NewGeminiAdapter(config GeminiAdapterConfig) (*GeminiAdapter, error) {
	return &GeminiAdapter{
		BaseAdapter:       NewBaseAdapter("gemini", "gemini", 200),
		cwdToUsage:       make(map[string]float64),
		contextUsageLimit: 0.25, // Default to 25%
	}, nil
}

// SetEngine sets the engine reference for sending responses
func (g *GeminiAdapter) SetEngine(engine Engine) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.engine = engine
}

// GeminiSession represents a Gemini session info
type GeminiSession struct {
	ID      string
	Summary string
	ModTime int64
}

// ResetSession starts a new session for Gemini CLI
func (g *GeminiAdapter) ResetSession(sessionName string) error {
	logger.WithField("session", sessionName).Info("resetting-gemini-session")
	// Send "gemini --new" command followed by enter
	// Note: We use SendKeys with "enter" keyword which is handled by watchdog
	return watchdog.SendKeys(sessionName, "gemini --new\n", g.inputDelayMs)
}

func (g *GeminiAdapter) ListSessions(sessionName string, formatter LinkFormatter) ([]string, error) {
	// Get current working directory from tmux session
	cwd, err := watchdog.GetCWD(sessionName)
	if err != nil {
		logger.WithField("error", err).Warn("failed-to-get-cwd-for-gemini-session-listing")
		return nil, fmt.Errorf("could not determine current work dir: %w", err)
	}
	
	sessions, err := listGeminiSessionsByWorkDir(cwd)
	if err != nil {
		return nil, err
	}

	var formatted []string
	for _, s := range sessions {
		link := s.ID
		summaryLink := ""
		if formatter != nil {
			link = formatter.FormatSessionLink(s.ID, "")
			if s.Summary != "" {
				summaryLink = formatter.FormatCommandLink(s.Summary, "")
			}
		}
		
		summary := s.Summary
		if summary == "" {
			summary = "No summary available"
			formatted = append(formatted, fmt.Sprintf("%s: `%s`", link, summary))
		} else {
			formatted = append(formatted, fmt.Sprintf("%s: [%s](%s)", link, summary, summaryLink))
		}
	}
	return formatted, nil
}

func (g *GeminiAdapter) GetSessionStats(sessionName string, formatter LinkFormatter) (map[string]interface{}, error) {
	cwd, err := watchdog.GetCWD(sessionName)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})
	stats["work_dir"] = cwd

	// Find the most recent session file to extract ID and title
	lastFile, err := g.lastSessionFile(cwd)
	if err == nil && lastFile != "" {
		id := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(lastFile), "session-"), ".json")
		stats["session_id"] = id
		summary := bot.TruncateRuneSafe(geminiSessionSummary(lastFile), 40)
		safeSummary := strings.ReplaceAll(strings.ReplaceAll(summary, "[", "("), "]", ")")
		
		var link, summaryLink string
		if formatter != nil {
			link = formatter.FormatSessionLink(id, "")
			summaryLink = formatter.FormatCommandLink(safeSummary, "")
		} else {
			link = "**" + id + "**"
			summaryLink = safeSummary
		}
		
		// Format: ID: Summary
		stats["session_title"] = fmt.Sprintf("%s: %s", link, summaryLink)
	}

	// Add usage percentage if available
	g.mu.Lock()
	if perc, ok := g.cwdToUsage[cwd]; ok {
		stats["usage_perc"] = perc
	}
	g.mu.Unlock()

	return stats, nil
}

// SwitchSession switches to a specific Gemini session using the /resume command.
func (g *GeminiAdapter) SwitchSession(sessionName, cliSessionID string) (string, error) {
	logger.WithFields(logrus.Fields{
		"session":     sessionName,
		"cli_session": cliSessionID,
	}).Info("switching-gemini-session-natively")

	cwd, err := watchdog.GetCWD(sessionName)
	if err != nil {
		return "", fmt.Errorf("could not determine cwd to validate session: %w", err)
	}

	fullID, err2 := resolveFullSessionID(cwd, cliSessionID)
	if err2 != nil {
		return "", err2
	}
	cliSessionID = fullID

	if err := g.SendInput(sessionName, fmt.Sprintf("/resume %s\n", cliSessionID)); err != nil {
		return "", err
	}

	return getGeminiSessionContext(cwd, cliSessionID), nil
}

// getGeminiSessionContext extracts the latest interaction to use as preview context
func getGeminiSessionContext(cwd, cliSessionID string) string {
	chatsDir, err := findGeminiChatsDir(cwd)
	if err != nil {
		return "(context unavailable)"
	}
	file := filepath.Join(chatsDir, fmt.Sprintf("session-%s.json", cliSessionID))
	userPrompt, response, err := (&GeminiAdapter{}).extractGeminiResponse(file, cwd)
	if err != nil {
		return "(no previous chat text)"
	}
	return fmt.Sprintf("🗣 **You**: %s\n\n🤖 **Gemini**: %s\n\n*(...)*", utils.TruncateRuneSafe(userPrompt, 150), utils.TruncateRuneSafe(response, 300))
}

// listGeminiSessionsByWorkDir is now in gemini_utils.go

// resolveFullSessionID attempts to find the full session UUID given a prefix or suffix.
// If exactly one session file matches the prefix or includes the pattern in the chats directory,
// it returns the full UUID. Otherwise it returns an error.
func resolveFullSessionID(workDir string, prefix string) (string, error) {
	chatsDir, err := findGeminiChatsDir(workDir)
	if err != nil {
		return prefix, err
	}

	if len(prefix) >= 36 { // Already a UUID size or close, check if file exists
		fullPath := filepath.Join(chatsDir, fmt.Sprintf("session-%s.json", prefix))
		if _, err := os.Stat(fullPath); err == nil {
			return prefix, nil
		}
		return prefix, fmt.Errorf("session ID %s does not exist", prefix)
	}

	// Try searching with wildcards to match prefixes (ssls) or suffixes (status bar)
	pattern := filepath.Join(chatsDir, fmt.Sprintf("session-*%s*.json", prefix))
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return prefix, fmt.Errorf("no session found matching: %s", prefix)
	}

	// If multiple matches, pick the most recent one
	if len(matches) > 1 {
		sort.Slice(matches, func(i, j int) bool {
			infoI, _ := os.Stat(matches[i])
			infoJ, _ := os.Stat(matches[j])
			return infoI.ModTime().After(infoJ.ModTime())
		})
	}

	// Extract the actual full ID from the match
	fullID := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(matches[0]), "session-"), ".json")
	return fullID, nil
}

// geminiSessionSummary is now in gemini_utils.go

// TruncateRuneSafe trims s to at most maxRunes Unicode code points and appends
// "..." if it was shortened. It also strips any invalid UTF-8 sequences so the
// output is always safe to send to Telegram.
// TruncateRuneSafe is now in pkg/utils/strings.go

// HandleHookData handles raw hook data from Gemini CLI
// Expected data format (JSON):
//
//	{"session_id": "...", "cwd": "...", ...}
//
// Gemini stores history in: ~/.gemini/tmp/{project_hash}/chats/session-*.json
// where project_hash = SHA256(project_path)
//
// This returns the cwd as the session identifier, which will be matched against
// the configured session's work_dir in the engine.
//
// Parameter data: raw hook data (JSON bytes)
// Returns: (cwd, lastUserPrompt, response, error)
func (g *GeminiAdapter) HandleHookData(data []byte) (string, string, string, error) {
	// Parse JSON
	var hookData map[string]interface{}
	if err := json.Unmarshal(data, &hookData); err != nil {
		logger.WithField("error", err).Error("failed-to-parse-hook-json-data")
		return "", "", "", fmt.Errorf("failed to parse JSON data: %w", err)
	}

	// Extract cwd (current working directory) - used to match the tmux session
	cwd, ok := hookData["cwd"].(string)
	if !ok {
		logger.Warn("missing-cwd-in-hook-data")
		return "", "", "", fmt.Errorf("missing cwd in hook data")
	}

	// Extract transcript_path if available
	transcriptPath := ""
	if v, ok := hookData["transcript_path"].(string); ok {
		transcriptPath = v
	}

	// Extract hook_event_name to check if this is a notification event
	hookEventName := ""
	if v, ok := hookData["hook_event_name"].(string); ok {
		hookEventName = v
	}

	logger.WithFields(logrus.Fields{
		"cwd":             cwd,
		"transcript_path": transcriptPath,
		"hook_event_name": hookEventName,
	}).Debug("hook-data-parsed")

	var response string
	var lastUserPrompt string
	var err error

	// Only extract response for non-notification events
	// Notification events don't have assistant responses to extract
	if !strings.EqualFold(hookEventName, "Notification") {
		lastUserPrompt, response, err = g.extractGeminiResponse(transcriptPath, cwd)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"transcript_path": transcriptPath,
				"cwd":             cwd,
				"error":           err,
			}).Warn("failed-to-extract-gemini-response")
		}
	} else {
		logger.WithField("hook_event_name", hookEventName).Debug("skipping-response-extraction-for-notification-event")
	}

	logger.WithFields(logrus.Fields{
		"cwd":          cwd,
		"prompt_len":   len(lastUserPrompt),
		"response_len": len(response),
	}).Info("response-extracted-from-gemini-history")

	// CONTEXT MONITORING: Parse context usage percentage
	// Expected format: "... X% context used ..."
	if strings.Contains(response, "context used") {
		re := regexp.MustCompile(`(\d+)%\s+context used`)
		matches := re.FindStringSubmatch(response)
		if len(matches) > 1 {
			perc, _ := strconv.ParseFloat(matches[1], 64)
			g.mu.Lock()
			g.cwdToUsage[cwd] = perc
			
			// Auto-reset if usage > limit (threshold 0.25 = 25%)
			if perc/100.0 >= g.contextUsageLimit {
				if g.engine != nil {
					// We don't have the session name here, but the engine knows it by cwd
					// The engine will perform its own check in SendResponseToSession
					// but we can trigger a pre-emptive warning here if we want to.
					// However, without sessionName we can't notify via SendResponseToSession.
					// We'll rely on the engine's check in SendResponseToSession for the actual reset.
					logger.WithFields(logrus.Fields{
						"cwd":   cwd,
						"usage": perc,
					}).Info("context-usage-limit-exceeded-pre-check")
				}
			}
			g.mu.Unlock()
		}
	}

	return cwd, lastUserPrompt, response, nil
}

// Gemini stores history in: ~/.gemini/tmp/{project_hash}/chats/session-*.json
func (g *GeminiAdapter) lastSessionFile(cwd string) (string, error) {
	// Build path to chats directory
	chatsDir, err := findGeminiChatsDir(cwd)
	if err != nil {
		return "", err
	}

	// Check if directory exists
	if _, err := os.Stat(chatsDir); os.IsNotExist(err) {
		return "", fmt.Errorf("chats directory not found: %s", chatsDir)
	}

	// Find all session-*.json files
	pattern := filepath.Join(chatsDir, "session-*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to find session files: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no session files found in %s", chatsDir)
	}

	// Sort by modification time, get the latest
	sort.Slice(matches, func(i, j int) bool {
		infoI, _ := os.Stat(matches[i])
		infoJ, _ := os.Stat(matches[j])
		return infoI.ModTime().After(infoJ.ModTime())
	})

	latestFile := matches[0]

	logger.WithFields(logrus.Fields{
		"latest_file": latestFile,
		"chats_dir":   chatsDir,
	}).Debug("found-latest-gemini-session-file")
	return latestFile, nil
}

// ExtractLatestInteraction exports the latest Gemini response extraction logic
func (g *GeminiAdapter) ExtractLatestInteraction(transcriptPath string, cwd string) (string, string, error) {
	return g.extractGeminiResponse(transcriptPath, cwd)
}

// extractGeminiResponse extracts the latest Gemini response from history
// JSON structure: {"messages": [{"type": "user", ...}, {"type": "gemini", "content": "...", "thoughts": [...]}, ...]}
func (g *GeminiAdapter) extractGeminiResponse(transcriptPath string, cwd string) (string, string, error) {
	var latestFile string
	if transcriptPath == "" {
		var err error
		latestFile, err = g.lastSessionFile(cwd)
		if err != nil {
			return "", "", err
		}
	} else {
		latestFile = transcriptPath
	}

	// Parse JSON
	data, err := os.ReadFile(latestFile)
	if err != nil {
		return "", "", fmt.Errorf("failed to read session file: %w", err)
	}

	var sessionData struct {
		Messages []struct {
			Type     string                   `json:"type"`
			Content  json.RawMessage          `json:"content"`
			Thoughts []map[string]interface{} `json:"thoughts,omitempty"`
		} `json:"messages"`
	}

	if err := json.Unmarshal(data, &sessionData); err != nil {
		return "", "", fmt.Errorf("failed to parse session JSON: %w", err)
	}

	messages := sessionData.Messages
	if len(messages) == 0 {
		return "", "", fmt.Errorf("no messages in session file")
	}

	// Find last user message index
	lastUserIndex := -1
	for i, msg := range messages {
		if msg.Type == "user" {
			lastUserIndex = i
		}
	}

	if lastUserIndex == -1 {
		return "", "", fmt.Errorf("no user message found in session")
	}

	var userPrompt string
	// Parse user content
	var parts []struct {
		Text string `json:"text"`
	}
	if json.Unmarshal(messages[lastUserIndex].Content, &parts) == nil && len(parts) > 0 {
		userPrompt = strings.TrimSpace(parts[0].Text)
	} else {
		var plain string
		if json.Unmarshal(messages[lastUserIndex].Content, &plain) == nil {
			userPrompt = strings.TrimSpace(plain)
		}
	}

	// Collect all Gemini messages after the last user message
	var contentParts []string
	for i := lastUserIndex + 1; i < len(messages); i++ {
		if messages[i].Type == "gemini" {
			var plain string
			if json.Unmarshal(messages[i].Content, &plain) == nil {
				content := strings.TrimSpace(plain)
				if content != "" {
					contentParts = append(contentParts, content)
				}
			}
		}
	}

	if len(contentParts) == 0 {
		return "", "", fmt.Errorf("no Gemini messages found after last user message")
	}

	// Join all content parts with double newline
	response := strings.Join(contentParts, "\n\n")

	logger.WithFields(logrus.Fields{
		"total_messages":  len(messages),
		"last_user_index": lastUserIndex,
		"gemini_messages": len(contentParts),
		"response_length": len(response),
	}).Info("extracted-gemini-response-from-session-file")

	return userPrompt, response, nil
}

// findGeminiChatsDir and computeProjectHash are now in gemini_utils.go
