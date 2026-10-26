package catalog

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	yearRE      = regexp.MustCompile(`^[12][0-9][0-9][0-9]$`)
	monthRE     = regexp.MustCompile(`^[12][0-9][0-9][0-9]-[0-9][0-9]$`)
	dayRE       = regexp.MustCompile(`^[12][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]$`)
	yearRangeRE = regexp.MustCompile(`^[12][0-9][0-9][0-9]--[12][0-9][0-9][0-9]$`)
	monthDayRE  = regexp.MustCompile(`^[0-1]?[0-9]-[0-3]?[0-9]$`)
)

func unionSortedPhotos(a, b []*Photo) []*Photo {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	res := make([]*Photo, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i].Rank < b[j].Rank {
			res = append(res, a[i])
			i++
		} else if b[j].Rank < a[i].Rank {
			res = append(res, b[j])
			j++
		} else {
			res = append(res, a[i])
			i++
			j++
		}
	}
	if i < len(a) {
		res = append(res, a[i:]...)
	}
	if j < len(b) {
		res = append(res, b[j:]...)
	}
	return res
}

func unionManySortedPhotos(lists [][]*Photo) []*Photo {
	var acc []*Photo
	for _, list := range lists {
		acc = unionSortedPhotos(acc, list)
	}
	return acc
}

func intersectSortedPhotos(a, b []*Photo) []*Photo {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}
	maxLen := len(a)
	if len(b) < maxLen {
		maxLen = len(b)
	}
	res := make([]*Photo, 0, maxLen)
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i].Rank < b[j].Rank {
			i++
		} else if b[j].Rank < a[i].Rank {
			j++
		} else {
			res = append(res, a[i])
			i++
			j++
		}
	}
	return res
}

func intersectManySortedPhotos(lists [][]*Photo) []*Photo {
	if len(lists) == 0 {
		return nil
	}
	acc := lists[0]
	for i := 1; i < len(lists) && len(acc) > 0; i++ {
		acc = intersectSortedPhotos(acc, lists[i])
	}
	return acc
}

func (c *Catalog) filterPhotos(pred func(*Photo) bool) []*Photo {
	var out []*Photo
	for _, album := range c.Albums {
		for _, p := range album.Photos {
			if pred(p) {
				out = append(out, p)
			}
		}
	}
	return out
}

func (c *Catalog) timeRangeQuery(start, end time.Time) []*Photo {
	return c.filterPhotos(func(p *Photo) bool {
		return p.ItemTime.After(start) && p.ItemTime.Before(end)
	})
}

func (c *Catalog) yearQuery(year string) []*Photo {
	t, err := time.Parse("2006 MST", year+" PST")
	if err != nil {
		return nil
	}
	return c.timeRangeQuery(t, t.AddDate(1, 0, 0))
}

func (c *Catalog) yearRangeQuery(yearRange string) []*Photo {
	tFrom, err1 := time.Parse("2006 MST", yearRange[0:4]+" PST")
	tTo, err2 := time.Parse("2006 MST", yearRange[6:10]+" PST")
	if err1 != nil || err2 != nil {
		return nil
	}
	return c.timeRangeQuery(tFrom, tTo.AddDate(1, 0, 0))
}

func (c *Catalog) monthQuery(yearMonth string) []*Photo {
	t, err := time.Parse("2006-01 MST", yearMonth+" PST")
	if err != nil {
		return nil
	}
	return c.timeRangeQuery(t, t.AddDate(0, 1, 0))
}

func (c *Catalog) dayQuery(yearMonthDay string) []*Photo {
	t, err := time.Parse("2006-01-02 MST", yearMonthDay+" PST")
	if err != nil {
		return nil
	}
	return c.timeRangeQuery(t, t.AddDate(0, 0, 1))
}

func parseMonthDay(s string) ([2]int, error) {
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return [2]int{}, fmt.Errorf("invalid MM-DD: %s", s)
	}
	month, err := strconv.Atoi(parts[0])
	if err != nil {
		return [2]int{}, err
	}
	day, err := strconv.Atoi(parts[1])
	if err != nil {
		return [2]int{}, err
	}
	return [2]int{month, day}, nil
}

func (c *Catalog) monthDayQuery(monthDay string) []*Photo {
	md, err := parseMonthDay(monthDay)
	if err != nil {
		return c.Index.PhotosByKeyword(monthDay, true)
	}
	return c.filterPhotos(func(p *Photo) bool {
		return p.ItemTime.Month() == time.Month(md[0]) && p.ItemTime.Day() == md[1]
	})
}

func (c *Catalog) keywordCountQuery(countStr string) []*Photo {
	cnt, err := strconv.Atoi(countStr)
	if err != nil {
		return nil
	}
	return c.filterPhotos(func(p *Photo) bool {
		return len(p.Keywords) == cnt
	})
}

func (c *Catalog) stereoQuery() []*Photo {
	return c.filterPhotos(func(p *Photo) bool {
		return p.Stereo != nil
	})
}

func (c *Catalog) directoriesQuery() []*Photo {
	out := make([]*Photo, 0, len(c.Albums))
	for _, album := range c.Albums {
		if len(album.Photos) > 0 {
			out = append(out, album.Photos[0])
		}
	}
	return out
}

