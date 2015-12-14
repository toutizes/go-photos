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

func getImageInfo(file string) (height int, width int, keywords []string, err error) {
  cmd := exec.Command(
    *BinRoot + "convert", file, "-format", "%h %w %[IPTC:2:25]", "info:")
  out, err := cmd.Output()
  if err != nil {
    log.Printf("%s: %s\n", file, err.Error())
  } else {
    splits := strings.SplitN(string(out), " ", 3)
    height, err = strconv.Atoi(splits[0])
    if err != nil {
      log.Printf("%s: %s\n", file, err.Error())
      return
    }
    width, err = strconv.Atoi(splits[1])
    if err != nil {
      log.Printf("%s: %s\n", file, err.Error())
      return
    }
    keywords = strings.Split(splits[2], ";")
  }
  return
}


func LoadImageFile(file string, image *store.Item) error {
  fi, err := os.Open(file)
  if err != nil {
    return err
  }
  defer fi.Close()
  ex, err := exif.Decode(bufio.NewReader(fi))
  if err != nil {
    return err
  }
  var image_time time.Time
  found_time := false
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
  if !found_time {
    log.Printf("Using file time for: %s\n", file)
    image_time = ProtoToTime(*image.FileTimestamp)
  }
  its := TimeToProto(image_time)
  image.ItemTimestamp = &its
  height, width, kwds, err := getImageInfo(file)
  if err == nil {
    image.Keywords = kwds
    image.Image = new(store.Image)
    image.Image.Height = proto.Int32(int32(height))
    image.Image.Width = proto.Int32(int32(width))
  } else {
    log.Printf("%s: %s", file, err.Error())
  }
  return nil
}
