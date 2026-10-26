package server

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"toutizes.com/go-photwo/backend/catalog"
)

const defaultAdminEmail = "matthieu.devin@gmail.com"

// Config holds configuration for constructing a Server.
type Config struct {
	CatalogOpts catalog.Options
	URLPrefix   string
	LogDir      string
	Verifier    TokenVerifier
	AdminEmail  string
}

// Server serves the photo HTTP API and static web application.
type Server struct {
	catalog    atomic.Pointer[catalog.Catalog]
	reloadMu   sync.Mutex
	opts       catalog.Options
	urlPrefix  string
	logDir     string
	verifier   TokenVerifier
	adminEmail string
}

// New creates a Server using an already-loaded initial Catalog.
func New(cat *catalog.Catalog, cfg Config) *Server {
	prefix := cfg.URLPrefix
	if prefix == "" {
		prefix = "/db"
	}
	admin := cfg.AdminEmail
	if admin == "" {
		admin = defaultAdminEmail
	}
	s := &Server{
		opts:       cfg.CatalogOpts,
		urlPrefix:  prefix,
		logDir:     cfg.LogDir,
		verifier:   cfg.Verifier,
		adminEmail: admin,
	}
	s.catalog.Store(cat)
	return s
}

// Catalog returns the current active Catalog snapshot.
func (s *Server) Catalog() *catalog.Catalog {
	return s.catalog.Load()
}

// Reload loads a fresh Catalog from disk (updating protobuf caches and minifying new images)
// and atomically swaps it into the Server.
func (s *Server) Reload() error {
	s.reloadMu.Lock()
	defer s.reloadMu.Unlock()

	reloadOpts := s.opts
	reloadOpts.UpdateDisk = true
	reloadOpts.Minify = true
	reloadOpts.ForceReload = false

	newCat, err := catalog.LoadCatalog(reloadOpts)
	if err != nil {
		return err
	}
	s.catalog.Store(newCat)
	return nil
}

func logRequest(tag string, r *http.Request) {
	log.Printf("%s: %s %s", tag, r.RemoteAddr, r.URL)
}

// GetContentType returns the MIME type for a file path based on its extension.
func GetContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
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

func isSafeSubpath(root, candidate string) bool {
	cleanRoot := filepath.Clean(root)
	cleanCand := filepath.Clean(candidate)
	if cleanCand == cleanRoot {
		return true
	}
	return strings.HasPrefix(cleanCand, cleanRoot+string(filepath.Separator))
}

func serveCachedFile(w http.ResponseWriter, fullPath string) {
	file, err := os.Open(fullPath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil || info.IsDir() {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", GetContentType(fullPath))
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))

	buf := make([]byte, 32*1024)
	if _, err := io.CopyBuffer(w, file, buf); err != nil {
		log.Printf("Error sending file %s: %v", fullPath, err)
	}
}

func (s *Server) handleFileFromRoot(w http.ResponseWriter, r *http.Request, prefix, root string) {
	if !strings.HasPrefix(r.URL.Path, prefix) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	rel := r.URL.Path[len(prefix):]
	fullPath := filepath.Join(root, filepath.FromSlash(rel))
	if !isSafeSubpath(root, fullPath) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	serveCachedFile(w, fullPath)
}

