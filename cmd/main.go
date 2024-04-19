package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

var (
	db *sqlx.DB
)

func init() {
	initDB()
}

func initDB() {
	dbName := os.Getenv("DB_NAME")
	dsn := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=%s", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), dbName, "disable")
	var err error
	db, err = sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to DB: ", err)
	}
}

func generateHash(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])[:10]
}

type RequestData struct {
	OriginalURL string `json:"original_url" validate:"required,http_url"`
}
type ValidationError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}
type validationErrors []ValidationError

func shortenURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var originalUrl, hash string
	var requestData RequestData
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		log.Println("Failed to decode request body: ", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
		responseJson, err := json.Marshal(validationErrors)
		if err != nil {
			log.Println("Failed to encode response JSON: ", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(responseJson)
		return
	}

	originalUrl = requestData.OriginalURL
	hash = generateHash(originalUrl)
	var existingHash string

	err = db.Get(&existingHash, "SELECT hash FROM short_url_mappings WHERE hash = $1", hash)
	shortUrl := os.Getenv("APP_URL") + "/" + hash
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, err = db.Exec("INSERT INTO short_url_mappings (original_url, hash) VALUES ($1, $2)", originalUrl, hash)
			if err != nil {
				log.Println("Failed to insert data: ", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		} else {
			log.Println("Failed to get data: ", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		shortUrl = os.Getenv("APP_URL") + "/" + existingHash
	}

	responseData := map[string]string{"short_url": shortUrl}
	responseJson, err := json.Marshal(responseData)
	if err != nil {
		log.Println("Failed to encode response JSON: ", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// クライアントにJSONレスポンスを返す
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJson)
}

func (r RequestData) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.OriginalURL, validation.Required.Error("URLは必須です"), is.URL.Error("URLの形式が正しくありません")),
	)
}

// Street: the length must be between 5 and 50; State: must be in a valid format.

func redirectHandlerOriginalURL(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Path[1:]
	var originalURL string
	err := db.Get(&originalURL, "SELECT original_url FROM short_url_mappings WHERE hash = $1", hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		log.Println("Failed to get data: ", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, originalURL, http.StatusFound)
}

func main() {
	http.HandleFunc("/shorten", shortenURLHandler)
	http.HandleFunc("/", redirectHandlerOriginalURL)
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
