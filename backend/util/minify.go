package main

import (
	"flag"
	"log"
)

import (
	"toutizes.com/go-photwo/backend/model"
)

var orig_root = flag.String("orig_root", "", "path to the original images")
var root = flag.String("root", "", "path to the database index, mini, etc")

func main() {
	flag.Parse()
	if *orig_root == "" {
		log.Fatal("Must pass --orig_root")
	}
	if *root == "" {
		log.Fatal("Must pass --root")
	}
  log.Printf("Loading db")
	db := model.NewDatabase2(*orig_root, *root, "")
	db.Load(false, false, false)
  log.Printf("Minifying")
  model.MinifyDatabase(db, true, false);
}
