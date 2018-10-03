package main

import (
    "fmt"
    "net/http"
    "os"
    //"page_scraper"
    "strings"
    "cache"
    "crypto/sha256"
    "path/filepath"
    "encoding/base64"
    //"compress/gzip"
    //"bytes"
    "io"
    //"io/ioutil"
)

var (
    clientCache cache.Cache
)



func handleHTTP(w http.ResponseWriter, req *http.Request) {
    isCached , _:= clientCache.CheckCache( "http://"+req.URL.Host + "/")

    if(strings.Contains(req.URL.String(),os.Getenv("DEPLOY_HOST_IP")) || strings.Contains(req.URL.String(),"127.0.0.1" )){
	fmt.Println(req.URL.String()+ "files")
	http.StripPrefix("/res/", http.FileServer(http.Dir("/root/res"))).ServeHTTP(w,req)
	fmt.Println("file request")
	return
    } else if(isMasked(req.URL.String())){
	handleFullProxy(w,req)
	return
    } else if(len(filepath.Ext(req.URL.String())) > 1 && isCached){
	oldPath := req.URL.Path
	req.URL.Path = encodeUrlFilePath(req.URL.String())
	if _, err := os.Stat(os.Getenv("RES_ROOT_DIR")+os.Getenv("RES_ENTRYPOINT") + req.URL.Path); os.IsNotExist(err) {
		req.URL.Path = oldPath
		fmt.Println(req.URL.String()+ " requesting to cache!!")
                clientCache.ProcessRequest(w,req)
	} else {
		fmt.Println(req.URL.String()+ " requesting from files server")
                http.FileServer(http.Dir(os.Getenv("RES_ROOT_DIR")+os.Getenv("RES_ENTRYPOINT"))).ServeHTTP(w,req)
	}
	return
    } else {
        fmt.Println(req.URL.String()+ " requesting to cache")
        clientCache.ProcessRequest(w,req)
    }

}

func encodeUrlFilePath(url string)(string){
	h := sha256.New()
        h.Write([]byte(url))
        sha := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return  "/"+filepath.Ext(url)[1:len(filepath.Ext(url))] + "/" + sha + filepath.Ext(url)
}

func isMasked(url string)(bool){
	if(strings.Contains(url,"detectportal.firefox.")|| strings.Contains(url,"https")){
		return true
	}
	return false

}

func handleFullProxy(w http.ResponseWriter, req *http.Request) {
    resp, err := http.DefaultTransport.RoundTrip(req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }
    defer resp.Body.Close()
    copyHeader(w.Header(), resp.Header)
    w.WriteHeader(resp.StatusCode)
    io.Copy(w, resp.Body)
}


func copyHeader(dst, src http.Header) {
    for k, vv := range src {
        for _, v := range vv {
            dst.Add(k, v)
        }
    }
}

func main() {
    clientCache = cache.Cache{}
    clientCache.InitializeCache(0,100000,10000000000)
    http.Handle( os.Getenv("RES_ENTRYPOINT") +"/", http.StripPrefix( os.Getenv("RES_ENTRYPOINT") +"/", http.FileServer(http.Dir(os.Getenv("RES_ROOT_DIR")+os.Getenv("RES_ENTRYPOINT")))))
    //http.ListenAndServe(":" + "8000", nil)
    server := &http.Server{
                        Addr: ":8888",
                        Handler: http.HandlerFunc(handleHTTP)}
    server.ListenAndServe()
}
