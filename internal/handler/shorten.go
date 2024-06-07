package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/Kiyo510/url-shorter/internal/config"
	"github.com/Kiyo510/url-shorter/internal/infrastructure/adaptor"
	"github.com/Kiyo510/url-shorter/internal/utils"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/redis/rueidis"
)

type RequestData struct {
	OriginalURL string `json:"original_url" validate:"required,http_url"`
}

type ValidationError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

type validationErrors []ValidationError

func (r RequestData) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.OriginalURL, validation.Required.Error("URLは必須です"), is.URL.Error("URLの形式が正しくありません")),
	)
}

func ShortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}
	var originalURL, hash string
	var requestData RequestData
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		log.Println("Failed to decode request body: ", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	err = requestData.Validate()

	if err != nil {
		var errs validation.Errors
		errors.As(err, &errs)
		var validationErrors validationErrors
		for k, err := range errs {
			validationErrors = append(validationErrors, ValidationError{Field: k, Error: err.Error()})
		}
		utils.RespondWithJSON(w, validationErrors, http.StatusBadRequest)
		return
	}

	originalURL = requestData.OriginalURL
	hash = utils.GenerateHash(originalURL)

	shortUrl, err := findOrCreateShortUrl(hash, originalURL)
	if err != nil {
		log.Println("Failed to find or create short URL: ", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	conf := config.RedisConf

	client, err := rueidis.NewClient(rueidis.ClientOption{InitAddress: []string{conf.Host + ":" + conf.Port}})
	if err != nil {
		log.Println("Failed to connect to redis: ", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	defer client.Close()
	ctx := context.Background()
	err = client.Do(ctx, client.B().Set().Key(hash).Value(originalURL).Nx().ExSeconds(100).Build()).Error()
	if err != nil {
		log.Println("Failed to set data to redis: ", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	utils.RespondWithJSON(w, map[string]string{"short_url": shortUrl}, http.StatusCreated)
}

func findOrCreateShortUrl(hash string, originalURL string) (string, error) {
	var existingUrl string

	dbAdaptor := adaptor.NewPostgresAdaptor()
	db, err := dbAdaptor()
	if err != nil {
		return "", fmt.Errorf("failed to connect to database: %w", err)
	}

	err = db.Get(&existingUrl, "SELECT original_url FROM short_url_mappings WHERE hash = $1", hash)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		_, err = db.Exec("INSERT INTO short_url_mappings (original_url, hash) VALUES ($1, $2)", originalURL, hash)
		if err != nil {
			return "", fmt.Errorf("failed to insert data: %w", err)
		}

		return config.AppConf.BaseUrl + "/" + hash, nil
	case err != nil:
		return "", fmt.Errorf("failed to get data: %w", err)
	case existingUrl == originalURL:
		return config.AppConf.BaseUrl + "/" + hash, nil
	case existingUrl != originalURL:
		log.Println("Hash collision detected. Retrying with a new hash.")
		const HashCollisionSuffix = "collision"

		return findOrCreateShortUrl(utils.GenerateHash(originalURL+HashCollisionSuffix), originalURL)
	default:
		return "", fmt.Errorf("unexpected error: %w", err)
	}
}
