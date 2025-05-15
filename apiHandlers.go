package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/JettMingin/chirpy-bootdev/internal/auth"
	"github.com/JettMingin/chirpy-bootdev/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}
type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) getChirp(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	reqChirpId, err := uuid.Parse(req.PathValue("chirpId"))
	if err != nil {
		responseErr, err := json.Marshal(
			map[string]string{"error": fmt.Sprintf("failed to parse uuid from req: %s", err)})
		if err != nil {
			log.Printf("Error marshaling error response: %s", err)
			res.WriteHeader(500)
			return
		}
		res.WriteHeader(400)
		res.Write(responseErr)
		return
	}
	dbChirp, err := cfg.DB.GetOneChirp(req.Context(), reqChirpId)
	if err != nil {
		responseErr, err := json.Marshal(
			map[string]string{"error": fmt.Sprintf("failed to get chirp from db: %s", err)})
		if err != nil {
			log.Printf("Error marshaling error response: %s", err)
			res.WriteHeader(500)
			return
		}
		res.WriteHeader(404)
		res.Write(responseErr)
		return
	}

	selectedChirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
	successRes, err := json.Marshal(selectedChirp)
	if err != nil {
		log.Printf("Error marshaling success response: %s", err)
		res.WriteHeader(500)
		return
	}
	res.WriteHeader(200)
	res.Write(successRes)
}

func (cfg *apiConfig) getAllChirps(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	dbChirps, err := cfg.DB.GetAllChirps(req.Context())
	if err != nil {
		responseErr, err := json.Marshal(
			map[string]string{"error": fmt.Sprintf("failed to get chirps from db: %s", err)})
		if err != nil {
			log.Printf("Error marshaling error response: %s", err)
			res.WriteHeader(500)
			return
		}
		res.WriteHeader(400)
		res.Write(responseErr)
		return
	}

	selectedChirps := []Chirp{}
	for _, row := range dbChirps {
		aChirp := Chirp{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
			Body:      row.Body,
			UserID:    row.UserID,
		}
		selectedChirps = append(selectedChirps, aChirp)
	}

	successRes, err := json.Marshal(selectedChirps)
	if err != nil {
		log.Printf("Error marshaling success response: %s", err)
		res.WriteHeader(500)
		return
	}
	res.WriteHeader(200)
	res.Write(successRes)
}

func (cfg *apiConfig) postChirp(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	reqData, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading req.Body: %s", err)
		res.WriteHeader(500)
		return
	}
	var newChirpReq map[string]string
	if err := json.Unmarshal(reqData, &newChirpReq); err != nil {
		log.Printf("Error Unmarshaling req.Body: %s", err)
		res.WriteHeader(500)
		return
	}

	cleanedChirp, isValid := validateChirpHelper(newChirpReq["body"])
	userIdStr, ok := newChirpReq["user_id"]
	parsedUserId, err := uuid.Parse(userIdStr)
	if !isValid || !ok || err != nil {
		responseErr, err := json.Marshal(
			map[string]string{"error": "req.Body did not contain valid chirp body or user id"})
		if err != nil {
			log.Printf("Error marshaling error response: %s", err)
			res.WriteHeader(500)
			return
		}
		res.WriteHeader(400)
		res.Write(responseErr)
		return
	}

	dbChirp, err := cfg.DB.CreateChirp(req.Context(),
		database.CreateChirpParams{Body: cleanedChirp, UserID: parsedUserId})
	if err != nil {
		log.Printf("Error creating new row in Chirps table: %s", err)
		res.WriteHeader(500)
		return
	}
	newChirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}

	successRes, err := json.Marshal(newChirp)
	if err != nil {
		log.Printf("Error marshaling success response: %s", err)
		res.WriteHeader(500)
		return
	}
	res.WriteHeader(201)
	res.Write(successRes)
}

