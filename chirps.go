package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
)

type Chirp struct {
	Body string `json:"body"`
}

func validateChirpHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
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
