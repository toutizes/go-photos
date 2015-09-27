package model

import (
  "regexp"
  "strings"
  "strconv"
  "time"
  "unicode"
)

type Query <- chan *Image

func KeywordQuery(db *Database, kwd string) Query {
  q := make(chan *Image)

  go func(q chan *Image, idx *Indexer) {
    defer close(q)
    for _, img := range idx.Images(kwd) {
      q <- img
    }
  } (q, db.Indexer())

  return q
}

func DirectoriesQuery(db *Database) Query {
  q := make(chan *Image)

  go func(q chan *Image, db *Database) {
    defer close(q)
    for _, dir := range db.Directories() {
      if len(dir.Images()) > 0 {
        q <- dir.Images()[0]
      }
    }
  } (q, db)

  return q
}

func DirectoryByNameQuery(db *Database, name string) Query {
  q := make(chan *Image)

  go func(q chan *Image, db *Database, name string) {
    defer close(q)
    for _, dir := range db.Directories() {
      if dir.RelPat() == name {
        for _, img := range dir.Images() {
          q <- img
        }
        return;
      }
    }
  }(q, db, name)

  return q
}

func andFill(qs []Query) []*Image {
  imgs := make([]*Image, len(qs))
  var ok bool
  for i, q := range qs {
    imgs[i], ok = <-q
    if !ok {
      return nil
    }
  }
  return imgs
}

func andAdvance(qs []Query, imgs []*Image, img0 *Image) (*Image, bool) {
  same := true
  var ok bool
  for i, q := range qs {
    for ; imgs[i].Id < img0.Id; {
      imgs[i], ok = <-q
      if !ok { return nil, false }
    }
    if imgs[i] != img0 {
      same = false
    }
  }
  if same {
    return img0, true
  } else {
    return nil, true
  }
}

func AndQuery(qs []Query) Query {
  if len(qs) == 0 {
    return nil
  }
  if len(qs) == 1 {
    return qs[0]
  }
  qq := make(chan *Image)

  go func(qq chan *Image) {
    defer close(qq)
    q0 := qs[0]
    qs := qs[1:]
    imgs := andFill(qs)
    if imgs == nil {
      return
    }
    for img0 := range q0 {
      img, ok := andAdvance(qs, imgs, img0)
      if !ok { return }
      if img != nil {
        qq <- img
      }
    }
  }(qq)

  return qq
}


func orFill(qs []Query) []*Image {
  has_any := false
  imgs := make([]*Image, len(qs))
  var ok bool
  for i, q := range qs {
    imgs[i], ok = <-q
    if ok {
      has_any = true
    }
  }
  if has_any {
    return imgs
  } else {
    return nil
  }
}

func orAdvance(qs []Query, imgs []*Image) *Image {
  var min_img *Image
  for _, img := range imgs {
    if img != nil  && (min_img == nil || img.Id < min_img.Id) {
      min_img = img
    }
  }
  if min_img == nil {
    return nil
  }
  for i, img := range imgs {
    if img == min_img {
      imgs[i] = <-qs[i]
    }
  }
  return min_img
}

func OrQuery(qs []Query) Query {
  if len(qs) == 0 {
    return nil
  }
  if len(qs) == 1 {
    return qs[0]
  }
  qq := make(chan *Image)

  go func(qq chan *Image) {
    defer close(qq)
    imgs := orFill(qs)
    if imgs == nil {
      return
    }
    for {
      img := orAdvance(qs, imgs)
      if img == nil {
        return
      }
      qq <- img
    }
  }(qq)

  return qq
}

func FilteredQuery(db *Database, filter func(*Image) bool) Query {
  q := make(chan *Image)

  go func(q chan *Image) {
    defer close(q)
    for _, dir := range db.Directories() {
      for _, img := range dir.Images() {
        if filter(img) {
          q <- img
        }
      }
    }
  }(q)

  return q
}

