package main

import (
	"fmt"
	"toutizes.com/go-photwo/backend/model"
)

func main() {
	db0 := model.NewDatabase2("/Users/matthieu/projects/test-photos", "/tmp/aserve/db-full.orig", "")
	db0.Load(false, false, false)
	db1 := model.NewDatabase2("/Users/matthieu/projects/test-photos", "/tmp/aserve/db-full", "")
	db1.Load(false, false, false)
	compare_db(db0, db1)
}

func compare_db(db0 *model.Database, db1 *model.Database) {
	var dirs0 = db0.Directories()
	var dirs1 = db1.Directories()
	if len(dirs0) != len(dirs1) {
		panic("Different number of directories")
	}
	for i := range dirs0 {
		if dirs0[i].RelPat() != dirs1[i].RelPat() {
			panic("Different directory names")
		}
		if len(dirs0[i].Images()) != len(dirs1[i].Images()) {
			panic("Different number of images")
		}
		for j := range dirs0[i].Images() {
			img0 := dirs0[i].Images()[j]
			img1 := dirs1[i].Images()[j]
			if img0.Name() != img1.Name() {
				fmt.Printf("Different image names: %v %v", img0.Name(), img1.Name())
			}
			// if img0.FileTime() != img1.FileTime() {
			// 	fmt.Printf("Different times: %v %v", img0.Name(), img1.Name())
			// }
			// if img0.ItemTime() != img1.ItemTime() {
			// 	panic("Different item times")
			// }
			// if img0.RotateDegrees() != img1.RotateDegrees() {
			// 	panic("Different rotate degrees")
			// }
			// if img0.Stereo() != img1.Stereo() {
			// 	panic("Different stereo")
			// }
			if len(img0.Keywords()) != len(img1.Keywords()) {
				fmt.Printf("%v: Different keywords count: %v %v", img0.Name(), len(img0.Keywords()), len(img1.Keywords()))
			}
			for k := range img0.Keywords() {
				if img0.Keywords()[k] != img1.Keywords()[k] {
					fmt.Printf("%v: Different keywords count: %v %v", img0.Name(), img0.Keywords()[k], img1.Keywords()[k])
				}
			}
		}
	}
}
