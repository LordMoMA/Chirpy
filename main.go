package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"

	"github.com/go-chi/chi"
)

type apiConfig struct {
	fileserverHits uint64
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	// Create a new apiConfig struct to hold the request count
	apiCfg := &apiConfig{}

	// Create a new router for the /api namespace
	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", healthzHandler)
	apiRouter.Post("/chirps", apiCfg.validateHandler)
	apiRouter.Get("/chirps", apiCfg.validateHandler)

	// create a new router for the admin
	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiCfg.metricsHandler)

	// Mount the apiRouter at /api in the main router
	r := chi.NewRouter()
	r.Mount("/api", apiRouter)
	r.Mount("/admin", adminRouter)

	// Serve static files from the root directory and add the middleware to track metrics
	r.Mount("/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot))))

	// Wrap the mux in a custom middleware function that adds CORS headers to the response
	corsMux := middlewareCors(r)

	// Create a new http.Server and use the corsMux as the handler
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	// Use the server's ListenAndServe method to start the server
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	// log.Fatal(srv.ListenAndServe())
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

// Chirp represents a single chirp message
type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

// DB represents a database connection
type DB struct {
	path string
	mux  *sync.RWMutex
}

// DBStructure represents the structure of the database file
type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

// NewDB creates a new database connection and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	if err := db.ensureDB(); err != nil {
		return nil, err
	}
	return db, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(dbStructure.Chirps) + 1
	chirp := Chirp{
		ID:   id,
		Body: body,
	}

	if err := validateChirp(chirp); err != nil {
		return Chirp{}, err
	}

	dbStructure.Chirps[id] = chirp
	if err := db.writeDB(dbStructure); err != nil {
		return Chirp{}, err
	}

	return chirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

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

	dbStructure := DBStructure{}

	bytes, err := ioutil.ReadFile(db.path)
	if err != nil {
		return dbStructure, err
	}

	if err := json.Unmarshal(bytes, &dbStructure); err != nil {
		return dbStructure, err
	}

	return dbStructure, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	file, err := json.MarshalIndent(dbStructure, "", " ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(db.path, file, 0644); err != nil {
		return err
	}

	return nil
}
