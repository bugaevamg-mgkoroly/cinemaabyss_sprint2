package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// Конфигурация миграции
var migrationPercent int

func init() {
	percentStr := os.Getenv("MOVIES_MIGRATION_PERCENT")
	migrationPercent, _ = strconv.Atoi(percentStr)
	if migrationPercent < 0 || migrationPercent > 100 {
		migrationPercent = 0
	}
}

// Обработчик для /api/movies
func moviesHandler(w http.ResponseWriter, r *http.Request) {
	// Решение: направить запрос в монолит или в микросервис movies
	useMicroservice := time.Now().UnixNano()%100 < int64(migrationPercent)


	var resp *http.Response
	var err error

	if useMicroservice {
		resp, err = http.Get("http://movies-service:8081/movies")
	} else {
		resp, err = http.Get("http://monolith:8080/api/movies")
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	defer resp.Body.Close()

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/movies", moviesHandler).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Proxy service listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
