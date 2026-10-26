package main

import (
	"log"
	"os"

	"toutizes.com/go-photwo/backend/catalog"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <image_path>", os.Args[0])
	}
	filepath := os.Args[1]
	h, w, kwds, err := catalog.ExtractImageInfo(filepath)
	if err != nil {
		log.Printf("%s: %v", filepath, err)
		return
	}
	log.Printf("%s: %d x %d", filepath, w, h)
	log.Printf("%s: %d keywords", filepath, len(kwds))
	for _, kwd := range kwds {
		log.Printf("%s", kwd)
	}
}
