package main

import (
	"fmt"

	toolkit "github.com/cmichels/buidling-a-module-go"
)


func main() {
  var tools toolkit.Tools


  s := tools.RandomString(10)


  fmt.Println("rando: ", s)
}
