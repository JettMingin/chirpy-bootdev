package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq"
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

func (cfg *apiConfig) metricsHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html")
	resHTML := `
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>
	`
	hitVal := cfg.fileServerHits.Load()
	res.WriteHeader(http.StatusOK)
	res.Write(fmt.Appendf(nil, resHTML, hitVal))
}

func validateChirpHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	type Chirp struct {
		Body string `json:"body"`
	}
	newChirp := Chirp{}
	if err := json.NewDecoder(req.Body).Decode(&newChirp); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		res.WriteHeader(500)
		return
	}

	type errorRes struct {
		Reserr string `json:"error"`
	}
	type successRes struct {
		CleanedBody string `json:"cleaned_body"`
	}

	if len(newChirp.Body) > 140 || len(newChirp.Body) < 1 {
		resBody := errorRes{
			Reserr: "invalid Chirp length",
		}
		data, err := json.Marshal(resBody)
		if err != nil {
			log.Printf("Error Mashalling Response JSON: %s", err)
			res.WriteHeader(500)
			return
		}
		res.WriteHeader(400)
		res.Write(data)
		return
	}

	var profanityRegexes = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bkerfuffle\b`),
		regexp.MustCompile(`(?i)\bsharbert\b`),
		regexp.MustCompile(`(?i)\bfornax\b`),
	}
	cleanedString := newChirp.Body
	for _, cussRegEx := range profanityRegexes {
		cleanedString = cussRegEx.ReplaceAllString(cleanedString, "****")
	}
	resBody := successRes{
		CleanedBody: cleanedString,
	}
	data, err := json.Marshal(resBody)
	if err != nil {
		log.Printf("Error Mashalling Response JSON: %s", err)
		res.WriteHeader(500)
		return
	}
	res.WriteHeader(200)
	res.Write(data)
}

func main() {
	servemux := http.NewServeMux()

	apiCfg := &apiConfig{}
	myFileServer := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	servemux.Handle("/app/", apiCfg.middlewareMetricsInc(myFileServer))

	servemux.HandleFunc("GET /api/healthz", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("OK"))
	})
	servemux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)

	servemux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)

	servemux.HandleFunc("POST /admin/reset", func(res http.ResponseWriter, req *http.Request) {
		apiCfg.fileServerHits.Store(0)
		res.WriteHeader(http.StatusOK)
	})

	//----------------------------------------------------------------------

	s := &http.Server{
		Addr:           ":8080",
		Handler:        servemux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}
