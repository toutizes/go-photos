package main

import (
	"flag"
	"fmt"
	"path"
	"runtime"
)

import (
	"toutizes.com/go-photwo/backend/model"
)

var orig_root = flag.String("orig_root", "", "path to the orig files")
var index_root = flag.String("index_root", "", "path to the index files")
var root = flag.String("root", "/tmp/db", "root of the index, mini, etc.")

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

func findSingles(db *model.Database, unique_names []string) {
	for i:= 0; i < len(unique_names); i++ {
		not_names := map[string]bool{}
		for j := 0; j < len(unique_names); j++ {
			if j != i {
				not_names[unique_names[j]] = true
			}
		}
		q := model.AndQuery([]model.Query{
			model.KeywordQuery(db, unique_names[i]), model.NotKeywordQuery(db, not_names)})
		
		fmt.Printf("%v:\n", unique_names[i])
		for img := range q {
			fmt.Printf("  %v\n", db.MidiPath(img.Id))
//			fmt.Printf("  %v\n", img.String())
		}
	}
}


func main() {
	flag.Parse()

	cpu := runtime.NumCPU()
	runtime.GOMAXPROCS(cpu / 2)

	db := database()
	// update_disk, minify, force_reload
	db.Load(false, false, false)
	names := []string{"julien", "coline", "catherine", "matthieu"}
	findSingles(db, names)
}
