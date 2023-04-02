package database

import (
	"time"
)

type RevokedToken struct {
	ID       string    `json:"id"`
	RevokedAt time.Time `json:"revoked_at"`
}

func (db *DB) RevokeToken(tokenID string, revokedAt time.Time) (RevokedToken,error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return RevokedToken{},err
	}
	
	token := RevokedToken{
		ID: tokenID,
		RevokedAt: revokedAt,
	}

	id := len(dbStructure.Tokens) + 1
	dbStructure.Tokens[id] = token

	if err := db.writeDB(dbStructure); err != nil {
		return RevokedToken{}, err
	}

	return token, nil
}