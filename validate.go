package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (cfg *apiConfig) validateHandler(w http.ResponseWriter, r *http.Request) {

	var chirp struct {
		Id   int    `json:"id"`
		Body string `json:"body"`
	}

	err := json.NewDecoder(r.Body).Decode(&chirp)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cleanedBody := replaceProfane(chirp.Body)

	respondWithJSON(w, http.StatusOK, map[string]string{"cleaned_body": cleanedBody})

}

func replaceProfane(text string) string {
	profane := []string{"kerfuffle", "sharbert", "fornax"}
	for _, word := range profane {
		text = strings.ReplaceAll(text, word, "****")
		text = strings.ReplaceAll(text, strings.ToUpper(word), "****")
	}
	return text
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
