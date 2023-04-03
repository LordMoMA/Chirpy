package database

import (
	"errors"
	"fmt"

	"github.com/cloudflare/cfssl/log"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Membership bool `json:"is_chirpy_red"`
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
	// fmt.Printf("hashed password in CreateUser: %s", hashedPassword)

	user := User{
		ID:       id,
		Email:    email,
		Password: string(hashedPassword),
		Membership: false,
	}

	dbStructure.Users[id] = user

	if err := db.writeDB(dbStructure); err != nil {
		return User{}, err
	}

	return user, nil
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

func (db *DB) UpdateUser(userID int, email, password string) (User, error) {
	// Load the current JSON data from the database file
	user, err := db.GetUser(userID)
	if err != nil {
		log.Error(err)
	}

	index := user.ID
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	// Replace the user at that index with the updated user
	user.Email = email
	user.Password = string(hashedPassword)
	dbStructure.Users[index] = user


	err = db.writeDB(dbStructure)
	if err != nil {
		log.Error(err)
	}

	return user, nil
}

func (db *DB) GetUser(userID int) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	for _, user := range dbStructure.Users {
		if user.ID == userID {
			return user, nil
		}
	}
	return User{}, errors.New("User not found")
}

func (db *DB) UpdateMembership(userID int, membership bool) (User, error) {
	// Load the current JSON data from the database file
	user, err := db.GetUser(userID)
	if err != nil {
		log.Error(err)
	}

	index := user.ID
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user.Membership = membership
	dbStructure.Users[index] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		log.Error(err)
	}

	return user, nil
}


