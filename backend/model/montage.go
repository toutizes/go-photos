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
	"time"
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
	log.Printf("Montage: %s\n", r.URL.Path)
	splits := strings.Split(r.URL.Path, "/")
	if len(splits) == 0 {
		return
	}
	spec := splits[len(splits)-1]
	geo, ids := montageSpec(spec)
	mont := path.Join(db.MontagePath(), spec+".jpg")
	if servePath(w, mont) {
		log.Printf("Montage served from cache: %s\n", r.URL.Path)
		return
	}
	createMontage2(db, geo, ids, mont)
	servePath(w, mont)
	log.Printf("Montage built and serve: %s\n", r.URL.Path)
}

func createMontage2(db *Database, geo string, ids []int, montPath string) {
	start_time := time.Now()
	var images []image.Image
	var nominal_width, total_width, height int

	// 1. Load and decode JPEG images
	for i, id := range ids {
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

		// The height and width of the first image determines the montage dimensions
		if i == 0 {
			height = img.Bounds().Dy()
			nominal_width = img.Bounds().Dx()
		}
		total_width += nominal_width
	}

	// 2. Create the destination image
	concatenatedImg := image.NewRGBA(image.Rect(0, 0, total_width, height))

	// 3. Draw the images horizontally
	for i, img := range images {
		draw.Draw(concatenatedImg,
			image.Rectangle{image.Point{i * nominal_width, 0}, image.Point{(i + 1) * nominal_width, height}},
			img, image.Point{0, 0}, draw.Over)
	}

	outFile, err := os.Create(montPath)
	if err != nil {
		log.Printf("%v: cannot write montage", montPath)
		return
	}
	defer outFile.Close()

	// Adjust quality as needed
	jpeg.Encode(outFile, concatenatedImg, &jpeg.Options{Quality: 90})
	log.Printf("Montage built %d ms\n",
		time.Since(start_time).Nanoseconds()/1000000)
}
