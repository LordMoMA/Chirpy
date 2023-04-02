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

	for k, v := range dbStructure.Tokens {
		if v.ID == tokenID {
			dbStructure.Tokens[k] = token
			if err := db.writeDB(dbStructure); err != nil {
				return token, err
			}
		}
	}
	return token, nil
}