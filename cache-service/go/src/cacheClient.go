package main

import (
	"./cache"
	"net/http"
)

func main() {
	clientCache := cache.Cache{}
	clientCache.InitializeCache(0,2000000, 1000)
	server := &http.Server{
		Addr: ":8888",
		Handler: http.HandlerFunc(clientCache.ProcessRequest)}
	server.ListenAndServe()
}
