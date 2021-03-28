package server

import (
	"net/http"

	"github.com/rokusei/gopass-server/api"
	"gorm.io/gorm"
)

func Run(db *gorm.DB) {
	apiConfig := api.APIConfig{
		DB: db,
	}
	apiHandler := api.NewAPI(apiConfig)
	http.ListenAndServe(":8080", apiHandler)
}
