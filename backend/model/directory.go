package model

import (
  "errors"
  "log"
  "sort"
  "strconv"
  "strings"
  "time"
)

import "github.com/golang/protobuf/proto"
import "toutizes.com/go-photwo/backend/store"

type Directory struct {
  rel_pat string
  index_time time.Time
  last_modified time.Time
  order_by string
  images []*Image
  // TODO: videos []Video
}

func (dir *Directory) OrderBy() string { return dir.order_by }
func (dir *Directory) Images() []*Image { return dir.images }
func (dir *Directory) RelPat() string { return dir.rel_pat }
func (dir *Directory) Time() time.Time { return dir.index_time }
func (dir *Directory) ItemTime() time.Time { 
  if len(dir.images) == 0 {
    return dir.index_time
  }
  return dir.images[0].ItemTime()
}


func (dir *Directory) Intern(indexer *Indexer) {
  for _, image := range dir.images {
    image.Intern(indexer)
  }
}

func tryGuess(s string, p string) (tim time.Time, err error) {
  if len(s) < len(p) {
    err = errors.New("")
    return
  }
  tim, err = time.ParseInLocation(p, s[0:len(p)], time.FixedZone("PST", -8 * 3600))
  return
}

func (dir *Directory) guessTimeFromName() (tim time.Time, err error) {
  splits := strings.Split(dir.rel_pat, "/")
  for i := len(splits) - 1; i >=0; i-- {
    tim, err = tryGuess(splits[i], "2006-01-02")
    if err == nil {
      return
    }
    tim, err = tryGuess(splits[i], "2006-01")
    if err == nil {
      return
    }
    tim, err = tryGuess(splits[i], "2006")
    if err == nil {
      return
    }
    tim, err = tryGuess(splits[i], "Jan _2, 2006")
    if err == nil {
      return
    }
    // Special case for days 1:9
    if len(splits[i]) >= 11 {
      tim, err = time.Parse("Jan _2, 2006 MST", splits[i][0:11] + " PST")
      if err == nil {
        return
      }
    }
  }
  err = errors.New("No guess.")
  return
}

// Sort images by name
type ByName[] *Image
func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name() < a[j].Name() }


// Finalize the directory after loading or creating.
func (dir *Directory) Finalize() {
  dir.last_modified = time.Unix(0, 0)
  for _, img := range dir.images {
    if img.FileTime().After(dir.last_modified) {
      dir.last_modified = img.FileTime()
    }
  }
  guessed, err := dir.guessTimeFromName()
  if err == nil {
    // TODO cleverly sort images before assigining time.  Use seq number in names
    // such as "Picture 1.jpg", "Image04.jpg".  Move all non-seq ones at the end?
    sort.Sort(ByName(dir.images))
    max_time := guessed.Add(365 * 24 * time.Hour)
    for i, img := range dir.images {
      if img.ItemTime().After(max_time) {
        fixed_time := guessed.Add(time.Duration(i) * time.Hour)
//         log.Printf("%s/%s: Fix time to '%s' (was '%s')\n", dir.rel_pat,
//                    img.Name(), fixed_time.String(), img.ItemTime().String())
        img.FixItemTime(fixed_time)
      }
    }
  }
}

func ProtoToDirectory(sdir *store.Directory, rel_pat string) *Directory {
  dir := new(Directory)
  dir.rel_pat = rel_pat
  if sdir.DirectoryTimestamp != nil {
    dir.index_time = ProtoToTime(*sdir.DirectoryTimestamp)
  } else {
    dir.index_time = time.Unix(0, 0)
  }
  images := make([]*Image, 0)
  for i := 0; i < len(sdir.Items); i++ {
    if sdir.Items[i].Video == nil {
      images = append(images, ProtoToImage(dir, sdir.Items[i]))
    }
  }
  dir.images = images 
  dir.Finalize()
  for _, img := range dir.images {
		if len(img.Keywords()) == 0 {
			log.Printf("No keywords: %s/%s\n", rel_pat, img.Name())
		} else if (img.Keywords()[0] == "") {
			log.Printf("Empty keyword: %s/%s (%v)\n", rel_pat, img.Name(), img.Keywords())
		}
	}
  return dir
}

func (dir *Directory) ToProto() *store.Directory {
  sdir := new(store.Directory)
  if dir.index_time != time.Unix(0, 0) {
    sdir.DirectoryTimestamp = proto.Int64(TimeToProto(dir.index_time))
  }
  for _, img := range dir.images {
    sdir.Items = append(sdir.Items, img.ToProto())
  }
  return sdir
}

type JsonDirectory struct {
  Id string                     // Album path
  Ats int64                     // Timetamp of first image in album
  Dts int64                     // Timestamp of album directory
  Nimgs int                     // Number of images in album
  Cov int                       // Cover image id
  CovName string                // Cover image name
  PreviewIds []int              // First 4 photo IDs for previews
  PreviewNames []string         // First 4 photo names for previews
}

func (dir *Directory) Json(jdir *JsonDirectory) {
  jdir.Id = dir.rel_pat
  if len(dir.images) > 0 {
    jdir.Ats = dir.images[0].ItemTime().Unix()  // Use first image's item timestamp
  } else {
    jdir.Ats = dir.index_time.Unix()  // Fallback if no images
  }
  jdir.Dts = dir.last_modified.Unix()
  jdir.Nimgs = len(dir.images)
  if len(dir.images) > 0 {
    img0 := dir.images[0]
    jdir.Cov = img0.Id
    jdir.CovName = img0.Name()
    
    // Populate images 1, 2, and 3 for previews (cover is already image 0)
    previewCount := len(dir.images) - 1  // Exclude the first image (cover)
    if previewCount > 3 {
      previewCount = 3
    }
    if previewCount > 0 {
      jdir.PreviewIds = make([]int, previewCount)
      jdir.PreviewNames = make([]string, previewCount)
      for i := 0; i < previewCount; i++ {
        jdir.PreviewIds[i] = dir.images[i+1].Id      // Start from index 1
        jdir.PreviewNames[i] = dir.images[i+1].Name() // Start from index 1
      }
    }
  }
}

func (dir *Directory) String() string {
  var parts []string
  parts = append(parts, "{name: ", dir.rel_pat, ", ")
  parts = append(parts, "n_images: ", strconv.Itoa(len(dir.images)), ", ")
  parts = append(parts, "last_modifed: ", dir.last_modified.String())
  parts = append(parts, "index_time: ", dir.index_time.String())
  parts = append(parts, "}")
  return strings.Join(parts, "")
}
