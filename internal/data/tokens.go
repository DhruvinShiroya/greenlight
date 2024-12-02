package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"github.com/DhruvinShiroya/greenlight/internal/validator"
	"golang.org/x/exp/rand"
)

const (
	ScopeActivation = "activation"
  ScopeAuthentication = "authentication"
)

// token for authentication
type Token struct {
	Plaintext string  `json:"token"`
	Hash      []byte `json:"-"`
	UserID    int64 `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string `json:"-"`
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	// create token instance containing the user ID
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomByte := make([]byte, 16)
	// used Read() to from crypto/rand package to fill the byte slice with random bytes
	// from OS's CSPRNG
	_, err := rand.Read(randomByte)
	if err != nil {
		return nil, err
	}

	// encode random byte and specify for NoPadding
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomByte)

	// generate hash of plaintext for database
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]
	return token, nil
}

func ValidateTokenPlainText(v *validator.Validator, tokenStr string) {
	v.Check(tokenStr != "", "Token", "must not be empty")
	v.Check(len(tokenStr) == 26, "Token", "must be 26 byte long")
}

// define Token Model
type TokenModel struct {
	DB *sql.DB
}

// new method to create token and insert into database
func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}
	err = m.Insert(token)
	return token, err
}

func (m TokenModel) Insert(token *Token) error {
	query := `INSERT INTO tokens(hash, user_id,expiry, scope) VALUES ($1, $2, $3, $4)`

	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (m TokenModel) DeleteAllForUser(userID int64, scope string) error {
	query := `DELETE FROM tokens WHERE scope = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}
