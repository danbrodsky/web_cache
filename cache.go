package main

import (
	"net/http"
	"sync"
)

// these should exist within main proxy
var (
	cache_table = table{Map: make(map[string]page)}
	cache_policy = 0 // 0 = LRU, 1 = LFU
)

type table struct {
	lock sync.Mutex
	Map map[string]page
}

type page struct {
	url_data string
	filename string
	expires_at uint64
}

// Possible return type for responses sent back to main proxy
type response struct {
}

func initializeCache(policy int) {
	cache_policy = policy
	loadCacheFromDisk()
}

// goroutine function for a new request (called directly from proxy), also responsible for responding to client
func cacheWorker(w http.ResponseWriter) {

}

// loads cache from disk
func loadCacheFromDisk() {

}

// saves cache to disk
func writeCacheToDisk() {

}

// check for url data in cache, return page if present
func checkCache(url string) (avail bool, site page) {
	if entry, ok := cache_table.Map[url]; ok {
		return true, entry
	}
	return false, page{}
}

// update contents of cache based on last request
func updateCache() {
	writeCacheToDisk()
}

func deleteFromCache() {
	// delete from disk
	// clear cache entry
}

// get url data from cache by request
func retrieveFromCache(url string ) (err error) {

}
