package model

import (
  "io/ioutil"
  "os"
  "path"
  "testing"
  "time"
)

import "toutizes.com/go-photwo/backend/store"

func makeDir(t *testing.T, prefix string, files []string) (pat string, mod_time time.Time) {
  pat, err := ioutil.TempDir("", prefix)
  file, err := os.Open(pat)
  if err != nil {
    t.Fatal(err)
  }
  defer file.Close()
  stat, err := file.Stat()
  if err != nil {
    t.Fatal(err)
  }
  mod_time = stat.ModTime()
  for _, f := range files {
    ioutil.WriteFile(path.Join(pat, f), nil, 0777)
  }
  return
}

func testImage(kwds []string, its int64) *store.Item {
  return &store.Item{Keywords:kwds, ItemTimestamp:&its}
}

func testImageNoTs(kwds []string) *store.Item {
  return &store.Item{Keywords:kwds}
}

func testNamedImage(name string, kwds []string) *store.Item {
  return &store.Item{Name:&name, Keywords:kwds, Image:&store.Image{}}
}

func Test_MergeImage(t *testing.T) {
  old_img := testImage([]string{"a", "b"}, 123)
  var mrg_img *store.Item

  mrg_img = mergeImage(old_img, testImageNoTs(nil))
  if len(mrg_img.Keywords) != 2 {
    t.Error("kwds")
  }
  if *mrg_img.ItemTimestamp != 123 {
    t.Error("ts")
  }

  mrg_img = mergeImage(old_img, testImage(nil, 456))
  if len(mrg_img.Keywords) != 2 {
    t.Error("kwds")
  }
  if *mrg_img.ItemTimestamp != 456 {
    t.Error("ts")
  }

  mrg_img = mergeImage(old_img, testImage([]string{"z"}, 789))
  if len(mrg_img.Keywords) != 1 {
    t.Error("kwds")
  }
  if *mrg_img.ItemTimestamp != 789 {
    t.Error("ts")
  }
}

func findItem(sdir *store.Directory, name string) *store.Item {
  for _, it := range sdir.Items {
    if *it.Name == name {
      return it
    }
  }
  return nil
}

func Test_UpdateDir(t *testing.T) {
  dir_files := []string{"foo.jpg", "bar.webm",  "fee", "gee.JPG", "bidon"}
  dir, _ := makeDir(t, "update_dir", dir_files)

  {
    sdir := store.Directory{}
    subs, _ := ioutil.ReadDir(dir)
    _ = UpdateDirectory(subs, &sdir)
    
    if len(sdir.Items) != 3 {
      t.Error("num items")
    }
    for _, name := range []string{"foo.jpg", "bar.webm",  "gee.JPG"} {
      it := findItem(&sdir, name)
      if it == nil {
        t.Error("missing item %s", name)
      }
    }
  }

  {
    sdir := store.Directory{Items: []*store.Item{
      testNamedImage("old.img", []string{"baba"}),
      testNamedImage("gee.JPG", []string{"ba ba"}),
    }}
    subs, _ := ioutil.ReadDir(dir)
    _ = UpdateDirectory(subs, &sdir)
    
    if len(sdir.Items) != 3 {
      t.Error("num items")
    }
    for _, name := range []string{"foo.jpg", "bar.webm",  "gee.JPG"} {
      it := findItem(&sdir, name)
      if it == nil {
        t.Error("missing item %s", name)
      }
    }
    gee := findItem(&sdir, "gee.JPG")
    if gee == nil {
      t.Error("missing gee")
    }
    if len(gee.Keywords) != 1 {
      t.Error("missing kwds")
    }
    if gee.Keywords[0] != "ba ba" {
      t.Error("missing ba ba")
    }
  }
}
