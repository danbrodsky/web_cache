package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

// these should exist within main proxy (accessible everywhere)
var (
	cacheTable = make(map[string]page)
	cachePolicy = 0 // 0 = LRU, 1 = LFU
	cachePath = "./cache/"
	cacheLock sync.Mutex
	cacheSize uint64
	cacheCapacity uint64

	Marshal = func(i interface{}) (io.Reader, error) {
		data, err := json.MarshalIndent(i, "", "\t")
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(data), nil
	}

	Unmarshal = func(r io.Reader, i interface{}) (error) {
		return json.NewDecoder(r).Decode(i)
	}
)

type page struct {
	urlData string
	filename string
	pageSize uint64
	expiresAt uint64
	timesAccessed uint64
	safe bool // boolean to keep track of if page finished being stored
	}

func initializeCache(policy int, capacity uint64) {
	cacheLock.Lock()
	defer cacheLock.Unlock()
	cachePolicy = policy
	cacheCapacity = capacity
	loadCacheFromDisk()
}

// loads cache from disk
func loadCacheFromDisk() (error){

	file, err := os.Open(cachePath + "cacheTable")
	if err != nil {
		return err
	}
	defer file.Close()
	return Unmarshal(file,cacheTable)
	}

// saves cache to disk
func writeCacheToDisk() (error){

	file, err := os.Create(cachePath + "cacheTable")
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := Marshal(cacheTable)
	if err != nil {
		return err
	}

	_, err = io.Copy(file,data)
	return err

}

// check for url data in cache, return page if present
func checkCache(url string) (avail bool, site page) {

	cacheLock.Lock()
	defer cacheLock.Unlock()

	if entry, ok := cacheTable[url]; ok {
		if !entry.safe || entry.expiresAt < uint64(time.Now().UnixNano()) {
			deleteFromCache(url)
			return false, page{}
		}
		return true, entry
	}
	return false, page{}
}

func deleteFromCache(url string) {

	delete(cacheTable,url)
	writeCacheToDisk()
}

func removeExpired() {

	cacheLock.Lock()
	defer cacheLock.Unlock()

	for url, site := range cacheTable {
		if site.expiresAt < uint64(time.Now().UnixNano()) {
			cacheCapacity -= site.pageSize
			deleteFromCache(url)
		}
	}
}

// complete new client request
func processRequest(url string) (error) {

	avail, site := checkCache(url)
	if avail {
		// TODO: return to client
		// TODO: increment timesAccessed, set expiresAt again
		return nil
	}
	// not in memory
	// TODO: send request to parser
	// TODO: update cache capacity
	if removeExpired(); cacheCapacity > cacheSize {
		if cachePolicy == 0 {
			err := LRU(url, site)
			if err != nil {
				return err
			}
		} else {
			err := LFU(url, site)
			if err != nil {
				return err
			}
		}
	}
	// TODO: save in cacheTable
	// TODO: return to client
	writeCacheToDisk()
	return nil
}

// TODO: Implement cache policies
func LRU(url string, site page) (error) {
	return nil
}

func LFU(url string, site page) (error) {
	return nil
}