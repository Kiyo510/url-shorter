package handler

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/Kiyo510/url-shorter/internal/config"
	"github.com/Kiyo510/url-shorter/internal/infrastructure/adaptor"
	"github.com/Kiyo510/url-shorter/internal/utils"
	"github.com/redis/rueidis"
)

func RedirectURL(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Path[1:]
	var originalURL string

	conf := config.RedisConf
	client, err := rueidis.NewClient(rueidis.ClientOption{InitAddress: []string{conf.Host + ":" + conf.Port}})
	if err != nil {
		log.Println("Failed to connect to redis: ", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	defer client.Close()
	ctx := context.Background()
	originalURL, err = client.Do(ctx, client.B().Get().Key(hash).Build()).ToString()
	if err == nil && !rueidis.IsRedisNil(err) {
		log.Println("Redirecting to original URL: ", originalURL)
		http.Redirect(w, r, originalURL, http.StatusFound)
		return
	}
	if err != nil && !rueidis.IsRedisNil(err) {
		log.Println("Failed to get data from redis: ", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

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
	err = client.Do(ctx, client.B().Set().Key(hash).Value(originalURL).Nx().ExSeconds(100).Build()).Error()
	if err != nil {
		log.Println("Failed to set data to redis: ", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}
