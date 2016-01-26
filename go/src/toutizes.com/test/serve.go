package main

import (
  "log"
  "os"
  "net/http"
  "runtime"
)

import (
  "toutizes.com/model"
)

func main() {
  cpu := runtime.NumCPU()
  runtime.GOMAXPROCS(cpu)

  db_root := os.Args[1]
  web_root := os.Args[2]

  var web_port string
  if len(os.Args) == 4 {
    web_port = os.Args[3]
  } else {
    web_port = "8080"
  }

  db := model.NewDatabase(db_root)
	update_disk := true
	minify := true
	force_reload := true
  db.Load(update_disk, minify, force_reload)
  println("Serving...")

  http.HandleFunc("/q",
    func (w http.ResponseWriter, r *http.Request) {
      model.HandleQuery(w, r, db)
    })
  http.HandleFunc("/montage/",
    func (w http.ResponseWriter, r *http.Request) {
      model.HandleMontage(w, r, db)
    })
  http.HandleFunc("/set",
    func (w http.ResponseWriter, r *http.Request) {
      model.HandleSet(w, r, db)
    })
  http.HandleFunc("/viewer",
    func (w http.ResponseWriter, r *http.Request) {
      model.HandleCommands(w, r, db)
    })
  http.HandleFunc("/f",
    func (w http.ResponseWriter, r *http.Request) {
      model.HandleFeed(w, r, db)
    })
  http.Handle("/",
    http.FileServer(http.Dir("/opt/local/apache2/" + web_root + "/")))
  log.Fatal(http.ListenAndServe(":" + web_port, nil))
}
