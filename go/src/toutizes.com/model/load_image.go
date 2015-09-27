package model

import (
  "bufio"
  "errors"
  "log"
  "os"
  "os/exec"
  "strings"
  "time"
)

import (
  "github.com/rwcarlsen/goexif/exif"
  "github.com/rwcarlsen/goexif/tiff"
)

import "toutizes.com/store"

func tagTime(ex *exif.Exif, ts exif.FieldName) (t time.Time, err error) {
  t = time.Unix(0, 0)
  tg, err := ex.Get(ts)
  if err != nil {
    return
  }
  if tg.TypeCategory() != tiff.StringVal {
    err = errors.New("Tag not a string")
    return
  }
  t, err = time.Parse("2006:01:02 15:04:05", tg.StringVal())
  return
}

func getKeywords(file string) (keywords []string, err error) {
  cmd := exec.Command(
    *BinRoot + "convert", file, "-format", "%[IPTC:2:25]", "info:")
  out, err := cmd.Output()
  if err != nil {
    log.Printf("%s: %s\n", file, err.Error())
  } else {
    keywords = strings.Split(string(out), ";")
  }
  return
}

func LoadImageFile(file string, image *store.Item) error {
  fi, err := os.Open(file)
  if err != nil {
    return err
  }
  defer fi.Close()
  exif, err := exif.Decode(bufio.NewReader(fi))
  if err != nil {
    return err
  }
  var image_time time.Time
  found_time := false
  dto, err := tagTime(exif, "DateTimeOriginal")
  if err == nil {
    image_time = dto
    found_time = true
  }
  if !found_time {
    dtd, err := tagTime(exif, "DateTimeDigitized")
    if err == nil {
      log.Printf("Using DateTimeDigitized for: %s (%s)\n", file, dtd)
      image_time = dtd
      found_time = true
    }
  }
  if !found_time {
    dt, err := tagTime(exif, "DateTime")
    if err == nil {
      log.Printf("Using DateTime for: %s (%s)\n", file, dt)
      image_time = dt
      found_time = true
    }
  }
  if !found_time {
    log.Printf("Using file time for: %s\n", file)
    image_time = ProtoToTime(*image.FileTimestamp)
  }
  its := TimeToProto(image_time)
  image.ItemTimestamp = &its
  kwds, err := getKeywords(file)
  if err == nil {
    image.Keywords = kwds
  }
  return nil
}
