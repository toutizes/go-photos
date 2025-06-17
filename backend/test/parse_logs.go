package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

import (
	"toutizes.com/go-photwo/backend/model"
)

var logDir = flag.String("log_dir", "", "path to the log files directory")

func main() {
	flag.Parse()
	
	if *logDir == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -log_dir <path_to_logs>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Note: Processes aserve.log and aserve.log.* files\n")
		os.Exit(1)
	}
	
	// Create log parser
	parser := model.NewLogParser(*logDir)
	
	// Parse all logs and group by user
	queriesByUser, err := parser.GroupQueriesByUser()
	if err != nil {
		log.Fatalf("Error parsing logs: %v", err)
	}
	
	// Display results
	fmt.Printf("Found queries for %d users:\n\n", len(queriesByUser))
	
	for username, queries := range queriesByUser {
		fmt.Printf("User: %s (%d queries)\n", username, len(queries))
		for i, query := range queries {
			if i < 5 { // Show first 5 queries
				fmt.Printf("  - [%s] %s (kind: %s)\n", 
					query.Timestamp.Format("2006-01-02 15:04:05"), 
					query.Query, 
					query.Kind)
			} else if i == 5 {
				fmt.Printf("  ... and %d more queries\n", len(queries)-5)
				break
			}
		}
		fmt.Println()
	}
	
	// Example: Get queries for a specific user
	if len(queriesByUser) > 0 {
		// Get first user as example
		var firstUser string
		for user := range queriesByUser {
			firstUser = user
			break
		}
		
		userQueries, err := parser.GetQueriesForUser(firstUser)
		if err != nil {
			log.Printf("Error getting queries for user %s: %v", firstUser, err)
		} else {
			fmt.Printf("Example - All queries for user '%s':\n", firstUser)
			for _, query := range userQueries {
				fmt.Printf("  [%s] %s (kind: %s)\n", 
					query.Timestamp.Format("2006-01-02 15:04:05"), 
					query.Query, 
					query.Kind)
			}
		}
	}
}