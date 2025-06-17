package model

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// QueryLogEntry represents a parsed query log entry
type QueryLogEntry struct {
	Timestamp time.Time
	Email     string
	Query     string
	Kind      string
}

// QueryWithTimestamp represents a query with its timestamp
type QueryWithTimestamp struct {
	Query     string    `json:"query"`
	Timestamp time.Time `json:"timestamp"`
	Kind      string    `json:"kind"`
}

// QuerysByUser maps username (email prefix) to list of queries with timestamps
type QuerysByUser map[string][]QueryWithTimestamp

// LogParser handles parsing of query logs from aserve.log files
type LogParser struct {
	logDir string
	// Regex to match log lines: 2025/06/16 16:02:46 Query from viemetivier@gmail.com: "\"manif no kings\"" (kind: )
	queryRegex *regexp.Regexp
}

// NewLogParser creates a new log parser for the given log directory
func NewLogParser(logDir string) *LogParser {
	// Regex pattern to match the log format with timestamp capture
	pattern := `^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}) Query from ([^@]+)@gmail\.com: (.+) \(kind: ([^)]*)\)`
	regex := regexp.MustCompile(pattern)
	
	return &LogParser{
		logDir:     logDir,
		queryRegex: regex,
	}
}

// parseLogLine parses a single log line and returns a QueryLogEntry if it matches
func (lp *LogParser) parseLogLine(line string) (*QueryLogEntry, bool) {
	matches := lp.queryRegex.FindStringSubmatch(strings.TrimSpace(line))
	if len(matches) != 5 {
		return nil, false
	}
	
	timestampStr := matches[1]
	username := matches[2]
	query := matches[3]
	kind := matches[4]
	
	// Parse timestamp - format: 2025/06/16 16:02:46
	timestamp, err := time.Parse("2006/01/02 15:04:05", timestampStr)
	if err != nil {
		return nil, false
	}
	
	// Clean up the query string by removing outer quotes if present
	if len(query) >= 2 && query[0] == '"' && query[len(query)-1] == '"' {
		query = query[1 : len(query)-1]
	}
	
	// Filter out meaningless queries
	if query == "albums:" || query == "" || strings.TrimSpace(query) == "" {
		return nil, false
	}
	
	return &QueryLogEntry{
		Timestamp: timestamp,
		Email:     username,
		Query:     query,
		Kind:      kind,
	}, true
}

// parseLogFile parses a single log file and returns parsed entries
func (lp *LogParser) parseLogFile(filePath string) ([]QueryLogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []QueryLogEntry
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := scanner.Text()
		if entry, ok := lp.parseLogLine(line); ok {
			entries = append(entries, *entry)
		}
	}
	
	return entries, scanner.Err()
}

// ParseAllLogs parses all aserve.log files in the configured directory
// Matches files like "aserve.log" and "aserve.log.2025-06-17:10:06:17"
func (lp *LogParser) ParseAllLogs() ([]QueryLogEntry, error) {
	var allEntries []QueryLogEntry
	
	err := filepath.Walk(lp.logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and only process aserve.log files (including rotated logs)
		if info.IsDir() {
			return nil
		}
		
		// Match files like "aserve.log" or "aserve.log.2025-06-17:10:06:17"
		filename := info.Name()
		if !strings.HasPrefix(filename, "aserve.log") {
			return nil
		}
		
		entries, parseErr := lp.parseLogFile(path)
		if parseErr != nil {
			// Log the error but continue processing other files
			return nil
		}
		
		allEntries = append(allEntries, entries...)
		return nil
	})
	
	return allEntries, err
}

// GroupQueriesByUser groups queries by username (email prefix) and sorts by decreasing timestamp
func (lp *LogParser) GroupQueriesByUser() (QuerysByUser, error) {
	entries, err := lp.ParseAllLogs()
	if err != nil {
		return nil, err
	}
	
	result := make(QuerysByUser)
	
	for _, entry := range entries {
		username := entry.Email
		if _, exists := result[username]; !exists {
			result[username] = make([]QueryWithTimestamp, 0)
		}
		result[username] = append(result[username], QueryWithTimestamp{
			Query:     entry.Query,
			Timestamp: entry.Timestamp,
			Kind:      entry.Kind,
		})
	}
	
	// Sort each user's queries by decreasing timestamp (most recent first)
	for username := range result {
		sort.Slice(result[username], func(i, j int) bool {
			return result[username][i].Timestamp.After(result[username][j].Timestamp)
		})
		
		// Remove consecutive duplicates and limit to 30 queries
		result[username] = removeDuplicatesAndLimit(result[username], 30)
	}
	
	return result, nil
}

// removeDuplicatesAndLimit removes consecutive duplicate queries and limits to maxCount
func removeDuplicatesAndLimit(queries []QueryWithTimestamp, maxCount int) []QueryWithTimestamp {
	if len(queries) == 0 {
		return queries
	}
	
	result := make([]QueryWithTimestamp, 0, len(queries))
	result = append(result, queries[0]) // Always keep the first query
	
	// Remove consecutive duplicates
	for i := 1; i < len(queries); i++ {
		// Only add if it's different from the previous query
		if queries[i].Query != queries[i-1].Query {
			result = append(result, queries[i])
		}
	}
	
	// Limit to maxCount
	if len(result) > maxCount {
		result = result[:maxCount]
	}
	
	return result
}

// GetQueriesForUser returns all queries for a specific user sorted by decreasing timestamp
func (lp *LogParser) GetQueriesForUser(username string) ([]QueryWithTimestamp, error) {
	allQueries, err := lp.GroupQueriesByUser()
	if err != nil {
		return nil, err
	}
	
	if queries, exists := allQueries[username]; exists {
		return queries, nil
	}
	
	return []QueryWithTimestamp{}, nil
}