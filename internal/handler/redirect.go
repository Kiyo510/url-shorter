package handler

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/Kiyo510/url-shorter/internal/infrastructure/adaptor"
	"github.com/Kiyo510/url-shorter/internal/utils"
)

func RedirectURL(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Path[1:]
	var originalURL string

	ada, err := adaptor.NewPostgresAdapter()
	if err != nil {
		log.Fatal(err)
	}

	err = ada.Get(&originalURL, "SELECT original_url FROM short_url_mappings WHERE hash = $1", hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		log.Println("Failed to get data: ", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}
