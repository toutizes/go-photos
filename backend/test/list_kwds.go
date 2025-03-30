package main

import (
  "os"
  "log"
)

import (
	"toutizes.com/go-photwo/backend/model"
)

func main() {
  filepath := os.Args[1]
  h, w, kwds, err := model.GetImageInfo2(filepath)
  if err != nil {
    log.Printf("%v: %v", filepath, err.Error())
    return
  }

  log.Printf("%v: %v x %v", filepath, w, h)
  log.Printf("%v: %v", filepath, len(kwds))
  for _, kwd := range kwds {
    log.Printf("%s", kwd)
  }
}
