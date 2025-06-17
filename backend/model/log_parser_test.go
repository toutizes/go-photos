package model

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLogParserParseLogLine(t *testing.T) {
	parser := NewLogParser("/tmp") // Directory doesn't matter for line parsing
	
	tests := []struct {
		name     string
		logLine  string
		expected *QueryLogEntry
		shouldMatch bool
	}{
		{
			name:    "Valid query log line",
			logLine: `2025/06/16 16:02:46 Query from viemetivier@gmail.com: "\"manif no kings\"" (kind: )`,
			expected: &QueryLogEntry{
				Timestamp: time.Date(2025, 6, 16, 16, 2, 46, 0, time.UTC),
				Email:     "viemetivier",
				Query:     `\"manif no kings\"`,
				Kind:      "",
			},
			shouldMatch: true,
		},
		{
			name:    "Query with kind specified",
			logLine: `2025/06/17 09:15:30 Query from john.doe@gmail.com: "vacation photos" (kind: album)`,
			expected: &QueryLogEntry{
				Timestamp: time.Date(2025, 6, 17, 9, 15, 30, 0, time.UTC),
				Email:     "john.doe",
				Query:     "vacation photos",
				Kind:      "album",
			},
			shouldMatch: true,
		},
		{
			name:    "Query with complex search term",
			logLine: `2025/06/17 14:30:22 Query from test.user@gmail.com: "\"beach sunset\" OR mountains" (kind: )`,
			expected: &QueryLogEntry{
				Timestamp: time.Date(2025, 6, 17, 14, 30, 22, 0, time.UTC),
				Email:     "test.user",
				Query:     `\"beach sunset\" OR mountains`,
				Kind:      "",
			},
			shouldMatch: true,
		},
		{
			name:        "Filtered albums query",
			logLine:     `2025/06/17 15:00:00 Query from user@gmail.com: "albums:" (kind: )`,
			shouldMatch: false,
		},
		{
			name:        "Filtered empty query",
			logLine:     `2025/06/17 15:01:00 Query from user@gmail.com: "" (kind: )`,
			shouldMatch: false,
		},
		{
			name:        "Filtered whitespace query",
			logLine:     `2025/06/17 15:02:00 Query from user@gmail.com: "   " (kind: )`,
			shouldMatch: false,
		},
		{
			name:        "Non-query log line",
			logLine:     `2025/06/16 16:02:46 Server started on port 8080`,
			shouldMatch: false,
		},
		{
			name:        "Different log format",
			logLine:     `2025/06/16 16:02:46 Error processing request`,
			shouldMatch: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, matched := parser.parseLogLine(tt.logLine)
			
			if matched != tt.shouldMatch {
				t.Errorf("Expected match=%v, got match=%v", tt.shouldMatch, matched)
				return
			}
			
			if !tt.shouldMatch {
				return // No need to check entry content for non-matching lines
			}
			
			if !entry.Timestamp.Equal(tt.expected.Timestamp) {
				t.Errorf("Expected timestamp=%v, got timestamp=%v", tt.expected.Timestamp, entry.Timestamp)
			}
			
			if entry.Email != tt.expected.Email {
				t.Errorf("Expected email=%s, got email=%s", tt.expected.Email, entry.Email)
			}
			
			if entry.Query != tt.expected.Query {
				t.Errorf("Expected query=%s, got query=%s", tt.expected.Query, entry.Query)
			}
			
			if entry.Kind != tt.expected.Kind {
				t.Errorf("Expected kind=%s, got kind=%s", tt.expected.Kind, entry.Kind)
			}
		})
	}
}

