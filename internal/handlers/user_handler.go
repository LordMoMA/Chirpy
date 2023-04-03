package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lordmoma/chirpy/internal/config"
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

// type updateResponse struct {
// 	Email string `json:"email"`
// }

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

func UpdateUserHandler(db *database.DB, apiCfg *config.ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the token from the headers
		// fmt.Printf("http request: %v", r)
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		//tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method jwt.SigningMethodHMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// Return the secret key used to sign the token
			return []byte(apiCfg.JwtSecret), nil
		})
		if err != nil {
			fmt.Printf("token error: %v", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Check if the token has expired
		claims, ok := token.Claims.(*jwt.RegisteredClaims)
		if !ok || claims.ExpiresAt == nil || claims.ExpiresAt.Before(time.Now().UTC()) {
			http.Error(w, "Token has expired", http.StatusUnauthorized)
			return
		}

		// Check if the token is a refresh token
		if claims.Issuer == "chirpy-refresh" {
			http.Error(w, "Invalid token: Refresh Token Rejected", http.StatusUnauthorized)
			return
		}

		// Extract the user ID from the token
		userID, err := strconv.Atoi(claims.Subject)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Parse the request body
		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Update the user in the database
		updatedUser, err := db.UpdateUser(userID, req.Email, req.Password)
		if err != nil {
			fmt.Printf("error updating user: %v", err)
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		fmt.Printf("updated user: %v\n", updatedUser)
		// Write the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		res := UserResponse{
			ID:    updatedUser.ID,
			Email: updatedUser.Email,
		}

		json.NewEncoder(w).Encode(res)
	}
}
