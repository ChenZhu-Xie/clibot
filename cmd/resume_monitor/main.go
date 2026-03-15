package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/keepmind9/clibot/internal/watchdog"
)

func main() {
	sessionName := flag.String("session", "", "tmux session name")
	targetUUID := flag.String("uuid", "", "target Gemini session UUID to resume")
	flag.Parse()

	if *sessionName == "" || *targetUUID == "" {
		fmt.Println("Usage: resume_monitor -session <name> -uuid <uuid>")
		os.Exit(1)
	}

	fmt.Printf("Starting Gemini Resume Monitor\n")
	fmt.Printf("Session: %s, Target UUID: %s\n", *sessionName, *targetUUID)

	// 1. Get CWD from tmux session
	cwd, err := watchdog.GetCWD(*sessionName)
	if err != nil {
		log.Fatalf("Failed to get CWD from session %s: %v", *sessionName, err)
	}
	fmt.Printf("Detected CWD: %s\n", cwd)

	// 2. Resolve session list
	sessions, err := listGeminiSessionsByWorkDir(cwd)
	if err != nil {
		log.Fatalf("Failed to list sessions: %v", err)
	}

	targetIdx := -1
	for i, s := range sessions {
		if strings.HasPrefix(s.ID, *targetUUID) {
			targetIdx = i
			break
		}
	}

	if targetIdx == -1 {
		log.Printf("Available sessions:\n")
		for i, s := range sessions {
			log.Printf("  [%d] %s\n", i, s.ID)
		}
		log.Fatalf("Target UUID %s not found in current work directory history", *targetUUID)
	}

	fmt.Printf("Target session %s found at index %d\n", sessions[targetIdx].ID, targetIdx)

	// 3. Simulation
	fmt.Println("Triggering /resume in tmux...")
	if err := watchdog.SendKeys(*sessionName, "/resume"); err != nil {
		log.Fatalf("Failed to send /resume command: %v", err)
	}

	fmt.Println("Waiting for menu to render (500ms)...")
	time.Sleep(500 * time.Millisecond)
	verifyUI(*sessionName, "Recent chats")

	// 4. Navigate down
	for i := 0; i < targetIdx; i++ {
		fmt.Printf("Sending Down Arrow (%d/%d)...\n", i+1, targetIdx)
		// \x1b[B is Arrow Down
		if err := watchdog.SendKeys(*sessionName, "\x1b[B"); err != nil {
			log.Fatalf("Failed to send Down Arrow: %v", err)
		}
		// Brief pause between arrows for UI stability
		time.Sleep(150 * time.Millisecond)
	}

	// 5. Final Confirmation
	fmt.Println("Target session reached. Pressing Enter (C-m)...")
	if err := watchdog.SendKeys(*sessionName, "C-m"); err != nil {
		log.Fatalf("Failed to send Enter: %v", err)
	}

	fmt.Println("\nMonitor: Simulation complete.")
	fmt.Println("Verification: Please check the tmux session to confirm the correct session was resumed.")
	fmt.Println("UI state after Enter:")
	verifyUI(*sessionName, "")
}

// --- Logic replicated from internal/cli/gemini.go for standalone operation ---

type GeminiSession struct {
	ID      string
	ModTime int64
}

func listGeminiSessionsByWorkDir(workDir string) ([]GeminiSession, error) {
	chatsDir, err := findGeminiChatsDir(workDir)
	if err != nil {
		return nil, err
	}

	matches, err := filepath.Glob(filepath.Join(chatsDir, "session-*.json"))
	if err != nil {
		return nil, err
	}

	sort.Slice(matches, func(i, j int) bool {
		infoI, _ := os.Stat(matches[i])
		infoJ, _ := os.Stat(matches[j])
		return infoI.ModTime().After(infoJ.ModTime())
	})

	var sessions []GeminiSession
	for _, file := range matches {
		info, _ := os.Stat(file)
		id := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(file), "session-"), ".json")
		sessions = append(sessions, GeminiSession{
			ID:      id,
			ModTime: info.ModTime().Unix(),
		})
	}
	return sessions, nil
}

func findGeminiChatsDir(workDir string) (string, error) {
	homeDir, _ := os.UserHomeDir()
	absWorkDir, _ := filepath.Abs(workDir)

	// Hash-based folder is standard for Gemini CLI
	projectHash := fmt.Sprintf("%x", sha256.Sum256([]byte(absWorkDir)))
	hashDir := filepath.Join(homeDir, ".gemini", "tmp", projectHash)
	
	if _, err := os.Stat(filepath.Join(hashDir, "chats")); err == nil {
		return filepath.Join(hashDir, "chats"), nil
	}

	// Fallback scanning
	tmpDir := filepath.Join(homeDir, ".gemini", "tmp")
	entries, _ := os.ReadDir(tmpDir)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		rootFile := filepath.Join(tmpDir, entry.Name(), ".project_root")
		if content, err := os.ReadFile(rootFile); err == nil {
			if strings.TrimSpace(strings.ToLower(string(content))) == strings.TrimSpace(strings.ToLower(absWorkDir)) {
				return filepath.Join(tmpDir, entry.Name(), "chats"), nil
			}
		}
	}
	
	return filepath.Join(hashDir, "chats"), nil
}

func verifyUI(sessionName, expectedText string) {
	pane, err := watchdog.CapturePaneClean(sessionName, 10)
	if err != nil {
		return
	}
	if expectedText != "" {
		if strings.Contains(pane, expectedText) {
			fmt.Printf("[OK] Found: %s\n", expectedText)
		} else {
			fmt.Printf("[WARN] Expected '%s' not found. Screen dump:\n---\n%s\n---\n", expectedText, pane)
		}
	} else {
		// Just dump if no expected text
		fmt.Printf("Screen dump:\n---\n%s\n---\n", pane)
	}
}