func validateChirpHelper(rawChirp string) (string, bool) {
	if len(rawChirp) > 140 || len(rawChirp) < 1 {
		return "", false
	}
	var profanityRegexes = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bkerfuffle\b`),
		regexp.MustCompile(`(?i)\bsharbert\b`),
		regexp.MustCompile(`(?i)\bfornax\b`),
	}
	cleanedString := rawChirp
	for _, cussRegEx := range profanityRegexes {
		cleanedString = cussRegEx.ReplaceAllString(cleanedString, "****")
	}
	return cleanedString, true
}

func (cfg *apiConfig) postUser(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	reqData, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading req.Body: %s", err)
		res.WriteHeader(500)
		return
	}
	var newUserData map[string]string
	if err := json.Unmarshal(reqData, &newUserData); err != nil {
		log.Printf("Error Unmarshaling req.Body: %s", err)
		res.WriteHeader(500)
		return
	}

	emailRegex := regexp.MustCompile(`(?i)^[0-9a-z]+@[a-z0-9]+\.[a-z]{1,3}$`)
	checkEmail, ok := newUserData["email"]
	if !ok || !emailRegex.MatchString(checkEmail) {
		responseErr, err := json.Marshal(
			map[string]string{"error": "req.Body did not contain valid email key or value"})
		if err != nil {
			log.Printf("Error marshaling error response: %s", err)
			res.WriteHeader(500)
			return
		}
		res.WriteHeader(400)
		res.Write(responseErr)
		return
	}
	checkPassword, ok := newUserData["password"]
	//very secure and complex pw safety checker with regex ofc
	if !ok {
		responseErr, err := json.Marshal(
			map[string]string{"error": "req.Body did not contain valid password field"})
		if err != nil {
			log.Printf("Error marshaling error response: %s", err)
			res.WriteHeader(500)
			return
		}
		res.WriteHeader(400)
		res.Write(responseErr)
		return
	}

	hashedPassword, err := auth.HashPassword(checkPassword)
	if err != nil {
		responseErr, err := json.Marshal(
			map[string]string{"error": "HashPassword failed"})
		if err != nil {
			log.Printf("Error marshaling error response: %s", err)
			res.WriteHeader(500)
			return
		}
		res.WriteHeader(400)
		res.Write(responseErr)
		return
	}

	dbUser, err := cfg.DB.CreateUser(req.Context(),
		database.CreateUserParams{Email: newUserData["email"], PwHash: hashedPassword})
	if err != nil {
		log.Printf("Error creating new row in Users table: %s", err)
		res.WriteHeader(500)
		return
	}
	newUser := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	responseSuc, err := json.Marshal(newUser)
	if err != nil {
		log.Printf("Error marshaling success response: %s", err)
		res.WriteHeader(500)
		return
	}
	res.WriteHeader(201)
	res.Write(responseSuc)
}

func (cfg *apiConfig) login(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	reqData, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading req.Body: %s", err)
		res.WriteHeader(500)
		return
	}
	var loginInfo map[string]string
	if err := json.Unmarshal(reqData, &loginInfo); err != nil {
		log.Printf("Error Unmarshaling req.Body: %s", err)
		res.WriteHeader(500)
		return
	}
	_, okEmail := loginInfo["email"]
	_, okPw := loginInfo["password"]
	if !okEmail || !okPw {
		responseErr, err := json.Marshal(
			map[string]string{"error": "req.Body did not contain 'email' or 'password' fields"})
		if err != nil {
			log.Printf("Error marshaling error response: %s", err)
			res.WriteHeader(500)
			return
		}
		res.WriteHeader(400)
		res.Write(responseErr)
		return
	}

	dbUser, dbErr := cfg.DB.LookupUser(req.Context(), loginInfo["email"])
	pwErr := auth.CheckPasswordHash(dbUser.PwHash, loginInfo["password"])
	if dbErr != nil || pwErr != nil {
		responseErr, err := json.Marshal(
			map[string]string{"error": "Incorrect email or password"})
		if err != nil {
			log.Printf("Error marshaling error response: %s", err)
			res.WriteHeader(500)
			return
		}
		res.WriteHeader(401)
		res.Write(responseErr)
		return
	}

	newUser := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	responseSuc, err := json.Marshal(newUser)
	if err != nil {
		log.Printf("Error marshaling success response: %s", err)
		res.WriteHeader(500)
		return
	}
	res.WriteHeader(200)
	res.Write(responseSuc)

}
