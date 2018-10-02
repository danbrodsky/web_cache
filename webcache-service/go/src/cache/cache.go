package cache

import (
	"../diskclient"
	"../page_scraper"
	"errors"
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

	diskCache diskclient.DC

)

func (c Cache) InitializeCache(policy int, capacity int64, expiry int64) (err error) {
	if initFlag {
		return errors.New("cache already initialized")
	}
	initFlag = true

	cacheLock.Lock()

	cachePolicy = policy
	cacheCapacity = capacity
	expiryTime = expiry

	dc, err := diskclient.Initialize(cacheCapacity)
	if err != nil {return err}

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
	cacheCapacity -= cacheTable[url].Size

	delete(cacheTable,url)
	diskCache.DeletePage(url)
}

// check for url data in cache, return page if present
func (c Cache) CheckCache(url string) (avail bool, site diskclient.Page) {
	// TODO: lock here?

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
		} else if !site.Safe {
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

// complete new client request
func (c Cache) ProcessRequest(w http.ResponseWriter, req *http.Request) {

	avail, site := c.CheckCache(req.Host+req.URL.Path)
	if avail {
		// in cache, return page
		updatedPage, err := c.UpdatePage(req.Host+req.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		// TODO: Add header write
		io.Copy(w, strings.NewReader(updatedPage.Html))
		return
	}
	// not in memory
	// TODO: send request to parser

	pageScraper := page_scraper.NewPageScraper(req.Host+req.URL.Path)
	newPage, err := pageScraper.GetPage() // assume this gives page stub

	newPage.TimesUsed = 1
	newPage.Timestamp = int64(time.Now().UnixNano())

	// TODO: lock here
	cacheTable[req.Host+req.URL.Path] = &newPage
	// write stub page in for now
	p, err := c.WritePageToDisk(req.Host+req.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}


	completePage, err := pageScraper.ScrapePage(newPage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// TODO: lock here
	completePage.Size = getPageSize(completePage)
	cacheTable[req.Host+req.URL.Path] = &completePage


	p, err = c.WritePageToDisk(req.Host+req.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	// successful write, mark as safe
	changePageToSafe(p)

	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	cacheLock.Lock()
	cacheTable[req.Host+req.URL.Path] = &newPage



	// TODO: add check in parser for page too large for cache to be returned immediately
	// TODO: update cache capacity
	if c.RemoveExpired(); cacheCapacity > cacheSize {
		if cachePolicy == 0 {
			err := c.LRU(site)
			if err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
		} else {
			err := c.LFU(site)
			if err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
		}
	}
	// TODO: save in cacheTable
	p, err = c.WritePageToDisk(req.Host+req.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	// successful write, mark as safe
	changePageToSafe(p)

	// TODO: return to client
	//io.Copy(w, strings.NewReader(cacheTable[req.Host+req.URL.Path].Html))
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