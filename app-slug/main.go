package main

import (
	"log"

	toolkit "github.com/cmichels/buidling-a-module-go"
)

func main() {
  toSlug := "NOW!@#$!@ is the time 123"


  var tools toolkit.Tools


  if slug, err := tools.Slugify(toSlug); err != nil{
    log.Println("err: ", err)
  }else{
    log.Println("slug: ", slug)
  }
}
