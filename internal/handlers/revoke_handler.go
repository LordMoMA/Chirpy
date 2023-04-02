package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lordmoma/chirpy/internal/database"
)

type RevokeResponse struct {
	ID       string    `json:"id"`
	RevokedAt time.Time `json:"revoked_at"`
}

func RevokeTokenHandler(db *database.DB, apiCfg *ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the token from the headers
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
			// fmt.Printf("token error: %v", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*jwt.RegisteredClaims)

		if !ok || claims.Issuer != "chirpy-refresh" {
			http.Error(w, "Invalid token: Must be a Refresh Token", http.StatusUnauthorized)
			return
		}

		// Revoke the refresh token
		token.Valid = false
		
		// store the token string and revoke time in the database
		currentTime := time.Now().UTC()
		
		revokedToken, err := db.RevokeToken(tokenString, currentTime) 
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		// Respond with a 200 status code
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(revokedToken)
	}
}