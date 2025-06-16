package main

import (
	"fmt"
	"os"
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

// RsyncPair is a pair of files to sync, one from Lightroom and one to Google Drive.
type RsyncPair struct {
	lr_from string
	tt_to   string
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
			return dir
		}
	}
	return ""
}

func filesToSync(paths []string) (pairs []RsyncPair) {
	has_chosen_lr_root := false
	var lr_root string
	var g_root = findGRoot()
	if g_root == "" {
		println("None of the g_roots exist!")
		return
	} else {
		println("Using g_root", g_root)
	}
	for _, p := range paths {
		if !has_chosen_lr_root {
			for _, r := range lr_roots {
				if strings.HasPrefix(p, r) {
					lr_root = r
					has_chosen_lr_root = true
					println("Using lr_root", lr_root)
					break
				}
			}
			if !has_chosen_lr_root {
				println("Not under a known lr_root:", p)
				continue
			}
		} else {
			if !strings.HasPrefix(p, lr_root) {
				println("Not under chosen lr_root:", p)
				continue
			}
		}

		rel_p := p[len(lr_root):]
		d := path.Dir(rel_p)
		if path.Base(d) != "final" {
			println("Not in \"final\": ", p)
			continue
		}
		var lr_to = path.Join(g_root, path.Dir(d), path.Base(rel_p))
		pairs = append(pairs, RsyncPair{lr_from: p, tt_to: lr_to})
	}
	return
}

type StringSet map[string]struct{}

func (s StringSet) Add(element string) {
	s[element] = struct{}{}
}

func (s StringSet) Contains(element string) bool {
	_, exists := s[element]
	return exists
}

func sync(pairs []RsyncPair) error {
	var err error
	start_time := time.Now()

	var seenDirs = make(StringSet)

	// Prepare the directories.
	for _, p := range pairs {
		to_d := path.Dir(p.tt_to)
		if !seenDirs.Contains(to_d) {
			if Type == "F" {
				fmt.Printf("rm -r %s\n", to_d)
				if !dry_run {
					err = os.RemoveAll(to_d)
					if err != nil {
						return err
					}
				}
			}
			fmt.Printf("mkdir -p %s\n", to_d)
			if !dry_run {
				err = os.MkdirAll(to_d, 0755)
				if err != nil {
					return err
				}
			}
			seenDirs.Add(to_d)
		}
	}

	// Move the files in the directories.
	for _, p := range pairs {
		fmt.Printf("mv %s %s\n", p.lr_from, p.tt_to)
		if !dry_run {
			err = os.Rename(p.lr_from, p.tt_to)
			if err != nil {
				return err
			}
		}
	}

	fmt.Printf("sync... %vs\n", time.Since(start_time).Seconds())
	return nil
}

func main() {
	fmt.Printf("Run: dry = %v\n", dry_run)
	if !dry_run {
		os.Remove("/tmp/lrlog")
		logFile, _ := os.OpenFile("/tmp/lrlog", os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0644)
		syscall.Dup2(int(logFile.Fd()), 1)
		syscall.Dup2(int(logFile.Fd()), 2)
	}
	fmt.Printf("Args: %v\n", os.Args[1:])
	sync(filesToSync(os.Args[1:]))
}
