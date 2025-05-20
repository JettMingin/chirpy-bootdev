package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/JettMingin/chirpy-bootdev/internal/auth"
	"github.com/JettMingin/chirpy-bootdev/internal/database"
	"github.com/google/uuid"
)

// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJpc3MiOiJjaGlycHkiLCJzdWIiOiI0YWM0MGE4Zi01MjczLTQ2ZWYtYTZhZi1hNDZiMWI3MjZkYTgiLCJleHAiOjE3NDc3ODI2ODQsImlhdCI6MTc0Nzc3OTA4NH0.FzPUR5KS4PkvsxCeZZUBSFoCyK3CG8a2xhcPUcAZKW8

func ErrorResponseWriter(res http.ResponseWriter, errMsg string, errorVal error, statusCode int) {
	if errMsg == "JSON" {
		errMsg = "Failed to encode a JSON response"
	}
	errorMap := map[string]string{
		"error":  errorVal.Error(),
		"errMsg": errMsg,
	}
	errorResponse, err := json.Marshal(errorMap)
	if err != nil {
		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		res.WriteHeader(500)
		res.Write([]byte("Internal Server Error - failed to encode proper response"))
	} else {
		res.WriteHeader(statusCode)
		res.Write(errorResponse)
	}
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}
type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func (cfg *apiConfig) getChirp(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	reqChirpId, err := uuid.Parse(req.PathValue("chirpId"))
	if err != nil {
		ErrorResponseWriter(res, "failed to parse uuid provided in url", err, 400)
		return
	}
	dbChirp, err := cfg.DB.GetOneChirp(req.Context(), reqChirpId)
	if err != nil {
		ErrorResponseWriter(res, "failed to find chirp with provided id in DB", err, 404)
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
		ErrorResponseWriter(res, "JSON", err, 500)
		return
	}
	res.WriteHeader(200)
	res.Write(successRes)
}

func (cfg *apiConfig) getAllChirps(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	dbChirps, err := cfg.DB.GetAllChirps(req.Context())
	if err != nil {
		ErrorResponseWriter(res, "failed to query for all chirps in DB", err, 500)
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
		ErrorResponseWriter(res, "JSON", err, 500)
		return
	}
	res.WriteHeader(200)
	res.Write(successRes)
}

func (cfg *apiConfig) postChirp(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	tokenString, err := auth.GetBearerToken(req.Header)
	if err != nil {
		ErrorResponseWriter(res, "No Authorization Header Set", err, 401)
		return
	}

	validUserId, err := auth.ValidateJWT(tokenString, cfg.TokenSecret)
	if err != nil {
		ErrorResponseWriter(res, "Bad Token, Unauthorized", err, 401)
		return
	}

	reqData, err := io.ReadAll(req.Body)
	if err != nil {
		ErrorResponseWriter(res, "Failed to read request body", err, 500)
		return
	}
	var newChirpReq map[string]string
	if err := json.Unmarshal(reqData, &newChirpReq); err != nil {
		ErrorResponseWriter(res, "Failed to decode request body", err, 500)
		return
	}

	cleanedChirp, isValid := validateChirpHelper(newChirpReq["body"])
	if !isValid {
		err := errors.New("invalid request body")
		ErrorResponseWriter(res, "Request Body missing 'body' field", err, 400)
		return
	}

	dbChirp, err := cfg.DB.CreateChirp(req.Context(),
		database.CreateChirpParams{Body: cleanedChirp, UserID: validUserId})
	if err != nil {
		ErrorResponseWriter(res, "Failed to write new chirp to DB", err, 500)
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
		ErrorResponseWriter(res, "JSON", err, 500)
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
		ErrorResponseWriter(res, "Failed to read request body", err, 500)
		return
	}
	var newUserData map[string]string
	if err := json.Unmarshal(reqData, &newUserData); err != nil {
		ErrorResponseWriter(res, "Failed to decode request body", err, 500)
		return
	}

	emailRegex := regexp.MustCompile(`(?i)^[0-9a-z]+@[a-z0-9]+\.[a-z]{1,3}$`)
	checkEmail, ok := newUserData["email"]
	if !ok || !emailRegex.MatchString(checkEmail) {
		err := errors.New("missing or invalid email field")
		ErrorResponseWriter(res, "email missing from body or invalid format", err, 400)
		return
	}
	checkPassword, ok := newUserData["password"]
	//very secure and complex pw safety checker with regex ofc
	if !ok || len(checkPassword) < 4 {
		err := errors.New("missing or invalid password field")
		ErrorResponseWriter(res, "password missing from body or invalid format", err, 400)
		return
	}

	hashedPassword, err := auth.HashPassword(checkPassword)
	if err != nil {
		ErrorResponseWriter(res, "Failed to hash password", err, 500)
		return
	}

	dbUser, err := cfg.DB.CreateUser(req.Context(),
		database.CreateUserParams{Email: newUserData["email"], PwHash: hashedPassword})
	if err != nil {
		ErrorResponseWriter(res, "Failed to write new user to DB", err, 500)
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
		ErrorResponseWriter(res, "JSON", err, 500)
		return
	}
	res.WriteHeader(201)
	res.Write(responseSuc)
}

