package main

import (
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	monolithURL      string
	moviesServiceURL string
	gradualMigration bool
	migrationPercent int
)

func init() {
	monolithURL = os.Getenv("MONOLITH_URL")
	moviesServiceURL = os.Getenv("MOVIES_SERVICE_URL")
	gradualMigration = os.Getenv("GRADUAL_MIGRATION") == "true"
	migrationPercent, _ = strconv.Atoi(os.Getenv("MOVIES_MIGRATION_PERCENT"))
}

func main() {
	http.HandleFunc("/api/movies", moviesHandler)
	http.HandleFunc("/", defaultHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      http.DefaultServeMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	println("Proxy service listening on :" + port)
	if err := server.ListenAndServe(); err != nil {
		println("Server error:", err.Error())
	}
}

func moviesHandler(w http.ResponseWriter, r *http.Request) {
	targetURL := monolithURL

	if gradualMigration {
		// Случайный выбор: монолит или микросервис
		if rand.Intn(100) < migrationPercent {
			targetURL = moviesServiceURL
		}
	}

	proxyRequest(w, r, targetURL)
}

func proxyRequest(w http.ResponseWriter, r *http.Request, targetURL string) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest(r.Method, targetURL+r.URL.Path, r.Body)
	req.Header = r.Header

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Копируем заголовки
	for k, v := range resp.Header {
		w.Header().Set(k, v[0])
	}
	w.WriteHeader(resp.StatusCode)

	// Копируем тело ответа
	_, _ = io.Copy(w, resp.Body)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not Found", http.StatusNotFound)
}
