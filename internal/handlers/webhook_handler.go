package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/lordmoma/chirpy/internal/config"
	"github.com/lordmoma/chirpy/internal/database"
)

// {
// 	"event": "user.upgraded",
// 	"data": {
// 	  "user_id": 3
// 	}
//   }
type WebhookRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserID int `json:"user_id"`
	} `json:"data"`
}

func WebhookHandler(db *database.DB, apiCfg *config.ApiConfig) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		var req WebhookRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Event != "user.upgraded" {
			w.Header(http.StatusOK)
		}
	}
}