package model

import (
	"hash"
	"hash/fnv"
  "log"
  "strconv"
  "strings"
)

type keywordCounts struct {
  keyword string
  count int
}

type Indexer struct {
  keyword_counts map[string]*keywordCounts
  images_by_keyword map[string][]*Image
  images_by_id map[int]*Image
}

func NewIndexer() *Indexer {
  idx := new(Indexer)
  idx.keyword_counts = make(map[string]*keywordCounts)
  return idx
}

func (idx *Indexer) Intern(s string) string {
  if kwcnt, ok := idx.keyword_counts[s]; ok {
    kwcnt.count += 1
    return kwcnt.keyword
  }

  // Make a copy of s, in case if was a substring of a larger string.
  b := []byte(s)
  is := string(b)
  // Add to the map.
  idx.keyword_counts[is] = &keywordCounts{count: 1, keyword: is}
  return is
}

func (idx *Indexer) Images(kwd string) []*Image {
  imgs, ok := idx.images_by_keyword[DropAccents(kwd, nil)]
  if ok {
    return imgs
  } else {
    return nil
  }
}

func (idx *Indexer) Image(image_id int) *Image {
  img, ok := idx.images_by_id[image_id]
  if ok {
    return img
  } else {
    return nil
  }
}

func (idx *Indexer) addImageByKeyword(img *Image, kwd string) {
  imgs, ok := idx.images_by_keyword[kwd]
  if !ok {
    log.Printf("Unregistered keyword: %s\n", kwd)
  } else {
    idx.images_by_keyword[kwd] = append(imgs, img)
  }
}

func imageHash(h hash.Hash32, dir *Directory, img *Image) int {
	h.Reset()
	h.Write([]byte(dir.RelPat()))
	h.Write([]byte(img.Name()))
	bytes, err := img.FileTime().MarshalBinary()
	if err == nil {
		h.Write(bytes)
	}
	return int(h.Sum32())
}

func (idx *Indexer) BuildIndex(db *Database) int {
  drop_cache := make(map[string]string, len(idx.keyword_counts))

  idx.images_by_keyword = make(map[string][]*Image)
  for _, kwcnt := range idx.keyword_counts {
    idx.images_by_keyword[DropAccents(kwcnt.keyword, drop_cache)] =
      make([]*Image, 0, kwcnt.count)
  }
	hasher := fnv.New32a()
  num_images := 0
  for _, dir := range db.Directories() {
    for _, img := range dir.Images() {
      img.Id = imageHash(hasher, dir, img)
      num_images += 1
      idx.addImageByKeyword(img, DropAccents(img.Name(), drop_cache))
      for _, kwd := range img.Keywords() {
        idx.addImageByKeyword(img, DropAccents(kwd, drop_cache))
      }
      for _, kwd := range img.SubKeywords() {
        idx.addImageByKeyword(img, DropAccents(kwd, drop_cache))
      }
    }
  }
  idx.images_by_id = make(map[int]*Image, num_images)
  for _, dir := range db.Directories() {
    for _, img := range dir.Images() {
      idx.images_by_id[img.Id] = img
    }
  }
  return num_images
}

func (idx *Indexer) String() string {
  var parts []string
  for kwd, imgs := range idx.images_by_keyword {
    parts = append(parts, kwd, ": ", strconv.Itoa(len(imgs)), "\n")
  }
  return strings.Join(parts, "")
}
