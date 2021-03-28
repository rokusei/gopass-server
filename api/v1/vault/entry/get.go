package entry

import (
	"encoding/json"
	"net/http"

	"github.com/rokusei/gopass-server/db"
	"gorm.io/gorm"
)

type getVaultEntryAPI struct {
	db *gorm.DB
}

func GetVaultEntryAPI(db *gorm.DB) http.Handler {
	return &getVaultEntryAPI{db}
}

func (c *getVaultEntryAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	authHash := r.FormValue("auth-hash")
	entryUUID := r.FormValue("entry-uuid")

	// Get the user, which in the process authenticates the request
	user, err := db.GetVerifiedUser(r.Context(), c.db, email, []byte(authHash))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the requested vault entry by UUID
	entry, err := db.GetVaultEntry(r.Context(), c.db, user, entryUUID)
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
