package main

import (
  "flag"
  "log"
  "net/http"
  "runtime"
)

import (
  "toutizes.com/model"
)

var static_root = flag.String("static_root", "", "path to the static files")
var port = flag.String("port", "80", "listen port")
var orig_root = flag.String("orig_root", "", "path to the original images")
var root = flag.String("root", "", "path to the database index, mini, etc")
var url_prefix = flag.String("url_prefix", "/db", "prefix for the urls.")
var num_cpu = flag.Int("num_cpu", 0, "Number of CPUs to use.  0 means MAXPROC.")


func Log(n string, r *http.Request) {
  log.Printf("%s %s\n", r.RemoteAddr, r.URL)
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
  db := model.NewDatabase2(*orig_root, *root)
  db.Load(true, true)
  log.Printf("Serving...")

  http.HandleFunc(*url_prefix + "/q",
    func (w http.ResponseWriter, r *http.Request) {
      Log("/q", r)
      model.HandleQuery(w, r, db)
    })
  http.HandleFunc(*url_prefix + "/montage/",
    func (w http.ResponseWriter, r *http.Request) {
      Log("/montage", r)
      model.HandleMontage2(w, r, db)
    })
  http.HandleFunc(*url_prefix + "/viewer",
    func (w http.ResponseWriter, r *http.Request) {
      Log("/viewer", r)
      model.HandleCommands(w, r, db)
    })
  http.Handle(
    *url_prefix + "/midi/",
    LogHandler(
      "/midi",
      http.StripPrefix(*url_prefix, http.FileServer(http.Dir(*root)))))
  http.Handle(
    *url_prefix + "/maxi/",
    LogHandler(
      "/maxi",
      http.StripPrefix(*url_prefix, http.FileServer(http.Dir(*root)))))
  http.Handle("/",
    LogHandler(
      "/ default",
      http.FileServer(http.Dir(*static_root + "/"))))

  for {
    err := http.ListenAndServe(":" + *port, nil)
    log.Printf("Web server error: %s\n", err)
  }
}