func (cfg *apiConfig) Login(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	reqData, err := io.ReadAll(req.Body)
	if err != nil {
		ErrorResponseWriter(res, "Failed to read request body", err, 500)
		return
	}
	var loginInfo map[string]string
	if err := json.Unmarshal(reqData, &loginInfo); err != nil {
		ErrorResponseWriter(res, "Failed to decode request body", err, 500)
		return
	}
	_, okEmail := loginInfo["email"]
	_, okPw := loginInfo["password"]
	if !okEmail || !okPw {
		err := errors.New("request body did not contain 'email' or 'password' field")
		ErrorResponseWriter(res, "invalid request body", err, 400)
		return
	}

	dbUser, err := cfg.DB.LookupUser(req.Context(), loginInfo["email"])
	if err != nil {
		ErrorResponseWriter(res, "DB lookup error, incorrect email", err, 401)
		return
	}
	if err := auth.CheckPasswordHash(dbUser.PwHash, loginInfo["password"]); err != nil {
		ErrorResponseWriter(res, "Invalid Password", err, 401)
		return
	}

	newUser := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	newToken, err := auth.MakeJWT(newUser.ID, cfg.TokenSecret, time.Duration(3600)*time.Second)
	if err != nil {
		ErrorResponseWriter(res, "Failed to create new JWT", err, 500)
		return
	}
	newUser.Token = newToken

	newRefereshToken := auth.MakeRefreshToken()
	refreshTokenParams := database.CreateTokenParams{
		Token:     newRefereshToken,
		UserID:    newUser.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	}
	_, rtErr := cfg.DB.CreateToken(req.Context(), refreshTokenParams)
	if rtErr != nil {
		ErrorResponseWriter(res, "Failed to write new refresh token to DB", rtErr, 500)
		return
	}
	newUser.RefreshToken = newRefereshToken

	responseSuc, err := json.Marshal(newUser)
	if err != nil {
		ErrorResponseWriter(res, "JSON", err, 500)
		return
	}
	res.WriteHeader(200)
	res.Write(responseSuc)
}

func (cfg *apiConfig) RefreshAccessToken(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	tokenString, err := auth.GetBearerToken(req.Header)
	if err != nil {
		ErrorResponseWriter(res, "Missing Authorization Header", err, 401)
		return
	}

	dbUserID, err := cfg.DB.GetUserFromRefreshToken(req.Context(), tokenString)
	if err != nil {
		ErrorResponseWriter(res, "failed to get user id from provided refresh token", err, 401)
		return
	}

	newAccessToken, err := auth.MakeJWT(dbUserID, cfg.TokenSecret, time.Duration(3600)*time.Second)
	if err != nil {
		ErrorResponseWriter(res, "Failed to create new JWT", err, 500)
		return
	}

	successResponse, err := json.Marshal(
		map[string]string{"token": newAccessToken})
	if err != nil {
		ErrorResponseWriter(res, "JSON", err, 500)
		return
	}
	res.WriteHeader(200)
	res.Write(successResponse)
}

