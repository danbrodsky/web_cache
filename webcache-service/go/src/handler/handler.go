package main

import (
    "fmt"
    "net/http"
    "os"
    "page_scraper"
    "compress/gzip"
    "bytes"
    //"io"
    //"io/ioutil"
)


func about_handler(w http.ResponseWriter, r *http.Request) {
    // ABOUT SECTION HTML CODE
    page,_ := page_scraper.NewPageScraper("http://vaastavanand.com/").Execute()
    fmt.Fprintf(w, page.Html)
}

func handleHTTP(w http.ResponseWriter, req *http.Request) {
    resp, err := http.DefaultTransport.RoundTrip(req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }
    fmt.Println(resp.Header)
    var b bytes.Buffer
    page,_ := page_scraper.NewPageScraper("http://vaastavanand.com/").Execute()
    gz := gzip.NewWriter(&b)
    if _, err := gz.Write([]byte(page.Html)); err != nil {
        panic(err)
    }
    if err := gz.Flush(); err != nil {
        panic(err)
    }
    if err := gz.Close(); err != nil {
        panic(err)
    }

    defer resp.Body.Close()
    copyHeader(w.Header(), resp.Header)
    w.WriteHeader(resp.StatusCode)
    w.Write(b.Bytes())
    //body, _ := ioutil.ReadAll(resp.Body)
    //bodyString := string(body)
    //fmt.Println(bodyString)
    //io.Copy(w, resp.Body)
}

func copyHeader(dst, src http.Header) {
    for k, vv := range src {
        for _, v := range vv {
            dst.Add(k, v)
        }
    }
}

func main() {
    http.HandleFunc("/about/", about_handler)
    fs_entrypoint := os.Getenv("RES_ROOT_DIR")
    fmt.Println(fs_entrypoint)
    http.Handle("/res/", http.StripPrefix("/res/", http.FileServer(http.Dir("/root/res"))))
    //http.ListenAndServe(":" + os.Getenv("DEPLOY_HOST_PORT"), nil)
    server := &http.Server{
                        Addr: ":8888",
                        Handler: http.HandlerFunc(handleHTTP)}
    server.ListenAndServe()
}
