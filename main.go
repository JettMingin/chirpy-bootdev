package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/JettMingin/chirpy-bootdev/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	DB             *database.Queries
	Platform       string
	TokenSecret    string
	PolkaKey       string
}

func main() {
	//set up env
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	superSecret := os.Getenv("TOKEN_SECRET")
	polkaKey := os.Getenv("POLKA_KEY")

	//set up db
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(err)
	}
	dbQueries := database.New(db)

	apiCfg := &apiConfig{
		DB:          dbQueries,
		Platform:    platform,
		TokenSecret: superSecret,
		PolkaKey:    polkaKey,
	}

	servemux := http.NewServeMux() //multiplex (seems to work like my map-based router in deno)

	myFileServer := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	servemux.Handle("/app/", apiCfg.middlewareMetricsInc(myFileServer))

	servemux.HandleFunc("GET /api/healthz", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("OK"))
	})

	servemux.HandleFunc("POST /api/users", apiCfg.postUser)
	servemux.HandleFunc("POST /api/login", apiCfg.Login)
	servemux.HandleFunc("PUT /api/users", apiCfg.UpdateUser)
	servemux.HandleFunc("POST /api/refresh", apiCfg.RefreshAccessToken)
	servemux.HandleFunc("POST /api/revoke", apiCfg.RevokeRefreshToken)

	servemux.HandleFunc("GET /api/chirps", apiCfg.getChirps)
	servemux.HandleFunc("GET /api/chirps/{chirpId}", apiCfg.getChirp)
	servemux.HandleFunc("POST /api/chirps", apiCfg.postChirp)
	servemux.HandleFunc("DELETE /api/chirps/{chirpId}", apiCfg.DeleteChirp)

	servemux.HandleFunc("POST /api/polka/webhooks", apiCfg.UpgradeUser)

	servemux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	servemux.HandleFunc("POST /admin/reset", apiCfg.postToReset)

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
