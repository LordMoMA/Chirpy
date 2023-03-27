package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"

	"github.com/cloudflare/cfssl/log"
	"github.com/go-chi/chi"
	"golang.org/x/crypto/bcrypt"
)

// Chirp represents a single chirp message
type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// DB represents a database connection
type DB struct {
	path string
	mux  *sync.RWMutex
}

// DBStructure represents the structure of the database file
type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

// NewDB creates a new database connection and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mux:  &sync.RWMutex{},
	}

	// Open the file with read and write permissions
	file, err := os.OpenFile(db.path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	// Assign the file to the DB
	db.path = file.Name()

	if err := db.ensureDB(); err != nil {
		return nil, err
	}
	fmt.Println("calling NewDB()... 📚")
	return db, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {

	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(dbStructure.Chirps) + 1
	chirp := Chirp{
		ID:   id,
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
		return chirps[i].ID < chirps[j].ID
	})

	return chirps, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	if _, err := os.Stat(db.path); err == nil {
		fmt.Printf("database file found at path %s\n", db.path)
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	dbStructure := DBStructure{
		Chirps: make(map[int]Chirp),
	}

	if err := db.writeDB(dbStructure); err != nil {
		return err
	}
	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	var dbStructure DBStructure

	file, err := os.Open(db.path)
	if err != nil {
		return dbStructure, err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&dbStructure)
	if err != nil {
		return dbStructure, err
	}
	return dbStructure, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	file, err := os.Create(db.path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(&dbStructure)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) CreateChirpsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var chirp Chirp
	if err := json.NewDecoder(r.Body).Decode(&chirp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create the chirp
	createdChirp, err := db.CreateChirp(chirp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the response
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(createdChirp)
}

func (db *DB) GetUserbyEmail(email string) (User, error) {

	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	for _, user := range dbStructure.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return User{}, errors.New("User not found")
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (db *DB) LoginHandler(w http.ResponseWriter, r *http.Request) {
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
	fmt.Printf("password: %s", req.Password)
	fmt.Printf("hashed password: %s", user.Password)
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "invalid password", http.StatusUnauthorized)
		log.Error(err)
		return
	}

	// Write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	res := User{
		ID:    user.ID,
		Email: user.Email,
	}

	json.NewEncoder(w).Encode(res)
}

func (db *DB) CreateUser(email, password string) (User, error) {

	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	// Check if user with the same email already exists
	for _, user := range dbStructure.Users {
		if user.Email == email {
			return User{}, fmt.Errorf("user with email %s already exists", email)
		}
	}

	id := len(dbStructure.Users) + 1

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	fmt.Printf("hashed password in CreateUser: %s", hashedPassword)

	user := User{
		ID:       id,
		Email:    email,
		Password: string(hashedPassword),
	}

	dbStructure.Users[id] = user

	if err := db.writeDB(dbStructure); err != nil {
		return User{}, err
	}

	return user, nil
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type UserResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func (db *DB) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//Hash the password using the bcrypt.GenerateFromPassword function
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Create the user
	createdUser, err := db.CreateUser(req.Email, string(hashedPassword))
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

func (db *DB) GetChirpIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	chirps, err := db.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, chirps[id-1])
}

func (db *DB) GetChirpsHandler(w http.ResponseWriter, r *http.Request) {
	chirps, err := db.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, chirps)
}

func respondWithError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
