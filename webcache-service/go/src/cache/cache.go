package cache

import (
	"../diskclient"
	"../page_scraper"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"unsafe"
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

	resEnv = "../res/"

	diskCache diskclient.DC

)

func (c Cache) InitializeCache(policy int, capacity int64, expiry int64) (err error) {
	if initFlag {
		return errors.New("cache already initialized")
	}
	initFlag = true

// 	cacheLock.Lock()

	cachePolicy = policy
	cacheCapacity = capacity
	expiryTime = expiry

	dc, err := diskclient.Initialize(cacheCapacity)
	if err != nil {return err}

	for url, site := range cacheTable {
		if !site.Safe {
			// fmt.Println("IVAN PLS ITS NOT WORTH IT")

			c.DeleteFromCache(url)
		}
	}

	diskCache = dc

// 	cacheLock.Unlock()
	fmt.Println("loading disk from cache")
	err = c.LoadCacheFromDisk()
	if err != nil {return err}
	return nil
}

// loads cache from disk
func (c Cache) LoadCacheFromDisk() (err error){

// 	cacheLock.Lock()
// 	defer cacheLock.Unlock()

	cache, err := diskCache.GetAllPages()
	if err != nil {
		return err
	}
	for _, page := range cache {
		cacheTable[page.Url] = &page
	}

	return nil
}

// saves cache to disk, set as unsafe
func (c Cache) WritePageToDisk(url string) (p diskclient.Page, err error){

// 	cacheLock.Lock()
// 	defer cacheLock.Unlock()

	cacheTable[url].Safe = false

	// remove previous page
	diskCache.DeletePage(url)
	diskCache.AddPage(*cacheTable[url])

	return *cacheTable[url], nil
}

func (c Cache) DeleteFromCache(url string) {

// 	cacheLock.Lock()
// 	defer cacheLock.Unlock()
	// fmt.Println("DELETING NOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO")

	diskCache.MarkUnsafe(url)

	for _, i := range cacheTable[url].Images {
		os.Remove(resEnv + i)
	}
	for _, l := range cacheTable[url].Links {
		os.Remove(resEnv + l)
	}
	for _, s := range cacheTable[url].Scripts {
		os.Remove(resEnv + s)
	}
	cacheCapacity -= cacheTable[url].Size

	delete(cacheTable,url)
	diskCache.DeletePage(url)
}

// check for url data in cache, return page if present
func (c Cache) CheckCache(url string) (avail bool, site diskclient.Page) {
	// TODO: lock here?
	// fmt.Println("checking cache for " + url)
	// fmt.Println(cacheTable)
	// fmt.Println(cacheTable[url])
	if entry, ok := cacheTable[url]; ok {
		if !entry.Safe || entry.Timestamp + expiryTime < int64(time.Now().UnixNano()) {
			// fmt.Println("MERCY ON US IVAN")
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
		// fmt.Println(site.Timestamp)
		// fmt.Println(expiryTime)
		// fmt.Println(int64(time.Now().UnixNano()))
		if site.Timestamp + expiryTime < int64(time.Now().UnixNano()) {
			// fmt.Println("DONT DO THIS IVAN")

			c.DeleteFromCache(url)
		}
	}
}

func (c Cache) UpdatePage(url string) (p diskclient.Page, err error){
// 	cacheLock.Lock()

	cacheTable[url].Timestamp = int64(time.Now().UnixNano())
	cacheTable[url].TimesUsed += 1

// 	cacheLock.Unlock()
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
// 	cacheLock.Lock()
	err = diskCache.MarkSafe(page.Url)
	if err != nil {
		return diskclient.Page{}, err
	}
	cacheTable[page.Url].Safe = true
// 	cacheLock.Unlock()

	return p, nil
}

func getPageSize(p diskclient.Page) (int64) {

	var size int64 = 0
	for _, i := range p.Images {
		fi, _ := os.Stat(i)
		size += fi.Size()
	}
	for _, l := range p.Links {
		fi, _ := os.Stat(l)
		size += fi.Size()
	}
	for _, s := range p.Scripts {
		fi, _ := os.Stat(s)
		size += fi.Size()
	}
	size += int64(unsafe.Sizeof(p.Html))

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

	avail, site := c.CheckCache(url)
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
	// TODO: add check in parser for page too large for cache to be returned immediately
	// TODO: update cache capacity
	if c.RemoveExpired(); cacheCapacity > cacheSize {
		if cachePolicy == 0 {
			err := c.LRU(site)
			if err != nil {
				// fmt.Println("error 5")
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
		} else {
			err := c.LFU(site)
			if err != nil {
				// fmt.Println("error 6")
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
		}
	}

	// write page in now that there's space
	cacheTable[url] = &completePage
	// fmt.Println("page resources loaded")
	completePage, err = c.WritePageToDisk(url)
	if err != nil {
		// fmt.Println("error 3")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	// successful write, mark as safe
	changePageToSafe(completePage)
	// fmt.Println("page written to disk")
	if err != nil {
		// fmt.Println("error 4")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	fmt.Println("result obtained first time")
	io.Copy(w, strings.NewReader(cacheTable[url].Html))
	return
}
func (c Cache) LRU(site diskclient.Page) (error) {
	if cacheCapacity < site.Size {
		return errors.New("size of page exceeds size of cache")
	}

	for cacheCapacity - site.Size < 0 {
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
		return errors.New("size of page exceeds size of cache")
	}

	for cacheCapacity - site.Size < 0 {
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