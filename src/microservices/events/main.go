package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Event struct {
	Type string                 `json:"type"` // User/Payment/Movie
	Data map[string]interface{} `json:"data"`
}

var producer *kafka.Produce
var consumer *kafka.Consumer

func initKafka() {
	// Настройка продюсера
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka:9092",
	})
	if err != nil {
		log.Fatal("Failed to create producer: ", err)
	}
	producer = producer

	// Настройка потребителя
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  "kafka:9092",
		"group.id":          "events-group",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatal("Failed to create consumer: ", err)
	}
	consumer = consumer

	err = consumer.SubscribeTopics([]string{"cinemaabyss.events"}, nil)
	if err != nil {
		log.Fatal("Failed to subscribe: ", err)
	}
}

func produceEvent(eventType string, data map[string]interface{}) {
	event := Event{Type: eventType, Data: data}
	msg, _ := json.Marshal(event)

	topic := "cinemaabyss.events"
	err := producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:      msg,
	}, nil)
	if err != nil {
		log.Printf("Failed to produce message: %v", err)
	} else {
		log.Printf("Event produced: %s", string(msg))
	}
}

func consumeEvents() {
	go func() {
		for {
			msg, err := consumer.ReadMessage(100)
			if err == nil {
				log.Printf("Consumed event: %s", string(msg.Value))
			} else if err.(kafka.Error).Code() == kafka.ErrTimedOut {
				continue
			} else {
				log.Printf("Consumer error: %v", err)
				break
			}
		}
	}()
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	var event Event
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	produceEvent(event.Type, event.Data)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Event received"))
}

func main() {
	initKafka()
	go consumeEvents()

	http.HandleFunc("/events", eventHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Events service listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