func (c *Catalog) directoryByNameQuery(name string) []*Photo {
	for _, album := range c.Albums {
		if album.RelPat == name {
			return album.Photos
		}
	}
	return nil
}

func (c *Catalog) directoryBySubnameQuery(name string) []*Photo {
	sub := strings.ToLower(name)
	var out []*Photo
	for _, album := range c.Albums {
		if strings.Contains(strings.ToLower(album.RelPat), sub) {
			out = append(out, album.Photos...)
		}
	}
	return out
}

func (c *Catalog) personQuery(p *Person) []*Photo {
	lists := make([][]*Photo, 0, 2*len(p.Names))
	for _, n := range p.Names {
		lists = append(lists, c.Index.PhotosByKeyword(n, false))
		words := strings.Fields(n)
		wordLists := make([][]*Photo, len(words))
		for i, w := range words {
			wordLists[i] = c.Index.PhotosByKeyword(w, false)
		}
		lists = append(lists, intersectManySortedPhotos(wordLists))
	}
	return unionManySortedPhotos(lists)
}

func (c *Catalog) keywordSynonymsQuery(kwd string) []*Photo {
	if p := c.Synonyms.Lookup(kwd); p != nil {
		return c.personQuery(p)
	}
	return c.Index.PhotosByKeyword(kwd, true)
}

func (c *Catalog) keywordMatchQuery(pat string) []*Photo {
	kwds := c.Index.MatchingKeywords(pat, true)
	n := len(kwds)
	if n == 0 {
		return nil
	}
	if max := 10 * len(pat); n > max {
		n = max
	}
	lists := make([][]*Photo, n)
	for i := 0; i < n; i++ {
		lists[i] = c.Index.PhotosByKeyword(kwds[i], true)
	}
	return unionManySortedPhotos(lists)
}

// QueryPhotos parses a Lightroom-style comma-separated query string and returns matching photos
// deduplicated by photo ID.
func (c *Catalog) QueryPhotos(q string) []*Photo {
	if c == nil || len(q) == 0 {
		return nil
	}
	rawTokens := strings.Split(q, ",")
	tokens := make([]string, 0, len(rawTokens))
	for _, tok := range rawTokens {
		tok = strings.TrimSpace(tok)
		if tok != "" {
			tokens = append(tokens, tok)
		}
	}
	if len(tokens) == 0 {
		return nil
	}

	clauseResults := make([][]*Photo, len(tokens))
	for i, t := range tokens {
		// Strip optional surrounding quotes (e.g. from legacy bookmarked queries)
		if len(t) >= 2 && t[0] == '"' && t[len(t)-1] == '"' {
			t = t[1 : len(t)-1]
		}
		lowerT := strings.ToLower(t)
		hasSpaces := strings.Contains(t, " ")

		switch {
		case strings.HasPrefix(lowerT, "count:"):
			clauseResults[i] = c.keywordCountQuery(t[len("count:"):])
		case strings.HasPrefix(lowerT, "stereo:"):
			clauseResults[i] = c.stereoQuery()
		case strings.HasPrefix(lowerT, "album:"):
			clauseResults[i] = c.directoryByNameQuery(t[len("album:"):])
		case strings.HasPrefix(lowerT, "in:"):
			clauseResults[i] = c.directoryBySubnameQuery(t[len("in:"):])
		case strings.HasPrefix(lowerT, "titre:"):
			clauseResults[i] = c.directoryBySubnameQuery(t[len("titre:"):])
		case t == "albums:":
			clauseResults[i] = c.directoriesQuery()
		case yearRE.MatchString(t):
			clauseResults[i] = unionSortedPhotos(c.yearQuery(t), c.Index.PhotosByKeyword(t, true))
		case monthRE.MatchString(t):
			clauseResults[i] = unionSortedPhotos(c.monthQuery(t), c.Index.PhotosByKeyword(t, true))
		case dayRE.MatchString(t):
			clauseResults[i] = unionSortedPhotos(c.dayQuery(t), c.Index.PhotosByKeyword(t, true))
		case monthDayRE.MatchString(t):
			clauseResults[i] = c.monthDayQuery(t)
		case yearRangeRE.MatchString(t):
			clauseResults[i] = c.yearRangeQuery(t)
		case hasSpaces:
			clauseResults[i] = c.keywordSynonymsQuery(lowerT)
		case c.Synonyms.IsName(lowerT):
			clauseResults[i] = unionSortedPhotos(c.keywordSynonymsQuery(lowerT), c.keywordMatchQuery(lowerT))
		default:
			clauseResults[i] = c.keywordMatchQuery(lowerT)
		}
	}

	merged := intersectManySortedPhotos(clauseResults)
	if len(merged) == 0 {
		return nil
	}
	seen := make(map[int]bool, len(merged))
	dedup := make([]*Photo, 0, len(merged))
	for _, p := range merged {
		if !seen[p.Id] {
			seen[p.Id] = true
			dedup = append(dedup, p)
		}
	}
	return dedup
}
