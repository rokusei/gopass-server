package db

import (
	"context"
	"encoding/hex"
	"errors"

	"crypto/rand"
	"crypto/sha512"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var UserAlreadyExists = errors.New("user already exists")
var UserDoesNotExist = errors.New("user does not exist")
var InvalidAuthHash = errors.New("invalid AuthenticationHash")

const GenerateUUIDRetries = 10
const AuthHashSize = 64

// a User is identified by an ID or Email
// their identity is verified by verifying a hash of their AuthenticationHash
// which is originally derived from their EncryptionKeyHash and MasterPassword
type User struct {
	gorm.Model
	UUID         string `gorm:"primaryKey"`
	EmailHash    string
	AuthHashHash []byte
	VaultID      uint
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
	if err := ctx.Err(); err != nil {
		return User{}, err
	}

	// check if authHash is right size
	if len(authenticationHash) != AuthHashSize {
		return User{}, InvalidAuthHash
	}

	h := sha512.New()
	h.Write([]byte(email))
	emailHash := hex.EncodeToString(h.Sum(nil))

	// check if user with this email exists
	u := User{}
	result := db.Where("EmailHash = ?", emailHash).Limit(1).Find(&u)
	if result.Error != nil {
		return User{}, result.Error
	}
	if u.UUID != "" {
		return User{}, UserAlreadyExists
	}

	// generate a UUID for the user
	userUUID, err := GenerateUUID()
	if err != nil {
		return User{}, err
	}

	vault := Vault{
		Entries: make([]VaultEntry, 0),
	}
	result = db.Create(&vault)

	// generate hash of the users authenticationHash using MaxCost
	authHashHash, _ := bcrypt.GenerateFromPassword(authenticationHash, bcrypt.DefaultCost)
	user := User{
		UUID:         userUUID,
		EmailHash:    emailHash,
		AuthHashHash: authHashHash,
		VaultID:      vault.ID,
	}

	// create user
	result = db.Create(&user)
	if result.Error != nil {
		return User{}, result.Error
	}
	db.Commit()
	return user, nil
}

// Fetch a reference to a user, requires authentication of the user's AuthenticationHash
func GetUser(ctx context.Context, db *gorm.DB, email string, authenticationHash []byte) (User, error) {
	if err := ctx.Err(); err != nil {
		return User{}, err
	}

	h := sha512.New()
	h.Write([]byte(email))
	emailHash := hex.EncodeToString(h.Sum(nil))
	user := User{}
	result := db.Where("EmailHash = ?", emailHash).Limit(1).Find(&user)
	if result.Error != nil {
		return User{}, result.Error
	}

	if len(user.AuthHashHash) < 1 || user.UUID == "" {
		return user, UserDoesNotExist
	}

	// verify user's authentication hash
	err := bcrypt.CompareHashAndPassword(user.AuthHashHash, authenticationHash)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
