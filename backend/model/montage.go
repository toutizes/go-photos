package model

import (
	"image"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

func montageSpec(spec string) (geometry string, ids []int) {
	splits := strings.Split(spec, "-")
	if len(splits) < 2 {
		return "", nil
	}
	ids = make([]int, len(splits)-1)
	for i, s := range splits[1:] {
		v, err := strconv.Atoi(s)
		if err != nil {
			return "", nil
		}
		ids[i] = v
	}
	geometry = splits[0]
	return geometry, ids
}

func servePath(w http.ResponseWriter, path string) bool {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return false
	} else {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(bytes)
		return true
	}
}

func HandleMontage2(w http.ResponseWriter, r *http.Request, db *Database) {
	splits := strings.Split(r.URL.Path, "/")
	if len(splits) == 0 {
		return
	}
	spec := splits[len(splits)-1]
	geo, ids := montageSpec(spec)
	mont := path.Join(db.MontagePath(), spec+".jpg")
	if servePath(w, mont) {
		return
	}
	createMontage2(db, geo, ids, mont)
	servePath(w, mont)
}

func createMontage2(db *Database, geo string, ids []int, montPath string) {
	var images []image.Image
	var width, height int

	// 1. Load and decode JPEG images
	for _, id := range ids {
		var path = db.MiniPath(id)
		file, err := os.Open(path)
		if err != nil {
			log.Printf("%v: cannot open", path)
			return
		}
		defer file.Close()

		img, err := jpeg.Decode(file)
		if err != nil {
			log.Printf("%v: cannot decode jpeg", path)
			return
		}

		images = append(images, img)

		// Assuming all images are square and of the same size
		if height == 0 {
			height = img.Bounds().Dy()
		}
		width += img.Bounds().Dx()
	}

	// 2. Create the destination image
	concatenatedImg := image.NewRGBA(image.Rect(0, 0, width, height))

	// 3. Draw the images horizontally
	xOffset := 0
	for _, img := range images {
		draw.Draw(concatenatedImg, img.Bounds().Add(image.Point{xOffset, 0}), img, img.Bounds().Min, draw.Over)
		xOffset += img.Bounds().Dx()
	}

	outFile, err := os.Create(montPath)
	if err != nil {
		log.Printf("%v: cannot write montage", montPath)
		return
	}
	defer outFile.Close()

	// Adjust quality as needed
	jpeg.Encode(outFile, concatenatedImg, &jpeg.Options{Quality: 90})
}
