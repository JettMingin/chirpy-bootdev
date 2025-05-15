package main

import (
	"fmt"
	"log"
	"net/http"
)

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

func (cfg *apiConfig) postToReset(res http.ResponseWriter, req *http.Request) {
	if cfg.Platform != "dev" {
		log.Println("req from invalid platform (env PLATFORM discrepancy)")
		res.WriteHeader(403)
		return
	}

	if err := cfg.DB.ResetUsers(req.Context()); err != nil {
		log.Printf("Error clearing all rows in Users table: %s", err)
		res.WriteHeader(500)
		return
	}

	cfg.fileServerHits.Store(0)
	res.WriteHeader(http.StatusOK)
}
