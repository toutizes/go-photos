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
		item_time:    baseTime.AddDate(0, 0, -10), // 10 days before baseTime
		Id:           123,
	}
	
	recentImg2 := &Image{
		dir:          recentDir,
		name:         "test2.jpg", 
		keywords:     []string{"vacation", "sunset"},
		sub_keywords: []string{"photography"},
		item_time:    baseTime.AddDate(0, 0, -5), // 5 days before baseTime (more recent)
		Id:           124,
	}
	
	oldImg := &Image{
		dir:          oldDir,
		name:         "old.jpg",
		keywords:     []string{"old", "archive"},
		sub_keywords: []string{},
		item_time:    baseTime.AddDate(0, -3, 0), // 3 months before baseTime
		Id:           125,
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
	
	// Check that "vacation" appears twice (highest count) and has recent images
	found := false
	for _, kw := range keywords {
		if kw.Keyword == "vacation" && kw.Count == 2 {
			found = true
			if len(kw.RecentImages) == 0 {
				t.Error("Expected recent images for vacation keyword")
			}
			// Check that recent images contain the expected image names (now *Image format)
			foundImg1 := false
			foundImg2 := false
			for _, img := range kw.RecentImages {
				if img.Name() == "test1.jpg" {
					foundImg1 = true
				}
				if img.Name() == "test2.jpg" {
					foundImg2 = true
				}
			}
			if !foundImg1 || !foundImg2 {
				t.Error("Expected to find test1.jpg and test2.jpg in recent images")
			}
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

func TestGetRecentActiveKeywordGroups(t *testing.T) {
	// Create a test database
	db := NewDatabase("/tmp/test-db-groups")
	
	// Use fixed dates for consistent testing
	baseTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	
	// Create test directories with different timestamps
	recentDir := &Directory{
		rel_pat:       "2025/2025-06-01",
		last_modified: baseTime.AddDate(0, 0, -15), // 15 days before baseTime
		images:        make([]*Image, 0),
	}
	
	// Create test images with keywords that should be grouped together
	// These images share keywords and should form groups
	groupImg1 := &Image{
		dir:          recentDir,
		name:         "beach1.jpg",
		keywords:     []string{"vacation", "beach", "summer"},
		sub_keywords: []string{},
		item_time:    baseTime.AddDate(0, 0, -5), // 5 days before baseTime
		Id:           201,
	}
	
	groupImg2 := &Image{
		dir:          recentDir,
		name:         "beach2.jpg",
		keywords:     []string{"vacation", "family", "sunset"}, // shares "vacation" with groupImg1
		sub_keywords: []string{},
		item_time:    baseTime.AddDate(0, 0, -3), // 3 days before baseTime (most recent)
		Id:           202,
	}
	
	separateImg := &Image{
		dir:          recentDir,
		name:         "city.jpg",
		keywords:     []string{"city", "urban"},
		sub_keywords: []string{},
		item_time:    baseTime.AddDate(0, 0, -7), // 7 days before baseTime
		Id:           203,
	}
	
	// Add images to directory
	recentDir.images = append(recentDir.images, groupImg1, groupImg2, separateImg)
	
	// Add directory to database
	db.directories = append(db.directories, recentDir)
	
	// Test the function with our fixed base time
	groups := db.GetRecentActiveKeywordGroupsAt(baseTime)
	
	// Verify results
	if len(groups) == 0 {
		t.Fatal("Expected keyword groups, got none")
	}
	
	// Find the group containing "vacation" (should be the most weighted group)
	var vacationGroup *KeywordGroup
	for i := range groups {
		for _, kw := range groups[i].Keywords {
			if kw.Keyword == "vacation" {
				vacationGroup = &groups[i]
				break
			}
		}
		if vacationGroup != nil {
			break
		}
	}
	
	if vacationGroup == nil {
		t.Fatal("Expected to find a group containing 'vacation' keyword")
	}
	
	// Check that vacation group has multiple keywords
	if len(vacationGroup.Keywords) < 2 {
		t.Errorf("Expected vacation group to have multiple keywords, got %d", len(vacationGroup.Keywords))
	}
	
	// Check that keywords are limited to 5 per group (should be less in this test case)
	if len(vacationGroup.Keywords) > 5 {
		t.Errorf("Expected max 5 keywords per group, got %d", len(vacationGroup.Keywords))
	}
	
	// Check that keywords in the group are sorted by weight (descending)
	for i := 1; i < len(vacationGroup.Keywords); i++ {
		if vacationGroup.Keywords[i-1].Weight < vacationGroup.Keywords[i].Weight {
			t.Error("Keywords within group should be sorted by weight in descending order")
		}
	}
	
	// Check total count reflects all unique images
	if vacationGroup.TotalCount < 2 {
		t.Errorf("Expected total count >= 2 for vacation group, got %d", vacationGroup.TotalCount)
	}
	
	// Check that recent images are provided (limited to 4 for display)
	if len(vacationGroup.RecentImages) == 0 {
		t.Error("Expected recent images in vacation group")
	}
	if len(vacationGroup.RecentImages) > 4 {
		t.Errorf("Expected max 4 recent images for display, got %d", len(vacationGroup.RecentImages))
	}
	
	// Verify groups are sorted by total weight (descending)
	for i := 1; i < len(groups); i++ {
		if groups[i-1].TotalWeight < groups[i].TotalWeight {
			t.Error("Groups should be sorted by total weight in descending order")
		}
	}
}

func TestGetRecentActiveKeywordGroupsEmpty(t *testing.T) {
	// Test with empty database
	db := NewDatabase("/tmp/test-db-groups-empty")
	
	groups := db.GetRecentActiveKeywordGroups()
	
	if len(groups) != 0 {
		t.Error("Expected empty result for empty database")
	}
}

func TestGetRecentActiveKeywordGroupsSingleKeywords(t *testing.T) {
	// Test that keywords appearing in only 1 image are filtered out
	db := NewDatabase("/tmp/test-db-groups-single")
	
	// Use fixed dates for consistent testing
	baseTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	
	recentDir := &Directory{
		rel_pat:       "2025/2025-06-01",
		last_modified: baseTime.AddDate(0, 0, -15),
		images:        make([]*Image, 0),
	}
	
	// Create images with keywords that appear only once (should be filtered)
	singleImg1 := &Image{
		dir:          recentDir,
		name:         "single1.jpg",
		keywords:     []string{"unique1"},
		sub_keywords: []string{},
		item_time:    baseTime.AddDate(0, 0, -5),
		Id:           301,
	}
	
	singleImg2 := &Image{
		dir:          recentDir,
		name:         "single2.jpg",
		keywords:     []string{"unique2"},
		sub_keywords: []string{},
		item_time:    baseTime.AddDate(0, 0, -3),
		Id:           302,
	}
	
	recentDir.images = append(recentDir.images, singleImg1, singleImg2)
	db.directories = append(db.directories, recentDir)
	
	groups := db.GetRecentActiveKeywordGroupsAt(baseTime)
	
	// Should be empty because all keywords appear in only 1 image
	if len(groups) != 0 {
		t.Errorf("Expected no groups for single-occurrence keywords, got %d", len(groups))
	}
}

func TestGetRecentActiveKeywordGroupsDuplicateKeyword(t *testing.T) {
	// Test that the same keyword can appear in multiple groups if it's associated with different primary images
	db := NewDatabase("/tmp/test-db-groups-duplicate")
	
	// Use fixed dates for consistent testing
	baseTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	
	recentDir := &Directory{
		rel_pat:       "2025/2025-06-01",
		last_modified: baseTime.AddDate(0, 0, -15),
		images:        make([]*Image, 0),
	}
	
	// First group: "nature" keyword with landscape images
	landscapeImg1 := &Image{
		dir:          recentDir,
		name:         "mountain1.jpg",
		keywords:     []string{"nature", "mountain", "landscape"},
		sub_keywords: []string{},
		item_time:    baseTime.AddDate(0, 0, -8), // 8 days before baseTime
		Id:           401,
	}
	
	landscapeImg2 := &Image{
		dir:          recentDir,
		name:         "mountain2.jpg",
		keywords:     []string{"nature", "hiking", "outdoor"},
		sub_keywords: []string{},
		item_time:    baseTime.AddDate(0, 0, -6), // 6 days before baseTime
		Id:           402,
	}
	
	// Second group: "nature" keyword with wildlife images (different primary image)
	wildlifeImg1 := &Image{
		dir:          recentDir,
		name:         "bird1.jpg",
		keywords:     []string{"nature", "bird", "wildlife"},
		sub_keywords: []string{},
		item_time:    baseTime.AddDate(0, 0, -4), // 4 days before baseTime (most recent, will be primary)
		Id:           403,
	}
	
	wildlifeImg2 := &Image{
		dir:          recentDir,
		name:         "bird2.jpg",
		keywords:     []string{"nature", "photography", "birding"},
		sub_keywords: []string{},
		item_time:    baseTime.AddDate(0, 0, -2), // 2 days before baseTime
		Id:           404,
	}
	
	// Add images to directory
	recentDir.images = append(recentDir.images, landscapeImg1, landscapeImg2, wildlifeImg1, wildlifeImg2)
	
	// Add directory to database
	db.directories = append(db.directories, recentDir)
	
	// Test the function
	groups := db.GetRecentActiveKeywordGroupsAt(baseTime)
	
	// Should have 2 groups since they have different primary images
	if len(groups) != 2 {
		t.Fatalf("Expected 2 groups, got %d", len(groups))
	}
	
	// Count how many groups contain the "nature" keyword
	natureGroupCount := 0
	var group1, group2 *KeywordGroup
	
	for i := range groups {
		hasNature := false
		for _, kw := range groups[i].Keywords {
			if kw.Keyword == "nature" {
				hasNature = true
				break
			}
		}
		if hasNature {
			natureGroupCount++
			if group1 == nil {
				group1 = &groups[i]
			} else {
				group2 = &groups[i]
			}
		}
	}
	
	// "nature" should appear in both groups
	if natureGroupCount != 2 {
		t.Errorf("Expected 'nature' keyword to appear in 2 groups, found in %d groups", natureGroupCount)
	}
	
	if group1 == nil || group2 == nil {
		t.Fatal("Failed to find both groups containing 'nature'")
	}
	
	// Verify that the groups have different primary images
	primaryImg1 := group1.RecentImages[0].Id
	primaryImg2 := group2.RecentImages[0].Id
	
	if primaryImg1 == primaryImg2 {
		t.Error("Expected groups to have different primary images")
	}
	
	// Verify that each group has the "nature" keyword but with different companions
	group1Keywords := make(map[string]bool)
	group2Keywords := make(map[string]bool)
	
	for _, kw := range group1.Keywords {
		group1Keywords[kw.Keyword] = true
	}
	for _, kw := range group2.Keywords {
		group2Keywords[kw.Keyword] = true
	}
	
	// Both should have "nature"
	if !group1Keywords["nature"] || !group2Keywords["nature"] {
		t.Error("Both groups should contain 'nature' keyword")
	}
	
	// They should have some different keywords
	hasDifferentKeywords := false
	for keyword := range group1Keywords {
		if keyword != "nature" && !group2Keywords[keyword] {
			hasDifferentKeywords = true
			break
		}
	}
	for keyword := range group2Keywords {
		if keyword != "nature" && !group1Keywords[keyword] {
			hasDifferentKeywords = true
			break
		}
	}
	
	if !hasDifferentKeywords {
		t.Error("Expected groups to have some different keywords besides 'nature'")
	}
	
	// Verify that total counts make sense
	if group1.TotalCount < 2 || group2.TotalCount < 2 {
		t.Errorf("Expected each group to have at least 2 images, got %d and %d", 
			group1.TotalCount, group2.TotalCount)
	}
}