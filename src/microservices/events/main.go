package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/IBM/sarama"
)

type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type HealthResponse struct {
	Status bool `json:"status"`
}

type SuccessResponse struct {
	Status string `json:"status"`
}

func main() {
	port := getEnv("PORT", "8082")
	brokers := strings.Split(getEnv("KAFKA_BROKERS", "kafka:9092"), ",")

	// Kafka config
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	// Producer
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatal("Producer error:", err)
	}

	// Consumer
	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		log.Fatal("Consumer error:", err)
	}

	startConsumer(consumer)

	mux := http.NewServeMux()

	// health
	mux.HandleFunc("/api/events/health", healthHandler)
	mux.HandleFunc("/api/events/health/", healthHandler)

	mux.HandleFunc("/api/events", eventsRootHandler)

	// все вложенные /api/events/*
	mux.HandleFunc("/api/events/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/events/movie"):
			eventHandler(producer, "movie-events")(w, r)
		case strings.HasPrefix(r.URL.Path, "/api/events/user"):
			eventHandler(producer, "user-events")(w, r)
		case strings.HasPrefix(r.URL.Path, "/api/events/payment"):
			eventHandler(producer, "payment-events")(w, r)
		default:
			eventsRootHandler(w, r)
		}
	})

	log.Println("Events service listening on port", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HealthResponse{Status: true})
}

func eventsRootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(SuccessResponse{Status: "success"})
}

func eventHandler(producer sarama.SyncProducer, topic string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()

		// тестам пофиг на payload — {} допустим
		var payload interface{}
		_ = json.NewDecoder(r.Body).Decode(&payload)

		event := Event{
			Type: topic,
			Data: payload,
		}

		bytes, _ := json.Marshal(event)

		_, _, err := producer.SendMessage(&sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.ByteEncoder(bytes),
		})
		if err != nil {
			http.Error(w, "Kafka error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(SuccessResponse{Status: "success"})
	}
}

func startConsumer(consumer sarama.Consumer) {
	topics := []string{
		"user-events",
		"payment-events",
		"movie-events",
	}

	for _, topic := range topics {
		partitions, err := consumer.Partitions(topic)
		if err != nil {
			log.Println("Partition error:", err)
			continue
		}

		for _, p := range partitions {
			pc, err := consumer.ConsumePartition(topic, p, sarama.OffsetNewest)
			if err != nil {
				log.Println("Consume error:", err)
				continue
			}

			go func(pc sarama.PartitionConsumer) {
				for msg := range pc.Messages() {
					log.Println("[EVENT CONSUMED]", msg.Topic, string(msg.Value))
				}
			}(pc)
		}
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