func TimeRangeQuery(db *Database, start time.Time, end time.Time) Query {
  filter := func(img *Image) bool {
    return img.ItemTime().After(start) && img.ItemTime().Before(end)
  }
  return FilteredQuery(db, filter)
}

func YearQuery(db *Database, year string) Query {
  tim, err := time.Parse("2006 MST", year + " PST")
  if err == nil {
    return TimeRangeQuery(db, tim, tim.AddDate(1, 0, 0))
  } else {
    return nil
  }
}

func MonthQuery(db *Database, year_month string) Query {
  tim, err := time.Parse("2006-01 MST", year_month + " PST")
  if err == nil {
    return TimeRangeQuery(db, tim, tim.AddDate(0, 1, 0))
  } else {
    return nil
  }
}

func DayQuery(db *Database, year_month_day string) Query {
  tim, err := time.Parse("2006-01-02 MST", year_month_day + " PST")
  if err == nil {
    return TimeRangeQuery(db, tim, tim.AddDate(0, 0, 1))
  } else {
    return nil
  }
}

func KeywordCountQuery(db *Database, count string) Query {
  cnt, err := strconv.Atoi(count)
  if err == nil {
    filter := func(img *Image) bool {
      return len(img.Keywords()) == cnt
    }
    return FilteredQuery(db, filter)
  } else {
    return nil
  }
}

func StereoQuery(db *Database) Query {
  filter := func(img *Image) bool {
    return img.stereo != nil
  }
  return FilteredQuery(db, filter)
}


// Query parser
const (
  in_space = 0
  in_token = 1
  in_string = 2
)

func tokenize(s string) []string {
  state := in_space
  var tokens []string
  from := 0
  for i, r := range s {
    switch {
    case unicode.IsSpace(r):
      switch state {
      case in_space:
      case in_token:
        tokens = append(tokens, s[from:i])
        state = in_space
      case in_string:
      }
    case r == '"':
      switch state {
      case in_space:
        state = in_string
        from = i + 1
      case in_token:
        tokens = append(tokens, s[from:i])
        from = i + 1
        state = in_string
      case in_string:
        tokens = append(tokens, s[from:i])
        state = in_space
      }
    default:
      switch state {
      case in_space:
        state = in_token
        from = i
      case in_token:
      case in_string:
        }
    }
  }
  if state != in_space {
    tokens = append(tokens, s[from:])
  }
  return tokens
}

const (
  dir_query = ":albums"
  year_re = "^[12][0-9][0-9][0-9]$"
  month_re = "^[12][0-9][0-9][0-9]-[0-9][0-9]$"
  day_re = "^[12][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]$"
)

func matches(pattern string, token string) bool {
  match, _ := regexp.MatchString(pattern, token)
  return match
}

func ParseQuery(s string, db *Database) Query {
  tokens := tokenize(s)
  qs := make([]Query, len(tokens))
  for i, t := range tokens {
    lower_t := strings.ToLower(t)
    var alt_q Query
    var time_q Query
    switch {
    case strings.HasPrefix(lower_t, "count:"):
      alt_q = KeywordCountQuery(db, t[len("count:"):])
    case strings.HasPrefix(lower_t, "stereo:"):
      alt_q = StereoQuery(db)
    case strings.HasPrefix(lower_t, "album:"):
      alt_q = DirectoryByNameQuery(db, t[len("album:"):])
    case t == "albums:":
      alt_q = DirectoriesQuery(db)
    case matches(year_re, t):
      time_q = YearQuery(db, t)
    case matches (month_re, t):
      time_q = MonthQuery(db, t)
    case matches (day_re, t):
      time_q = DayQuery(db, t)
    }
    if alt_q != nil {
      qs[i] = alt_q
    } else {
      kwd_q := KeywordQuery(db, DropAccents(lower_t, nil))
      if time_q != nil {
        qs[i] = OrQuery([]Query{kwd_q, time_q})
      } else {
        qs[i] = kwd_q
      }
    }
  }
  return AndQuery(qs)
}