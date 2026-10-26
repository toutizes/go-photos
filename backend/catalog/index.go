package catalog

import (
	"hash"
	"hash/fnv"
	"sort"
	"strings"
)

type keywordCountEntry struct {
	keyword string
	count   int
}

// Index holds the inverted keyword/sub-keyword maps and ID lookup table for a Catalog.
type Index struct {
	keywordCounts map[string]*keywordCountEntry
	byKeyword     map[string][]*Photo
	bySubKeyword  map[string][]*Photo
	byID          map[int]*Photo
}

// NewIndex creates a new Index ready for string interning and building.
func NewIndex() *Index {
	return &Index{
		keywordCounts: make(map[string]*keywordCountEntry),
	}
}

// Intern deduplicates strings in memory and counts occurrences for slice preallocation.
func (idx *Index) Intern(s string) string {
	if entry, ok := idx.keywordCounts[s]; ok {
		entry.count++
		return entry.keyword
	}
	cloned := strings.Clone(s)
	idx.keywordCounts[cloned] = &keywordCountEntry{keyword: cloned, count: 1}
	return cloned
}

// internAlbum interns all photo names, keywords, and sub-keywords in an album.
func (idx *Index) internAlbum(album *Album) {
	for _, p := range album.Photos {
		p.Name = idx.Intern(p.Name)
		for i, kwd := range p.Keywords {
			p.Keywords[i] = idx.Intern(kwd)
		}
		for i, sub := range p.SubKeywords {
			p.SubKeywords[i] = idx.Intern(sub)
		}
	}
}

func computePhotoHash(h hash.Hash32, album *Album, p *Photo) int {
	h.Reset()
	h.Write([]byte(album.RelPat))
	h.Write([]byte(p.Name))
	if b, err := p.FileTime.MarshalBinary(); err == nil {
		h.Write(b)
	}
	return int(h.Sum32())
}

func appendUniqueByRank(slice []*Photo, p *Photo) []*Photo {
	if n := len(slice); n > 0 && slice[n-1] == p {
		return slice
	}
	return append(slice, p)
}

// Build populates photo IDs, ranks, and inverted keyword indices across all albums.
func (idx *Index) Build(albums []*Album) int {
	dropCache := make(map[string]string, len(idx.keywordCounts))
	idx.byKeyword = make(map[string][]*Photo, len(idx.keywordCounts)*2)
	idx.bySubKeyword = make(map[string][]*Photo, len(idx.keywordCounts)*2)

	for _, entry := range idx.keywordCounts {
		idx.byKeyword[entry.keyword] = make([]*Photo, 0, entry.count)
		idx.bySubKeyword[entry.keyword] = make([]*Photo, 0, entry.count)
		if dropped := DropAccents(entry.keyword, dropCache); dropped != entry.keyword {
			idx.byKeyword[dropped] = make([]*Photo, 0, entry.count)
			idx.bySubKeyword[dropped] = make([]*Photo, 0, entry.count)
		}
	}

	hasher := fnv.New32a()
	numPhotos := 0
	for _, album := range albums {
		for _, p := range album.Photos {
			p.Id = computePhotoHash(hasher, album, p)
			p.Rank = numPhotos
			numPhotos++

			nameKey := DropAccents(p.Name, dropCache)
			idx.byKeyword[nameKey] = appendUniqueByRank(idx.byKeyword[nameKey], p)

			for _, kwd := range p.Keywords {
				idx.byKeyword[kwd] = appendUniqueByRank(idx.byKeyword[kwd], p)
				if dropped := DropAccents(kwd, dropCache); dropped != kwd {
					idx.byKeyword[dropped] = appendUniqueByRank(idx.byKeyword[dropped], p)
				}
			}
			for _, sub := range p.SubKeywords {
				idx.bySubKeyword[sub] = appendUniqueByRank(idx.bySubKeyword[sub], p)
				if dropped := DropAccents(sub, dropCache); dropped != sub {
					idx.bySubKeyword[dropped] = appendUniqueByRank(idx.bySubKeyword[dropped], p)
				}
			}
		}
	}

	idx.byID = make(map[int]*Photo, numPhotos)
	for _, album := range albums {
		for _, p := range album.Photos {
			idx.byID[p.Id] = p
		}
	}
	return numPhotos
}

// PhotoByID returns the Photo with the given ID, or nil if not found.
func (idx *Index) PhotoByID(id int) *Photo {
	if idx == nil {
		return nil
	}
	return idx.byID[id]
}

// PhotosByKeyword returns photos matching kwd in either the main keyword index or
// (when includeSubkeywords is true) the sub-keyword index, sorted by ascending Rank.
func (idx *Index) PhotosByKeyword(kwd string, includeSubkeywords bool) []*Photo {
	main := idx.byKeyword[kwd]
	if !includeSubkeywords {
		return main
	}
	sub := idx.bySubKeyword[kwd]
	if len(sub) == 0 {
		return main
	}
	if len(main) == 0 {
		return sub
	}
	// Both main and sub are ordered by ascending Rank; merge them in Rank order.
	return unionSortedPhotos(main, sub)
}

// MatchingKeywords returns all indexed keywords (and sub-keywords if includeSubkeywords is true)
// that contain pat as a substring, sorted by length ascending (ties broken lexicographically).
func (idx *Index) MatchingKeywords(pat string, includeSubkeywords bool) []string {
	set := make(map[string]struct{})
	for kwd, photos := range idx.byKeyword {
		if len(photos) > 0 && strings.Contains(kwd, pat) {
			set[kwd] = struct{}{}
		}
	}
	if includeSubkeywords {
		for kwd, photos := range idx.bySubKeyword {
			if len(photos) > 0 && strings.Contains(kwd, pat) {
				set[kwd] = struct{}{}
			}
		}
	}
	matches := make([]string, 0, len(set))
	for kwd := range set {
		matches = append(matches, kwd)
	}
	sort.Slice(matches, func(i, j int) bool {
		if len(matches[i]) != len(matches[j]) {
			return len(matches[i]) < len(matches[j])
		}
		return matches[i] < matches[j]
	})
	return matches
}
