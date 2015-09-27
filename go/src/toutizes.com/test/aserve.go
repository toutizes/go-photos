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
var db_root = flag.String("db_root", "", "path to the database root")
var url_prefix = flag.String("url_prefix", "/db", "prefix for the usls.")


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
  cpu := runtime.NumCPU()
  runtime.GOMAXPROCS(cpu)

  if *static_root == "" {
    log.Fatal("Must pass --static_root")
  }
  if *db_root == "" {
    *db_root = *static_root
  }
  db := model.NewDatabase(*db_root)
  db.Load(true, true)
  println("Serving...")

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
      http.StripPrefix(*url_prefix, http.FileServer(http.Dir(*db_root)))))
  http.Handle(
    *url_prefix + "/maxi/",
    LogHandler(
      "/maxi",
      http.StripPrefix(*url_prefix, http.FileServer(http.Dir(*db_root)))))
  http.Handle("/",
    LogHandler(
      "/ default",
      http.FileServer(http.Dir(*static_root + "/"))))

  for {
    err := http.ListenAndServe(":" + *port, nil)
    log.Printf("Web server error: %s\n", err)
  }
}