package model

import (
  "strconv"
  "strings"
  "time"
)

import "github.com/golang/protobuf/proto"
import "toutizes.com/store"

type Stereo struct {
  Dx, Dy, AnaDx, AnaDy float32
}

type Image struct {
  dir *Directory
  name string
  keywords []string
  sub_keywords []string
  file_time time.Time
  item_time time.Time
  rotate_degrees int32
  stereo *Stereo
  Id int
}

func (img *Image) Directory() *Directory { return img.dir }
func (img *Image) Name() string { return img.name }
func (img *Image) Keywords() []string { return img.keywords }
func (img *Image) SubKeywords() []string { return img.sub_keywords }
func (img *Image) FileTime() time.Time { return img.file_time }
func (img *Image) ItemTime() time.Time { return img.item_time }
func (img *Image) RotateDegrees() int32 { return img.rotate_degrees }
func (img *Image) Stereo() *Stereo { return img.stereo }

func (img *Image) Intern(indexer *Indexer) {
  img.name = indexer.Intern(img.name)
  for i, kwd := range img.keywords {
    img.keywords[i] = indexer.Intern(kwd)
  }
  for i, kwd := range img.sub_keywords {
    img.sub_keywords[i] = indexer.Intern(kwd)
  }
}

func addSubKeywords(img *Image) {
  added := make(map[string]bool, len(img.keywords) + 3)
  for _, kwd := range img.Keywords() {
    added[kwd] = true
  }
  for _, kwd := range img.keywords {
    for _, sub_kwd := range strings.Split(kwd, " ") {
      if len(sub_kwd) > 0 && !added[sub_kwd] {
        added[sub_kwd] = true
        img.sub_keywords = append(img.sub_keywords, sub_kwd)
      }
    }
  }
}

// Constructor from a proto.
func ProtoToImage(dir *Directory, sitem *store.Item) *Image {
  img := new(Image)
  img.name = *sitem.Name
  img.dir = dir
  if sitem.FileTimestamp != nil {
    img.file_time = ProtoToTime(*sitem.FileTimestamp)
  } else {
    img.file_time = time.Unix(0, 0)
  }
  if sitem.ItemTimestamp != nil {
    img.item_time = ProtoToTime(*sitem.ItemTimestamp)
  } else {
    img.item_time = time.Unix(0, 0)
  }
  has_subs := false
  img.keywords = make([]string, len(sitem.Keywords))
  // if len(sitem.Keywords) == 0 {
  //   log.Printf("%s %s: no keywords\n", dir.RelPat(), img.name);
  // }
  for i := 0; i < len(sitem.Keywords); i++ {
    kwd := sitem.Keywords[i]
    img.keywords[i] = kwd
    if strings.ContainsRune(kwd, ' ') {
      has_subs = true
    }
  }
  if has_subs {
    addSubKeywords(img)
  }
  if simg := sitem.Image; simg != nil {
    if simg.RotateDegrees != nil {
      img.rotate_degrees = *simg.RotateDegrees
    }
    if sstereo := simg.Stereo; sstereo != nil {
      stereo := new(Stereo)
      stereo.Dx = *sstereo.Dx
      stereo.Dy = *sstereo.Dy
      stereo.AnaDx = *sstereo.AnaDx
      stereo.AnaDy = *sstereo.AnaDy
      img.stereo = stereo
    }
  }
  return img
}

func (img *Image) ToProto() *store.Item {
  sitem := new(store.Item)
  sitem.Name = proto.String(img.name)
  if img.file_time != time.Unix(0, 0) {
    sitem.FileTimestamp = proto.Int64(TimeToProto(img.file_time))
  }
  if img.item_time != time.Unix(0, 0) {
    sitem.ItemTimestamp = proto.Int64(TimeToProto(img.item_time))
  }
  sitem.Keywords = img.keywords
  sitem.Image = new(store.Image)
  if img.rotate_degrees != 0 {
    sitem.Image.RotateDegrees = proto.Int32(img.rotate_degrees)
  }
  if img.stereo != nil {
    sitem.Image.Stereo = new(store.Stereo)
    sitem.Image.Stereo.Dx = proto.Float32(img.stereo.Dx)
    sitem.Image.Stereo.Dy = proto.Float32(img.stereo.Dy)
    sitem.Image.Stereo.AnaDx = proto.Float32(img.stereo.AnaDx)
    sitem.Image.Stereo.AnaDy = proto.Float32(img.stereo.AnaDy)
  }
  return sitem
}

type JsonImage struct {
  Id int                        // Image id
  Ad string                     // Album directory
  In string                     // Image filename in the album dir
  Its int64                     // Image taken timestamp
  Fts int64                     // Image file timestamp
  Kwd []string                  // Keywords
  Stereo *Stereo                // Stereo info
}

func (img *Image) Json(jimg *JsonImage) {
  jimg.Id = img.Id
  jimg.Ad = img.dir.RelPat()
  jimg.In = img.name
  jimg.Its = img.item_time.Unix()
  jimg.Fts = img.file_time.Unix()
  jimg.Kwd = img.keywords
  jimg.Stereo = img.stereo
}

func (img *Image) String() string {
  var parts []string
  parts = append(parts, "{name: '", img.name, "', ")
  parts = append(parts, "file_time: ", img.file_time.String(), ", ")
  parts = append(parts, "item_time: ", img.item_time.String(), ", ")
  parts = append(parts, "rotate: ", strconv.Itoa(int(img.rotate_degrees)), ", ")
  parts = append(parts, "keywords: [", strings.Join(img.keywords, ", "), "], ")
  parts = append(parts, "sub_keywords: [", strings.Join(img.sub_keywords, ", "), "]")
  parts = append(parts, "}")
  return strings.Join(parts, "")
}
