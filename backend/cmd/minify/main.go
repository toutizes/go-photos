package main

import (
	"flag"
	"log"

	"toutizes.com/go-photwo/backend/catalog"
)

var (
	origRoot        = flag.String("orig_root", "", "path to the original images")
	root            = flag.String("root", "", "path to the database index, mini, midi")
	minifierThreads = flag.Int("minifier_threads", 2, "Number of threads for minifying.")
	_               = flag.String("bin_root", "/usr/local/bin/", "directory for the binaries (deprecated)")
)

func main() {
	flag.Parse()
	if *origRoot == "" {
		log.Fatal("Must pass --orig_root")
	}
	if *root == "" {
		log.Fatal("Must pass --root")
	}
	log.Printf("Loading db")
	cat, err := catalog.LoadCatalog(catalog.Options{
		OrigRoot:        *origRoot,
		Root:            *root,
		UpdateDisk:      false,
		Minify:          false,
		ForceReload:     false,
		MinifierThreads: *minifierThreads,
	})
	if err != nil {
		log.Fatalf("Error loading catalog: %v", err)
	}
	log.Printf("Minifying")
	catalog.MinifyCatalog(cat, *minifierThreads, true)
}
