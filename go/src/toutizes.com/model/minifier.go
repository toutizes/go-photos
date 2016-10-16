

package model

import (
  "flag"
  "fmt"
  "io/ioutil"
  "log"
  "os"
  "os/exec"
  "path"
	"strings"
  "time"
)

var minifier_threads = flag.Int("minifier_threads", 4, "Number of threads for minifying.")


func indexTimes(dir string) (m map[string]time.Time) {
  fis, err :=  ioutil.ReadDir(dir)
  if err != nil {
    os.MkdirAll(dir, 0777)
    return
  }
  m = make(map[string]time.Time, len(fis))
  for _, fi := range fis {
    m[fi.Name()] = fi.ModTime()
  }
  return
}

func mustScale(img *Image, mini_times map[string]time.Time) bool {
  mini_time, ok := mini_times[img.Name()]
  return !ok || mini_time.Before(img.FileTime())
}

func doScaleN(in_dir string, imgs []*Image, size int, out_dir string) {
  args := make([]string, 0, len(imgs) + 10)
  args = append(args, []string{
    *BinRoot + "mogrify",
    "-resize", fmt.Sprintf("%dx%d", size, size),
    "-quality", "90",
    "-path", out_dir }...)
	if (size > 1000) {
    args = append(args, []string{ "-interlace", "Plane"}...)
	}
  for _, img := range imgs {
    args = append(args, path.Join(in_dir, img.Name()))
  }
  cmd := exec.Command(args[0], args[1:]...)
  output, err := cmd.CombinedOutput()
  if err != nil {
    log.Printf("doScaleN failed (%s): %s\n", strings.Join(args, " "), err)
    log.Printf("Full output: %s\n", string(output))
  }
}

func minify(db *Database, dir *Directory) int {
  start_time := time.Now()
  rel_pat := dir.RelPat()
  mini_dir := db.FullMiniPath(rel_pat)
  mini_times := indexTimes(mini_dir)
  midi_dir := db.FullMidiPath(rel_pat)
  midi_times := indexTimes(midi_dir)
  orig_dir := db.FullOrigPath(rel_pat)
  num_resized := 0
  var to_minis []*Image
  var to_midis []*Image
  for _, img := range dir.Images() {
    if mustScale(img, mini_times) {
      to_minis = append(to_minis, img)
      num_resized += 1 
    }
    if mustScale(img, midi_times) {
      to_midis = append(to_midis, img)
      num_resized += 1 
    }
  }
  if to_minis != nil {
    doScaleN(orig_dir, to_minis, 180, mini_dir)
  }
  if to_midis != nil {
    doScaleN(orig_dir, to_midis, 2048, midi_dir)
  }
  if num_resized > 0 {
    log.Printf("%s: resized %d in %d ms\n", rel_pat, num_resized,
      time.Since(start_time).Nanoseconds() / 1000000)
  } 
  return num_resized
}

func minifyWorker(db *Database, dir_ch <- chan *Directory, res_ch chan <- int) {
  for dir := range dir_ch {
    res_ch <-minify(db, dir)
  }
}

// rm -rf /tmp/mini/2005-01*
// 8: Minifed 229 images in 21472 ms
// 8:   mogri 229 images in 13872 ms
// 4: Minifed 229 images in 20249 ms
// 4:   mogri 229 images in 14058 ms
// 2: Minifed 229 images in 30756 ms
// rm -rf /tmp/mini/2005-0{1,2,3,4}*
// 8: Minifed 927 images in 68639 ms
// 8:   mogri 927 images in 47370
// 6:   mogri 927 images in 43174 ms
// 4: Minifed 927 images in 81521 ms
// 4:   mogri 927 images in 49567 ms


func MinifyDatabase(db *Database) int {
  N := *minifier_threads;
  dir_ch := make(chan *Directory, N)
  res_ch := make(chan int, N)
  for i := 0; i < N; i++ {
    go minifyWorker(db, dir_ch, res_ch)
  }
  dirs := db.Directories()
  left := len(dirs)
  pushed := 0
  minified := 0
  for ; left > 0; {
    has_pushed := false
    if pushed < len(dirs) {
      select {
      case dir_ch <- dirs[pushed]:
        pushed += 1
        has_pushed = true
      default:
      }
    }
    if !has_pushed {
      minified += <- res_ch
      left -= 1
    }
  }
  close(dir_ch)
  close(res_ch)
  return minified
}
