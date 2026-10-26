package catalog

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"google.golang.org/protobuf/proto"
	"toutizes.com/go-photwo/backend/store"
)

// Options configures paths and loading behavior for a Catalog.
type Options struct {
	OrigRoot        string
	Root            string
	StaticRoot      string
	UpdateDisk      bool
	Minify          bool
	ForceReload     bool
	MinifierThreads int
}

// Catalog is an immutable in-memory snapshot of all albums, photos, search indices,
// and precomputed recent keyword groups.
type Catalog struct {
	OrigRoot            string
	Root                string
	IndexRoot           string
	MidiRoot            string
	MiniRoot            string
	StaticRoot          string
	Albums              []*Album
	Index               *Index
	Synonyms            *Synonyms
	RecentKeywordGroups []KeywordGroupResponse
}

// NewCatalog initializes an empty Catalog with standard directory paths under root.
func NewCatalog(origRoot, root, staticRoot string) *Catalog {
	return &Catalog{
		OrigRoot:   origRoot,
		Root:       root,
		IndexRoot:  filepath.Join(root, "index"),
		MidiRoot:   filepath.Join(root, "midi"),
		MiniRoot:   filepath.Join(root, "mini"),
		StaticRoot: staticRoot,
		Index:      NewIndex(),
		Synonyms:   NewSynonyms(),
	}
}

// IndexPath returns the binary protobuf cache path for a relative album path.
func (c *Catalog) IndexPath(relPat string) string {
	return filepath.Join(c.IndexRoot, relPat, "index.pbin")
}

// MiniPath returns the path to a photo's 360x360 square thumbnail.
func (c *Catalog) MiniPath(p *Photo) string {
	if p == nil {
		return ""
	}
	return filepath.Join(c.MiniRoot, p.Album.RelPat, p.Name)
}

// MiniPathByID returns the path to a photo's 360x360 square thumbnail by photo ID.
func (c *Catalog) MiniPathByID(id int) string {
	return c.MiniPath(c.Index.PhotoByID(id))
}

// MidiPath returns the path to a photo's <=2048px display image.
func (c *Catalog) MidiPath(p *Photo) string {
	if p == nil {
		return ""
	}
	return filepath.Join(c.MidiRoot, p.Album.RelPat, p.Name)
}

// MidiPathByID returns the path to a photo's <=2048px display image by photo ID.
func (c *Catalog) MidiPathByID(id int) string {
	return c.MidiPath(c.Index.PhotoByID(id))
}

// OrigPath returns the path to a photo's original full-resolution file.
func (c *Catalog) OrigPath(p *Photo) string {
	if p == nil {
		return ""
	}
	return filepath.Join(c.OrigRoot, p.Album.RelPat, p.Name)
}

// OrigPathByID returns the path to a photo's original full-resolution file by photo ID.
func (c *Catalog) OrigPathByID(id int) string {
	return c.OrigPath(c.Index.PhotoByID(id))
}

func readIndexProto(idxPath string, sdir *store.Directory) error {
	buf, err := os.ReadFile(idxPath)
	if err != nil {
		return err
	}
	return proto.Unmarshal(buf, sdir)
}

func writeIndexProto(binPath string, sdir *store.Directory) error {
	log.Printf("write %s", binPath)
	if err := os.MkdirAll(filepath.Dir(binPath), 0777); err != nil {
		return err
	}
	data, err := proto.Marshal(sdir)
	if err != nil {
		return err
	}
	return os.WriteFile(binPath, data, 0666)
}

func readDirFileInfos(dir string) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	infos := make([]os.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}

type loadJob struct {
	relPat    string
	origMTime time.Time
	err       error
}

type loadResult struct {
	album    *Album
	relPat   string
	origSubs []os.FileInfo
	err      error
}

func (c *Catalog) handleLoadDir(updateDisk, forceReload bool, job *loadJob) (*loadResult, error) {
	if job.err != nil {
		return nil, job.err
	}
	var sdir store.Directory
	_ = readIndexProto(c.IndexPath(job.relPat), &sdir)

	origDir := filepath.Join(c.OrigRoot, job.relPat)
	subs, err := readDirFileInfos(origDir)
	if err != nil {
		return nil, err
	}

	if sdir.DirectoryTimestamp == nil ||
		ProtoToTime(*sdir.DirectoryTimestamp).Before(job.orgdMTimeRounded()) ||
		forceReload {
		if err := UpdateDirectory(origDir, subs, forceReload, &sdir); err != nil {
			return nil, err
		}
		orgdTS := TimeToProto(job.orgdMTimeRounded())
		sdir.DirectoryTimestamp = &orgdTS
		if updateDisk {
			if werr := writeIndexProto(c.IndexPath(job.relPat), &sdir); werr != nil {
				log.Printf("%s: %v", job.relPat, werr)
			}
		}
	}

	album := ProtoToAlbum(&sdir, job.relPat)
	return &loadResult{album: album, relPat: job.relPat, origSubs: subs}, nil
}

