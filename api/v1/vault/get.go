package vault

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rokusei/gopass-server/db"
	"gorm.io/gorm"
)

type getVaultAPI struct {
	db *gorm.DB
}

func GetVaultAPI(db *gorm.DB) http.Handler {
	return &getVaultAPI{db}
}

func (c *getVaultAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	authHash := r.FormValue("auth-hash")

	//fmt.Printf("email: %s, authHash: %s", email, authHash)

	user, err := db.GetVerifiedUser(r.Context(), c.db, email, []byte(authHash))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vault, err := db.GetVault(r.Context(), c.db, user)
	fmt.Printf("user: %#v\nvault: %#v\nvault entries: %#v", user, vault, vault.VaultEntries)

	b, err := json.Marshal(vault)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(b)
}
