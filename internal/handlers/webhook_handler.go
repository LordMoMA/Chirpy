package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/lordmoma/chirpy/internal/config"
	"github.com/lordmoma/chirpy/internal/database"
)


type WebhookRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserID int `json:"user_id"`
	} `json:"data"`
}

type WebhookResponse struct {
	Membership bool `json:"is_chirpy_red"`
}

func WebhookHandler(db *database.DB, apiCfg *config.ApiConfig) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		fmt.Printf("http request: %v", r)
		authHeader := r.Header.Get("Authorization")
		fmt.Printf("authoHeader: %v", authHeader)
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		apiString := strings.TrimPrefix(authHeader, "ApiKey ")

		if apiString != apiCfg.APIKey {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		var req WebhookRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Event != "user.upgraded" {
			w.WriteHeader(http.StatusOK)
		}

		// get the users from db
		user, err := db.UpdateMembership(req.Data.UserID, true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		res := WebhookResponse{
			Membership: user.Membership,
		}

		json.NewEncoder(w).Encode(res)

	}
}