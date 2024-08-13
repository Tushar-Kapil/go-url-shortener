package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var rdb *redis.Client

func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "url-shortener-redis-ekg2sg.serverless.aps1.cache.amazonaws.com:6379",
		Password: "",
		DB: 0,
	})

	http.HandleFunc("/shorten", shortenUrl)

	fmt.Println("Server Running On Port 8080")
	http.ListenAndServe(":8080", nil);
}

type ShortenRequest struct {
	LongUrl string `json:"longUrl"`
}

type ShortenResponse struct {
	ShortUrl string `json:"shortUrl"`
}

func shortenUrl(w http.ResponseWriter , r *http.Request){
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if req.LongUrl == "" {
		http.Error(w, "Missing 'longUrl' field", http.StatusBadRequest)
		return
	}
	shortUrl := generateShortUrl()

	err = rdb.Set(ctx, shortUrl, req.LongUrl, 24*time.Hour).Err()

	if err != nil {
		http.Error(w, "Failed to store URL", http.StatusInternalServerError)
		return
	}

	response := ShortenResponse{ShortUrl: "http://localhost:8080/" + shortUrl}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

func generateShortUrl() string {
	b := make([]byte, 6)
	_, err := rand.Read(b)

	if err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(b)[:8]
}