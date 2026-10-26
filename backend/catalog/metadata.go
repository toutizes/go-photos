package catalog

import (
	"bufio"
	"bytes"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	iptc "github.com/dsoprea/go-iptc"
	jpegstructure "github.com/dsoprea/go-jpeg-image-structure/v2"
	"github.com/rwcarlsen/goexif/exif"
	"golang.org/x/text/encoding/charmap"
	"google.golang.org/protobuf/proto"
	"toutizes.com/go-photwo/backend/store"
)

const (
	ftDir = iota
	ftImg
	ftVid
	ftIgn
)

var ignoredNames = map[string]bool{
	".minis":    true,
	".comments": true,
	"Data":      true,
	"Albums":    true,
	"Desktop":   true,
	"Originals": true,
	"Thumbs":    true,
}

func isImageName(name string) bool {
	return strings.HasSuffix(name, ".jpg") ||
		strings.HasSuffix(name, ".JPG") ||
		strings.HasSuffix(name, ".jpeg") ||
		strings.HasSuffix(name, ".JPEG") ||
		strings.HasSuffix(name, ".BMP") ||
		strings.HasSuffix(name, ".TIF") ||
		strings.HasSuffix(name, ".gif") ||
		strings.HasSuffix(name, ".GIF")
}

func isVideoName(name string) bool {
	return strings.HasSuffix(name, ".webm") ||
		strings.HasSuffix(name, ".WEBM")
}

func classifyFile(file os.FileInfo) int {
	switch {
	case ignoredNames[file.Name()]:
		return ftIgn
	case file.IsDir():
		return ftDir
	case isImageName(file.Name()):
		return ftImg
	case isVideoName(file.Name()):
		return ftVid
	default:
		return ftIgn
	}
}

func tagTime(ex *exif.Exif, ts exif.FieldName) (time.Time, error) {
	tg, err := ex.Get(ts)
	if err != nil {
		return time.Unix(0, 0), err
	}
	strVal, err := tg.StringVal()
	if err != nil {
		return time.Unix(0, 0), err
	}
	return time.Parse("2006:01:02 15:04:05", strVal)
}

// LoadImageFile reads EXIF capture time, dimensions, and IPTC keywords from file into item.
func LoadImageFile(file string, item *store.Item) error {
	fi, err := os.Open(file)
	if err != nil {
		log.Printf("Error opening %s: %v", file, err)
		return err
	}
	defer fi.Close()

	var imageTime time.Time
	foundTime := false
	if ex, err := exif.Decode(bufio.NewReader(fi)); err == nil {
		if dto, err := tagTime(ex, exif.DateTimeOriginal); err == nil {
			imageTime = dto
			foundTime = true
		} else if dtd, err := tagTime(ex, exif.DateTimeDigitized); err == nil {
			log.Printf("Using DateTimeDigitized for: %s (%s)", file, dtd)
			imageTime = dtd
			foundTime = true
		} else if dt, err := tagTime(ex, exif.DateTime); err == nil {
			log.Printf("Using DateTime for: %s (%s)", file, dt)
			imageTime = dt
			foundTime = true
		}
	}
	if !foundTime {
		log.Printf("Using file time for: %s", file)
		if item.FileTimestamp != nil {
			imageTime = ProtoToTime(*item.FileTimestamp)
		} else {
			imageTime = time.Unix(0, 0)
		}
	}
	its := TimeToProto(imageTime)
	item.ItemTimestamp = &its

	height, width, kwds, err := ExtractImageInfo(file)
	if err != nil {
		log.Printf("Info error for %s: %v", file, err)
	}
	if len(kwds) > 0 {
		item.Keywords = kwds
	}
	if height > 0 && width > 0 {
		item.Image = &store.Image{
			Height: proto.Int32(int32(height)),
			Width:  proto.Int32(int32(width)),
		}
	}
	return nil
}