func TestLogParserGroupQueriesByUser(t *testing.T) {
	// Create a temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "log_parser_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create test log files with the expected naming pattern
	testLogContent := `2025/06/16 16:02:46 Query from viemetivier@gmail.com: "\"manif no kings\"" (kind: )
2025/06/16 16:03:15 Server info: Processing request
2025/06/17 09:15:30 Query from john.doe@gmail.com: "vacation photos" (kind: album)
2025/06/17 10:20:45 Query from viemetivier@gmail.com: "sunset beach" (kind: )
2025/06/17 10:21:00 Query from viemetivier@gmail.com: "sunset beach" (kind: )
2025/06/17 14:30:22 Query from test.user@gmail.com: "mountains" (kind: )
2025/06/17 14:35:00 Query from test.user@gmail.com: "albums:" (kind: )
`
	
	// Test current log file
	logFilePath := filepath.Join(tempDir, "aserve.log")
	err = os.WriteFile(logFilePath, []byte(testLogContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write current log file: %v", err)
	}
	
	// Test rotated log file
	rotatedLogContent := `2025/06/15 12:00:00 Query from old.user@gmail.com: "old query" (kind: )
`
	rotatedLogPath := filepath.Join(tempDir, "aserve.log.2025-06-17:10:06:17")
	err = os.WriteFile(rotatedLogPath, []byte(rotatedLogContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write rotated log file: %v", err)
	}
	
	// Create a file that should be ignored (not aserve.log)
	ignoredLogPath := filepath.Join(tempDir, "other.log")
	ignoredLogContent := `2025/06/17 15:00:00 Query from ignored.user@gmail.com: "ignored query" (kind: )
`
	err = os.WriteFile(ignoredLogPath, []byte(ignoredLogContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write ignored log file: %v", err)
	}
	
	// Test the parser
	parser := NewLogParser(tempDir)
	queriesByUser, err := parser.GroupQueriesByUser()
	if err != nil {
		t.Fatalf("Failed to parse logs: %v", err)
	}
	
	// Verify results - queries should be sorted by decreasing timestamp (most recent first)
	// Note: consecutive "sunset beach" queries are deduplicated, keeping the first (most recent) one
	expectedUsers := map[string][]QueryWithTimestamp{
		"viemetivier": {
			{Query: "sunset beach", Timestamp: time.Date(2025, 6, 17, 10, 21, 0, 0, time.UTC), Kind: ""},
			{Query: `\"manif no kings\"`, Timestamp: time.Date(2025, 6, 16, 16, 2, 46, 0, time.UTC), Kind: ""},
		},
		"john.doe": {
			{Query: "vacation photos", Timestamp: time.Date(2025, 6, 17, 9, 15, 30, 0, time.UTC), Kind: "album"},
		},
		"test.user": {
			{Query: "mountains", Timestamp: time.Date(2025, 6, 17, 14, 30, 22, 0, time.UTC), Kind: ""},
		},
		"old.user": {
			{Query: "old query", Timestamp: time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC), Kind: ""},
		},
	}
	
	if len(queriesByUser) != len(expectedUsers) {
		t.Errorf("Expected %d users, got %d users", len(expectedUsers), len(queriesByUser))
	}
	
	for expectedUser, expectedQueries := range expectedUsers {
		actualQueries, exists := queriesByUser[expectedUser]
		if !exists {
			t.Errorf("Expected user %s not found in results", expectedUser)
			continue
		}
		
		if len(actualQueries) != len(expectedQueries) {
			t.Errorf("User %s: expected %d queries, got %d queries", 
				expectedUser, len(expectedQueries), len(actualQueries))
			continue
		}
		
		for i, expectedQuery := range expectedQueries {
			if actualQueries[i].Query != expectedQuery.Query {
				t.Errorf("User %s, query %d: expected query=%s, got query=%s", 
					expectedUser, i, expectedQuery.Query, actualQueries[i].Query)
			}
			if !actualQueries[i].Timestamp.Equal(expectedQuery.Timestamp) {
				t.Errorf("User %s, query %d: expected timestamp=%v, got timestamp=%v", 
					expectedUser, i, expectedQuery.Timestamp, actualQueries[i].Timestamp)
			}
			if actualQueries[i].Kind != expectedQuery.Kind {
				t.Errorf("User %s, query %d: expected kind=%s, got kind=%s", 
					expectedUser, i, expectedQuery.Kind, actualQueries[i].Kind)
			}
		}
	}
}

func TestRemoveDuplicatesAndLimit(t *testing.T) {
	tests := []struct {
		name     string
		input    []QueryWithTimestamp
		maxCount int
		expected []QueryWithTimestamp
	}{
		{
			name:     "Empty input",
			input:    []QueryWithTimestamp{},
			maxCount: 10,
			expected: []QueryWithTimestamp{},
		},
		{
			name: "Remove consecutive duplicates",
			input: []QueryWithTimestamp{
				{Query: "cats", Timestamp: time.Date(2025, 6, 17, 15, 0, 0, 0, time.UTC)},
				{Query: "cats", Timestamp: time.Date(2025, 6, 17, 14, 0, 0, 0, time.UTC)},
				{Query: "dogs", Timestamp: time.Date(2025, 6, 17, 13, 0, 0, 0, time.UTC)},
				{Query: "cats", Timestamp: time.Date(2025, 6, 17, 12, 0, 0, 0, time.UTC)},
			},
			maxCount: 10,
			expected: []QueryWithTimestamp{
				{Query: "cats", Timestamp: time.Date(2025, 6, 17, 15, 0, 0, 0, time.UTC)},
				{Query: "dogs", Timestamp: time.Date(2025, 6, 17, 13, 0, 0, 0, time.UTC)},
				{Query: "cats", Timestamp: time.Date(2025, 6, 17, 12, 0, 0, 0, time.UTC)},
			},
		},
		{
			name: "Limit to maxCount",
			input: []QueryWithTimestamp{
				{Query: "query1", Timestamp: time.Date(2025, 6, 17, 15, 0, 0, 0, time.UTC)},
				{Query: "query2", Timestamp: time.Date(2025, 6, 17, 14, 0, 0, 0, time.UTC)},
				{Query: "query3", Timestamp: time.Date(2025, 6, 17, 13, 0, 0, 0, time.UTC)},
				{Query: "query4", Timestamp: time.Date(2025, 6, 17, 12, 0, 0, 0, time.UTC)},
			},
			maxCount: 2,
			expected: []QueryWithTimestamp{
				{Query: "query1", Timestamp: time.Date(2025, 6, 17, 15, 0, 0, 0, time.UTC)},
				{Query: "query2", Timestamp: time.Date(2025, 6, 17, 14, 0, 0, 0, time.UTC)},
			},
		},
		{
			name: "Both duplicates and limit",
			input: []QueryWithTimestamp{
				{Query: "a", Timestamp: time.Date(2025, 6, 17, 15, 0, 0, 0, time.UTC)},
				{Query: "a", Timestamp: time.Date(2025, 6, 17, 14, 0, 0, 0, time.UTC)},
				{Query: "b", Timestamp: time.Date(2025, 6, 17, 13, 0, 0, 0, time.UTC)},
				{Query: "c", Timestamp: time.Date(2025, 6, 17, 12, 0, 0, 0, time.UTC)},
				{Query: "d", Timestamp: time.Date(2025, 6, 17, 11, 0, 0, 0, time.UTC)},
			},
			maxCount: 3,
			expected: []QueryWithTimestamp{
				{Query: "a", Timestamp: time.Date(2025, 6, 17, 15, 0, 0, 0, time.UTC)},
				{Query: "b", Timestamp: time.Date(2025, 6, 17, 13, 0, 0, 0, time.UTC)},
				{Query: "c", Timestamp: time.Date(2025, 6, 17, 12, 0, 0, 0, time.UTC)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeDuplicatesAndLimit(tt.input, tt.maxCount)
			
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d results, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, expected := range tt.expected {
				if result[i].Query != expected.Query {
					t.Errorf("Result %d: expected query=%s, got query=%s", i, expected.Query, result[i].Query)
				}
				if !result[i].Timestamp.Equal(expected.Timestamp) {
					t.Errorf("Result %d: expected timestamp=%v, got timestamp=%v", i, expected.Timestamp, result[i].Timestamp)
				}
			}
		})
	}
}