func (j *loadJob) orgdMTimeRounded() time.Time {
	return j.origMTime.Round(time.Second)
}

// Load scans OrigRoot, updates or reads per-directory protobuf caches, builds the search
// index, optionally minifies images, and precomputes recent keyword groups.
func (c *Catalog) Load(opts Options) error {
	c.Synonyms = LoadSynonyms(c.StaticRoot)

	const workers = 3
	jobCh := make(chan *loadJob, workers)
	resCh := make(chan *loadResult, workers)
	for i := 0; i < workers; i++ {
		go func() {
			for job := range jobCh {
				res, err := c.handleLoadDir(opts.UpdateDisk, opts.ForceReload, job)
				if err == nil {
					resCh <- res
				} else {
					log.Printf("Worker err: %v", err)
					resCh <- &loadResult{err: err}
				}
			}
		}()
	}

	log.Printf("Loading database")
	startTime := time.Now()
	stat, err := os.Stat(c.OrigRoot)
	if err != nil {
		log.Printf("%s: %v", c.OrigRoot, err)
		close(jobCh)
		return err
	}

	fileTimes := map[string]time.Time{"": stat.ModTime()}
	newJob := func(relPat string) *loadJob {
		mt, ok := fileTimes[relPat]
		if !ok {
			return &loadJob{relPat: relPat, err: errors.New("missing time for " + relPat)}
		}
		return &loadJob{relPat: relPat, origMTime: mt.Round(time.Second)}
	}

	queue := []*loadJob{newJob("")}
	left := len(queue)
	for left > 0 {
		pushed := false
		if len(queue) > 0 {
			select {
			case jobCh <- queue[0]:
				queue = queue[1:]
				pushed = true
			default:
			}
		}
		if !pushed {
			res := <-resCh
			left--
			if res.err == nil {
				for _, fi := range res.origSubs {
					rel := fi.Name()
					if res.relPat != "" {
						rel = res.relPat + "/" + fi.Name()
					}
					fileTimes[rel] = fi.ModTime()
				}
				c.Index.internAlbum(res.album)
				c.Albums = append(c.Albums, res.album)
				for _, fi := range res.origSubs {
					if !fi.Mode().IsRegular() {
						rel := fi.Name()
						if res.relPat != "" {
							rel = res.relPat + "/" + fi.Name()
						}
						queue = append(queue, newJob(rel))
						left++
					}
				}
			}
		}
	}
	close(jobCh)
	close(resCh)

	log.Printf("Loaded %d directories in %.1f ms", len(c.Albums), time.Since(startTime).Seconds()*1000)

	sort.Slice(c.Albums, func(i, j int) bool {
		if c.Albums[i].IndexTime.Equal(c.Albums[j].IndexTime) {
			return c.Albums[i].RelPat < c.Albums[j].RelPat
		}
		return c.Albums[i].IndexTime.Before(c.Albums[j].IndexTime)
	})

	startTime = time.Now()
	numImages := c.Index.Build(c.Albums)
	log.Printf("Indexed %d images in %.1f ms", numImages, time.Since(startTime).Seconds()*1000)

	if opts.Minify {
		startTime = time.Now()
		numMinified := MinifyCatalog(c, opts.MinifierThreads, false)
		log.Printf("Minified %d images in %.1f ms", numMinified, time.Since(startTime).Seconds()*1000)
	}

	startTime = time.Now()
	c.RecentKeywordGroups = ComputeRecentKeywordGroups(c.Albums)
	log.Printf("Computed %d recent keyword groups in %.1f ms",
		len(c.RecentKeywordGroups), time.Since(startTime).Seconds()*1000)
	return nil
}

// LoadCatalog creates and loads a new Catalog with the given options.
func LoadCatalog(opts Options) (*Catalog, error) {
	c := NewCatalog(opts.OrigRoot, opts.Root, opts.StaticRoot)
	if err := c.Load(opts); err != nil {
		return nil, err
	}
	return c, nil
}
