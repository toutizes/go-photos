package main

import (
  "os"
)

import (
	"toutizes.com/go-photwo/backend/model"
)

func main() {
  filepath := os.Args[1]
  ts := model.GetKeywords2(filepath)
  log.Printf("Keywords: %v", ts)
}
