package cache

import (
	"../../../../page_disk-service/go/src/disk_client"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Cache struct {
}

type C interface {
	InitializeCache(policy int, capacity int64, expiry int64) (error)

	LoadCacheFromDisk() (error)

	WritePageToDisk(url string) (p disk_client.Page, err error)

	CheckCache(url string) (avail bool, site disk_client.Page)

	DeleteFromCache(url string)

	RemoveExpired()

	ProcessRequest(w http.ResponseWriter, req *http.Request)

	UpdatePage(url string) (p disk_client.Page, err error)

	LRU(site disk_client.Page) (error)

	LFU(site disk_client.Page) (error)
}

// these should exist within main proxy (accessible everywhere)
var (
	initFlag = false
	cacheTable = make(map[string]*disk_client.Page)
	cachePolicy = 0 // 0 = LRU, 1 = LFU
	cacheLock sync.Mutex
	cacheSize int64
	cacheCapacity int64
	expiryTime int64

	diskCache disk_client.DC

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

	dc, err := disk_client.Initialize(cacheCapacity)
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

// saves cache to disk
func (c Cache) WritePageToDisk(url string) (p disk_client.Page, err error){

	cacheLock.Lock()
	defer cacheLock.Unlock()

	cacheTable[url].Safe = false

	// remove previous page
	diskCache.DeletePage(url)

	diskCache.AddPage(*cacheTable[url])

	// successful write, mark as safe
	err = diskCache.MarkSafe(url)

	if err != nil {
		return disk_client.Page{}, err
	}
	cacheTable[url].Safe = true
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
func (c Cache) CheckCache(url string) (avail bool, site disk_client.Page) {
	// TODO: lock here?

	if entry, ok := cacheTable[url]; ok {
		if !entry.Safe || entry.Timestamp + expiryTime < int64(time.Now().UnixNano()) {
			c.DeleteFromCache(url)
			return false, disk_client.Page{}
		}
		return true, *entry
	}
	return false, disk_client.Page{}
}

func (c Cache) RemoveExpired() {
	// TODO: lock here?
	for url, site := range cacheTable {
		if site.Timestamp + expiryTime < int64(time.Now().UnixNano()) {
			c.DeleteFromCache(url)
		}
	}
}

func (c Cache) UpdatePage(url string) (p disk_client.Page, err error){
	cacheLock.Lock()

	cacheTable[url].Timestamp = int64(time.Now().UnixNano())
	cacheTable[url].TimesUsed += 1

	cacheLock.Unlock()
	return c.WritePageToDisk(url)

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
	//pageScraper := page_scraper.NewPageScraper(req.Host+req.URL.Path)
	//newPage, err := pageScraper.Execute()
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusServiceUnavailable)
	//	return
	//}
	//
	//cacheLock.Lock()
	//cacheTable[req.Host+req.URL.Path] = &newPage

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
	c.WritePageToDisk(req.Host+req.URL.Path)
	// TODO: return to client
	//io.Copy(w, strings.NewReader(cacheTable[req.Host+req.URL.Path].Html))
	return
}

func (c Cache) LRU(site disk_client.Page) (error) {
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

func (c Cache) LFU(site disk_client.Page) (error) {
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