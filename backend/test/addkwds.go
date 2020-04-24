package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

import (
	"toutizes.com/go-photwo/backend/model"
)

var orig_root = flag.String("orig_root", "", "path to the orig files")
var index_root = flag.String("index_root", "", "path to the index files")
var root = flag.String("root", "/tmp/db", "root or the index, mini, etc.")
var old_index_root = flag.String("old_index_root", "", "path to the old index files")

func database() *model.Database {
	if *orig_root == "" {
		*orig_root = path.Join(*root, "orig")
	}
	if *old_index_root == "" {
		*old_index_root = path.Join(*root, "oldex")
	}
	if *index_root == "" {
		*index_root = path.Join(*root, "index")
	}
	return model.NewDatabase5(*index_root, *orig_root, path.Join(*root, "mini"),
		                        path.Join(*root, "midi"), path.Join(*root, "montage"))
}

func parseKeywords(s string) (keywords []string) {
	if strings.HasPrefix(s, "convert: unknown image property") || len(s) == 0 {
		return nil
	}
	if strings.HasSuffix(s, "\n") {
		s = s[0:len(s) - 1]
	}
	return strings.Split(s, ";")
}

func appendIfNew(kwds []string, k string, present map[string]bool) []string {
	_, p := present[k]
	if !p {
		present[k] = true
		return append(kwds, k)
	} else {
		return kwds
	}
}

func keywordsToSet(img *model.Image, file_kwds []string, 
	                 oldex map[string][]string) ([]string, bool) {
	// log.Printf("%s file: %s\n", img.Name(), strings.Join(file_kwds, ", "))
	// log.Printf("%s db  : %s\n", img.Name(), strings.Join(img.Keywords(), ", "))
	present := make(map[string]bool)
	var kwds []string
	var k string
	for _, k = range img.Keywords() {
		kwds = appendIfNew(kwds, k, present)
	}
	for _, k = range file_kwds {
		kwds = appendIfNew(kwds, k, present)
	}
	old_kwds, has_old := oldex[img.Name()]
	if has_old {
		for _, k = range old_kwds {
			kwds = appendIfNew(kwds, k, present)
		}
	}
	return kwds, len(kwds) != len(file_kwds)
}

func setKeywords(img *model.Image, path string, kwds []string) error {
	// charset issues?
	// log.Printf("%s: kwds: %d\n", path, len(kwds))
	tmp := "/tmp/k.txt"
	text := "2#25#Keyword=" + strings.Join(kwds, ";")
  err := ioutil.WriteFile(tmp, []byte(text), 0777)
  if err != nil {
    return err
  }
	cmd := exec.Command(
		*model.BinRoot + "mogrify",  "-comment",  "JFIFComment",  "-profile",  
		"8BIMTEXT:" + tmp, path)
	out, err := cmd.Output()
	if err != nil {
		log.Printf("%v: %s\n", cmd.Args, out)
		return err
	}
	os.Chtimes(path, img.FileTime(), img.FileTime())
	return nil
}

func toUtf8(iso8859_1_buf []byte) string {
	buf := make([]rune, len(iso8859_1_buf))
	for i, b := range iso8859_1_buf {
		buf[i] = rune(b)
	}
	return string(buf)
}

func collapseQuoted(kwds []string) []string {
	var qwds []string
	quoted := false
	q := ""
	for _, k := range kwds {
		if quoted {
			if strings.HasSuffix(k, "\"") {
				qwds = append(qwds, q + " " + k[:len(k) - 1])
				quoted = false
				q = ""
			} else {
				q += " " + k
			}
		} else if strings.HasPrefix(k, "\"") {
			quoted = true
			q = k[1:]
		} else if k == "" {
			continue
		} else {
			qwds = append(qwds, k)
		}
	}
	if q != "" {
		log.Printf("Bad quotes: %v\n", kwds)
	}
	return qwds
}

func loadOldIndex(dir string) (map[string][]string, error) {
	contents, err := ioutil.ReadFile(path.Join(dir, "index.txt"))
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(toUtf8(contents)), "\n")
	result := make(map[string][]string, len(lines))
	for _, line := range lines {
		if strings.HasSuffix(line, ": ") {
			continue
		}
		splits := strings.Split(line, " ")
		if len(splits) < 4 {
			continue
		}
		result[splits[0]] = collapseQuoted(splits[3:])
	}
	return result, nil
}

func main() {
	flag.Parse()

	cpu := runtime.NumCPU()
	runtime.GOMAXPROCS(cpu / 2)

	db := database()
	db.Load(false /* update_disk */, 
		      false /* minify */, 
		      false /* force reload*/)
	var to_clean []string
	for _, dir := range db.Directories() {
		dir_path := db.FullOrigPath(dir.RelPat())
		oldex, err := loadOldIndex(path.Join(*old_index_root, dir.RelPat()))
		if err != nil {
			log.Printf("Old index no load: %s\n", dir.RelPat())
			continue;
		}
		has_changes := false
		for _, img := range dir.Images() {
			img_path := path.Join(dir_path, img.Name())
			cmd := exec.Command(
				*model.BinRoot + "convert", img_path, "-format", "%[IPTC:2:25]", "info:")
			out, err := cmd.Output()
			var file_kwds []string
			if err == nil {
				file_kwds = parseKeywords(string(out))
			}
			if len(file_kwds) > 0 {
				// Skip images that already have file keywords.
				continue
			}
			kwds, must_set := keywordsToSet(img, file_kwds, oldex)
			if must_set {
				// log.Printf("To set %s: %s\n", img.Name(), strings.Join(kwds, ", "))
				setKeywords(img, img_path, kwds)
				has_changes = true
			} else {
				// log.Printf("Good   %s\n", img.Name())
			}
		}
		if has_changes {
			to_clean = append(to_clean, "\"" + dir.RelPat() + "/index.pbin\"")
			if len(to_clean) > 10 {
				log.Printf("rm %s\n", strings.Join(to_clean, " "))
				to_clean = make([]string, 0)
			}
		}
	}
	log.Printf("rm %s\n", strings.Join(to_clean, " "))
}
