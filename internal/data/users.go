package data

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Define user struct for individual user
// hide details like password and version number with `json:"-"` tag for details not to addpear in json
// also create a custom password type
type User struct {
	ID        int64     `json:"id"`
	CreateAt  time.Time `json:"create_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

// custom password for saving plain password text and hashed version.
type password struct {
	plaintext *string
	hash      []byte
}

// set() method calculates hash of a plaintext password, and stores both
// the hash and plaintext version in the struct
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 10)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

// Matches() method checks whether the provided plaintext password matches the hashed
// password, return true if it doesn or false
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}
