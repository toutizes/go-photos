package model

import (
	"mime"
  "archive/zip"
  "fmt"
  "io"
  "log"
  "net/http"
  "net/url"
  "os"
  "path"
  "path/filepath"
  "strconv"
  "strings"
  ioutil "io/ioutil"
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

// Add function to get content type
func GetContentType(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".svg":
		return "image/svg+xml"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".ico":
		return "image/x-icon"
	default:
		if ct := mime.TypeByExtension(ext); ct != "" {
			return ct
		}
		return "application/octet-stream"
	}
}


// func HandleFile(w http.ResponseWriter, r *http.Request, prefix string,
//                 root string) {
//   vals, err := url.ParseQuery(r.URL.RawQuery)
//   if err != nil {
//     log.Printf("%v\n", err.Error())
//     w.WriteHeader(http.StatusBadRequest)
//     return
//   }
//   dl, has_dl := vals["dl"]
//   var rel_path = r.URL.Path
//   if (!strings.HasPrefix(rel_path, prefix)) {
//     w.WriteHeader(http.StatusForbidden)
//     return
//   }
//   rel_path = strings.TrimPrefix(rel_path, prefix)
//   var abs_path = path.Join(root, rel_path)
//   if (has_dl && dl[0] == "true") {
//     w.Header().Set("Content-Disposition", 
//       "attachment; filename=" + strconv.Quote(path.Base(rel_path)))
//     w.Header().Set("Content-Type", "image/jpeg")
//   }
//   // For google auth.
//   // https://developers.google.com/identity/gsi/web/guides/get-google-api-clientid#cross_origin_opener_policy
//   w.Header().Set("Cross-Origin-Opener-Policy", "same-origin-allow-popups")
//   http.ServeFile(w, r, abs_path)
// }
func HandleFile(w http.ResponseWriter, r *http.Request, prefix string, root string) {
	// Strip the prefix from the path
	path := r.URL.Path[len(prefix):]
	fullPath := filepath.Join(root, path)

	// Prevent directory traversal
	if !strings.HasPrefix(filepath.Clean(fullPath), root) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Open the file
	file, err := os.Open(fullPath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Get file info for Content-Length
	info, err := file.Stat()
	if err != nil {
		http.Error(w, "Unable to get file info", http.StatusInternalServerError)
		return
	}

	// Set content type based on file extension
	contentType := GetContentType(path)
	w.Header().Set("Content-Type", contentType)
	
	// Set caching headers
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))

	// Use larger buffer for copying
	buf := make([]byte, 32*1024) // 32KB buffer
	_, err = io.CopyBuffer(w, file, buf)
	if err != nil {
		log.Printf("Error sending file %s: %v", path, err)
	}
}
