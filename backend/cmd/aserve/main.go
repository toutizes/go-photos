package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"time"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
	"toutizes.com/go-photwo/backend/catalog"
	"toutizes.com/go-photwo/backend/server"
)

var (
	staticRoot      = flag.String("static_root", "", "path to the static files")
	origRoot        = flag.String("orig_root", "", "path to the original images")
	letsencryptRoot = flag.String("letsencrypt_root", "", "path to letsencrypt files")
	root            = flag.String("root", "", "path to the database index, mini, midi")
	urlPrefix       = flag.String("url_prefix", "/db", "prefix for the urls")
	numCPU          = flag.Int("num_cpu", 0, "Number of CPUs to use. 0 means MAXPROC.")
	updateDB        = flag.Bool("update_db", true, "If true update the database files.")
	forceReload     = flag.Bool("force_reload", false, "If true force a reload of images.")
	useHTTPS        = flag.Bool("use_https", false, "If true listen for HTTPS on 8443 and redirect 8080.")
	firebaseCreds   = flag.String("firebase_creds", "", "Path to the Firebase service account credentials JSON file")
	logDir          = flag.String("log_dir", "", "Path to directory containing query log files for analysis")
	minifierThreads = flag.Int("minifier_threads", 2, "Number of threads for minifying.")

	// Kept for CLI backwards-compatibility with existing startup scripts.
	_ = flag.Bool("use_lr_parser", true, "Use Lightroom-style query parser (always enabled)")
	_ = flag.String("bin_root", "/usr/local/bin/", "directory for the binaries (deprecated)")
)

func main() {
	flag.Parse()

	if *numCPU == 0 {
		*numCPU = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(*numCPU)
	log.Printf("Cpus: %d", runtime.GOMAXPROCS(*numCPU))

	if *staticRoot == "" {
		log.Fatal("Must pass --static_root")
	}
	if *origRoot == "" {
		log.Fatal("Must pass --orig_root")
	}
	if *root == "" {
		log.Fatal("Must pass --root")
	}
	if *firebaseCreds == "" {
		log.Fatal("Must pass --firebase_creds")
	}

	opt := option.WithCredentialsFile(*firebaseCreds)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v", err)
	}
	authClient, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Error initializing Firebase Auth client: %v", err)
	}

	catOpts := catalog.Options{
		OrigRoot:        *origRoot,
		Root:            *root,
		StaticRoot:      *staticRoot,
		UpdateDisk:      *updateDB,
		Minify:          *updateDB,
		ForceReload:     *forceReload,
		MinifierThreads: *minifierThreads,
	}
	cat, err := catalog.LoadCatalog(catOpts)
	if err != nil {
		log.Fatalf("Error loading catalog: %v", err)
	}

	srv := server.New(cat, server.Config{
		CatalogOpts: catOpts,
		URLPrefix:   *urlPrefix,
		LogDir:      *logDir,
		Verifier:    &server.FirebaseVerifier{Client: authClient},
	})

	log.Printf("Serving...")

	mainHandler := srv.Handler()
	httpServer := &http.Server{
		Addr:           ":8080",
		Handler:        mainHandler,
		MaxHeaderBytes: 1 << 20, // 1MB
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    120 * time.Second,
	}

	// Serve localhost maintenance listener on :8081 (used by cron-pull.sh for reload).
	go func() {
		localServer := &http.Server{
			Addr:           "127.0.0.1:8081",
			Handler:        srv.LocalHandler(),
			MaxHeaderBytes: 1 << 20,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   120 * time.Second,
			IdleTimeout:    120 * time.Second,
		}
		if err := localServer.ListenAndServe(); err != nil {
			log.Printf("Local server error: %v", err)
		}
	}()

	if *useHTTPS {
		redirectMux := http.NewServeMux()
		redirectMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("HTTPS /: %s %s", r.RemoteAddr, r.URL)
			http.Redirect(w, r, "https://toutizes.com"+r.RequestURI, http.StatusMovedPermanently)
		})
		redirectServer := &http.Server{
			Addr:         ":8080",
			Handler:      redirectMux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		}
		go func() {
			if err := redirectServer.ListenAndServe(); err != nil {
				log.Printf("Redirect server error: %v", err)
			}
		}()

		tlsServer := &http.Server{
			Addr:           ":8443",
			Handler:        mainHandler,
			MaxHeaderBytes: 1 << 20,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   30 * time.Second,
			IdleTimeout:    120 * time.Second,
		}
		certFile := filepath.Join(*letsencryptRoot, "live/toutizes.com/fullchain.pem")
		keyFile := filepath.Join(*letsencryptRoot, "live/toutizes.com/privkey.pem")
		if err := tlsServer.ListenAndServeTLS(certFile, keyFile); err != nil {
			log.Printf("Web server error: %v", err)
		}
	} else {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Printf("Web server error: %v", err)
		}
	}
}
