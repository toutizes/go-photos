package model

import (
	"hash"
	"hash/fnv"
  "log"
  "sort"
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
  images_by_subkeyword map[string][]*Image
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
  return idx.ImagesWithSubkeywords(kwd, true)
}

func (idx *Indexer) ImagesWithSubkeywords(kwd string, includeSubkeywords bool) []*Image {
  // Get images from main keyword index
  imgs, ok := idx.images_by_keyword[kwd]
  if !ok {
    imgs = make([]*Image, 0)
  }
  
  // If not including sub-keywords, return main keyword images only
  if !includeSubkeywords {
    return imgs
  }
  
  // If including sub-keywords, merge with sub-keyword index
  subImgs, subOk := idx.images_by_subkeyword[kwd]
  if !subOk || len(subImgs) == 0 {
    // No sub-keyword images, return main keyword images
    return imgs
  }
  
  // Merge the two slices, avoiding duplicates while preserving order
  seen := make(map[int]bool)
  merged := make([]*Image, 0, len(imgs)+len(subImgs))
  
  // Add main keyword images first (they keep their priority)
  for _, img := range imgs {
    if !seen[img.Id] {
      seen[img.Id] = true
      merged = append(merged, img)
    }
  }
  
  // Add sub-keyword images that aren't already included
  for _, img := range subImgs {
    if !seen[img.Id] {
      seen[img.Id] = true
      merged = append(merged, img)
    }
  }
  
  return merged
}

func (idx *Indexer) Image(image_id int) *Image {
  img, ok := idx.images_by_id[image_id]
  if ok {
    return img
  } else {
    return nil
  }
}


type StringBySize []string

func (a StringBySize) Len() int           { return len(a) }
func (a StringBySize) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a StringBySize) Less(i, j int) bool { return len(a[i]) < len(a[j]) }

func (idx *Indexer) MatchingKeywords(pat string) []string {
	return idx.MatchingKeywordsWithSubkeywords(pat, true)
}

func (idx *Indexer) MatchingKeywordsWithSubkeywords(pat string, includeSubkeywords bool) []string {
	keywordSet := make(map[string]bool)
	
	// Add matching main keywords
	for kwd := range idx.images_by_keyword {
		if strings.Contains(kwd, pat) {
			keywordSet[kwd] = true
		}
	}
	
	// Add matching sub-keywords if requested
	if includeSubkeywords {
		for kwd := range idx.images_by_subkeyword {
			if strings.Contains(kwd, pat) {
				keywordSet[kwd] = true
			}
		}
	}
	
	// Convert to slice
	a := make([]string, 0, len(keywordSet))
	for kwd := range keywordSet {
		a = append(a, kwd)
	}
	
	sort.Sort(StringBySize(a))
	return a
}

func (idx *Indexer) addImageByKeyword(img *Image, kwd string) {
  imgs, ok := idx.images_by_keyword[kwd]
  if !ok {
    log.Printf("Unregistered keyword: %s\n", kwd)
  } else {
    idx.images_by_keyword[kwd] = append(imgs, img)
  }
}

func (idx *Indexer) addImageBySubkeyword(img *Image, kwd string) {
  imgs, ok := idx.images_by_subkeyword[kwd]
  if !ok {
    log.Printf("Unregistered sub-keyword: %s\n", kwd)
  } else {
    idx.images_by_subkeyword[kwd] = append(imgs, img)
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
  idx.images_by_subkeyword = make(map[string][]*Image)
  for _, kwcnt := range idx.keyword_counts {
    idx.images_by_keyword[kwcnt.keyword] = make([]*Image, 0, kwcnt.count)
    idx.images_by_subkeyword[kwcnt.keyword] = make([]*Image, 0, kwcnt.count)
		dropped := DropAccents(kwcnt.keyword, drop_cache)
		if dropped != kwcnt.keyword {
			idx.images_by_keyword[dropped] = make([]*Image, 0, kwcnt.count)
			idx.images_by_subkeyword[dropped] = make([]*Image, 0, kwcnt.count)
		}
  }
	hasher := fnv.New32a()
  num_images := 0
  for _, dir := range db.Directories() {
    for _, img := range dir.Images() {
      img.Id = imageHash(hasher, dir, img)
			img.Rank = num_images
      num_images += 1
      idx.addImageByKeyword(img, DropAccents(img.Name(), drop_cache))
      for _, kwd := range img.Keywords() {
        idx.addImageByKeyword(img, kwd)
				dropped := DropAccents(kwd, drop_cache)
				if dropped != kwd {
					idx.addImageByKeyword(img, DropAccents(kwd, drop_cache))
				}
      }
      for _, kwd := range img.SubKeywords() {
        idx.addImageBySubkeyword(img, kwd)
				dropped := DropAccents(kwd, drop_cache)
				if dropped != kwd {
					idx.addImageBySubkeyword(img, DropAccents(kwd, drop_cache))
				}
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
  parts = append(parts, "Keywords:\n")
  for kwd, imgs := range idx.images_by_keyword {
    parts = append(parts, "  ", kwd, ": ", strconv.Itoa(len(imgs)), "\n")
  }
  parts = append(parts, "Sub-keywords:\n")
  for kwd, imgs := range idx.images_by_subkeyword {
    parts = append(parts, "  ", kwd, ": ", strconv.Itoa(len(imgs)), "\n")
  }
  return strings.Join(parts, "")
}
