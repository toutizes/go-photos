package model

import (
  "encoding/json"
  "errors"
  "fmt"
  "net/http"
  "strconv"
  "log"
)

type ImageResults struct {
  Images []JsonImage
}

type DirectoryResults struct {
  Directories []JsonDirectory
}

type StringResults struct {
  Message string
}

type KeywordResults struct {
  Keywords []KeywordCount `json:"keywords"`
}

func queryImages(q string, db *Database) []*Image {
  if len(q) > 0 {
    qry := ParseQuery(q, db)
    imgs := make([]*Image, 0)
    for img := range qry {
      imgs = append(imgs, img)
    }
    return imgs
  } else {
    return nil
  }
}

func returnImages(w http.ResponseWriter, imgs []*Image) {
  enc := json.NewEncoder(w)
  res := make([]JsonImage, len(imgs))
  for i, img := range imgs {
    img.Json(&res[i])
  }
  enc.Encode(&res)
}

func returnDirectories(w http.ResponseWriter, imgs []*Image) {
  dirs := make(map[*Directory]bool)
  for _, img := range imgs {
    dirs[img.Directory()] = true
  }
  enc := json.NewEncoder(w)
  res := make([]JsonDirectory, len(dirs))
  i := 0
  for dir, _ := range dirs {
    dir.Json(&res[i])
    i += 1
  }
  enc.Encode(&res)
}

func HandleQuery(w http.ResponseWriter, r *http.Request, db *Database) {
  q := r.FormValue("q")
  kind := r.FormValue("kind")
  
  // Get user email from context
  userEmail := r.Context().Value("userEmail").(string)
  log.Printf("Query from %s: %q (kind: %s)", userEmail, q, kind)
  
  imgs := queryImages(q, db)
  switch {
  case kind == "album":
    returnDirectories(w, imgs)
  default:
    returnImages(w, imgs)
  }
}

func parseInt(r *http.Request, s string, err error) (int, bool, error) {
  if err != nil {
    return 0, false, err
  }
  vals, ok := r.Form[s]
  if !ok {
    return 0, false, nil
  }
  i, e := strconv.ParseInt(vals[0], 10, 32)
  return int(i), true, e
}

func parseFloat(r *http.Request, s string, err error) (float32, bool, error) {
  if err != nil {
    return 0.0, false, err
  }
  vals, ok := r.Form[s]
  if !ok {
    return 0, false, nil
  }
  f, e := strconv.ParseFloat(vals[0], 32)
  return float32(f), true, e
}

func HandleSet(w http.ResponseWriter, r *http.Request, db *Database) {
  var res StringResults
  err := r.ParseForm()
  id, has_id, err := parseInt(r, "id", err)
  dx, has_dx, err := parseFloat(r, "dx", err)
  dy, has_dy, err := parseFloat(r, "dy", err)
  var image *Image
  if err == nil {
    if !has_id {
      err = errors.New("Missing image id")
    }
  }
  if err == nil {
    image = db.Indexer().Image(id)
    if image == nil {
      err = errors.New(fmt.Sprintf("Unknown image id: %d", id))
    }
  }
  if err == nil {
    if has_dx && has_dy {
      // add and update the stereo info.
      if image.stereo == nil {
        image.stereo = new(Stereo)
      }
      image.stereo.Dx = dx
      image.stereo.Dy = dy
    } else if !has_dx && !has_dy {
      // delete the stereo info.
      image.stereo = nil
    } else {
      err = errors.New("Pass both dx and dy or none of them")
    }
  }
  if err == nil {
    err = db.SaveDirectory(image.Directory())
  }
  if err == nil {
    res.Message = "ok"
  } else {
    res.Message = err.Error()
  }
  enc := json.NewEncoder(w)
  enc.Encode(&res)
}

func HandleRecentKeywords(w http.ResponseWriter, r *http.Request, db *Database) {
  // Get user email from context
  userEmail := r.Context().Value("userEmail").(string)
  log.Printf("Recent keywords request from %s", userEmail)
  
  keywords := db.GetRecentActiveKeywords()
  
  enc := json.NewEncoder(w)
  result := KeywordResults{Keywords: keywords}
  enc.Encode(&result)
}

