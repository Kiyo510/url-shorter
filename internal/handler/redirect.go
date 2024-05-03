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

	dbAdaptor := adaptor.NewPostgresAdaptor()
	db, err := dbAdaptor()
	if err != nil {
		log.Println("Failed to connect to database: ", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	err = db.Get(&originalURL, "SELECT original_url FROM short_url_mappings WHERE hash = $1", hash)
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
