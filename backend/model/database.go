package model

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"time"
)

import "github.com/golang/protobuf/proto"
import "toutizes.com/go-photwo/backend/store"

type Database struct {
	root        string
	indx_root   string
	orig_root   string
	midi_root   string
	mini_root   string
	mont_root   string
	static_root string
	directories []*Directory
	indexer     *Indexer
	file_times  FileTimes
	recentActiveKeywords []KeywordCount
}

func NewDatabase(root string) *Database {
	db := new(Database)
	db.indx_root = path.Join(root, "index")
	db.orig_root = path.Join(root, "originals")
	db.midi_root = path.Join(root, "midi")
	db.mini_root = path.Join(root, "mini")
	db.mont_root = path.Join(root, "montage")
	db.static_root = root
	db.indexer = NewIndexer()
	db.file_times = NewFileTimes()
	return db
}

func DatabaseToReload(odb *Database) *Database {
	db := new(Database)
	db.indx_root = odb.indx_root
	db.orig_root = odb.orig_root
	db.midi_root = odb.midi_root
	db.mini_root = odb.mini_root
	db.mont_root = odb.mont_root
	db.static_root = odb.static_root
	db.indexer = NewIndexer()
	db.file_times = NewFileTimes()
	return db
}

func NewDatabase2(orig_root, root string, static_root string) *Database {
	db := new(Database)
	db.orig_root = orig_root
	db.indx_root = path.Join(root, "index")
	db.midi_root = path.Join(root, "midi")
	db.mini_root = path.Join(root, "mini")
	db.mont_root = path.Join(root, "montage")
	db.static_root = static_root
	db.indexer = NewIndexer()
	db.file_times = NewFileTimes()
	return db
}

func NewDatabase5(index_root, orig_root, mini_root, midi_root, mont_root string) *Database {
	db := new(Database)
	db.indx_root = index_root
	db.orig_root = orig_root
	db.midi_root = mini_root
	db.mini_root = midi_root
	db.mont_root = mont_root
	db.static_root = index_root
	db.indexer = NewIndexer()
	db.file_times = NewFileTimes()
	return db
}

func (db *Database) Directories() []*Directory { return db.directories }
func (db *Database) Indexer() *Indexer         { return db.indexer }
func (db *Database) FileTimes() FileTimes      { return db.file_times }
func (db *Database) MontagePath() string       { return db.mont_root }
func (db *Database) IndexPath(rel_pat string) string {
	return path.Join(db.indx_root, rel_pat, "index.pbin")
}
func (db *Database) IndexTextPath(rel_pat string) string {
	return path.Join(db.indx_root, rel_pat, "index.pbtxt")
}
func (db *Database) FullMiniPath(rel_pat string) string {
	return path.Join(db.mini_root, rel_pat)
}
func (db *Database) FullMidiPath(rel_pat string) string {
	return path.Join(db.midi_root, rel_pat)
}
func (db *Database) FullOrigPath(rel_pat string) string {
	return path.Join(db.orig_root, rel_pat)
}
func (db *Database) MiniPath(image_id int) string {
	img := db.indexer.Image(image_id)
	if img == nil {
		return ""
	}
	return path.Join(db.mini_root, img.Directory().RelPat(), img.Name())
}
func (db *Database) MidiPath(image_id int) string {
	img := db.indexer.Image(image_id)
	if img == nil {
		return ""
	}
	return path.Join(db.midi_root, img.Directory().RelPat(), img.Name())
}
func (db *Database) OrigPath(image_id int) string {
	img := db.indexer.Image(image_id)
	if img == nil {
		return ""
	}
	return path.Join(db.orig_root, img.Directory().RelPat(), img.Name())
}
func (db *Database) Swap(ndb *Database) {
	// Lets' ignore threading issues.
	db.indexer = ndb.indexer
	db.directories = ndb.directories
}

func (db *Database) SaveDirectory(dir *Directory) (err error) {
	bin_path := db.IndexPath(dir.RelPat())
	txt_path := db.IndexTextPath(dir.RelPat())
	return writeIndex(bin_path, txt_path, dir.ToProto())
}

// Database Loader
// Struct pass to the loader
type loaderLoad struct {
	rel_pat    string    // Directory relative path.
	orgd_mtime time.Time // Mod time of orginals directory.
	err        error
}

// Struct returned by the loader workers.
type loaderResult struct {
	dir       *Directory    // Constructed directory
	rel_pat   string        // Directory relative path.
	orgd_subs []os.FileInfo // FileInfo for sub directories of orgd.
	err       error         // Error, if any
}

