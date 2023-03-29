package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudflare/cfssl/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lordmoma/chirpy/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Expire   int    `json:"expires_in_seconds"`
}

type LoginResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Token string `json:"token"`
}

func LoginHandler(db *database.DB, apiCfg *ApiConfig) http.HandlerFunc {
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

		// fmt.Printf("JWT_SECRET-2: %s\n", apiCfg.JwtSecret)
		// jwtSecret := []byte(os.Getenv("JWT_SECRET"))
		if len(apiCfg.JwtSecret) == 0 {
			http.Error(w, "JWT_SECRET not set", http.StatusInternalServerError)
			return
		}

		var expirationTime time.Duration

		expiresInSeconds, err := strconv.Atoi(r.FormValue("expires_in_seconds"))
		if err == nil {
			if expiresInSeconds > 86400 {
				expirationTime = 86400 * time.Second
			} else {
				expirationTime = time.Duration(expiresInSeconds) * time.Second
			}
		} else {
			expirationTime = 24 * time.Hour
		}

		claims := jwt.RegisteredClaims{
			Issuer:    "chirpy",
			Subject:   strconv.Itoa(user.ID),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expirationTime)),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString([]byte(apiCfg.JwtSecret))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		res := LoginResponse{
			ID:    user.ID,
			Email: user.Email,
			Token: signedToken,
		}

		json.NewEncoder(w).Encode(res)
	}
}
