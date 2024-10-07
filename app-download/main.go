package main

import (
	"log"
	"net/http"

	toolkit "github.com/cmichels/buidling-a-module-go"
)



func main() {
  
  mux := routes()

  err := http.ListenAndServe(":8080", mux)

  if err != nil {
    log.Fatal(err)
  }
  
}


func routes() http.Handler {
  mux := http.NewServeMux()

  mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("."))))
  mux.HandleFunc("/download", downloadFile)

  return mux
}


func downloadFile(w http.ResponseWriter, r *http.Request)  {
  
  var tools toolkit.Tools

  tools.DownloadStaticFile(w, r, "./files", "pic.jpg", "puppy.jpg")
}
