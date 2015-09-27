package main

import (
  "fmt"
  "os"
)

func check(p string, err error) {
  if err != nil {
    fmt.Printf("%20s: %v\n", p, err)
  }
}

func ls(p string) {
  file, err := os.Open(p)
  defer file.Close()
  if err != nil {
    check(p, err)
    return
  }
  var stat os.FileInfo
  stat, err = file.Stat()
  if err != nil {
    check(p, err)
    return
  }
  fmt.Printf("%20s: %v\n", p, stat.ModTime())
}

type Foo struct {
  a int
  b string
}

func foo() *Foo {
  return &Foo{a:1}
}

func main1() {
  for _, p := range os.Args[1:] {
    ls(p)
  }
}

func main() {
  fmt.Printf("%v\n", foo())
}