func (s *Server) handleWebApp(w http.ResponseWriter, r *http.Request) {
	staticRoot := s.opts.StaticRoot
	if cat := s.Catalog(); cat != nil && cat.StaticRoot != "" {
		staticRoot = cat.StaticRoot
	}
	webRoot := filepath.Join(staticRoot, "flutter")
	rel := strings.TrimPrefix(r.URL.Path, "/app/")
	if rel == "" {
		rel = "index.html"
	}
	filePath := filepath.Join(webRoot, filepath.FromSlash(rel))
	if !isSafeSubpath(webRoot, filePath) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if info, err := os.Stat(filePath); err != nil || info.IsDir() {
		filePath = filepath.Join(webRoot, "index.html")
	}

	w.Header().Set("Content-Type", GetContentType(filePath))
	AddCorsHeaders(w, r)
	if strings.HasSuffix(filePath, "index.html") {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
	}
	http.ServeFile(w, r, filePath)
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	q := r.FormValue("q")
	kind := r.FormValue("kind")
	userEmail := UserEmail(r.Context())
	log.Printf("Query from %s: %q (kind: %s)", userEmail, q, kind)

	cat := s.Catalog()
	photos := cat.QueryPhotos(q)

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)

	if kind == "album" {
		seen := make(map[*catalog.Album]bool)
		albums := make([]*catalog.Album, 0)
		for _, p := range photos {
			if !seen[p.Album] {
				seen[p.Album] = true
				albums = append(albums, p.Album)
			}
		}
		res := make([]catalog.JsonDirectory, len(albums))
		for i, a := range albums {
			a.FillJson(&res[i])
		}
		_ = enc.Encode(&res)
		return
	}

	res := make([]catalog.JsonImage, len(photos))
	for i, p := range photos {
		p.FillJson(&res[i])
	}
	_ = enc.Encode(&res)
}

func (s *Server) handleRecentKeywordGroups(w http.ResponseWriter, r *http.Request) {
	userEmail := UserEmail(r.Context())
	log.Printf("Recent keyword groups request from %s", userEmail)

	var groups []catalog.KeywordGroupResponse
	if cat := s.Catalog(); cat != nil {
		groups = cat.RecentKeywordGroups
	}
	if groups == nil {
		groups = []catalog.KeywordGroupResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&catalog.KeywordGroupResults{Groups: groups})
}

func (s *Server) handleThumb(w http.ResponseWriter, r *http.Request) {
	prefix := s.urlPrefix + "/thumb/"
	idStr := strings.TrimPrefix(r.URL.Path, prefix)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Bad thumbnail id", http.StatusBadRequest)
		return
	}
	cat := s.Catalog()
	if cat == nil {
		http.Error(w, "Catalog unavailable", http.StatusServiceUnavailable)
		return
	}
	miniPath := cat.MiniPathByID(id)
	if miniPath == "" {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}
	serveCachedFile(w, miniPath)
}

func (s *Server) handleDownload(w http.ResponseWriter, vals url.Values) {
	qVals, okQ := vals["q"]
	sVals, okS := vals["s"]
	if !okQ || len(qVals) == 0 || !okS || len(sVals) == 0 {
		http.Error(w, "Missing q or s parameter", http.StatusBadRequest)
		return
	}
	cat := s.Catalog()
	if cat == nil {
		http.Error(w, "Catalog unavailable", http.StatusServiceUnavailable)
		return
	}
	photos := cat.QueryPhotos(qVals[0])
	if len(photos) == 0 {
		return
	}
	if len(photos) > 100 {
		photos = photos[:100]
	}

	w.Header().Set("Content-Disposition", `attachment; filename="images.zip"`)
	w.Header().Set("Content-Type", "application/zip")

	zw := zip.NewWriter(w)
	defer zw.Close()

	useMidi := sVals[0] == "M"
	for _, p := range photos {
		entryPath := path.Join(p.Album.RelPat, p.Name)
		zf, err := zw.Create(entryPath)
		if err != nil {
			log.Printf("Zip create error %s: %v", entryPath, err)
			break
		}
		var srcPath string
		if useMidi {
			srcPath = cat.MidiPath(p)
		} else {
			srcPath = cat.OrigPath(p)
		}
		srcFile, err := os.Open(srcPath)
		if err != nil {
			log.Printf("Zip open error %s: %v", srcPath, err)
			break
		}
		_, copyErr := io.Copy(zf, srcFile)
		_ = srcFile.Close()
		if copyErr != nil {
			log.Printf("Zip copy error %s: %v", srcPath, copyErr)
			break
		}
	}
}

