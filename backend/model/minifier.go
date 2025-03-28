package model

import (
	"flag"
	"github.com/nfnt/resize"
	"image"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

var minifier_threads = flag.Int("minifier_threads", 4, "Number of threads for minifying.")

func indexTimes(dir string) (m map[string]time.Time) {
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		os.MkdirAll(dir, 0777)
		return
	}
	m = make(map[string]time.Time, len(fis))
	for _, fi := range fis {
		m[fi.Name()] = fi.ModTime()
	}
	return
}

func mustScale(img *Image, mini_times map[string]time.Time) bool {
	mini_time, ok := mini_times[img.Name()]
	return !ok || mini_time.Before(img.FileTime())
}

func origPath(db *Database, img *Image) string {
	return path.Join(db.orig_root, img.Directory().RelPat(), img.Name())
}

func midiPath(db *Database, img *Image) string {
	return path.Join(db.midi_root, img.Directory().RelPat(), img.Name())
}

func miniPath(db *Database, img *Image) string {
	return path.Join(db.mini_root, img.Directory().RelPat(), img.Name())
}

func centerSquareCrop(img image.Image) image.Rectangle {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var squareSize, xOffset, yOffset int

	if width < height {
		squareSize = width
		yOffset = (height - width) / 2
		xOffset = 0
	} else {
		squareSize = height
		xOffset = (width - height) / 2
		yOffset = 0
	}

	return image.Rect(xOffset, yOffset, xOffset+squareSize, yOffset+squareSize)
}

// Scale the image if it's larger than the max dim. Otherwise
// return the image as was.
func scaleImage(src image.Image, maxDim uint, centerCrop bool) image.Image {
	var bounds image.Rectangle
	if centerCrop {
		bounds = centerSquareCrop(src)
	} else {
		bounds = src.Bounds()
	}
	width := bounds.Dx()
	height := bounds.Dy()

	scale := float32(maxDim) / float32(max(width, height))
	newW := uint(scale * float32(width))
	newH := uint(scale * float32(height))
	if !centerCrop {
		if scale > 1.0 {
			return src
		}
		return resize.Resize(newW, newH, src, resize.Lanczos3)
	}
	croppedImg := image.NewRGBA(bounds)
	draw.Draw(croppedImg, croppedImg.Bounds(), src, bounds.Min, draw.Src)
	if scale > 1.0 {
		return croppedImg
	}
	return resize.Resize(newW, newH, croppedImg, resize.Lanczos3)
}

func doScaleImg(db *Database, img *Image) int {
	orig := origPath(db, img)
	f, err := os.Open(orig)
	if err != nil {
		log.Printf("%v: cannot open", orig)
		return 0
	}
	data, err := jpeg.Decode(f)
	if err != nil {
		log.Printf("%v: cannot decode jpg", orig)
		return 0
	}
	if img.width == 0 || img.height == 0 {
		log.Printf("%v: zero dimension, cannot scale", orig)
		return 0
	}
	midiData := scaleImage(data, 2048, false)
	midiO, err := os.Create(midiPath(db, img))
	if err != nil {
		log.Printf("%v: cannot create", midiPath(db, img))
		return 0
	}
	options := &jpeg.Options{Quality: 90} // Quality ranges from 1 to 100
	jpeg.Encode(midiO, midiData, options)
	miniData := scaleImage(midiData, 360, true)
	miniO, err := os.Create(miniPath(db, img))
	if err != nil {
		log.Printf("%v: cannot create", miniPath(db, img))
		return 0
	}
	jpeg.Encode(miniO, miniData, options)
	return 1
}

func feedImages(db *Database, force bool, img_ch chan<- *Image) int {
	fed := 0
	for _, dir := range db.Directories() {
		rel_pat := dir.RelPat()
		mini_dir := db.FullMiniPath(rel_pat)
		mini_times := indexTimes(mini_dir)
		midi_dir := db.FullMidiPath(rel_pat)
		midi_times := indexTimes(midi_dir)
		for _, img := range dir.Images() {
			if force || mustScale(img, mini_times) || mustScale(img, midi_times) {
				fed += 1
				img_ch <- img
			}
		}
	}
	return fed
}

// rm -rf /tmp/mini/2005-01*
// 8: Minifed 229 images in 21472 ms
// 8:   mogri 229 images in 13872 ms
// 4: Minifed 229 images in 20249 ms
// 4:   mogri 229 images in 14058 ms
// 2: Minifed 229 images in 30756 ms
// Pure go impl:
// 4: Minifed 229 images in 11643 ms
// rm -rf /tmp/mini/2005-0{1,2,3,4}*
// 8: Minifed 927 images in 68639 ms
// 8:   mogri 927 images in 47370
// 6:   mogri 927 images in 43174 ms
// 4: Minifed 927 images in 81521 ms
// 4:   mogri 927 images in 49567 ms

func minifyWorker2(db *Database, img_ch <-chan *Image, wg *sync.WaitGroup, id int) {
	defer wg.Done()
	N := 10
	i := 0
	start_time := time.Now()
	for img := range img_ch {
		i += 1
		doScaleImg(db, img)
		if (i % N) == 0 {
			log.Printf("%d: Resized %d in %d ms\n", id, N,
				time.Since(start_time).Nanoseconds()/1000000)
			start_time = time.Now()
		}
	}
}

func MinifyDatabase(db *Database, force_minis bool, force_midis bool) int {
	force := force_minis || force_midis
	img_ch := make(chan *Image)
	var wg sync.WaitGroup
	for i := 0; i < *minifier_threads; i++ {
		wg.Add(1)
		go minifyWorker2(db, img_ch, &wg, i)
	}
	resized := feedImages(db, force, img_ch)
	close(img_ch)
	wg.Wait()
	return resized
}
