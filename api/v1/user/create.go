package user

import (
	"encoding/json"
	"net/http"

	"github.com/rokusei/gopass-server/db"
	"gorm.io/gorm"
)

type createUserAPI struct {
	db *gorm.DB
}

func CreateUserAPI(db *gorm.DB) http.Handler {
	return &createUserAPI{db}
}

func (c *createUserAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	authHash := r.FormValue("auth-hash")

	user, err := db.CreateUser(r.Context(), c.db, email, []byte(authHash))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(b)
}