func (cfg *apiConfig) RevokeRefreshToken(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	tokenString, err := auth.GetBearerToken(req.Header)
	if err != nil {
		ErrorResponseWriter(res, "Missing Authorization Header", err, 401)
		return
	}
	if err := cfg.DB.RevokeRefreshTokenFromDB(req.Context(), tokenString); err != nil {
		ErrorResponseWriter(res, "Failed to revoke refresh token in DB", err, 500)
		return
	}
	res.WriteHeader(204)
}

func (cfg *apiConfig) UpdateUser(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	tokenString, err := auth.GetBearerToken(req.Header)
	if err != nil {
		ErrorResponseWriter(res, "Missing Authorization Header", err, 401)
		return
	}
	userID, err := auth.ValidateJWT(tokenString, cfg.TokenSecret)
	if err != nil {
		ErrorResponseWriter(res, "Invalid JWT", err, 401)
		return
	}

	reqData, err := io.ReadAll(req.Body)
	if err != nil {
		ErrorResponseWriter(res, "Failed to read request body", err, 500)
		return
	}
	var newUserData map[string]string
	if err := json.Unmarshal(reqData, &newUserData); err != nil {
		ErrorResponseWriter(res, "Failed to decode request body", err, 500)
		return
	}

	emailRegex := regexp.MustCompile(`(?i)^[0-9a-z]+@[a-z0-9]+\.[a-z]{1,3}$`)
	checkEmail, ok := newUserData["email"]
	if !ok || !emailRegex.MatchString(checkEmail) {
		err := errors.New("missing or invalid email field")
		ErrorResponseWriter(res, "email missing from body or invalid format", err, 400)
		return
	}
	checkPassword, ok := newUserData["password"]
	//very secure and complex pw safety checker with regex ofc
	if !ok || len(checkPassword) < 4 {
		err := errors.New("missing or invalid password field")
		ErrorResponseWriter(res, "password missing from body or invalid format", err, 400)
		return
	}

	hashedPassword, err := auth.HashPassword(checkPassword)
	if err != nil {
		ErrorResponseWriter(res, "Failed to hash password", err, 500)
		return
	}

	updatedUserInfo, err := cfg.DB.UpdateUser(req.Context(),
		database.UpdateUserParams{Email: newUserData["email"], PwHash: hashedPassword, ID: userID})
	if err != nil {
		ErrorResponseWriter(res, "Failed to update user-info in DB", err, 500)
		return
	}
	updatedUser := User{
		ID:        updatedUserInfo.ID,
		CreatedAt: updatedUserInfo.CreatedAt,
		UpdatedAt: updatedUserInfo.UpdatedAt,
		Email:     updatedUserInfo.Email,
	}
	responseSuc, err := json.Marshal(updatedUser)
	if err != nil {
		ErrorResponseWriter(res, "JSON", err, 500)
		return
	}
	res.WriteHeader(200)
	res.Write(responseSuc)
}

func (cfg *apiConfig) DeleteChirp(res http.ResponseWriter, req *http.Request) {
	reqChirpId, err := uuid.Parse(req.PathValue("chirpId"))
	if err != nil {
		res.WriteHeader(400)
		return
	}
	dbChirp, err := cfg.DB.GetOneChirp(req.Context(), reqChirpId)
	if err != nil {
		res.WriteHeader(404)
		return
	}
	tokenString, err := auth.GetBearerToken(req.Header)
	if err != nil {
		res.WriteHeader(401)
		return
	}
	validUserId, err := auth.ValidateJWT(tokenString, cfg.TokenSecret)
	if err != nil {
		res.WriteHeader(401)
		return
	}

	if validUserId != dbChirp.UserID {
		res.WriteHeader(403)
		return
	}

	if err := cfg.DB.DeleteOneChirp(req.Context(), dbChirp.ID); err != nil {
		res.WriteHeader(500)
		return
	}
	res.WriteHeader(204)
}
