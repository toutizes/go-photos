package model

import (
  "archive/zip"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "net/url"
  "path"
  "strconv"
  "strings"
)

func HandleDownload(w http.ResponseWriter, r *http.Request, db *Database,
                    vals url.Values) {
  q, ok := vals["q"]
  if !ok {
    log.Printf("Missing q: %v\n", vals)
    return
  }
  s, ok := vals["s"]
  if !ok {
    log.Printf("Missing s: %v\n", vals)
    return
  }
  imgs := queryImages(q[0], db)
  if imgs == nil {
    log.Printf("No images for query: %v\n", vals)
    return
  } 
  if len(imgs) >= 100 {
    imgs = imgs[:100]
  }
  header := w.Header()
  header.Set("Content-Disposition", `attachment; filename="images.zip"`)
  z := zip.NewWriter(w)
  for _, img := range imgs {
    f, err := z.Create(path.Join(img.Directory().RelPat(), img.Name()))
    var b []byte
    if err == nil {
      if s[0] == "M" {
        b, err = ioutil.ReadFile(db.MidiPath(img.Id))
      } else {
        b, err = ioutil.ReadFile(db.OrigPath(img.Id))
      }
    }
    if err == nil {
      _, err = f.Write(b)
    }
    if err != nil {
      log.Printf("Error %s", err.Error())
      break
    }
  }
  z.Close()
}

func HandleReload(w http.ResponseWriter, r *http.Request, odb *Database) {
  ndb := DatabaseToReload(odb)
  err := ndb.Load(true, true, false)
  if err != nil {
    fmt.Fprintf(w, err.Error())
  } else {
    odb.Swap(ndb)
    fmt.Fprintf(w, "Reloaded")
  }
}

func HandleCommands(w http.ResponseWriter, r *http.Request, db *Database) {
  vals, err := url.ParseQuery(r.URL.RawQuery)
  if err != nil {
    log.Printf("%v\n", err.Error())
    return
  }
  comm, ok := vals["command"]
  if !ok {
    log.Printf("Missing command: %v\n", vals)
    return
  }
  switch comm[0] {
  case "download": HandleDownload(w, r, db, vals)
  case "reload": HandleReload(w, r, db)
  default: log.Printf("Unknown command: %v\n", vals)
  }
}

func HandleFile(w http.ResponseWriter, r *http.Request, prefix string,
                root string) {
  vals, err := url.ParseQuery(r.URL.RawQuery)
  if err != nil {
    log.Printf("%v\n", err.Error())
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  dl, has_dl := vals["dl"]
  var rel_path = r.URL.Path
  if (!strings.HasPrefix(rel_path, prefix)) {
    w.WriteHeader(http.StatusForbidden)
    return
  }
  rel_path = strings.TrimPrefix(rel_path, prefix)
  var abs_path = path.Join(root, rel_path)
  if (has_dl && dl[0] == "true") {
    w.Header().Set("Content-Disposition", 
      "attachment; filename=" + strconv.Quote(path.Base(rel_path)))
    w.Header().Set("Content-Type", "image/jpeg")
  }
  http.ServeFile(w, r, abs_path)
}
