package model

import (
  "bufio"
  "errors"
  "fmt"
  "log"
  "os"
  "os/exec"
  "strconv"
  "strings"
  "time"
)

import (
  "github.com/golang/protobuf/proto"
  "github.com/rwcarlsen/goexif/exif"
)

import "toutizes.com/store"

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

func tagInt(ex *exif.Exif, ts exif.FieldName) (val int32, err error) {
  val = -1
  tg, err := ex.Get(ts)
  if err != nil {
    return
  }
  if tg.Count != 1 {
    err = errors.New(fmt.Sprintf("Tag count not 1: %s", ts))
    return
  }
  int_val, err := tg.Int(0)
  if err == nil {
    val = int32(int_val)
  }
  return
}

		
func parseKeywords(s string) (keywords []string) {
	if strings.HasPrefix(s, "convert: unknown image property") || len(s) == 0 {
		return nil
	}
	if strings.HasSuffix(s, "\n") {
		s = s[0:len(s) - 1]
	}
	return strings.Split(s, ";")
}

func getImageInfo(file string) (height int, width int, keywords []string, err error) {
  args := []string{
		*BinRoot + "convert", file, "-format", "%h\\ %w\\n%[IPTC:2:25]\\n", "info:"}
  cmd := exec.Command(args[0], args[1:]...)
  out, err := cmd.CombinedOutput()
	if err != nil {
		// If there is no IPTC entry.
		log.Printf("IPTC failed, trying just sizes: %s\n", file)
		log.Printf("Cmd: %s =>\n%s\n", strings.Join(args, " "), string(out))
		cmd = exec.Command(*BinRoot + "convert", file, "-format", "%h %w\\n", "info:")
		out, err = cmd.Output()
	}
  if err != nil {
		return
  }
	// First line is "height width"
	// Second line is "kwd;kwd;kwd"
	lines := strings.Split(string(out), "\n")
	if len(lines) >= 1 {
		h_w := strings.SplitN(lines[0], " ", 2)
		height, err = strconv.Atoi(h_w[0])
		if err != nil {
			return
		}
		width, err = strconv.Atoi(h_w[1])
		if err != nil {
			return
		}
	}
	if len(lines) >= 2 {
		keywords = parseKeywords(lines[1])
	}
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
  height, width, kwds, err := getImageInfo(file)
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
