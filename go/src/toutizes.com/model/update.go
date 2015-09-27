package model

import (
  "os"
  "path"
  "strings"
)

import "toutizes.com/store"

const (
  ft_dir = 0
  ft_img = 1
  ft_vid= 2
  ft_ign = 3
)

var ignored_names = map[string]bool{
  // Old imageDB names.
  ".minis":true, ".comments":true,
  // iPhoto names.
  "Data":true, "Albums":true, "Desktop":true, "Originals":true, "Thumbs":true}
  
func isImageName(name string) bool {
  return strings.HasSuffix(name, ".jpg") ||
    strings.HasSuffix(name, ".JPG") ||
    strings.HasSuffix(name, ".jpeg") ||
    strings.HasSuffix(name, ".BMP") ||
    strings.HasSuffix(name, ".TIF") ||
    strings.HasSuffix(name, ".gif") ||
    strings.HasSuffix(name, ".GIF")
}

func isVideoName(name string) bool {
  return strings.HasSuffix(name, ".webm") ||
    strings.HasSuffix(name, ".WEBM")
}

func fileType(file os.FileInfo) int {
  switch {
  case ignored_names[file.Name()]:
    return ft_ign
  case file.IsDir():
    return ft_dir
  case isImageName(file.Name()):
    return ft_img
  case isVideoName(file.Name()):
    return ft_vid
  default:
    return ft_ign
  }
}

func makeItem(file os.FileInfo) *store.Item {
  ts := TimeToProto(file.ModTime())
  name := file.Name()
  return &store.Item{
    Name: &name,
    FileTimestamp: &ts,
  }
}

func makeImage(file os.FileInfo) *store.Item {
  img := makeItem(file)
  img.Image = &store.Image{}
  return img
}

func makeVideo(file os.FileInfo) *store.Item {
  vid := makeItem(file)
  vid.Video = &store.Video{}
  return vid
}

func filterEntries(files []os.FileInfo) (imgs []*store.Item, vids []*store.Item, dirs []string) {
  for _, file := range files {
    switch fileType(file) {
    case ft_dir: dirs = append(dirs, file.Name())
    case ft_img: imgs = append(imgs, makeImage(file))
    case ft_vid: vids = append(vids, makeVideo(file))
    }
  }
  return
}

func mergeVideo(old_vid *store.Item, new_vid *store.Item) *store.Item {
  if old_vid != nil {
    new_vid.Keywords = old_vid.Keywords
    new_vid.ItemTimestamp = old_vid.ItemTimestamp
  }
  return new_vid
}

func mergeImage(old_img *store.Item, new_img *store.Item) *store.Item {
  if old_img != nil {
    if len(new_img.Keywords) == 0 {
      // Did not find keywords in the image, use old ones.
      new_img.Keywords = old_img.Keywords
    }
    if new_img.ItemTimestamp == nil {
      // Did not find item timestamp in image, use old one.
      new_img.ItemTimestamp = old_img.ItemTimestamp
    }
    new_img.Image = old_img.Image
  }
  return new_img
}

func UpdateDirectory(origd string, subs []os.FileInfo, sdir *store.Directory) error {
  old_itms := make(map[string]*store.Item, len(sdir.Items))
  for _, itm := range sdir.Items {
    old_itms[*itm.Name] = itm
  }
  imgs, vids, dirs := filterEntries(subs)
  sdir.SubDirectories = dirs

  new_items := make([]*store.Item, 0, len(imgs) + len(vids))

  for _, new_vid := range vids {
    old_vid, _ := old_itms[*new_vid.Name]
    new_items = append(new_items, mergeVideo(old_vid, new_vid))
  }

  for _, new_img := range imgs {
    old_img, ok := old_itms[*new_img.Name]
    if !ok || old_img.FileTimestamp == nil || 
      *old_img.FileTimestamp != *new_img.FileTimestamp {
      LoadImageFile(path.Join(origd, *new_img.Name), new_img)
    }
    new_items = append(new_items, mergeImage(old_img, new_img))
  }

  sdir.Items = new_items
  return nil
}
