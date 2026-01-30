package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	kafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Event struct {
	Type string                 `json:"type"` // User/Payment/Movie
	Data map[string]interface{} `json:"data"`
}

var producer *kafka.Producer
var consumer *kafka.Consumer

func initKafka() {
	var err error

	producer, err = kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("KAFKA_BROKERS"),
	})
	if err != nil {
		log.Fatal("Failed to create producer: ", err)
	}

	consumer, err = kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("KAFKA_BROKERS"),
		"group.id":          os.Getenv("GROUP_ID"),
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatal("Failed to create consumer: ", err)
	}

	topic := os.Getenv("TOPIC_NAME")
	if topic == "" {
		topic = "cinemaabyss.events"
	}
	if err := consumer.SubscribeTopics([]string{topic}, nil); err != nil {
		log.Fatal("Failed to subscribe to topic: ", err)
	}
}

func produceEvent(eventType string, data map[string]interface{}) {
	event := Event{Type: eventType, Data: data}
	msg, _ := json.Marshal(event)

	topic := "cinemaabyss.events"
	err := producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          msg,
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
	brokers := os.Getenv("KAFKA_BROKERS")
	groupID := os.Getenv("GROUP_ID")

	if brokers == "" || groupID == "" {
		log.Fatal("KAFKA_BROKERS and GROUP_ID must be set")
	}

	initKafka()
	go consumeEvents()

	http.HandleFunc("/events", eventHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Events service listening on :%s", port)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)
	<-sigchan

	if consumer != nil {
		consumer.Close()
	}
	if producer != nil {
		producer.Close()
	}
}
