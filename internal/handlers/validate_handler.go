package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

func ValidateHandler(cfg *ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var chirp struct {
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
}

func replaceProfane(text string) string {
	profane := []string{"kerfuffle", "sharbert", "fornax"}
	for _, word := range profane {
		text = strings.ReplaceAll(text, word, "****")
		text = strings.ReplaceAll(text, strings.ToUpper(word), "****")
	}
	return text
}
