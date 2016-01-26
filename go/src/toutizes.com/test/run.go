package main

import (
	"flag"
	"fmt"
	"path"
	"runtime"
)

import (
	"toutizes.com/model"
)

var orig_root = flag.String("orig_root", "", "path to the orig files")
var index_root = flag.String("index_root", "", "path to the index files")
var root = flag.String("root", "/tmp/db", "root or the index, mini, etc.")

func database() *model.Database {
	if *orig_root == "" {
		*orig_root = path.Join(*root, "orig")
	}
	if *index_root == "" {
		*index_root = path.Join(*root, "index")
	}
	return model.NewDatabase5(*index_root, *orig_root, path.Join(*root, "mini"),
		                        path.Join(*root, "midi"), path.Join(*root, "montage"))
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
	runtime.GOMAXPROCS(cpu / 2)

	db := database()
	db.Load(true, true, false)
	q := model.ParseQuery("2000", db)
	fmt.Printf("images: %d\n", len(q))
	for img := range q {
		fmt.Printf("%v\n", img.String())
	}
}
