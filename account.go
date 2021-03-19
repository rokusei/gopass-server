package gopass_server

import (
	"context"
	"encoding/hex"
	"errors"

	"crypto/rand"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrInvalidAuthHash = errors.New("invalid AuthenticationHash")

const GenerateUUIDRetries = 10
const AuthHashSize = 64

// a User is identified by an ID or Email
// their identity is verified by verifying a hash of their AuthenticationHash
// which is originally derived from their EncryptionKeyHash and MasterPassword
type User struct {
	gorm.Model
	UUID         string `gorm:"primaryKey"`
	Email        string
	AuthHashHash []byte
	Vault        Vault
}

// GenerateUUID generates a hex encoded 16byte UUID (128 bits) using crypto/rand
func GenerateUUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	h := hex.EncodeToString(b)
	return h, nil
}

// CreateUser generates a bcrypt hash of their AuthenticationHash
// and creates a new vault for them in the database
// Note: This hash doesn't require a salt for two reasons:
//      1) the AuthenticationHash has a large random salt already prepended
//      2) we're hashing a 101,102 iteration PBKDF2 hash, so good luck creating a rainbow table of those
func CreateUser(ctx context.Context, db *gorm.DB, email string, authenticationHash []byte) (User, error) {
	// check if authHash is right size
	if len(authenticationHash) != AuthHashSize {
		return User{}, ErrInvalidAuthHash
	}

	// check if user with this email exists
	result := db.First(&User{}).Where("Email = ?", email)
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return User{}, ErrUserAlreadyExists
	}

	// generate a UUID for the user
	userUUID, err := GenerateUUID()
	if err != nil {
		return User{}, err
	}

	// generate hash of the users authenticationHash using MaxCost
	authHashHash, _ := bcrypt.GenerateFromPassword(authenticationHash, bcrypt.MaxCost)
	user := User{
		UUID:         userUUID,
		Email:        email,
		AuthHashHash: authHashHash,
		Vault:        Vault{},
	}

	// create user, commit changes
	result = db.Create(user).Commit()
	if result.Error != nil {
		return User{}, result.Error
	}
	return user, nil
}

// Fetch a reference to a user, requires authentication of the user's AuthenticationHash
func GetUser(ctx context.Context, db *gorm.DB, id uint64, authenticationHash []byte) (User, error) {
	var user User
	result := db.First(user, id)
	if result.Error != nil {
		return User{}, result.Error
	}

	// verify user's authentication hash
	err := bcrypt.CompareHashAndPassword(user.AuthHashHash, authenticationHash)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
