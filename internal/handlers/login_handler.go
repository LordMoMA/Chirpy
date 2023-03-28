package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cloudflare/cfssl/log"
	"github.com/lordmoma/chirpy/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Expire   int    `json:"expire"`
}

func LoginHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body
		var req LoginRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// get the user by email
		user, err := db.GetUserbyEmail(req.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Compare the hashed password with the password provided in the request

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			http.Error(w, "invalid password", http.StatusUnauthorized)
			log.Error(err)
			return
		}

		// Write the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		res := database.User{
			ID:    user.ID,
			Email: user.Email,
		}

		json.NewEncoder(w).Encode(res)
	}
}
