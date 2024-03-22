package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
	return hex.EncodeToString(hash[:10])
}

type RequestData struct {
	OriginalURL string `json:"original_url"`
}

func shortenURLHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: GETの場合はエラーを返す
	//TODO: バリデーション
	//TODO: Errorメッセージを返す

	var originalUrl, hash string
	if r.Method == "POST" {
		var requestData RequestData
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			log.Println("Failed to decode request body: ", err)
		}

		originalUrl = requestData.OriginalURL
		hash = generateHash(originalUrl)
	}

	_, err := db.Exec("INSERT INTO short_url_mappings (original_url, hash) VALUES ($1, $2)", originalUrl, hash)
	if err != nil {
		log.Fatal("Failed to insert data: ", err)
	}

	w.Write([]byte(`{"result": "ok"}`))
}

func main() {
	http.HandleFunc("/shorten", shortenURLHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
