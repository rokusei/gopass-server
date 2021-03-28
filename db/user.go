package db

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrUserNotVerified = errors.New("user is not verified")
var ErrUserAlreadyExists = errors.New("user already exists")
var ErrUserDoesNotExist = errors.New("user does not exist")
var ErrInvalidAuthHash = errors.New("invalid AuthenticationHash")

const GenerateUUIDRetries = 10
const AuthHashSize = 64

// a User is identified by an ID or Email
// their identity is verified by verifying a hash of their AuthenticationHash
// which is originally derived from their EncryptionKeyHash and MasterPassword
type User struct {
	gorm.Model
	ID           uint         `gorm:"primarykey" json:"-"`
	UUID         string       `json:"ID"`
	EmailHash    string       `json:"-"`
	AuthHashHash []byte       `json:"-"`
	Verification Verification `gorm:"embedded"`
	VaultID      uint         `json:"-"`
	Vault        Vault
}

type Verification struct {
	gorm.Model
	Hash      string `json:"-"`
	Completed bool   `gorm:"default:false"`
	Attempts  uint   `gorm:"default:0"`
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
func CreateUser(ctx context.Context, db *gorm.DB, email string, authenticationHash []byte) (*User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// check if authHash is right size
	if len(authenticationHash) != AuthHashSize {
		return nil, ErrInvalidAuthHash
	}

	// check if user with this email exists
	emailHash := StringToEncodedHash(email)
	u := User{}
	result := db.Debug().Where("email_hash = ?", emailHash).Limit(1).Find(&u)
	if result.Error != nil {
		return nil, result.Error
	}
	if u.EmailHash != "" {
		return nil, ErrUserAlreadyExists
	}

	uuid, err := GenerateUUID()
	if err != nil {
		return nil, err
	}

	vuuid, err := GenerateUUID()
	if err != nil {
		return nil, err
	}

	// slightly hacky way to get a cryptographically secure random number
	// between 100000 and 999999 using the crypto rand libraries
	b := make([]byte, 8)
	_, err = rand.Read(b)
	if err != nil {
		return nil, err
	}
	i := binary.BigEndian.Uint64(b)
	vc := (100000) + (i % 900000)
	fmt.Printf("%v", vc)

	// generate hash of the users authenticationHash using MaxCost
	authHashHash, _ := bcrypt.GenerateFromPassword(authenticationHash, bcrypt.DefaultCost)
	user := User{
		UUID:         uuid,
		EmailHash:    emailHash,
		AuthHashHash: authHashHash,
		Verification: Verification{
			Hash: StringToEncodedHash(fmt.Sprintf("%v", vc)),
		},
		Vault: Vault{
			UUID:         vuuid,
			VaultEntries: make([]VaultEntry, 0),
		},
	}

	// create user
	result = db.Debug().Create(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	db.Debug().Save(&user)
	return &user, nil
}

// Fetch a reference to a user, requires authentication of the user's AuthenticationHash
// Does not check if the user is verified
func GetUser(ctx context.Context, db *gorm.DB, email string, authenticationHash []byte) (*User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	user, err := AuthenticateUser(ctx, db, email, authenticationHash)
	if err != nil {
		return nil, err
	}

	err = db.Model(&user).Association("Vault").Find(&user.Vault)
	if err != nil {
		return nil, err
	}

	err = db.Debug().Model(&user.Vault).Association("VaultEntries").Find(&user.Vault.VaultEntries)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Fetch a reference to a user, requires authentication of the user's AuthenticationHash
// Ensures that the user is verified
func GetVerifiedUser(ctx context.Context, db *gorm.DB, email string, authenticationHash []byte) (*User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	user, err := AuthenticateUser(ctx, db, email, authenticationHash)
	if err != nil {
		return nil, err
	}
	if user.Verification.Hash != "" && !user.Verification.Completed {
		return nil, ErrUserNotVerified
	}
	return user, nil
}

func AuthenticateUser(ctx context.Context, db *gorm.DB, email string, authenticationHash []byte) (*User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Validate email, fetch that email's user
	user := User{}
	result := db.Where("email_hash = ?", StringToEncodedHash(email)).Limit(1).Find(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	if len(user.AuthHashHash) < 1 || user.EmailHash == "" {
		return nil, ErrUserDoesNotExist
	}

	// compare provided authentication hash to the bcrypt hash of the fetched user
	err := bcrypt.CompareHashAndPassword(user.AuthHashHash, authenticationHash)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func StringToEncodedHash(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