func readIndex(idx string, sdir *store.Directory) error {
	buffer, err := ioutil.ReadFile(idx)
	if err == nil {
		// err = proto.UnmarshalText(string(buffer), sdir)
		err = proto.Unmarshal(buffer, sdir)
	}
	return err
}

func writeIndex(bin_path string, txt_path string, sdir *store.Directory) (err error) {
	log.Printf("write %s, %s\n", bin_path, txt_path)
	err = os.MkdirAll(path.Dir(bin_path), 0777)
	if err != nil {
		return
	}
	bin_data, err := proto.Marshal(sdir)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(bin_path, bin_data, 0777)
	if err != nil {
		return
	}
	txt_data := []byte(proto.MarshalTextString(sdir))
	err = ioutil.WriteFile(txt_path, txt_data, 0777)
	return
}

func (db *Database) handleLoad(
	update_disk bool, force_reload bool, lod *loaderLoad) (*loaderResult, error) {
	if lod.err != nil {
		return nil, lod.err
	}
	var sdir store.Directory
	// Ignore missing index, means new directory.
	readIndex(db.IndexPath(lod.rel_pat), &sdir)
	origd := db.FullOrigPath(lod.rel_pat)
	subs, err := ioutil.ReadDir(origd)
	if err != nil {
		return nil, err
	}
	if sdir.DirectoryTimestamp == nil ||
		ProtoToTime(*sdir.DirectoryTimestamp).Before(lod.orgd_mtime) ||
		force_reload {
		err = UpdateDirectory(origd, subs, force_reload, &sdir)
		if err == nil {
			orgd_ts := TimeToProto(lod.orgd_mtime)
			sdir.DirectoryTimestamp = &orgd_ts
			if update_disk {
				err = writeIndex(db.IndexPath(lod.rel_pat), db.IndexTextPath(lod.rel_pat),
					&sdir)
				if err != nil {
					log.Printf("%s: %s\n", lod.rel_pat, err.Error())
					err = nil
				}
			}
		}
	}
	if err != nil {
		return nil, err
	}
	dir := ProtoToDirectory(&sdir, lod.rel_pat)
	return &loaderResult{dir: dir, rel_pat: lod.rel_pat, orgd_subs: subs}, nil
}

// Worker function loading and checking directories.
func (db *Database) loaderWorker(update_disk bool, force_reload bool,
	lod_ch <-chan *loaderLoad,
	res_ch chan<- *loaderResult) {
	for lod := range lod_ch {
		res, err := db.handleLoad(update_disk, force_reload, lod)
		if err == nil {
			res_ch <- res
		} else {
			log.Printf("Worker err: %v\n", err)
			res_ch <- &loaderResult{err: err}
		}
	}
}

// Add a directory to the database.
func (db *Database) addDirectory(dir *Directory) {
	dir.Intern(db.indexer)
	db.directories = append(db.directories, dir)
}

// Clear and recreate the montage directory
func (db *Database) resetMontageDirectory() {
	os.RemoveAll(db.MontagePath())
	os.MkdirAll(db.MontagePath(), 0777)
}

func (db *Database) newLoad(rel_pat string) *loaderLoad {
	mod_time, ok := db.file_times.ModTime(rel_pat)
	// Avoid weird time comparisons by rounding to seconds as we only
	// store seconds in the protos.
	mod_time = mod_time.Round(time.Second)
	if ok {
		return &loaderLoad{rel_pat: rel_pat, orgd_mtime: mod_time}
	} else {
		return &loaderLoad{rel_pat: rel_pat, orgd_mtime: mod_time,
			err: errors.New("missing time for " + rel_pat)}
	}
}

type ByMostRecent []*Directory

func (a ByMostRecent) Len() int           { return len(a) }
func (a ByMostRecent) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByMostRecent) Less(i, j int) bool { return a[i].Time().Before(a[j].Time()) }

