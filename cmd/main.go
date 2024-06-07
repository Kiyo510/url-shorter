package main

import (
	"log"
	"net/http"

	"github.com/Kiyo510/url-shorter/internal/config"
	"github.com/Kiyo510/url-shorter/internal/handler"
)

func main() {
	config.LoadDBConfig()
	config.LoadAppConfig()
	config.LoadRedisConfig()

	http.HandleFunc("/shorten", handler.ShortenURL)
	http.HandleFunc("/", handler.RedirectURL)
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
