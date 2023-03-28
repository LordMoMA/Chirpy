package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/lordmoma/chirpy/internal/database"
)

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type UserResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func CreateUserHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body
		var req CreateUserRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Create the user
		createdUser, err := db.CreateUser(req.Email, req.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		res := UserResponse{
			ID:    createdUser.ID,
			Email: createdUser.Email,
		}

		json.NewEncoder(w).Encode(res)
	}
}
