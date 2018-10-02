package page_scraper

import (
    "errors"
    "io"
    "net/http"
    "encoding/base64"
    "os"
    "fmt"
    "crypto/sha256"
    "path/filepath"
)


type ResourceScraper struct {
}



type RS interface {
	ScrapeResource(url string) (fn string, e error)
}

func NewResourceScraper() (resourceScraper RS) {
	resourceScraper = ResourceScraper{}
	return resourceScraper
}

func (rs ResourceScraper) ScrapeResource(url string) (rn string, e error){
    //fileUrl := "https://www.cs.ubc.ca/~wolf/pics/fullsize2.jpg"
    var ext = filepath.Ext(url)
    if(len(ext) < 1){
	return rn, errors.New("no extension found")
    }
    path := ext[1:len(ext)]
    fmt.Println(path)
    os.MkdirAll(path, os.ModePerm)
    rn,err := DownloadFile(path, ext, url)
    if err != nil {
        return "",err
    }
    return rn,nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(fp string, ext string, url string) (uri string, erro error) {

    // Get the data
    resp, err := http.Get(url)
    if err != nil {
        return uri,err
    }
    defer resp.Body.Close()
    if resp.StatusCode == http.StatusOK {
        h := sha256.New()
        h.Write([]byte(url))
	sha := base64.URLEncoding.EncodeToString(h.Sum(nil))
        uri = filepath.Join(fp, sha+ext)
	fmt.Println(uri)
    }

    // Create the file
    out, err := os.Create(uri)
    if err != nil {
         return uri,err
    }
    defer out.Close()

    // Write the body to file
    _, err = io.Copy(out, resp.Body)
    if err != nil {
        return uri,err
    }

    return uri,nil
}
