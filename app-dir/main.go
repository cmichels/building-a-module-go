package main

import (
	toolkit "github.com/cmichels/buidling-a-module-go"
)

func main() {
  var tools toolkit.Tools


  tools.CreateDirIfNotExists("./test-dir")
}
