package main

import (
	// "fmt"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, req)
	})
}

func main() {
	servemux := http.NewServeMux()

	apiCfg := &apiConfig{}
	myFileServer := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	servemux.Handle("/app/", apiCfg.middlewareMetricsInc(myFileServer))

	servemux.HandleFunc("/healthz", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("OK"))
	})

	servemux.HandleFunc("/metrics", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		hitVal := apiCfg.fileServerHits.Load()
		hitString := "Hits: " + strconv.Itoa(int(hitVal))
		res.Write([]byte(hitString))
	})

	servemux.HandleFunc("/reset", func(res http.ResponseWriter, req *http.Request) {
		apiCfg.fileServerHits.Store(0)
		res.WriteHeader(http.StatusOK)
	})

	s := &http.Server{
		Addr:           ":8080",
		Handler:        servemux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}
