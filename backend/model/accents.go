package model

import (
  "strings"
)

var french_accents = map[rune]byte{
  'à': 'a',
  'â': 'a',
  'é': 'e',
  'è': 'e',
  'ê': 'e',
  'ë': 'e',
  'î': 'i',
  'ï': 'i',
  'ô': 'o',
  'ù': 'u',
  'û': 'u',
  'ü': 'u',
  'ç': 'c',
}

func DropAccents(s string, cache map[string]string) string {
  changed := false
  if cache != nil {
    ss, ok := cache[s]
    if ok {
      return ss
    }
  }
	lower_s := strings.ToLower(s)
	// Save ram by not keeping the lowercase version if the original was
	// already lowercase.
	if lower_s == s {
		lower_s = s
	}
	// Why am I not using 'rune' here??  I do not remember.  I think I should.
  bs := make([]byte, 0, len(lower_s))
  for _, r := range lower_s {
    rr, ok := french_accents[r]
    if ok {
      changed = true
      bs = append(bs, rr)
    } else {
      bs = append(bs, byte(r))
    }
  }
  var res string
  if changed {
    res = string(bs)
  } else {
    res = lower_s
  }
  if cache != nil {
		// Always use the original string as the cache key.
    cache[s] = res
  }
  return res
}
