package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lordmoma/chirpy/internal/config"
	"github.com/lordmoma/chirpy/internal/database"
)

type RefreshResponse struct {
	Token string `json:"token"`
}

func AccessTokenHandler(db *database.DB, apiCfg *config.ApiConfig) http.HandlerFunc {
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
			fmt.Printf("token error: %v", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*jwt.RegisteredClaims)
		
		// Check if the token is a refresh token
		if claims.Issuer != "chirpy-refresh" {
			http.Error(w, "Invalid token: Must be a Refresh Token", http.StatusUnauthorized)
			return
		}
		
		// Check if the token has been revoked	
		if !ok || claims.ExpiresAt == nil || claims.ExpiresAt.Before(time.Now().UTC()) {
			http.Error(w, "Refresh Token has been revoked!", http.StatusUnauthorized)
			return
		}

		newAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
		})
		
		newAccessTokenString, err := newAccessToken.SignedString([]byte(apiCfg.JwtSecret))

		if err != nil {
			http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
			return
		}

		res := RefreshResponse{
			Token: newAccessTokenString,
		}
		// Return the new access token
		json.NewEncoder(w).Encode(res)
	}
}