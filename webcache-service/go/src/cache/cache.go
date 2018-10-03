package cache

import (
	"diskclient"
	"page_scraper"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"crypto/sha256"
	"path/filepath"
        "encoding/base64"
)

type Cache struct {
}

type C interface {
	InitializeCache(policy int, capacity int64, expiry int64) (error)

	LoadCacheFromDisk() (error)

	WritePageToDisk(url string) (p diskclient.Page, err error)

	CheckCache(url string) (avail bool, site diskclient.Page)

	DeleteFromCache(url string)

	RemoveExpired()

	ProcessRequest(w http.ResponseWriter, req *http.Request)

	UpdatePage(url string) (p diskclient.Page, err error)

	LRU(site diskclient.Page) (error)

	LFU(site diskclient.Page) (error)
}

// these should exist within main proxy (accessible everywhere)
var (
	initFlag = false
	cacheTable = make(map[string]*diskclient.Page)
	cachePolicy = 0 // 0 = LRU, 1 = LFU
	cacheLock sync.Mutex
	cacheSize int64
	cacheCapacity int64
	expiryTime int64
	ROOTDIR string

	resEnv = "../res/"

	diskCache diskclient.DC

)

func (c Cache) InitializeCache(policy int, capacity int64, expiry int64) (err error) {
	if initFlag {
		return errors.New("cache already initialized")
	}
	initFlag = true
	ROOTDIR = os.Getenv("RES_ROOT_DIR")
   	cacheLock.Lock()

	cachePolicy = policy
	cacheCapacity = capacity
	expiryTime = expiry * 1000000000

	dc, err := diskclient.Initialize(cacheCapacity)
	if err != nil {return err}

	for url, site := range cacheTable {
		if !site.Safe {
			c.DeleteFromCache(url)
		}
	}

	diskCache = dc

   	cacheLock.Unlock()

	err = c.LoadCacheFromDisk()
	if err != nil {return err}
	return nil
}

// loads cache from disk
func (c Cache) LoadCacheFromDisk() (err error){

   	cacheLock.Lock()
   	defer cacheLock.Unlock()

	cache, err := diskCache.GetAllPages()
	if err != nil {
		return err
	}
	for _, page := range cache {
		cacheTable[page.Url] = &page
	}
	fmt.Println("cache loaded from disk")

	return nil
}

// saves cache to disk, set as unsafe
func (c Cache) WritePageToDisk(url string) (p diskclient.Page, err error){

   	cacheLock.Lock()
   	defer cacheLock.Unlock()

	cacheTable[url].Safe = false

	// remove previous page
	diskCache.DeletePage(url)
	diskCache.AddPage(*cacheTable[url])

	return *cacheTable[url], nil
}

func (c Cache) DeleteFromCache(url string) {

   	cacheLock.Lock()
   	defer cacheLock.Unlock()


	diskCache.MarkUnsafe(url)

	for _, i := range cacheTable[url].Images {
		fmt.Println("deleting!!" + encodeUrlFilePath(i))
		os.Remove(encodeUrlFilePath(i))
	}
	for _, l := range cacheTable[url].Links {
		fmt.Println("deleting!!"  + encodeUrlFilePath(l))
		os.Remove(encodeUrlFilePath(l))
	}
	for _, s := range cacheTable[url].Scripts {
		fmt.Println("deleting!!"  + encodeUrlFilePath(s))
		os.Remove(encodeUrlFilePath(s))
	}


	cacheSize -= cacheTable[url].Size

	delete(cacheTable,url)
	diskCache.DeletePage(url)
}

func encodeUrlFilePath(uri string)(string){
        if(strings.Contains(uri,"http")){
		fmt.Println(uri)
		h := sha256.New()
		h.Write([]byte(uri))
		sha := base64.URLEncoding.EncodeToString(h.Sum(nil))
		if(len(filepath.Ext(uri)) <= 1){
			return uri
		}
		return  ROOTDIR + os.Getenv("RES_ENTRYPOINT") + "/"+filepath.Ext(uri)[1:len(filepath.Ext(uri))] + "/" + sha + filepath.Ext(uri)
	}
        return  ROOTDIR + uri
}

// check for url data in cache, return page if present
func (c Cache) CheckCache(url string) (avail bool, site diskclient.Page) {
	// TODO: lock here?
	// fmt.Println("checking cache for " + url)

	if entry, ok := cacheTable[url]; ok {
		if !entry.Safe || entry.Timestamp + expiryTime < int64(time.Now().UnixNano()) {
			c.DeleteFromCache(url)
			return false, diskclient.Page{}
		}
		return true, *entry
	}
	return false, diskclient.Page{}
}

func (c Cache) RemoveExpired() {
	// TODO: lock here?
	for url, site := range cacheTable {



		if site.Timestamp + expiryTime < int64(time.Now().UnixNano()) {


			c.DeleteFromCache(url)
		}
	}
}

