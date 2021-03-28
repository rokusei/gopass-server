package api

import (
	"net/http"

	"github.com/rokusei/gopass-server/api/v1/user"
	"github.com/rokusei/gopass-server/api/v1/vault"
	"github.com/rokusei/gopass-server/api/v1/vault/entry"
	"gorm.io/gorm"
)

type APIConfig struct {
	DB *gorm.DB
}

type api struct {
	*http.ServeMux
}

func NewAPI(apiConfig APIConfig) http.Handler {
	mux := http.NewServeMux()

	// user
	mux.Handle("/user", user.GetUserAPI(apiConfig.DB))
	mux.Handle("/user/create", user.CreateUserAPI(apiConfig.DB))

	// vault
	mux.Handle("/vault", vault.GetVaultAPI(apiConfig.DB))

	// vault/entry
	mux.Handle("/vault/entry", entry.GetVaultEntryAPI(apiConfig.DB))
	mux.Handle("/vault/entry/create", entry.CreateVaultEntryAPI(apiConfig.DB))
	mux.Handle("/vault/entry/update", entry.UpdateVaultEntryAPI(apiConfig.DB))
	return &api{mux}
}
