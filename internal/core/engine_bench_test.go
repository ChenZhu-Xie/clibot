package core

import (
	"strings"
	"testing"
)

// BenchmarkIsSpecialCommand_ExactMatch benchmarks exact match commands (fast path)
func BenchmarkIsSpecialCommand_ExactMatch(b *testing.B) {
	tests := []string{
		"help",
		"status",
		"slist",
		"whoami",
		"snew",
		"sdel",
		"suse",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			isSpecialCommand(test)
		}
	}
}

// BenchmarkIsSpecialCommand_NormalInput benchmarks normal input (not a command)
func BenchmarkIsSpecialCommand_NormalInput(b *testing.B) {
	tests := []string{
		"help me write this code",
		"what is the status",
		"can you help me",
		"write a function to parse this",
		"analyze this code please",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			isSpecialCommand(test)
		}
	}
}

// BenchmarkIsSpecialCommand_UnicodeInput benchmarks non-ASCII Unicode input
func BenchmarkIsSpecialCommand_UnicodeInput(b *testing.B) {
	tests := []string{
		"帮我写代码",
		"请分析这个函数",
		"解释一下这段代码",
		"优化这个算法",
		"帮我重构代码",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			isSpecialCommand(test)
		}
	}
}

// BenchmarkIsSpecialCommand_LongInput benchmarks very long input
func BenchmarkIsSpecialCommand_LongInput(b *testing.B) {
	// Simulate long natural language input
	longInputs := make([]string, 10)
	for i := range longInputs {
		longInputs[i] = strings.Repeat("help me understand and analyze this complex piece of code ", 10)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range longInputs {
			isSpecialCommand(input)
		}
	}
}

// BenchmarkIsSpecialCommand_Mixed benchmarks mixed workload
func BenchmarkIsSpecialCommand_Mixed(b *testing.B) {
	tests := []string{
		"help",               // 20% - exact match
		"status",             // 20% - exact match
		"help me write code", // 25% - normal input
		"帮我优化这段代码",           // 25% - Unicode input
		"slist",              // 10% - exact match
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			isSpecialCommand(test)
		}
	}
}

// BenchmarkMapLookup compares pure map lookup performance
func BenchmarkMapLookup(b *testing.B) {
	m := map[string]struct{}{
		"help":   {},
		"status": {},
		"slist":  {},
		"whoami": {},
	}

	keys := []string{"help", "status", "slist", "whoami"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, key := range keys {
			_, _ = m[key]
		}
	}
}

// BenchmarkOldApproach benchmarks the old Fields-first approach (for comparison)
func BenchmarkOldApproach(b *testing.B) {
	specialCommands := map[string]struct{}{
		"help":   {},
		"status": {},
		"slist":  {},
		"whoami": {},
	}

	tests := []string{
		"help",
		"help me write code",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range tests {
			// Old approach: always use Fields
			parts := strings.Fields(input)
			if len(parts) > 0 {
				if _, exists := specialCommands[parts[0]]; exists {
					// Command found
				}
			}
		}
	}
}

// BenchmarkNewApproach benchmarks the new exact-match-first approach
func BenchmarkNewApproach(b *testing.B) {
	specialCommands := map[string]struct{}{
		"help":     {},
		"status":   {},
		"sessions": {},
		"whoami":   {},
	}

	tests := []string{
		"help",
		"help me write code",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range tests {
			// New approach: exact match first
			if _, exists := specialCommands[input]; exists {
				// Fast path hit
				continue
			}
		}
	}
}
