package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	db *sqlx.DB
)

func init() {
	initDB()
}

func initDB() {
	// TODO:あとで環境変数にする
	var err error
	db, err = sqlx.Open("postgres", "user=postgres password=postgres host=localhost dbname=postgres sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to DB: ", err)
	}
}

func generateHash(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])[:10]
}

type RequestData struct {
	OriginalURL string `json:"original_url"`
}

func shortenURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	//TODO: バリデーション

	var originalUrl, hash string
	var requestData RequestData
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		log.Println("Failed to decode request body: ", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	originalUrl = requestData.OriginalURL
	hash = generateHash(originalUrl)
	var existingHash string
	err = db.Get(&existingHash, "SELECT hash FROM short_url_mappings WHERE hash = $1", hash)
	//TODO: URLのホストは環境変数にする
	shortUrl := "http://localhost:8080/" + hash
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
		shortUrl = "http://localhost:8080/" + existingHash
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

func main() {
	http.HandleFunc("/shorten", shortenURLHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
