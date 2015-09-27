package main

import (
  "encoding/json"
  "os"
  "runtime"
)

import (
  "toutizes.com/model"
)

func collectQ(q model.Query) []int {
  ids := make([]int, 0)
  for img := range q {
    // fmt.Printf("%v\n", img)
    ids = append(ids, img.Id)
  }
  return ids
}

func main0() {
  cpu := runtime.NumCPU()
  runtime.GOMAXPROCS(cpu)
  var db *model.Database
  if len(os.Args) == 5 {
    db = model.NewDatabase4(os.Args[1], os.Args[2],
      os.Args[3], os.Args[4])
  } else if len(os.Args) == 3 {
    db = model.NewDatabase2(os.Args[1], os.Args[2])
  } else {
    db = model.NewDatabase(os.Args[1])
  }
  err := db.Load(false, false)
  if err != nil {
    println(err.Error())
  }
  ids := collectQ(model.ParseQuery("albums:", db))
  print(len(ids))
}

func main() {
  cpu := runtime.NumCPU()
  runtime.GOMAXPROCS(cpu)

  db_root := os.Args[1]
//  web_root := os.Args[2]

  db := model.NewDatabase(db_root)
  db.Load(false, false)

  enc := json.NewEncoder(os.Stdout)

  q := model.ParseQuery("matthieu Paris", db)
  for img := range q {
    jimg := model.JsonImage{}
    img.Json(&jimg)
    enc.Encode(&jimg)
  }
}
