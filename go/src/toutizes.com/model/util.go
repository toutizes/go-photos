package model

import (
  "os"
  "time"
)

// Convert between proto time and time.Time
func ProtoToTime(ptime int64) time.Time {
  return time.Unix(ptime / 1000, 0)
}

func TimeToProto(t time.Time) int64 { 
  return 1000 * t.Unix()
}

func DirModTime(dir string) (time time.Time, err error) {
  file, err := os.Open(dir)
  if err != nil {
    return
  }
  defer file.Close()
  stat, err := file.Stat()
  if err != nil {
    return
  }
  time = stat.ModTime()
  return
}

// Return the mod time of "path" or nil if "path" does not exist.
func ChangedSince(path string, timestamp time.Time) (changed bool, err error) {
  t, err := DirModTime(path)
  if err != nil {
    return
  }
  changed = t.After(timestamp)
  return
}
