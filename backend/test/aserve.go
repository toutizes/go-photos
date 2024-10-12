package main

import (
  "flag"
  "log"
  "net/http"
  "runtime"

  // "github.com/alexedwards/scs/v2"
)

import (
  "toutizes.com/go-photwo/backend/model"
)

var static_root = flag.String("static_root", "", "path to the static files")
var orig_root = flag.String("orig_root", "", "path to the original images")
var root = flag.String("root", "", "path to the database index, mini, etc")
var url_prefix = flag.String("url_prefix", "/db", "prefix for the urls.")
var num_cpu = flag.Int("num_cpu", 0, "Number of CPUs to use.  0 means MAXPROC.")
var update_db = flag.Bool("update_db", true, "If true update the database files.")
var force_reload = flag.Bool("force_reload", false, 
  "If true force a reload of images.")
var use_https = flag.Bool("use_https", false, "If true listen for HTTPS in 443.")

// var sessionManager *scs.SessionManager
// var cookieSalt = "da89HIuneDMBa8eThg-9VYcDScApDUKIXaiFXcbvMys"

func Log(n string, r *http.Request) {
  log.Printf("%s: %s %s\n", n, r.RemoteAddr, r.URL)
}

func LogHandler(n string, handler http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    Log(n, r)
    handler.ServeHTTP(w, r)
  })
}

func main() {
  flag.Parse()

  if *num_cpu == 0 {
    *num_cpu = runtime.NumCPU()
  }
  log.Printf("Cpus: %d\n", runtime.GOMAXPROCS(*num_cpu))

  if *static_root == "" {
    log.Fatal("Must pass --static_root")
  }
  if *orig_root == "" {
    log.Fatal("Must pass --orig_root")
  }
  if *root == "" {
    log.Fatal("Must pass --root")
  }
  model.LoadSynonyms(*static_root)
  db := model.NewDatabase2(*orig_root, *root)
  db.Load(*update_db, *update_db, *force_reload)
  log.Printf("Serving...")

  // // Initialize the session manager
  // sessionManager = scs.New()
  // sessionManager.Store = scs.NewCookieStore([]byte(cookieSalt))

  var mux = http.NewServeMux()
  mux.HandleFunc(*url_prefix + "/q",
    func (w http.ResponseWriter, r *http.Request) {
      Log("/q", r)
      model.HandleQuery(w, r, db)
    })
  mux.HandleFunc(*url_prefix + "/montage/",
    func (w http.ResponseWriter, r *http.Request) {
      Log("/montage", r)
      model.HandleMontage2(w, r, db)
    })
  mux.HandleFunc(*url_prefix + "/viewer",
    func (w http.ResponseWriter, r *http.Request) {
      Log("/viewer", r)
      model.HandleCommands(w, r, db)
    })
  mux.HandleFunc(*url_prefix + "/mini/",
    func(w http.ResponseWriter, r *http.Request) {
      Log("/mini", r)
      model.HandleFile(w, r, *url_prefix, *root)
    })
  mux.HandleFunc(*url_prefix + "/midi/",
    func(w http.ResponseWriter, r *http.Request) {
      Log("/midi", r)
      model.HandleFile(w, r, *url_prefix, *root)
    })
  mux.HandleFunc(*url_prefix + "/maxi/",
    func(w http.ResponseWriter, r *http.Request) {
      Log("/maxi", r)
      model.HandleFile(w, r, *url_prefix + "/maxi/", *orig_root)
    })
  // mux.Handle("/",
  //   LogHandler(
  //     "/ default",
  //     http.FileServer(http.Dir(*static_root + "/"))))
  mux.Handle(
		"/db/",
    LogHandler(
      "/db files",
      http.FileServer(http.Dir(*static_root))))
  mux.HandleFunc("/",
    func (w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/db/pic.html", 301)
    })

  // Always serve on http:8081, useful for access from localhost.
  go http.ListenAndServe(":8081", mux)

  // Serve images on https:8443/http:8080 (redirect)
  if (*use_https) {
    // Redirect HTTP to HTTPS
    var http_mux = http.NewServeMux()
    http_mux.HandleFunc("/",
      func (w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "https://toutizes.com" + r.RequestURI, 301)
      })
    go http.ListenAndServe(":8080", http_mux)

    // Listen on HTTPS.
    err := http.ListenAndServeTLS(":8443", 
      "/etc/letsencrypt/live/toutizes.com/fullchain.pem",
      "/etc/letsencrypt/live/toutizes.com/privkey.pem",
      mux)
    log.Printf("Web server error: %s\n", err)
  } else {
    err := http.ListenAndServe(":8080", mux)
    log.Printf("Web server error: %s\n", err)
  }    
}