func (s *Server) handleCommands(w http.ResponseWriter, r *http.Request) {
	vals, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, "Bad query", http.StatusBadRequest)
		return
	}
	comm, ok := vals["command"]
	if !ok || len(comm) == 0 {
		http.Error(w, "Missing command", http.StatusBadRequest)
		return
	}
	switch comm[0] {
	case "download":
		s.handleDownload(w, vals)
	case "reload":
		if err := s.Reload(); err != nil {
			fmt.Fprint(w, err.Error())
		} else {
			fmt.Fprint(w, "Reloaded")
		}
	default:
		http.Error(w, "Unknown command", http.StatusBadRequest)
	}
}

// AllUserQueriesResult is the JSON payload for /db/user-queries.
type AllUserQueriesResult struct {
	Users map[string][]QueryWithTimestamp `json:"users"`
}

func (s *Server) handleUserQueries(w http.ResponseWriter, r *http.Request) {
	if s.logDir == "" {
		http.Error(w, "Log directory not configured", http.StatusInternalServerError)
		return
	}
	userEmail := UserEmail(r.Context())
	log.Printf("User queries request from %s", userEmail)

	if userEmail != s.adminEmail {
		log.Printf("Access denied for user queries endpoint: %s", userEmail)
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	parser := NewLogParser(s.logDir)
	allQueries, err := parser.GroupQueriesByUser()
	if err != nil {
		log.Printf("Error getting all user queries: %v", err)
		http.Error(w, "Error retrieving user queries", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&AllUserQueriesResult{Users: allQueries})
}

func (s *Server) authed(tag string, fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(tag, r)
		AuthMiddleware(s.verifier, func(w http.ResponseWriter, r *http.Request) {
			AddCorsHeaders(w, r)
			fn(w, r)
		})(w, r)
	}
}

// Handler returns the main HTTP handler with Firebase authentication on all /db/* endpoints.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	p := s.urlPrefix

	mux.HandleFunc(p+"/q", s.authed("/q", s.handleQuery))
	mux.HandleFunc(p+"/thumb/", s.authed("/thumb", s.handleThumb))
	mux.HandleFunc(p+"/viewer", s.authed("/viewer", s.handleCommands))
	mux.HandleFunc(p+"/mini/", s.authed("/mini", func(w http.ResponseWriter, r *http.Request) {
		root := s.opts.Root
		if cat := s.Catalog(); cat != nil {
			root = cat.Root
		}
		s.handleFileFromRoot(w, r, p, root)
	}))
	mux.HandleFunc(p+"/midi/", s.authed("/midi", func(w http.ResponseWriter, r *http.Request) {
		root := s.opts.Root
		if cat := s.Catalog(); cat != nil {
			root = cat.Root
		}
		s.handleFileFromRoot(w, r, p, root)
	}))
	mux.HandleFunc(p+"/maxi/", s.authed("/maxi", func(w http.ResponseWriter, r *http.Request) {
		origRoot := s.opts.OrigRoot
		if cat := s.Catalog(); cat != nil {
			origRoot = cat.OrigRoot
		}
		s.handleFileFromRoot(w, r, p+"/maxi/", origRoot)
	}))
	mux.HandleFunc(p+"/recent-keyword-groups", s.authed("/recent-keyword-groups", s.handleRecentKeywordGroups))
	mux.HandleFunc(p+"/user-queries", s.authed("/user-queries", s.handleUserQueries))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logRequest("/", r)
		http.Redirect(w, r, "/app/", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/app/", func(w http.ResponseWriter, r *http.Request) {
		logRequest("/app", r)
		s.handleWebApp(w, r)
	})

	return mux
}

// LocalHandler returns the HTTP handler for the localhost :8081 listener, allowing
// local cron jobs (e.g. cmd/cron-pull.sh) to trigger /db/viewer?command=reload without a JWT.
func (s *Server) LocalHandler() http.Handler {
	mainHandler := s.Handler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == s.urlPrefix+"/viewer" && r.URL.Query().Get("command") == "reload" {
			logRequest("local /viewer reload", r)
			if err := s.Reload(); err != nil {
				fmt.Fprint(w, err.Error())
			} else {
				fmt.Fprint(w, "Reloaded")
			}
			return
		}
		mainHandler.ServeHTTP(w, r)
	})
}