func (c Cache) UpdatePage(url string) (p diskclient.Page, err error){
   	cacheLock.Lock()

	cacheTable[url].Timestamp = int64(time.Now().UnixNano())
	cacheTable[url].TimesUsed += 1

   	cacheLock.Unlock()

	p, err = c.WritePageToDisk(url)
	if err != nil {
		return diskclient.Page{}, err
	}
	// successful write, mark as safe
	p, err = changePageToSafe(p)
	if err != nil {
		return diskclient.Page{}, err
	}

	return p, nil

}

func changePageToSafe(page diskclient.Page) (p diskclient.Page, err error) {
   	cacheLock.Lock()
	err = diskCache.MarkSafe(page.Url)
	if err != nil {
		return diskclient.Page{}, err
	}
	cacheTable[page.Url].Safe = true
   	cacheLock.Unlock()

	return p, nil
}

func getPageSize(p diskclient.Page) (int64) {
	var size int64 = 0
	for _, i := range p.Images {
		fmt.Println("getting page size !!!!!!! " + encodeUrlFilePath(i))
		fi, _ := os.Stat( encodeUrlFilePath(i))
		size += fi.Size()
	}
	for _, l := range p.Links {
		fi, _ := os.Stat( encodeUrlFilePath(l))
		size += fi.Size()
	}
	for _, s := range p.Scripts {
		fi, _ := os.Stat( encodeUrlFilePath(s))
		size += fi.Size()
	}
	size += int64(len(p.Html))

	return size
}

func blacklisted(url string) (blacklisted bool){
	if strings.Contains(url, "firefox") ||
		strings.Contains(url, "mozilla") ||
		strings.Contains(url, "google"){
		return true
	}
	return false
}

func (c Cache) createNewPage(url string) (page diskclient.Page, err error) {
	pageScraper := page_scraper.NewPageScraper(url)
	newPage, err := pageScraper.GetPage()

	newPage.TimesUsed = 1
	newPage.Timestamp = int64(time.Now().UnixNano())

	// TODO: lock here
	cacheTable[url] = &newPage
	// write stub page in for now

	newPage, err = c.WritePageToDisk(url)
	if err != nil {
		return diskclient.Page{}, err
	}


	completePage, err := pageScraper.ScrapePage(newPage)
	if err != nil {
		return diskclient.Page{}, err
	}
	completePage.Timestamp = newPage.Timestamp
	completePage.TimesUsed = newPage.TimesUsed
	completePage.Size = getPageSize(completePage)

	return completePage, nil
}

// complete new client request
func (c Cache) ProcessRequest(w http.ResponseWriter, req *http.Request) {

	url := "http://"+req.Host+req.URL.Path

	if blacklisted(url) {
		return
	}

	avail, _ := c.CheckCache(url)
	if avail {
		// in cache, return page
		_, err := c.UpdatePage(url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		// TODO: Add header write
		fmt.Println("result obtained from cache")
		io.Copy(w, strings.NewReader(cacheTable[url].Html))
		return
	}

	// not in memory, create a new page
	completePage, err := c.createNewPage(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	cacheSize += completePage.Size

	fmt.Println("current cache size")
	fmt.Println(cacheSize)
	fmt.Println("current page size")
	fmt.Println(completePage.Size)

	if c.RemoveExpired(); cacheCapacity < cacheSize {
		fmt.Println("capacity exceeded")
		if cachePolicy == 0 {
			err := c.LRU(completePage)
			if err != nil {
				// cache too small, return page
				fmt.Println("cache too small")
				io.Copy(w, strings.NewReader(completePage.Html))
				return
			}
		} else {

			err := c.LFU(completePage)
			if err != nil {
				// cache too small, return page
				fmt.Println("cache too small")
				io.Copy(w, strings.NewReader(completePage.Html))
				return
			}
		}
	}

	// write page in now that there's space
	cacheTable[url] = &completePage

	completePage, err = c.WritePageToDisk(url)
	if err != nil {

		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	// successful write, mark as safe
	changePageToSafe(completePage)
	// fmt.Println("page written to disk")
	if err != nil {

		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	fmt.Println("result obtained first time")
	io.Copy(w, strings.NewReader(completePage.Html))
	return
}
func (c Cache) LRU(site diskclient.Page) (error) {

	if cacheCapacity < site.Size {
		cacheSize -= site.Size
		return errors.New("size of page exceeds size of cache")
	}

	for cacheCapacity - cacheSize < 0 {
		oldest := int64(time.Now().UnixNano())
		oldestUrl := ""
		for url, page := range cacheTable {
			if page.Timestamp < oldest {
				oldest = page.Timestamp
				oldestUrl = url
			}
		}
		c.DeleteFromCache(oldestUrl)
	}
	return nil
}

func (c Cache) LFU(site diskclient.Page) (error) {


	if cacheCapacity < site.Size {
		cacheSize -= site.Size
		return errors.New("size of page exceeds size of cache")
	}

	for cacheCapacity - cacheSize < 0 {
		leastUsed := int64(10000)
		leastUsedUrl := ""
		for url, page := range cacheTable {
			if page.TimesUsed < leastUsed {
				leastUsed = page.TimesUsed
				leastUsedUrl = url
			}
		}
		c.DeleteFromCache(leastUsedUrl)
	}
	return nil
}
