package model

import (
  "encoding/base64"
  "net/http"
  "strings"
  "time"
)

var header = "<rss version='2.0' xmlns:content='http://purl.org/rss/1.0/modules/content/' xmlns:dc='http://purl.org/dc/elements/1.1/' xmlns:media='http://search.yahoo.com/mrss/' xmlns:atom='http://www.w3.org/2005/Atom' xmlns:georss='http://www.georss.org/georss'><channel><title>Les Photos chez Toutizes</title><link>http://toutizes.com/</link><description>Photos chez Toutizes, rafraichies regulierement.</description><language>fr-fr</language><atom:link rel='hub' href='http://toutizes.superfeedr.com/'/><atom:link rel='self' href='http://toutizes.com/db/f'/>\n"

var footer = "</channel></rss>\n"


func(dir *Directory) Title() string {
  return "Album " + dir.RelPat()
}

func(dir *Directory) Summary() string {
  return "dir dir dir bla bla bla"
}

func min(a int, b int) int {
  if a < b { return a } else { return b }
}

func(img *Image) Summary() string {
  ikwds := img.Keywords()
  return strings.Join(ikwds[0: min(len(ikwds), 5)], ", ")
}

// This is the same b64 encoding as expected in the pic.html hash element.
// See hash.js.
func album_url(url_root string, dir *Directory) string {
  q := "q\x00album:" + dir.RelPat() + "\x00mode\x00I"
  data := []byte(q)
  b64q := base64.StdEncoding.EncodeToString(data)
  return url_root + "pic.html#" + b64q
}

func (dir *Directory) RSSItem(url_root string) string {
  var strs []string
  strs = append(strs, "<item>")
  strs = append(strs, "<title>", dir.Title(), "</title>")
  dir_url := album_url(url_root, dir)
  strs = append(strs, "<link>", dir_url, "</link>")
  strs = append(strs, "<guid>", dir_url, "</guid>")
  strs = append(strs,
    "<pubDate>", dir.Time().Format(time.RFC1123), "</pubDate>")
  strs = append(strs, "<dc:creator>Toutizes</dc:creator>")
  strs = append(strs, "<title><![CDATA[", dir.Summary(), "]]></title>")
  strs = append(strs, "<description><![CDATA[")
  imgs := dir.Images()
  if len(imgs) > 4 {
    imgs = imgs[:4]
  }
  for _, img := range imgs {
    // strs = append(strs, "<p>", img.Summary(), "</p>")
    strs = append(strs, "<figure>")
    strs = append(strs, "<img src='" +  url_root + "midi/", dir.RelPat(),
      "/", img.Name(), "'/>")
    strs = append(strs, "<figcaption>", img.Summary(), "</figcaption>")
    strs = append(strs, "</figure><p></p>\n")
  }
  strs = append(strs, "]]></description>")
  strs = append(strs, "</item>\n")
  return strings.Join(strs, "")
}

func HandleFeed(w http.ResponseWriter, r *http.Request, db *Database) {
  var max_dirs = 5;
  w.Header().Set("Content-Type", "application/atom+xml")
  w.Write([]byte(header))
  dirs := db.Directories()
  url_root := "http://" + r.Header["X-Forwarded-Host"][0] + "/db/"
  for i := len(dirs) - 1; i >= 0; i-- {
    if max_dirs == 0 {
      break
    }
    if len(dirs[i].Images()) > 0 {
      w.Write([]byte(dirs[i].RSSItem(url_root)))
      max_dirs--
    }
  }
  w.Write([]byte(footer))
}

