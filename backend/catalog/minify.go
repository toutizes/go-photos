package catalog

import (
	"image"
	"image/draw"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/nfnt/resize"
)

func indexFileTimes(dir string) map[string]time.Time {
	entries, err := os.ReadDir(dir)
	if err != nil {
		_ = os.MkdirAll(dir, 0777)
		return nil
	}
	m := make(map[string]time.Time, len(entries))
	for _, entry := range entries {
		if info, err := entry.Info(); err == nil {
			m[entry.Name()] = info.ModTime()
		}
	}
	return m
}

func mustScale(img *Photo, times map[string]time.Time) bool {
	t, ok := times[img.Name]
	return !ok || t.Before(img.FileTime)
}

func centerSquareCrop(img image.Image) image.Rectangle {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	var size, xOff, yOff int
	if w < h {
		size = w
		yOff = (h - w) / 2
	} else {
		size = h
		xOff = (w - h) / 2
	}
	return image.Rect(xOff, yOff, xOff+size, yOff+size)
}

func scaleImage(src image.Image, maxDim uint, centerCrop bool) image.Image {
	var bounds image.Rectangle
	if centerCrop {
		bounds = centerSquareCrop(src)
	} else {
		bounds = src.Bounds()
	}
	w := bounds.Dx()
	h := bounds.Dy()

	scale := float32(maxDim) / float32(max(w, h))
	newW := uint(scale * float32(w))
	newH := uint(scale * float32(h))
	if !centerCrop {
		if scale > 1.0 {
			return src
		}
		return resize.Resize(newW, newH, src, resize.Lanczos3)
	}
	cropped := image.NewRGBA(bounds)
	draw.Draw(cropped, cropped.Bounds(), src, bounds.Min, draw.Src)
	if scale > 1.0 {
		return cropped
	}
	return resize.Resize(newW, newH, cropped, resize.Lanczos3)
}

func scalePhoto(c *Catalog, p *Photo) bool {
	orig := c.OrigPath(p)
	f, err := os.Open(orig)
	if err != nil {
		log.Printf("%s: cannot open: %v", orig, err)
		return false
	}
	src, err := jpeg.Decode(f)
	_ = f.Close()
	if err != nil {
		log.Printf("%s: cannot decode jpg: %v", orig, err)
		return false
	}
	if p.Width == 0 || p.Height == 0 {
		log.Printf("%s: zero dimension, cannot scale", orig)
		return false
	}

	opts := &jpeg.Options{Quality: 90}

	midiImg := scaleImage(src, 2048, false)
	midiOutPath := c.MidiPath(p)
	midiFile, err := os.Create(midiOutPath)
	if err != nil {
		log.Printf("%s: cannot create: %v", midiOutPath, err)
		return false
	}
	_ = jpeg.Encode(midiFile, midiImg, opts)
	_ = midiFile.Close()

	miniImg := scaleImage(midiImg, 360, true)
	miniOutPath := c.MiniPath(p)
	miniFile, err := os.Create(miniOutPath)
	if err != nil {
		log.Printf("%s: cannot create: %v", miniOutPath, err)
		return false
	}
	_ = jpeg.Encode(miniFile, miniImg, opts)
	_ = miniFile.Close()
	return true
}

// MinifyCatalog generates missing or stale midi (<=2048px) and mini (360x360 square) JPEGs.
func MinifyCatalog(c *Catalog, threads int, force bool) int {
	if threads <= 0 {
		threads = 2
	}
	imgCh := make(chan *Photo)
	var wg sync.WaitGroup
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			const logBatch = 10
			count := 0
			start := time.Now()
			for p := range imgCh {
				count++
				scalePhoto(c, p)
				if count%logBatch == 0 {
					log.Printf("%d: Resized %d in %d ms", workerID, logBatch, time.Since(start).Milliseconds())
					start = time.Now()
				}
			}
		}(i)
	}

	fed := 0
	for _, album := range c.Albums {
		miniDir := filepath.Join(c.MiniRoot, album.RelPat)
		miniTimes := indexFileTimes(miniDir)
		midiDir := filepath.Join(c.MidiRoot, album.RelPat)
		midiTimes := indexFileTimes(midiDir)
		for _, p := range album.Photos {
			if force || mustScale(p, miniTimes) || mustScale(p, midiTimes) {
				fed++
				imgCh <- p
			}
		}
	}
	close(imgCh)
	wg.Wait()
	return fed
}
