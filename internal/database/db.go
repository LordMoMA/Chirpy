package database

import (
	"encoding/json"
	"os"
	"sync"
)

// DB represents a database connection
type DB struct {
	path string
	mux  *sync.RWMutex
	// dbStructure DBStructure
}
type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
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
	// // Decode the JSON file into DBStructure
	// err = json.NewDecoder(file).Decode(&db.dbStructure)
	// if err != nil {
	// 	return nil, err
	// }

	return db, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	// Check if the database file exists
	_, err := os.Stat(db.path)
	if os.IsNotExist(err) {
		// Create an empty database file

		dbStructure := DBStructure{
			Chirps: make(map[int]Chirp),
			Users:  make(map[int]User),
		}

		if err := db.writeDB(dbStructure); err != nil {
			return err
		}
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

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&dbStructure); err != nil {
		return err
	}

	// err = json.NewEncoder(file).Encode(&dbStructure)
	// if err != nil {
	// 	return err
	// }

	return nil
}
