package model

import (
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
  images []*Image
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
  if image_id < 0 || image_id >= len(idx.images) {
    return nil
  } else {
    return idx.images[image_id]
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

func (idx *Indexer) BuildIndex(db *Database) int {
  drop_cache := make(map[string]string, len(idx.keyword_counts))

  idx.images_by_keyword = make(map[string][]*Image)
  for _, kwcnt := range idx.keyword_counts {
    idx.images_by_keyword[DropAccents(kwcnt.keyword, drop_cache)] =
      make([]*Image, 0, kwcnt.count)
  }
  num_images := 0
  for _, dir := range db.Directories() {
    for _, img := range dir.Images() {
      img.Id = num_images
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
  idx.images = make([]*Image, num_images)
  for _, dir := range db.Directories() {
    for _, img := range dir.Images() {
      idx.images[img.Id] = img
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
