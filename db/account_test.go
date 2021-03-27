package db_test

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"regexp"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/rokusei/gopass"
	"github.com/rokusei/gopass-server/db"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func Test_CreateUser(t *testing.T) {
	mdb, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mdb.Close()
	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: mdb,
	}), &gorm.Config{})
	require.NoError(t, err)

	testCases := []struct {
		email      string
		masterPass string
	}{
		{
			email:      "abc@123.com",
			masterPass: "abc123",
		},
	}

	for _, test := range testCases {
		authHash, _, _, err := gopass.GenerateAuthEncHashes(test.masterPass)
		require.NoError(t, err)

		h := sha512.New()
		h.Write([]byte(test.email))
		emailHash := hex.EncodeToString(h.Sum(nil))

		//mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "users"
		WHERE EmailHash = $1 AND "users"."deleted_at" IS NULL LIMIT 1`)).
			WithArgs(emailHash).
			WillReturnRows(sqlmock.NewRows([]string{}))

		mock.ExpectQuery(regexp.QuoteMeta(
			`INSERT INTO "vaults" ("created_at","updated_at","deleted_at") VALUES ($1,$2,$3) RETURNING "id"`)).
			WillReturnRows(sqlmock.NewRows([]string{"123"}))
		mock.ExpectQuery(regexp.QuoteMeta(
			`INSERT INTO "users" ("created_at","updated_at","deleted_at","uuid","email_hash","auth_hash_hash","vault_id") 
			VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`)).
			WillReturnRows(sqlmock.NewRows([]string{"a", "b", "c"}))
		mock.ExpectCommit()

		u, err := db.CreateUser(context.Background(), gdb, test.email, authHash)
		require.NoError(t, err)
		require.Equal(t, emailHash, u.EmailHash)
		err = bcrypt.CompareHashAndPassword(u.AuthHashHash, authHash)
		require.NoError(t, err)
	}
}

func Test_CreateUserDeadlineExceeded(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*-1))
	defer cancel()
	_, err := db.CreateUser(ctx, &gorm.DB{}, "", []byte{})
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func Test_CreateUserDeadlineCancelled(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*10))
	cancel()
	_, err := db.CreateUser(ctx, &gorm.DB{}, "", []byte{})
	require.ErrorIs(t, err, context.Canceled)
}

func Test_GetUser(t *testing.T) {
	mdb, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mdb.Close()
	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: mdb,
	}), &gorm.Config{})
	require.NoError(t, err)

	testCases := []struct {
		email      string
		masterPass string
	}{
		{
			email:      "abc@123.com",
			masterPass: "abc123",
		},
	}

	for _, test := range testCases {
		authHash, _, _, err := gopass.GenerateAuthEncHashes(test.masterPass)
		require.NoError(t, err)

		h := sha512.New()
		h.Write([]byte(test.email))
		emailHash := hex.EncodeToString(h.Sum(nil))

		// Create User Queries
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "users"
		WHERE EmailHash = $1 AND "users"."deleted_at" IS NULL LIMIT 1`)).
			WithArgs(emailHash).
			WillReturnRows(sqlmock.NewRows([]string{}))
		mock.ExpectQuery(regexp.QuoteMeta(
			`INSERT INTO "vaults" ("created_at","updated_at","deleted_at") VALUES ($1,$2,$3) RETURNING "id"`)).
			WillReturnRows(sqlmock.NewRows([]string{"123"}))
		mock.ExpectQuery(regexp.QuoteMeta(
			`INSERT INTO "users" ("created_at","updated_at","deleted_at","uuid","email_hash","auth_hash_hash","vault_id") 
			VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`)).
			WillReturnRows(sqlmock.NewRows([]string{"1"}))

		_, err = db.CreateUser(context.Background(), gdb, test.email, authHash)
		require.NoError(t, err)

		authHashHash, _ := bcrypt.GenerateFromPassword(authHash, bcrypt.DefaultCost)
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "users"
			WHERE EmailHash = $1 AND "users"."deleted_at" IS NULL LIMIT 1`)).
			WithArgs(emailHash).
			WillReturnRows(sqlmock.NewRows([]string{"uuid", "email_hash", "auth_hash_hash"}).AddRow("123", emailHash, authHashHash))
		u, err := db.GetUser(context.Background(), gdb, test.email, authHash)
		require.NoError(t, err)
		require.Equal(t, emailHash, u.EmailHash)
		err = bcrypt.CompareHashAndPassword(u.AuthHashHash, authHash)
		require.NoError(t, err)
	}
}

func Test_GetUserDeadlineExceeded(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*-1))
	defer cancel()
	_, err := db.GetUser(ctx, &gorm.DB{}, "", []byte{})
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func Test_GetUserDeadlineCancelled(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*10))
	cancel()
	_, err := db.GetUser(ctx, &gorm.DB{}, "", []byte{})
	require.ErrorIs(t, err, context.Canceled)
}
