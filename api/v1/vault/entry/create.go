package entry

import (
	"encoding/json"
	"net/http"

	"github.com/rokusei/gopass-server/db"
	"gorm.io/gorm"
)

type createVaultEntryAPI struct {
	db *gorm.DB
}

func CreateVaultEntryAPI(db *gorm.DB) http.Handler {
	return &createVaultEntryAPI{db}
}

func (c *createVaultEntryAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	authHash := r.FormValue("auth-hash")
	encEntry := r.FormValue("encrypted-entry")

	// Get the user, which in the process authenticates the request
	user, err := db.GetVerifiedUser(r.Context(), c.db, email, []byte(authHash))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create the vault entry
	entry, err := db.CreateVaultEntry(r.Context(), c.db, user, []byte(encEntry))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(entry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
}
