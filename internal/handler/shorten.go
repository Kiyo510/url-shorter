package handler

import (
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

	var originalUrl, hash string
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

	originalUrl = requestData.OriginalURL
	hash = utils.GenerateHash(originalUrl)
	shortUrl, err := findOrCreateShortUrl(hash, originalUrl)
	if err != nil {
		log.Println("Failed to find or create short URL: ", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	utils.RespondWithJSON(w, map[string]string{"short_url": shortUrl}, http.StatusCreated)
}

func findOrCreateShortUrl(hash string, originalUrl string) (string, error) {
	var existingHash string

	dbAdaptor := adaptor.NewPostgresAdaptor()
	db, err := dbAdaptor()
	if err != nil {
		return "", fmt.Errorf("failed to connect to database: %w", err)
	}

	err = db.Get(&existingHash, "SELECT hash FROM short_url_mappings WHERE original_url = $1", originalUrl)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		_, err = db.Exec("INSERT INTO short_url_mappings (original_url, hash) VALUES ($1, $2)", originalUrl, hash)
		if err != nil {
			return "", fmt.Errorf("failed to insert data: %w", err)
		}
	case err != nil:
		return "", fmt.Errorf("failed to get data: %w", err)
	default:
		hash = existingHash
	}

	shortUrl := config.AppConf.BaseUrl + "/" + hash

	return shortUrl, nil
}
