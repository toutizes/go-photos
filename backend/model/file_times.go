package model

import (
  "os"
  "path"
  "time"
)

type FileTimes map[string]time.Time

func NewFileTimes() FileTimes {
  return make(FileTimes, 1000)
}

func (fts FileTimes) RecordOne(file string, t time.Time) {
  fts[file] = t
}

func (fts FileTimes) Record(root string, infos []os.FileInfo) {
  for _, i := range infos {
    fts[path.Join(root, i.Name())] = i.ModTime()
  }
}

func (fts FileTimes) ModTime(file string) (time.Time, bool) {
  t, ok := fts[file]
  return t, ok
}
