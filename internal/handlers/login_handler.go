package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudflare/cfssl/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lordmoma/chirpy/internal/config"
	"github.com/lordmoma/chirpy/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	// Expire   int    `json:"expires_in_seconds"`
}

type LoginResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func LoginHandler(db *database.DB, apiCfg *config.ApiConfig) http.HandlerFunc {
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

		// Create the JWT token
		// jwtSecret := []byte(os.Getenv("JWT_SECRET"))
		if len(apiCfg.JwtSecret) == 0 {
			http.Error(w, "JWT_SECRET not set", http.StatusInternalServerError)
			return
		}

		// Set expiration time for access token
		accessTokenExpirationTime := time.Now().Add(1 * time.Hour)

		// Set expiration time for refresh token
		refreshTokenExpirationTime := time.Now().Add(60 * 24 * time.Hour)

		accessTokenClaims := jwt.RegisteredClaims{
			Issuer:    "chirpy-access",
			Subject:   strconv.Itoa(user.ID),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(accessTokenExpirationTime.UTC()),
		}

		accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
		accessTokenString, err := accessToken.SignedString([]byte(apiCfg.JwtSecret))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		refreshTokenClaims := jwt.RegisteredClaims{
			Issuer:    "chirpy-refresh",
			Subject:   strconv.Itoa(user.ID),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(refreshTokenExpirationTime.UTC()),
		}

		refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
		refreshTokenString, err := refreshToken.SignedString([]byte(apiCfg.JwtSecret))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		res := LoginResponse{
			ID:           user.ID,
			Email:        user.Email,
			AccessToken:  accessTokenString,
			RefreshToken: refreshTokenString,
		}
		json.NewEncoder(w).Encode(res)
	}
}
