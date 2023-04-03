package database

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lordmoma/chirpy/internal/config"
)

type Chirp struct {
	ID   int    `json:"id"`
	AuthorID int    `json:"author_id"`
	Body string `json:"body"`
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(r *http.Request, apiCfg *config.ApiConfig, body string) (Chirp, error) {

	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return Chirp{}, fmt.Errorf("authorization header missing")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(apiCfg.JwtSecret), nil
	})
	if err != nil {
		return Chirp{}, err
	}
	if !token.Valid {
		return Chirp{}, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)

	if !ok || claims.ExpiresAt == nil || claims.ExpiresAt.Before(time.Now().UTC()) {
		return Chirp{}, fmt.Errorf("refresh Token has been revoked")
	}

	id := len(dbStructure.Chirps) + 1

	authorID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return Chirp{}, err
	}

	chirp := Chirp{
		ID:   id,
		AuthorID: authorID,
		Body: body,
	}

	dbStructure.Chirps[id] = chirp
	if err := db.writeDB(dbStructure); err != nil {
		return Chirp{}, err
	}
	return chirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {

	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := make([]Chirp, 0, len(dbStructure.Chirps))
	for _, chirp := range dbStructure.Chirps {
		chirps = append(chirps, chirp)
	}

	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID && chirps[i].AuthorID < chirps[j].AuthorID
	})

	return chirps, nil
}
