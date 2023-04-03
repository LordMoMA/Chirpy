package database

import (
	"errors"
	"sort"
)

type Chirp struct {
	ID   int    `json:"id"`
	AuthorID int    `json:"author_id"`
	Body string `json:"body"`
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(authorID int, body string) (Chirp, error) {

	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(dbStructure.Chirps) + 1

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
		return chirps[i].AuthorID < chirps[j].AuthorID && chirps[i].ID < chirps[j].ID
	})

	return chirps, nil
}

func (db *DB) DeleteChirp(authorID, id int) error {

	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	if _, ok := dbStructure.Chirps[id]; !ok {
		return errors.New("chirp not found")
	}

	if dbStructure.Chirps[id].AuthorID != authorID {
		return errors.New("unauthorised to delete a chirp")
	}

	delete(dbStructure.Chirps, id)

	if err := db.writeDB(dbStructure); err != nil {
		return err
	}

	return nil

}