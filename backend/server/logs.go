package server

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// QueryLogEntry represents a single parsed search log line.
type QueryLogEntry struct {
	Timestamp time.Time
	Email     string
	Query     string
	Kind      string
}

// QueryWithTimestamp is the JSON structure returned in /db/user-queries.
type QueryWithTimestamp struct {
	Query     string    `json:"query"`
	Timestamp time.Time `json:"timestamp"`
	Kind      string    `json:"kind"`
}

// QueriesByUser maps a user's full email address to their recent queries.
type QueriesByUser map[string][]QueryWithTimestamp

var queryLogRegex = regexp.MustCompile(`^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}) Query from ([^\s@]+@[^\s@:]+): (.+) \(kind: ([^)]*)\)`)

// LogParser parses query history from aserve.log* files in a directory.
type LogParser struct {
	logDir string
}

// NewLogParser creates a LogParser for logDir.
func NewLogParser(logDir string) *LogParser {
	return &LogParser{logDir: logDir}
}

func (lp *LogParser) parseLogLine(line string) (*QueryLogEntry, bool) {
	matches := queryLogRegex.FindStringSubmatch(strings.TrimSpace(line))
	if len(matches) != 5 {
		return nil, false
	}

	ts, err := time.Parse("2006/01/02 15:04:05", matches[1])
	if err != nil {
		return nil, false
	}
	email := matches[2]
	query := matches[3]
	kind := matches[4]

	if len(query) >= 2 && query[0] == '"' && query[len(query)-1] == '"' {
		query = query[1 : len(query)-1]
	}
	if query == "albums:" || strings.TrimSpace(query) == "" {
		return nil, false
	}

	return &QueryLogEntry{
		Timestamp: ts,
		Email:     email,
		Query:     query,
		Kind:      kind,
	}, true
}

func (lp *LogParser) parseLogFile(filePath string) ([]QueryLogEntry, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []QueryLogEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if entry, ok := lp.parseLogLine(scanner.Text()); ok {
			entries = append(entries, *entry)
		}
	}
	return entries, scanner.Err()
}

// ParseAllLogs walks logDir and parses all aserve.log* files.
func (lp *LogParser) ParseAllLogs() ([]QueryLogEntry, error) {
	var all []QueryLogEntry
	err := filepath.Walk(lp.logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasPrefix(info.Name(), "aserve.log") {
			return nil
		}
		if entries, parseErr := lp.parseLogFile(path); parseErr == nil {
			all = append(all, entries...)
		}
		return nil
	})
	return all, err
}

func removeDuplicatesAndLimit(queries []QueryWithTimestamp, maxCount int) []QueryWithTimestamp {
	if len(queries) == 0 {
		return queries
	}
	res := make([]QueryWithTimestamp, 0, len(queries))
	res = append(res, queries[0])
	for i := 1; i < len(queries); i++ {
		if queries[i].Query != queries[i-1].Query {
			res = append(res, queries[i])
		}
	}
	if len(res) > maxCount {
		res = res[:maxCount]
	}
	return res
}

// GroupQueriesByUser returns per-user query histories sorted newest-first,
// with consecutive duplicates removed and capped at 30 entries per user.
func (lp *LogParser) GroupQueriesByUser() (QueriesByUser, error) {
	entries, err := lp.ParseAllLogs()
	if err != nil {
		return nil, err
	}
	result := make(QueriesByUser)
	for _, entry := range entries {
		result[entry.Email] = append(result[entry.Email], QueryWithTimestamp{
			Query:     entry.Query,
			Timestamp: entry.Timestamp,
			Kind:      entry.Kind,
		})
	}
	for email := range result {
		sort.Slice(result[email], func(i, j int) bool {
			return result[email][i].Timestamp.After(result[email][j].Timestamp)
		})
		result[email] = removeDuplicatesAndLimit(result[email], 30)
	}
	return result, nil
}
