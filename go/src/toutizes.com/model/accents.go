package model

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
  bs := make([]byte, 0, len(s))
  for _, r := range s {
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
    res = s
  }
  if cache != nil {
    cache[s] = res
  }
  return res
}
