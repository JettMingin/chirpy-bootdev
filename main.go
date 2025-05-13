package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/JettMingin/chirpy-bootdev/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	DB             *database.Queries
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) postUser(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	reqData, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading req.Body: %s", err)
		res.WriteHeader(500)
		return
	}
	var newEmail map[string]string
	if err := json.Unmarshal(reqData, &newEmail); err != nil {
		log.Printf("Error Unmarshalling req.Body: %s", err)
		res.WriteHeader(500)
		return
	}

	fmt.Println("newEmail map:", newEmail)

	emailRegex := regexp.MustCompile(`(?i)^[0-9a-z]+@[a-z0-9]+\.[a-z]{1,3}$`)
	checkEmail, ok := newEmail["email"]
	if !ok || !emailRegex.MatchString(checkEmail) {
		var errorMap = map[string]string{
			"error": "req.Body did not contain valid email key or value",
		}
		data, err := json.Marshal(errorMap)
		if err != nil {
			log.Printf("Error marshalling error response: %s", err)
			res.WriteHeader(500)
			return
		}
		res.WriteHeader(400)
		res.Write(data)
		return
	}

	// newUser := User{}
	//process the email and add a new user to db

	res.WriteHeader(200)
	//respond with the new user json, not this message
	res.Write([]byte("response from postUser"))
}

func main() {
	//set up db
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(err)
	}
	dbQueries := database.New(db)

	apiCfg := &apiConfig{
		DB: dbQueries,
	}

	servemux := http.NewServeMux() //multiplex (seems to work like my map-based router in deno)

	myFileServer := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	servemux.Handle("/app/", apiCfg.middlewareMetricsInc(myFileServer))

	servemux.HandleFunc("GET /api/healthz", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("OK"))
	})

	servemux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
	servemux.HandleFunc("POST /api/users", apiCfg.postUser)

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