// ExtractImageInfo decodes image dimensions and extracts IPTC Record 2 Dataset 25 keywords.
// If the image has valid dimensions but no IPTC segment, it returns height, width, nil, nil.
func ExtractImageInfo(path string) (height int, width int, keywords []string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, nil, err
	}

	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, nil, err
	}
	height = config.Height
	width = config.Width

	parser := jpegstructure.NewJpegMediaParser()
	intfc, jerr := parser.ParseBytes(data)
	if jerr != nil {
		return height, width, nil, nil
	}
	sl, ok := intfc.(*jpegstructure.SegmentList)
	if !ok {
		return height, width, nil, nil
	}
	_, segment, ierr := sl.FindIptc()
	if ierr != nil {
		return height, width, nil, nil
	}
	tags, terr := segment.Iptc()
	if terr != nil {
		return height, width, nil, nil
	}

	kwdBytes := tags[iptc.StreamTagKey{RecordNumber: 2, DatasetNumber: 25}]
	keywords = make([]string, 0, len(kwdBytes))
	for _, raw := range kwdBytes {
		var decoded string
		if utf8.Valid(raw) {
			decoded = string(raw)
		} else {
			decoder := charmap.ISO8859_1.NewDecoder()
			if decodedBytes, derr := decoder.Bytes(raw); derr == nil {
				decoded = string(decodedBytes)
			} else {
				log.Printf("%s: Failed to decode ISO-8859-1 keyword bytes: %v", path, derr)
				decoded = string(raw)
			}
		}
		for _, kwd := range strings.Split(decoded, ";") {
			keywords = append(keywords, kwd)
		}
	}
	return height, width, keywords, nil
}

func makeItem(file os.FileInfo) *store.Item {
	ts := TimeToProto(file.ModTime())
	name := file.Name()
	return &store.Item{
		Name:          &name,
		FileTimestamp: &ts,
	}
}

func mergeVideo(oldVid, newVid *store.Item) *store.Item {
	if oldVid != nil {
		newVid.Keywords = oldVid.Keywords
		newVid.ItemTimestamp = oldVid.ItemTimestamp
	}
	return newVid
}

func mergeImage(oldImg, newImg *store.Item) *store.Item {
	if oldImg != nil {
		if len(newImg.Keywords) == 0 {
			for _, kwd := range oldImg.Keywords {
				kwd = strings.ReplaceAll(kwd, "\n", "")
				if len(kwd) > 0 {
					newImg.Keywords = append(newImg.Keywords, kwd)
				}
			}
		}
		if newImg.ItemTimestamp == nil {
			newImg.ItemTimestamp = oldImg.ItemTimestamp
		}
		if oldImg.Image != nil {
			if newImg.Image == nil {
				newImg.Image = &store.Image{}
			}
			if oldImg.Image.Stereo != nil {
				newImg.Image.Stereo = oldImg.Image.Stereo
			}
			if oldImg.Image.RotateDegrees != nil {
				newImg.Image.RotateDegrees = oldImg.Image.RotateDegrees
			}
			if oldImg.Image.Height != nil {
				newImg.Image.Height = oldImg.Image.Height
			}
			if oldImg.Image.Width != nil {
				newImg.Image.Width = oldImg.Image.Width
			}
		}
	}
	return newImg
}

// UpdateDirectory reconciles filesystem entries in origDir with cached store.Directory metadata.
func UpdateDirectory(origDir string, subs []os.FileInfo, forceReload bool, sdir *store.Directory) error {
	oldItems := make(map[string]*store.Item, len(sdir.Items))
	for _, itm := range sdir.Items {
		if itm.Name != nil {
			oldItems[*itm.Name] = itm
		}
	}

	var imgs, vids []*store.Item
	var dirs []string
	for _, f := range subs {
		switch classifyFile(f) {
		case ftDir:
			dirs = append(dirs, f.Name())
		case ftImg:
			itm := makeItem(f)
			itm.Image = &store.Image{}
			imgs = append(imgs, itm)
		case ftVid:
			itm := makeItem(f)
			itm.Video = &store.Video{}
			vids = append(vids, itm)
		}
	}
	sdir.SubDirectories = dirs

	newItems := make([]*store.Item, 0, len(imgs)+len(vids))
	for _, newVid := range vids {
		oldVid := oldItems[*newVid.Name]
		newItems = append(newItems, mergeVideo(oldVid, newVid))
	}
	for _, newImg := range imgs {
		oldImg, ok := oldItems[*newImg.Name]
		if !ok || oldImg.FileTimestamp == nil || *oldImg.FileTimestamp != *newImg.FileTimestamp || forceReload {
			fullPath := filepath.Join(origDir, *newImg.Name)
			if err := LoadImageFile(fullPath, newImg); err != nil {
				log.Printf("%s: %v", fullPath, err)
			}
		}
		newItems = append(newItems, mergeImage(oldImg, newImg))
		if len(newImg.Keywords) == 0 {
			log.Printf("%s: no keywords", filepath.Join(origDir, *newImg.Name))
		}
	}
	sdir.Items = newItems
	return nil
}

