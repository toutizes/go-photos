package main

import (
  "fmt"
  "net/http"
  "net/url"
)

func main() {
  resp, err :=
    http.PostForm("http://pubsubhubbub.appspot.com",
    url.Values{
      "hub.url":{"http://toutizes.com/db/f"},
      "hub.mode":{"publish"}})
  if err != nil {
    println(err.Error())
  }    
  fmt.Printf("%v\n", resp)
}
