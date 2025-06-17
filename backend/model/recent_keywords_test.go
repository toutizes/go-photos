package model

import (
	"testing"
	"time"
)

func TestGetRecentActiveKeywords(t *testing.T) {
	// Create a test database
	db := NewDatabase("/tmp/test-db")
	
	// Use fixed dates for consistent testing
	baseTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	
	// Create test directories with different timestamps
	recentDir := &Directory{
		rel_pat:       "2025/2025-06-01",
		last_modified: baseTime.AddDate(0, 0, -15), // 15 days before baseTime
		images:        make([]*Image, 0),
	}
	
	oldDir := &Directory{
		rel_pat:       "2024/2024-01-01", 
		last_modified: baseTime.AddDate(0, -2, 0), // 2 months before baseTime
		images:        make([]*Image, 0),
	}
	
	// Create test images with keywords
	recentImg1 := &Image{
		dir:          recentDir,
		name:         "test1.jpg",
		keywords:     []string{"vacation", "beach", "family"},
		sub_keywords: []string{"summer"},
	}
	
	recentImg2 := &Image{
		dir:          recentDir,
		name:         "test2.jpg", 
		keywords:     []string{"vacation", "sunset"},
		sub_keywords: []string{"photography"},
	}
	
	oldImg := &Image{
		dir:          oldDir,
		name:         "old.jpg",
		keywords:     []string{"old", "archive"},
		sub_keywords: []string{},
	}
	
	// Add images to directories
	recentDir.images = append(recentDir.images, recentImg1, recentImg2)
	oldDir.images = append(oldDir.images, oldImg)
	
	// Add directories to database
	db.directories = append(db.directories, recentDir, oldDir)
	
	// Test the function with our fixed base time
	keywords := db.GetRecentActiveKeywordsAt(baseTime)
	
	// Verify results
	if len(keywords) == 0 {
		t.Fatal("Expected keywords, got none")
	}
	
	// Check that "vacation" appears twice (highest count)
	found := false
	for _, kw := range keywords {
		if kw.Keyword == "vacation" && kw.Count == 2 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'vacation' keyword with count 2")
	}
	
	// Check that old keywords are not included
	for _, kw := range keywords {
		if kw.Keyword == "old" || kw.Keyword == "archive" {
			t.Error("Old keywords should not be included in recent keywords")
		}
	}
	
	// Verify keywords are sorted by count (descending)
	for i := 1; i < len(keywords); i++ {
		if keywords[i-1].Count < keywords[i].Count {
			t.Error("Keywords should be sorted by count in descending order")
		}
	}
}

func TestGetRecentActiveKeywordsEmpty(t *testing.T) {
	// Test with empty database
	db := NewDatabase("/tmp/test-db-empty")
	
	keywords := db.GetRecentActiveKeywords()
	
	if len(keywords) != 0 {
		t.Error("Expected empty result for empty database")
	}
}

func TestGetRecentActiveKeywordsNoRecentDirs(t *testing.T) {
	// Test with only old directories
	db := NewDatabase("/tmp/test-db-old")
	
	// Use fixed dates for consistent testing
	baseTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	
	oldDir := &Directory{
		rel_pat:       "2023/2023-01-01",
		last_modified: baseTime.AddDate(-1, 0, 0), // 1 year before baseTime
		images:        make([]*Image, 0),
	}
	
	oldImg := &Image{
		dir:          oldDir,
		name:         "old.jpg",
		keywords:     []string{"old", "archive"},
		sub_keywords: []string{},
	}
	
	oldDir.images = append(oldDir.images, oldImg)
	db.directories = append(db.directories, oldDir)
	
	keywords := db.GetRecentActiveKeywordsAt(baseTime)
	
	if len(keywords) != 0 {
		t.Error("Expected empty result when no recent directories exist")
	}
}

func TestRecentKeywordsCaching(t *testing.T) {
	// Test that GetRecentActiveKeywords uses cached data when available
	db := NewDatabase("/tmp/test-db-cache")
	
	// Use fixed dates for consistent testing
	baseTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	
	// Create test directory with images
	recentDir := &Directory{
		rel_pat:       "2025/2025-06-01",
		last_modified: baseTime.AddDate(0, 0, -15), // 15 days before baseTime
		images:        make([]*Image, 0),
	}
	
	recentImg := &Image{
		dir:          recentDir,
		name:         "test.jpg",
		keywords:     []string{"cached", "test"},
		sub_keywords: []string{},
	}
	
	recentDir.images = append(recentDir.images, recentImg)
	db.directories = append(db.directories, recentDir)
	
	// Manually set cached data
	db.recentActiveKeywords = []KeywordCount{
		{Keyword: "cached", Count: 1},
		{Keyword: "test", Count: 1},
	}
	
	// Test that GetRecentActiveKeywords returns cached data
	keywords := db.GetRecentActiveKeywords()
	
	if len(keywords) != 2 {
		t.Fatalf("Expected 2 cached keywords, got %d", len(keywords))
	}
	
	foundCached := false
	foundTest := false
	for _, kw := range keywords {
		if kw.Keyword == "cached" && kw.Count == 1 {
			foundCached = true
		}
		if kw.Keyword == "test" && kw.Count == 1 {
			foundTest = true
		}
	}
	
	if !foundCached {
		t.Error("Expected to find cached keyword in results")
	}
	if !foundTest {
		t.Error("Expected to find test keyword in results")
	}
}

func TestRecentKeywordsFallback(t *testing.T) {
	// Test that GetRecentActiveKeywords falls back to computation when cache is empty
	db := NewDatabase("/tmp/test-db-fallback")
	
	// Use fixed dates for consistent testing
	baseTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	
	// Create test directory with images
	recentDir := &Directory{
		rel_pat:       "2025/2025-06-01",
		last_modified: baseTime.AddDate(0, 0, -15), // 15 days before baseTime
		images:        make([]*Image, 0),
	}
	
	recentImg := &Image{
		dir:          recentDir,
		name:         "test.jpg",
		keywords:     []string{"fallback"},
		sub_keywords: []string{},
	}
	
	recentDir.images = append(recentDir.images, recentImg)
	db.directories = append(db.directories, recentDir)
	
	// Don't set cache (recentActiveKeywords will be nil)
	// GetRecentActiveKeywords should fall back to computation
	keywords := db.GetRecentActiveKeywords()
	
	if len(keywords) != 1 {
		t.Fatalf("Expected 1 keyword from fallback computation, got %d", len(keywords))
	}
	
	if keywords[0].Keyword != "fallback" || keywords[0].Count != 1 {
		t.Errorf("Expected fallback keyword with count 1, got %s with count %d", 
			keywords[0].Keyword, keywords[0].Count)
	}
}