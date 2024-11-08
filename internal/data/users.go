package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/DhruvinShiroya/greenlight/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

// Define user struct for individual user
// hide details like password and version number with `json:"-"` tag for details not to addpear in json
// also create a custom password type
type User struct {
	ID        int64     `json:"id"`
	CreateAt  time.Time `json:"created_at"`
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

func ValidatePassword(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "password length must be greater than 8 bytes long")
	v.Check(len(password) <= 75, "password", "password length must be less than 75 bytes long")
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "email must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be provided valid email address this is error")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 200, "name", "must be under 200 bytes")

	// call for email validation
	ValidateEmail(v, user.Email)
	// validate password
	if user.Password.plaintext != nil {
		ValidatePassword(v, *user.Password.plaintext)
	}

	// if there is password hash which is nil, will indicate that
	// our code has logic error and hence put sanity check to include here
	// and raise panic instead
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

// error for duplicate email
var (
	ErrDuplicateEmail = errors.New("Please provide new email, which is not registred")
)

// create userModel for db interactions
type UserModel struct {
	DB *sql.DB
}

// add new user to the database
func (m UserModel) Insert(user *User) error {
	query := `
    INSERT INTO users (name, email, activated, password_hash)
    VALUES ($1 , $2, $3, $4)
    RETURNING id, created_at, version`

	args := []interface{}{user.Name, user.Email, user.Activated, user.Password.hash}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreateAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

// Retrive user by Email
func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `SELECT SELECT id, created_at, name, email, password_hash, activated, version
                  FROM users WHERE email = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreateAt,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// update user details
func (m UserModel) UpdateUser(user *User) error {
	query := `
    UPDATE users
    SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
    WHERE id = $5 AND version = $6
    RETURNING version
  `

	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}
