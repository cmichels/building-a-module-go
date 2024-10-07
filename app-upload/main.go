package main

import (
	"fmt"
	"log"
	"net/http"

	toolkit "github.com/cmichels/buidling-a-module-go"
)


func main() {
  mux := routes()
  
  log.Println("starting port on 8080")


  if err := http.ListenAndServe(":8080", mux); err != nil{
    log.Fatalf("failed to start server. cause by [%s]",err)
  }

}



func routes() http.Handler {
  
  mux := http.NewServeMux()

  mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("."))))
  mux.HandleFunc("/upload", uploadlfiles)
  mux.HandleFunc("/upload-one", uploadOneFile)

  return mux
}


func uploadlfiles(w http.ResponseWriter, r *http.Request)  {
  
  if r.Method != "POST"{
    http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    return
  }


  t := toolkit.Tools{
    MaxFileSize: 1024 * 1024 * 1024,
    AllowedFileTypes: []string{"image/jpeg", "image/png", "image/gif"},
  }


  files, err := t.UploadFiles(r, "./uploads")
  if err != nil{
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  out := ""
  for _, item := range files {
    out += fmt.Sprintf("uploaded [%s]. renamed to [%s]", item.OriginalFileName, item.NewFileName)
  }


  _,_  = w.Write([]byte(out))

}


func uploadOneFile(w http.ResponseWriter, r *http.Request)  {
  
  if r.Method != "POST"{
    http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    return
  }


  t := toolkit.Tools{
    MaxFileSize: 1024 * 1024 * 1024,
    AllowedFileTypes: []string{"image/jpeg", "image/png", "image/gif"},
  }


  file, err := t.UploadOneFile(r, "./uploads")
  if err != nil{
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  out := fmt.Sprintf("uploaded [%s]. renamed to [%s]", file.OriginalFileName, file.NewFileName)

  _,_  = w.Write([]byte(out))


}