func addSubKeywords(p *Photo) {
	added := make(map[string]bool, len(p.Keywords)+3)
	for _, kwd := range p.Keywords {
		added[kwd] = true
	}
	for _, kwd := range p.Keywords {
		for _, sub := range strings.Split(kwd, " ") {
			if len(sub) > 0 && !added[sub] {
				added[sub] = true
				p.SubKeywords = append(p.SubKeywords, sub)
			}
		}
	}
}

// ProtoToPhoto constructs an in-memory Photo from a store.Item.
func ProtoToPhoto(album *Album, sitem *store.Item) *Photo {
	p := &Photo{
		Album:    album,
		Name:     sitem.GetName(),
		FileTime: time.Unix(0, 0),
		ItemTime: time.Unix(0, 0),
	}
	if sitem.FileTimestamp != nil {
		p.FileTime = ProtoToTime(*sitem.FileTimestamp)
	}
	if sitem.ItemTimestamp != nil {
		p.ItemTime = ProtoToTime(*sitem.ItemTimestamp)
	}
	hasSubs := false
	p.Keywords = make([]string, len(sitem.Keywords))
	for i, kwd := range sitem.Keywords {
		p.Keywords[i] = kwd
		if strings.ContainsRune(kwd, ' ') {
			hasSubs = true
		}
	}
	if hasSubs {
		addSubKeywords(p)
	}
	if simg := sitem.Image; simg != nil {
		p.Height = simg.GetHeight()
		p.Width = simg.GetWidth()
		p.RotateDegrees = simg.GetRotateDegrees()
		if sst := simg.Stereo; sst != nil {
			p.Stereo = &Stereo{
				Dx:    sst.GetDx(),
				Dy:    sst.GetDy(),
				AnaDx: sst.GetAnaDx(),
				AnaDy: sst.GetAnaDy(),
			}
		}
	}
	return p
}

var pstZone = time.FixedZone("PST", -8*3600)

func tryGuessDate(s, layout string) (time.Time, error) {
	if len(s) < len(layout) {
		return time.Time{}, errors.New("too short")
	}
	return time.ParseInLocation(layout, s[:len(layout)], pstZone)
}

func guessTimeFromPath(relPat string) (time.Time, error) {
	splits := strings.Split(relPat, "/")
	for i := len(splits) - 1; i >= 0; i-- {
		seg := splits[i]
		for _, layout := range []string{"2006-01-02", "2006-01", "2006", "Jan _2, 2006"} {
			if tim, err := tryGuessDate(seg, layout); err == nil {
				return tim, nil
			}
		}
		if len(seg) >= 11 {
			if tim, err := time.Parse("Jan _2, 2006 MST", seg[:11]+" PST"); err == nil {
				return tim, nil
			}
		}
	}
	return time.Time{}, errors.New("no date guess")
}

// Finalize computes LastModified, sorts photos by filename when a directory date can be
// inferred, and clamps photo ItemTime values more than 1 year after the directory date.
func (a *Album) Finalize() {
	a.LastModified = time.Unix(0, 0)
	for _, img := range a.Photos {
		if img.FileTime.After(a.LastModified) {
			a.LastModified = img.FileTime
		}
	}
	if guessed, err := guessTimeFromPath(a.RelPat); err == nil {
		sort.Slice(a.Photos, func(i, j int) bool {
			return a.Photos[i].Name < a.Photos[j].Name
		})
		maxTime := guessed.Add(365 * 24 * time.Hour)
		for i, img := range a.Photos {
			if img.ItemTime.After(maxTime) {
				img.ItemTime = guessed.Add(time.Duration(i) * time.Hour)
			}
		}
	}
}

// ProtoToAlbum constructs an in-memory Album from a store.Directory.
func ProtoToAlbum(sdir *store.Directory, relPat string) *Album {
	album := &Album{
		RelPat:    relPat,
		IndexTime: time.Unix(0, 0),
	}
	if sdir.DirectoryTimestamp != nil {
		album.IndexTime = ProtoToTime(*sdir.DirectoryTimestamp)
	}
	photos := make([]*Photo, 0, len(sdir.Items))
	for _, item := range sdir.Items {
		if item.Video == nil {
			photos = append(photos, ProtoToPhoto(album, item))
		}
	}
	album.Photos = photos
	album.Finalize()
	return album
}