func (db *Database) Load(update_disk, minify, force_reload bool) error {
	LoadSynonyms(db.static_root)
	N := 3
	pat_ch := make(chan *loaderLoad, N)
	res_ch := make(chan *loaderResult, N)
	for i := 0; i < N; i++ {
		go db.loaderWorker(update_disk, force_reload, pat_ch, res_ch)
	}
	log.Printf("Loading database\n")
	start_time := time.Now()
	t, err := DirModTime(db.orig_root)
	if err != nil {
		log.Printf("%s: %s\n", db.orig_root, err.Error())
		return err
	}
	queue := make([]*loaderLoad, 0, N)
	db.file_times.RecordOne("", t)
	queue = append(queue, db.newLoad(""))
	left := len(queue)
	for left > 0 {
		has_pushed := false
		if len(queue) > 0 {
			select {
			case pat_ch <- queue[0]:
				queue = queue[1:]
				has_pushed = true
			default:
			}
		}
		if !has_pushed {
			res := <-res_ch
			left -= 1
			if res.err == nil {
				db.file_times.Record(res.rel_pat, res.orgd_subs)
				db.addDirectory(res.dir)
				for _, fi := range res.orgd_subs {
					if !fi.Mode().IsRegular() {
						jp := path.Join(res.rel_pat, fi.Name())
						queue = append(queue, db.newLoad(jp))
						left += 1
					}
				}
			}
		}
	}
	close(pat_ch)
	close(res_ch)
	log.Printf("Loaded %d directories in %g ms\n",
		len(db.directories),
		time.Since(start_time).Seconds()*1000)
	log.Printf("%d files\n", len(db.file_times))
	start_time = time.Now()
	sort.Sort(ByMostRecent(db.directories))
	start_time = time.Now()
	num_images := db.indexer.BuildIndex(db)
	log.Printf("Indexed %d images in %g ms\n",
		num_images,
		time.Since(start_time).Seconds()*1000)
	if minify {
		start_time = time.Now()
		go db.resetMontageDirectory()
		num_minified := MinifyDatabase(db, false, false)
		log.Printf("Minifed %d images in %g ms\n",
			num_minified,
			time.Since(start_time).Seconds()*1000)
	}
	
	// Compute and cache recent active keywords
	start_time = time.Now()
	db.recentActiveKeywords = db.GetRecentActiveKeywordsAt(time.Now())
	log.Printf("Computed %d recent active keywords in %g ms\n",
		len(db.recentActiveKeywords),
		time.Since(start_time).Seconds()*1000)
	
	db.file_times = nil // free that.
	return nil
}

// KeywordCount represents a keyword with its occurrence count and sample images
type KeywordCount struct {
	Keyword      string   `json:"keyword"`
	Count        int      `json:"count"`
	RecentImages []*Image `json:"recent_images"`
}

// GetRecentActiveKeywords returns keywords from albums with directory timestamp less than one month old,
// sorted by number of occurrences in these recent photos. Uses cached data computed during Load().
func (db *Database) GetRecentActiveKeywords() []KeywordCount {
	if db.recentActiveKeywords != nil {
		return db.recentActiveKeywords
	}
	// Fallback to computation if cache is empty (e.g., during testing)
	return db.GetRecentActiveKeywordsAt(time.Now())
}

// GetRecentActiveKeywordsAt returns keywords from albums with directory timestamp less than one month before the given time,
// sorted by number of occurrences in these recent photos. This version allows for testing with a specific time.
func (db *Database) GetRecentActiveKeywordsAt(now time.Time) []KeywordCount {
	oneMonthAgo := now.AddDate(0, -1, 0)
	keywordImages := make(map[string][]*Image)
	
	// Find directories with last_modified time within the last month
	for _, dir := range db.directories {
		if dir.last_modified.After(oneMonthAgo) {
			// Collect images for each keyword from this recent directory
			for _, img := range dir.images {
				// Process main keywords
				for _, keyword := range img.keywords {
					if keyword != "" { // Skip empty keywords
						keywordImages[keyword] = append(keywordImages[keyword], img)
					}
				}
				// Process sub-keywords
				for _, keyword := range img.sub_keywords {
					if keyword != "" { // Skip empty keywords
						keywordImages[keyword] = append(keywordImages[keyword], img)
					}
				}
			}
		}
	}
	
	// Convert map to slice and prepare recent images for each keyword
	result := make([]KeywordCount, 0, len(keywordImages))
	for keyword, images := range keywordImages {
		// Sort images by item timestamp (most recent first)
		sort.Slice(images, func(i, j int) bool {
			return images[i].ItemTime().After(images[j].ItemTime())
		})
		
		// Take up to 4 most recent images
		maxImages := 4
		if len(images) < maxImages {
			maxImages = len(images)
		}
		
		result = append(result, KeywordCount{
			Keyword:      keyword,
			Count:        len(images),
			RecentImages: images[:maxImages],
		})
	}
	
	// Sort by count (descending), then by keyword name for consistency
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].Keyword < result[j].Keyword
		}
		return result[i].Count > result[j].Count
	})
	
	return result
}
