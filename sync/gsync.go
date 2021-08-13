package main

import ( "fmt"; "net/http"; "os"; "os/exec"; "path"; "strings"; "syscall"; "time" )

var lr_roots = []string{
  "/Users/matthieu/Pictures/Lightroom/Photos",
  "/Volumes/overflow/matthieu/Lightroom/Photos",
}
var g_root = "/Users/matthieu/Google Drive/Photos"

type RsyncPair struct {
  lr_from string
  tt_to string
}

func dirsToSync(paths []string) (pairs []RsyncPair) {
  dirs := make(map[string]bool)
  has_chosen_root := false
  var lr_root_used string
  for _, p := range paths {
    if !has_chosen_root {
      for _, r := range lr_roots {
        if strings.HasPrefix(p, r) {
          lr_root_used = r
          has_chosen_root = true
          break;
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

// For an incremental update do not pass --delete.
func sync(pair RsyncPair) {
	mkdir_args := []string {"-p", pair.tt_to}
  args := []string {
    "--recursive", "--relative", "--update", /*"--delete",*/ "--remove-source-files",
    "--perms", "--omit-dir-times=false", "--times", "--timeout", "600",
    pair.lr_from, pair.tt_to,
  }
  start_time := time.Now()
  if true {
		fmt.Printf("mkdir %v\n", strings.Join(mkdir_args, "' '"))
		exec.Command("mkdir", mkdir_args...).Run()
		fmt.Printf("rsync %v\n", strings.Join(args, "' '"))
    exec.Command("rsync", args...).Run()
    fmt.Printf("rsync... %vs\n", time.Since(start_time).Seconds())
  }
}

func main() {
  os.Remove("/tmp/lrlog")
  logFile, _ := os.OpenFile("/tmp/lrlog", os.O_WRONLY | os.O_CREATE | os.O_SYNC, 0644)
  syscall.Dup2(int(logFile.Fd()), 1)
  syscall.Dup2(int(logFile.Fd()), 2)
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
