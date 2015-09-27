package main

import (
  "fmt"
  "os"
  "runtime"
)

import (
  "toutizes.com/model"
  "toutizes.com/store"
)

func dumpQ(q model.Query) {
  for img := range q {
    fmt.Printf("%d: %v\n", img.Id, img)
  }
}

func collectQ(q model.Query) []int {
  ids := make([]int, 0)
  for img := range q {
    // fmt.Printf("%v\n", img)
    ids = append(ids, img.Id)
  }
  return ids
}

func kw(db *model.Database, kwd string) model.Query {
  return model.KeywordQuery(db, kwd)
}

func and(q ...model.Query) model.Query {
  return model.AndQuery(q)
}

func or(q ...model.Query) model.Query {
  return model.OrQuery(q)
}

func contains(s string, ss []string) bool {
  for _, x := range ss {
    if s == x {
      return true
    }
  }
  return false
}

func main0() {
  cpu := runtime.NumCPU()
  runtime.GOMAXPROCS(cpu)
  db := model.NewDatabase(os.Args[1])
  db.Load(false, false)
  ids := collectQ(model.ParseQuery(":albums", db))
  print(len(ids))
}

func main1() {
  fmt.Printf("%v\n", model.DropAccents("héllô", nil))
}

func main() {
  file := os.Args[1]
  fi, _ := os.Stat(file)
  ts := model.TimeToProto(fi.ModTime())
  name := file
  img := &store.Item{
    Name: &name,
    FileTimestamp: &ts,
  }
  model.LoadImageFile(file, img)
  fmt.Printf("\n")
}

func main3() {
  cpu := runtime.NumCPU()
  runtime.GOMAXPROCS(cpu)

  db_root := os.Args[1]

  db := model.NewDatabase(db_root)
  db.Load(false, false)

  q := model.ParseQuery("2015 degustations montoulieu", db)
  for img := range q {
    fmt.Printf("%s: %s\n", img.Name(), img.ItemTime())
  }
}
