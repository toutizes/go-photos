package model

import (
	"bufio"
	"bytes"
	"golang.org/x/text/encoding/charmap"
	"image"
	_ "image/jpeg"
	"log"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

import (
	"github.com/golang/protobuf/proto"
	"github.com/rwcarlsen/goexif/exif"
	// For parsing JPEG segments
	jpegstructure "github.com/dsoprea/go-jpeg-image-structure/v2"
	// For parsing IPTC data
	iptc "github.com/dsoprea/go-iptc"
)

import "toutizes.com/go-photwo/backend/store"

// Replacer that converts mac ascii extended codes to unicode.
var macAccentsCleaner = strings.NewReplacer(
	// "\x80", "Ä",
	// "\x81", "Å",
	"\x82", "Ç",
	"\x83", "É",
	"\x84", "Ñ",
	"\x85", "Ö",
	"\x86", "Ü",
	"\x87", "á",
	"\x88", "à",
	"\x89", "â",
	// "\x8A", "ä",
	// "\x8B", "ã",
	// "\x8C", "å",
	"\x8D", "ç",
	"\x8E", "é",
	"\x8F", "è",
	"\x90", "ê",
	"\x91", "ë",
	// "\x92", "í",
	// "\x93", "ì",
	"\x94", "î",
	"\x95", "ï",
	"\x96", "ñ",
	// "\x97", "ó",
	// "\x98", "ò",
	// "\x99", "ô",
	// "\x9A", "ö",
	// "\x9B", "õ",
	// "\x9C", "ú",
	"\x9D", "ù",
	"\x9E", "û",
	"\x9F", "ü",
	// "\xA0", "†",
	// "\xA1", "°",
	// "\xA2", "¢",
	// "\xA3", "£",
	// "\xA4", "§",
	// "\xA5", "•",
	// "\xA6", "¶",
	// "\xA7", "ß",
	// "\xA8", "®",
	// "\xA9", "©",
	// "\xAA", "™",
	// "\xAB", "´",
	// "\xAC", "¨",
	// "\xAD", "≠",
	// "\xAE", "Æ",
	// "\xAF", "Ø",
	// "\xB0", "∞",
	// "\xB1", "±",
	// "\xB2", "≤",
	// "\xB3", "≥",
	// "\xB4", "¥",
	// "\xB5", "µ",
	// "\xB6", "∂",
	// "\xB7", "∑",
	// "\xB8", "∏",
	// "\xB9", "π",
	// "\xBA", "∫",
	// "\xBB", "ª",
	// "\xBC", "º",
	// "\xBD", "Ω",
	// "\xBE", "æ",
	// "\xBF", "ø",
	// "\xC0", "¿",
	// "\xC1", "¡",
	// "\xC2", "¬",
	// "\xC3", "√",
	// "\xC4", "ƒ",
	// "\xC5", "≈",
	// "\xC6", "∆",
	// "\xC7", "«",
	// "\xC8", "»",
	// "\xC9", "…",
	// "\xCA", "n",
	"\xCB", "À",
	// "\xCC", "Ã",
	// "\xCD", "Õ",
	"\xCE", "Œ",
	"\xCF", "œ",
	// "\xD0", "–",
	// "\xD1", "—",
	// "\xD2", "“",
	// "\xD3", "”",
	// "\xD4", "‘",
	// "\xD5", "’",
	// "\xD6", "÷",
	// "\xD7", "◊",
	// "\xD8", "ÿ",
	// "\xD9", "Ÿ",
	// "\xDA", "⁄",
	// "\xDB", "€",
	// "\xDC", "‹",
	// "\xDD", "›",
	// "\xDE", "ﬁ",
	// "\xDF", "ﬂ",
	// "\xE0", "‡",
	// "\xE1", "·",
	// "\xE2", "‚",
	// "\xE3", "„",
	// "\xE4", "‰",
	"\xE5", "Â",
	"\xE6", "Ê",
	// "\xE7", "Á",
	"\xE8", "Ë",
	"\xE9", "È",
	// "\xEA", "Í",
	"\xEB", "Î",
	"\xEC", "Ï",
	// "\xED", "Ì",
	// "\xEE", "Ó",
	"\xEF", "Ô",
	// "\xF0", "",
	// "\xF1", "Ò",
	// "\xF2", "Ú",
	"\xF3", "Û",
	"\xF4", "Ù",
	// "\xF5", "ı",
	// "\xF6", "ˆ",
	// "\xF7", "˜",
	// "\xF8", "¯",
	// "\xF9", "˘",
	// "\xFA", "˙",
	// "\xFB", "˚",
	// "\xFC", "¸",
	// "\xFD", "˝",
	// "\xFE", "˛",
	// "\xFF", "ˇ"
)

func tagTime(ex *exif.Exif, ts exif.FieldName) (t time.Time, err error) {
	t = time.Unix(0, 0)
	tg, err := ex.Get(ts)
	if err != nil {
		return
	}
	str_val, err := tg.StringVal()
	if err != nil {
		return
	}
	t, err = time.Parse("2006:01:02 15:04:05", str_val)
	return
}

func LoadImageFile(file string, image *store.Item) error {
	fi, err := os.Open(file)
	if err != nil {
		log.Printf("Error: %s\n", err.Error())
		return err
	}
	defer fi.Close()
	found_time := false
	var image_time time.Time
	ex, err := exif.Decode(bufio.NewReader(fi))
	if err == nil {
		dto, err := tagTime(ex, exif.DateTimeOriginal)
		if err == nil {
			image_time = dto
			found_time = true
		}
		if !found_time {
			dtd, err := tagTime(ex, exif.DateTimeDigitized)
			if err == nil {
				log.Printf("Using DateTimeDigitized for: %s (%s)\n", file, dtd)
				image_time = dtd
				found_time = true
			}
		}
		if !found_time {
			dt, err := tagTime(ex, exif.DateTime)
			if err == nil {
				log.Printf("Using DateTime for: %s (%s)\n", file, dt)
				image_time = dt
				found_time = true
			}
		}
	}
	if !found_time {
		log.Printf("Using file time for: %s\n", file)
		image_time = ProtoToTime(*image.FileTimestamp)
	}
	its := TimeToProto(image_time)
	image.ItemTimestamp = &its
	height, width, kwds, err := GetImageInfo2(file)
	if err == nil {
		if kwds != nil && len(kwds) > 0 {
			image.Keywords = kwds
		}
		image.Image = new(store.Image)
		image.Image.Height = proto.Int32(int32(height))
		image.Image.Width = proto.Int32(int32(width))
	} else {
		log.Printf("Info got error %s: %s", file, err.Error())
	}
	return nil
}

func GetImageInfo2(filepath string) (height int, width int, keywords []string, err error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return
	}

	reader := bytes.NewReader(data)

	// Decode image configuration using the reader based on the byte slice
	config, _, err := image.DecodeConfig(reader)
	if err != nil {
		log.Printf("Error decoding image config: %s\n", err.Error())
		return
	}

	// Named return values
	height = config.Height
	width = config.Width

	parser := jpegstructure.NewJpegMediaParser()
	intfc, err := parser.ParseBytes(data)
	if err != nil {
		log.Printf("Parsing jpeg media: %s\n", err.Error())
		return
	}

	sl := intfc.(*jpegstructure.SegmentList)
	_, segment, err := sl.FindIptc()
	if err != nil {
		log.Printf("Finding iptc: %s\n", err.Error())
		return
	}

	tags, err := segment.Iptc()
	if err != nil {
		log.Printf("Iptc tags: %s\n", err.Error())
		return
	}

	kwdBytes := tags[iptc.StreamTagKey{RecordNumber: 2, DatasetNumber: 25}]

	keywords = make([]string, 0, len(kwdBytes))
	for _, bytes := range kwdBytes {
		var decodedString string
		if utf8.Valid(bytes) {
      // fmt.Printf("valid utf8: %v\n", bytes)
      // for _, b := range bytes {
      //   fmt.Printf(" %x %d %c\n", b, b, b)
      // }
			decodedString = string(bytes)
      // fmt.Printf("decoded utf8: %s\n", decodedString)
		} else {
			// Assume Latin-1
			decoder := charmap.ISO8859_1.NewDecoder()
			decodedBytes, err := decoder.Bytes(bytes)
      // fmt.Printf("decoded latin1: %v",decodedBytes)
			if err != nil {
				// Should be rare for ISO-8859-1, but handle just in case
				log.Printf("%s: Failed to decode supposed ISO-8859-1 keyword bytes: %v", filepath, err)
				// Pray! (As we did before)
				decodedString = string(bytes)
			} else {
				decodedString = string(decodedBytes) // Now contains valid Go UTF-8 string
			}
		}
		for _, kwd := range strings.Split(decodedString, ";") {
			keywords = append(keywords, kwd)
		}
	}

	return
}
