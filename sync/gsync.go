package main

import ( 
  "fmt"
  "net/http"
  "os"
  "os/exec"
  "path"
  "strings"
  "syscall"
  "time" 
)

var lr_roots = []string{
  "/Users/matthieu/Pictures/Photos",
  "/Users/matthieu/Pics from Bernard/",
  // "/Volumes/overflow/matthieu/Lightroom/Photos",
}

var g_roots = []string{
  "/Users/matthieu/Google Drive/My Drive/Photos",
  "/Users/matthieu/Google Drive/Mon Drive/Photos"}

var dry_run = false
// F: full (deletes previous destination contents)
// I: incremental (keeps previous destination contents)
var Type = "F"

type RsyncPair struct {
  lr_from string
  tt_to string
}

func directoryExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func findGRoot() string {
	for _, dir := range g_roots {
		if directoryExists(dir) {
      return dir;
		}
	}
  return "";
}

func dirsToSync(paths []string) (pairs []RsyncPair) {
  dirs := make(map[string]bool)
  has_chosen_root := false
  var lr_root_used string
  var g_root = findGRoot();
  if g_root == "" {
    println("None of the g_roots exist!");
    return;
  }
  for _, p := range paths {
    if !has_chosen_root {
      for _, r := range lr_roots {
        if strings.HasPrefix(p, r) {
          lr_root_used = r
          has_chosen_root = true
          break
        }
      }
      if !has_chosen_root {
        println("Not under a known lr_root:", p)
        continue
      }
    } else {
      if !strings.HasPrefix(p, lr_root_used) {
        println("Not under lr_root:", p)
        continue
      }
    }

    p = p[len(lr_root_used):]
    d := path.Dir(p)
    if path.Base(d) != "final"{
      println("Not in \"final\": ", p)
      continue
    }
    _, ok := dirs[path.Dir(d)]
    if !ok {
      dirs[path.Dir(d)] = true
      pairs = append(pairs, RsyncPair{
        // Here, /./ marks the start of the relative path to use in the copy.
        lr_from:path.Join(lr_root_used, d) + "/./",
        tt_to:path.Join(g_root, path.Dir(d))})
    }
  }
  return
}

func quoteShell(s string) string {
  s = strings.Replace(s, " ", "\\ ", -1)
  s = strings.Replace(s, "&", "\\&", -1)
  return s
}

func sync(pair RsyncPair) {
	mkdir_args := []string {"-p", pair.tt_to}
  args := []string {
    "--recursive", "--relative", "--update", "--remove-source-files",
    "--perms", "--omit-dir-times=false", "--times", "--timeout", "600",
  }
  if (Type == "F") {
    args = append(args, "--delete")
  }
  args = append(args, pair.lr_from)
  args = append(args, pair.tt_to)
  start_time := time.Now()
  fmt.Printf("mkdir %v\n", strings.Join(mkdir_args, "' '"))
  if !dry_run {
		exec.Command("mkdir", mkdir_args...).Run()
	}	
  fmt.Printf("rsync '%v'\n", strings.Join(args, "' '"))
  if !dry_run {
    exec.Command("rsync", args...).Run()
  }
  fmt.Printf("rsync... %vs\n", time.Since(start_time).Seconds())
}

func main() {
  if !dry_run {
    os.Remove("/tmp/lrlog")
    logFile, _ := os.OpenFile("/tmp/lrlog", os.O_WRONLY | os.O_CREATE | os.O_SYNC, 0644)
    syscall.Dup2(int(logFile.Fd()), 1)
    syscall.Dup2(int(logFile.Fd()), 2)
  }
  fmt.Printf("Args: %v\n", os.Args[1:])
  pairs := dirsToSync(os.Args[1:])
  for _, p := range pairs {
    sync(p)
  }
  if false {
    fmt.Printf("reloading...\n")
    start_time := time.Now()
    http.Get("http://icute.local/db/viewer?command=reload");
    fmt.Printf("reloading... %vs\n", time.Since(start_time).Seconds())
  }
}
