package main

import (
  "flag"
  "fmt"
  "log"
  "runtime"
)

import (
  "toutizes.com/model"
)

var orig_root = flag.String("orig_root", "", "path to the orig files")
var root = flag.String("root", "/tmp/db", "root or the index, mini, etc.")

func database() *model.Database {
  if *orig_root == "" {
    log.Fatal("Must pass --orig_root")
  }
  return model.NewDatabase2(*orig_root, *root)
}

func collectQ(q model.Query) []int {
  ids := make([]int, 0)
  for img := range q {
    // fmt.Printf("%v\n", img)
    ids = append(ids, img.Id)
  }
  return ids
}

func main() {
  flag.Parse()

  cpu := runtime.NumCPU()
  runtime.GOMAXPROCS(cpu)

  db := database()
  db.Load(false, false)
  q := model.ParseQuery("2008", db)
  fmt.Printf("images: %d\n", len(q))
  for img := range q {
   fmt.Printf("%v\n", img.String())
  }
}
