package entry

import (
	"encoding/json"
	"net/http"

	"github.com/rokusei/gopass-server/db"
	"gorm.io/gorm"
)

type updateVaultEntryAPI struct {
	db *gorm.DB
}

func UpdateVaultEntryAPI(db *gorm.DB) http.Handler {
	return &updateVaultEntryAPI{db}
}

func (u *updateVaultEntryAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	authHash := r.FormValue("auth-hash")
	entryUUID := r.FormValue("entry-uuid")
	encEntry := r.FormValue("encrypted-entry")

	// Get the user, which in the process authenticates the request
	user, err := db.GetVerifiedUser(r.Context(), u.db, email, []byte(authHash))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the specified entry by UUID
	entry, err := db.UpdateVaultEntry(r.Context(), u.db, user, entryUUID, []byte(encEntry))
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
