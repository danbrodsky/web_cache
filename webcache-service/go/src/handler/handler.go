package main

import (
    "fmt"
    "net/http"
    "os"
)


func about_handler(w http.ResponseWriter, r *http.Request) {
    // ABOUT SECTION HTML CODE
    fmt.Fprintf(w, "<title>Go/about/</title>")
    fmt.Fprintf(w, "Expert web design by JT Skrivanek")
}

func main() {
  //  http.HandleFunc("/about/", about_handler)
    fs_entrypoint = os.Getenv("RES_ROOT_DIR")
    http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("../root/res"))))
    http.ListenAndServe(":8000", nil)
}
